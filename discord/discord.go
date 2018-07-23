package discord

import (
	"fmt"
	"log"
	"os"
	"time"

	"mrpoundsign.com/poundbot/rustconn"

	"github.com/bwmarrin/discordgo"
	cache "github.com/patrickmn/go-cache"
)

const logSymbol = "🏟️ "
const logRunnerSymbol = logSymbol + "🏃 "

type RunnerConfig struct {
	Token      string
	LinkChan   string
	StatusChan string
}

type discord struct {
	session       *discordgo.Session
	linkChanID    string
	statusChanID  string
	token         string
	status        chan bool
	userCache     *cache.Cache
	LinkChan      chan string
	StatusChan    chan string
	RaidAlertChan chan rustconn.RaidNotification
	GetIDChan     chan string
}

func DiscordRunner(rc *RunnerConfig) *discord {
	return &discord{
		linkChanID:    rc.LinkChan,
		statusChanID:  rc.StatusChan,
		token:         rc.Token,
		userCache:     cache.New(5*time.Minute, 10*time.Minute),
		LinkChan:      make(chan string),
		StatusChan:    make(chan string),
		RaidAlertChan: make(chan rustconn.RaidNotification),
		GetIDChan:     make(chan string),
	}
}

func (d *discord) Start() error {
	session, err := discordgo.New("Bot " + d.token)
	if err == nil {
		d.session = session
		d.session.AddHandler(d.messageCreate)
		d.session.AddHandler(d.ready)
		d.session.AddHandler(d.disconnected)
		d.session.AddHandler(d.resumed)

		d.status = make(chan bool)

		go d.runner()

		d.connect()
	}
	return err
}

func (d *discord) Close() {
	log.Println(logSymbol + "🛑 Disconnecting")
	d.session.Close()
}

func (d *discord) runner() {
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
				case connectedState = <-d.status:
					if !connectedState {
						log.Println(logRunnerSymbol + "☎️ Received disconnected message")
						break Reading
					} else {
						log.Println(logRunnerSymbol + "❓ Received unexpected connected message")
					}
				case t := <-d.LinkChan:
					_, err := d.session.ChannelMessageSend(
						d.linkChanID,
						fmt.Sprintf("📝 @everyone New Update: %s", t),
					)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error sending to channel: %v\n", err)
					}

				case t := <-d.StatusChan:
					_, err := d.session.ChannelMessageSend(
						d.statusChanID,
						fmt.Sprintf(logRunnerSymbol+t),
					)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error sending to channel: %v\n", err)
					}
				case t := <-d.RaidAlertChan:
					log.Printf(logRunnerSymbol+" Got raid alert: %v", t)
					user, err := d.getUser(t.DiscordID)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error finding user %s: %s\n", t.DiscordID, err)
						break
					}

					// var toChannel discordgo.Channel;
					// channels, err := d.session.UserChannels()
					// if err != nil {
					// 	log.Printf(logRunnerSymbol+" Error loading user channels: %v", err)
					// 	continue
					// }
					// for _, channel := range channels {
					// 	for _
					// }
					channel, err := d.session.UserChannelCreate(user.ID)
					if err != nil {
						log.Printf(logRunnerSymbol+" Error creating user channel: %v", err)
					} else {
						d.session.ChannelMessageSend(channel.ID, fmt.Sprintf("%v", t.Items))
					}
				}
			}
		}

		log.Println(logRunnerSymbol + " Waiting for connected state")

		// Wait for connected
	Connecting:
		for {
			select {
			case connectedState = <-d.status:
				if connectedState {
					log.Println(logRunnerSymbol + "📞 Received connected message")
					break Connecting
				} else {
					log.Println(logRunnerSymbol + "☎️ Received disconnected message")
				}
			}

			time.Sleep(1 * time.Second)
		}
	}

}

func (d *discord) connect() {
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
func (d *discord) disconnected(s *discordgo.Session, event *discordgo.Disconnect) {
	d.status <- false
	log.Println(logSymbol + "☎️ Disconnected!")
}

func (d *discord) resumed(s *discordgo.Session, event *discordgo.Resumed) {
	log.Println(logSymbol + "📞 Resumed!")
	d.status <- true
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (d *discord) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println(logSymbol + "📞 ✔️ Ready!")
	s.UpdateStatus(0, "I'm a real boy!")

	uguilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		log.Println(err)
		return
	}
	var foundLinkChan = false
	var foundStatusChan = false

ChannelSearch:
	for _, g := range uguilds {
		// log.Printf(logSymbol + " %s: %s\n", g.ID, g.Name)
		channels, err := s.GuildChannels(g.ID)
		if err != nil {
			log.Println(err)
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
			if foundLinkChan && foundStatusChan {
				break ChannelSearch
			}
			// log.Printf("%s, %s", c.Name, c.ID)
		}
	}

	if foundLinkChan && foundStatusChan {
		d.status <- true
	} else {
		log.Fatalln("Could not find both link and status channels.")
		os.Exit(3)
	}
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func (d *discord) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// check if the message is "!test"
	// if strings.HasPrefix(m.Content, "!test") {
	log.Printf(logSymbol+" Message (%v) %s from %s %s on %s\n", m.Type, m.Content, m.Author.Username, m.Author.String(), m.ChannelID)
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
	s.ChannelMessageSend(m.ChannelID, "I don't do any interactions, yet.")
}

// Returns nil user if they don't exist; Returns error if there was a communications error
func (d *discord) getUser(id string) (user discordgo.User, err error) {
	u, found := d.userCache.Get(id)
	if found {
		user = u.(discordgo.User)
	} else {
		guilds, err := d.session.UserGuilds(100, "", "")
		if err == nil {
			for _, guild := range guilds {
				users, err := d.session.GuildMembers(guild.ID, "", 1000)
				if err != nil {
					return user, err
				}

				for _, user := range users {
					if user.User.String() == id {
						d.cacheUser(*user.User)
						return *user.User, nil
					}
				}
			}
		}
	}
	return
}

func (d *discord) cacheUser(u discordgo.User) {
	d.userCache.Set(u.String(), u, cache.DefaultExpiration)
}
