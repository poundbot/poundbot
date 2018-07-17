package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/go-twitter/twitter"
)

type discord struct {
	session        *discordgo.Session
	twitterChannel string
	token          string
	status         chan bool
	// kill           chan bool
	TweetChan chan *twitter.Tweet
}

func DiscordRunner(token, channel string) *discord {
	return &discord{
		twitterChannel: channel,
		token:          token,
		TweetChan:      make(chan *twitter.Tweet),
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
		// d.kill = make(chan bool)

		go d.runner()

		d.connect()
	}
	return err
}

func (d *discord) Close() {
	fmt.Println("🛑 DISCORD: Closing")
	// d.kill <- true
	fmt.Println("🛑 DISCORD: Disconnecting")
	d.session.Close()
}

func (d *discord) runner() {
	connectedState := false
Connected:
	for {
		if connectedState {
			fmt.Println("🏃 DISCORD: Waiting for messages")
			for {
				select {
				case connectedState = <-d.status:
					fmt.Println("🏃 DISCORD: Connection state changed to ", connectedState)
					break Connected

				// case <-d.kill:
				// 	fmt.Println("🏃 DISCORD: Exiting")
				// 	return

				case t := <-d.TweetChan:
					if strings.Contains(strings.ToLower(t.Text), "#almupdate") {
						_, err := d.session.ChannelMessageSend(
							d.twitterChannel,
							fmt.Sprintf(
								"📝 @everyone New Update: https://twitter.com/%s/status/%d",
								t.User.ScreenName,
								t.ID,
							),
						)
						if err != nil {
							fmt.Println(err)
						}
					} else {
						fmt.Println("🏃 DISCORD: Not posting tweet: ", t.Text)
					}
				}
			}
		}

		fmt.Println("🏃 DISCORD: Waiting for connected state")

		// Wait for connected
		for {
			select {
			case connectedState = <-d.status:
				if connectedState {
					fmt.Println("🏃 DISCORD: Connected")
					break Connected
				} else {
					fmt.Println("🏃 DISCORD: Disconnected")
				}
				// case <-d.kill:
				// 	fmt.Println("🏃 DISCORD: Exiting")
				// 	return
			}

			time.Sleep(1 * time.Second)
		}
	}

}

func (d *discord) connect() {
	fmt.Println("⚪ DISCORD: Connecting")
	d.status <- false
	for {
		err := d.session.Open()
		if err != nil {
			fmt.Println("⚠️ DISCORD: Error connecting: ", err)
			fmt.Println("🔁 DISCORD: Attempting discord reconnect...")
			time.Sleep(1 * time.Second)
		} else {
			fmt.Println("✔️ DISCORD: Connected!")
			return
		}
	}
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (d *discord) disconnected(s *discordgo.Session, event *discordgo.Disconnect) {
	fmt.Println("🛑 DISCORD: Disconnected!")
	d.connect()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func (d *discord) ready(s *discordgo.Session, event *discordgo.Ready) {
	d.status <- true

	uguilds, err := s.UserGuilds(100, "", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, g := range uguilds {
		fmt.Printf("%s: %s\n", g.ID, g.Name)
		channels, err := s.GuildChannels(g.ID)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, c := range channels {
			fmt.Printf("%v %s: %s\n", c.Type, c.ID, c.Name)
		}
	}
	// Set the playing status.
	s.UpdateStatus(0, "With JonnyNof's Tiny Penis!")
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
			fmt.Printf("Message %s from %s\n", m.Content, m.ChannelID)
			s.ChannelMessageSend(m.ChannelID, "I don't do any interactions, yet.")
			for _, embed := range m.Embeds {
				fmt.Println(embed.Type)
			}
			return
		}
	}
	// }
}
