// Hello!
//
// This is the source code for @TeknumCaptchaBot where you can find
// the ugly code behind @TeknumCaptchaBot's captcha feature and more.
//
// If you are learning Go for the first time and about to browse this
// repository as one of your first steps, you might want to read the
// other repository on the organization. It's far easier.
// Here: https://github.com/teknologi-umum/polarite
//
// Unless, you're stubborn and want to learn the hard way, all I can
// say is just... good luck.
//
// This source code is very ugly. Let me tell you that up front.
package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// Internals
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/analytics/server"
	"teknologi-umum-bot/cmd"
	"teknologi-umum-bot/shared"

	// Database and cache
	"github.com/allegro/bigcache/v3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	// Others third party stuff
	"github.com/getsentry/sentry-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v2"
)

// This init function checks if there's any configuration
// missing from the .env file.
func init() {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		log.Fatal("Please provide the ENVIRONMENT value on the .env file")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Please provide the BOT_TOKEN value on the .env file")
	}

	sentryDSN := os.Getenv("SENTRY_DSN")
	if env == "production" && sentryDSN == "" {
		log.Fatal("Please provide the SENTRY_DSN value on the .env file")
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL == "" || !strings.HasPrefix(dbURL, "postgres") {
		log.Fatal("Please provide the correct DATABASE_URL value on the .env file")
	}

	if redisURL := os.Getenv("REDIS_URL"); redisURL == "" || !strings.HasPrefix(redisURL, "redis") {
		log.Fatal("Please provide the correct REDIS_URL value on the .env file")
	}

	if mongoURL := os.Getenv("MONGO_URL"); mongoURL == "" || !strings.HasPrefix(mongoURL, "mongodb") {
		log.Fatal("Please provide the correct MONGO_URL value on the .env file")
	}
	if os.Getenv("TZ") == "" {
		err := os.Setenv("TZ", "UTC")
		if err != nil {
			log.Fatalln("during setting TZ environment variable:", err)
		}
	}

	log.Println("Passed the environment variable check")
}

func main() {
	// Context for initiating database connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Connect to PostgreSQL
	db, err := sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("during opening a postgres client:", errors.WithStack(err))
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal("during closing the postgres client:", errors.WithStack(err))
		}
	}(db)

	// Setup mongodb connection
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal("during connecting to mongo client:", errors.WithStack(err))
	}
	defer func(client *mongo.Client) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal("during closing the mongo connection:", errors.WithStack(err))
		}
	}(mongoClient)

	// Mongo health check
	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("during mongodb ping:", err)
	}

	// Setup in memory cache
	cache, err := bigcache.NewBigCache(bigcache.Config{
		Shards:             1024,
		LifeWindow:         time.Minute * 5,
		CleanWindow:        time.Minute * 1,
		Verbose:            true,
		HardMaxCacheSize:   1024 * 1024 * 1024,
		MaxEntrySize:       500,
		MaxEntriesInWindow: 50,
	})
	if err != nil {
		log.Fatal("during creating a in memory cache:", errors.WithStack(err))
	}
	defer func(cache *bigcache.BigCache) {
		err := cache.Close()
		if err != nil {
			log.Fatal(errors.WithStack(err))
		}
	}(cache)

	// Setup Sentry for error handling.
	logger, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Debug:            os.Getenv("ENVIRONMENT") == "development",
		Environment:      os.Getenv("ENVIRONMENT"),
	})
	if err != nil {
		log.Fatal("during initiating a new sentry client:", errors.WithStack(err))
	}
	defer logger.Flush(5 * time.Second)

	// Running migration on database first.
	err = analytics.MustMigrate(db)
	if err != nil {
		log.Fatal("during initial database migration:", errors.WithStack(err))
	}

	// Setup Telegram Bot
	b, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Reporter: func(err error) {
			_ = logger.CaptureException(
				err,
				&sentry.EventHint{OriginalException: err},
				nil,
			)
		},
	})
	if err != nil {
		log.Fatal("during init of bot client:", errors.WithStack(err))
	}
	defer b.Stop()

	// This is for recovering from panic.
	defer func() {
		r := recover()
		if r != nil {
			_ = logger.CaptureException(r.(error), &sentry.EventHint{
				OriginalException: r.(error),
			}, nil)
			log.Println(r.(error))
		}
	}()

	deps := cmd.New(cmd.Dependency{
		Memory: cache,
		Bot:    b,
		Logger: logger,
		DB:     db,
	})

	// This is basically just for health check.
	b.Handle("/start", func(m *tb.Message) {
		if m.FromGroup() {
			_, err := b.Send(m.Chat, "ok")
			if err != nil {
				shared.HandleBotError(err, logger, b, m)
				return
			}
		}
	})

	// Captcha handlers
	b.Handle(tb.OnUserJoined, deps.OnUserJoinHandler)
	b.Handle(tb.OnText, deps.OnTextHandler)
	b.Handle(tb.OnPhoto, deps.OnNonTextHandler)
	b.Handle(tb.OnAnimation, deps.OnNonTextHandler)
	b.Handle(tb.OnVideo, deps.OnNonTextHandler)
	b.Handle(tb.OnDocument, deps.OnNonTextHandler)
	b.Handle(tb.OnSticker, deps.OnNonTextHandler)
	b.Handle(tb.OnVoice, deps.OnNonTextHandler)
	b.Handle(tb.OnVideoNote, deps.OnNonTextHandler)
	b.Handle(tb.OnUserLeft, deps.OnUserLeftHandler)

	log.Println("Bot started!")
	go func() {
		b.Start()
	}()

	go func() {
		// Parse mongo url
		parsedURL, err := url.Parse(os.Getenv("MONGO_URL"))
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to parse MONGO_URL"))
		}

		server.New(server.Config{
			DB:          db,
			Memory:      cache,
			Mongo:       mongoClient,
			Logger:      logger,
			MongoDBName: parsedURL.Path[1:],
			Port:        os.Getenv("PORT"),
		})
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Shutdown signal received, exiting...")
}
