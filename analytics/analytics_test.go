package analytics_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/teknologi-umum/captcha/analytics"

	"github.com/allegro/bigcache/v3"
	"github.com/getsentry/sentry-go"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var dependency *analytics.Dependency

func TestMain(m *testing.M) {
	databaseUrl, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		databaseUrl = "postgresql://postgres:password@localhost:5432/captcha?sslmode=disable"
	}

	dbURL, err := pq.ParseURL(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	_ = sentry.Init(sentry.ClientOptions{})

	memory, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Hour*1))
	if err != nil {
		log.Fatal(err)
	}

	err = analytics.MustMigrate(db)
	if err != nil {
		log.Fatal(err)
	}

	dependency = &analytics.Dependency{
		DB:       db,
		Memory:   memory,
		TeknumID: "123456789",
	}

	setupCtx, setupCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer setupCancel()

	err = Seed(setupCtx)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	Cleanup()

	err = Teardown()
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

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE analytics RESTART IDENTITY CASCADE")
	if err != nil {
		if r := tx.Rollback(); r != nil {
			log.Fatal(r)
		}
		log.Fatal(err)
	}

	_, err = tx.ExecContext(ctx, "TRUNCATE TABLE analytics_hourly RESTART IDENTITY CASCADE")
	if err != nil {
		if r := tx.Rollback(); r != nil {
			log.Fatal(r)
		}
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		if r := tx.Rollback(); r != nil {
			log.Fatal(r)
		}
		log.Fatal(err)
	}

	err = dependency.Memory.Reset()
	if err != nil {
		log.Fatal(err)
	}
}

func Teardown() error {
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cleanupCancel()

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

	tx, err := c.BeginTxx(cleanupCtx, &sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: false})
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
		if r := tx.Rollback(); r != nil {
			return r
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
	c, err := dependency.DB.Connx(ctx)
	if err != nil {
		return err
	}
	defer func(c *sqlx.Conn) {
		err := c.Close()
		if err != nil && !errors.Is(err, sql.ErrConnDone) {
			log.Print(err)
		}
	}(c)

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	// The lack of group_id value is intentional, because I want to check for
	// null SQL values.
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO analytics
			(user_id, username, display_name, counter, created_at, joined_at, updated_at)
			VALUES
			($1, $2, $3, $4, $5, $6, $7)`,
		90,
		"user1",
		"User 1",
		1,
		time.Now(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	// create a dummy hourly type
	hourly := []analytics.HourlyMap{
		{
			TodaysDate: time.Now().Add(time.Hour * -48).Format("2006-01-02"),
			ZeroHour:   14,
			OneHour:    15,
			TwoHour:    16,
		},
		{
			TodaysDate: time.Now().Add(time.Hour * -24).Format("2006-01-02"),
			ZeroHour:   3,
			OneHour:    4,
			TwoHour:    5,
		},
		{
			TodaysDate: time.Now().Format("2006-01-02"),
			ZeroHour:   6,
			OneHour:    7,
			TwoHour:    8,
		},
	}

	for _, hour := range hourly {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO analytics_hourly
				(todays_date, zero_hour, one_hour, two_hour)
				VALUES
				($1, $2, $3, $4)`,
			hour.TodaysDate,
			hour.ZeroHour,
			hour.OneHour,
			hour.TwoHour,
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
		return err
	}

	return nil
}
