package server

import (
	"log"
	"net/http"
	"teknologi-umum-bot/analytics"

	"github.com/allegro/bigcache/v3"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type Dependency struct {
	DB *sqlx.DB
	Memory *bigcache.BigCache
}

type User = analytics.UserMap

func Server(db *sqlx.DB, memory *bigcache.BigCache) {
	deps := &Dependency{DB:db}

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := deps.GetAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	r.Get("/total", func(rw http.ResponseWriter, r *http.Request) {
		data, err := deps.GetTotal(r.Context())
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	})

	log.Println("Starting server on port 8080")
	http.ListenAndServe(":8080", r)
}
