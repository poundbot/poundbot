package types

type DiscordInfo struct {
	DiscordName string
	Snowflake   string
}

type Ack func(bool)

type DiscordAuth struct {
	GuildSnowflake string
	PlayerID       string
	DiscordInfo    `bson:",inline"`
	Pin            int
	Ack            Ack `bson:"-" json:"-"`
}

func (d DiscordAuth) GetPlayerID() string {
	return d.PlayerID
}

func (d DiscordAuth) GetDiscordID() string {
	return d.Snowflake
}
