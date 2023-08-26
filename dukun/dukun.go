package dukun

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDB's collection schema as taken from
// https://github.com/teknologi-umum/bot/blob/5b6df1e179b878f597efde10e20612d04ba0df02/src/services/dukun/index.js#L19-L31
//
// const dukunSchema = new mongoose.Schema(
// 	{
// 	  userID: Number,
// 	  firstName: String,
// 	  lastName: String,
// 	  userName: String,
// 	  points: Number,
// 	  master: Boolean,
// 	  createdAt: Date,
// 	  updatedAt: Date
// 	},
// 	{ collection: "suhu" }
// );

type Dependency struct {
	Mongo  *mongo.Client
	DBName string
}

type Dukun struct {
	UserID    int64     `json:"userID" bson:"userID"`
	FirstName string    `json:"firstName" bson:"firstName"`
	LastName  string    `json:"lastName" bson:"lastName"`
	UserName  string    `json:"userName" bson:"userName"`
	Points    int       `json:"points" bson:"points"`
	Master    bool      `json:"master" bson:"master"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

func (d *Dependency) GetAllDukun(ctx context.Context) ([]Dukun, error) {
	collection := d.Mongo.Database(d.DBName).Collection("suhu")
	cols, err := collection.Find(ctx, bson.D{{}})
	if err != nil {
		return []Dukun{}, err
	}
	defer cols.Close(ctx)

	var dukuns []Dukun
	for cols.Next(ctx) {
		var dukun Dukun
		err := cols.Decode(&dukun)
		if err != nil {
			return []Dukun{}, err
		}
		dukuns = append(dukuns, dukun)
	}

	return dukuns, nil
}
