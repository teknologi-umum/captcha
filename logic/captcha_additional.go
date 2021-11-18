package logic

import (
	"encoding/json"
	"log"
	"strconv"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Collect AdditionalMsg that was sent because the user did something
// and put it on cache.
//
// It is not recommended to use it with a goroutine.
// This should be a normal blocking function.
func (d *Dependencies) collectAdditionalAndCache(captcha Captcha, m *tb.Message, wrongMsg *tb.Message) {
	// Because the wrongMsg is another message sent by us, which correlates to the
	// captcha message, we need to put the message ID into the cache.
	// So that we can delete it later.
	captcha.AdditionalMsgs = append(captcha.AdditionalMsgs, strconv.Itoa(wrongMsg.ID))

	// Update the cache with the added AdditionalMsgs
	data, err := json.Marshal(captcha)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	err = d.Cache.Set(strconv.Itoa(m.Sender.ID), data)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}
}

func (d *Dependencies) collectUsrMsgsAndCache(captcha Captcha, m *tb.Message) {
	log.Println("Func running: collectUsrMsgsAndCache")
	captcha.UserMsgs = append(captcha.UserMsgs, strconv.Itoa(m.ID))
	log.Println("Local var: captcha.UserMsgs:", captcha.UserMsgs)
	// Update the cache with the added UserMsgs
	data, err := json.Marshal(captcha)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}

	err = d.Cache.Set(strconv.Itoa(m.Sender.ID), data)
	if err != nil {
		handleError(err, d.Logger, d.Bot, m)
		return
	}
	log.Println("Func collecUsrMsgsAndCache no error")
}
