package discord

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func (r *Runner) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	mcLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "messageCreate", "gID": m.GuildID})
	if !r.mls.Obtain(m.ID, "discord") {
		return
	}
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	var err error

	// Detect PM
	if len(m.GuildID) == 0 {
		r.interact(s, m)
		return
	}

	account, err := r.as.GetByDiscordGuild(m.GuildID)
	if err != nil {
		mcLog.WithError(err).Error("Could not get account for guild")
		return
	}

	if len(account.OwnerSnowflake) == 0 {
		mcLog.Info("Guild is missing owner")
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			mcLog.WithError(err).Error("Error getting guild from Discord")
			return
		}

		mcLog.WithField("oID", guild.OwnerID).Info("Setting owner")
		account.OwnerSnowflake = guild.OwnerID
		err = r.as.UpsertBase(account.BaseAccount)
		if err != nil {
			mcLog.WithError(err).Error("Storage error updating account")
			return
		}
	}

	var response instructResponse
	respond := false

	// Detect prefix
	if strings.HasPrefix(m.Message.Content, account.GetCommandPrefix()) {
		m.Message.Content = strings.TrimPrefix(m.Message.Content, account.GetCommandPrefix())
		response = instruct(s.State.User.ID, m.ChannelID, m.Author.ID, m.Content, account, r.as)
		respond = true
	}

	// Detect mention
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			response = instruct(s.State.User.ID, m.ChannelID, m.Author.ID, m.Content, account, r.as)
			respond = true
		}
	}

	if respond {
		switch response.responseType {
		case instructResponsePrivate:
			err = r.sendPrivateMessage(m.Author.ID, response.message)
		case instructResponseChannel:
			err = r.sendChannelMessage(r.session.State.User.ID, m.ChannelID, response.message)
		}
		return
	}

	if len(account.Servers) == 0 {
		return
	}

	for _, server := range account.Servers {
		cTags, ok := server.TagsForChannelID(m.ChannelID)
		if !ok {
			continue
		}
		csLog := mcLog.WithFields(logrus.Fields{"serverID": server.Key[:4]})
		for _, cTag := range cTags {
			csLog.WithFields(logrus.Fields{"t": cTag}).Trace("inserting message")
			cm := types.ChatMessage{
				ServerKey:   server.Key,
				Tag:         cTag,
				DisplayName: m.Author.Username,
				Message:     m.Message.Content,
				DiscordInfo: types.DiscordInfo{
					Snowflake:   m.Author.ID,
					DiscordName: m.Author.String(),
				},
			}
			go func() {
				user, err := r.us.GetByDiscordID(m.Author.ID)
				if err == nil {
					found, clan := server.UsersClan(user.PlayerIDs)
					if found {
						cm.ClanTag = clan.Tag
					}
				}
				if len(cm.Message) > 128 {
					cm.Message = truncateString(cm.Message, 128)
					err = r.sendChannelMessage(r.session.State.User.ID, cm.ChannelID, localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "TruncatedMessage",
							Other: "*Truncated message to {{.Message}}",
						},
						TemplateData: map[string]string{"Message": cm.Message},
					}))
					if err != nil {
						csLog.WithError(err).Error("Error sendingmessage")
					}
				}
				err = r.cqs.InsertMessage(cm)
				if err != nil {
					mcLog.WithError(err).Error("Storage error saving message")
				}
			}()
		}
	}
}

type channelPermissionsGetter interface {
	UserChannelPermissions(userID, channelID string) (apermissions int, err error)
}

type messageChannelsGetter interface {
	channelPermissionsGetter
	Guild(guildID string) (st *discordgo.Guild, err error)
}

func sendChannelList(userID, guildID string, ch chan<- types.ServerChannelsResponse, mgg messageChannelsGetter) error {
	sclLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "sendChannelList", "gID": guildID, "uID": userID})
	defer close(ch)
	guild, err := mgg.Guild(guildID)
	if err != nil {
		ch <- types.ServerChannelsResponse{OK: false}
		sclLog.WithError(err).Warn("Could not find guild")
		return errors.Wrap(err, fmt.Sprintf("could not find guild with id %s", guildID))
	}

	r := types.ServerChannelsResponse{OK: true}
	for _, channel := range guild.Channels {
		canSend, err := canSendToChannel(mgg, userID, channel.ID)
		if err != nil {
			ch <- types.ServerChannelsResponse{OK: false}
			sclLog.WithError(err).Warn("Can not send to channel")
			return errors.Wrap(err, "channel send failed")
		}

		canEmbed, err := canEmbedToChannel(mgg, userID, channel.ID)
		if err != nil {
			ch <- types.ServerChannelsResponse{OK: false}
			sclLog.WithError(err).Warn("Cannot embed to channel")
			return errors.Wrap(err, "channel embed failed")
		}

		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		r.Channels = append(r.Channels, types.ServerChannel{ID: channel.ID, Name: channel.Name, CanSend: canSend, CanStyle: canEmbed})
	}
	ch <- r
	return nil
}

func (r *Runner) sendChannelMessage(userID, channelID, message string) error {
	scmLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "sendChannelMessage", "cID": channelID, "uID": userID})
	canSend, err := canSendToChannel(r.session.State, userID, channelID)
	if err != nil {
		scmLog.WithError(err).Warn("Cannot send to channel")
		return errors.Wrap(err, "cannot send to channel")
	}

	if !canSend {
		return errors.New("not permitted to send to channel")
	}

	_, err = r.session.ChannelMessageSend(channelID, message)
	if err != nil {
		scmLog.WithError(err).Warn("error sending message to channel")
	}
	return errors.Wrap(err, "error sending message to channel")
}

func (r *Runner) sendChannelEmbed(userID, channelID, message string, color int) error {
	sceLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "sendChannelMessage", "cID": channelID, "uID": userID})
	canEmbed, err := canEmbedToChannel(r.session.State, userID, channelID)
	if err != nil {
		sceLog.WithError(err).Warn("Cannot embed to channel")
		return errors.Wrap(err, "cannot embed to channel")
	}

	if !canEmbed {
		return errors.New("not permitted to embed to channel")
	}

	_, err = r.session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Description: message,
		Color:       color,
	})
	if err != nil {
		sceLog.WithError(err).Warn("error embedding message to channel")
	}
	return errors.Wrap(err, "error embedding message to channel")
}

func (r *Runner) sendPrivateMessage(snowflake, message string) error {
	spmLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "sendPrivateMessage", "cID": snowflake})
	channel, err := r.session.UserChannelCreate(snowflake)

	if err != nil {
		spmLog.WithError(err).Error("Error creating user channel")
		return errors.Wrap(err, "could not create user channel")
	}

	_, err = r.session.ChannelMessageSend(
		channel.ID,
		message,
	)

	return errors.Wrap(err, "error sending private message")
}

func canSendToChannel(pg channelPermissionsGetter, userID, channelID string) (bool, error) {
	cstcLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "canSendToChannel", "uID": userID, "cID": channelID})
	perms, err := pg.UserChannelPermissions(userID, channelID)

	if err != nil {
		cstcLog.WithError(err).WithField("cID", channelID).Trace("canSendToChannel: error reading permissions for channel")
		return false, errors.Wrap(err, "could not read permissions for channel")
	}

	if discordgo.PermissionSendMessages&^perms != 0 {
		cstcLog.WithField("cID", channelID).Trace("canSendToChannel: cannot send to channel")
		return false, nil
	}
	return true, nil
}

func canEmbedToChannel(pg channelPermissionsGetter, userID, channelID string) (bool, error) {
	cetcLog := log.WithFields(logrus.Fields{"sys": "RUN", "ssys": "canEmbedToChannel", "uID": userID, "cID": channelID})
	perms, err := pg.UserChannelPermissions(userID, channelID)

	if err != nil {
		cetcLog.WithError(err).WithField("cID", channelID).Trace("canEmbedToChannel: error sending to channel")
		return false, nil
	}

	if discordgo.PermissionEmbedLinks&^perms != 0 {
		cetcLog.WithField("cID", channelID).Trace("canEmbedToChannel: cannot embed to channel")
		return false, err
	}
	return true, nil
}
