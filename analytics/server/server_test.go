package server_test

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"teknologi-umum-bot/analytics"
	"teknologi-umum-bot/analytics/server"
	"teknologi-umum-bot/dukun"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var dependency *server.Dependency

func TestMain(m *testing.M) {
	mongoUrl, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		mongoUrl = "mongodb://root:password@localhost:27017"
	}

	mongoDbName, ok := os.LookupEnv("MONGO_DBNAME")
	if !ok {
		mongoDbName = "captcha"
	}

	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		databaseUrl = "postgresql://postgres:password@localhost:5432/captcha?sslmode=disable"
	}

	setupCtx, setupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setupCancel()

	dbURL, err := pq.ParseURL(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	mongoClient, err := mongo.Connect(setupCtx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Fatal(err)
	}

	if err = mongoClient.Ping(setupCtx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	memory, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Hour * 1))
	if err != nil {
		log.Fatal(err)
	}

	err = analytics.MustMigrate(db)
	if err != nil {
		log.Fatal(err)
	}

	dependency = &server.Dependency{
		DB:          db,
		Memory:      memory,
		Mongo:       mongoClient,
		MongoDBName: mongoDbName,
	}

	err = Seed(setupCtx)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cleanupCancel()

	Cleanup()

	err = Teardown(cleanupCtx)
	if err != nil {
		log.Print(err)
	}

	err = memory.Close()
	if err != nil {
		log.Print(err)
	}
	err = db.Close()
	if err != nil {
		log.Print(err)
	}

	err = mongoClient.Disconnect(cleanupCtx)
	if err != nil {
		log.Print(err)
	}

	os.Exit(exitCode)
}

func Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := dependency.DB.Connx(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			log.Fatal(err)
		}
	}(c)

	tx, err := c.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
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

	collection := dependency.Mongo.Database(dependency.MongoDBName).Collection("dukun")
	err = collection.Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = dependency.Memory.Reset()
	if err != nil {
		log.Fatal(err)
	}
}

func Teardown(cleanupCtx context.Context) error {
	c, err := dependency.DB.Connx(cleanupCtx)
	if err != nil {
		return err
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			log.Print(err)
		}
	}(c)

	tx, err := c.BeginTxx(cleanupCtx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: false})
	if err != nil {
		return err
	}

	queries := []string{
		"DROP INDEX IF EXISTS idx_counter",
		"DROP INDEX IF EXISTS idx_active",
		"DROP TABLE IF EXISTS captcha_swarm",
		"DROP TABLE IF EXISTS analytics",
		"DROP TABLE IF EXISTS analytics_hourly",
	}

	for _, query := range queries {
		_, err = tx.ExecContext(cleanupCtx, query)
		if err != nil {
			if r := tx.Rollback(); r != nil {
				return r
			}

			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	err = dependency.Memory.Reset()
	if err != nil {
		return err
	}

	return nil
}

func Seed(ctx context.Context) error {
	// create a dummy user struct slice
	users := []server.User{
		{UserID: 500, GroupID: analytics.NullInt64{Int64: 123456, Valid: true}, Username: "user5", DisplayName: "User 5", Counter: 1},
		{UserID: 600, GroupID: analytics.NullInt64{Int64: 123456, Valid: true}, Username: "user6", DisplayName: "User 6", Counter: 2},
		{UserID: 700, GroupID: analytics.NullInt64{Int64: 123456, Valid: true}, Username: "user7", DisplayName: "User 7", Counter: 3},
	}

	// convert users slice to single slice with no keys, just values.
	var usersSlice [][]interface{}
	for _, v := range users {
		usersSlice = append(
			usersSlice,
			[]interface{}{
				v.UserID,
				v.GroupID,
				v.Username,
				v.DisplayName,
				v.Counter,
				time.Now(),
				time.Now(),
				time.Now(),
			},
		)
	}

	// create a dummy hourly type
	hourly := []server.Hourly{
		{
			TodaysDate: "2021-01-01",
			ZeroHour:   14,
			OneHour:    15,
			TwoHour:    16,
		},
		{
			TodaysDate: "2021-01-02",
			ZeroHour:   3,
			OneHour:    4,
			TwoHour:    5,
		},
		{
			TodaysDate: "2021-01-03",
			ZeroHour:   6,
			OneHour:    7,
			TwoHour:    8,
		},
	}

	// convert hourly slice to a single interface{} slice
	var hourlySlice [][]interface{}
	for _, v := range hourly {
		hourlySlice = append(
			hourlySlice,
			[]interface{}{
				v.TodaysDate,
				v.ZeroHour,
				v.OneHour,
				v.TwoHour,
			},
		)
	}

	c, err := dependency.DB.Connx(ctx)
	if err != nil {
		return err
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil {
			log.Print(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
	if err != nil {

		return err
	}

	for _, user := range usersSlice {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO analytics
				(user_id, group_id, username, display_name, counter, created_at, joined_at, updated_at)
				VALUES
				($1, $2, $3, $4, $5, $6, $7, $8)`,
			user...,
		)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return e
			}

			return err
		}
	}

	for _, hourly := range hourlySlice {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO analytics_hourly
				(todays_date, zero_hour, one_hour, two_hour)
				VALUES
				($1, $2, $3, $4)`,
			hourly...,
		)
		if err != nil {
			if e := tx.Rollback(); e != nil {
				return e
			}

			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	// Feed some dukun
	collection := dependency.Mongo.Database(dependency.MongoDBName).Collection("dukun")
	_, err = collection.InsertOne(ctx, dukun.Dukun{
		UserID:    1,
		FirstName: "Jason",
		LastName:  "Bourne",
		UserName:  "jasonbourne",
		Points:    100,
		Master:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	now := time.Date(2020, 8, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	err = dependency.Memory.Set("analytics:last_updated:users", []byte(now))
	if err != nil {
		return err
	}

	err = dependency.Memory.Set("analytics:last_updated:hourly", []byte(now))
	if err != nil {
		return err
	}

	err = dependency.Memory.Set("analytics:last_updated:total", []byte(now))
	if err != nil {
		return err
	}

	err = dependency.Memory.Set("analytics:last_updated:dukun", []byte(now))
	if err != nil {
		return err
	}

	return nil
}
