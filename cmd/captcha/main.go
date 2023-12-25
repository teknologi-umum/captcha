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
	"github.com/teknologi-umum/captcha/ascii"
	"github.com/teknologi-umum/captcha/badwords"
	"github.com/teknologi-umum/captcha/captcha"
	"github.com/teknologi-umum/captcha/setir"
	"github.com/teknologi-umum/captcha/underattack"
	"github.com/teknologi-umum/captcha/underattack/datastore"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	// Internals
	"github.com/teknologi-umum/captcha/analytics"
	"github.com/teknologi-umum/captcha/analytics/server"
	"github.com/teknologi-umum/captcha/shared"
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

var version string

func main() {
	configuration, err := ParseConfiguration("")
	if err != nil {
		log.Fatalf("Parsing configuration: %s", err.Error())
		return
	}

	// Setup Sentry for error handling.
	err = sentry.Init(sentry.ClientOptions{
		Dsn:                configuration.SentryDSN,
		Debug:              configuration.Environment == "development",
		Environment:        configuration.Environment,
		SampleRate:         1.0,
		EnableTracing:      true,
		TracesSampleRate:   0.2,
		ProfilesSampleRate: 0.05,
		Release:            version,
	})
	if err != nil {
		log.Fatal("during initiating a new sentry client:", errors.WithStack(err))
	}
	defer sentry.Flush(30 * time.Second)

	var db *sqlx.DB
	if configuration.FeatureFlag.Analytics || (configuration.FeatureFlag.UnderAttack && configuration.UnderAttack.DatastoreProvider == "postgres") {
		// Connect to PostgreSQL
		db, err = sqlx.Open("postgres", configuration.Database.PostgresUrl)
		if err != nil {
			log.Fatal("during opening a postgres client:", errors.WithStack(err))
		}
	}
	defer func(db *sqlx.DB) {
		if db != nil {
			err := db.Close()
			if err != nil {
				log.Print("during closing the postgres client:", errors.WithStack(err))
			}
		}
	}(db)

	// Context for initiating database connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	var mongoClient *mongo.Client
	var mongoDBName string
	if configuration.FeatureFlag.BadwordsInsertion || configuration.FeatureFlag.Dukun {
		// Setup mongodb connection
		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(configuration.Database.MongoUrl))
		if err != nil {
			log.Fatal("during connecting to mongo client:", errors.WithStack(err))
		}

		// Mongo health check
		if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
			log.Fatal("during mongodb ping:", err)
		}

		// Get the MongoDB database name from the given MONGO_URL environment variable.
		parsedURL, err := url.Parse(configuration.Database.MongoUrl)
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to parse MONGO_URL"))
		}
		mongoDBName = parsedURL.Path[1:]
	}
	defer func(client *mongo.Client) {
		if client != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			defer cancel()
			err := client.Disconnect(ctx)
			if err != nil {
				log.Print("during closing the mongo connection:", errors.WithStack(err))
			}
		}
	}(mongoClient)

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
		log.Fatal("during creating a in memory cache: ", errors.WithStack(err))
	}
	defer func(cache *bigcache.BigCache) {
		err := cache.Close()
		if err != nil {
			log.Print(errors.WithStack(err))
		}
	}(cache)

	// Setup Telegram Bot
	b, err := tb.NewBot(tb.Settings{
		Token:       configuration.BotToken,
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

	var analyticsDependency *analytics.Dependency
	if configuration.FeatureFlag.Analytics {
		// Check if database is initialized
		if db == nil {
			log.Println("To enable analytics, database must been set")
			return
		}
		err = analytics.MustMigrate(db)
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal("during initial database migration: ", errors.WithStack(err))
		}

		analyticsDependency = &analytics.Dependency{
			Memory:      cache,
			Bot:         b,
			DB:          db,
			HomeGroupID: configuration.HomeGroupID,
		}
	}

	var badwordsDependency *badwords.Dependency
	if configuration.FeatureFlag.BadwordsInsertion {
		// Check if mongodb is initialized
		if mongoClient == nil && mongoDBName == "" {
			log.Println("To enable badwords insertion, mongodb mnust been set")
			return
		}

		badwordsDependency = &badwords.Dependency{
			Mongo:       mongoClient,
			MongoDBName: mongoDBName,
			AdminIDs:    configuration.AdminIds,
		}
	}

	var underAttackDependency *underattack.Dependency
	if configuration.FeatureFlag.UnderAttack {
		var underAttackDatastore underattack.Datastore
		switch configuration.UnderAttack.DatastoreProvider {
		case "postgres":
			underAttackDatastore, err = datastore.NewPostgresDatastore(db.DB)
			if err != nil {
				log.Fatal(err)
				return
			}
		case "memory":
			fallthrough
		default:
			underAttackDatastore, err = datastore.NewInMemoryDatastore(cache)
			if err != nil {
				log.Fatal(err)
				return
			}
		}

		// Migrate datastore while we're here
		err = underAttackDatastore.Migrate(context.Background())
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal(err)
			return
		}

		underAttackDependency = &underattack.Dependency{
			Datastore: underAttackDatastore,
			Memory:    cache,
			Bot:       b,
		}
	}

	var setirDependency *setir.Dependency
	setirDependency, err = setir.New(b, configuration.AdminIds, configuration.HomeGroupID)
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
		return
	}

	program, err := New(Dependency{
		FeatureFlag: configuration.FeatureFlag,
		Captcha: &captcha.Dependencies{
			Memory:        cache,
			Bot:           b,
			TeknumGroupID: configuration.HomeGroupID,
		},
		Ascii:       &ascii.Dependencies{Bot: b},
		Analytics:   analyticsDependency,
		Badwords:    badwordsDependency,
		UnderAttack: underAttackDependency,
		Setir:       setirDependency,
	})
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal(err)
		return
	}

	httpServer := server.New(server.Config{
		DB:               db,
		Memory:           cache,
		Mongo:            mongoClient,
		MongoDBName:      mongoDBName,
		ListeningAddress: net.JoinHostPort(configuration.HTTPServer.ListeningHost, configuration.HTTPServer.ListeningPort),
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
	b.Handle(tb.OnUserJoined, program.OnUserJoinHandler)
	b.Handle(tb.OnText, program.OnTextHandler)
	b.Handle(tb.OnPhoto, program.OnNonTextHandler)
	b.Handle(tb.OnAnimation, program.OnNonTextHandler)
	b.Handle(tb.OnVideo, program.OnNonTextHandler)
	b.Handle(tb.OnDocument, program.OnNonTextHandler)
	b.Handle(tb.OnSticker, program.OnNonTextHandler)
	b.Handle(tb.OnVoice, program.OnNonTextHandler)
	b.Handle(tb.OnVideoNote, program.OnNonTextHandler)
	b.Handle(tb.OnUserLeft, program.OnUserLeftHandler)

	// Under attack handlers
	b.Handle("/underattack", program.EnableUnderAttackModeHandler)
	b.Handle("/disableunderattack", program.DisableUnderAttackModeHandler)

	// Bad word handlers
	b.Handle("/badwords", program.BadWordHandler)

	// <redacted>
	b.Handle("/setir", program.SetirHandler)

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt)

	go func() {
		// Start an HTTP server instance
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
