package mgo

import (
	"time"

	"mrpoundsign.com/poundbot/db"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

// A Clans implements db.ClansStore
type Clans struct {
	collection *mgo.Collection // the clans collection
	users      db.UsersStore   // users collection accessor
}

// Upsert implements github.com/mrpoundsign/poundbot/db.ClansStore.Upsert
//
// Note: it also uses Users to handle clan importing
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

// Remove implements db.ClansStore.Remove
func (c Clans) Remove(tag string) error {
	_, err := c.collection.RemoveAll(bson.M{"tag": tag})
	return err
}

// RemoveNotIn implements db.ClansStore.RemoveNotIn
func (c Clans) RemoveNotIn(tags []string) error {
	_, err := c.collection.RemoveAll(
		bson.M{"tag": bson.M{"$nin": tags}},
	)
	return err
}
