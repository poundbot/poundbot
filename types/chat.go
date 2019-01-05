package types

const (
	ChatSourceRust    = "rust"
	ChatSourceDiscord = "discord"
)

type ChatMessage struct {
	SteamInfo   `bson:",inline" json:",inline"`
	DiscordInfo `bson:",inline" json:"-"`
	ServerKey   string `json:"-"`
	ClanTag     string
	DisplayName string
	Message     string
	Source      string
	ChannelID   string `json:"-"`
	Timestamp   `bson:",inline" json:",inline"`
}
