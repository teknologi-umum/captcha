package captcha

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/shared"

	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

// deleteMessage creates a timer of one minute to delete a certain message.
func (d *Dependencies) deleteMessage(ctx context.Context, messages []tb.Editable) {
	span := sentry.StartSpan(ctx, "captcha.delete_message")
	ctx = span.Context()
	defer span.Finish()

	c := make(chan struct{}, 1)
	time.AfterFunc(time.Minute*1, func() {
		for {
			err := d.Bot.DeleteMany(ctx, messages)
			if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
				var floodError tb.FloodError
				if errors.As(err, &floodError) {
					if floodError.RetryAfter == 0 {
						floodError.RetryAfter = 15
					}

					slog.WarnContext(ctx, "Received FloodError", slog.String("error", err.Error()), slog.Int("retry_after", floodError.RetryAfter))
					time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
					continue
				}

				if strings.Contains(err.Error(), "Gateway Timeout (504)") {
					slog.WarnContext(ctx, "Received Gateway Timeout, retrying in 10 seconds", slog.String("error", err.Error()))
					time.Sleep(time.Second * 10)
					continue
				}

				shared.HandleError(ctx, err)
			}

			break
		}

		slog.DebugContext(ctx, "Message successfully deleted", slog.Any("messages", messages))
		c <- struct{}{}
	})

	<-c
}

func (d *Dependencies) deleteMessageBlocking(ctx context.Context, messages []tb.Editable) error {
	span := sentry.StartSpan(ctx, "captcha.delete_message_blocking")
	ctx = span.Context()
	defer span.Finish()

	for {
		err := d.Bot.DeleteMany(ctx, messages)
		if err != nil && !strings.Contains(err.Error(), "message to delete not found") {
			var floodError tb.FloodError
			if errors.As(err, &floodError) {
				if floodError.RetryAfter == 0 {
					floodError.RetryAfter = 15
				}

				slog.DebugContext(ctx, fmt.Sprintf("Received flood error, retrying in %d seconds", floodError.RetryAfter), slog.Any("messages", messages))
				time.Sleep(time.Second * time.Duration(floodError.RetryAfter))
				continue
			}

			if strings.Contains(err.Error(), "Gateway Timeout (504)") {
				slog.DebugContext(ctx, "Received Gateway Timeout, retrying in 10 seconds", slog.Any("messages", messages))
				time.Sleep(time.Second * 10)
				continue
			}

			return fmt.Errorf("error deleting message: %w", err)
		}

		break
	}

	slog.DebugContext(ctx, "Message successfully deleted", slog.Any("messages", messages))
	return nil
}
