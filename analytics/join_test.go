package analytics_test

import (
	"testing"

	tb "gopkg.in/telebot.v3"
)

func TestNewUser(t *testing.T) {
	user := &tb.User{
		ID:        1,
		Username:  "reinaldy",
		FirstName: "Reinaldy",
		LastName:  "Reinaldy",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()

	dependency.NewUser(&tb.Message{Chat: &tb.Chat{ID: 10}}, user)
}
