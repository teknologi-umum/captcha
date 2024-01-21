package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"

	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	tb "github.com/teknologi-umum/captcha/internal/telebot"
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

	span := sentry.StartSpan(ctx, "captcha.non_text_listener", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha NonTextListener"))
	defer span.Finish()
	ctx = span.Context()

	// Check if the answer is correct or not.
	// If not, ask them to give the correct answer and time remaining.
	// If yes, delete the message and remove the user from the captcha:users list.
	//
	// Get the answer and all the data surrounding captcha from
	// this specific user ID from the cache.
	var captcha Captcha
	err = d.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(strconv.FormatInt(m.Chat.ID, 10) + ":" + strconv.FormatInt(m.Sender.ID, 10)))
		if err != nil {
			return err
		}

		rawValue, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		err = json.Unmarshal(rawValue, &captcha)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return
		}
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	// Check if the answer is a media
	remainingTime := time.Until(captcha.Expiry)
	wrongMsg, err := d.Bot.Send(
		ctx,
		m.Chat,
		"Hai, <a href=\"tg://user?id="+strconv.FormatInt(m.Sender.ID, 10)+"\">"+
			utils.SanitizeInput(m.Sender.FirstName)+
			utils.ShouldAddSpace(m.Sender)+
			utils.SanitizeInput(m.Sender.LastName)+
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

	err = d.deleteMessageBlocking(ctx, []tb.Editable{&tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: strconv.Itoa(m.ID),
	}})
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
