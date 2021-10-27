package handlers

import (
	"math/rand"
	"strings"
	"time"

	"github.com/aldy505/decrr"
	tb "gopkg.in/tucnak/telebot.v2"
)

var currentWelcomeMessages [5]string = [5]string{
	`Halo, {user}!

Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya.

Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Mau ngomongin apa aja juga boleh kok.

Ngga perlu pasang profile picture dan username kayak grup-grup sebelah.

Repository public kita lagi open buat <a href="https://hacktoberfest.digitalocean.com/">Hacktoberfest</a> juga nih. Daftar repository nya bisa dilihat di message ini: https://t.me/teknologi_umum/93094`,
	`Eh haloo {user}!

Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya.

Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Jangan bikin kita diciduk tukang bakso bawa HT.

Kalo mau OOT juga ga perlu izin, toh ini grup buat OOT.

Repository public kita lagi open buat <a href="https://hacktoberfest.digitalocean.com/">Hacktoberfest</a> juga nih. Daftar repository nya bisa dilihat di message ini: https://t.me/teknologi_umum/93094`,
	`Welcome {user}!

Saya ngga tau mau ngomong apa lagi selain jangan lupa matiin notification, grup ini berisik banget.

Harusnya saya promosi soal Hacktoberfest yang lagi diselenggarain di beberapa repository grup ini, cuma ngga ah, malas hahaha.`,
	`Haloo {user}!

Selamat datang di grup Teknologi Umum, yuk langsung matiin notification biar hidup kamu ngga sengsara.

Tapi grup ini akur kok, sejauh ini ngga pernah ada drama. Semoga betah ya!

Kalo kamu mau ngumpulin point <a href="https://hacktoberfest.digitalocean.com/">Hacktoberfest</a> yang cuma modal nulis artikel doang, bisa ambil issue ini: https://github.com/teknologi-umum/blog/issues/39`,
	`Hai, {user}!

Selamat datang di grup Teknologi Umum!

Coba ketik (dan kirim) /quiz deh, nanti grup ini tiba-tiba hidup.

Oh iya, grup ini ngga ada aturan. Tapi jangan sampe bikin kita diciduk tukang bakso bawa HT.

Repository public kita lagi open buat <a href="https://hacktoberfest.digitalocean.com/">Hacktoberfest</a> juga nih. Daftar repository nya bisa dilihat di message ini: https://t.me/teknologi_umum/93094`,
}

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
		panic(decrr.Wrap(err))
	}

	go deleteMessage(d.Bot, msg)
}

func randomNum() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(4)
}
