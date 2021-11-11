package logic

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/allegro/bigcache/v3"
	tb "gopkg.in/tucnak/telebot.v2"
)

// This is the handler for listening to incoming user message.
// It will uh... do a pretty long task of validating the input message.
func (d *Dependencies) WaitForAnswer(m *tb.Message) {
	// Check if the message author is in the captcha:users list or not
	// If not, return
	// If yes, check if the answer is correct or not
	check, err := userExists(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	if !check {
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

	// Check if the answer is correct or not
	if m.Text != captcha.Answer {
		remainingTime := time.Until(captcha.Expiry)
		wrongMsg, err := d.Bot.Send(
			m.Chat,
			"Wrong answer, please try again. You have "+strconv.Itoa(int(remainingTime.Seconds()))+" more second to solve the captcha.",
			&tb.SendOptions{
				ParseMode: tb.ModeHTML,
				ReplyTo:   m,
			},
		)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}

		// Because the wrongMsg is another message sent by us, which correlates to the
		// captcha message, we need to put the message ID into the cache.
		// So that we can delete it later.
		captcha.AdditionalMsgs = append(captcha.AdditionalMsgs, strconv.Itoa(wrongMsg.ID))

		// Update the cache with the added AdditionalMsgs
		data, err = json.Marshal(captcha)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}

		err = d.Cache.Set(strconv.Itoa(m.Sender.ID), data)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}

		return
	}

	// Congratulate the user, delete the message, then delete user from captcha:users
	// Send the welcome message to the user.
	go sendWelcomeMessage(d.Bot, m, d.Logger)

	// Delete the question message.
	msgToBeDeleted := tb.StoredMessage{
		ChatID:    m.Chat.ID,
		MessageID: captcha.QuestionID,
	}
	err = d.Bot.Delete(&msgToBeDeleted)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	// Delete any additional message.
	for _, msgID := range captcha.AdditionalMsgs {
		msgToBeDeleted = tb.StoredMessage{
			ChatID:    m.Chat.ID,
			MessageID: msgID,
		}
		err = d.Bot.Delete(&msgToBeDeleted)
		if err != nil {
			handleError(err, d.Logger, d.Bot, m)
			return
		}
	}

	// TODO: Delete the user answers. But uhh, I don't really think
	// that that's necessary. But, we'll see.

	err = removeUserFromCache(d.Cache, strconv.Itoa(m.Sender.ID))
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}
}

// It... remove the user from cache. What else do you expect?
func removeUserFromCache(cache *bigcache.BigCache, key string) error {
	err := cache.Delete(key)
	if err != nil {
		return err
	}

	users, err := cache.Get("captcha:users")
	if err != nil {
		return err
	}

	str := strings.Replace(string(users), key+",", "", 1)
	err = cache.Set("captcha:users", []byte(str))
	if err != nil {
		return err
	}

	return nil
}
