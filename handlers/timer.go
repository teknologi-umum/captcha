package handlers

import (
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Ga penting, udah ga relevan dengan logic yang baru.
// Kalo mau hapus, ya hapus aja gaperlu izin.
type CaptchaTimer map[string]CaptchaTimerConfig

type CaptchaTimerConfig struct {
	Timer           *time.Timer
	Sender          *tb.User
	Chat            *tb.Chat
	MessageImage    *tb.Message
	MessageQuestion *tb.Message
}
