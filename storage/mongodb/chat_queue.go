package mongodb

import (
	"context"
	"time"

	"github.com/poundbot/poundbot/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChatQueue struct {
	collection *mongo.Collection
}

func (cq ChatQueue) InsertMessage(m types.ChatMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := cq.collection.InsertOne(ctx, m)
	return err
}

func (cq ChatQueue) GetGameServerMessage(sk, tag string, to time.Duration) (types.ChatMessage, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()

	var cursor *mongo.Cursor
	// defer cursor.Close(ctx)

	cursor, err := cq.collection.Find(
		ctx,
		bson.M{
			"serverkey":    sk,
			"tag":          tag,
			"senttoserver": false,
		},
	)
	if err != nil {
		return types.ChatMessage{}, false
	}

	var cm types.ChatMessage
	if cursor.TryNext(ctx) {
		cursor.Decode(&cm)
		ur, err := cq.collection.UpdateOne(
			ctx,
			bson.M{"_id": cm.ID, "senttoserver": false},
			bson.M{"$set": bson.M{"senttoserver": true}},
		)

		if err != nil {
			log.WithError(err).Trace("error updating message")
			return cm, false
		}

		if ur.ModifiedCount == 0 {
			log.Trace("message not modified")
			return cm, false
		}

		return cm, true
	}

	// If TryNext returns false, the next document is not yet available, the cursor was exhausted and was closed, or
	// an error occured. TryNext should only be called again for the empty batch case.
	if err := cursor.Err(); err != nil {
		log.WithField("cmd", "GetGameServerMessage").Error(err)
	}

	return cm, false
}
