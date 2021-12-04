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
// say is just.. good luck.
//
// This source code is very ugly. Let me tell you that up front.
package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/cmd"

	"time"

	"github.com/aldy505/decrr"
	"github.com/allegro/bigcache/v3"
	sentry "github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
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

	sentry := os.Getenv("SENTRY_DSN")
	if env == "production" && sentry == "" {
		log.Fatal("Please provide the SENTRY_DSN value on the .env file")
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL == "" || !strings.HasPrefix(dbURL, "postgresql://") {
		log.Fatal("Please provide the correct DATABASE_URL value on the .env file")
	}

	if redisURL := os.Getenv("REDIS_URL"); redisURL == "" || !strings.HasPrefix(redisURL, "redis://") {
		log.Fatal("Please provide the correct REDIS_URL value on the .env file")
	}

	if tz := os.Getenv("TZ"); tz == "" {
		log.Println("You are encouraged to provide the TZ value to UTC, but eh..")
	}
}

func main() {
	db, err := sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}
	defer db.Close()

	// Setup in memory cache
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 12))
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}
	defer cache.Close()

	// Setup Sentry for error handling.
	logger, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Debug:            os.Getenv("ENVIRONMENT") == "development",
		Environment:      os.Getenv("ENVIRONMENT"),
	})
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}
	defer logger.Flush(5 * time.Second)

	// Running migration on database first.
	err = analytics.MustMigrate(db)
	if err != nil {
		log.Fatal(decrr.Wrap(err))
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
		log.Fatal(decrr.Wrap(err))
	}
	defer b.Stop()

	// This is for recovering from panic.
	defer func() {
		r := recover()
		if r != nil {
			_ = logger.CaptureException(r.(error), &sentry.EventHint{
				OriginalException: r.(error),
			}, nil)
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
			b.Send(m.Chat, "ok")
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

	b.Handle("/ascii", deps.AsciiCmdHandler)

	log.Println("Bot started!")
	go func() {
		b.Start()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Shutdown signal received, exiting...")
}
