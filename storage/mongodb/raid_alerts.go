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
	for _, steamID := range ed.Owners {
		user, err := r.users.Get(steamID)
		if err == nil {
			_, err := r.collection.Upsert(
				user.SteamInfo,
				bson.M{
					"$setOnInsert": bson.M{
						"alertat":    time.Now().UTC().Add(alertIn),
						"servername": ed.ServerName,
					},
					"$inc": bson.M{
						fmt.Sprintf("items.%s", ed.Name): 1,
					},
					"$addToSet": bson.M{
						"gridpositions": ed.GridPos,
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
	return r.collection.Remove(alert.SteamInfo)
}
