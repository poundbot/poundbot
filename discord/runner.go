package discord

import (
	"fmt"
	"strings"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

var iclock = pbclock.Clock

type RunnerConfig struct {
	Token string
}

type Client struct {
	session       *discordgo.Session
	cqs           storage.ChatQueueStore
	as            storage.AccountsStore
	mls           storage.MessageLocksStore
	das           storage.DiscordAuthsStore
	us            storage.UsersStore
	token         string
	status        chan bool
	ChatChan      chan types.ChatMessage
	RaidAlertChan chan types.RaidAlert
	DiscordAuth   chan types.DiscordAuth
	AuthSuccess   chan types.DiscordAuth
	shutdown      bool
}

func Runner(token string, as storage.AccountsStore, das storage.DiscordAuthsStore,
	us storage.UsersStore, mls storage.MessageLocksStore, cqs storage.ChatQueueStore) *Client {
	return &Client{
		cqs:           cqs,
		mls:           mls,
		as:            as,
		das:           das,
		us:            us,
		token:         token,
		ChatChan:      make(chan types.ChatMessage),
		DiscordAuth:   make(chan types.DiscordAuth),
		AuthSuccess:   make(chan types.DiscordAuth),
		RaidAlertChan: make(chan types.RaidAlert),
	}
}

// Start starts the runner
func (c *Client) Start() error {
	session, err := discordgo.New("Bot " + c.token)
	if err == nil {
		c.session = session
		c.session.AddHandler(c.messageCreate)
		c.session.AddHandler(c.ready)
		c.session.AddHandler(disconnected(c.status))
		c.session.AddHandler(c.resumed)
		c.session.AddHandler(newGuildCreate(c.as, c.us))
		c.session.AddHandler(newGuildDelete(c.as))
		c.session.AddHandler(newGuildMemberAdd(c.us, c.as))
		c.session.AddHandler(newGuildMemberRemove(c.us, c.as))

		c.status = make(chan bool)

		go c.runner()

		connect(session)
	}
	return err
}

// Stop stops the runner
func (c *Client) Stop() {
	log.WithFields(logrus.Fields{"ssys": "RUNNER"}).Info(
		"Disconnecting...",
	)
	// log.Println(logPrefix + "[CONN] Disconnecting...")
	c.shutdown = true
	c.session.Close()
}

func (c *Client) runner() {
	rLog := log.WithFields(logrus.Fields{"ssys": "RUNNER"})
	defer rLog.Warn("Runner exited")

	connectedState := false

	for {
		if connectedState {
			rLog.Info("Waiting for messages.")
		Reading:
			for {
				select {
				case connectedState = <-c.status:
					if !connectedState {
						rLog.Warn("Received disconnected message")
						if c.shutdown {
							return
						}
						break Reading
					}

					rLog.Info("Received unexpected connected message")

				case t := <-c.RaidAlertChan:
					raLog := rLog.WithFields(logrus.Fields{"chan": "RAID", "playerID": t.PlayerID})
					raLog.Trace("Gor raid alert")
					go func() {
						raUser, err := c.us.GetByPlayerID(t.PlayerID)
						if err != nil {
							raLog.WithError(err).Error("Player not found trying to send raid alert")
							return
						}

						user, err := c.session.User(raUser.Snowflake)
						if err != nil {
							raLog.WithField("userID", raUser.Snowflake).WithError(err).Error(
								"Discord user not found trying to send raid alert",
							)
							return
						}

						c.sendPrivateMessage(user.ID, t.String())
					}()

				case t := <-c.DiscordAuth:
					dLog := rLog.WithFields(logrus.Fields{
						"chan":    "DAUTH",
						"guildID": t.GuildSnowflake,
						"name":    t.DiscordInfo.DiscordName,
						"userID":  t.Snowflake,
					})
					dLog.Trace("Got discord auth")
					dUser, err := c.getUserByName(t.GuildSnowflake, t.DiscordInfo.DiscordName)
					if err != nil {
						dLog.WithError(err).Error("Discord user not found")
						err = c.das.Remove(t)
						if err != nil {
							dLog.WithError(err).Error("Error removing discord auth for PlayerID from the database.")
						}
						break
					}

					t.Snowflake = dUser.ID

					err = c.das.Upsert(t)
					if err != nil {
						dLog.WithError(err).Error("Error upserting PlayerID ito the database")
						break
					}

					err = c.sendPrivateMessage(t.Snowflake,
						localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{
								ID:    "UserPINPrompt",
								Other: "Enter the PIN provided in-game to validate your account.\nOnce you are validated, you will begin receiving raid alerts!",
							},
						}),
					)

					if err != nil {
						dLog.WithError(err).Error("Could not send PIN request to user")
					}

				case t := <-c.ChatChan:
					var clan = ""
					if t.ClanTag != "" {
						clan = fmt.Sprintf("[%s] ", t.ClanTag)
					}
					_, err := c.session.ChannelMessageSend(
						t.ChannelID,
						fmt.Sprintf("☢️ @%s **%s%s**: %s",
							iclock().Now().UTC().Format("01-02 15:04 MST"),
							clan, escapeDiscordString(t.DisplayName), escapeDiscordString(t.Message)),
					)
					if err != nil {
						log.WithFields(logrus.Fields{"ssys": "CHAT", "playerid": t.PlayerID, "chanid": t.ChannelID}).WithError(err).Error(
							"Error sending chat to channel.",
						)
					}
				}
			}
		}
	Connecting:
		for {
			rLog.Info("Waiting for connected state...")
			connectedState = <-c.status
			if connectedState {
				rLog.WithField("ssys", "CONN").Info("Received connected message")
				break Connecting
			}
			rLog.WithField("ssys", "CONN").Info("Received disconnected message")
		}
	}

}

func (c *Client) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.WithField("ssys", "CONN").Info("Resumed connection")
	c.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (c *Client) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithField("ssys", "CONN").Info("Connection Ready")

	s.UpdateStatus(0, localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "DiscordStatus",
			Other: "!pb help",
		}}),
	)
	guilds := make([]types.BaseAccount, len(s.State.Guilds))
	for i, guild := range s.State.Guilds {
		guilds[i] = types.BaseAccount{GuildSnowflake: guild.ID, OwnerSnowflake: guild.OwnerID}
	}
	c.as.RemoveNotInDiscordGuildList(guilds)
	c.status <- true
}

func (c Client) sendChannelMessage(channelID, message string) error {
	_, err := c.session.ChannelMessageSend(channelID, message)
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

// Returns nil user if they don't exist; Returns error if there was a communications error
func (c *Client) getUserByName(guildSnowflake, name string) (discordgo.User, error) {
	users, err := c.session.GuildMembers(guildSnowflake, "", 1000)
	if err != nil {
		return discordgo.User{}, fmt.Errorf("discord user not found %s in %s", name, guildSnowflake)
	}

	for _, user := range users {
		if strings.ToLower(user.User.String()) == strings.ToLower(name) {
			return *user.User, nil
		}
	}

	return discordgo.User{}, fmt.Errorf("discord user not found %s", name)
}

func (c *Client) getDiscordAuth(snowflake string) (types.DiscordAuth, error) {
	return c.das.GetByDiscordID(snowflake)
}
