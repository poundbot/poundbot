package discord

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/poundbot/poundbot/types"
	"github.com/sirupsen/logrus"
)

type gameDiscordMessageSender func(string, string) error
type guildFinder func(string) (*discordgo.Guild, error)

// gameMessageHandler handles the messages interface from games
func gameMessageHandler(m types.GameMessage, gf guildFinder, sendMessage gameDiscordMessageSender) {
	defer close(m.ErrorResponse)

	mhLog := log.WithFields(logrus.Fields{
		"cmd":         "ChatChan",
		"channelName": m.ChannelName,
	})

	sendErrorResponse := func(errorCh chan<- error, err error) {
		select {
		case errorCh <- err:
		case <-time.After(time.Second / 2):
			mhLog.WithError(err).Error("no response sending message error to channel")
		}
	}

	channelID := ""

	if m.Snowflake == "" {
		sendErrorResponse(m.ErrorResponse, fmt.Errorf("no server defined"))
		mhLog.Error("no guild id provided with channel name")
		return
	}
	guild, err := gf(m.Snowflake)
	if err != nil {
		sendErrorResponse(m.ErrorResponse, fmt.Errorf("server not found"))
		mhLog.WithError(err).Error("Could not get guild from session")
		return
	}
	for _, gChan := range guild.Channels {
		mhLog.WithField("guildChan", gChan.Name).Trace("checking for channel match")
		if gChan.Name == m.ChannelName {
			channelID = gChan.ID
			break
		}
	}

	if channelID == "" {
		sendErrorResponse(m.ErrorResponse, errors.New("channel not found"))
		mhLog.Info("could not find channel")
		return
	}

	err = sendMessage(channelID, escapeDiscordString(m.Message))
	if err != nil {
		m.ErrorResponse <- errors.New("could not send to channel")
		mhLog.WithError(err).Error("Error sending chat to channel")
		return
	}
}

// gameChatHandler handles game chat messages
func gameChatHandler(cm types.ChatMessage, gf guildFinder, sendMessage gameDiscordMessageSender) {
	ccLog := log.WithFields(logrus.Fields{
		"cmd":         "ChatChan",
		"playerID":    cm.PlayerID,
		"guildID":     cm.Snowflake,
		"channelID":   cm.ChannelID,
		"name":        cm.DisplayName,
		"dName":       cm.DiscordName,
		"channelName": cm.ChannelName,
	})
	var clan = ""
	if cm.ClanTag != "" {
		clan = fmt.Sprintf("[%s] ", cm.ClanTag)
	}

	channelID := ""

	if cm.ChannelName != "" {
		if cm.Snowflake == "" {
			ccLog.Error("no guild id provided with channel name")
			return
		}

		guild, err := gf(cm.Snowflake)
		if err != nil {
			ccLog.WithError(err).Error("Could not get guild from session")
			return
		}

		for _, gChan := range guild.Channels {
			ccLog.WithField("guildChan", gChan.Name).Trace("checking for channel match")
			if gChan.Name == cm.ChannelName {
				channelID = gChan.ID
				break
			}
		}
	}

	err := sendMessage(
		channelID,
		fmt.Sprintf("☢️ @%s **%s%s**: %s",
			iclock().Now().UTC().Format("01-02 15:04 MST"),
			clan, escapeDiscordString(cm.DisplayName), escapeDiscordString(cm.Message)),
	)
	if err != nil {
		ccLog.WithError(err).Error("Error sending chat to channel.")
	}
}
