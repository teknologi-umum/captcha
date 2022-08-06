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

var dependency *badwords.Dependency

func TestMain(m *testing.M) {
	mongoUrl, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		mongoUrl = "mongodb://root:password@localhost:27017/"
	}

	dbName, ok := os.LookupEnv("MONGO_DBNAME")
	if !ok {
		dbName = "captcha"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoUrl))
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	dependency = &badwords.Dependency{
		Mongo:       db,
		MongoDBName: dbName,
	}

	exitCode := m.Run()

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cleanupCancel()

	collection := db.Database(dependency.MongoDBName).Collection("badstuff")
	err = collection.Drop(cleanupCtx)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Disconnect(cleanupCtx)
	if err != nil {
		log.Print(err)
	}

	os.Exit(exitCode)
}

func TestAuthenticate(t *testing.T) {
	err := os.Setenv("ADMIN_ID", "30,40,50,60")
	if err != nil {
		t.Fatal(err)
	}

	ok := dependency.Authenticate("30")
	if !ok {
		t.Errorf("failed to authenticate with the value of %d", 30)
	}

	notOK := dependency.Authenticate("10")
	if notOK {
		t.Errorf("should be ok, but success with the value of %d", 10)
	}

	os.Clearenv()
}

func TestAddBadWord(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := dependency.AddBadWord(ctx, "some bad word")
	if err != nil {
		t.Errorf("error during inserting bad word into mongo: %v", err)
	}
}
