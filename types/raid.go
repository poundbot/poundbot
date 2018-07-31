package types

import (
	"fmt"
	"strings"
	"time"
)

type EntityDeath struct {
	Name      string
	GridPos   string
	Owners    []uint64
	CreatedAt time.Time
}

type RaidInventory struct {
	Name  string
	Count int
}

type RaidNotification struct {
	MongoID       `bson:",inline"`
	DiscordInfo   `bson:",inline"`
	GridPositions []string       `bson:"grid_positions"`
	Items         map[string]int `bson:"items"`
	AlertAt       time.Time      `bson:"alert_at"`
}

func (rn RaidNotification) String() string {
	index := 0
	items := make([]string, len(rn.Items))
	for k, v := range rn.Items {
		items[index] = fmt.Sprintf("%s(%d)", k, v)
		index++
	}

	return fmt.Sprintf(`
	RAID ALERT! You are being raided!
	
	Locations: 
	  %s
	
	Destroyed:
	  %s
	`, strings.Join(rn.GridPositions, ", "), strings.Join(items, ", "))
}
