package captcha

import (
	"encoding/json"
	"strconv"
	"strings"
	"teknologi-umum-bot/shared"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
)

// WaitForAnswer is the handler for listening to incoming user message.
// It will uh... do a pretty long task of validating the input message.
func (d *Dependencies) WaitForAnswer(m *tb.Message) {
	// Check if the message author is in the captcha:users list or not
	// If not, return
	// If yes, check if the answer is correct or not
	exists, err := d.userExists(m.Sender.ID, m.Chat.ID)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}

	if !exists {
		return
	}

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

		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}

	var captcha Captcha
	err = json.Unmarshal(captchaData, &captcha)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}

	err = d.collectUserMessageAndCache(&captcha, m)
	if err != nil {
		shared.HandleBotError(errors.Wrap(err, "collecting user message"), d.Logger, d.Bot, m)
		return
	}

	// If the user submitted something that's a number but contains spaces,
	// we will trim the spaces down. This is because I'm lazy to not let
	// the user pass if they're actually answering the right answer
	// but got spaces on their answer. You get the idea.
	answer := removeSpaces(m.Text)

	// Check if the answer is not a number
	if _, err := strconv.Atoi(answer); errors.Is(err, strconv.ErrSyntax) {
		remainingTime := time.Until(captcha.Expiry)
		wrongMsg, err := d.Bot.Send(
			m.Chat,
			"Jawaban captcha salah, hanya terdiri dari angka kok. Kamu punya "+
				strconv.Itoa(int(remainingTime.Seconds()))+
				" detik lagi untuk menyelesaikan captcha.",
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyTo:   m,
			},
		)
		if err != nil {
			if strings.Contains(err.Error(), "replied message not found") {
				return
			}

			if strings.Contains(err.Error(), "retry after") {
				// If this happen, probably we're in a spam bot surge and would
				// probably doesn't care with the user captcha after all.
				// If they're human, they'll complete the captcha anyway,
				// or would ask to be unbanned later.
				// So, we'll just put a return here.
				return
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				// Yep, including this one.
				return
			}

			shared.HandleBotError(err, d.Logger, d.Bot, m)
			return
		}

		err = d.collectAdditionalAndCache(&captcha, m, wrongMsg)
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, m)
			return
		}

		return
	}

	// Check if the answer is correct or not
	if answer != captcha.Answer {
		remainingTime := time.Until(captcha.Expiry)
		wrongMsg, err := d.Bot.Send(
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
				return
			}

			if strings.Contains(err.Error(), "retry after") {
				// If this happen, probably we're in a spam bot surge and would
				// probably doesn't care with the user captcha after all.
				// If they're human, they'll complete the captcha anyway,
				// or would ask to be unbanned later.
				// So, we'll just put a return here.
				return
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				// Yep, including this one.
				return
			}

			shared.HandleBotError(err, d.Logger, d.Bot, m)
			return
		}

		err = d.collectAdditionalAndCache(&captcha, m, wrongMsg)
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, m)
			return
		}

		return
	}

	err = d.removeUserFromCache(m.Sender.ID, m.Chat.ID)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}

	// Congratulate the user, delete the message, then delete user from captcha:users
	// Send the welcome message to the user.
	err = d.sendWelcomeMessage(m)
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
		return
	}

	go d.Analytics.UpdateSwarm(m.Sender, m.Chat.ID, true)

	// Delete user's messages.
	for _, msgID := range captcha.UserMessages {
		if msgID == "" {
			continue
		}
		err = d.deleteMessageBlocking(&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, m)
			return
		}
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMessages {
		if msgID == "" {
			continue
		}
		err = d.deleteMessageBlocking(&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			shared.HandleBotError(err, d.Logger, d.Bot, m)
			return
		}
	}

	// Delete the question message.
	err = d.deleteMessageBlocking(&tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	})
	if err != nil {
		shared.HandleBotError(err, d.Logger, d.Bot, m)
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
