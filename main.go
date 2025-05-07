package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// 전역 HTTP 클라이언트 재사용
var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		MaxConnsPerHost:     10,
		MaxIdleConnsPerHost: 10,
	},
}

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

var loc, _ = time.LoadLocation("Asia/Seoul")

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
		url, exists := data.Result[name]
		if !exists || url == "" {
			return false, nil
		}
	}
	return true, nil
}

func getImage(ctx context.Context, key string) (string, error) {
	url := fmt.Sprintf("https://pf.kakao.com/rocket-web/web/profiles/%s/posts", key)
	log.Info().Str("url", url).Msg("Fetching image")
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
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

	now := time.Now().In(loc)
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix() * 1000

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

func fetchAllImages(ctx context.Context, searchTarget map[string]string) map[string]string {
	var wg sync.WaitGroup
	var mu sync.Mutex // map 쓰기 동기화를 위한 뮤텍스

	data := make(map[string]string, len(searchTarget))
	for name, key := range searchTarget {
		wg.Add(1)
		go func(name, key string) {
			defer wg.Done()
			img, err := getImage(ctx, key)
			mu.Lock()
			defer mu.Unlock()
			if err == nil && img != "" {
				data[name] = img
			} else {
				data[name] = ""
			}
		}(name, key)
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
		return errors.New("status code is not 200")
	}
	return nil
}

func main() {
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(loc)
	}

	searchTarget := map[string]string{
		"uncle":  "_FxbaQC", // 삼촌밥차
		"mouse":  "_CiVis",  // 슈마우스
		"jundam": "_vKxgdn", // 정담
	}

	// 예시: 전체 함수에 컨텍스트 추가
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 작업 필요 유무 체크
	check, _ := checkWorkDone(searchTarget)
	if check {
		log.Info().Msg("Work done")
		return
	}

	data := fetchAllImages(ctx, searchTarget)

	if len(data) == 0 {
		log.Error().Msg("No image found")
		return
	}

	checkData := false
	for key, value := range data {
		log.Info().Str(key, value).Msg("Upload image")
		if value != "" {
			checkData = true
		}
	}
	if checkData {
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
