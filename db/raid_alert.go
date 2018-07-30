package db

import (
	"fmt"
	"log"
	"time"

	"github.com/globalsign/mgo/bson"
	"mrpoundsign.com/poundbot/types"
)

func AddRaidAlertInfo(s *Session, ed types.EntityDeath) {
	for _, steamID := range ed.Owners {
		u, err := GetUser(s, types.SteamInfo{steamID})
		if err == nil {
			s.RaidAlertCollection().Upsert(
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
		}
	}
}

func FindReadyRaidAlerts(s *Session, alerts *[]types.RaidNotification) {
	err := s.RaidAlertCollection().Find(
		bson.M{
			"alert_at": bson.M{
				"$lte": time.Now().UTC(),
			},
		},
	).All(alerts)
	if err != nil {
		log.Printf("Error getting raid alerts! %s\n", err)
	}
}

func RemoveRaidAlert(s *Session, alert types.RaidNotification) {
	s.RaidAlertCollection().Remove(alert.DiscordInfo)
}
