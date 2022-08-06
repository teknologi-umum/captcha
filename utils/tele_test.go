package utils_test

import (
	"teknologi-umum-bot/utils"
	"testing"

	tb "gopkg.in/telebot.v3"
)

func TestShouldAddSpace(t *testing.T) {
	s := utils.ShouldAddSpace(&tb.User{LastName: ""})
	if s != "" {
		t.Error("ShouldAddSpace should return empty string")
	}

	s = utils.ShouldAddSpace(&tb.User{LastName: "Reinaldy"})
	if s != " " {
		t.Error("ShouldAddSpace should return a space")
	}
}

func TestIsAdmin(t *testing.T) {
	var admins []tb.ChatMember
	admins = append(admins, tb.ChatMember{User: &tb.User{ID: 1}})
	admins = append(admins, tb.ChatMember{User: &tb.User{ID: 2}})
	admins = append(admins, tb.ChatMember{User: &tb.User{ID: 3}})

	if utils.IsAdmin(admins, &tb.User{ID: 1}) == false {
		t.Error("IsAdmin should return true")
	}

	if utils.IsAdmin(admins, &tb.User{ID: 4}) == true {
		t.Error("IsAdmin should return false")
	}
}
