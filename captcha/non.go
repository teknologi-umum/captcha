package captcha

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"teknologi-umum-captcha/shared"
	"teknologi-umum-captcha/utils"

	tb "gopkg.in/telebot.v3"
)

// NonTextListener is the handler for every incoming payload that
// is not a text format.
func (d *Dependencies) NonTextListener(ctx context.Context, m *tb.Message) {
	// Check if the message author is in the captcha:users list or not
	// If not, return
	// If yes, check if the answer is correct or not
	exists, err := d.userExists(m.Sender.ID, m.Chat.ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	if !exists {
		return
	}

	// Check if the answer is correct or not.
	// If not, ask them to give the correct answer and time remaining.
	// If yes, delete the message and remove the user from the captcha:users list.
	//
	// Get the answer and all the data surrounding captcha from
	// this specific user ID from the cache.
	data, err := d.Memory.Get(strconv.FormatInt(m.Chat.ID, 10) + ":" + strconv.FormatInt(m.Sender.ID, 10))
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	var captcha Captcha
	err = json.Unmarshal(data, &captcha)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	// Check if the answer is a media
	remainingTime := time.Until(captcha.Expiry)
	wrongMsg, err := d.Bot.Send(
		m.Chat,
		"Hai, <a href=\"tg://user?id="+strconv.FormatInt(m.Sender.ID, 10)+"\">"+
			sanitizeInput(m.Sender.FirstName)+
			utils.ShouldAddSpace(m.Sender)+
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
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	err = d.deleteMessageBlocking(&tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: strconv.Itoa(m.ID),
	})
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	err = d.collectAdditionalAndCache(&captcha, m, wrongMsg)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}
}
