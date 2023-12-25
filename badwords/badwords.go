package badwords

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Dependency contains the dependency injection struct
// for the badwords package.
type Dependency struct {
	Mongo       *mongo.Client
	MongoDBName string
	AdminIDs    []string
}

// AddBadWords will add a new word into the MongoDB database.
func (d *Dependency) AddBadWord(ctx context.Context, word string) error {
	col := d.Mongo.Database(d.MongoDBName).Collection("badstuff")
	_, err := col.InsertOne(ctx, bson.D{{Key: "value", Value: word}})
	if err != nil {
		return err
	}
	return nil
}

// Authenticate will check if the user is allowed to add a new
// badword into the database.
func (d *Dependency) Authenticate(id string) bool {
	admins := d.AdminIDs

	for _, admin := range admins {
		if admin == id {
			return true
		}
	}
	return false
}
