package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/teknologi-umum/captcha/analytics"
	"github.com/teknologi-umum/captcha/shared"

	"github.com/allegro/bigcache/v3"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

// Dependency specifies the dependency injection struct
// for the server package to use.
type Dependency struct {
	DB     *sqlx.DB
	Memory *bigcache.BigCache
}

// User is a type alias for analytics.GroupMember and should be
// similar for the behavior and other stuffs.
type User = analytics.GroupMember

// Hourly is a type alias for analytics.HourlyMap
type Hourly = analytics.HourlyMap

// Endpoint specifies a type to be used as enum.
type Endpoint int

const (
	// UserEndpoint indicates the endpoint for getting users data
	UserEndpoint Endpoint = iota
	// HourlyEndpoint indicates the endpoint for getting the data per hour
	HourlyEndpoint
	// TotalEndpoint indicates the endpoint for getting the total amount
	// of messages that was sent per the database's data.
	TotalEndpoint
	// DukunEndpoint indicates the endpoint for getting the whole
	// dukun points as used by the Javascript bot.
	DukunEndpoint
)

var ErrInvalidValue = errors.New("invalid value")

// Config is the configuration struct for the server package.
// Only the Port field is optional. It will be set to 8080 if not set.
type Config struct {
	DB               *sqlx.DB
	Memory           *bigcache.BigCache
	Environment      string
	ListeningAddress string
}

// New creates and runs an HTTP server instance for fetching analytics data
// that can be used later by other third party sites or bots.
//
// Requires 3 parameter that should be sent from the main goroutine.
func New(config Config) *http.Server {
	// Give default port
	if config.ListeningAddress == "" {
		config.ListeningAddress = ":8080"
	}

	deps := &Dependency{
		DB:     config.DB,
		Memory: config.Memory,
	}

	secureMiddleware := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		SSLRedirect:        config.Environment == "production",
		IsDevelopment:      config.Environment == "development",
	})
	corsMiddleware := cors.New(cors.Options{
		Debug:          config.Environment == "development",
		AllowedOrigins: []string{},
		AllowedMethods: []string{"GET", "OPTIONS"},
	})

	sentryMiddleware := sentryhttp.New(sentryhttp.Options{Repanic: false})

	r := chi.NewRouter()
	r.Use(sentryMiddleware.Handle)
	r.Use(secureMiddleware.Handler)
	r.Use(corsMiddleware.Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Available routes:\n\n- GET /users\n- GET /hourly\n- GET /total"))
		if err != nil {
			shared.HandleHttpError(r.Context(), err, r)
			return
		}
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := deps.GetAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(r.Context(), err, r)
			return
		}

		lastUpdated, err := deps.LastUpdated(UserEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(r.Context(), err, r)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Last-Updated", lastUpdated.String())
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			shared.HandleHttpError(r.Context(), err, r)
			return
		}
	})

	r.Get("/hourly", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := deps.GetHourly(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(r.Context(), err, r)
			return
		}

		lastUpdated, err := deps.LastUpdated(HourlyEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(r.Context(), err, r)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Last-Updated", lastUpdated.String())
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			shared.HandleHttpError(r.Context(), err, r)
			return
		}
	})

	r.Get("/total", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := deps.GetTotal(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(r.Context(), err, r)
			return
		}

		lastUpdated, err := deps.LastUpdated(TotalEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(r.Context(), err, r)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Last-Updated", lastUpdated.String())
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			shared.HandleHttpError(r.Context(), err, r)
			return
		}
	})

	return &http.Server{
		Handler:           r,
		Addr:              config.ListeningAddress,
		ReadTimeout:       time.Minute,
		WriteTimeout:      time.Minute,
		ReadHeaderTimeout: time.Minute,
		IdleTimeout:       time.Minute,
	}
}
