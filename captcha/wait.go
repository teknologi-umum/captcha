package captcha

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/utils"

	"github.com/allegro/bigcache/v3"
	"github.com/pkg/errors"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// waitOrDelete will start a timer. If the timer is expired, it will kick the user from the group.
func (d *Dependencies) waitOrDelete(ctx context.Context, msgUser *tb.Message) {
	// Let's start the timer, shall we?
	t := time.NewTimer(Timeout)

	for _, ok := <-t.C; ok; {
		// Now, when the timer is already finished, we want to check
		// whether the User ID is still in the cache.
		//
		// If they're still in the cache, we will say goodbye and
		// kick them from the group.
		check := d.cacheExists(strconv.FormatInt(msgUser.Chat.ID, 10) + ":" + strconv.FormatInt(msgUser.Sender.ID, 10))

		if check {
			// Fetch the captcha data first
			var captcha Captcha
			user, err := d.Memory.Get(strconv.FormatInt(msgUser.Chat.ID, 10) + ":" + strconv.FormatInt(msgUser.Sender.ID, 10))
			if err != nil {
				shared.HandleBotError(ctx, err, d.Bot, msgUser)
				break
			}

			err = json.Unmarshal(user, &captcha)
			if err != nil {
				shared.HandleBotError(ctx, err, d.Bot, msgUser)
				break
			}

		KICKMSG_RETRY:
			// Goodbye, user!
			kickMsg, err := d.Bot.Send(
				ctx,
				msgUser.Chat,
				"<a href=\"tg://user?id="+strconv.FormatInt(msgUser.Sender.ID, 10)+"\">"+
					utils.SanitizeInput(msgUser.Sender.FirstName)+
					utils.ShouldAddSpace(msgUser.Sender)+
					utils.SanitizeInput(msgUser.Sender.LastName)+
					"</a> nggak nyelesain captcha, mari kita kick!",
				&tb.SendOptions{
					ParseMode: tb.ModeHTML,
				})
			if err != nil {
				var floodError tb.FloodError
				if errors.As(err, &floodError) {
					if floodError.RetryAfter == 0 {
						floodError.RetryAfter = 15
					}

					time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
					goto KICKMSG_RETRY
				}

				if strings.Contains(err.Error(), "Gateway Timeout (504)") {
					time.Sleep(time.Second * 10)
					goto KICKMSG_RETRY
				}

				shared.HandleBotError(ctx, err, d.Bot, msgUser)
				break
			}

		BAN_RETRY:
			// Even if the keyword is Ban, it's just kicking them.
			// If the RestrictedUntil value is below zero, it means
			// they are banned forever.
			err = d.Bot.Ban(ctx, msgUser.Chat, &tb.ChatMember{
				RestrictedUntil: time.Now().Add(BanDuration).Unix(),
				User:            msgUser.Sender,
			}, true)
			if err != nil {
				var floodError tb.FloodError
				if errors.As(err, &floodError) {
					if floodError.RetryAfter == 0 {
						floodError.RetryAfter = 15
					}

					time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
					goto BAN_RETRY
				}

				if strings.Contains(err.Error(), "Gateway Timeout (504)") {
					time.Sleep(time.Second * 10)
					goto BAN_RETRY
				}

				shared.HandleBotError(ctx, err, d.Bot, msgUser)
				break
			}

			// Delete all the message that we've sent unless the last one.
			msgToBeDeleted := []tb.Editable{&tb.StoredMessage{
				ChatID:    msgUser.Chat.ID,
				MessageID: captcha.QuestionID,
			}}

			for _, msgID := range captcha.AdditionalMessages {
				if msgID == "" {
					continue
				}

				msgToBeDeleted = append(msgToBeDeleted, &tb.StoredMessage{
					ChatID:    msgUser.Chat.ID,
					MessageID: msgID,
				})
			}

			err = d.deleteMessageBlocking(ctx, msgToBeDeleted)
			if err != nil {
				shared.HandleBotError(ctx, err, d.Bot, msgUser)
				break
			}

			go d.deleteMessage(
				ctx,
				[]tb.Editable{&tb.StoredMessage{
					MessageID: strconv.Itoa(kickMsg.ID),
					ChatID:    kickMsg.Chat.ID,
				}},
			)

			err = d.Memory.Delete(strconv.FormatInt(msgUser.Chat.ID, 10) + ":" + strconv.FormatInt(msgUser.Sender.ID, 10))
			if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
				shared.HandleBotError(ctx, err, d.Bot, msgUser)
				break
			}
		}

		break
	}
}
