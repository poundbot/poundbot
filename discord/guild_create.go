package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

type guildCreate struct {
	as storage.AccountsStore
}

func NewGuildCreate(as storage.AccountsStore) func(*discordgo.Session, *discordgo.GuildCreate) {
	gc := guildCreate{as: as}
	return gc.guildCreate
}

func (g guildCreate) guildCreate(s *discordgo.Session, gc *discordgo.GuildCreate) {
	account, err := g.as.GetByDiscordGuild(gc.ID)
	if err == nil {
		account.OwnerSnowflake = gc.OwnerID
		return
	} else if err.Error() != "not found" {
		log.Printf("Error: GuildCreate: %v", err)
		return
	} else {
		account.BaseAccount = types.BaseAccount{GuildSnowflake: gc.ID, OwnerSnowflake: gc.OwnerID}
	}

	g.as.UpsertBase(account.BaseAccount)
}
