package types

import "github.com/globalsign/mgo/bson"

type ChatMessage struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	PlayerID    string
	DiscordInfo `bson:",inline" json:"-"`
	ServerKey   string `json:"-"`
	ClanTag     string
	DisplayName string
	Message     string
	ChannelID   string `json:"-"`
	SentToUser  bool
}
