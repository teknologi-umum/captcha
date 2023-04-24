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
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	// Internals
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/analytics/server"
	"teknologi-umum-bot/cmd"
	"teknologi-umum-bot/shared"
	"teknologi-umum-bot/underattack"

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
	tb "gopkg.in/telebot.v3"
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
	// Setup Sentry for error handling.
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Debug:            os.Getenv("ENVIRONMENT") == "development",
		Environment:      os.Getenv("ENVIRONMENT"),
	})
	if err != nil {
		log.Fatal("during initiating a new sentry client:", errors.WithStack(err))
	}
	defer sentry.Flush(30 * time.Second)

	// Connect to PostgreSQL
	db, err := sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("during opening a postgres client:", errors.WithStack(err))
	}
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Print("during closing the postgres client:", errors.WithStack(err))
		}
	}(db)

	// Context for initiating database connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	// Setup mongodb connection
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal("during connecting to mongo client:", errors.WithStack(err))
	}
	defer func(client *mongo.Client) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		defer cancel()
		err := client.Disconnect(ctx)
		if err != nil {
			log.Print("during closing the mongo connection:", errors.WithStack(err))
		}
	}(mongoClient)

	// Mongo health check
	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal("during mongodb ping:", err)
	}

	// Get the MongoDB database name from the given MONGO_URL environment variable.
	parsedURL, err := url.Parse(os.Getenv("MONGO_URL"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse MONGO_URL"))
	}
	mongoDBName := parsedURL.Path[1:]

	// Setup in memory cache
	cache, err := bigcache.New(context.Background(), bigcache.Config{
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
			log.Print(errors.WithStack(err))
		}
	}(cache)

	// Running migration on database first.
	err = analytics.MustMigrate(db)
	if err != nil {
		log.Fatal("during initial database migration:", errors.WithStack(err))
	}
	err = underattack.MustMigrate(db)
	if err != nil {
		log.Fatal("during initial database migration:", errors.WithStack(err))
	}

	// Setup Telegram Bot
	b, err := tb.NewBot(tb.Settings{
		Token:       os.Getenv("BOT_TOKEN"),
		Poller:      &tb.LongPoller{Timeout: 10 * time.Second},
		Synchronous: false,
		OnError: func(err error, ctx tb.Context) {
			if strings.Contains(err.Error(), "Conflict: terminated by other getUpdates request") {
				// This error means the bot is currently being deployed
				return
			}

			sentry.CaptureException(err)
		},
		Client: &http.Client{
			Timeout: time.Hour,
			Transport: &http.Transport{
				Proxy:                 http.ProxyFromEnvironment,
				TLSHandshakeTimeout:   time.Minute * 3,
				ForceAttemptHTTP2:     true,
				IdleConnTimeout:       time.Minute * 3,
				ExpectContinueTimeout: time.Minute,
				DialContext: (&net.Dialer{
					Timeout:   time.Minute * 3,
					KeepAlive: time.Minute,
				}).DialContext,
			},
		},
	})
	if err != nil {
		sentry.CaptureException(fmt.Errorf("initializing bot client: %w", err))
		log.Fatal("during init of bot client:", errors.WithStack(err))
	}
	defer func() {
		_, err := b.Close()
		if err != nil {
			sentry.CaptureException(err)
			log.Print(errors.WithStack(err))
		}
	}()
	defer b.Stop()

	// This is for recovering from panic.
	defer func() {
		r := recover()
		if r != nil {
			sentry.CaptureException(r.(error))
			log.Println(r.(error))
		}
	}()

	deps := cmd.New(cmd.Dependency{
		Memory:      cache,
		Bot:         b,
		DB:          db,
		Mongo:       mongoClient,
		MongoDBName: mongoDBName,
		TeknumID:    os.Getenv("TEKNUM_ID"),
	})

	httpServer := server.New(server.Config{
		DB:          db,
		Memory:      cache,
		Mongo:       mongoClient,
		MongoDBName: mongoDBName,
		Port:        os.Getenv("PORT"),
	})

	// This is basically just for health check.
	b.Handle("/start", func(c tb.Context) error {
		if c.Message().FromGroup() {
			_, err := c.Bot().Send(c.Message().Chat, "ok")
			if err != nil {
				shared.HandleBotError(ctx, err, b, c.Message())
				return nil
			}
		}

		return nil
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

	// Under attack handlers
	b.Handle("/underattack", deps.EnableUnderAttackModeHandler)
	b.Handle("/disableunderattack", deps.DisableUnderAttackModeHandler)

	// Badword handlers
	b.Handle("/badwords", deps.BadWordHandler)
	b.Handle("/cukup", deps.CukupHandler)

	// <redacted>
	b.Handle("/setir", deps.SetirHandler)

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt)

	go func() {
		// Start a HTTP server instance
		log.Printf("Starting http server on %s", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			log.Printf("%s", err.Error())
		}
	}()

	go func() {
		<-exitSignal
		log.Println("Shutdown signal received, exiting...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
		defer shutdownCancel()

		err = httpServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Printf("Shutting down HTTP server: %s", err.Error())
			sentry.CaptureException(err)
		}
	}()

	// Lesson learned: do not start bot on a goroutine
	log.Println("Bot started!")
	b.Start()
}
