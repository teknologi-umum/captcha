package handlers

import (
	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	"github.com/qiniu/qmgo"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Dependencies struct {
	Cache *bigcache.BigCache
	Mongo *qmgo.Client
	Redis *redis.Client
	Bot   *tb.Bot
}
