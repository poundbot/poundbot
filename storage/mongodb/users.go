package mongodb

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

const userPlayerIDsField = "playerids"
const userSnowflakeField = "snowflake"

// A Users implements db.UsersStore
type Users struct {
	collection *mgo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) GetByPlayerID(gameUserID string) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{userPlayerIDsField: gameUserID}).One(&user)
	return user, err
}

func (u Users) GetByDiscordID(snowflake string) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{userSnowflakeField: snowflake}).One(&user)
	return user, err
}

func (u Users) GetPlayerIDsByDiscordIDs(snowflakes []string) ([]string, error) {
	var playerIDs []string
	err := u.collection.Find(bson.M{userSnowflakeField: bson.M{"$in": snowflakes}}).
		Distinct(userPlayerIDsField, &playerIDs)
	return playerIDs, err
}

func (u Users) UpsertPlayer(info storage.UserInfoGetter) error {
	_, err := u.collection.Upsert(
		bson.M{userSnowflakeField: info.GetDiscordID()},
		bson.M{
			"$setOnInsert": bson.M{
				userSnowflakeField: info.GetDiscordID(),
				"createdat":        time.Now().UTC(),
			},
			"$set":      bson.M{"updatedat": time.Now().UTC()},
			"$addToSet": bson.M{userPlayerIDsField: info.GetPlayerID()},
		},
	)

	return err
}

func (u Users) RemovePlayerID(snowflake, playerID string) error {
	if playerID == "all" {
		err := u.collection.Remove(
			bson.M{userSnowflakeField: snowflake},
		)
		return err
	}

	err := u.collection.Update(
		bson.M{userSnowflakeField: snowflake},
		bson.M{
			"$set": bson.M{
				"updatedat": time.Now().UTC(),
			},
			"$pull": bson.M{userPlayerIDsField: playerID},
		},
	)
	return err
}
