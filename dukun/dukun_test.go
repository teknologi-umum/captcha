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

var dependency *dukun.Dependency

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

	dependency = &dukun.Dependency{
		Mongo:  db,
		DBName: dbName,
	}

	err = seed()
	if err != nil {
		log.Fatal(err)
	}

	exitCode := m.Run()

	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cleanupCancel()

	collection := db.Database(dependency.DBName).Collection("dukun")
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

func seed() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Feed some dukun
	collection := dependency.Mongo.Database(dependency.DBName).Collection("dukun")
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
		return err
	}

	return nil
}

func TestGetAllDukun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	data, err := dependency.GetAllDukun(ctx)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 1 {
		t.Error("Expected 1 dukun, got", len(data))
	}
}
