package captcha

import (
	"encoding/json"
	"strconv"
	"sync"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/utils"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// waitOrDelete will start a timer. If the timer is expired, it will kick the user from the group.
func (d *Dependencies) waitOrDelete(msgUser *tb.Message, cond *sync.Cond) {
	// Let's start the timer, shall we?
	t := time.NewTimer(Timeout)

	// We need to wait for the timer to expire.
	cond.L.Lock()

	for _, ok := <-t.C; ok; {
		// Now, when the timer is already finished, we want to check
		// whether the User ID is still in the cache.
		//
		// If they're still in the cache, we will say goodbye and
		// kick them from the group.
		check := cacheExists(d.Memory, strconv.Itoa(msgUser.Sender.ID))

		if check {
			// Fetch the captcha data first
			var captcha Captcha
			user, err := d.Memory.Get(strconv.Itoa(msgUser.Sender.ID))
			if err != nil {
				shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
				return
			}

			err = json.Unmarshal(user, &captcha)
			if err != nil {
				shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
				return
			}

			// Goodbye, user!
			kickMsg, err := d.Bot.Send(
				msgUser.Chat,
				"<a href=\"tg://user?id="+strconv.Itoa(msgUser.Sender.ID)+"\">"+
					sanitizeInput(msgUser.Sender.FirstName)+
					utils.ShouldAddSpace(msgUser.Sender)+
					sanitizeInput(msgUser.Sender.LastName)+
					"</a> nggak nyelesain captcha, mari kita kick!",
				&tb.SendOptions{
					ParseMode: tb.ModeHTML,
				})
			if err != nil {
				shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
				return
			}

			// Even if the keyword is Ban, it's just kicking them.
			// If the RestrictedUntil value is below zero, it means
			// they are banned forever.
			err = d.Bot.Ban(msgUser.Chat, &tb.ChatMember{
				RestrictedUntil: time.Now().Unix() + int64(BanDuration),
				User:            msgUser.Sender,
			}, true)
			if err != nil {
				shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
				return
			}

			// Delete all the message that we've sent unless the last one.
			msgToBeDeleted := tb.StoredMessage{
				ChatID:    msgUser.Chat.ID,
				MessageID: captcha.QuestionID,
			}
			err = d.Bot.Delete(&msgToBeDeleted)
			if err != nil {
				shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
				return
			}

			for _, msgID := range captcha.AdditionalMessages {
				msgToBeDeleted = tb.StoredMessage{
					ChatID:    msgUser.Chat.ID,
					MessageID: msgID,
				}
				err = d.Bot.Delete(&msgToBeDeleted)
				if err != nil {
					shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
					return
				}
			}

			go deleteMessage(
				d.Bot,
				tb.StoredMessage{
					MessageID: strconv.Itoa(kickMsg.ID),
					ChatID:    kickMsg.Chat.ID,
				},
				d.Logger,
			)

			err = d.Memory.Delete(strconv.Itoa(msgUser.Sender.ID))
			if err != nil {
				shared.HandleBotError(err, d.Logger, d.Bot, msgUser)
				return
			}

			// We're done here. Let's send the value to the done channel.
			return
		}
	}
	cond.Broadcast()
	cond.L.Unlock()
}
