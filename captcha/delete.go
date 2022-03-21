package captcha

import (
	"strconv"
	"strings"
	"teknologi-umum-bot/shared"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// deleteMessage creates a timer of one minute to delete a certain message.
func (d *Dependencies) deleteMessage(message *tb.StoredMessage) {
	c := make(chan struct{}, 1)
	time.AfterFunc(time.Minute*1, func() {
	DELETEMSG_RETRY:
		err := d.Bot.Delete(message)
		if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
			if strings.Contains(err.Error(), "retry after") {
				// Acquire the retry number
				retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
				if err != nil {
					// If there's an error, we'll just retry after 10 second
					retry = 10
				}

				// Let's wait a bit and retry
				time.Sleep(time.Second * time.Duration(retry))
				goto DELETEMSG_RETRY
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				time.Sleep(time.Second * 10)
				goto DELETEMSG_RETRY
			}

			shared.HandleError(err, d.Logger)
		}
		c <- struct{}{}
	})

	<-c
}

func (d *Dependencies) deleteMessageBlocking(message *tb.StoredMessage) error {
DELETEMSG_RETRY:
	err := d.Bot.Delete(message)
	if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
		if strings.Contains(err.Error(), "retry after") {
			// Acquire the retry number
			retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
			if err != nil {
				// If there's an error, we'll just retry after 10 second
				retry = 10
			}

			// Let's wait a bit and retry
			time.Sleep(time.Second * time.Duration(retry))
			goto DELETEMSG_RETRY
		}

		if strings.Contains(err.Error(), "Gateway Timeout (504)") {
			time.Sleep(time.Second * 10)
			goto DELETEMSG_RETRY
		}

		return err
	}

	return nil
}
