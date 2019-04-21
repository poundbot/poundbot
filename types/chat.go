package types

type ChatMessage struct {
	PlayerID    string
	DiscordInfo `bson:",inline" json:"-"`
	ServerKey   string `json:"-"`
	ClanTag     string
	DisplayName string
	Message     string
	ChannelID   string `json:"-"`
}
