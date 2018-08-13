package mongodb

import (
	"fmt"
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
func (u Users) Get(steamID uint64, user *types.User) error {
	err := u.collection.Find(bson.M{"steamid": steamID}).One(&user)
	return err
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

// RemoveClan implements db.UsersStore.RemoveClan
func (u Users) RemoveClan(serverKey, tag string) error {
	s := fmt.Sprintf("servers.%s.clantag", serverKey)
	_, err := u.collection.UpdateAll(
		bson.M{s: tag},
		bson.M{"$unset": bson.M{s: 1}},
	)
	return err
}

// RemoveClansNotIn implements db.UsersStore.RemoveClansNotIn
func (u Users) RemoveClansNotIn(serverKey string, tags []string) error {
	s := fmt.Sprintf("servers.%s.clantag", serverKey)
	_, err := u.collection.UpdateAll(
		bson.M{s: bson.M{"$nin": tags}},
		bson.M{"$unset": bson.M{s: 1}},
	)
	return err
}

// SetClanIn implements db.UsersStore.SetClanIn
func (u Users) SetClanIn(serverKey, tag string, steamIds []uint64) error {
	_, err := u.collection.UpdateAll(
		bson.M{"steamid": bson.M{"$in": steamIds}},
		bson.M{"$set": bson.M{fmt.Sprintf("servers.%s.clantag", serverKey): tag}},
	)
	return err
}
