package logic

import (
	"encoding/json"
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

	// Delete the question message.
	msgToBeDeleted := tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	}
	err = d.Bot.Delete(&msgToBeDeleted)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMsgs {
		msgToBeDeleted = tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		}
		err = d.Bot.Delete(&msgToBeDeleted)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}

	err = removeUserFromCache(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}
}
