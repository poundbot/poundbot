package types

import "time"

// SteamInfo steam id translater between server and DB
// also used as a selector on the DB
type SteamInfo struct {
	SteamID uint64 `bson:"steam_id" json:"SteamID"`
}

// BaseUser core user information for upserts
type BaseUser struct {
	SteamInfo   `bson:",inline"`
	DiscordInfo `bson:",inline"`
	DisplayName string `bson:"display_name"`
	ClanTag     string `bson:"clan_tag"`
}

// User full user model
type User struct {
	BaseUser  `bson:",inline"`
	CreatedAt time.Time `bson:"created_at"`
}
