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
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/teknologi-umum/captcha/analytics"
	"github.com/teknologi-umum/captcha/analytics/server"
	"github.com/teknologi-umum/captcha/ascii"
	"github.com/teknologi-umum/captcha/badwords"
	"github.com/teknologi-umum/captcha/captcha"
	"github.com/teknologi-umum/captcha/deletion"
	"github.com/teknologi-umum/captcha/reminder"
	"github.com/teknologi-umum/captcha/setir"
	"github.com/teknologi-umum/captcha/shared"
	"github.com/teknologi-umum/captcha/underattack"
	"github.com/teknologi-umum/captcha/underattack/datastore"

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
	"github.com/rs/zerolog/log"
	tb "github.com/teknologi-umum/captcha/internal/telebot"
)

var version string

func main() {
	var configurationFilePath string
	flag.StringVar(&configurationFilePath, "configuration-file", "", "Path to configuration file")
	flag.Parse()
	if value, ok := os.LookupEnv("CONFIGURATION_FILE"); ok {
		configurationFilePath = value
	}

	configuration, err := ParseConfiguration(configurationFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Parsing configuration")
		return
	}

	// Setup Sentry for error handling.
	err = sentry.Init(sentry.ClientOptions{
		Dsn:           configuration.SentryDSN,
		Debug:         configuration.Environment == "development",
		DebugWriter:   log.Logger,
		Environment:   configuration.Environment,
		SampleRate:    1.0,
		EnableTracing: true,
		TracesSampler: func(ctx sentry.SamplingContext) float64 {
			if ctx.Span.Name == "GET /" || ctx.Span.Name == "POST /bot[Filtered]/getUpdates" {
				return 0
			}

			return 0.2
		},
		ProfilesSampleRate: 0.05,
		Release:            version,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("during initiating a new sentry client:")
	}
	defer sentry.Flush(10 * time.Second)

	var db *sqlx.DB
	if configuration.FeatureFlag.Analytics || (configuration.FeatureFlag.UnderAttack && configuration.UnderAttack.DatastoreProvider == "postgres") {
		// Connect to PostgreSQL
		db, err = sqlx.Open("postgres", configuration.Database.PostgresUrl)
		if err != nil {
			log.Fatal().Err(err).Msg("during opening a postgres client")
		}
	}
	defer func(db *sqlx.DB) {
		if db != nil {
			log.Debug().Msg("Closing postgresql")
			err := db.Close()
			if err != nil {
				log.Warn().Err(err).Msg("during closing the postgres client")
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
			log.Fatal().Err(err).Msg("during connecting to mongo client")
		}

		// Mongo health check
		if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
			log.Fatal().Err(err).Msg("during mongodb ping")
		}

		// Get the MongoDB database name from the given MONGO_URL environment variable.
		parsedURL, err := url.Parse(configuration.Database.MongoUrl)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse MONGO_URL")
		}
		mongoDBName = parsedURL.Path[1:]
	}
	defer func(client *mongo.Client) {
		if client != nil {
			log.Debug().Msg("Closing mongo")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			defer cancel()
			err := client.Disconnect(ctx)
			if err != nil {
				log.Warn().Err(err).Msg("during closing the mongo connection")
			}
		}
	}(mongoClient)

	// Setup in memory cache
	cache, err := bigcache.New(context.Background(), bigcache.Config{
		Shards:             1024,
		LifeWindow:         time.Hour * 12,
		CleanWindow:        time.Hour,
		Verbose:            configuration.Environment != "production",
		Logger:             &log.Logger,
		HardMaxCacheSize:   1024 * 1024 * 1024,
		MaxEntrySize:       500,
		MaxEntriesInWindow: 50,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("during creating a in memory cache")
	}
	defer func(cache *bigcache.BigCache) {
		log.Debug().Msg("Closing bigcache")
		err := cache.Close()
		if err != nil {
			log.Print(errors.WithStack(err))
		}
	}(cache)

	fileStorage, err := badger.Open(badger.DefaultOptions(configuration.Database.BadgerPath))
	if err != nil {
		log.Fatal().Err(err).Msg("during creating badger db")
		return
	}
	defer func(db *badger.DB) {
		log.Debug().Msg("Closing badger")
		err := fileStorage.Close()
		if err != nil {
			log.Print(errors.WithStack(err))
		}
	}(fileStorage)

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
			Transport: &SentryTransportWrapper{
				OriginalTransport: &http.Transport{
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
		},
	})
	if err != nil {
		sentry.CaptureException(fmt.Errorf("initializing bot client: %w", err))
		log.Fatal().Err(err).Msg("during init of bot client")
	}
	defer func() {
		b.Stop()

		_, err := b.Close(context.Background())
		if err != nil {
			sentry.CaptureException(err)
			log.Print(errors.WithStack(err))
		}
	}()

	// This is for recovering from panic.
	defer func() {
		r := recover()
		if r != nil {
			sentry.CaptureException(r.(error))
			log.Error().Err(err).Msg("recovering from panic")
		}
	}()

	var analyticsDependency *analytics.Dependency
	if configuration.FeatureFlag.Analytics {
		// Check if database is initialized
		if db == nil {
			log.Print("To enable analytics, database must been set")
			return
		}
		err = analytics.MustMigrate(db)
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal().Err(err).Msg("during initial database migration")
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
			log.Print("To enable badwords insertion, mongodb mnust been set")
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
				log.Fatal().Err(err).Msg("creating postgres datastore for under attack feature")
				return
			}
		case "memory":
			fallthrough
		default:
			underAttackDatastore, err = datastore.NewInMemoryDatastore(cache)
			if err != nil {
				log.Fatal().Err(err).Msg("creating in memory datastore for under attack feature")
				return
			}
		}

		// Migrate datastore while we're here
		err = underAttackDatastore.Migrate(context.Background())
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal().Err(err).Msg("migrating under attack tables schema")
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
		log.Fatal().Err(err).Msg("creating setir dependency")
		return
	}

	var reminderDependency *reminder.Dependency
	if configuration.FeatureFlag.Reminder {
		reminderDependency, err = reminder.New(cache)
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal().Err(err).Msg("creating reminder dependency")
			return
		}
	}

	var deletionDependency *deletion.Dependency
	if configuration.FeatureFlag.Deletion {
		deletionDependency, err = deletion.New()
		if err != nil {
			sentry.CaptureException(err)
			log.Fatal().Err(err).Msg("creating deletion dependency")
			return
		}
	}

	program, err := New(Dependency{
		FeatureFlag: configuration.FeatureFlag,
		Captcha: &captcha.Dependencies{
			Memory:        cache,
			Bot:           b,
			TeknumGroupID: configuration.HomeGroupID,
			DB:            fileStorage,
		},
		Ascii:       &ascii.Dependencies{Bot: b},
		Analytics:   analyticsDependency,
		Badwords:    badwordsDependency,
		UnderAttack: underAttackDependency,
		Setir:       setirDependency,
		Reminder:    reminderDependency,
		Deletion:    deletionDependency,
	})
	if err != nil {
		sentry.CaptureException(err)
		log.Fatal().Err(err).Msg("creating main program")
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
			_, err := c.Bot().Send(context.Background(), c.Message().Chat, "ok")
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

	// Reminder (temporary feature)
	b.Handle("/remind", program.ReminderHandler)

	// Deletion (temporary feature)
	b.Handle("/delete", program.DeletionHandler)

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
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Listening HTTP server")
		}
	}()

	go func() {
		<-exitSignal
		log.Print("Shutdown signal received, exiting...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
		defer shutdownCancel()

		b.Stop()

		err = httpServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Error().Err(err).Msg("Shutting down the http server")
			sentry.CaptureException(err)
		}

		log.Debug().Msg("Starting a 10 second countdown until killing the application")
		time.Sleep(time.Second * 10)
		os.Exit(0)
	}()

	go func() {
		// Run the cleanup worker
		program.Captcha.Cleanup()
	}()

	// Lesson learned: do not start bot on a goroutine
	log.Print("Bot started!")
	b.Start()
	log.Print("Application is shutting down...")
}
