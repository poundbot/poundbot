package discord

import (
	"github.com/bwmarrin/discordgo"
)

// Disconnected is a handler for the Disconnected discord call
func Disconnected(status chan bool, logPrefix string) func(s *discordgo.Session, event *discordgo.Disconnect) {
	return func(s *discordgo.Session, event *discordgo.Disconnect) {
		status <- false
		log.Println(logPrefix + "[CONN] Disconnected!")
	}
}
