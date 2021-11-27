package analytics_test

import (
	"log"
	"os"
	"teknologi-umum-bot/analytics"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var DB *sqlx.DB
var Redis *redis.Client

func TestMain(m *testing.M) {
	dbURL, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	DB, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	redisURL, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal(err)
	}

	Redis = redis.NewClient(redisURL)
	defer Redis.Close()

	err = analytics.MustMigrate(DB)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}
