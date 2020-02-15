package discord

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/poundbot/poundbot/messages"
	"github.com/poundbot/poundbot/types"

	"github.com/sirupsen/logrus"
)

type dmUserStorage interface {
	GetByDiscordID(snowflake string) (types.User, error)
	RemovePlayerID(snowflake, playerID string) error
}

type dmAuthStorage interface {
	AddRegisteredPlayerIDs(accountID string, playerIDs []string) error
}

type dmDiscordAccountStorage interface {
	GetByDiscordID(snowflake string) (types.DiscordAuth, error)
}

type dm struct {
	us       dmUserStorage
	as       dmAuthStorage
	das      dmDiscordAccountStorage
	authChan chan<- types.DiscordAuth
}

func (i dm) process(m *discordgo.MessageCreate) string {
	pLog := log.WithFields(logrus.Fields{"sys": "dm.process()"})
	message := strings.TrimSpace(m.Content)

	isPIN, err := regexp.MatchString("\\A[0-9]+\\z", message)
	if err != nil {
		pLog.Error(err)
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "InternalError",
				Other: "Internal error. Please try again.",
			}})
	}

	if isPIN {
		return i.validatePIN(message, m.Author.ID)
	}

	parts := strings.Fields(message)

	switch parts[0] {
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandStatus",
			Other: "status",
		},
	}):
		return i.status(m.Author.ID)
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandHelp",
			Other: "help",
		},
	}):
		return i.help(m.Author.ID)
	case localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructCommandUnregister",
			Other: "unregister",
		},
	}):
		return i.unregister(m.Author.ID, parts[1:])
	}

	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "InstructInvalidCommand",
			Other: "Invalid command. See `help`",
		}})
}

func (i dm) status(authorID string) string {
	u, err := i.us.GetByDiscordID(authorID)
	if err != nil {
		return "You are not registered anywhere."
	}
	return fmt.Sprintf("Your registered IDs are: %s", strings.Join(u.PlayerIDs, ","))
}

func (i dm) unregister(authorID string, parts []string) string {
	u, err := i.us.GetByDiscordID(authorID)
	if err != nil {
		return "You are not registered anywhere."
	}

	if len(parts) == 0 {
		return "Usage: `unregister <game>`"
	}

	if parts[0] == "all" {
		u.PlayerIDs = []string{}
		i.us.RemovePlayerID(u.Snowflake, "all")
		return "You have been removed from all games."
	}

	game := parts[0]
	for _, pID := range u.PlayerIDs {
		if strings.HasPrefix(pID, fmt.Sprintf("%s:", game)) {
			err := i.us.RemovePlayerID(u.Snowflake, pID)
			if err != nil {
				// return "Could not remove ID, try again."
				return fmt.Sprintf("Could not remove ID, try again. %s", err)
			}
			return fmt.Sprintf("%s removed", pID)
		}
	}

	return fmt.Sprintf("Could not find an ID for game %s.\n%s", game, i.status(authorID))
}

func (i dm) help(authorID string) string {
	return messages.DMHelpText()
}

func (i dm) validatePIN(pin, authorID string) string {
	vpLog := log.WithFields(logrus.Fields{"sys": "dm.validatePIN()"})
	da, err := i.das.GetByDiscordID(authorID)
	if err != nil {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PINNotRequested",
				Other: "ERROR: PIN is not required at this time. Check `status` or `help`.",
			}})
	}

	vpLog = vpLog.WithFields(logrus.Fields{"playerid": da.PlayerID, "guildid": da.GuildSnowflake})

	if !(pinString(da.Pin) == pin) {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PINInvalid",
				Other: "Invalid PIN. Please try again.",
			}})
	}

	authResult := make(chan string)
	da.Ack = func(authenticated bool) {
		defer close(authResult)
		if authenticated {
			authResult <- localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "PINAuthenticated",
					Other: "You have authenticated!",
				}})
			err = i.as.AddRegisteredPlayerIDs(da.GuildSnowflake, []string{da.PlayerID})
			if err != nil {
				vpLog.Error(err)
			}
			return
		}

		authResult <- localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PINInternalError",
				Other: "Internal error. Please try again.",
			}})
	}
	i.authChan <- da
	return <-authResult
}
