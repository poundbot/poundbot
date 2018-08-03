package mgo

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

// A Users implements db.UsersAccessLayer
type Users struct {
	collection *mgo.Collection
}

// Get implements db.UsersAccessLayer.Get
func (u Users) Get(si types.SteamInfo) (*types.User, error) {
	var user types.User
	err := u.collection.Find(si).One(&user)
	return &user, err
}

// BaseUpsert implements db.UsersAccessLayer.BaseUpsert
func (u Users) BaseUpsert(user types.BaseUser) error {
	_, err := u.collection.Upsert(
		user.SteamInfo,
		bson.M{
			"$setOnInsert": bson.M{"created_at": time.Now().UTC().Add(5 * time.Minute)},
			"$set":         user,
		},
	)
	return err
}

// RemoveClan implements db.UsersAccessLayer.RemoveClan
func (u Users) RemoveClan(tag string) error {
	_, err := u.collection.UpdateAll(
		bson.M{"clan_tag": tag},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
	return err
}

// RemoveClansNotIn implements db.UsersAccessLayer.RemoveClansNotIn
func (u Users) RemoveClansNotIn(tags []string) error {
	_, err := u.collection.UpdateAll(
		bson.M{"clan_tag": bson.M{"$nin": tags}},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
	return err
}

// SetClanIn implements db.UsersAccessLayer.SetClanIn
func (u Users) SetClanIn(tag string, steamIds []uint64) error {
	_, err := u.collection.UpdateAll(
		bson.M{"steam_id": bson.M{"$in": steamIds}},
		bson.M{"$set": bson.M{"clan_tag": tag}},
	)
	return err
}
