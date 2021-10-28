package handlers

import (
	"strconv"
	"strings"
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Ini ga penting.
type Captcha struct {
	Question   string `json:"question" bson:"question"`
	Answer     string `json:"answer" bson:"answer"`
	ContentURL string `json:"contenturl" bson:"contenturl"`
}

var CaptchaQuestion = `Halo, {user}!

Sebelum lanjut, selesaikan captcha ini dulu ya. Semuanya angka.

<pre>{captcha}</pre>`

func (d *Dependencies) CaptchaUserJoin(m *tb.Message) {
	// randNum generates a random number (4 digit) in string format
	var randNum string = utils.GenerateRandomNumber()
	// captcha generates ascii art from the randNum value
	var captcha string = utils.GenerateAscii(randNum)

	// Replacing the templte from CaptchaQuestion
	question := strings.Replace(
		strings.Replace(CaptchaQuestion, "{captcha}", captcha, 1),
		"{user}",
		"<a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+m.Sender.FirstName+"</a>",
		1,
	)

	// Send the question first.
	msgQuestion, err := d.Bot.Send(m.Sender, question, &tb.SendOptions{ParseMode: tb.ModeHTML, ReplyTo: m})
	if err != nil {
		panic(err)
	}

	// See https://redis.io/commands/hset for the Redis documentation
	// Jadi isi dari key captcha:users ini Hash. Tambahin terus setiap ada user baru.
	//
	// Ini dimasukin Redis cuma buat mantain persistence aja, karena nggak pake SQL database.
	// Ujung-ujungnya value di key captcha:users bakal masuk in-memory data.
	// Pas kita listen the whole chat buat cek user yang ngirim jawaban captcha, kita carinya
	// dari in-memory data, bukan dari Redis.
	//
	// Kalo bener, yang di Redis dan di in-memory di hapus.
	//
	// Buat dapetin semua elemen di hash: https://redis.io/commands/hgetall
	// Buat dapetin length dari suatu hash: https://redis.io/commands/hlen
	// Buat dapetin keys nya doang: https://redis.io/commands/hkeys
	// Buat delete 1 key-value pair dari hash: https://redis.io/commands/hdel
	err = d.Redis.HSet(d.Context, "captcha:users", []string{strconv.Itoa(m.Sender.ID), randNum}).Err()
	if err != nil {
		panic(err)
	}

	// Jadi TODO setelah kode diatas adalah:
	// 1. Masukin datanya ke in-memory (satu buat daftar users, satu buat jawaban), nama key nya terserah.
	// 2. Start the timer
	// Buat stop timer, bisa baca-baca: https://stackoverflow.com/questions/50223771/how-to-stop-a-timer-correctly
	// Iya, kita harus pake select.
	// Documentation Go soal time.NewTimer() https://pkg.go.dev/time#NewTimer

	// Ini kode lama, aku bisa jamin ga works. Tapi cek comment yang aku masukin.
	expiryTimer := time.AfterFunc(5*time.Minute, func() {
		err := d.Cache.Delete(strconv.Itoa(m.Sender.ID) + ":expiry")
		if err != nil {
			panic(err)
		}

		err = d.Cache.Delete(strconv.Itoa(m.Sender.ID) + ":answer")
		if err != nil {
			panic(err)
		}

		_, err = d.Bot.Send(m.Sender,
			"<a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+m.Sender.FirstName+" "+m.Sender.LastName+"</a> didn't solve the captcha. Alright, time to kick them.",
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
		if err != nil {
			panic(err)
		}

		// Ini buat kick orang. Walaupun keywordnya ban,
		// selama RestrictedUntil ini nggak minus, dia hanya ke kick.
		err = d.Bot.Ban(m.Chat, &tb.ChatMember{
			RestrictedUntil: time.Now().Unix() + int64(1*time.Hour),
			User:            m.Sender,
		}, true)
		if err != nil {
			panic(err)
		}

		// Ini buat delete message. Dia terima *tb.Message
		err = d.Bot.Delete(msgQuestion)
		if err != nil {
			panic(err)
		}
	})

	// Ini ga works.
	d.Timers <- CaptchaTimer{
		strconv.Itoa(m.Sender.ID): CaptchaTimerConfig{
			Timer:           expiryTimer,
			Sender:          m.Sender,
			Chat:            m.Chat,
			MessageQuestion: msgQuestion,
		},
	}
}

// Aku ga inget ini apa. Rombak aja.
func (d *Dependencies) CaptchaUserMessage(m *tb.Message) {
	// Listen to user message, check if current message matches
	_, err := d.Cache.Get(strconv.Itoa(m.Sender.ID) + ":expiry")
	if err != nil {
		return
	}

}
