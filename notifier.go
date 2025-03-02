package main

import (
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Notifier 는 텔레그램 전송 기능을 담당합니다.
type Notifier struct {
	Bot         *tgbotapi.BotAPI
	ChatIDList  []string
	DelaySecond int
}

// NewNotifier 생성자
func NewNotifier(botToken string, chatIDList []string, delaySecond int) (*Notifier, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}
	return &Notifier{
		Bot:         bot,
		ChatIDList:  chatIDList,
		DelaySecond: delaySecond,
	}, nil
}

// SendMessage 텔레그램으로 메시지를 전송합니다.
func (n *Notifier) SendMessage(message string) {
	for _, chatIDStr := range n.ChatIDList {
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			log.Printf("chatID 변환 오류: %v", err)
			continue
		}
		msg := tgbotapi.NewMessage(chatID, message)
		msg.ParseMode = "HTML"
		_, err = n.Bot.Send(msg)
		if err != nil {
			log.Printf("텔레그램 전송 오류: %v", err)
		}
	}
}
