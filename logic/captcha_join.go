package logic

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Captcha struct keep all the data needed for the captcha
// for a certain user.
//
// It will be converted to JSON format (as array of bytes or []byte)
// and then will be stored to the in memory cache, with the key
// of the corresponding Telegram User ID.
type Captcha struct {
	// Store the correct answer for the captcha
	Answer string `json:"answer"`
	// Expiry time for the captcha
	Expiry         time.Time `json:"expiry"`
	ChatID         int64     `json:"chat_id"`
	QuestionID     string    `json:"question_id"`
	AdditionalMsgs []string  `json:"additional_msgs"`
}

const (
	// How long a user will be banned in the group.
	BAN_DURATION = 1 * time.Minute
	// How long the captcha question will be valid.
	// After this time, the user will be kicked.
	// Or banned exactly, for one hour.
	CAPTCHA_TIMEOUT = 1 * time.Minute
)

// Default captcha question.
var CaptchaQuestion = `Halo, {user}!

Sebelum lanjut, selesaikan captcha ini dulu ya. Semuanya angka. Kamu punya waktu 1 menit dari sekarang!

Kalau angkanya pecah, dirotate layarnya kebentuk landscape ya.

<pre>{captcha}</pre>`

// This is the most frustrating function that I've written at this point of time.
//
// As the function name says, it will prompt a captcha to the incoming user that
// has just joined the group.
//
// At the end of the function, it will create 2 goroutines in which
// both of them are responsible for kicking the user out of the group.
func (d *Dependencies) CaptchaUserJoin(m *tb.Message) {
	// Check if the user is an admin or bot first.
	// If they are, return.
	// If they're not, continue execute the captcha.
	admins, err := d.Bot.AdminsOf(m.Chat)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	if m.UserJoined.ID != 0 {
		m.Sender = m.UserJoined
	}

	if m.Sender.IsBot || m.Private() || isAdmin(admins, m.Sender) {
		return
	}

	// randNum generates a random number (4 digit) in string format
	var randNum string = utils.GenerateRandomNumber()
	// captcha generates ascii art from the randNum value
	var captcha string = utils.GenerateAscii(randNum)

	// Replacing the template from CaptchaQuestion
	question := strings.Replace(
		strings.Replace(CaptchaQuestion, "{captcha}", captcha, 1),
		"{user}",
		"<a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+
			sanitizeInput(m.Sender.FirstName)+shouldAddSpace(m)+sanitizeInput(m.Sender.LastName)+
			"</a>",
		1,
	)

	// Send the question first.
	msgQuestion, err := d.Bot.Send(
		m.Chat,
		question,
		&tb.SendOptions{
			ParseMode:             tb.ModeHTML,
			ReplyTo:               m,
			DisableWebPagePreview: true,
		},
	)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// OK. We've sent the question. Now we are going to prepare the data that will
	// be kept on the in-memory store.
	//
	// The AdditionalMsgs key will be added later when there is an additional message
	// sent by the bot.
	captchaData, err := json.Marshal(Captcha{
		Expiry:     time.Now().Add(CAPTCHA_TIMEOUT),
		ChatID:     m.Chat.ID,
		Answer:     randNum,
		QuestionID: strconv.Itoa(msgQuestion.ID),
	})
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Yes, the cache key is their User ID in string format.
	err = d.Cache.Set(strconv.Itoa(m.Sender.ID), captchaData)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	err = d.Cache.Append("captcha:users", []byte(strconv.Itoa(m.Sender.ID)+","))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	cond := sync.NewCond(&sync.Mutex{})
	go waitOrDelete(d.Cache, d.Logger, d.Bot, m, msgQuestion, cond)
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

func sanitizeInput(inp string) string {
	var str string
	str = strings.ReplaceAll(inp, ">", "&gt;")
	str = strings.ReplaceAll(str, "<", "&lt;")
	return str
}

func shouldAddSpace(m *tb.Message) string {
	if m.Sender.LastName != "" {
		return " "
	}

	return ""
}
