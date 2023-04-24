package underattack_test

import (
	"context"
	"database/sql"
	"log"
	"os"
	"teknologi-umum-bot/underattack"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var dependency *underattack.Dependency

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

	memory, err := bigcache.New(context.Background(), bigcache.DefaultConfig(time.Hour*1))
	if err != nil {
		log.Fatal(err)
	}

	err = underattack.MustMigrate(db)
	if err != nil {
		log.Fatal(err)
	}

	dependency = &underattack.Dependency{
		Memory: memory,
		DB:     db,
		Bot:    nil,
		Logger: nil,
	}

	setupCtx, setupCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer setupCancel()

	err = Seed(setupCtx)
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

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

func Seed(ctx context.Context) error {
	c, err := dependency.DB.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err := c.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	tx, err := c.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO
			under_attack
			(group_id, is_under_attack, expires_at, notification_message_id, updated_at)
			VALUES
			($1, $2, $3, $4, $5)`,
		1,
		true,
		time.Now().Add(time.Hour*1),
		1002,
		time.Now(),
	)
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	err = tx.Commit()
	if err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}

		return err
	}

	return nil
}
