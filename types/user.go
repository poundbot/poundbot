package types

// SteamInfo steam id translater between server and DB
// also used as a selector on the DB
type SteamInfo struct {
	SteamID uint64
}

// BaseUser core user information for upserts
type BaseUser struct {
	SteamInfo   `bson:",inline" json:",inline"`
	DiscordInfo `bson:",inline" json:",inline"`
	DisplayName string
}

// User full user model
type User struct {
	BaseUser  `bson:",inline" json:",inline"`
	Timestamp `bson:",inline" json:",inline"`
}
