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
var currentWelcomeMessages = [8]string{
	"Halo, {user} \n" +

		"Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya. \n" +

		"Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Mau ngomongin apa aja juga boleh kok. \n" +

		"Ngga perlu pasang profile picture dan username kayak grup-grup sebelah.",
	"Hai {user}! \n" +

		"Selamat datang di grup Teknologi Umum. Disini kita berisik banget, jadi langsung matiin notificationnya ya. \n" +

		"Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Jangan bikin kita diciduk tukang bakso bawa HT. \n" +

		"Kalo mau OOT juga ga perlu izin, toh ini grup buat OOT.",
	"Welcome {user}! \n" +

		"Saya ngga tau mau ngomong apa lagi selain jangan lupa matiin notification, grup ini berisik banget.",
	"Haloo {user}! \n" +

		"Selamat datang di grup Teknologi Umum, yuk langsung matiin notification biar hidup kamu ngga sengsara. \n" +

		"Tapi grup ini akur kok, sejauh ini ngga pernah ada drama. Semoga betah ya!",
	"Hai, {user}! \n" +

		"Selamat datang di grup Teknologi Umum! \n" +

		"Coba ketik (dan kirim) /quiz deh, nanti grup ini tiba-tiba hidup. \n" +

		"Oh iya, grup ini ngga ada aturan. Tapi jangan sampe bikin kita diciduk tukang bakso bawa HT.",
	"Haii {user}! \n" +

		"Selama di grup ini, jangan sungkan & malu-malu ya. Biarin aja grup ini berisik. Jangan lupa matiin notification juga. \n" +

		"Semoga betah yaa!\n " +

		"Main-main ke website dan Github organization grup ini di https://teknologiumum.com \n" +
		"dan https://github.com/teknologi-umum",
	"Hi {user}! Keren banget bisa selesain captcha aneh barusan \n." +

		"Banyak member grup ini yang udah kerja di tempat-tempat keren, dan mereka juga sering ngelirik ke profile \n" +
		"Github. Pastiin profile Github-mu isinya project yang keren juga ya! Nggak usah malu-malu kalau menurutmu masih \n" +
		"biasa aja :D \n" +

		"Jangan lupa matiin notifikasi, grup ini berisik banget, apalagi kalo lagi ngegibah.",
	"您好 {user}! \n" +
		"欢迎您在 Teknologi Umum, 我们每天都很嘈杂, 请把你的筒子声音关掉, \n" +
		"您要问什么, 请问吧. 希望您很高兴在这里"}

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
