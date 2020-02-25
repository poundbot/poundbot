package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A RaidAlerts implements storage.RaidAlertsStore
type RaidAlerts struct {
	collection *mongo.Collection
	users      storage.UsersStore
}

// AddInfo implements storage.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(alertIn time.Duration, ed types.EntityDeath) error {

	for _, pid := range ed.OwnerIDs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Checking if the user exists, just bail if not
		_, err := r.users.GetByPlayerID(pid)
		if err != nil {
			continue
		}

		_, err = r.collection.UpdateOne(
			ctx,
			bson.M{"playerid": pid},
			bson.M{
				"$setOnInsert": bson.M{
					"alertat":    time.Now().UTC().Add(alertIn),
					"servername": ed.ServerName,
					"serverkey":  ed.ServerKey,
				},
				"$inc": bson.M{
					fmt.Sprintf("items.%s", ed.Name): 1,
				},
				"$addToSet": bson.M{
					"gridpositions": ed.GridPos,
				},
			},
			options.Update().SetUpsert(true),
		)
	}
	return nil
}

// GetReady implements storage.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady() ([]types.RaidAlert, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var alerts []types.RaidAlert

	c, err := r.collection.Find(
		ctx,
		bson.M{
			"alertat": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	)
	if err != nil {
		return alerts, err
	}
	err = c.All(ctx, &alerts)
	return alerts, err
}

// Remove implements storage.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidAlert) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	d, err := r.collection.DeleteOne(ctx, bson.M{"_id": alert.ID})
	if d.DeletedCount != 1 {
		return errors.New("coul not find raid alert to delete")
	}
	return err
}
