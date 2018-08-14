package handler

import (
	"bitbucket.org/mrpoundsign/poundbot/storage"
	"github.com/bwmarrin/discordgo"
)

type guildDelete struct {
	as storage.AccountsStore
}

func NewGuildDelete(as storage.AccountsStore) func(*discordgo.Session, *discordgo.GuildDelete) {
	gd := guildDelete{as: as}
	return gd.guildDelete
}

func (g *guildDelete) guildDelete(s *discordgo.Session, gd *discordgo.GuildDelete) {
	g.as.Remove(gd.Guild.ID)
}
