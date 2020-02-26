package discord

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
)

type roleGuildGetter interface {
	Guild(guildID string) (*discordgo.Guild, error)
}

type rolePlayerGetter interface {
	GetByPlayerID(PlayerID string) (types.User, error)
	GetByDiscordID(snowflake string) (types.User, error)
}

type roleMemberAdder interface {
	GuildMemberRoleAdd(guildID, userID, roleID string) (err error)
	GuildMemberRoleRemove(guildID, userID, roleID string) (err error)
}

func rolesSetHandler(userID string, rs types.RoleSet, state roleGuildGetter, rpg rolePlayerGetter, rma roleMemberAdder) {
	rsLog := log.WithFields(logrus.Fields{"cmd": "rolesSetHandler", "rsRole": rs.Role, "gID": rs.GuildID})
	rsLog.Tracef("roles set: %s", rs)
	guild, err := state.Guild(rs.GuildID)
	if err != nil {
		rsLog.WithError(err).Error("Could not find guild")
		return
	}

	var me *discordgo.Member
	maxRolePermit := -1
	for _, member := range guild.Members {
		if member.User.ID == userID {
			me = member
		}
	}

	if me == nil {
		rsLog.WithError(errors.New("could not find myself in guild")).Error("can't find me")
		return
	}

	var gRole *discordgo.Role
	for _, role := range guild.Roles {
		rsLog.Tracef("%s is %d", role.Name, role.Position)
		if role.ID == rs.Role || role.Name == rs.Role {
			gRole = role
			// break
		}
		if role.Position > maxRolePermit && role.Permissions&discordgo.PermissionManageRoles != 0 {
			for _, roleID := range me.Roles {
				rsLog.Tracef("Checking permissions on %s for %s", roleID, role.Name)
				if role.ID == roleID {
					maxRolePermit = role.Position
					break
				}
			}
		}
	}

	if gRole == nil {
		rsLog.Tracef("could not find role %s", rs.Role)
		return
	}

	rsLog = rsLog.WithFields(logrus.Fields{"rName": gRole.Name, "rID": gRole.ID})

	if maxRolePermit < gRole.Position {
		rsLog.WithField("rPerms", gRole.Permissions).Trace("I can't do that, dave.")
		return
	}

	for _, member := range guild.Members {
		rsLog = rsLog.WithFields(logrus.Fields{"uID": member.User.ID, "uName": member.User.Username})
		hasRole := false
		for _, role := range member.Roles {
			if role == gRole.ID {
				hasRole = true
				break
			}
		}

		u, err := rpg.GetByDiscordID(member.User.ID)
		if err != nil {
			if err != mgo.ErrNotFound {
				rsLog.WithError(err).Error("storage error finding user")
				continue
			}
		}

		shouldHaveRole := false
		for _, pID := range rs.PlayerIDs {
			for _, uPID := range u.PlayerIDs {
				if uPID == pID {
					shouldHaveRole = true
					break
				}
			}
		}

		if shouldHaveRole || hasRole {
			rsLog.Trace("checking user")
		}

		if shouldHaveRole && !hasRole {
			rsLog.Tracef("adding role")
			if err := rma.GuildMemberRoleAdd(guild.ID, member.User.ID, gRole.ID); err != nil {
				rsLog.WithField("uID", u.Snowflake).WithError(err).Error("Could not set role")
			}
			continue
		}

		if hasRole && !shouldHaveRole {
			rsLog.Tracef("removing role")
			if err := rma.GuildMemberRoleRemove(guild.ID, member.User.ID, gRole.ID); err != nil {
				rsLog.WithField("uID", u.Snowflake).WithError(err).Error("Could not remove role")
			}
			continue
		}
	}
}
