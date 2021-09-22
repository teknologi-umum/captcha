package main

import (
	"context"
	"log"
	"os"
	"teknologi-umum-bot/handlers"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	"github.com/qiniu/qmgo"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	// Setup cache
	// TODO: Need to specify more config options for handling maximum cache stored and all that
	// Documentation for that: https://pkg.go.dev/github.com/allegro/bigcache?utm_source=godoc#Config
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(24 * time.Hour))
	if err != nil {
		log.Fatal(err)
	}

	// Setup mongo
	mongo, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: os.Getenv("MONGO_URL")})
	if err != nil {
		log.Fatal(err)
	}

	// Setup redis
	parsedRedisURL, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	rds := redis.NewClient(parsedRedisURL)

	// Setup bot
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	deps := &handlers.Dependencies{
		Cache: cache,
		Mongo: mongo,
		Redis: rds,
		Bot:   b,
	}

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Sender, "Hello jancuk!")
	})

	b.Handle(tb.OnUserJoined, deps.CaptchaUserJoin)

	b.Start()
}
