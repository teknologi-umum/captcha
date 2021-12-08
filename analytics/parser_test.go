package analytics_test

import (
	"teknologi-umum-bot/analytics"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestParseToUser(t *testing.T) {
	message := &tb.Message{
		Chat: &tb.Chat{
			ID:   123456789,
			Type: tb.ChatGroup,
		},
		Sender: &tb.User{
			ID:        1,
			FirstName: "Reinaldy",
			LastName:  "Reinaldy",
			Username:  "reinaldy",
		},
	}

	userMap := analytics.ParseToUser(message)
	if userMap.UserID != 1 {
		t.Errorf("UserID should be 1, got: %d", userMap.UserID)
	}
	if userMap.DisplayName != "Reinaldy Reinaldy" {
		t.Errorf("DisplayName should be Reinaldy Reinaldy, got: %s", userMap.DisplayName)
	}
	if userMap.Username != "reinaldy" {
		t.Errorf("Username should be reinaldy, got: %s", userMap.Username)
	}
	if userMap.GroupID != 123456789 {
		t.Errorf("GroupID should be 123456789, got: %d", userMap.GroupID)
	}
}
