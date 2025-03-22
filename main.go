package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type CheckResponseData struct {
	Result string `json:"result"`
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

func checkWorkDone() (bool, error) {
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

	if data.Result != "" {
		return true, nil
	}
	return false, nil
}

func getImage(key string) (string, error) {
	url := fmt.Sprintf("https://pf.kakao.com/rocket-web/web/profiles/%s/posts", key)
	fmt.Printf("Fetching image from %s\n", url)
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
	// fmt.Println(string(body))

	var data PostsData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", err
	}

	if data.Items == nil || len(data.Items) == 0 {
		return "", errors.New("no items found")
	}

	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location()).Unix() * 1000
	// todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix() * 1000

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

func fetchAllImages(keys []string) []string {
	var wg sync.WaitGroup

	var urls []string
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			img, err := getImage(key)
			fmt.Println(img, err)
			urls = append(urls, img)
		}(key)
	}

	wg.Wait()
	return urls
}

func main() {
	// 작업 필요 유무 체크
	check, err := checkWorkDone()
	fmt.Println(check, err)

	keys := []string{"_FxbaQC", "_CiVis", "_vKxgdn"}

	urls := fetchAllImages(keys)
	fmt.Println(urls)
}
