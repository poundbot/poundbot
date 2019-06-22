package mongodb

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

// A RaidAlerts implements storage.RaidAlertsStore
type RaidAlerts struct {
	collection *mgo.Collection
	users      storage.UsersStore
}

// AddInfo implements storage.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(alertIn time.Duration, ed types.EntityDeath) error {
	for _, pid := range ed.OwnerIDs {
		// Checking if the user exists, just bail if not
		_, err := r.users.GetByPlayerID(pid)
		if err != nil {
			continue
		}

		_, err = r.collection.Upsert(
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
		)
	}
	return nil
}

// GetReady implements storage.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady() ([]types.RaidAlert, error) {
	var alerts []types.RaidAlert
	// change := mgo.Change{Remove: true}
	err := r.collection.Find(
		bson.M{
			"alertat": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	).All(&alerts)
	return alerts, err
}

// Remove implements storage.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidAlert) error {
	return r.collection.Remove(bson.M{"_id": alert.ID})
}
