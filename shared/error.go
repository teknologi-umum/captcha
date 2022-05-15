package shared

import (
	"log"
	"net/http"
	"os"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// HandleError handles common errors.
func HandleError(e error, logger *sentry.Client) {
	if e == nil {
		return
	}

	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
	}

	_ = logger.CaptureException(
		errors.WithStack(e),
		&sentry.EventHint{OriginalException: e},
		nil,
	)
}

// HandleBotError is the handler for an error which a function has a
// bot and a message instance.
//
// For other errors that don't have one of those struct instance, use
// HandleError instead.
func HandleBotError(e error, logger *sentry.Client, bot *tb.Bot, m *tb.Message) {
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

	scope := sentry.NewScope()
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
		_ = logger.CaptureException(
			errors.WithStack(err),
			&sentry.EventHint{OriginalException: err},
			scope,
		)
	}

	_ = logger.CaptureException(
		errors.WithStack(e),
		&sentry.EventHint{OriginalException: e},
		scope,
	)
}

// HandleHttpError handles error that has a http.Request struct instance
func HandleHttpError(e error, logger *sentry.Client, r *http.Request) {
	if e == nil {
		return
	}

	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
	}

	scope := sentry.NewScope()
	scope.SetContext("http:request", map[string]interface{}{
		"method": r.Method,
		"url":    r.URL.String(),
	})

	_ = logger.CaptureException(
		errors.WithStack(e),
		&sentry.EventHint{OriginalException: e},
		scope,
	)
}
