package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type CheckResponseData struct {
	Result map[string]string `json:"result"`
}

type PostsData struct {
	Items []ItemsData `json:"items"`
}

type ItemsData struct {
	Created_at int64       `json:"created_at"`
	Media      []MediaData `json:"media"`
	Type       string      `json:"type"`
}

type MediaData struct {
	Type      string `json:"type"`
	Large_url string `json:"large_url"`
}

type CrawlingData struct {
}

func checkWorkDone(searchTarget map[string]string) (bool, error) {
	// 작업 완료 여부 체크
	res, err := http.Get("https://lunch.muz.kr?check=true")
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false, errors.New("status code is not 200")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	// fmt.Println(string(body))
	var data CheckResponseData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return false, err
	}
	// searchData 의 value 값이 모두 들어 있는지
	for name, _ := range searchTarget {
		if _, ok := data.Result[name]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func getImage(key string) (string, error) {
	url := fmt.Sprintf("https://pf.kakao.com/rocket-web/web/profiles/%s/posts", key)
	log.Info().Str("url", url).Msg("Fetching image")
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", errors.New("status code is not 200")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var data PostsData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	if data.Items == nil || len(data.Items) == 0 {
		return "", errors.New("no items found")
	}

	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix() * 1000

	// 오늘 작성된 포스트만 필터링
	var todayMedia []MediaData
	for _, item := range data.Items {
		if item.Created_at >= todayMidnight && item.Type == "image" {
			todayMedia = item.Media
			break
		}
	}

	if len(todayMedia) == 0 {
		return "", errors.New("no image found")
	}

	if key == "_FxbaQC" {
		// 삼촌
		return todayMedia[len(todayMedia)-1].Large_url, nil
	} else if key == "_CiVis" {
		// 슈마우스
		return todayMedia[0].Large_url, nil
	} else if key == "_vKxgdn" {
		// 정담
		return todayMedia[0].Large_url, nil
	}
	return "", errors.New("no image found")
}

func fetchAllImages(searchTarget map[string]string) map[string]string {
	var wg sync.WaitGroup

	data := make(map[string]string)
	for name, key := range searchTarget {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			img, err := getImage(key)
			if (err == nil) && (img != "") {
				data[name] = img
			} else {
				data[name] = ""
			}
		}(key)
	}

	wg.Wait()
	return data
}

func uploadImage(data map[string]string) error {
	body, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal data")
		return err
	}
	resp, err := http.Post("https://lunch.muz.kr", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("Failed to upload image")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error().Msg("status code is not 200")
		return err
	}
	return nil
}

func main() {
	searchTarget := map[string]string{
		"uncle":  "_FxbaQC", // 삼촌밥차
		"mouse":  "_CiVis",  // 슈마우스
		"jundam": "_vKxgdn", // 정담
	}

	// 작업 필요 유무 체크
	check, err := checkWorkDone(searchTarget)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check work done")
		return
	}
	if check {
		log.Info().Msg("Work done")
		return
	}

	data := fetchAllImages(searchTarget)

	if len(data) != 0 {
		for key, value := range data {
			log.Info().Str(key, value).Msg("Upload image")
		}
		err := uploadImage(data)
		if err != nil {
			log.Error().Err(err).Msg("Image upload failed")
		} else {
			log.Info().Msg("Image uploaded")
		}
	} else {
		log.Error().Msg("No image found")
	}
}
