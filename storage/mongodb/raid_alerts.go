package mongodb

import (
	"fmt"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// A RaidAlerts implements storage.RaidAlertsStore
type RaidAlerts struct {
	collection *mgo.Collection
	users      storage.UsersStore
}

// AddInfo implements storage.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(alertIn time.Duration, ed types.EntityDeath) error {
	var u types.User
	for _, steamID := range ed.Owners {
		err := r.users.Get(steamID, &u)
		if err == nil {
			_, err := r.collection.Upsert(
				u.SteamInfo,
				bson.M{
					"$setOnInsert": bson.M{
						"alertat": time.Now().UTC().Add(alertIn),
					},
					"$inc": bson.M{
						fmt.Sprintf("items.%s", ed.Name): 1,
					},
					"$addToSet": bson.M{
						"grid_positions": ed.GridPos,
					},
				},
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetReady implements storage.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady(alerts *[]types.RaidNotification) error {
	// change := mgo.Change{Remove: true}
	err := r.collection.Find(
		bson.M{
			"alertat": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	).All(alerts)
	return err
}

// Remove implements storage.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidNotification) error {
	return r.collection.Remove(alert.SteamInfo)
}
