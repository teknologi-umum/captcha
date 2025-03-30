package captcha

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

func (d *Dependencies) removeUserFromGroup(ctx context.Context, chat *tb.Chat, sender *tb.User, captcha Captcha) error {
	span := sentry.StartSpan(ctx, "captcha.remove_user_from_group")
	ctx = span.Context()
	defer span.Finish()

BanRetry:
	// Even if the keyword is Ban, it's just kicking them.
	// If the RestrictedUntil value is below zero, it means
	// they are banned forever.
	err := d.Bot.Ban(ctx, chat, &tb.ChatMember{
		RestrictedUntil: time.Now().Add(BanDuration).Unix(),
		User:            sender,
	}, true)
	if err != nil {
		var floodError tb.FloodError
		if errors.As(err, &floodError) {
			if floodError.RetryAfter == 0 {
				floodError.RetryAfter = 15
			}

			time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
			goto BanRetry
		}

		if strings.Contains(err.Error(), "Gateway Timeout (504)") {
			time.Sleep(time.Second * 10)
			goto BanRetry
		}

		return err
	}

	// Delete all the message that we've sent unless the last one.
	msgToBeDeleted := []tb.Editable{&tb.StoredMessage{
		ChatID:    chat.ID,
		MessageID: captcha.QuestionID,
	}}

	for _, msgID := range captcha.AdditionalMessages {
		if msgID == "" {
			continue
		}

		msgToBeDeleted = append(msgToBeDeleted, &tb.StoredMessage{
			ChatID:    chat.ID,
			MessageID: msgID,
		})
	}

	err = d.deleteMessageBlocking(ctx, msgToBeDeleted)
	if err != nil {
		return err
	}

	err = d.DB.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(strconv.FormatInt(chat.ID, 10) + ":" + strconv.FormatInt(sender.ID, 10)))
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}

	return nil
}
