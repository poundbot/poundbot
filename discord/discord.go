package discord

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"

	"github.com/bwmarrin/discordgo"
	cache "github.com/patrickmn/go-cache"
)

const logSymbol = "🏟️ "
const logRunnerSymbol = logSymbol + "🏃 "

type RunnerConfig struct {
	Token       string
	LinkChan    string
	StatusChan  string
	GeneralChan string
}

type Client struct {
	session          *discordgo.Session
	linkChanID       string
	statusChanID     string
	generalChanID    string
	token            string
	status           chan bool
	userCache        *cache.Cache
	authRequestCache *cache.Cache
	LinkChan         chan string
	StatusChan       chan string
	GeneralChan      chan string
	GeneralOutChan   chan types.ChatMessage
	RaidAlertChan    chan types.RaidNotification
	DiscordAuth      chan types.DiscordAuth
	AuthSuccess      chan types.DiscordAuth
	shutdown         bool
}

func Runner(rc *RunnerConfig) *Client {
	return &Client{
		linkChanID:       rc.LinkChan,
		statusChanID:     rc.StatusChan,
		generalChanID:    rc.GeneralChan,
		token:            rc.Token,
		userCache:        cache.New(5*time.Minute, 10*time.Minute),
		authRequestCache: cache.New(60*time.Minute, 24*time.Hour),
		LinkChan:         make(chan string),
		StatusChan:       make(chan string),
		GeneralChan:      make(chan string),
		GeneralOutChan:   make(chan types.ChatMessage),
		DiscordAuth:      make(chan types.DiscordAuth),
		AuthSuccess:      make(chan types.DiscordAuth),
		RaidAlertChan:    make(chan types.RaidNotification),
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

		c.status = make(chan bool)

		go c.runner()

		c.connect()
	}
	return err
}

func (c *Client) Close() {
	log.Println(logSymbol + "🛑 Disconnecting")
	c.shutdown = true
	c.session.Close()
}

func (c *Client) runner() {
	defer func() {
		log.Println(logRunnerSymbol + " Runner Exiting")
	}()
	connectedState := false

	for {
		if connectedState {
			log.Println(logRunnerSymbol + " Waiting for messages")
		Reading:
			for {
				select {
				case connectedState = <-c.status:
					if !connectedState {
						log.Println(logRunnerSymbol + "☎️ Received disconnected message")
						if c.shutdown {
							return
						}
						break Reading
					} else {
						log.Println(logRunnerSymbol + "❓ Received unexpected connected message")
					}
				case t := <-c.LinkChan:
					_, err := c.session.ChannelMessageSend(
						c.linkChanID,
						fmt.Sprintf("📝 @everyone New Update: %s", t),
					)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error sending to channel: %v\n", err)
					}

				case t := <-c.StatusChan:
					_, err := c.session.ChannelMessageSend(
						c.statusChanID,
						fmt.Sprintf(logRunnerSymbol+t),
					)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error sending to channel: %v\n", err)
					}
				case t := <-c.RaidAlertChan:
					user, err := c.getUser(t.DiscordID)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error finding user %s: %s\n", t.DiscordID, err)
						break
					}

					channel, err := c.session.UserChannelCreate(user.ID)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error creating user channel: %v", err)
					} else {
						c.session.ChannelMessageSend(channel.ID, t.String())
					}
				case t := <-c.DiscordAuth:
					_, err := c.sendPrivateMessage(t.DiscordID, `
					A request has been made for you to authenticate your ALM user.
					Enter the PIN provided in-game to validate your account.
					`)
					if err == nil {
						c.cacheDiscordAuth(t)
					}
				case t := <-c.GeneralChan:
					_, err := c.session.ChannelMessageSend(c.generalChanID, t)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error sending to channel: %v\n", err)
					}
				}
			}
		}
	Connecting:
		for {
			log.Println(logRunnerSymbol + " Waiting for connected state")
			connectedState = <-c.status
			if connectedState {
				log.Println(logRunnerSymbol + "📞 Received connected message")
				break Connecting
			} else {
				log.Println(logRunnerSymbol + "☎️ Received disconnected message")
			}
		}
	}

	// Wait for connected

}

func (d *Client) connect() {
	log.Println(logSymbol + "☎️ Connecting")
	for {
		err := d.session.Open()
		if err != nil {
			log.Println(logSymbol+"⚠️ Error connecting:", err)
			log.Println(logSymbol + "🔁 Attempting discord reconnect...")
			time.Sleep(1 * time.Second)
		} else {
			log.Println(logSymbol + "📞 ✔️ Connected!")
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// This function will be called (due to AddHandler above) when the bot receives
// the "disconnect" event from Discord.
func (d *Client) disconnected(s *discordgo.Session, event *discordgo.Disconnect) {
	d.status <- false
	log.Println(logSymbol + "☎️ Disconnected!")
}

func (d *Client) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.Println(logSymbol + "📞 Resumed!")
	d.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (d *Client) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println(logSymbol + "📞 ✔️ Ready!")
	s.UpdateStatus(0, "I'm a real boy!")

	uguilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		log.Println(logSymbol + err.Error())
		return
	}
	var foundLinkChan = false
	var foundStatusChan = false
	var foundGeneralChan = false

ChannelSearch:
	for _, g := range uguilds {
		// log.Printf(logSymbol + " %s: %s\n", g.ID, g.Name)
		channels, err := s.GuildChannels(g.ID)
		if err != nil {
			log.Println(logSymbol + err.Error())
			return
		}
		for _, c := range channels {
			if c.ID == d.linkChanID {
				log.Printf(logSymbol+"📞 ✔️ Found link channel on server %s, %s: %s\n", g.Name, c.ID, c.Name)
				foundLinkChan = true
				if c.Type != discordgo.ChannelTypeGuildText {
					log.Fatalf(logSymbol+"📞 🛑 Invalid channel type: %v\n", c.Type)
					os.Exit(3)
				}
			}
			if c.ID == d.statusChanID {
				log.Printf(logSymbol+"📞 ✔️ Found status channel on server %s, %s: %s\n", g.Name, c.ID, c.Name)
				foundStatusChan = true
				if c.Type != discordgo.ChannelTypeGuildText {
					log.Fatalf(logSymbol+"📞 🛑 Invalid channel type: %v\n", c.Type)
					os.Exit(3)
				}
			}

			if c.ID == d.generalChanID {
				log.Printf(logSymbol+"📞 ✔️ Found general channel on server %s, %s: %s\n", g.Name, c.ID, c.Name)
				foundGeneralChan = true
				if c.Type != discordgo.ChannelTypeGuildText {
					log.Fatalf(logSymbol+"📞 🛑 Invalid channel type: %v\n", c.Type)
					os.Exit(3)
				}
			}

			if foundLinkChan && foundStatusChan && foundGeneralChan {
				break ChannelSearch
			}
		}
	}

	if foundLinkChan && foundStatusChan && foundGeneralChan {
		d.status <- true
	} else {
		log.Fatalln("Could not find both link and status channels.")
		os.Exit(3)
	}
}

func (c *Client) sendPrivateMessage(discordID, message string) (m *discordgo.Message, err error) {
	user, err := c.getUser(discordID)
	if err != nil {
		log.Printf(logRunnerSymbol+" Error finding user %s: %s\n", discordID, err)
		return
	}

	channel, err := c.session.UserChannelCreate(user.ID)
	if err != nil {
		log.Printf(logRunnerSymbol+" Error creating user channel: %v", err)
		return
	} else {
		return c.session.ChannelMessageSend(
			channel.ID,
			message,
		)
	}

}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (d *Client) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!test"
	// if strings.HasPrefix(m.Content, "!test") {
	// log.Printf(logSymbol+" Message (%v) %s from %s %s on %s\n", m.Type, m.Content, m.Author.Username, m.Author.String(), m.ChannelID)
	if m.ChannelID == d.generalChanID {
		go func(ch chan types.ChatMessage, cm types.ChatMessage) {
			if len(cm.Message) > 128 {
				cm.Message = truncateString(cm.Message, 128)
				d.session.ChannelMessageSend(d.generalChanID, fmt.Sprintf("*Truncated message to %s*", cm.Message))
			}
			ch <- cm
		}(d.GeneralOutChan, types.ChatMessage{
			DisplayName: m.Author.Username,
			Message:     m.Message.Content,
			Source:      types.ChatSourceDiscord,
		})
	}
	dChan, err := d.session.Channel(m.ChannelID)
	if err != nil {
		log.Printf(logSymbol+"❓ Could not get channel data for %s\n", m.ChannelID)
		return
	} else if dChan.GuildID == "" {
		goto Interact
	}

	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			goto Interact
		}
	}

	// Break out and do not interact.
	return

Interact:
	da, err := d.getDiscordAuth(m.Author.String())
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "I don't do any interactions, yet.")
	} else {
		if pinString(da.Pin) == strings.TrimSpace(m.Content) {
			da.Ack = func(authed bool) {
				if authed {
					s.ChannelMessageSend(m.ChannelID, "You have been authenticated.")
				} else {
					s.ChannelMessageSend(m.ChannelID, "Internal error. Please try again. If the problem persists, please contact MrPoundsign")
				}
			}
			d.AuthSuccess <- da
		} else {
			s.ChannelMessageSend(m.ChannelID, "Invalid pin. Please try again.")
		}
	}
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (c *Client) getUser(id string) (user discordgo.User, err error) {
	u, found := c.userCache.Get(strings.ToLower(id))
	if found {
		user = u.(discordgo.User)
	} else {
		guilds, err := c.session.UserGuilds(100, "", "")
		if err == nil {
			for _, guild := range guilds {
				users, err := c.session.GuildMembers(guild.ID, "", 1000)
				if err != nil {
					return user, err
				}

				for _, user := range users {
					if strings.ToLower(user.User.String()) == strings.ToLower(id) {
						c.cacheUser(*user.User)
						return *user.User, nil
					}
				}
			}
		}
	}
	return
}

func (c *Client) cacheUser(u discordgo.User) {
	c.userCache.Set(u.String(), u, cache.DefaultExpiration)
}

func (c *Client) cacheDiscordAuth(da types.DiscordAuth) {
	cacheID := strings.ToLower(da.DiscordID)
	log.Printf(logRunnerSymbol+"Caching auth record %v as %s", da, cacheID)
	c.authRequestCache.Set(strings.ToLower(da.DiscordID), da, cache.DefaultExpiration)
}

func (c *Client) getDiscordAuth(discordID string) (da types.DiscordAuth, err error) {
	item, found := c.authRequestCache.Get(strings.ToLower(discordID))
	if found {
		log.Printf(logRunnerSymbol+" Found %v", item.(types.DiscordAuth))
		da = item.(types.DiscordAuth)
		return
	}
	return da, errors.New("no auth record matching pin")
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
