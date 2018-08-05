package mgo

import (
	"fmt"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/db"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A RaidAlerts implements db.RaidAlertsStore
type RaidAlerts struct {
	collection *mgo.Collection
	users      db.UsersStore
}

// AddInfo implements db.RaidAlertsStore.AddInfo
func (r RaidAlerts) AddInfo(ed types.EntityDeath) error {
	for _, steamID := range ed.Owners {
		u, err := r.users.Get(types.SteamInfo{SteamID: steamID})
		if err == nil {
			_, err := r.collection.Upsert(
				u.DiscordInfo,
				bson.M{

					"$setOnInsert": bson.M{
						"alert_at": time.Now().UTC().Add(5 * time.Minute),
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

// GetReady implements db.RaidAlertsStore.GetReady
func (r RaidAlerts) GetReady(alerts *[]types.RaidNotification) error {
	err := r.collection.Find(
		bson.M{
			"alert_at": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	).All(alerts)
	return err
}

// Remove implements db.RaidAlertsStore.Remove
func (r RaidAlerts) Remove(alert types.RaidNotification) error {
	return r.collection.Remove(alert.DiscordInfo)
}
