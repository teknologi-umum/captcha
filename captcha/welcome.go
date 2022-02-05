package captcha

import (
	"math/rand"
	"strconv"
	"strings"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/utils"
	"time"

	"github.com/getsentry/sentry-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

// currentWelcomeMessages is a collection of welcome messages
// that have a dynamic user value, written as {user}.
//
// This should be sent to the user with a random pick.
var currentWelcomeMessages = [10]string{
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
	`Wah mantap, {user} bisa nyelesain captcha yang aneh itu.

Di grup ini nggak usah sungkan-sungkan kalau mau tanya atau bicara ya. Kita semua manusia bar-bar.
Sehari bisa muncul lebih dari 500 message, jadi jangan lupa matiin notifikasi ya!`,
	`Haii {user}!

Selama di grup ini, jangan sungkan & malu-malu ya. Biarin aja grup ini berisik. Jangan lupa matiin notification juga.

Semoga betah yaa!

Main-main ke website dan Github organization grup ini di https://teknologiumum.com
dan https://github.com/teknologi-umum`,
	`Hi {user}! Keren banget bisa selesain captcha aneh barusan.

Banyak member grup ini yang udah kerja di tempat-tempat keren, dan mereka juga sering ngelirik ke profile
Github. Pastiin profile Github-mu isinya project yang keren juga ya! Nggak usah malu-malu kalau menurutmu masih
biasa aja :D

Jangan lupa matiin notifikasi, grup ini berisik banget, apalagi kalo lagi ngegibah.`,
	`Keren, captchanya selesai! Gimana kabarnya {user}?

Semoga hari ini bukan hari yang buruk ya. Apalagi abis masuk grup ini, rasanya kayak bakal kenal sama temen-temen
baru yang nggak berhenti-berhentinya ngomongin topik yang random banget. Kadang ngomongin pemrograman/teknologi,
kadang juga ngomongin makanan atau ngomongin seberapa ngeselinnya satu provider internet yang berawalan dengan
Indi dan berakhiran dengan home.

Nah, karena kita nggak ada berhenti-berhentinya, jangan lupa matiin notification ya biar nggak berasa artis.

Fun fact: ada orang yang pernah mencoba kayak gitu, jadi notificationnya dibiarin nyala. Dia hanya bertahan 3 jam.
Abis itu memutuskan untuk mematikan notifikasinya.`,
	`Hai {user}!

Nama grup ini Teknologi Umum. Berarti kita membicarakan hal berbau teknologi dan/atau hal umum.

Jadi kalo ada yang lagi ngomongin cara terbaik untuk ngoding di PHP tapi 10 detik kemudian ada yang ngomong,
"ini kenapa angkot depan rumah ngeselin banget sih", itu valid. Nggak ada topik OOT disini.

Kamu nggak perlu pasang profile picture/username. Cukup matiin notification aja, berisik banget disini.

Kita juga punya website dan Github organization. Bisa di cek di: https://teknologiumum.com
dan https://github.com/teknologi-umum`,
}

// deleteMessage creates a timer of one minute to delete a certain message.
func deleteMessage(bot *tb.Bot, message tb.StoredMessage, logger *sentry.Client) {
	c := make(chan struct{}, 1)
	time.AfterFunc(time.Minute*1, func() {
		err := bot.Delete(message)
		if err != nil {
			shared.HandleError(err, logger)
		}
		c <- struct{}{}
	})

	<-c
}

// sendWelcomeMessage literally does what it's written.
func sendWelcomeMessage(bot *tb.Bot, m *tb.Message, logger *sentry.Client) error {
	msg, err := bot.Send(
		m.Chat,
		strings.Replace(
			currentWelcomeMessages[randomNum()],
			"{user}",
			"<a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+
				sanitizeInput(m.Sender.FirstName)+utils.ShouldAddSpace(m.Sender)+sanitizeInput(m.Sender.LastName)+
				"</a>",
			1,
		),
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

	go deleteMessage(
		bot,
		tb.StoredMessage{MessageID: strconv.Itoa(msg.ID), ChatID: m.Chat.ID},
		logger,
	)
	return nil
}

func randomNum() int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(len(currentWelcomeMessages) - 1)
}
