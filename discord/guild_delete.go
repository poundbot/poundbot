package discord

import (
	"github.com/bwmarrin/discordgo"
)

type guildRemover interface {
	Remove(string) error
}

func NewGuildDelete(gr guildRemover) func(*discordgo.Session, *discordgo.GuildDelete) {
	return func(s *discordgo.Session, gd *discordgo.GuildDelete) {
		guildDelete(gr, gd.Guild.ID)
	}
}

func guildDelete(gr guildRemover, gID string) {
	gr.Remove(gID)
}
