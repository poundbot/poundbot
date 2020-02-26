package mongodb

import (
	"context"
	"time"

	"github.com/poundbot/poundbot/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type chatQueueCloner interface {
	newConnection() (*mongo.Database, error)
}

type ChatQueue struct {
	collection *mongo.Collection
	cloner     chatQueueCloner
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

	log.Trace(to.Seconds())

	var cursor *mongo.Cursor

	db, err := cq.cloner.newConnection()
	if err != nil {
		log.WithError(err).Error("could not create session")
		return types.ChatMessage{}, false
	}
	defer db.Client().Disconnect(context.Background())

	coll := db.Collection(cq.collection.Name())

	cursor, err = coll.Find(
		ctx,
		bson.M{
			"serverkey":    sk,
			"tag":          tag,
			"senttoserver": false,
		},
		options.Find().SetCursorType(options.Tailable).SetMaxTime(to),
	)
	if err != nil {
		log.WithError(err).Error("blah")
		return types.ChatMessage{}, false
	}
	defer cursor.Close(context.Background())

	var cm types.ChatMessage
	for {
		if cursor.Next(ctx) {
			// if cursor.TryNext(ctx) {
			cursor.Decode(&cm)
			ur, err := coll.UpdateOne(
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
				continue
				// return cm, false
			}

			return cm, true
		}

		// log.Trace("next")

		// If TryNext returns false, the next document is not yet available, the cursor was exhausted and was closed, or
		// an error occured. TryNext should only be called again for the empty batch case.
		// if err := cursor.Err(); err != nil {
		// 	log.WithField("cmd", "GetGameServerMessage").WithError(err).Error("blah")
		// }
	}

	return cm, false
}
