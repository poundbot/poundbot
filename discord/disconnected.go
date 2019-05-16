package discord

import (
	"github.com/bwmarrin/discordgo"
)

// Disconnected is a handler for the Disconnected discord call
func disconnected(status chan<- bool) func(s *discordgo.Session, event *discordgo.Disconnect) {
	return func(s *discordgo.Session, event *discordgo.Disconnect) {
		status <- false
		log.WithField("sys", "disconnected").Warn("Disconnected!")
	}
}
