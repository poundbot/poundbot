package discord

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

type discord struct {
	session        *discordgo.Session
	twitterChannel string
	token          string
	status         chan bool
	// kill           chan bool
	LinkChan chan string
}

func DiscordRunner(token, channel string) *discord {
	return &discord{
		twitterChannel: channel,
		token:          token,
		LinkChan:       make(chan string),
	}
}

func (d *discord) Start() error {
	session, err := discordgo.New("Bot " + d.token)
	if err == nil {
		d.session = session
		d.session.AddHandler(d.messageCreate)
		d.session.AddHandler(d.ready)
		d.session.AddHandler(d.disconnected)

		d.status = make(chan bool)

		go d.runner()

		d.connect()
	}
	return err
}

func (d *discord) Close() {
	log.Println("🏟️🛑 Disconnecting")
	d.session.Close()
}

func (d *discord) runner() {
	defer func() {
		log.Println("🏟️🏃 Runner Exiting")
	}()
	connectedState := false

	for {
		if connectedState {
			log.Println("🏟️🏃 Waiting for messages")
		Reading:
			for {
				select {
				case connectedState = <-d.status:
					log.Println("🏟️🏃⚠️ Received disconnected message ")
					break Reading
				case t := <-d.LinkChan:
					_, err := d.session.ChannelMessageSend(
						d.twitterChannel,
						fmt.Sprintf("📝 @everyone New Update: %s", t),
					)
					if err != nil {
						log.Printf("🏟️🏃 Error sending to channel: %v\n", err)
					}
				}
			}
		}

		log.Println("🏟️🏃 Waiting for connected state")

		// Wait for connected
	Connecting:
		for {
			select {
			case connectedState = <-d.status:
				if connectedState {
					log.Println("🏟️🏃 Connected")
					break Connecting
				} else {
					log.Println("🏟️🏃 Disconnected")
				}
			}

			time.Sleep(1 * time.Second)
		}
	}

}

func (d *discord) connect() {
	log.Println("🏟️⚪ Connecting")
	d.status <- false
	for {
		err := d.session.Open()
		if err != nil {
			log.Println("🏟️⚠️ Error connecting: ", err)
			log.Println("🏟️🔁 Attempting discord reconnect...")
			time.Sleep(1 * time.Second)
		} else {
			log.Println("🏟️✔️ Connected!")
			return
		}
	}
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (d *discord) disconnected(s *discordgo.Session, event *discordgo.Disconnect) {
	log.Println("🏟️🛑 Disconnected!")
	d.connect()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (d *discord) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("🏟️✔️ Ready!")
	s.UpdateStatus(0, "I'm a real boy!")

	uguilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		log.Println(err)
		return
	}

ChannelSearch:
	for _, g := range uguilds {
		// log.Printf("🏟️ %s: %s\n", g.ID, g.Name)
		channels, err := s.GuildChannels(g.ID)
		if err != nil {
			log.Println(err)
			return
		}
		for _, c := range channels {
			if c.ID == d.twitterChannel {
				log.Printf("🏟️✔️ Found channel on server %s, %s: %s\n", g.Name, c.ID, c.Name)
				if c.Type != discordgo.ChannelTypeGuildText {
					log.Printf("🏟️🛑 Invalid channel type: %v", c.Type)
					os.Exit(3)
				}
				d.status <- true
				break ChannelSearch
			}
		}
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
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			log.Printf("🏟️ Message %s from %s\n", m.Content, m.ChannelID)
			s.ChannelMessageSend(m.ChannelID, "I don't do any interactions, yet.")
			for _, embed := range m.Embeds {
				log.Println(embed.Type)
			}
			return
		}
	}
	// }
}
