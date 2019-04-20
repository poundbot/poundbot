package migrations

import (
	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func init() {
	migrate.Register(func(db *mgo.Database) error { //Up
		_, err := db.C("accounts").UpdateAll(
			bson.M{},
			bson.M{"$unset": bson.M{"servers.$[].clans": 1}})
		return err

	}, func(db *mgo.Database) error { //Down
		return nil
	})
}
