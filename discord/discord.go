package discord

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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
	// LinkChan    string
	// StatusChan  string
	// GeneralChan string
}

type Client struct {
	session *discordgo.Session
	// linkChanID     string
	// statusChanID   string
	// generalChanID  string
	as     storage.AccountsStore
	cs     storage.ChatsStore
	das    storage.DiscordAuthsStore
	us     storage.UsersStore
	token  string
	status chan bool
	// LinkChan       chan string
	// StatusChan     chan string
	GeneralChan chan types.ChatMessage
	// GeneralOutChan chan types.ChatMessage
	RaidAlertChan chan types.RaidAlert
	DiscordAuth   chan types.DiscordAuth
	AuthSuccess   chan types.DiscordAuth
	shutdown      bool
}

func Runner(token string, as storage.AccountsStore, cs storage.ChatsStore, das storage.DiscordAuthsStore, us storage.UsersStore) *Client {
	return &Client{
		as:    as,
		cs:    cs,
		das:   das,
		us:    us,
		token: token,
		// LinkChan:       make(chan string),
		// StatusChan:     make(chan string),
		GeneralChan: make(chan types.ChatMessage),
		// GeneralOutChan: make(chan types.ChatMessage),
		DiscordAuth:   make(chan types.DiscordAuth),
		AuthSuccess:   make(chan types.DiscordAuth),
		RaidAlertChan: make(chan types.RaidAlert),
	}
}

func (c *Client) Start() error {
	session, err := discordgo.New("Bot " + c.token)
	if err == nil {
		c.session = session
		c.session.AddHandler(c.messageCreate)
		c.session.AddHandler(c.ready)
		c.session.AddHandler(c.disconnected)
		c.session.AddHandler(c.resumed)
		c.session.AddHandler(handler.NewGuildCreate(c.as))
		c.session.AddHandler(handler.NewGuildDelete(c.as))

		c.status = make(chan bool)

		go c.runner()

		c.connect()
	}
	return err
}

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
						log.Printf(logRunnerPrefix + "User not found")
						break Reading
					}

					user, err := c.session.User(u.Snowflake)
					if err != nil {
						log.Printf(logRunnerPrefix+" Error finding user %d: %d\n", t.SteamID, err)
						break Reading
					}

					channel, err := c.session.UserChannelCreate(user.ID)
					if err != nil {
						log.Printf(logRunnerPrefix+" Error creating user channel: %v", err)
					} else {
						c.session.ChannelMessageSend(channel.ID, t.String())
					}

				case t := <-c.DiscordAuth:
					dUser, err := c.getUserByName(t.DiscordInfo.DiscordName)
					if err != nil {
						log.Printf(logRunnerPrefix+"User %s not found\n", t.DiscordInfo.DiscordName)
						err = c.das.Remove(t.SteamInfo)
						if err != nil {
							log.Printf(logRunnerPrefix+" - Error removing SteamID %d from the database\n", t.SteamInfo.SteamID)
						}
						break Reading
					}

					t.BaseUser.Snowflake = dUser.ID

					err = c.das.Upsert(t)
					if err != nil {
						log.Printf(logRunnerPrefix+" - Error upserting SteamID %d from the database\n", t.SteamInfo.SteamID)
						break Reading
					}

					_, err = c.sendPrivateMessage(t.Snowflake, messages.PinPrompt)
					if err != nil {
						log.Println(logRunnerPrefix + "Could not send PIN request to user")
					}

				case t := <-c.GeneralChan:
					var clan = ""
					if t.ClanTag != "" {
						clan = fmt.Sprintf("[%s] ", t.ClanTag)
					}
					_, err := c.session.ChannelMessageSend(t.ChannelID, fmt.Sprintf("☢️ **%s%s**: %s", clan, t.DisplayName, t.Message))
					if err != nil {
						log.Printf(logRunnerPrefix+" Error sending to channel: %v\n", err)
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

// This function will be called (due to AddHandler above) when the bot receives
// the "disconnect" event from Discord.
func (c *Client) disconnected(s *discordgo.Session, event *discordgo.Disconnect) {
	c.status <- false
	log.Println(logPrefix + "[CONN] Disconnected!")
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

	var da types.DiscordAuth
	var account types.Account
	var server types.Server
	var err error

	// Detect PM
	if m.GuildID == "" {
		goto Interact
	}

	err = c.as.GetByDiscordGuild(m.GuildID, &account)
	if err != nil {
		log.Printf(logPrefix+"Could not get account for %s\n", m.GuildID)
		return
	}

	// Detect mention
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			goto Instruct
		}
	}

	if len(account.Servers) == 0 {
		return
	}

	// TODO: Get the actual server
	server = account.Servers[0]

	if server.ChatChanID == m.ChannelID {
		go func(cm types.ChatMessage) {
			if len(cm.Message) > 128 {
				cm.Message = truncateString(cm.Message, 128)
				c.session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("*Truncated message to %s*", cm.Message))
			}
			c.cs.Log(cm)

		}(types.ChatMessage{
			ServerKey:   server.Key,
			DisplayName: m.Author.Username,
			Message:     m.Message.Content,
			Source:      types.ChatSourceDiscord,
		})
	}

	// Break out and do not interact.
	return

Interact:
	err = c.getDiscordAuth(m.Author.ID, &da)
	if err != nil {
		return
	} else {
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
	}
	return

Instruct:
	log.Printf("Instruct: %s, %s", account.GuildSnowflake, m.ContentWithMentionsReplaced())
	message := strings.Trim(
		strings.Replace(m.Content, fmt.Sprintf("<@%s>", s.State.User.ID), "", -1),
		"\n ",
	)

	switch message {
	case "help":
		c.sendPrivateMessage(m.Author.ID, messages.HelpText)
		break
	case "server init":
		if len(account.Servers) > 0 {
			c.sendPrivateMessage(m.Author.ID, "You already have a server")
			return
		}

		account.Servers = []types.Server{
			types.Server{Key: uuid.NewV4().String(), ChatChanID: m.ChannelID, RaidDelay: "1m"},
		}

		c.as.AddServer(account.GuildSnowflake, account.Servers[0])
		c.sendServerKey(m.Author.ID, account.Servers[0].Key)
		return
	case "server reset":
		if len(account.Servers) < 1 {
			c.sendPrivateMessage(m.Author.ID, "You don't have a server defined. Try *help*")
			return
		}
		account.Servers[0].Key = uuid.NewV4().String()
		c.as.UpdateServer(account.GuildSnowflake, account.Servers[0])
		c.sendServerKey(m.Author.ID, account.Servers[0].Key)
		return
	case "server chat here":
		if len(account.Servers) < 1 {
			c.sendPrivateMessage(m.Author.ID, "You don't have a server defined. Try *help*")
			return
		}
		account.Servers[0].ChatChanID = m.ChannelID
		c.as.UpdateServer(account.GuildSnowflake, account.Servers[0])
		return
	}
}

func (c *Client) sendPrivateMessage(snowflake, message string) (m *discordgo.Message, err error) {
	channel, err := c.session.UserChannelCreate(snowflake)
	if err != nil {
		log.Printf(logRunnerPrefix+" Error creating user channel: %v", err)
		return
	} else {
		return c.session.ChannelMessageSend(
			channel.ID,
			message,
		)
	}
}

func (c *Client) sendServerKey(snowflake, u1 string) (m *discordgo.Message, err error) {
	return c.sendPrivateMessage(snowflake, messages.ServerKeyMessage(u1))
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (c *Client) getUserByName(name string) (user discordgo.User, err error) {

	guilds, err := c.session.UserGuilds(100, "", "")
	if err != nil {
		return discordgo.User{}, errors.New(fmt.Sprintf("discord user not found %s", name))
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

	return discordgo.User{}, errors.New(fmt.Sprintf("discord user not found %s", name))
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
