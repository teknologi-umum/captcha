package server

import (
	"log"
	"net/http"
	"os"
	"teknologi-umum-bot/analytics"

	"github.com/allegro/bigcache/v3"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

type Dependency struct {
	DB     *sqlx.DB
	Memory *bigcache.BigCache
}

type User = analytics.UserMap

func Server(db *sqlx.DB, memory *bigcache.BigCache) {
	deps := &Dependency{DB: db}

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
		data, err := deps.GetAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	r.Get("/total", func(rw http.ResponseWriter, r *http.Request) {
		data, err := deps.GetTotal(r.Context())
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	})

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", r)
}
