package handlers

import (
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

var currentWelcomeMessage string = `Halo, {user}!

Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya.

Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Mau ngomongin apa aja juga boleh kok.

Ngga perlu pasang profile picture dan username kayak grup-grup sebelah.

Repository public kita lagi open buat <a href="https://hacktoberfest.digitalocean.com/">Hacktoberfest</a> juga nih. Daftar repository nya bisa dilihat di message ini: https://t.me/teknologi_umum/93094`

func deleteMessage(bot *tb.Bot, message tb.Editable) {
	c := make(chan struct{}, 1)
	time.AfterFunc(time.Minute*1, func() {
		bot.Delete(message)
		c <- struct{}{}
	})

	<-c
}

func (d *Dependencies) WelcomeMessage(m *tb.Message) {
	msg, err := d.Bot.Send(
		m.Chat,
		strings.Replace(currentWelcomeMessage, "{user}", m.Sender.FirstName+" "+m.Sender.LastName, 1),
		&tb.SendOptions{
			ReplyTo:               m,
			ParseMode:             tb.ModeHTML,
			DisableWebPagePreview: true,
			DisableNotification:   false,
			AllowWithoutReply:     true,
		},
	)
	if err != nil {
		panic(err)
	}

	go deleteMessage(d.Bot, msg)
}
