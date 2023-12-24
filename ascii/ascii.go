package ascii

import (
	"context"
	"errors"

	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	"github.com/getsentry/sentry-go"
	tb "gopkg.in/telebot.v3"
)

// Dependencies contains dependency injection struct
// to be used for the Ascii package.
type Dependencies struct {
	Bot    *tb.Bot
	Logger *sentry.Client
}

// Ascii simply sends ASCII art message for fun.
func (d *Dependencies) Ascii(ctx context.Context, m *tb.Message) {
	if m.Payload == "" {
		return
	}

	gen := utils.GenerateAscii(m.Payload)

	_, err := d.Bot.Send(
		m.Chat,
		"<pre>"+gen+"</pre>",
		&tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true},
	)
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
				shared.HandleBotError(ctx, err, d.Bot, m)
				return
			}
		} else {
			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}
	}
}
