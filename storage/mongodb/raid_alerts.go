package mongodb

import (
	"fmt"
	"time"
	"errors"

	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// A RaidAlerts implements storage.RaidAlertsStore
type RaidAlerts struct {
	collection mongo.Collection
	users      storage.UsersStore
}

// AddInfo implements storage.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(alertIn time.Duration, ed types.EntityDeath) error {
	for _, pid := range ed.OwnerIDs {
		// Checking if the user exists, just bail if not
		_, err := r.users.Get(pid)
		if err != nil {
			continue
		}

		u := true
		uo := options.UpdateOptions{Upsert: &u}

		_, err = r.collection.UpdateOne(
			nil,
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
			&uo,
		)
	}
	return nil
}

// GetReady implements storage.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady() ([]types.RaidAlert, error) {
	var alerts []types.RaidAlert

	cur, err := r.collection.Find(
		nil,
		bson.M{
			"alertat": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	)
	if err != nil {
		return alerts, err
	}
	defer cur.Close(nil)

	for cur.Next(nil) {
		var ra types.RaidAlert
		err := cur.Decode(&ra)
		if err != nil {
			return []types.RaidAlert{}, nil
		}
		alerts = append(alerts, ra)
	}

	return alerts, err
}

// Remove implements storage.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidAlert) error {
	dr, err := r.collection.DeleteOne(nil, alert)
	if dr.DeletedCount != 1 {
		return errors.New("not found")
	}
	return err
}
