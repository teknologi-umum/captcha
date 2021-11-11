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
	_ "github.com/joho/godotenv/autoload"
	tb "gopkg.in/tucnak/telebot.v2"
)

func main() {
	// Setup in memory cache
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 12))
	if err != nil {
		log.Fatal(decrr.Wrap(err))
	}
	defer cache.Close()

	// Setup redis
	// parsedRedisURL, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// rds := redis.NewClient(parsedRedisURL)
	// defer rds.Close()

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
			b.Stop()
			// cache.Flush()
		}
	}()

	// Kalo mau di rename di file handlers/dependencies.go juga gapapa.
	// Cache -> In memory cache (https://github.com/allegro/bigcache)
	//   Si BigCache ini dia punya function namanya Append(), nanti kedepannya bisa dimanfaatkan
	//   buat berurusan sama slice/list of string.
	// Redis -> Ya redis (https://pkg.go.dev/github.com/go-redis/redis/v8@v8.11.3)
	deps := &handlers.Dependencies{
		Cache: cache,
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

	// (aldy505): Jadi ini deps.WelcomeMessage diganti sama deps.CaptchaUserJoin
	// Setelah itu, nggak tau ya, kayaknya bakal listen to the whole chat.
	// If (message.Sender.ID is in array of ongoing captchas) {
	//   check if the message contains the captcha answer, while still keeping the timer running
	//   if (message is the captcha answer) {
	//      - cancel the timer
	//      - remove the user ID from redis and in memory cache
	//      - send congratulations message
	//   } else {
	//      do nothing, keep the timer running
	//   }
	// } else {
	//   do nothing, keep the timer running
	// }
	//
	// Documentation soal package telebot ada disini: https://pkg.go.dev/gopkg.in/tucnak/telebot.v2
	//
	// Dari sini kamu ke handlers/captcha.go
	b.Handle(tb.OnUserJoined, deps.CaptchaUserJoin)
	b.Handle(tb.OnText, deps.WaitForAnswer)
	b.Handle("/ascii", deps.Ascii)
	b.Handle("/captcha", deps.CaptchaUserJoin)

	b.SetCommands([]tb.Command{
		{Text: "ascii", Description: "Sends ASCII generated text."},
	})

	b.Start()
}
