package handlers

import (
	"os"
	"strconv"

	"github.com/ztrue/tracerr"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependencies) SetirManual(m *tb.Message) {
	if strconv.Itoa(m.Sender.ID) != os.Getenv("ADMIN_ID") || m.Chat.Type != tb.ChatPrivate {
		return
	}

	home, err := strconv.Atoi(os.Getenv("HOME_GROUP_ID"))
	if err != nil {
		panic(tracerr.Wrap(err))
	}

	if m.IsReply() {
		replyToID, err := strconv.Atoi(m.Payload)
		if err != nil {
			panic(tracerr.Wrap(err))
		}
		_, err = d.Bot.Send(tb.ChatID(home), m.ReplyTo.Text, &tb.SendOptions{
			ParseMode:         tb.ModeHTML,
			AllowWithoutReply: true,
			ReplyTo: &tb.Message{
				ID: replyToID,
				Chat: &tb.Chat{
					ID: int64(home),
				},
			},
		})
		if err != nil {
			_, err = d.Bot.Send(m.Chat, "Failed sending that message: "+err.Error())
			if err != nil {
				panic(tracerr.Wrap(err))
			}
		} else {
			_, err = d.Bot.Send(m.Chat, "Message sent")
			if err != nil {
				panic(tracerr.Wrap(err))
			}
		}
		return
	}

	_, err = d.Bot.Send(tb.ChatID(home), m.Payload, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
	if err != nil {
		_, err = d.Bot.Send(m.Chat, "Failed sending that message: "+err.Error())
		if err != nil {
			panic(tracerr.Wrap(err))
		}
		return
	}

	_, err = d.Bot.Send(m.Chat, "Message sent")
	if err != nil {
		panic(tracerr.Wrap(err))
	}
}
