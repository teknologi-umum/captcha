package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/getsentry/sentry-go"

	"github.com/teknologi-umum/captcha/shared"

	"github.com/pkg/errors"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// WaitForAnswer is the handler for listening to incoming user message.
// It will uh... do a pretty long task of validating the input message.
func (d *Dependencies) WaitForAnswer(ctx context.Context, m *tb.Message) {
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

	span := sentry.StartSpan(ctx, "captcha.wait_for_answer", sentry.WithTransactionSource(sentry.SourceTask),
		sentry.WithTransactionName("Captcha WaitForAnswer"))
	defer span.Finish()
	ctx = span.Context()

	// Check if the answer is correct or not.
	// If not, ask them to give the correct answer and time remaining.
	// If yes, delete the message and remove the user from the captcha:users list.
	//
	// Get the answer and all the captchaData surrounding captcha from
	// this specific user ID from the cache.
	var captchaData []byte
	err = d.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(strconv.FormatInt(m.Chat.ID, 10) + ":" + strconv.FormatInt(m.Sender.ID, 10)))
		if err != nil {
			return err
		}

		captchaData, err = item.ValueCopy(nil)
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

	var captcha Captcha
	err = json.Unmarshal(captchaData, &captcha)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	err = d.collectUserMessageAndCache(&captcha, m)
	if err != nil {
		shared.HandleBotError(ctx, errors.Wrap(err, "collecting user message"), d.Bot, m)
		return
	}

	// If the user submitted something that's a number but contains spaces,
	// we will trim the spaces down. This is because I'm lazy to not let
	// the user pass if they're actually answering the right answer
	// but got spaces on their answer. You get the idea.
	answer := strings.ToUpper(removeSpaces(m.Text))

	// Check if the answer is correct or not
	if answer != captcha.Answer {
		remainingTime := time.Until(captcha.Expiry)
		// If the current time is after the expiry time, we should return immediately.
		// We don't need to delete any message since sometimes those messages are a valid one,
		// and we are having a race due to network issues. This is not something that
		// we can easily fix, since network is not always consistent.
		if remainingTime < 0 {
			return
		}

		wrongMsg, err := d.Bot.Send(
			ctx,
			m.Chat,
			"Jawaban captcha salah, harap coba lagi. Kamu punya "+
				strconv.Itoa(int(remainingTime.Seconds()))+
				" detik lagi untuk menyelesaikan captcha.",
			&tb.SendOptions{
				ParseMode:             tb.ModeHTML,
				ReplyTo:               m,
				DisableWebPagePreview: true,
				AllowWithoutReply:     true,
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "retry after") {
				// If this happens, probably we're in a spam bot surge and would
				// probably don't care with the user captcha after all.
				// If they're human, they'll complete the captcha anyway,
				// or would ask to be unbanned later.
				// So, we'll just put a return here.
				return
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				// Yep, including this one.
				return
			}

			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}

		err = d.collectAdditionalAndCache(&captcha, m, wrongMsg)
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}

		return
	}

	err = d.removeUserFromCache(m.Sender.ID, m.Chat.ID)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "debug",
		Category: "captcha.accepted",
		Message:  "User completed a captcha",
		Data: map[string]interface{}{
			"user": m.Sender,
			"chat": m.Chat,
		},
		Level:     sentry.LevelDebug,
		Timestamp: time.Now(),
	}, &sentry.BreadcrumbHint{})

	// Congratulate the user, delete the message, then delete user from captcha:users
	// Send the welcome message to the user.
	err = d.sendWelcomeMessage(ctx, m)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}

	var messageToBeDeleted []tb.Editable
	// Delete user's messages.
	for _, msgID := range captcha.UserMessages {
		if msgID == "" {
			continue
		}
		messageToBeDeleted = append(messageToBeDeleted, &tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMessages {
		if msgID == "" {
			continue
		}
		messageToBeDeleted = append(messageToBeDeleted, &tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
	}

	// Delete the question message.
	messageToBeDeleted = append(messageToBeDeleted, &tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	})

	err = d.deleteMessageBlocking(ctx, messageToBeDeleted)
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}
}

// It... remove the user from cache. What else do you expect?
func (d *Dependencies) removeUserFromCache(userID int64, groupID int64) error {
	err := d.DB.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("captcha:users:" + strconv.FormatInt(groupID, 10)))
		if err != nil {
			return err
		}

		users, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		str := bytes.Replace(users, []byte(";"+strconv.FormatInt(userID, 10)), []byte(""), 1)

		err = txn.Set([]byte("captcha:users:"+strconv.FormatInt(groupID, 10)), str)
		if err != nil {
			return err
		}

		err = txn.Delete([]byte(strconv.FormatInt(groupID, 10) + ":" + strconv.FormatInt(userID, 10)))
		if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Uh… You should understand what this function does.
// It's pretty self-explanatory.
func removeSpaces(text string) string {
	return strings.ReplaceAll(text, " ", "")
}
