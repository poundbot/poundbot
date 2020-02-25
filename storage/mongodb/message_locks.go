package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// type messageLock struct {
// 	MessageID   string
// 	MessageType string
// 	LockedAt    time.time
// }

// A MessageLocks implements db.MessageLocksStore
type MessageLocks struct {
	collection *mongo.Collection
}

func (ml MessageLocks) Obtain(mID, mType string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ci, err := ml.collection.UpdateOne(
		ctx,
		bson.M{"messageid": mID},
		bson.M{
			"$setOnInsert": bson.M{
				"lockedat": iclock().Now().UTC(),
			},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Printf("Could not create log for %s:%s", mID, mType)
		return false
	}
	if ci.UpsertedID == nil {
		return false
	}
	if ci.MatchedCount != 0 {
		log.Printf("Matched but no UpsertedId for %s:%s", mID, mType)
		return false
	}
	return true
}
