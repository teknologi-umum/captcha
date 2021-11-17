package utils

import tb "gopkg.in/tucnak/telebot.v2"

func ShouldAddSpace(m *tb.User) string {
	if m.LastName != "" {
		return " "
	}

	return ""
}

// Check whether or not a user is in the admin list
func IsAdmin(admins []tb.ChatMember, user *tb.User) bool {
	for _, v := range admins {
		if v.User.ID == user.ID {
			return true
		}
	}
	return false
}
