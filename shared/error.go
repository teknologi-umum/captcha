package shared

import (
	"log"
	"net/http"
	"os"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/aldy505/decrr"
	"github.com/getsentry/sentry-go"
)

// We handle error by apologizing to the user and then sending the error to Sentry.
func HandleError(e error, logger *sentry.Client, bot *tb.Bot, m *tb.Message) {
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
			decrr.Wrap(err),
			&sentry.EventHint{OriginalException: err},
			scope,
		)
	}

	_ = logger.CaptureException(
		decrr.Wrap(e),
		&sentry.EventHint{OriginalException: e},
		scope,
	)
}

func HandleHttpError(e error, r *http.Request, logger *sentry.Client) {
	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
	}

	scope := sentry.NewScope()
	scope.SetContext("http:request", map[string]interface{}{
		"method": r.Method,
		"url":    r.URL.String(),
	})

	_ = logger.CaptureException(
		decrr.Wrap(e),
		&sentry.EventHint{OriginalException: e},
		scope,
	)
}
