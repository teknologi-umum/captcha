package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"

	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// CaptchaUserLeave handles the event when a user left the group.
// This will check if the user is in the memory of current active
// captcha or not.
//
// If it is, the captcha will be deleted.
func (d *Dependencies) CaptchaUserLeave(ctx context.Context, m *tb.Message) {
	// Check if the user is an admin or bot first.
	// If they are, return.
	// If they're not, continue to execute the captcha.
	admins, err := d.Bot.AdminsOf(ctx, m.Chat)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	if m.Sender.IsBot || m.Private() || utils.IsAdmin(admins, m.Sender) {
		return
	}

	// We need to check if the user is in the captcha:users cache
	// or not.
	check, err := d.userExists(m.Sender.ID, m.Chat.ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	if !check {
		return
	}

	span := sentry.StartSpan(ctx, "captcha.captcha_user_leave", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha CaptchaUserLeave"))
	defer span.Finish()
	ctx = span.Context()

	// OK, they exist in the cache. Now we've got to delete
	// all the message that we've sent before.
	var captcha Captcha
	err = d.DB.View(func(txn *badger.Txn) error {
		defer txn.Discard()

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

		return txn.Commit()
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return
		}
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	err = d.removeUserFromCache(m.Sender.ID, m.Chat.ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	// Build message to be deleted
	messagesToBeDeleted := []tb.Editable{
		// Delete question message
		&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: captcha.QuestionID,
		},
	}

	// Delete user's messages.
	for _, msgID := range captcha.UserMessages {
		if msgID == "" {
			continue
		}

		messagesToBeDeleted = append(messagesToBeDeleted, &tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMessages {
		if msgID == "" {
			continue
		}

		messagesToBeDeleted = append(messagesToBeDeleted, &tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
	}

	err = d.deleteMessageBlocking(ctx, messagesToBeDeleted)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}
}
