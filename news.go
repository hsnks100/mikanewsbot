package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	baseURL = "https://openapi.naver.com/v1/search/news.json"
)

type NewsResponse struct {
	LastBuildDate string     `json:"lastBuildDate"`
	Total         int        `json:"total"`
	Start         int        `json:"start"`
	Display       int        `json:"display"`
	Items         []NewsItem `json:"items"`
}

type NewsItem struct {
	Title        string `json:"title"`
	OriginalLink string `json:"originallink"`
	Link         string `json:"link"`
	Description  string `json:"description"`
	PubDate      string `json:"pubDate"`
	Timestamp    int64  `json:"timestamp"`
}

// fetchNaverNewsJSON 네이버 뉴스 API를 호출하여 JSON 데이터를 반환합니다.
func fetchNaverNewsJSON(clientID, clientSecret, query string) (*NewsResponse, error) {
	escapedQuery := url.QueryEscape(query)
	reqURL := fmt.Sprintf("%s?query=%s", baseURL, escapedQuery)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Naver-Client-Id", clientID)
	req.Header.Add("X-Naver-Client-Secret", clientSecret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result NewsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// cleanTitle 단순 HTML 태그 제거 등으로 제목을 정리합니다.
func cleanTitle(s string) string {
	s = strings.ReplaceAll(s, "<b>", "")
	s = strings.ReplaceAll(s, "</b>", "")
	return s
}

// cleanDesc 단순 HTML 태그 제거 등으로 설명을 정리합니다.
func cleanDesc(s string) string {
	s = strings.ReplaceAll(s, "<b>", "")
	s = strings.ReplaceAll(s, "</b>", "")
	return s
}

// GetNewsList 각 쿼리에 대해 네이버 뉴스 API를 호출하고 뉴스 항목을 정리하여 반환합니다.
func GetNewsList(clientID, clientSecret string, queryList []string, timeWindowHours int) ([]NewsItem, error) {
	var newsList []NewsItem
	for _, query := range queryList {
		response, err := fetchNaverNewsJSON(clientID, clientSecret, query)
		if err != nil {
			log.Printf("뉴스 API 호출 오류: %v", err)
			continue
		}
		items := response.Items
		for _, item := range items {
			newsMap := item
			title := cleanTitle(newsMap.Title)
			desc := cleanDesc(newsMap.Description)
			link := newsMap.OriginalLink
			if link == "" {
				link = newsMap.Link
			}
			pubDateRaw := newsMap.PubDate
			pubTime, err := time.Parse(time.RFC1123Z, pubDateRaw)
			if err != nil {
				// RFC1123Z 파싱 실패 시 RFC1123로 시도
				pubTime, err = time.Parse(time.RFC1123, pubDateRaw)
				if err != nil {
					log.Printf("날짜 파싱 오류: %v", err)
					continue // 파싱에 실패하면 해당 뉴스 항목은 건너뜁니다.
				}
			}
			if time.Since(pubTime) > 24*time.Hour {
				continue // 이내의 뉴스만 필터링합니다.
			}

			formattedDate := pubTime.In(time.Local).Format("2006/01/02 15:04")
			newsList = append(newsList, NewsItem{
				Title:       title,
				Description: desc,
				Link:        link,
				PubDate:     formattedDate,
				Timestamp:   time.Now().Unix(),
			})
		}
	}

	// 날짜를 기준으로 오름차순 정렬 (문자열을 시간으로 파싱)
	sort.Slice(newsList, func(i, j int) bool {
		t1, err1 := time.Parse("2006/01/02 15:04", newsList[i].PubDate)
		t2, err2 := time.Parse("2006/01/02 15:04", newsList[j].PubDate)
		if err1 != nil || err2 != nil {
			return false
		}
		return t1.Before(t2)
	})
	return newsList, nil
}

// FilterRecentNews 최근 48시간 이내의 뉴스 항목만 필터링합니다.
func FilterRecentNews(newsList []NewsItem) []NewsItem {
	currentTime := time.Now().Unix()
	var filtered []NewsItem
	for _, news := range newsList {
		if currentTime-news.Timestamp <= 48*3600 {
			filtered = append(filtered, news)
		}
	}
	return filtered
}
