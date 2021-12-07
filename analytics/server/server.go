package server

import (
	"errors"
	"log"
	"net/http"
	"os"
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/shared"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
	"go.mongodb.org/mongo-driver/mongo"
)

// Dependency specifies the dependency injection struct
// for the server package to use.
type Dependency struct {
	DB     *sqlx.DB
	Memory *bigcache.BigCache
	Logger *sentry.Client
	Mongo  *mongo.Database
}

// User is a type alias for analytics.UserMap and should be
// similar for the behavior and other stuffs.
type User = analytics.UserMap

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
)

var ErrInvalidValue = errors.New("invalid value")

// Server creates and runs an HTTP server instance for fetching analytics data
// that can be used later by other third party sites or bots.
//
// Requires 3 parameter that should be sent from the main goroutine.
func Server(db *sqlx.DB, mongoDB *mongo.Database, memory *bigcache.BigCache, logger *sentry.Client) {
	deps := &Dependency{
		DB:     db,
		Memory: memory,
		Logger: logger,
		Mongo:  mongoDB,
	}

	secureMiddleware := secure.New(secure.Options{
		BrowserXssFilter:   true,
		ContentTypeNosniff: true,
		SSLRedirect:        os.Getenv("ENV") == "production",
		IsDevelopment:      os.Getenv("ENV") == "development",
	})
	corsMiddleware := cors.New(cors.Options{
		Debug:          os.Getenv("ENV") == "development",
		AllowedOrigins: []string{},
		AllowedMethods: []string{"GET", "OPTIONS"},
	})

	r := chi.NewRouter()
	r.Use(secureMiddleware.Handler)
	r.Use(corsMiddleware.Handler)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Available routes:\n\n- GET /users\n- GET /hourly\n- GET /total"))
		if err != nil {
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}
	})

	r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		data, err := deps.GetAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}

		lastUpdated, err := deps.LastUpdated(UserEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Last-Updated", lastUpdated.String())
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}
	})

	r.Get("/hourly", func(w http.ResponseWriter, r *http.Request) {
		data, err := deps.GetHourly(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}

		lastUpdated, err := deps.LastUpdated(HourlyEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Last-Updated", lastUpdated.String())
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}
	})

	r.Get("/total", func(w http.ResponseWriter, r *http.Request) {
		data, err := deps.GetTotal(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}

		lastUpdated, err := deps.LastUpdated(TotalEndpoint)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		h.Set("Last-Updated", lastUpdated.String())
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			shared.HandleHttpError(err, deps.Logger, r)
			return
		}
	})

	log.Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		shared.HandleError(err, logger)
	}
}