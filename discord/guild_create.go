package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/poundbot/poundbot/types"
)

type guildCreateAccountStorer interface {
	UpsertBase(types.BaseAccount) error
	SetRegisteredPlayerIDs(ServerID string, IDs []string) error
	GetByDiscordGuild(string) (types.Account, error)
}

type guildCreateUserGetter interface {
	GetPlayerIDsByDiscordIDs(snowflakes []string) ([]string, error)
}

type guildCreate struct {
	as guildCreateAccountStorer
	ug guildCreateUserGetter
}

func NewGuildCreate(as guildCreateAccountStorer, ug guildCreateUserGetter) func(*discordgo.Session, *discordgo.GuildCreate) {
	gc := guildCreate{as: as, ug: ug}
	return gc.guildCreate
}

func (g guildCreate) guildCreate(s *discordgo.Session, gc *discordgo.GuildCreate) {
	log.Printf("Guild Create %s:%s", gc.ID, gc.Name)
	userIDs := make([]string, len(gc.Members))
	for i, member := range gc.Members {
		userIDs[i] = member.User.ID
	}
	account, err := g.as.GetByDiscordGuild(gc.ID)
	if err == nil {
		account.OwnerSnowflake = gc.OwnerID
	} else if err.Error() != "not found" {
		log.Printf("Error: GuildCreate: %v", err)
		return
	} else {
		account.BaseAccount = types.BaseAccount{GuildSnowflake: gc.ID, OwnerSnowflake: gc.OwnerID}
	}

	g.as.UpsertBase(account.BaseAccount)

	playerIDs, err := g.ug.GetPlayerIDsByDiscordIDs(userIDs)
	if err != nil {
		log.Printf("guildCreate: Error getting playerIDs: %v", err)
		return
	}

	log.Printf("guild: %s(%s), playerIDs:%v", gc.Name, gc.ID, playerIDs)

	err = g.as.SetRegisteredPlayerIDs(account.GuildSnowflake, playerIDs)
	if err != nil {
		log.Printf("guildCreate: Error setting playerIDs: %v", err)
	}
}
