package mongodb

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// type messageLock struct {
// 	MessageID   string
// 	MessageType string
// 	LockedAt    time.time
// }

// A MessageLocks implements db.MessageLocksStore
type MessageLocks struct {
	collection *mgo.Collection
}

func (ml MessageLocks) Obtain(mID, mType string) bool {
	ci, err := ml.collection.Upsert(
		bson.M{"messageid": mID},
		bson.M{
			"$setOnInsert": bson.M{
				"lockedat": iclock().Now().UTC(),
			},
		},
	)
	if err != nil {
		log.Printf("Could not create log for %s:%s", mID, mType)
		return false
	}
	if ci.UpsertedId == nil {
		return false
	}
	if ci.Matched != 0 {
		log.Printf("Matched but no UpsertedId for %s:%s", mID, mType)
		return false
	}
	return true
}
