package analytics_test

import (
	"testing"

	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

func TestNewMsg(t *testing.T) {
	m := &tb.Message{
		Chat: &tb.Chat{
			ID:   123456789,
			Type: tb.ChatGroup,
		},
		Sender: &tb.User{
			ID:        123456789,
			FirstName: "Reinaldy",
			LastName:  "Reinaldy",
			Username:  "reinaldy",
		},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()

	err := dependency.NewMessage(m)
	if err != nil {
		t.Error(err)
	}
}

func TestNewMsgNotGroup(t *testing.T) {
	m := &tb.Message{
		Chat: &tb.Chat{
			ID:   123456789,
			Type: tb.ChatPrivate,
		},
		Sender: &tb.User{
			ID:        123456789,
			FirstName: "Reinaldy",
			LastName:  "Reinaldy",
			Username:  "reinaldy",
		},
	}

	err := dependency.NewMessage(m)
	if err != nil {
		t.Error(err)
	}
}
