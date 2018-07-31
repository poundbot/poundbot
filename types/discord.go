package types

type DiscordInfo struct {
	DiscordID string `bson:"discord_id"`
}

type Ack func(bool)

type DiscordAuth struct {
	MongoID    `bson:",inline"`
	BaseUser   `bson:",inline"`
	Pin        int
	SentToUser bool
	Ack        Ack `bson:"-"`
}
