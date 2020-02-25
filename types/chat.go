package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type ChatMessage struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	DiscordInfo  `bson:",inline" json:"-"`
	ChannelID    string `json:"-"`
	ClanTag      string
	DisplayName  string
	Message      string
	PlayerID     string
	ServerKey    string `json:"-"`
	SentToServer bool   `json:"-"`
	Tag          string
}
