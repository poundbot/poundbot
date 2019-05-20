package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
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

func newGuildCreate(as guildCreateAccountStorer, ug guildCreateUserGetter) func(*discordgo.Session, *discordgo.GuildCreate) {
	gc := guildCreate{as: as, ug: ug}
	return gc.guildCreate
}

func (g guildCreate) guildCreate(s *discordgo.Session, gc *discordgo.GuildCreate) {
	gcLog := log.WithFields(logrus.Fields{"sys": "guildCreate", "guildID": gc.ID, "guildName": gc.Name})

	userIDs := make([]string, len(gc.Members))
	for i, member := range gc.Members {
		userIDs[i] = member.User.ID
	}

	account, err := g.as.GetByDiscordGuild(gc.ID)
	if err != nil {
		if err != mgo.ErrNotFound {
			// Some other storage error
			log.WithError(err).Error("Error loading account")
			return
		}
		account.BaseAccount = types.BaseAccount{GuildSnowflake: gc.ID, OwnerSnowflake: gc.OwnerID}
	} else {
		account.OwnerSnowflake = gc.OwnerID
	}

	err = g.as.UpsertBase(account.BaseAccount)
	if err != nil {
		gcLog.WithError(err).Error("Error upserting account")
		return
	}

	playerIDs, err := g.ug.GetPlayerIDsByDiscordIDs(userIDs)
	if err != nil {
		gcLog.WithError(err).Error("Error getting playerIDs")
		return
	}

	gcLog = gcLog.WithField("playerIDs", playerIDs)

	gcLog.Trace("Adding players")

	err = g.as.SetRegisteredPlayerIDs(account.GuildSnowflake, playerIDs)
	if err != nil {
		gcLog.WithError(err).Error("Error setting playerIDs")
	}
}

type guildRemover interface {
	Remove(string) error
}

func newGuildDelete(gr guildRemover) func(*discordgo.Session, *discordgo.GuildDelete) {
	return func(s *discordgo.Session, gd *discordgo.GuildDelete) {
		guildDelete(gr, gd.Guild.ID)
	}
}

func guildDelete(gr guildRemover, gID string) {
	gr.Remove(gID)
}

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
	gmaLog := log.WithFields(logrus.Fields{"sys": "guildMemberAdd", "guildID": gID, "userID": uID})
	user, err := uf.GetByDiscordID(uID)
	if err != nil {
		gmaLog.WithError(err).Trace("Error finding user")
		return
	}
	_, err = gma.GetByDiscordGuild(gID)
	if err != nil {
		if err != mgo.ErrNotFound {
			gmaLog.WithError(err).Trace("Could not get account for guild")
		}
		return
	}
	err = gma.AddRegisteredPlayerIDs(gID, user.PlayerIDs)
	if err != nil {
		gmaLog.WithError(err).Error("Storage error: Could not add player IDs to account")
	}
}

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
	gmrLog := log.WithFields(logrus.Fields{"sys": "guildMemberRemove", "guildID": gID, "userID": uID})
	user, err := uf.GetByDiscordID(uID)
	if err != nil {
		gmrLog.WithError(err).Error("Error finding user")
		return
	}
	account, err := gmr.GetByDiscordGuild(gID)
	if err != nil {
		if err != mgo.ErrNotFound {
			gmrLog.WithError(err).Error("Could not get account for guild ID")
		}
		return
	}
	err = gmr.RemoveRegisteredPlayerIDs(account.GuildSnowflake, user.PlayerIDs)
	if err != nil {
		gmrLog.WithError(err).Error("Could not remove player IDs to account")
	}
}
