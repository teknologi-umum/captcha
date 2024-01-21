package captcha

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	"github.com/pkg/errors"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// waitOrDelete will start a timer. If the timer is expired, it will kick the user from the group.
func (d *Dependencies) waitOrDelete(ctx context.Context, msgUser *tb.Message) {
	span := sentry.StartSpan(ctx, "captcha.wait_or_delete")
	defer span.Finish()
	// Let's start the timer, shall we?
	time.Sleep(Timeout)

	// Now, when the timer is already finished, we want to check
	// whether the User ID is still in the cache.
	//
	// If they're still in the cache, we will say goodbye and
	// kick them from the group.
	check := d.cacheExists(strconv.FormatInt(msgUser.Chat.ID, 10) + ":" + strconv.FormatInt(msgUser.Sender.ID, 10))

	if check {
		// Fetch the captcha data first
		var captcha Captcha
		err := d.DB.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(strconv.FormatInt(msgUser.Chat.ID, 10) + ":" + strconv.FormatInt(msgUser.Sender.ID, 10)))
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

			shared.HandleBotError(ctx, err, d.Bot, msgUser)
			return
		}

	KICKMSG_RETRY:
		// Goodbye, user!
		kickMsg, err := d.Bot.Send(
			ctx,
			msgUser.Chat,
			"<a href=\"tg://user?id="+strconv.FormatInt(msgUser.Sender.ID, 10)+"\">"+
				utils.SanitizeInput(msgUser.Sender.FirstName)+
				utils.ShouldAddSpace(msgUser.Sender)+
				utils.SanitizeInput(msgUser.Sender.LastName)+
				"</a> tidak menyelesaikan captcha, saya kick!",
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
			})
		if err != nil {
			var floodError tb.FloodError
			if errors.As(err, &floodError) {
				if floodError.RetryAfter == 0 {
					floodError.RetryAfter = 15
				}

				time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
				goto KICKMSG_RETRY
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				time.Sleep(time.Second * 10)
				goto KICKMSG_RETRY
			}

			shared.HandleBotError(ctx, err, d.Bot, msgUser)
		}

		go d.deleteMessage(
			ctx,
			[]tb.Editable{&tb.StoredMessage{
				MessageID: strconv.Itoa(kickMsg.ID),
				ChatID:    kickMsg.Chat.ID,
			}},
		)

		err = d.removeUserFromGroup(ctx, msgUser.Chat, msgUser.Sender, captcha)
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, msgUser)
		}
	}
}
