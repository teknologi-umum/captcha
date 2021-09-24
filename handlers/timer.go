package handlers

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

type CaptchaTimer map[string]CaptchaTimerConfig

type CaptchaTimerConfig struct {
	Timer           *time.Timer
	Sender          *tb.User
	Chat            *tb.Chat
	MessageImage    *tb.Message
	MessageQuestion *tb.Message
}
