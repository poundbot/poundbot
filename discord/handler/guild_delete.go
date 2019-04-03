package handler

import (
	"github.com/bwmarrin/discordgo"
)

type guildRemover interface {
	Remove(string) error
}

func NewGuildDelete(gr guildRemover) func(*discordgo.Session, *discordgo.GuildDelete) {
	return func(s *discordgo.Session, gd *discordgo.GuildDelete) {
		guildDelete(gr, gd)
	}
}

func guildDelete(gr guildRemover, gd *discordgo.GuildDelete) {
	gr.Remove(gd.Guild.ID)
}
