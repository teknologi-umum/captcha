package logic

import (
	"encoding/json"
	"strconv"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// This is the handler for every incoming payload that
// is not a text format.
//
func (d *Dependencies) NonTextListener(m *tb.Message) {
	// Check if the message author is in the captcha:users list or not
	// If not, return
	// If yes, check if the answer is correct or not
	exists, err := userExists(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	if !exists {
		return
	}

	// Check if the answer is correct or not.
	// If not, ask them to give the correct answer and time remaining.
	// If yes, delete the message and remove the user from the captcha:users list.
	//
	// Get the answer and all of the data surrounding captcha from
	// this specific user ID from the cache.
	data, err := d.Cache.Get(strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	var captcha Captcha
	err = json.Unmarshal(data, &captcha)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Check if the answer is a media
	remainingTime := time.Until(captcha.Expiry)
	wrongMsg, err := d.Bot.Send(
		m.Chat,
		"Hai, <a href=\"tg://user?id="+strconv.Itoa(m.Sender.ID)+"\">"+
			sanitizeInput(m.Sender.FirstName)+
			shouldAddSpace(m)+
			sanitizeInput(m.Sender.LastName)+
			"</a>. "+
			"Selesain captchanya dulu yuk, baru kirim yang aneh-aneh. Kamu punya "+
			strconv.Itoa(int(remainingTime.Seconds()))+
			" detik lagi, kalau nggak, saya kick!",
		&tb.SendOptions{
			ParseMode:             tb.ModeHTML,
			DisableWebPagePreview: true,
		},
	)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	err = d.Bot.Delete(m)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	collectAdditionalAndCache(d.Cache, d.Bot, d.Logger, captcha, m, wrongMsg)
}
