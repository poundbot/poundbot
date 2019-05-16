package types

import "github.com/globalsign/mgo/bson"

type ChatMessage struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	DiscordInfo `bson:",inline" json:"-"`
	ChannelID   string `json:"-"`
	ChannelName string
	ClanTag     string
	DisplayName string
	Message     string
	PlayerID    string
	SentToUser  bool
	ServerKey   string `json:"-"`
	Tag         string
}
