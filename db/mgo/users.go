package mgo

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/db"
	"mrpoundsign.com/poundbot/types"
)

type Users struct {
	collection *mgo.Collection
}

func (u Users) Get(si types.SteamInfo) (*types.User, error) {
	var user types.User
	err := u.collection.Find(si).One(&user)
	return &user, err
}

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

func (u Users) RemoveClan(tag string) error {
	_, err := u.collection.UpdateAll(
		bson.M{"clan_tag": tag},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
	return err
}

func (u Users) RemoveClansNotIn(tags []string) error {
	_, err := u.collection.UpdateAll(
		bson.M{"clan_tag": bson.M{"$nin": tags}},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
	return err
}

func (u Users) SetClanIn(tag string, steam_ids []uint64) error {
	_, err := u.collection.UpdateAll(
		bson.M{"steam_id": bson.M{"$in": steam_ids}},
		bson.M{"$set": bson.M{"clan_tag": tag}},
	)
	return err
}

var _ = db.UsersAccessLayer(Users{})
