package mongodb

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/types"
)

// A Users implements db.UsersStore
type Users struct {
	collection *mgo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) Get(gameUserID string) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{"GameUserID": gameUserID}).One(&user)
	return user, err
}

func (u Users) GetSnowflake(snowflake string) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{"snowflake": snowflake}).One(&user)
	return user, err
}

// UpsertBase implements db.UsersStore.UpsertBase
func (u Users) UpsertBase(user types.BaseUser) error {
	_, err := u.collection.Upsert(
		user.SteamInfo,
		bson.M{
			"$setOnInsert": bson.M{"createdat": time.Now().UTC()},
			"$set":         user,
		},
	)

	return err
}
