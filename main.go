package main

import (
	"context"
	"log"
	"os"
	"strings"
	"teknologi-umum-bot/handlers"
	"time"

	"github.com/aldy505/decrr"
	"github.com/allegro/bigcache/v3"
	sentry "github.com/getsentry/sentry-go"
	_ "github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/qiniu/qmgo"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	// Setup in memory cache
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 6))
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}

	// Setup mongo
	// mongo, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: os.Getenv("MONGO_URL")})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// mongoDB := mongo.Database(os.Getenv("MONGO_DB_NAME"))

	// // Setup redis
	// parsedRedisURL, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// rds := redis.NewClient(parsedRedisURL)

	// Setup sentry
	logger, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Environment:      strings.Join(os.Environ(), " "),
	})
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}
	defer logger.Flush(5 * time.Second)

	// Setup bot
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		if r := recover().(error); r != nil {
			logger.CaptureException(r, &sentry.EventHint{
				OriginalException: r,
			},
				nil)
			// rds.Close()
			// mongo.Close(deps.Context)
			b.Stop()
			cache.Close()
		}
	}()

	deps := &handlers.Dependencies{
		Cache: cache,
		// Mongo:   mongoDB,
		// Redis:   rds,
		Bot:     b,
		Context: context.Background(),
		Logger:  logger,
	}

	b.Handle("/start", func(m *tb.Message) {
		if m.FromGroup() {
			b.Send(m.Chat, "ok")
		}
	})

	b.Handle(tb.OnUserJoined, deps.WelcomeMessage)

	b.Handle("/setir", deps.SetirManual)
	b.Handle("/ascii", deps.Ascii)

	b.Start()
}
