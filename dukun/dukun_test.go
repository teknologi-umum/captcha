package dukun_test

import (
	"context"
	"log"
	"os"
	"teknologi-umum-bot/dukun"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var db *mongo.Client

func TestMain(m *testing.M) {
	Setup()

	defer Teardown()
	defer Cleanup()
	os.Exit(m.Run())
}

func Setup() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
}

func Teardown() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := db.Disconnect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := db.Database(os.Getenv("MONGO_DBNAME")).Collection("dukun")
	err := collection.Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func TestGetAllDukun(t *testing.T) {
	t.Cleanup(Cleanup)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Feed some dukun
	collection := db.Database(os.Getenv("MONGO_DBNAME")).Collection("dukun")
	_, err := collection.InsertOne(ctx, dukun.Dukun{
		UserID:    1,
		FirstName: "Jason",
		LastName:  "Bourne",
		UserName:  "jasonbourne",
		Points:    100,
		Master:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		t.Error(err)
	}

	deps := &dukun.Dependency{Mongo: db, DBName: os.Getenv("MONGO_DBNAME")}

	data, err := deps.GetAllDukun(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 1 {
		t.Error("Expected 1 dukun, got", len(data))
	}
}
