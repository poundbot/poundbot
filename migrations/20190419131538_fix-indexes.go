package migrations

import (
	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
)

func init() {
	migrate.Register(func(db *mgo.Database) error { //Up
		err := db.C("discord_auths").DropIndexName("steamid_1")
		if err != nil {
			return err
		}
		return db.C("users").DropIndexName("steamid_1")
	}, func(db *mgo.Database) error { //Down
		return nil
	})
}
