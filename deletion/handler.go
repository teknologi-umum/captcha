package deletion

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

func (d *Dependency) Handler(ctx context.Context, c tb.Context) error {
	input := strings.TrimPrefix(strings.TrimPrefix(c.Text(), "/delete@TeknumCaptchaBot"), "/delete")
	if input == "" {
		return nil
	}
	if !c.Message().FromGroup() {
		return nil
	}
	if !c.Message().IsReply() {
		return nil
	}
	if c.Message().ReplyTo.Sender.ID != c.Message().Sender.ID {
		return nil
	}

	span := sentry.StartSpan(ctx, "bot.deletion_handler", sentry.WithDescription("Deletion Handler"))
	ctx = span.Context()
	defer span.Finish()

	// This is an experimental feature, so sending telemetry is a must.
	// It eases the debugging process.
	sentry.GetHubFromContext(ctx).AddBreadcrumb(&sentry.Breadcrumb{
		Type:     "info",
		Category: "deletion.handler",
		Message:  "A deletion input just came in",
		Data: map[string]interface{}{
			"Deletion Text":    input,
			"Chat ID":          c.Chat().ID,
			"Chat Username":    c.Chat().Username,
			"Chat Full Name":   c.Chat().FirstName + " " + c.Chat().LastName,
			"Chat Title":       c.Chat().Title,
			"Message ID":       c.Message().ID,
			"Sender ID":        c.Sender().ID,
			"Sender Username":  c.Sender().Username,
			"Sender Full Name": c.Sender().FirstName + " " + c.Sender().LastName,
			"From Group":       c.Message().FromGroup(),
			"From Channel":     c.Message().FromChannel(),
			"Is Forwarded":     c.Message().IsForwarded(),
		},
		Level:     "debug",
		Timestamp: time.Now(),
	}, &sentry.BreadcrumbHint{})

	duration, err := ParseDuration(ctx, input)
	if err != nil {
		if errors.Is(err, ErrParseClock) {
			return nil
		}

		sentry.GetHubFromContext(ctx).CaptureException(err)
		return nil
	}

	// Avoid abuse
	if duration <= 0 || duration >= time.Hour*24*7 {
		return nil
	}

	go func(duration2 time.Duration) {
		time.Sleep(duration2)
		_ = c.Bot().Delete(ctx, c.Message().ReplyTo)
	}(duration)

	err = c.Bot().SetMessageReaction(ctx, c.Message(), "\U0001FAE1")
	if err != nil {
		sentry.GetHubFromContext(ctx).CaptureException(err)
	}

	time.Sleep(time.Second * 10)
	_ = c.Bot().Delete(ctx, c.Message())

	return nil
}
