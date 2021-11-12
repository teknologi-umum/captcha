package logic

import (
	"strconv"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependencies) CaptchaUserLeave(m *tb.Message) {
	// Check if the user is an admin or bot first.
	// If they are, return.
	// If they're not, continue execute the captcha.
	admins, err := d.Bot.AdminsOf(m.Chat)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	if m.Sender.IsBot || m.Private() || isAdmin(admins, m.Sender) {
		return
	}

	// We need to check if the user is in the captcha:users cache
	// or not.
	check, err := userExists(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	if !check {
		return
	}

	// If the user is in the cache, we need to remove their data.
	err = d.Cache.Delete(strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}
}
