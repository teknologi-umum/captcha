package analytics_test

import (
	"teknologi-umum-bot/analytics"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestNewMsg(t *testing.T) {
	t.Cleanup(Cleanup)

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

	d := &analytics.Dependency{
		DB:     db,
		Memory: memory,
		TeknumID: "123456789",
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: %v", r)
		}
	}()

	err := d.NewMessage(m)
	if err != nil {
		t.Error(err)
	}
}

func TestNewMsgNotGroup(t *testing.T) {
	t.Cleanup(Cleanup)

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

	d := &analytics.Dependency{
		DB:     db,
		Memory: memory,
	}

	err := d.NewMessage(m)
	if err != nil {
		t.Error(err)
	}
}
