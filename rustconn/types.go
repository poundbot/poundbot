package rustconn

import (
	"fmt"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

type Ack func(bool)

type EntityDeath struct {
	Name      string
	GridPos   string
	Owners    []uint64
	CreatedAt time.Time
}

type MongoID struct {
	ID bson.ObjectId `bson:"_id,omitempty"`
}

type DiscordInfo struct {
	DiscordID string `bson:"discord_id"`
}

type SteamInfo struct {
	SteamID uint64 `bson:"steam_id"`
}

type User struct {
	MongoID     `bson:",inline"`
	DiscordInfo `bson:",inline"`
	SteamInfo   `bson:",inline"`
	CreatedAt   time.Time `bson:"created_at"`
}

func (s SteamInfo) UpsertID() SteamInfo {
	return s
}

type RemoveUser struct {
	SteamInfo `bson:",inline"`
}

type DiscordAuth struct {
	MongoID     `bson:",inline"`
	DiscordInfo `bson:",inline"`
	SteamInfo   `bson:",inline"`
	Pin         int
	SentToUser  bool
	Ack         Ack `bson:"-"`
}

type RaidNotification struct {
	MongoID       `bson:",inline"`
	DiscordInfo   `bson:",inline"`
	GridPositions []string       `bson:"grid_positions"`
	Items         map[string]int `bson:"items"`
	AlertAt       time.Time      `bson:"alert_at"`
}

type RaidInventory struct {
	Name  string
	Count int
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
