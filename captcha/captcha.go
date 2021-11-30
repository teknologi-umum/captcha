package captcha

import (
	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v8"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Dependencies struct {
	Memory *bigcache.BigCache
	Redis  *redis.Client
	Bot    *tb.Bot
	Logger *sentry.Client
}
