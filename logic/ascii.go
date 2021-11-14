package logic

import (
	"errors"
	"teknologi-umum-bot/utils"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Send ASCII art message for fun.
func (d *Dependencies) Ascii(m *tb.Message) {
	if m.Payload == "" {
		return
	}

	gen := utils.GenerateAscii(m.Payload)

	_, err := d.Bot.Send(m.Chat, "<pre>"+gen+"</pre>", &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
	if err != nil {
		if errors.Is(err, tb.ErrEmptyMessage) {
			_, err := d.Bot.Send(
				m.Chat,
				"That text is not supported yet",
				&tb.SendOptions{
					ParseMode:         tb.ModeHTML,
					AllowWithoutReply: true,
					ReplyTo:           m,
				},
			)
			if err != nil {
				handleError(err, d.Logger, d.Bot, m)
				return
			}
		} else {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}
}
