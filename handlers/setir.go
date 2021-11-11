package handlers

import (
	"os"
	"strconv"
	"strings"
	"teknologi-umum-bot/utils"

	"github.com/aldy505/decrr"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (d *Dependencies) SetirManual(m *tb.Message) {
	admin := strings.Split(os.Getenv("ADMIN_ID"), ",")
	if !utils.IsIn(admin, strconv.Itoa(m.Sender.ID)) || m.Chat.Type != tb.ChatPrivate {
		return
	}

	home, err := strconv.Atoi(os.Getenv("HOME_GROUP_ID"))
	if err != nil {
		panic(decrr.Wrap(err))
	}

	if m.IsReply() {
		var replyToID int

		if strings.HasPrefix(m.Payload, "https://t.me/") {
			replyToID, err = strconv.Atoi(strings.Split(m.Payload, "/")[4])
			if err != nil {
				panic(decrr.Wrap(err))
			}
		} else {
			replyToID, err = strconv.Atoi(m.Payload)
			if err != nil {
				panic(decrr.Wrap(err))
			}
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
				panic(decrr.Wrap(err))
			}
		} else {
			_, err = d.Bot.Send(m.Chat, "Message sent")
			if err != nil {
				panic(decrr.Wrap(err))
			}
		}
		return
	}

	if strings.HasPrefix(m.Payload, "https://") {
		var toBeSent interface{}
		if strings.HasSuffix(m.Payload, ".jpg") || strings.HasSuffix(m.Payload, ".png") || strings.HasSuffix(m.Payload, ".jpeg") {
			toBeSent = &tb.Photo{File: tb.FromURL(m.Payload)}
		} else if strings.HasSuffix(m.Payload, ".gif") {
			toBeSent = &tb.Animation{File: tb.FromURL(m.Payload)}
		} else {
			return
		}

		_, err = d.Bot.Send(tb.ChatID(home), toBeSent, &tb.SendOptions{AllowWithoutReply: true})
		if err != nil {
			_, err = d.Bot.Send(m.Chat, "Failed sending that photo: "+err.Error())
			if err != nil {
				panic(decrr.Wrap(err))
			}
			return
		}

		_, err = d.Bot.Send(m.Chat, "Photo sent")
		if err != nil {
			panic(decrr.Wrap(err))
		}
		return

	}

	_, err = d.Bot.Send(tb.ChatID(home), m.Payload, &tb.SendOptions{ParseMode: tb.ModeHTML, AllowWithoutReply: true})
	if err != nil {
		_, err = d.Bot.Send(m.Chat, "Failed sending that message: "+err.Error())
		if err != nil {
			panic(decrr.Wrap(err))
		}
		return
	}

	_, err = d.Bot.Send(m.Chat, "Message sent")
	if err != nil {
		panic(decrr.Wrap(err))
	}
}
