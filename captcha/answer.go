package captcha

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/teknologi-umum/captcha/shared"

	"github.com/allegro/bigcache/v3"
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
	captchaData, err := d.Memory.Get(strconv.FormatInt(m.Chat.ID, 10) + ":" + strconv.FormatInt(m.Sender.ID, 10))
	if err != nil {
		if errors.Is(err, bigcache.ErrEntryNotFound) {
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
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "replied message not found") {
				// Don't retry to send the message if the user won't know
				// which message we're replying to.
				return
			}

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

	// go d.Analytics.UpdateSwarm(m.Sender, m.Chat.ID, true)

	// Delete user's messages.
	for _, msgID := range captcha.UserMessages {
		if msgID == "" {
			continue
		}
		err = d.deleteMessageBlocking(ctx, &tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMessages {
		if msgID == "" {
			continue
		}
		err = d.deleteMessageBlocking(ctx, &tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			shared.HandleBotError(ctx, err, d.Bot, m)
			return
		}
	}

	// Delete the question message.
	err = d.deleteMessageBlocking(ctx, &tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	})
	if err != nil {
		shared.HandleBotError(ctx, err, d.Bot, m)
		return
	}
}

// It... remove the user from cache. What else do you expect?
func (d *Dependencies) removeUserFromCache(userID int64, groupID int64) error {
	users, err := d.Memory.Get("captcha:users:" + strconv.FormatInt(groupID, 10))
	if err != nil {
		return err
	}

	str := strings.Replace(string(users), ";"+strconv.FormatInt(userID, 10), "", 1)
	err = d.Memory.Set("captcha:users:"+strconv.FormatInt(groupID, 10), []byte(str))
	if err != nil {
		return err
	}

	err = d.Memory.Delete(strconv.FormatInt(groupID, 10) + ":" + strconv.FormatInt(userID, 10))
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		return err
	}

	return nil
}

// Uh… You should understand what this function does.
// It's pretty self-explanatory.
func removeSpaces(text string) string {
	return strings.ReplaceAll(text, " ", "")
}
