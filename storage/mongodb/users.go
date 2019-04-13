package mongodb

import (
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// A Users implements db.UsersStore
type Users struct {
	collection *mgo.Collection
}

// Get implements db.UsersStore.Get
func (u Users) Get(steamID uint64) (types.User, error) {
	var user types.User
	err := u.collection.Find(bson.M{"steamid": steamID}).One(&user)
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
