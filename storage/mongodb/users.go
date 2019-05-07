package mongodb

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

// A Users implements db.UsersStore
type Users struct {
	collection *mgo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) GetByPlayerID(gameUserID string) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{"playerids": gameUserID}).One(&user)
	return user, err
}

func (u Users) GetByDiscordID(snowflake string) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{"snowflake": snowflake}).One(&user)
	return user, err
}

func (u Users) GetPlayerIDsByDiscordIDs(snowflakes []string) ([]string, error) {
	var playerIDs []string;
	err := u.collection.Find(bson.M{"snowflake": bson.M{"$in": snowflakes}}).Distinct("playerids", &playerIDs)
	return playerIDs, err
}

func (u Users) UpsertPlayer(info storage.UserInfoGetter) error {
	_, err := u.collection.Upsert(
		bson.M{"snowflake": info.GetDiscordID()},
		bson.M{
			"$setOnInsert": bson.M{
				"snowflake": info.GetDiscordID(),
				"createdat": time.Now().UTC(),
			},
			"$addToSet": bson.M{"playerids": info.GetPlayerID()},
		},
	)

	return err
}
