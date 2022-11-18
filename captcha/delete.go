package captcha

import (
	"fmt"
	"strconv"
	"strings"
	"teknologi-umum-bot/shared"
	"time"

	tb "gopkg.in/telebot.v3"
)

// deleteMessage creates a timer of one minute to delete a certain message.
func (d *Dependencies) deleteMessage(message *tb.StoredMessage) {
	c := make(chan struct{}, 1)
	time.AfterFunc(time.Minute*1, func() {
		for {
			err := d.Bot.Delete(message)
			if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
				if strings.Contains(err.Error(), "retry after") {
					// Acquire the retry number
					retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
					if err != nil {
						// If there's an error, we'll just retry after 15 second
						retry = 15
					}

					// Let's wait a bit and retry
					time.Sleep(time.Second * time.Duration(retry))
					continue
				}

				if strings.Contains(err.Error(), "Gateway Timeout (504)") {
					time.Sleep(time.Second * 10)
					continue
				}

				shared.HandleError(err, d.Logger)
			}

			break
		}

		c <- struct{}{}
	})

	<-c
}

func (d *Dependencies) deleteMessageBlocking(message *tb.StoredMessage) error {
	for {
		err := d.Bot.Delete(message)
		if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
			if strings.Contains(err.Error(), "retry after") {
				// Acquire the retry number
				retry, err := strconv.Atoi(strings.Split(strings.Split(err.Error(), "telegram: retry after ")[1], " ")[0])
				if err != nil {
					// If there's an error, we'll just retry after 15 second
					retry = 15
				}

				// Let's wait a bit and retry
				time.Sleep(time.Second * time.Duration(retry))
				continue
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				time.Sleep(time.Second * 10)
				continue
			}

			return fmt.Errorf("error deleting message: %w", err)
		}

		break
	}

	return nil
}
