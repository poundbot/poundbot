package types

type ChatMessage struct {
	SteamInfo   `bson:",inline" json:",inline"`
	DiscordInfo `bson:",inline" json:"-"`
	ServerKey   string `json:"-"`
	ClanTag     string
	DisplayName string
	Message     string
	ChannelID   string `json:"-"`
}
