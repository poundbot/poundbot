package types

type DiscordInfo struct {
	DiscordName string
	Snowflake   string
}

type Ack func(bool)

type DiscordAuth struct {
	GuildSnowflake string
	BaseUser       `bson:",inline"`
	Pin            int
	SentToUser     bool
	Ack            Ack `bson:"-" json:"-"`
}
