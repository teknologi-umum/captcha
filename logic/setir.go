package logic

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependencies) Setir(m *tb.Message) {
	_, err := d.Bot.Send(m.Chat, m.Payload, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
	if err != nil {
		_, err = d.Bot.Send(m.Chat, "Failed sending that message: "+err.Error())
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
		return
	}
}
