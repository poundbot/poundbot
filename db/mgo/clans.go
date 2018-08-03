package mgo

import (
	"time"

	"mrpoundsign.com/poundbot/db"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

// A Clans implements db.ClansAccessLayer
type Clans struct {
	collection *mgo.Collection     // the clans collection
	users      db.UsersAccessLayer // users collection accessor
}

// Upsert implements github.com/mrpoundsign/poundbot/db.ClansAccessLayer.Upsert
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

// Remove implements db.ClansAccessLayer.Remove
func (c Clans) Remove(tag string) error {
	_, err := c.collection.RemoveAll(bson.M{"tag": tag})
	return err
}

// RemoveNotIn implements db.ClansAccessLayer.RemoveNotIn
func (c Clans) RemoveNotIn(tags []string) error {
	_, err := c.collection.RemoveAll(
		bson.M{"tag": bson.M{"$nin": tags}},
	)
	return err
}
