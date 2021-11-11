package logic

import (
	"log"
	"os"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/getsentry/sentry-go"
)

// We handle error by apologizing to the user and then sending the error to Sentry.
func handleError(e error, logger *sentry.Client, bot *tb.Bot, m *tb.Message) {
	_, err := bot.Send(
		m.Chat,
		"Oh no, something went wrong with me! Can you guys help me to ping my masters?",
		&tb.SendOptions{ParseMode: tb.ModeHTML},
	)
	if err != nil {
		// Come on? Another error?
		_ = logger.CaptureException(
			err,
			&sentry.EventHint{OriginalException: err},
			nil,
		)
	}

	_ = logger.CaptureException(
		e,
		&sentry.EventHint{OriginalException: err},
		nil,
	)

	if os.Getenv("ENVIRONMENT") == "development" {
		log.Println(e)
	}
}
