package analytics_test

import (
	"teknologi-umum-bot/analytics"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestNewUser(t *testing.T) {
	defer Cleanup(DB, Redis)

	user := &tb.User{
		ID:        1,
		Username:  "reinaldy",
		FirstName: "Reinaldy",
		LastName:  "Reinaldy",
	}

	d := &analytics.Dependency{
		DB:    DB,
		Redis: Redis,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()

	d.NewUser(&tb.Message{}, user)
}
