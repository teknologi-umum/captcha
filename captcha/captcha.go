package captcha

import (
	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Dependencies struct {
	Memory *bigcache.BigCache
	Bot    *tb.Bot
	Logger *sentry.Client
}
