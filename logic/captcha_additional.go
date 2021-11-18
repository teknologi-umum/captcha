package logic

import (
	"encoding/json"
	"strconv"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Collect AdditionalMsg that was sent because the user did something
// and put it on cache.
//
// It is not recommended to use it with a goroutine.
// This should be a normal blocking function.
func (d *Dependencies) collectAdditionalAndCache(captcha *Captcha, m *tb.Message, wrongMsg *tb.Message) error {
	// Because the wrongMsg is another message sent by us, which correlates to the
	// captcha message, we need to put the message ID into the cache.
	// So that we can delete it later.
	captcha.AdditionalMsgs = append(captcha.AdditionalMsgs, strconv.Itoa(wrongMsg.ID))

	// Update the cache with the added AdditionalMsgs
	data, err := json.Marshal(captcha)
	if err != nil {
		return err
	}

	err = d.Cache.Set(strconv.Itoa(m.Sender.ID), data)
	if err != nil {
		return err
	}

	return nil
}

func (d *Dependencies) collectUserMsgAndCache(captcha *Captcha, m *tb.Message) error {
	// We store directly the message ID that was sent by the user into the UserMsgs slices.
	captcha.UserMsgs = append(captcha.UserMsgs, strconv.Itoa(m.ID))

	// Update the cache with the added UserMsgs
	data, err := json.Marshal(captcha)
	if err != nil {
		return err
	}

	err = d.Cache.Set(strconv.Itoa(m.Sender.ID), data)
	if err != nil {
		return err
	}

	return nil
}
