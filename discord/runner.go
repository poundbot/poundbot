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
	session         *discordgo.Session
	cqs             storage.ChatQueueStore
	as              storage.AccountsStore
	mls             storage.MessageLocksStore
	das             storage.DiscordAuthsStore
	us              storage.UsersStore
	token           string
	status          chan bool
	ChatChan        chan types.ChatMessage
	RaidAlertChan   chan types.RaidAlert
	DiscordAuth     chan types.DiscordAuth
	GameMessageChan chan types.GameMessage
	AuthSuccess     chan types.DiscordAuth
	shutdown        bool
}

func Runner(token string, as storage.AccountsStore, das storage.DiscordAuthsStore,
	us storage.UsersStore, mls storage.MessageLocksStore, cqs storage.ChatQueueStore) *Client {
	return &Client{
		cqs:             cqs,
		mls:             mls,
		as:              as,
		das:             das,
		us:              us,
		token:           token,
		ChatChan:        make(chan types.ChatMessage),
		DiscordAuth:     make(chan types.DiscordAuth),
		AuthSuccess:     make(chan types.DiscordAuth),
		RaidAlertChan:   make(chan types.RaidAlert),
		GameMessageChan: make(chan types.GameMessage),
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
	defer c.session.Close()
	log.WithFields(logrus.Fields{"sys": "RUNNER"}).Info(
		"Disconnecting...",
	)
	// log.Println(logPrefix + "[CONN] Disconnecting...")
	c.shutdown = true
}

func (c *Client) runner() {
	rLog := log.WithFields(logrus.Fields{"sys": "RUNNER"})
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
					raLog := rLog.WithFields(logrus.Fields{"chan": "RAID", "pID": t.PlayerID})
					raLog.Trace("Got raid alert")
					go func() {
						raUser, err := c.us.GetByPlayerID(t.PlayerID)
						if err != nil {
							raLog.WithError(err).Error("Player not found trying to send raid alert")
							return
						}

						user, err := c.session.User(raUser.Snowflake)
						if err != nil {
							raLog.WithField("uID", raUser.Snowflake).WithError(err).Error(
								"Discord user not found trying to send raid alert",
							)
							return
						}

						err = c.sendPrivateMessage(user.ID, t.String())
						if err != nil {
							raLog.WithError(err).Error("could not create private channel to send to user")
						}
					}()

				case da := <-c.DiscordAuth:
					go c.discordAuthHandler(da)

				case m := <-c.GameMessageChan:
					go gameMessageHandler(m, c.session.State.Guild, c.sendChannelMessage, c.sendChannelEmbed)

				case cm := <-c.ChatChan:
					go gameChatHandler(cm, c.session.State.Guild, c.sendChannelMessage)
				}
			}
		}
	Connecting:
		for {
			rLog.Info("Waiting for connected state...")
			connectedState = <-c.status
			if connectedState {
				rLog.WithField("sys", "CONN").Info("Received connected message")
				break Connecting
			}
			rLog.WithField("sys", "CONN").Info("Received disconnected message")
		}
	}

}

func (c *Client) discordAuthHandler(da types.DiscordAuth) {
	dLog := log.WithFields(logrus.Fields{
		"chan": "DAUTH",
		"gID":  da.GuildSnowflake,
		"name": da.DiscordInfo.DiscordName,
		"uID":  da.Snowflake,
	})
	dLog.Trace("Got discord auth")
	dUser, err := c.getUserByName(da.GuildSnowflake, da.DiscordInfo.DiscordName)
	if err != nil {
		dLog.WithError(err).Error("Discord user not found")
		err = c.das.Remove(da)
		if err != nil {
			dLog.WithError(err).Error("Error removing discord auth for PlayerID from the database.")
		}
		return
	}

	da.Snowflake = dUser.ID

	err = c.das.Upsert(da)
	if err != nil {
		dLog.WithError(err).Error("Error upserting PlayerID ito the database")
		return
	}

	err = c.sendPrivateMessage(da.Snowflake,
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
}

func (c *Client) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.WithField("sys", "CONN").Info("Resumed connection")
	c.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (c *Client) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithField("sys", "CONN").Info("Connection Ready")

	if err := s.UpdateStatus(0, localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "DiscordStatus",
			Other: "!pb help",
		}}),
	); err != nil {
		log.WithError(err).Error("failed to update bot status")
	}

	guilds := make([]types.BaseAccount, len(s.State.Guilds))
	for i, guild := range s.State.Guilds {
		guilds[i] = types.BaseAccount{GuildSnowflake: guild.ID, OwnerSnowflake: guild.OwnerID}
	}
	if err := c.as.RemoveNotInDiscordGuildList(guilds); err != nil {
		log.WithError(err).Error("could not sync discord guilds")
	}
	c.status <- true
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (c *Client) getUserByName(guildSnowflake, name string) (discordgo.User, error) {
	guild, err := c.session.State.Guild(guildSnowflake)
	if err != nil {
		return discordgo.User{}, fmt.Errorf("guild %s not found searching for user %s", guildSnowflake, name)
	}

	for _, user := range guild.Members {
		if strings.ToLower(user.User.String()) == strings.ToLower(name) {
			return *user.User, nil
		}
	}

	return discordgo.User{}, fmt.Errorf("discord user not found %s", name)
}

func (c *Client) getDiscordAuth(snowflake string) (types.DiscordAuth, error) {
	return c.das.GetByDiscordID(snowflake)
}
