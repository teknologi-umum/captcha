package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"teknologi-umum-bot/utils"
	"time"

	"github.com/allegro/bigcache/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Ini penting.
type Captcha struct {
	Answer         string    `json:"answer"`
	Expiry         time.Time `json:"expiry"`
	ChatID         int64     `json:"chat_id"`
	QuestionID     string    `json:"question_id"`
	AdditionalMsgs []string  `json:"additional_msgs"`
}

const (
	BAN_DURATION    = 1 * time.Minute
	CAPTCHA_TIMEOUT = 1 * time.Minute
)

var CaptchaQuestion = `Halo, {user}!

Sebelum lanjut, selesaikan captcha ini dulu ya. Semuanya angka. Kamu punya waktu 1 menit dari sekarang!

<pre>{captcha}</pre>`

func (d *Dependencies) CaptchaUserJoin(m *tb.Message) {
	// Check if the user is an admin or bot first.
	// If they are, return.
	// If they're not, continue execute the captcha.

	admins, err := d.Bot.AdminsOf(m.Chat)
	if err != nil {
		panic(err)
	}

	if m.Sender.IsBot || m.Private() || isAdmin(admins, m.Sender) {
		_, err = d.Bot.Send(m.Chat, "Kamu admin, nggak bisa")
		if err != nil {
			panic(err)
		}
		return
	}

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
	msgQuestion, err := d.Bot.Send(m.Chat, question, &tb.SendOptions{ParseMode: tb.ModeHTML, ReplyTo: m})
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
	// err = d.Redis.HSet(d.Context, "captcha:users", []string{strconv.Itoa(m.Sender.ID), randNum}).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// Jadi TODO setelah kode diatas adalah:
	// 1. Masukin datanya ke in-memory (satu buat daftar users, satu buat jawaban), nama key nya terserah.
	// 2. Start the timer
	// Buat stop timer, bisa baca-baca: https://stackoverflow.com/questions/50223771/how-to-stop-a-timer-correctly
	// Iya, kita harus pake select.
	// Documentation Go soal time.NewTimer() https://pkg.go.dev/time#NewTimer

	captchaData, err := json.Marshal(Captcha{
		Expiry:     time.Now().Add(CAPTCHA_TIMEOUT),
		ChatID:     m.Chat.ID,
		Answer:     randNum,
		QuestionID: strconv.Itoa(msgQuestion.ID),
	})
	if err != nil {
		panic(err)
	}

	log.Println("sender id:", m.Sender.ID)
	err = d.Cache.Set(strconv.Itoa(m.Sender.ID), captchaData)
	if err != nil {
		panic(err)
	}

	err = d.Cache.Append("captcha:users", []byte(strconv.Itoa(m.Sender.ID)+","))
	if err != nil {
		panic(err)
	}

	users, err := d.Cache.Get("captcha:users")
	if err != nil {
		panic(err)
	}

	log.Println(string(users))

	cond := sync.NewCond(&sync.Mutex{})
	done := make(chan bool, 1)
	go waitOrDelete(d.Cache, d.Bot, m, msgQuestion, cond, &done)
}

// Check whether or not a user is in the admin list
func isAdmin(admins []tb.ChatMember, user *tb.User) bool {
	for _, v := range admins {
		if v.User.ID == user.ID {
			return true
		}
	}
	return false
}

func waitOrDelete(cache *bigcache.BigCache, bot *tb.Bot, msgUser *tb.Message, msgQst *tb.Message, cond *sync.Cond, done *chan bool) {
	t := time.NewTimer(CAPTCHA_TIMEOUT)
	log.Println("timer started")
	go func() {
		cond.L.Lock()
		for _, ok := <-t.C; ok; {
			log.Println("entering the for loop")
			log.Println("sender id:", strconv.Itoa(msgUser.Sender.ID))
			check := cacheExists(cache, strconv.Itoa(msgUser.Sender.ID))
			log.Println("check:", check)
			if check {
				// Fetch the captcha data first
				var captcha Captcha
				user, err := cache.Get(strconv.Itoa(msgUser.Sender.ID))
				if err != nil {
					panic(err)
				}

				err = json.Unmarshal(user, &captcha)
				if err != nil {
					panic(err)
				}

				kickMsg, err := bot.Send(msgUser.Chat,
					"<a href=\"tg://user?id="+strconv.Itoa(msgUser.Sender.ID)+"\">"+msgUser.Sender.FirstName+" "+msgUser.Sender.LastName+"</a> didn't solve the captcha. Alright, time to kick them.",
					&tb.SendOptions{
						ParseMode: tb.ModeHTML,
					})
				if err != nil {
					panic(err)
				}

				// Ini buat kick orang. Walaupun keywordnya ban,
				// selama RestrictedUntil ini nggak minus, dia hanya ke kick.
				err = bot.Ban(msgUser.Chat, &tb.ChatMember{
					RestrictedUntil: time.Now().Unix() + int64(BAN_DURATION),
					User:            msgUser.Sender,
				}, true)
				if err != nil {
					panic(err)
				}

				// Delete all the message that we've sent unless the last one.
				msgToBeDeleted := tb.StoredMessage{
					ChatID:    msgUser.Chat.ID,
					MessageID: captcha.QuestionID,
				}
				err = bot.Delete(&msgToBeDeleted)
				if err != nil {
					panic(err)
				}

				for _, msgID := range captcha.AdditionalMsgs {
					msgToBeDeleted = tb.StoredMessage{
						ChatID:    msgUser.Chat.ID,
						MessageID: msgID,
					}
					err = bot.Delete(&msgToBeDeleted)
					if err != nil {
						panic(err)
					}
				}

				go deleteMessage(bot, kickMsg)

				err = cache.Delete(strconv.Itoa(msgUser.Sender.ID))
				if err != nil {
					panic(err)
				}
				*done <- false
				return
			}
			*done <- true
			break
		}
		cond.Broadcast()
		cond.L.Unlock()
	}()
	<-*done
}

func cacheExists(cache *bigcache.BigCache, key string) bool {
	_, err := cache.Get(key)
	return !errors.Is(err, bigcache.ErrEntryNotFound)
}

func (d *Dependencies) WaitForAnswer(m *tb.Message) {
	// Check if the message author is in the captcha:users list or not
	// If not, return
	// If yes, check if the answer is correct or not
	check, err := userExists(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		panic(err)
	}

	if !check {
		return
	}

	// Check if the answer is correct or not
	// If not, ask them to give the correct answer and time remaining
	// If yes, delete the message and remove the user from the captcha:users list

	// Get the answer from the cache
	data, err := d.Cache.Get(strconv.Itoa(m.Sender.ID))
	if err != nil {
		panic(err)
	}

	var captcha Captcha
	err = json.Unmarshal(data, &captcha)
	if err != nil {
		panic(err)
	}

	// Check if the answer is correct or not
	if m.Text != captcha.Answer {
		remainingTime := time.Until(captcha.Expiry)
		wrongMsg, err := d.Bot.Send(
			m.Chat,
			"Wrong answer, please try again. You have "+strconv.Itoa(int(remainingTime.Seconds()))+" more second to solve the captcha.",
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyTo:   m,
			},
		)
		if err != nil {
			panic(err)
		}

		captcha.AdditionalMsgs = append(captcha.AdditionalMsgs, strconv.Itoa(wrongMsg.ID))

		// Update the cache
		data, err = json.Marshal(captcha)
		if err != nil {
			panic(err)
		}

		err = d.Cache.Set(strconv.Itoa(m.Sender.ID), data)
		if err != nil {
			panic(err)
		}

		return
	}

	// Congratulate the user, delete the message, then delete user from captcha:users
	_, err = d.Bot.Send(
		m.Chat,
		"<a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+m.Sender.FirstName+" "+m.Sender.LastName+"</a> solved the captcha!",
		&tb.SendOptions{
			ParseMode: tb.ModeHTML,
			ReplyTo:   m,
		},
	)
	if err != nil {
		panic(err)
	}

	msgToBeDeleted := tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	}
	err = d.Bot.Delete(&msgToBeDeleted)
	if err != nil {
		panic(err)
	}

	for _, msgID := range captcha.AdditionalMsgs {
		msgToBeDeleted = tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		}
		err = d.Bot.Delete(&msgToBeDeleted)
		if err != nil {
			panic(err)
		}
	}

	err = removeUserFromCache(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		panic(err)
	}

}

func userExists(cache *bigcache.BigCache, key string) (bool, error) {
	users, err := cache.Get("captcha:users")
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, err
	}

	// Split the users which is in the type of []byte
	// to []string first. Then we'll iterate through it.
	// Also, we'd like to pop the last array, because it's
	// just an empty string.
	str := strings.Split(string(users), ",")[:len(strings.Split(string(users), ","))-1]
	for _, v := range str {
		if v == key {
			return true, nil
		}
	}
	return false, nil
}

func removeUserFromCache(cache *bigcache.BigCache, key string) error {
	err := cache.Delete(key)
	if err != nil {
		return err
	}

	users, err := cache.Get("captcha:users")
	if err != nil {
		return err
	}

	str := strings.Replace(string(users), key+",", "", 1)
	err = cache.Set("captcha:users", []byte(str))
	if err != nil {
		return err
	}

	return nil
}
