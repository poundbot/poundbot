package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
)

type guildMemberRemover interface {
	GetByDiscordGuild(key string) (types.Account, error)
	RemoveRegisteredPlayerIDs(accountSnowflake string, playerIDs []string) error
}

func newGuildMemberRemove(uf userFinder, gmr guildMemberRemover) func(*discordgo.Session, *discordgo.GuildMemberRemove) {
	return func(s *discordgo.Session, dgmr *discordgo.GuildMemberRemove) {
		guildMemberRemove(uf, gmr, dgmr.GuildID, dgmr.Member.User.ID)
	}
}

func guildMemberRemove(uf userFinder, gmr guildMemberRemover, gID, uID string) {
	user, err := uf.GetByDiscordID(uID)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Printf("guildMemberRemove: error finding user %s: %v", uID, err)
		}
		return
	}
	account, err := gmr.GetByDiscordGuild(gID)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Printf("guildMemberRemove: Could not get account for guild ID %s: %v", gID, err)
		}
		return
	}
	err = gmr.RemoveRegisteredPlayerIDs(account.GuildSnowflake, user.PlayerIDs)
	if err != nil {
		log.Printf("guildMemberRemove: Could not remove player IDs to account %s: %v", account.GuildSnowflake, err)
	}
}
