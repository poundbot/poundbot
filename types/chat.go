package types

const (
	ChatSourceRust    = "rust"
	ChatSourceDiscord = "discord"
)

type ChatMessage struct {
	SteamInfo   `bson:",inline"`
	ClanTag     string
	DisplayName string
	Message     string
	Source      string
}
