package underattack_test

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/teknologi-umum/captcha/underattack/datastore"
	"log"
	"os"
	"testing"
	"time"

	"github.com/teknologi-umum/captcha/underattack"

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

	_ = sentry.Init(sentry.ClientOptions{})

	memoryDatastore, err := datastore.NewInMemoryDatastore(memory)
	if err != nil {
		log.Fatal(err)
	}

	dependency = &underattack.Dependency{
		Memory:    memory,
		Datastore: memoryDatastore,
		Bot:       nil,
	}

	setupCtx, setupCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer setupCancel()

	err = dependency.Datastore.Migrate(setupCtx)
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
