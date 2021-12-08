package server_test

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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var db *sqlx.DB
var memory *bigcache.BigCache
var mongoClient *mongo.Client

func TestMain(m *testing.M) {
	Setup()

	defer Teardown()
	defer Cleanup()
	os.Exit(m.Run())
}

func Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := db.Connx(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE analytics RESTART IDENTITY")
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.Fatal(e)
		}
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE analytics_hourly RESTART IDENTITY")
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.Fatal(e)
		}
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.Fatal(e)
		}
		log.Fatal(err)
	}

	collection := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("dukun")
	err = collection.Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = memory.Reset()
	if err != nil {
		log.Fatal(err)
	}
}

func Setup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbURL, err := pq.ParseURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	db, err = sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal(err)
	}

	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
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
	defer func(memory *bigcache.BigCache) {
		err := memory.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(memory)
	defer func(db *sqlx.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	defer func(mongoClient *mongo.Client) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := mongoClient.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}(mongoClient)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := db.Connx(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "DROP TABLE IF EXISTS analytics")
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.Fatal(e)
		}
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "DROP TABLE IF EXISTS analytics_hourly")
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.Fatal(e)
		}
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			log.Fatal(e)
		}
		log.Fatal(err)
	}

	err = memory.Reset()
	if err != nil {
		log.Fatal(err)
	}
}
