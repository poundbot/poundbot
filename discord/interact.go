package discord

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

func (r *Runner) interact(s *discordgo.Session, m *discordgo.MessageCreate) {
	da, err := r.getDiscordAuth(m.Author.ID)
	if err != nil {
		return
	}

	if !(pinString(da.Pin) == strings.TrimSpace(m.Content)) {
		s.ChannelMessageSend(m.ChannelID, localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PinInvalid",
				Other: "Invalid pin. Please try again.",
			}}))
		return
	}

	da.Ack = func(authed bool) {
		if authed {
			s.ChannelMessageSend(m.ChannelID, localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "PinAuthenticated",
					Other: "You have authenticated!",
				}}))
			err = r.as.AddRegisteredPlayerIDs(da.GuildSnowflake, []string{da.PlayerID})
			if err != nil {
				log.WithFields(logrus.Fields{"sys": "interact()", "playerid": da.PlayerID, "guildid": da.GuildSnowflake, "err": err}).Error(
					"Could not add player discord account.",
				)
			}
			return
		}
		s.ChannelMessageSend(m.ChannelID, localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PinInternalError",
				Other: "Internal error. Please try again.",
			}}))
	}
	r.AuthSuccess <- da
}
