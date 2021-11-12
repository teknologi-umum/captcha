package logic

import (
	"math/rand"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

var currentWelcomeMessages [5]string = [5]string{
	`Halo, {user}!

Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya.

Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Mau ngomongin apa aja juga boleh kok.

Ngga perlu pasang profile picture dan username kayak grup-grup sebelah.`,
	`Hai {user}!

Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya.

Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Jangan bikin kita diciduk tukang bakso bawa HT.

Kalo mau OOT juga ga perlu izin, toh ini grup buat OOT.`,
	`Welcome {user}!

Saya ngga tau mau ngomong apa lagi selain jangan lupa matiin notification, grup ini berisik banget.`,
	`Haloo {user}!

Selamat datang di grup Teknologi Umum, yuk langsung matiin notification biar hidup kamu ngga sengsara.

Tapi grup ini akur kok, sejauh ini ngga pernah ada drama. Semoga betah ya!`,
	`Hai, {user}!

Selamat datang di grup Teknologi Umum!

Coba ketik (dan kirim) /quiz deh, nanti grup ini tiba-tiba hidup.

Oh iya, grup ini ngga ada aturan. Tapi jangan sampe bikin kita diciduk tukang bakso bawa HT.`,
}

func deleteMessage(bot *tb.Bot, message tb.Editable) {
	c := make(chan struct{}, 1)
	time.AfterFunc(time.Minute*1, func() {
		bot.Delete(message)
		c <- struct{}{}
	})

	<-c
}

func sendWelcomeMessage(bot *tb.Bot, m *tb.Message, logger *sentry.Client) error {
	msg, err := bot.Send(
		m.Chat,
		strings.Replace(currentWelcomeMessages[randomNum()], "{user}", m.UserJoined.FirstName+" "+m.UserJoined.LastName, 1),
		&tb.SendOptions{
			ReplyTo:               m,
			ParseMode:             tb.ModeHTML,
			DisableWebPagePreview: true,
			DisableNotification:   false,
			AllowWithoutReply:     true,
		},
	)
	if err != nil {
		return err
	}

	go deleteMessage(bot, msg)
	return nil
}

func randomNum() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(4)
}
