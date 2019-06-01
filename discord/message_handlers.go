package discord

import (
	"errors"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (c *Client) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	mcLog := log.WithFields(logrus.Fields{"sys": "RUN", "gID": m.GuildID})
	if !c.mls.Obtain(m.ID, "discord") {
		return
	}
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	var err error

	// Detect PM
	if m.GuildID == "" {
		c.interact(s, m)
		return
	}

	account, err := c.as.GetByDiscordGuild(m.GuildID)
	if err != nil {
		mcLog.WithError(err).Error("Could not get account for guild")
		return
	}

	if account.OwnerSnowflake == "" {
		mcLog.Info("Guild is missing owner")
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			mcLog.WithError(err).Error("Error getting guild from Discord")
			return
		}

		mcLog.WithField("gID", guild.OwnerID).Info("Setting owner")
		account.OwnerSnowflake = guild.OwnerID
		err = c.as.UpsertBase(account.BaseAccount)
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
		response = instruct(s.State.User.ID, m.ChannelID, m.Author.ID, m.Content, account, c.as)
		respond = true
	}

	// Detect mention
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			response = instruct(s.State.User.ID, m.ChannelID, m.Author.ID, m.Content, account, c.as)
			respond = true
		}
	}

	if respond {
		switch response.responseType {
		case instructResponsePrivate:
			err = c.sendPrivateMessage(m.Author.ID, response.message)
		case instructResponseChannel:
			err = c.sendChannelMessage(c.session.State.User.ID, m.ChannelID, response.message)
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
				user, err := c.us.GetByDiscordID(m.Author.ID)
				if err == nil {
					found, clan := server.UsersClan(user.PlayerIDs)
					if found {
						cm.ClanTag = clan.Tag
					}
				}
				if len(cm.Message) > 128 {
					cm.Message = truncateString(cm.Message, 128)
					err = c.sendChannelMessage(c.session.State.User.ID, cm.ChannelID, localizer.MustLocalize(&i18n.LocalizeConfig{
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
				err = c.cqs.InsertMessage(cm)
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
	defer close(ch)
	guild, err := mgg.Guild(guildID)
	if err != nil {
		ch <- types.ServerChannelsResponse{OK: false}
		return err
	}

	r := types.ServerChannelsResponse{OK: true}
	for _, channel := range guild.Channels {
		canSend, err := canSendToChannel(mgg, userID, channel.ID)
		if err != nil {
			ch <- types.ServerChannelsResponse{OK: false}
			return err
		}

		canEmbed, err := canEmbedToChannel(mgg, userID, channel.ID)
		if err != nil {
			ch <- types.ServerChannelsResponse{OK: false}
			return err
		}

		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		r.Channels = append(r.Channels, types.ServerChannel{ID: channel.ID, Name: channel.Name, CanSend: canSend, CanStyle: canEmbed})
	}
	ch <- r
	return nil
}

func (c Client) sendChannelMessage(userID, channelID, message string) error {
	canSend, err := canSendToChannel(c.session, userID, channelID)
	if err != nil {
		return err
	}

	if !canSend {
		return errors.New("cannot send to channel")
	}

	_, err = c.session.ChannelMessageSend(channelID, message)
	return err
}

func (c Client) sendChannelEmbed(userID, channelID, message string, color int) error {
	canEmbed, err := canEmbedToChannel(c.session, userID, channelID)
	if err != nil {
		return err
	}

	if !canEmbed {
		return errors.New("cannot send to channel")
	}

	_, err = c.session.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
		Description: message,
		Color:       color,
	})
	return err
}

func (c Client) sendPrivateMessage(snowflake, message string) error {
	channel, err := c.session.UserChannelCreate(snowflake)

	if err != nil {
		log.WithError(err).Error("Error creating user channel")
		return err
	}

	_, err = c.session.ChannelMessageSend(
		channel.ID,
		message,
	)

	return err
}

func canSendToChannel(pg channelPermissionsGetter, userID, channelID string) (bool, error) {
	perms, err := pg.UserChannelPermissions(userID, channelID)

	if err != nil {
		log.WithError(err).WithField("cID", channelID).Trace("canSendToChannel: error reading permissions for channel")
		return false, err
	}

	if discordgo.PermissionSendMessages&^perms != 0 {
		log.WithField("cID", channelID).Trace("canSendToChannel: cannot send to channel")
		return false, nil
	}
	return true, nil
}

func canEmbedToChannel(pg channelPermissionsGetter, userID, channelID string) (bool, error) {
	perms, err := pg.UserChannelPermissions(userID, channelID)

	if err != nil {
		log.WithError(err).WithField("cID", channelID).Trace("canSendToChannel: error sending to channel")
		return false, nil
	}

	if discordgo.PermissionEmbedLinks&^perms != 0 {
		log.WithField("cID", channelID).Trace("canSendToChannel: cannot embed to channel")
		return false, err
	}
	return true, nil
}
