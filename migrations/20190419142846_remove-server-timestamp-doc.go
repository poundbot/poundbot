package migrations

import (
	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	migrate.Register(func(db *mgo.Database) error { //Up
		_, err := db.C("accounts").UpdateAll(
			bson.M{"servers": bson.M{"$exists": true}},
			bson.M{"$unset": bson.M{"servers.$[].timestamp": 1}},
		)
		return err

	}, func(db *mgo.Database) error { //Down
		return nil
	})
}
