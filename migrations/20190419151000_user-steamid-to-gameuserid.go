package migrations

import (
	"errors"
	"fmt"

	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func init() {
	migrate.Register(func(db *mgo.Database) error { //Up
		var steamids []uint64
		db.C("users").Find(bson.M{}).Distinct("steamid", &steamids)

		for _, id := range steamids {
			db.C("users").Update(
				bson.M{"steamid": id},
				bson.M{
					"$set":   bson.M{"gameuserid": fmt.Sprintf("%d", id)},
					"$unset": bson.M{"steamid": 1},
				},
			)
		}
		return nil
	}, func(db *mgo.Database) error { //Down
		return errors.New("ploop")
	},
	)
}
