package main

// Config 는 전체 설정을 담는 구조체입니다.
type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	News     NewsConfig     `mapstructure:"news"`
}

// TelegramConfig 는 텔레그램 관련 설정을 담습니다.
type TelegramConfig struct {
	ChatIDs      []string `mapstructure:"chat_ids"`
	DelaySeconds int      `mapstructure:"delay_seconds"`
}

// NewsConfig 는 뉴스 검색 및 저장 관련 설정을 담습니다.
type NewsConfig struct {
	QueryList           []string `mapstructure:"query_list"`
	TimeWindowHours     int      `mapstructure:"time_window_hours"`
	PullIntervalSeconds int      `mapstructure:"pull_interval_seconds"`
}
