package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

const oldNewsFile = "data/old_dandok_list.json"

func loadOldNews() []NewsItem {
	var oldNews []NewsItem
	if _, err := os.Stat(oldNewsFile); err == nil {
		data, err := ioutil.ReadFile(oldNewsFile)
		if err == nil {
			json.Unmarshal(data, &oldNews)
		}
	}
	return oldNews
}

func saveNews(news []NewsItem) {
	data, err := json.Marshal(news)
	if err != nil {
		log.Printf("뉴스 목록 저장 중 오류: %v", err)
		return
	}
	ioutil.WriteFile(oldNewsFile, data, 0644)
}

func filterNewNews(oldNews, fetchedNews []NewsItem) []NewsItem {
	oldTitles := make(map[string]bool)
	for _, news := range oldNews {
		oldTitles[news.Title] = true
	}
	var newNews []NewsItem
	for _, news := range fetchedNews {
		if !oldTitles[news.Title] {
			newNews = append(newNews, news)
		}
	}
	return newNews
}

func main() {
	//logFile, err := os.OpenFile("dandok.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatalf("로그 파일 열기 실패: %v", err)
	//}
	//log.SetOutput(logFile)
	//defer logFile.Close()
	if err := godotenv.Load(); err != nil {
		log.Println("경고: .env 파일을 로드하지 못했습니다. 환경변수는 시스템 설정을 따릅니다.")
	}
	clientID := os.Getenv("NAVER_CLIENT_ID")
	if clientID == "" {
		log.Fatal("NAVER_CLIENT_ID 환경변수가 설정되어 있지 않습니다.")
	}
	clientSecret := os.Getenv("NAVER_CLIENT_SECRET")
	if clientSecret == "" {
		log.Fatal("NAVER_CLIENT_SECRET 환경변수가 설정되어 있지 않습니다.")
	}
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN 환경변수가 설정되어 있지 않습니다.")
	}
	v := viper.New()
	v.SetConfigType("yaml")
	configFile, err := os.Open("config.yaml")
	if err != nil {
		log.Fatalf("설정 파일 열기 실패: %v", err)
	}
	err = v.ReadConfig(configFile)
	if err != nil {
		log.Fatalf("설정 파일 읽기 실패: %v", err)
	}
	var config Config
	err = v.Unmarshal(&config, func(config *mapstructure.DecoderConfig) {
		config.ErrorUnused = true
		config.ErrorUnset = true
	})
	if err != nil {
		log.Fatalf("설정 파일 파싱 실패: %v", err)
	}

	// 텔레그램 Notifier 초기화
	notifier, err := NewNotifier(botToken, config.Telegram.ChatIDs, 1)
	if err != nil {
		log.Fatalf("텔레그램 봇 초기화 오류: %v", err)
	}
	log.Printf("Authorized on account %s", notifier.Bot.Self.UserName)

	// 시작 시 검색어 목록 전송
	for _, chatIDStr := range notifier.ChatIDList {
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			log.Printf("chatID 변환 오류: %v", err)
			continue
		}
		startMsg := tgbotapi.NewMessage(chatID, "미니미 검색어 목록입니다:")
		startMsg.ParseMode = "HTML"
		notifier.Bot.Send(startMsg)
		for _, query := range config.News.QueryList {
			qMsg := tgbotapi.NewMessage(chatID, query)
			qMsg.ParseMode = "HTML"
			notifier.Bot.Send(qMsg)
		}
	}

	for {
		// 기존 뉴스 불러오기 및 최근 뉴스만 필터링
		oldNews := FilterRecentNews(loadOldNews())

		// 뉴스 데이터 가져오기
		fetchedNews, err := GetNewsList(clientID, clientSecret, config.News.QueryList, config.News.TimeWindowHours)
		if err != nil {
			errMsg := fmt.Sprintf("뉴스 검색 중 예외 발생: %v", err)
			log.Println(errMsg)
			notifier.SendMessage(errMsg)
		}

		// 새로운 뉴스 항목 선별
		newNews := filterNewNews(oldNews, fetchedNews)
		log.Printf("새로운 뉴스 %d개 검색", len(newNews))

		// 새로운 뉴스가 있을 경우 텔레그램 메시지 전송
		for _, news := range newNews {
			text := fmt.Sprintf("%s\n%s\n\n%s\n\n<a href=\"%s\">기사 링크</a>",
				news.PubDate, news.Title, news.Description, news.Link)
			notifier.SendMessage(text)
		}
		// 기존 뉴스 목록 갱신 후 저장
		combinedNews := append(oldNews, newNews...)
		saveNews(combinedNews)
		log.Printf("마지막 갱신: %d", time.Now().Unix())
		// 2분 대기
		time.Sleep(time.Duration(config.News.PullIntervalSeconds) * time.Second)
	}
}
