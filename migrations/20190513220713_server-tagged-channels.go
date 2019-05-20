package migrations

import (
	"fmt"

	migrate "github.com/eminetto/mongo-migrate"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func init() {
	err := migrate.Register(func(db *mgo.Database) error { //Up
		type server struct {
			ChatChanID string
		}

		type account struct {
			ID      bson.ObjectId `bson:"_id"`
			Servers []server
		}

		var accounts []account

		coll := db.C("accounts")
		if err := coll.Find(bson.M{}).All(&accounts); err != nil {
			return err
		}

		for _, account := range accounts {
			sets := bson.M{}
			unsets := bson.M{}
			for i := range account.Servers {
				sets[fmt.Sprintf("servers.%d.channels", i)] = []bson.M{
					{"channel_id": account.Servers[i].ChatChanID, "tags": []string{"chat"}},
				}
				unsets[fmt.Sprintf("servers.%d.chatchanid", i)] = 1
				unsets[fmt.Sprintf("servers.%d.serverchanid", i)] = 1
			}
			if len(sets) > 0 {
				if err := coll.Update(
					bson.M{"_id": account.ID},
					bson.M{
						"$set":   sets,
						"$unset": unsets,
					},
				); err != nil {
					return err
				}
			}
		}
		return nil

	}, func(db *mgo.Database) error { //Down
		return nil
	})

	if err != nil {
		panic(err)
	}
}
