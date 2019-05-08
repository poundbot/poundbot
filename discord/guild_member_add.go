package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
)

type userFinder interface {
	GetByDiscordID(snowflake string) (types.User, error)
}

type guildMemberAdder interface {
	GetByDiscordGuild(key string) (types.Account, error)
	AddRegisteredPlayerIDs(accountSnowflake string, playerIDs []string) error
}

func newGuildMemberAdd(uf userFinder, gma guildMemberAdder) func(*discordgo.Session, *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, dgma *discordgo.GuildMemberAdd) {
		guildMemberAdd(uf, gma, dgma.GuildID, dgma.Member.User.ID)
	}
}

func guildMemberAdd(uf userFinder, gma guildMemberAdder, gID, uID string) {
	user, err := uf.GetByDiscordID(uID)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Printf("guildMemberAdd: error finding user %s: %v", uID, err)
		}
		return
	}
	account, err := gma.GetByDiscordGuild(gID)
	if err != nil {
		if err != mgo.ErrNotFound {
			log.Printf("guildMemberAdd: Could not get account for guild ID %s: %v", gID, err)
		}
		return
	}
	err = gma.AddRegisteredPlayerIDs(account.GuildSnowflake, user.PlayerIDs)
	if err != nil {
		log.Printf("guildMemberAdd: Could not add player IDs to account %s: %v", account.GuildSnowflake, err)
	}
}
