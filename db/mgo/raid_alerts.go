package mgo

import (
	"fmt"
	"time"

	"mrpoundsign.com/poundbot/db"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

type RaidAlerts struct {
	collection *mgo.Collection
	users      db.UsersAccessLayer
}

func (r RaidAlerts) AddInfo(ed types.EntityDeath) error {
	for _, steamID := range ed.Owners {
		u, err := r.users.Get(types.SteamInfo{steamID})
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

func (r RaidAlerts) Remove(alert types.RaidNotification) error {
	return r.collection.Remove(alert.DiscordInfo)
}

var _ = db.RaidAlertsAccessLayer(RaidAlerts{})