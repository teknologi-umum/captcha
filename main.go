package main

import (
	"context"
	"log"
	"os"
	"strings"
	"teknologi-umum-bot/handlers"
	"time"

	"github.com/allegro/bigcache/v3"
	sentry "github.com/getsentry/sentry-go"
	"github.com/go-redis/redis/v8"
	_ "github.com/joho/godotenv/autoload"
	"github.com/qiniu/qmgo"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	// Setup in memory cache
	cache, err := bigcache.NewBigCache(bigcache.Config{
		CleanWindow:      0,
		MaxEntrySize:     2048,
		HardMaxCacheSize: 1024 * 16,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Setup mongo
	mongo, err := qmgo.NewClient(context.Background(), &qmgo.Config{Uri: os.Getenv("MONGO_URL")})
	if err != nil {
		log.Fatal(err)
	}
	mongoDB := mongo.Database(os.Getenv("MONGO_DB_NAME"))

	// Setup redis
	parsedRedisURL, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	rds := redis.NewClient(parsedRedisURL)

	// Setup sentry
	logger, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Environment:      strings.Join(os.Environ(), " "),
	})
	if err != nil {
		log.Fatal(err)
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

	timers := make(chan handlers.CaptchaTimer, 100)

	deps := &handlers.Dependencies{
		Cache:   cache,
		Mongo:   mongoDB,
		Redis:   rds,
		Bot:     b,
		Context: context.Background(),
		Logger:  logger,
		Timers:  timers,
	}

	b.Handle("/start", func(m *tb.Message) {
		b.Send(m.Sender, "Hello jancuk!")
	})

	b.Handle(tb.OnUserJoined, deps.CaptchaUserJoin)

	b.Start()

	defer func() {
		if r := recover(); r != nil {
			logger.Recover(r, nil, nil)
			rds.Close()
			mongo.Close(deps.Context)
			b.Stop()
			cache.Close()
		}
	}()
}
