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
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"teknologi-umum-bot/logic"
	"time"

	"github.com/aldy505/decrr"
	"github.com/allegro/bigcache/v3"
	sentry "github.com/getsentry/sentry-go"
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
}

func main() {
	// Setup in memory cache
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 12))
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}
	defer cache.Close()

	// This Redis line below is commented out because it's not needed
	// for now. Yet, while I'm a shaman, I'm not sure if I'll need it
	// or not. So, I'll just leave it here.
	//
	// Setup redis
	// parsedRedisURL, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// rds := redis.NewClient(parsedRedisURL)
	// defer rds.Close()

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
		log.Fatal(err)
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

	deps := &logic.Dependencies{
		Cache:   cache,
		Bot:     b,
		Context: context.Background(),
		Logger:  logger,
	}

	// This is basically just for health check.
	b.Handle("/start", func(m *tb.Message) {
		if m.FromGroup() {
			b.Send(m.Chat, "ok")
		}
	})

	// Captcha handlers
	b.Handle(tb.OnUserJoined, deps.CaptchaUserJoin)
	b.Handle(tb.OnText, deps.WaitForAnswer)
	b.Handle(tb.OnPhoto, deps.NonTextListener)
	b.Handle(tb.OnAnimation, deps.NonTextListener)
	b.Handle(tb.OnVideo, deps.NonTextListener)
	b.Handle(tb.OnDocument, deps.NonTextListener)
	b.Handle(tb.OnSticker, deps.NonTextListener)
	b.Handle(tb.OnVoice, deps.NonTextListener)
	b.Handle(tb.OnVideoNote, deps.NonTextListener)
	b.Handle(tb.OnUserLeft, deps.CaptchaUserLeave)

	b.Handle("/ascii", deps.Ascii)

	err = b.SetCommands([]tb.Command{
		{Text: "ascii", Description: "Sends ASCII generated text."},
	})
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}

	log.Println("Bot started!")
	go func() {
		b.Start()
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Shutdown signal received, exiting...")
}
