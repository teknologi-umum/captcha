package logic

import (
	"encoding/json"
	"strconv"
	"teknologi-umum-bot/utils"

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

	if m.Sender.IsBot || m.Private() || utils.IsAdmin(admins, m.Sender) {
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

	// OK, they exists in the cache. Now we've got to delete
	// all the message that we've sent before.
	data, err := d.Cache.Get(strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	var captcha Captcha
	err = json.Unmarshal(data, &captcha)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	err = d.removeUserFromCache(strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Delete the question message.
	err = d.Bot.Delete(&tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	})
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Delete user's messages.
	for _, msgID := range captcha.UserMsgs {
		if msgID == "" {
			continue
		}
		err = d.Bot.Delete(&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMsgs {
		if msgID == "" {
			continue
		}
		err = d.Bot.Delete(&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}
}
