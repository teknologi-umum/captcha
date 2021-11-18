package logic

import (
	"strconv"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependencies) DoOrDel(m *tb.Message) {
	err := d.Bot.Delete(&tb.StoredMessage{
		ChatID:    m.ReplyTo.Chat.ID,
		MessageID: strconv.Itoa(m.ReplyTo.ID),
	})
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}
}
