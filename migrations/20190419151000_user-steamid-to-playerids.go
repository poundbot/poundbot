package migrations

import (
	"errors"
	"fmt"
	"log"

	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	migrate.Register(func(db *mgo.Database) error { //Up
		var steamids []uint64
		db.C("users").Find(bson.M{}).Distinct("steamid", &steamids)

		for _, id := range steamids {
			log.Printf("Fixing %d", id)
			err := db.C("users").Update(
				bson.M{"steamid": id},
				bson.M{
					"$set":   bson.M{"playerids": []string{fmt.Sprintf("rust:%d", id)}},
					"$unset": bson.M{"steamid": 1},
				},
			)
			if err != nil {
				return err
			}
		}
		return nil
	}, func(db *mgo.Database) error { //Down
		return errors.New("ploop")
	},
	)
}
