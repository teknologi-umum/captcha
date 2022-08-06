package badwords_test

import (
	"context"
	"log"
	"os"
	"teknologi-umum-bot/badwords"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client

func TestMain(m *testing.M) {
	Setup()

	exitCode := m.Run()

	Teardown()
	Cleanup()

	os.Exit(exitCode)
}

func Setup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal(err)
	}

	if err = mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
}

func Cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := mongoClient.Database(os.Getenv("MONGO_DBNAME")).Collection("dukun")
	err := collection.Drop(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func Teardown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := mongoClient.Disconnect(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func TestAuthenticate(t *testing.T) {
	deps := &badwords.Dependency{
		Mongo:       mongoClient,
		MongoDBName: os.Getenv("MONGO_DBNAME"),
	}

	err := os.Setenv("ADMIN_ID", "30,40,50,60")
	if err != nil {
		t.Fatal(err)
	}

	ok := deps.Authenticate("30")
	if !ok {
		t.Errorf("failed to authenticate with the value of %d", 30)
	}

	notOK := deps.Authenticate("10")
	if notOK {
		t.Errorf("should be ok, but success with the value of %d", 10)
	}

	os.Clearenv()
}

func TestAddBadWord(t *testing.T) {
	deps := &badwords.Dependency{
		Mongo:       mongoClient,
		MongoDBName: os.Getenv("MONGO_DBNAME"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := deps.AddBadWord(ctx, "some bad word")
	if err != nil {
		t.Errorf("error during inserting bad word into mongo: %v", err)
	}
}
