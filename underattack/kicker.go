package underattack

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"

	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

func (d *Dependency) Kicker(ctx context.Context, c tb.Context) error {
	span := sentry.StartSpan(ctx, "underattack.kicker")
	defer span.Finish()

	for {
		err := c.Bot().Ban(ctx, c.Chat(), &tb.ChatMember{User: c.Sender(), RestrictedUntil: tb.Forever()})
		if err != nil {
			var floodError tb.FloodError
			if errors.As(err, &floodError) {
				if floodError.RetryAfter == 0 {
					floodError.RetryAfter = 15
				}

				time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
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
		err := d.Bot.Delete(ctx, c.Message())
		if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
			var floodError tb.FloodError
			if errors.As(err, &floodError) {
				if floodError.RetryAfter == 0 {
					floodError.RetryAfter = 15
				}

				time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
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
