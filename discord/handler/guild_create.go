package handler

import (
	"log"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/bwmarrin/discordgo"
)

type guildCreate struct {
	as storage.AccountsStore
}

func NewGuildCreate(as storage.AccountsStore) func(*discordgo.Session, *discordgo.GuildCreate) {
	gc := guildCreate{as: as}
	return gc.guildCreate
}

func (g *guildCreate) guildCreate(s *discordgo.Session, gc *discordgo.GuildCreate) {
	var a types.Account
	err := g.as.GetByDiscordGuild(gc.ID, &a)
	if err == nil {
		a.OwnerSnowflake = gc.OwnerID
		return
	} else if err.Error() != "not found" {
		log.Printf("Error: GuildCreate: %v", err)
		return
	} else {
		a.BaseAccount = types.BaseAccount{GuildSnowflake: gc.ID, OwnerSnowflake: gc.OwnerID}
	}

	g.as.UpsertBase(a.BaseAccount)
}
