package logic

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// This is the handler for listening to incoming user message.
// It will uh... do a pretty long task of validating the input message.
func (d *Dependencies) WaitForAnswer(m *tb.Message) {
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

	err = d.collectUserMsgAndCache(&captcha, m)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
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
			handleError(err, d.Logger, d.Bot, m)
			return
		}

		err = d.collectAdditionalAndCache(&captcha, m, wrongMsg)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
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
			handleError(err, d.Logger, d.Bot, m)
			return
		}

		err = d.collectAdditionalAndCache(&captcha, m, wrongMsg)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}

		return
	}

	err = d.removeUserFromCache(strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Congratulate the user, delete the message, then delete user from captcha:users
	// Send the welcome message to the user.
	go sendWelcomeMessage(d.Bot, m, d.Logger)

	// Delete the question message.
	err = d.Bot.Delete(&tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	})
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Delete user's messages.
	for _, msgID := range captcha.UserMsgs {
		if msgID == "" {
			continue
		}
		err = d.Bot.Delete(&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMsgs {
		if msgID == "" {
			continue
		}
		err = d.Bot.Delete(&tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		})
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}
}

// It... remove the user from cache. What else do you expect?
func (d *Dependencies) removeUserFromCache(key string) error {
	users, err := d.Cache.Get("captcha:users")
	if err != nil {
		return err
	}

	str := strings.Replace(string(users), ";"+key, "", 1)
	err = d.Cache.Set("captcha:users", []byte(str))
	if err != nil {
		return err
	}

	err = d.Cache.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

// Uh.. You should understand what this function does.
// It's pretty self explanatory.
func removeSpaces(text string) string {
	return strings.ReplaceAll(text, " ", "")
}
