package analytics_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"teknologi-umum-bot/analytics"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var db *sqlx.DB
var memory *bigcache.BigCache

func TestMain(m *testing.M) {
	Setup()

	defer Teardown()

	os.Exit(m.Run())
}

func Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := db.Connx(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE analytics")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE analytics_hourly")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	err = memory.Reset()
	if err != nil {
		log.Fatal(err)
	}
}

func Setup() {
	dbURL, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	db, err = sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	memory, err = bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 1))
	if err != nil {
		log.Fatal(err)
	}

	err = analytics.MustMigrate(db)
	if err != nil {
		log.Fatal(err)
	}
}

func Teardown() {
	memory.Close()
	db.Close()
}
