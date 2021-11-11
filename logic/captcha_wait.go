package logic

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

// It will start a timer. If the timer is expired, it will kick the user from the group.
func waitOrDelete(cache *bigcache.BigCache, logger *sentry.Client, bot *tb.Bot, msgUser *tb.Message, msgQst *tb.Message, cond *sync.Cond, done *chan bool) {
	// Let's start the timer, shall we?
	t := time.NewTimer(CAPTCHA_TIMEOUT)

	// We need to wait for the timer to expire.
	go func() {
		cond.L.Lock()

		for _, ok := <-t.C; ok; {
			// Now, when the timer is already finished, we want to check
			// whether or not the User ID is still in the cache.
			//
			// If they're still in the cache, we will say goodbye and
			// kick them from the group.
			check := cacheExists(cache, strconv.Itoa(msgUser.Sender.ID))

			if check {
				// Fetch the captcha data first
				var captcha Captcha
				user, err := cache.Get(strconv.Itoa(msgUser.Sender.ID))
				if err != nil {
					handleError(err, logger, bot, msgUser)
					return
				}

				err = json.Unmarshal(user, &captcha)
				if err != nil {
					handleError(err, logger, bot, msgUser)
					return
				}

				// Goodbye, user!
				kickMsg, err := bot.Send(msgUser.Chat,
					"<a href=\"tg://user?id="+strconv.Itoa(msgUser.Sender.ID)+"\">"+msgUser.Sender.FirstName+" "+msgUser.Sender.LastName+"</a> didn't solve the captcha. Alright, time to kick them.",
					&tb.SendOptions{
						ParseMode: tb.ModeHTML,
					})
				if err != nil {
					handleError(err, logger, bot, msgUser)
					return
				}

				// Even if the keyword is Ban, it's just kicking them.
				// If the RestrictedUntil value is below zero, it means
				// they are banned forever.
				err = bot.Ban(msgUser.Chat, &tb.ChatMember{
					RestrictedUntil: time.Now().Unix() + int64(BAN_DURATION),
					User:            msgUser.Sender,
				}, true)
				if err != nil {
					handleError(err, logger, bot, msgUser)
					return
				}

				// Delete all the message that we've sent unless the last one.
				msgToBeDeleted := tb.StoredMessage{
					ChatID:    msgUser.Chat.ID,
					MessageID: captcha.QuestionID,
				}
				err = bot.Delete(&msgToBeDeleted)
				if err != nil {
					handleError(err, logger, bot, msgUser)
					return
				}

				for _, msgID := range captcha.AdditionalMsgs {
					msgToBeDeleted = tb.StoredMessage{
						ChatID:    msgUser.Chat.ID,
						MessageID: msgID,
					}
					err = bot.Delete(&msgToBeDeleted)
					if err != nil {
						handleError(err, logger, bot, msgUser)
						return
					}
				}

				go deleteMessage(bot, kickMsg)

				err = cache.Delete(strconv.Itoa(msgUser.Sender.ID))
				if err != nil {
					handleError(err, logger, bot, msgUser)
					return
				}

				// We're done here. Let's send the value to the done channel.
				*done <- false
				return
			}
			*done <- true
			break
		}
		cond.Broadcast()
		cond.L.Unlock()
	}()
	<-*done
}
