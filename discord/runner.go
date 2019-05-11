package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"

	"github.com/bwmarrin/discordgo"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/sirupsen/logrus"
)

const logPrefix = "[DISCORD]"
const logRunnerPrefix = logPrefix + "[RUNNER]"

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
		c.session.AddHandler(Disconnected(c.status, logPrefix))
		c.session.AddHandler(c.resumed)
		c.session.AddHandler(NewGuildCreate(c.as, c.us))
		c.session.AddHandler(NewGuildDelete(c.as))
		c.session.AddHandler(newGuildMemberAdd(c.us, c.as))
		c.session.AddHandler(newGuildMemberRemove(c.us, c.as))

		c.status = make(chan bool)

		go c.runner()

		c.connect()
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
	defer log.Println(" Runner exited")

	connectedState := false

	for {
		if connectedState {
			log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
				"Waiting for messages.",
			)
		Reading:
			for {
				select {
				case connectedState = <-c.status:
					if !connectedState {
						log.WithFields(logrus.Fields{"ssys": "CONN"}).Warn(
							"Received disconnected message",
						)
						if c.shutdown {
							return
						}
						break Reading
					}

					log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
						"Received unexpected connected message",
					)

				case t := <-c.RaidAlertChan:
					go func() {
						raUser, err := c.us.GetByPlayerID(t.PlayerID)
						if err != nil {
							log.WithFields(logrus.Fields{"ssys": "RAID", "PlayerID": t.PlayerID}).WithError(err).Error(
								"Player not found trying to send raid alert",
							)
							return
						}

						user, err := c.session.User(raUser.Snowflake)
						if err != nil {
							log.WithFields(logrus.Fields{"ssys": "RAID", "Snowflake": raUser.Snowflake}).WithError(err).Error(
								"Discord user not found trying to send raid alert",
							)
							return
						}

						c.sendPrivateMessage(user.ID, t.String())
					}()

				case t := <-c.DiscordAuth:
					dUser, err := c.getUserByName(t.GuildSnowflake, t.DiscordInfo.DiscordName)
					if err != nil {
						log.WithFields(logrus.Fields{"ssys": "DAUTH", "guildid": t.GuildSnowflake, "name": t.DiscordInfo.DiscordName}).WithError(err).Error(
							"Discord user not found",
						)
						err = c.das.Remove(t)
						if err != nil {
							log.WithFields(logrus.Fields{"ssys": "DAUTH", "guildid": t.GuildSnowflake, "playerid": t.PlayerID}).WithError(err).Error(
								"Error removing discord auth for PlayerID from the database.",
							)
						}
						break
					}

					t.Snowflake = dUser.ID

					err = c.das.Upsert(t)
					if err != nil {
						log.WithFields(logrus.Fields{"ssys": "DAUTH", "guildid": t.GuildSnowflake, "name": t.DiscordInfo.DiscordName}).WithError(err).Error(
							"Error upserting PlayerID ito the database",
						)
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
						log.WithFields(logrus.Fields{"ssys": "CHAT", "discordid": t.Snowflake}).WithError(err).Error(
							"Could not send PIN request to user",
						)
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
			log.Println("Waiting for connected state...")
			connectedState = <-c.status
			if connectedState {
				log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
					"Received connected message",
				)
				break Connecting
			}
			log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
				"Received disconnected message",
			)
		}
	}

}

func (c *Client) connect() {
	log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
		"Connecting",
	)
	for {
		err := c.session.Open()
		if err != nil {
			log.WithFields(logrus.Fields{"ssys": "CONN"}).WithError(err).Warn(
				"Error connecting",
			)
			log.WithFields(logrus.Fields{"ssys": "CONN"}).Warn(
				"Attempting Reconnect...",
			)
			time.Sleep(1 * time.Second)
		} else {
			log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
				"Connected",
			)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *Client) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
		"Resumed connection",
	)
	c.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (c *Client) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.WithFields(logrus.Fields{"ssys": "CONN"}).Info(
		"Connection Ready",
	)
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

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (c *Client) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
		log.Printf(logPrefix+" Could not get account for guild %s\n", m.GuildID)
		return
	}

	if account.OwnerSnowflake == "" {
		log.WithFields(logrus.Fields{"ssys": "RUNNER", "guildid": m.GuildID}).Info(
			"Guild is missing owner",
		)
		guild, err := s.Guild(m.GuildID)
		if err != nil {
			// TODO handle not finding the guild here
			return
		}

		log.WithFields(logrus.Fields{"ssys": "RUNNER", "guildid": m.GuildID, "ownersnowflake": guild.OwnerID}).Info(
			"Setting owner",
		)
		account.OwnerSnowflake = guild.OwnerID
		c.as.UpsertBase(account.BaseAccount)
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
			err = c.sendChannelMessage(m.ChannelID, response.message)
		}
		return
	}

	if len(account.Servers) == 0 {
		return
	}

	// Find the server for the channel and send the message to it
	for _, server := range account.Servers {
		if server.ChatChanID == m.ChannelID {
			cm := types.ChatMessage{
				ServerKey:   server.Key,
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
					c.session.ChannelMessageSend(m.ChannelID, localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "TruncatedMessage",
							Other: "*Truncated message to {{.Message}}",
						},
						TemplateData: map[string]string{"Message": cm.Message},
					}))
				}
				err = c.cqs.InsertMessage(cm)
				if err != nil {
					log.Printf("discord: Could not insert message: %v", err)
				}
			}()
			return
		}
	}
}

func (c Client) sendChannelMessage(channelID, message string) error {
	_, err := c.session.ChannelMessageSend(channelID, message)
	return err
}

func (c Client) sendPrivateMessage(snowflake, message string) error {
	channel, err := c.session.UserChannelCreate(snowflake)

	if err != nil {
		log.Printf(logRunnerPrefix+" Error creating user channel: %v", err)
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
