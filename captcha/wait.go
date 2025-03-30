package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	ctx = span.Context()
	defer span.Finish()
	// Let's start the timer, shall we?
	slog.DebugContext(ctx, "Starting timer for wait or delete procedure", slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
	time.Sleep(Timeout)
	slog.DebugContext(ctx, "Timer expired, checking if the user has completed their captcha", slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))

	// Now, when the timer is already finished, we want to check
	// whether the User ID is still in the cache.
	//
	// If they're still in the cache, we will say goodbye and
	// kick them from the group.
	check := d.cacheExists(strconv.FormatInt(msgUser.Chat.ID, 10) + ":" + strconv.FormatInt(msgUser.Sender.ID, 10))

	if check {
		slog.DebugContext(ctx, "User still exists in cache", slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
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
				slog.DebugContext(ctx, "Captcha data not found in cache, won't try to kick the user", slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
				return
			}

			slog.ErrorContext(ctx, "Failed to get captcha data from cache", slog.String("error", err.Error()), slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
			shared.HandleBotError(ctx, err, d.Bot, msgUser)
			return
		}

		slog.DebugContext(ctx, "Will try to kick the user", slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))

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

				slog.WarnContext(ctx, fmt.Sprintf("Received FloodError, retrying in %d seconds", floodError.RetryAfter), slog.String("error", err.Error()), slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID), slog.Int("retry_after", floodError.RetryAfter))
				time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
				goto KICKMSG_RETRY
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				slog.WarnContext(ctx, "Received Gateway Timeout, retrying in 10 seconds", slog.String("error", err.Error()), slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
				time.Sleep(time.Second * 10)
				goto KICKMSG_RETRY
			}

			slog.ErrorContext(ctx, "Failed to send a kick message to user", slog.String("error", err.Error()), slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
			shared.HandleBotError(ctx, err, d.Bot, msgUser)
		}

		if kickMsg != nil {
			slog.DebugContext(ctx, "Deleting the kick message", slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
			go d.deleteMessage(
				ctx,
				[]tb.Editable{&tb.StoredMessage{
					MessageID: strconv.Itoa(kickMsg.ID),
					ChatID:    msgUser.Chat.ID,
				}},
			)
		}

		err = d.removeUserFromGroup(ctx, msgUser.Chat, msgUser.Sender, captcha)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to remove user from group", slog.String("error", err.Error()), slog.Int64("group_id", msgUser.Chat.ID), slog.Int64("user_id", msgUser.Sender.ID))
			shared.HandleBotError(ctx, err, d.Bot, msgUser)
		}
	}
}
