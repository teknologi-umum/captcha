package captcha

import (
	"teknologi-umum-bot/analytics"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	tb "gopkg.in/telebot.v3"
)

// Dependencies contains the dependency injection struct for
// methods in the captcha package.
type Dependencies struct {
	Memory    *bigcache.BigCache
	Bot       *tb.Bot
	Logger    *sentry.Client
	Analytics *analytics.Dependency
	TeknumID  string
}
