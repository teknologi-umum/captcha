package analytics_test

import (
	"teknologi-umum-bot/analytics"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
)

func TestParseToUser(t *testing.T) {
	user := &tb.User{
		ID:        1,
		FirstName: "Reinaldy",
		LastName:  "Reinaldy",
		Username:  "reinaldy",
	}

	userMap := analytics.ParseToUser(user)
	if userMap.UserID != 1 {
		t.Errorf("UserID should be 1, got: %d", userMap.UserID)
	}
	if userMap.DisplayName != "Reinaldy Reinaldy" {
		t.Errorf("DisplayName should be Reinaldy Reinaldy, got: %s", userMap.DisplayName)
	}
	if userMap.Username != "reinaldy" {
		t.Errorf("Username should be reinaldy, got: %s", userMap.Username)
	}
}
