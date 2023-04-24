package shared

import (
	"context"
	"log"
	"net/http"
	"os"

	tb "gopkg.in/telebot.v3"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// HandleError handles common errors.
func HandleError(ctx context.Context, e error) {
	if e == nil {
		return
	}

	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
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
func HandleBotError(ctx context.Context, e error, bot *tb.Bot, m *tb.Message) {
	if e == nil {
		return
	}

	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
	}

	_, err := bot.Send(
		m.Chat,
		"Oh no, something went wrong with me! Can you guys help me to ping my masters?",
		&tb.SendOptions{ParseMode: tb.ModeHTML},
	)

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

	if err != nil {
		// Come on? Another error?
		hub.CaptureException(errors.WithStack(err))
	}

	hub.CaptureException(errors.WithStack(e))
}

// HandleHttpError handles error that has a http.Request struct instance
func HandleHttpError(ctx context.Context, e error, r *http.Request) {
	if e == nil {
		return
	}

	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
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
