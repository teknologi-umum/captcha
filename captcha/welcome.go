package captcha

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"teknologi-umum-bot/utils"

	tb "gopkg.in/telebot.v3"
)

// currentWelcomeMessages is a collection of welcome messages
// that have a dynamic user value, written as {user}.
//
// This should be sent to the user with a random pick.
var currentWelcomeMessages = [11]string{
	"Halo, {user}\n\n" +
		"Selamat datang di grup {groupname}. Disini kita berisik banget, jadi langsung matiin notificationnya ya. " +
		"Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Mau ngomongin apa aja juga boleh kok. " +
		"Ngga perlu pasang profile picture dan username kayak grup-grup sebelah.",
	"Hai {user}! \n\n" +
		"Selamat datang di grup {groupname}. Disini kita berisik banget, jadi langsung matiin notificationnya ya. " +
		"Disini sebenernya nggak ada aturan, tapi ya wajar-wajar aja lah. Jangan bikin kita diciduk tukang bakso bawa HT. " +
		"Kalo mau OOT juga ga perlu izin, toh ini grup buat OOT.",
	"Welcome {user}!\n\n" +
		"Saya ngga tau mau ngomong apa lagi selain jangan lupa matiin notification, grup ini berisik banget.",
	"Haloo {user}!\n\n" +
		"Selamat datang di grup {groupname}, yuk langsung matiin notification biar hidup kamu ngga sengsara. " +
		"Tapi grup ini akur kok, sejauh ini ngga pernah ada drama. Semoga betah ya!",
	"Hai, {user}!\n\n" +
		"Selamat datang di grup {groupname}!\n\n" +
		"Coba ketik (dan kirim) /joke@TeknologiUmumBot deh, nanti grup ini tiba-tiba hidup.\n\n" +
		"Oh iya, grup ini ngga ada aturan. Tapi jangan sampe bikin kita diciduk tukang bakso bawa HT.",
	"Haii {user}!\n\n" +
		"Selama di grup ini, jangan sungkan & malu-malu ya. Biarin aja grup ini berisik. Jangan lupa matiin notification juga. " +
		"Semoga betah yaa!\n\n" +
		"Main-main ke website dan Github organization grup ini di https://teknologiumum.com " +
		"dan https://github.com/teknologi-umum",
	"Hi {user}! Keren banget bisa selesain captcha aneh barusan.\n\n" +
		"Banyak member grup ini yang udah kerja di tempat-tempat keren, dan mereka juga sering ngelirik ke profile " +
		"Github. Pastiin profile Github-mu isinya project yang keren juga ya! Nggak usah malu-malu kalau menurutmu masih " +
		"biasa aja :D\n\n" +
		"Jangan lupa matiin notifikasi, grup ini berisik banget, apalagi kalo lagi ngegibah.",
	"您好 {user}! \n\n" +
		"欢迎您在 {groupname}, 我们每天都很嘈杂, 请把你的筒子声音关掉, " +
		"您要问什么, 请问吧. 希望您很高兴在这里",
	"こんにちは！\n\n" +
		"Tenang, ini bukan grup jejepangan kok. Tapi kalau mau ngomongin topik apapun, dari Jepang, Korea, sampe negara yang " +
		"nggak banyak orang tau boleh banget. Ini grup untuk bahas hal-hal yang dianggap OOT di grup lain.\n\n" +
		"Semoga hidupmu di grup ini menyenangkan!",
	"Halo, {user}!\n\n" +
		"Kalau kamu lagi senggang, dan kebetulan kamu adalah programmer, coba cek pinned message dan coba kerjakan " +
		"kuis-kuis yang ada. Kebanyakan kuisnya anonim kok, jadi kamu nggak perlu takut salah-benar.\n\n" +
		"Semoga harimu menyenangkan!",
	"Hai {user}!\n\n" +
		"Yeay! Kamu bisa menyelesaikan captcha super aneh itu. Selamat datang di {groupname} ya.\n\n" +
		"Cerita sedikit soal diri kamu dong, sekarang kerjaannya apa dan suka melakukan apa pas senggang?",
}

var regularWelcomeMessage = "Halo, {user}!\n\n" +
	"Selamat datang di {groupname}. Jangan lupa untuk baca pinned message, ya. Semoga hari mu menyenangkan."

// sendWelcomeMessage literally does what it's written.
func (d *Dependencies) sendWelcomeMessage(ctx context.Context, m *tb.Message) error {
	var msgToSend string = regularWelcomeMessage

	if strconv.FormatInt(m.Chat.ID, 10) == d.TeknumID {
		msgToSend = currentWelcomeMessages[randomNum()]
	}

	for {
		msg, err := d.Bot.Send(
			m.Chat,
			strings.NewReplacer(
				"{user}",
				"<a href=\"tg://user?id="+strconv.FormatInt(m.Sender.ID, 10)+"\">"+
					sanitizeInput(m.Sender.FirstName)+utils.ShouldAddSpace(m.Sender)+sanitizeInput(m.Sender.LastName)+
					"</a>",
				"{groupname}",
				sanitizeInput(m.Chat.Title),
			).Replace(msgToSend),
			&tb.SendOptions{
				ReplyTo:               m,
				ParseMode:             tb.ModeHTML,
				DisableWebPagePreview: true,
				DisableNotification:   false,
				AllowWithoutReply:     true,
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "retry after") {
				// Acquire the retry number
				retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
				if err != nil {
					// If there's an error, we'll just retry after 15 second
					retry = 15
				}

				// Let's wait a bit and retry
				time.Sleep(time.Second * time.Duration(retry))
				continue
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				time.Sleep(time.Second * 10)
				continue
			}
			return fmt.Errorf("failed to send welcome message: %w", err)
		}

		go d.deleteMessage(
			ctx,
			&tb.StoredMessage{MessageID: strconv.Itoa(msg.ID), ChatID: m.Chat.ID},
		)

		break
	}

	return nil
}

func randomNum() int {
	return rand.Intn(len(currentWelcomeMessages))
}
