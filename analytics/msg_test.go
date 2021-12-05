package analytics_test

import (
	"teknologi-umum-bot/analytics"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestNewMsg(t *testing.T) {
	t.Cleanup(Cleanup)

	m := &tb.Message{
		Sender: &tb.User{
			ID:        123456789,
			FirstName: "Reinaldy",
			LastName:  "Reinaldy",
			Username:  "reinaldy",
		},
	}

	d := &analytics.Dependency{
		DB:     db,
		Memory: memory,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()

	d.NewMsg(m)
}
