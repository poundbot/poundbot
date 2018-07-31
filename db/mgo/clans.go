package mgo

import (
	"time"

	"mrpoundsign.com/poundbot/db"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

type Clans struct {
	collection *mgo.Collection
	users      db.UsersAccessLayer
}

func (c Clans) Upsert(cl types.Clan) error {
	_, err := c.collection.Upsert(
		bson.M{
			"tag": cl.Tag,
		},
		bson.M{
			"$setOnInsert": bson.M{"created_at": time.Now().UTC().Add(5 * time.Minute)},
			"$set":         cl.BaseClan,
		},
	)
	if err != nil {
		return err
	}

	return c.users.SetClanIn(cl.Tag, cl.Members)
}

func (c Clans) Remove(tag string) error {
	_, err := c.collection.RemoveAll(bson.M{"tag": tag})
	return err
}

func (c Clans) RemoveNotIn(tags []string) error {
	_, err := c.collection.RemoveAll(
		bson.M{"tag": bson.M{"$nin": tags}},
	)
	return err
}

var _ = db.ClansAccessLayer(Clans{})
