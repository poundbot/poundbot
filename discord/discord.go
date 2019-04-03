package discord

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/chatcache"
	"bitbucket.org/mrpoundsign/poundbot/discord/handler"
	"bitbucket.org/mrpoundsign/poundbot/messages"
	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/types"

	"github.com/bwmarrin/discordgo"
	uuid "github.com/satori/go.uuid"
)

const logPrefix = "[DISCORD]"
const logRunnerPrefix = logPrefix + "[RUNNER]"

type RunnerConfig struct {
	Token string
}

type Client struct {
	session       *discordgo.Session
	as            storage.AccountsStore
	cc            chatcache.ChatCache
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

func Runner(token string, cc chatcache.ChatCache, as storage.AccountsStore, das storage.DiscordAuthsStore, us storage.UsersStore) *Client {
	return &Client{
		as:            as,
		cc:            cc,
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
		c.session.AddHandler(handler.Disconnected(c.status, logPrefix))
		c.session.AddHandler(c.resumed)
		c.session.AddHandler(handler.NewGuildCreate(c.as))
		c.session.AddHandler(handler.NewGuildDelete(c.as))

		c.status = make(chan bool)

		go c.runner()

		c.connect()
	}
	return err
}

// Stop stops the runner
func (c *Client) Stop() {
	log.Println(logPrefix + "[CONN] Disconnecting")
	c.shutdown = true
	c.session.Close()
}

func (c *Client) runner() {
	defer func() {
		log.Println(logRunnerPrefix + " Runner Exiting")
	}()
	connectedState := false

	for {
		if connectedState {
			log.Println(logRunnerPrefix + " Waiting for messages")
		Reading:
			for {
				select {
				case connectedState = <-c.status:
					if !connectedState {
						log.Println(logRunnerPrefix + "[CONN] Received disconnected message")
						if c.shutdown {
							return
						}
						break Reading
					} else {
						log.Println(logRunnerPrefix + "[CONN] Received unexpected connected message")
					}

				case t := <-c.RaidAlertChan:
					var u types.User

					err := c.us.Get(t.SteamID, &u)
					if err != nil {
						log.Printf(logRunnerPrefix + "[COMM] User not found trying to send raid alert")
						break
					}

					user, err := c.session.User(u.Snowflake)
					if err != nil {
						log.Printf(logRunnerPrefix+"[COMM] Error finding user %d: %d\n", t.SteamID, err)
						break
					}

					channel, err := c.session.UserChannelCreate(user.ID)
					if err != nil {
						log.Printf(logRunnerPrefix+"[COMM] Error creating user channel: %v", err)
					} else {
						c.session.ChannelMessageSend(channel.ID, t.String())
					}

				case t := <-c.DiscordAuth:
					dUser, err := c.getUserByName(t.DiscordInfo.DiscordName)
					if err != nil {
						log.Printf(logRunnerPrefix+"[COMM] User %s not found\n", t.DiscordInfo.DiscordName)
						err = c.das.Remove(t.SteamInfo)
						if err != nil {
							log.Printf(logRunnerPrefix+"[DB] - Error removing SteamID %d from the database\n", t.SteamInfo.SteamID)
						}
						break
					}

					t.BaseUser.Snowflake = dUser.ID

					err = c.das.Upsert(t)
					if err != nil {
						log.Printf(logRunnerPrefix+"[DB] - Error upserting SteamID %d from the database\n", t.SteamInfo.SteamID)
						break
					}

					_, err = c.sendPrivateMessage(t.Snowflake, messages.PinPrompt)
					if err != nil {
						log.Println(logRunnerPrefix + "[COMM] Could not send PIN request to user")
					}

				case t := <-c.ChatChan:
					var clan = ""
					if t.ClanTag != "" {
						clan = fmt.Sprintf("[%s] ", t.ClanTag)
					}
					_, err := c.session.ChannelMessageSend(t.ChannelID, fmt.Sprintf("☢️ @%s **%s%s**: %s", t.Timestamp.CreatedAt.Format("2006-01-02 15:04:05 MST"), clan, t.DisplayName, t.Message))
					if err != nil {
						log.Printf(logRunnerPrefix+"[COMM] Error sending to channel: %v\n", err)
					}
				}
			}
		}
	Connecting:
		for {
			log.Println(logRunnerPrefix + " Waiting for connected state")
			connectedState = <-c.status
			if connectedState {
				log.Println(logRunnerPrefix + "[CONN] Received connected message")
				break Connecting
			} else {
				log.Println(logRunnerPrefix + "[CONN] Received disconnected message")
			}
		}
	}

	// Wait for connected

}

func (c *Client) connect() {
	log.Println(logPrefix + "[CONN] Connecting")
	for {
		err := c.session.Open()
		if err != nil {
			log.Println(logPrefix+"[CONN][WARN] Error connecting:", err)
			log.Println(logPrefix + "[CONN] Attempting discord reconnect...")
			time.Sleep(1 * time.Second)
		} else {
			log.Println(logPrefix + "[CONN] Connected!")
			return
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *Client) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.Println(logPrefix + "[CONN] Resumed!")
	c.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (c *Client) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println(logPrefix + "[CONN] Ready!")
	s.UpdateStatus(0, "I'm a real boy!")
	guilds := make([]string, len(s.State.Guilds))
	for i, guild := range s.State.Guilds {
		guilds[i] = guild.ID
	}
	c.as.RemoveNotInDiscordGuildList(guilds)
	c.status <- true
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (c *Client) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

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

	var account types.Account
	err = c.as.GetByDiscordGuild(m.GuildID, &account)
	if err != nil {
		log.Printf(logPrefix+"Could not get account for %s\n", m.GuildID)
		return
	}

	// Detect mention
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			c.instruct(s, m, account)
			return
		}
	}

	if len(account.Servers) == 0 {
		return
	}

	// Find the server for the channel and send the message to it
	for _, server := range account.Servers {
		if server.ChatChanID == m.ChannelID {
			go func(cm types.ChatMessage, cc chan types.ChatMessage) {
				if len(cm.Message) > 128 {
					cm.Message = truncateString(cm.Message, 128)
					c.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("*Truncated message to %s*", cm.Message))
				}
				cm.CreatedAt = time.Now().UTC()
				select {
				case cc <- cm:
					return
				case <-time.After(10 * time.Second):
					return
				}

			}(types.ChatMessage{
				ServerKey:   server.Key,
				DisplayName: m.Author.Username,
				Message:     m.Message.Content,
				Source:      types.ChatSourceDiscord,
			}, c.cc.GetOutChannel(server.Key))
		}
	}
}

func (c *Client) interact(s *discordgo.Session, m *discordgo.MessageCreate) {
	var da types.DiscordAuth
	err := c.getDiscordAuth(m.Author.ID, &da)
	if err != nil {
		return
	}

	if pinString(da.Pin) == strings.TrimSpace(m.Content) {
		da.Ack = func(authed bool) {
			if authed {
				s.ChannelMessageSend(m.ChannelID, "You have been authenticated.")
			} else {
				s.ChannelMessageSend(m.ChannelID, "Internal error. Please try again. If the problem persists, please contact MrPoundsign")
			}
		}
		c.AuthSuccess <- da
	} else {
		s.ChannelMessageSend(m.ChannelID, "Invalid pin. Please try again.")
	}

	return
}

func (c *Client) instruct(s *discordgo.Session, m *discordgo.MessageCreate, account types.Account) {
	log.Printf("Instruct: %s, %s", account.GuildSnowflake, m.ContentWithMentionsReplaced())
	parts := strings.Fields(
		strings.Replace(m.Content, fmt.Sprintf("<@%s>", s.State.User.ID), "", -1),
	)

	if m.Author.ID != account.OwnerSnowflake {
		return
	}

	if len(parts) == 0 {
		return
	}

	command := parts[0]
	parts = parts[1:]

	switch command {
	case "help":
		c.sendPrivateMessage(m.Author.ID, messages.HelpText())
		break
	case "server":
		if len(parts) == 0 {
			c.session.ChannelMessageSend(m.ChannelID, "TODO: Server Usage. See `help`.")
			return
		}

		switch parts[0] {
		case "list":
			out := "**Server List**\nID : Name : RaidDelay : Key\n----"
			for i, server := range account.Servers {
				out = fmt.Sprintf("%s\n%d : %s : %s : ||`%s`||", out, i, server.RaidDelay, server.Name, server.Key)
			}
			c.sendPrivateMessage(m.Author.ID, out)
			return
		case "add":
			if len(parts) < 2 {
				c.session.ChannelMessageSend(m.ChannelID, "Usage: `server add <name>`")
				return
			}
			server := types.Server{
				Name:       strings.Join(parts[1:], " "),
				Key:        uuid.NewV4().String(),
				ChatChanID: m.ChannelID,
				RaidDelay:  "1m",
			}
			c.as.AddServer(account.GuildSnowflake, server)
			c.sendServerKey(m.Author.ID, server.Key)
			return
		}

		serverID := 0
		instructions := parts
		serverID, err := strconv.Atoi(instructions[0])
		if err == nil {
			instructions = instructions[1:]
		} else if len(account.Servers) > 1 {
			c.session.ChannelMessageSend(m.ChannelID, "You have multiple servers. Use server `#`.")
			return
		}

		switch instructions[0] {
		case "reset":
			if len(account.Servers) <= serverID {
				c.session.ChannelMessageSend(m.ChannelID, "Server not defined. Try `help`")
				return
			}
			oldKey := account.Servers[serverID].Key
			account.Servers[serverID].Key = uuid.NewV4().String()
			c.as.UpdateServer(account.GuildSnowflake, oldKey, account.Servers[serverID])
			c.sendServerKey(m.Author.ID, account.Servers[serverID].Key)
			return
		case "rename":
			if len(account.Servers) <= serverID {
				c.session.ChannelMessageSend(m.ChannelID, "Server not defined. Try `help`")
				return
			}
			if len(instructions) < 2 {
				c.session.ChannelMessageSend(m.ChannelID, "Usage: `server rename [id] <name>`")
				return
			}
			account.Servers[serverID].Name = strings.Join(instructions[1:], " ")
			c.as.UpdateServer(account.GuildSnowflake, account.Servers[serverID].Key, account.Servers[serverID])
			c.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Server %d name set to %s", serverID, account.Servers[serverID].Name))
			return
		case "delete":
			if len(account.Servers) <= serverID {
				c.session.ChannelMessageSend(m.ChannelID, "Server not defined. Try `help`")
				return
			}
			c.as.RemoveServer(account.GuildSnowflake, account.Servers[serverID].Key)
			c.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Server %d(%s) removed", serverID, account.Servers[serverID].Name))
			return
		case "chathere":
			if len(account.Servers) <= serverID {
				c.session.ChannelMessageSend(m.ChannelID, "Server not defined. Try `help`")
				return
			}
			account.Servers[serverID].ChatChanID = m.ChannelID
			c.as.UpdateServer(account.GuildSnowflake, account.Servers[serverID].Key, account.Servers[serverID])
			return
		case "raiddelay":
			if len(account.Servers) <= serverID {
				c.session.ChannelMessageSend(m.ChannelID, "Server not defined. Try `help`")
				return
			}
			if len(instructions) != 2 {
				c.session.ChannelMessageSend(m.ChannelID, "Usage: `server rename [id] <name>`")
				return
			}
			_, err := time.ParseDuration(instructions[1])
			if err != nil {
				c.session.ChannelMessageSend(m.ChannelID, "Invalid duration format. Examples:\n`1m` = 1 minute, `1h` = 1 hour, `1s` = 1 second")
				return
			}

			account.Servers[serverID].RaidDelay = instructions[1]
			c.as.UpdateServer(account.GuildSnowflake, account.Servers[serverID].Key, account.Servers[serverID])

			return
		}

		c.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Invalid command %s. Try `help`", instructions[0]))
	}
}

func (c *Client) sendPrivateMessage(snowflake, message string) (m *discordgo.Message, err error) {
	channel, err := c.session.UserChannelCreate(snowflake)

	if err != nil {
		log.Printf(logRunnerPrefix+" Error creating user channel: %v", err)
		return
	}

	return c.session.ChannelMessageSend(
		channel.ID,
		message,
	)
}

func (c *Client) sendServerKey(snowflake, key string) (m *discordgo.Message, err error) {
	message := messages.ServerKeyMessage(key)
	return c.sendPrivateMessage(snowflake, message)
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (c *Client) getUserByName(name string) (user discordgo.User, err error) {

	guilds, err := c.session.UserGuilds(100, "", "")
	if err != nil {
		return discordgo.User{}, fmt.Errorf("discord user not found %s", name)
	}

	for _, guild := range guilds {
		users, err := c.session.GuildMembers(guild.ID, "", 1000)
		if err != nil {
			return user, err
		}

		for _, user := range users {
			if strings.ToLower(user.User.String()) == strings.ToLower(name) {
				return *user.User, nil
			}
		}
	}

	return discordgo.User{}, fmt.Errorf("discord user not found %s", name)
}

func (c *Client) getDiscordAuth(snowflake string, da *types.DiscordAuth) error {
	return c.das.GetSnowflake(snowflake, da)
}

func pinString(pin int) string {
	return fmt.Sprintf("%04d", pin)
}

func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		bnoden = str[0:num] + "..."
	}
	return bnoden
}
