package rustconn

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type EntityDeath struct {
	Name      string
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

func (u User) UpsertID() SteamInfo {
	return SteamInfo{SteamID: u.SteamID}
	// return &steamInfo{SteamID: u.SteamID}
}

type DiscordAuth struct {
	MongoID     `bson:",inline"`
	DiscordInfo `bson:",inline"`
	SteamInfo   `bson:",inline"`
	Pin         int
}

type RaidNotification struct {
	DiscordID string
	Items     []RaidInventory
}

type RaidInventory struct {
	Name  string
	Count int
}

// func (u User) UpsertData() (*bson.Document, error) {
// 	tf, err := mongo.TransformDocument(bson.Document{"$set": u})
// 	if err != nil {
// 		return tf, err
// 	}
// 	return &tf, nil
// }
