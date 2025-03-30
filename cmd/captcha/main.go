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
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/teknologi-umum/captcha/ascii"
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

	// Others third party stuff
	"github.com/getsentry/sentry-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pkg/errors"
	slogmulti "github.com/samber/slog-multi"
	"github.com/teknologi-umum/captcha/internal/requestid"
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
		slog.Error("Parsing configuration", slog.String("error", err.Error()))
		return
	}

	slogSentryBreadcrumb := &SlogSentryBreadcrumb{
		Enable: configuration.SentryDSN != "",
		Level:  parseSlogLevel(configuration.LogLevel),
	}
	slog.SetDefault(slog.New(
		slogmulti.Pipe(
			slogmulti.NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
				clonedRecord := record.Clone()
				reqId := requestid.GetRequestIdFromContext(ctx)
				if reqId != "" {
					clonedRecord.AddAttrs(slog.String("request_id", reqId))
				}
				next(ctx, clonedRecord)
				return nil
			}),
		).
			Handler(slogmulti.Fanout(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level: parseSlogLevel(configuration.LogLevel),
				}),
				slogSentryBreadcrumb,
			)),
	))

	slogWrapper := &slogWriterWrapper{logger: slog.Default()}

	// Setup Sentry for error handling.
	err = sentry.Init(sentry.ClientOptions{
		Dsn:           configuration.SentryDSN,
		Debug:         configuration.Environment == "development",
		DebugWriter:   slogWrapper,
		Environment:   configuration.Environment,
		SampleRate:    configuration.SentryConfig.SentrySampleRate,
		EnableTracing: true,
		TracesSampler: func(ctx sentry.SamplingContext) float64 {
			if ctx.Span.Name == "GET /" || ctx.Span.Name == "POST /bot[Filtered]/getUpdates" || ctx.Span.Name == "POST /bot[Filtered]/getMe" {
				return 0
			}

			return configuration.SentryConfig.SentryTracesSampleRate
		},
		Release: version,
	})
	if err != nil {
		slog.Error("initializing new sentry client", slog.String("error", err.Error()))
		os.Exit(1)
		return
	}
	defer sentry.Flush(10 * time.Second)

	var db *sqlx.DB
	if configuration.FeatureFlag.Analytics || (configuration.FeatureFlag.UnderAttack && configuration.UnderAttack.DatastoreProvider == "postgres") {
		// Connect to PostgreSQL
		db, err = sqlx.Open("postgres", configuration.Database.PostgresUrl)
		if err != nil {
			slog.Error("opening a postgres client", slog.String("error", err.Error()))
			os.Exit(1)
			return
		}
	}
	defer func(db *sqlx.DB) {
		if db != nil {
			slog.Debug("Closing postgresql")
			err := db.Close()
			if err != nil {
				slog.Warn("closing the postgres client", slog.String("error", err.Error()))
			}
		}
	}(db)

	// Context for initiating database connection.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())

	// Setup in memory cache
	cache, err := bigcache.New(context.Background(), bigcache.Config{
		Shards:             1024,
		LifeWindow:         time.Hour * 12,
		CleanWindow:        time.Hour,
		Verbose:            configuration.Environment != "production",
		Logger:             slogWrapper,
		HardMaxCacheSize:   1024 * 1024 * 1024,
		MaxEntrySize:       500,
		MaxEntriesInWindow: 50,
	})
	if err != nil {
		slog.Error("creating a in memory cache", slog.String("error", err.Error()))
		os.Exit(1)
		return
	}
	defer func(cache *bigcache.BigCache) {
		slog.Debug("Closing bigcache")
		err := cache.Close()
		if err != nil {
			slog.Warn("closing the bigcache", slog.String("error", err.Error()))
			os.Exit(1)
			return
		}
	}(cache)

	fileStorage, err := badger.Open(badger.DefaultOptions(configuration.Database.BadgerPath))
	if err != nil {
		slog.Error("creating badger db", slog.String("error", err.Error()))
		os.Exit(1)
		return
	}
	defer func(db *badger.DB) {
		slog.Debug("Closing badger")
		err := fileStorage.Close()
		if err != nil {
			slog.Warn("closing the badger db", slog.String("error", err.Error()))
			os.Exit(1)
			return
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
		slog.Error("initializing bot client", slog.String("error", err.Error()))
		os.Exit(1)
		return
	}
	defer func() {
		slog.Debug("Closing bot client")
		b.Stop()

		_, err := b.Close(context.Background())
		if err != nil {
			sentry.CaptureException(err)
			slog.Warn("closing the bot client", slog.String("error", err.Error()))
		}
	}()

	// This is for recovering from panic.
	defer func() {
		r := recover()
		if r != nil {
			sentry.CaptureException(r.(error))
			slog.ErrorContext(ctx, "recovering from panic", slog.String("error", err.Error()))
		}
	}()

	var underAttackDependency *underattack.Dependency
	if configuration.FeatureFlag.UnderAttack {
		var underAttackDatastore underattack.Datastore
		switch configuration.UnderAttack.DatastoreProvider {
		case "postgres":
			underAttackDatastore, err = datastore.NewPostgresDatastore(db.DB)
			if err != nil {
				slog.ErrorContext(ctx, "creating postgres datastore for under attack feature", slog.String("error", err.Error()))
				os.Exit(1)
				return
			}
		case "memory":
			fallthrough
		default:
			underAttackDatastore, err = datastore.NewInMemoryDatastore(cache)
			if err != nil {
				slog.ErrorContext(ctx, "creating in memory datastore for under attack feature", slog.String("error", err.Error()))
				os.Exit(1)
				return
			}
		}

		// Migrate datastore while we're here
		err = underAttackDatastore.Migrate(context.Background())
		if err != nil {
			sentry.CaptureException(err)
			slog.ErrorContext(ctx, "migrating under attack tables schema", slog.String("error", err.Error()))
			os.Exit(1)
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
		slog.ErrorContext(ctx, "creating setir dependency", slog.String("error", err.Error()))
		os.Exit(1)
		return
	}

	var reminderDependency *reminder.Dependency
	if configuration.FeatureFlag.Reminder {
		reminderDependency, err = reminder.New(cache)
		if err != nil {
			sentry.CaptureException(err)
			slog.ErrorContext(ctx, "creating reminder dependency", slog.String("error", err.Error()))
			os.Exit(1)
			return
		}
	}

	var deletionDependency *deletion.Dependency
	if configuration.FeatureFlag.Deletion {
		deletionDependency, err = deletion.New()
		if err != nil {
			sentry.CaptureException(err)
			slog.ErrorContext(ctx, "creating deletion dependency", slog.String("error", err.Error()))
			os.Exit(1)
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
		UnderAttack: underAttackDependency,
		Setir:       setirDependency,
		Reminder:    reminderDependency,
		Deletion:    deletionDependency,
	})
	if err != nil {
		sentry.CaptureException(err)
		slog.ErrorContext(ctx, "creating main program", slog.String("error", err.Error()))
		os.Exit(1)
		return
	}

	var httpServer *http.Server
	if configuration.FeatureFlag.HttpServer {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		httpServer = &http.Server{
			Addr:              net.JoinHostPort(configuration.HTTPServer.ListeningHost, configuration.HTTPServer.ListeningPort),
			Handler:           h,
			ReadTimeout:       time.Minute,
			ReadHeaderTimeout: time.Minute,
			WriteTimeout:      time.Minute,
			IdleTimeout:       time.Minute,
		}
	}

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

	// <redacted>
	b.Handle("/setir", program.SetirHandler)

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, os.Interrupt)

	go func() {
		if httpServer == nil {
			return
		}

		// Start an HTTP server instance
		slog.InfoContext(ctx, "Starting http server", slog.String("address", httpServer.Addr))
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.ErrorContext(ctx, "Listening HTTP server", slog.String("error", err.Error()))
			sentry.CaptureException(err)
		}
	}()

	go func() {
		<-exitSignal
		slog.InfoContext(ctx, "Shutdown signal received, exiting...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
		defer shutdownCancel()

		b.Stop()

		if httpServer != nil {
			err := httpServer.Shutdown(shutdownCtx)
			if err != nil {
				slog.ErrorContext(ctx, "Shutting down the http server", slog.String("error", err.Error()))
				sentry.CaptureException(err)
			}
		}

		slog.InfoContext(ctx, "Starting a 10 second countdown until killing the application")
		time.Sleep(time.Second * 10)
		os.Exit(0)
	}()

	go func() {
		// Run the cleanup worker
		program.Captcha.Cleanup()
	}()

	// Lesson learned: do not start bot on a goroutine
	slog.InfoContext(ctx, "Bot started!")
	b.Start()
	slog.InfoContext(ctx, "Application is shutting down...")
}
