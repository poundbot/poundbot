package types

// GamesInfo steam id translater between server and DB
// also used as a selector on the DB
type GamesInfo struct {
	PlayerIDs []string `bson:",omitempty"`
}

// BaseUser core user information for upserts
type BaseUser struct {
	GamesInfo   `bson:",inline" json:",inline"`
	DiscordInfo `bson:",inline" json:",inline"`
}

// User full user model
type User struct {
	BaseUser  `bson:",inline" json:",inline"`
	Timestamp `bson:",inline" json:",inline"`
}
