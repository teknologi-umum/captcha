package shared

import (
	"context"
	"net/http"

	tb "github.com/teknologi-umum/captcha/internal/telebot"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// HandleError handles common errors.
func HandleError(ctx context.Context, e error) {
	if e == nil {
		return
	}
	hub := sentry.GetHubFromContext(ctx)
	if hub != nil {
		hub.CaptureException(errors.WithStack(e))
	} else {
		sentry.CaptureException(errors.WithStack(e))
	}
}

// HandleBotError is the handler for an error which a function has a
// bot and a message instance.
//
// For other errors that don't have one of those struct instance, use
// HandleError instead.
func HandleBotError(ctx context.Context, e error, bot tb.API, m *tb.Message) {
	if e == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		sentry.CaptureException(errors.WithStack(e))
		return
	}

	scope := hub.Scope()
	scope.SetContext("tg:sender", map[string]interface{}{
		"id":       m.Sender.ID,
		"name":     m.Sender.FirstName + " " + m.Sender.LastName,
		"username": m.Sender.Username,
	})
	scope.SetContext("tg:message", map[string]interface{}{
		"id":   m.ID,
		"text": m.Text,
		"unix": m.Unixtime,
	})

	hub.CaptureException(errors.WithStack(e))

	_, err := bot.Send(
		ctx,
		m.Chat,
		"Oh no, something went wrong with me! Can you guys help me to ping my masters?",
		&tb.SendOptions{ParseMode: tb.ModeHTML},
	)
	if err != nil {
		// Come on? Another error?
		hub.CaptureException(errors.WithStack(err))
	}
}

// HandleHttpError handles error that has a http.Request struct instance
func HandleHttpError(ctx context.Context, e error, r *http.Request) {
	if e == nil {
		return
	}

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		sentry.CaptureException(errors.WithStack(e))
		return
	}

	scope := hub.Scope()
	scope.SetContext("http:request", map[string]interface{}{
		"method": r.Method,
		"url":    r.URL.String(),
	})

	hub.CaptureException(errors.WithStack(e))
}
