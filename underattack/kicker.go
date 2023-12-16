package underattack

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"strconv"
	"strings"
	"time"

	tb "gopkg.in/telebot.v3"
)

func (d *Dependency) Kicker(ctx context.Context, c tb.Context) error {
	span := sentry.StartSpan(ctx, "underattack.kicker")
	defer span.Finish()
	
	for {
		err := c.Bot().Ban(c.Chat(), &tb.ChatMember{User: c.Sender(), RestrictedUntil: tb.Forever()})
		if err != nil {
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

			return fmt.Errorf("error banning user: %w", err)
		}

		break
	}

	for {
		err := d.Bot.Delete(c.Message())
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
