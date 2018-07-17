package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/spf13/viper"
)

var discordStatus = make(chan bool)
var incTweets = make(chan *twitter.Tweet)
var discordChannel = ""

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.AddConfigPath("/etc/almbot/")
	viper.AddConfigPath("$HOME/.almbot")
	viper.AddConfigPath(".")
	viper.SetConfigFile("config.json")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	twitterConsumerKey := viper.GetString("twitter.consumer.key")
	twitterConsumerSecret := viper.GetString("twitter.consumer.secret")
	twitterAccessToken := viper.GetString("twitter.access.token")
	twitterAccessSecret := viper.GetString("twitter.access.secret")
	// discordAccount := viper.GetString("discord.account")
	discordToken := viper.GetString("discord.token")
	discordChannel = viper.GetString("discord.twitter-channel")

	fmt.Println("Discord channel is ", discordChannel)

	// fmt.Printf(discordAccount + " " + discordToken)
	// os.Exit(3)

	config := oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	token := oauth1.NewToken(twitterAccessToken, twitterAccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	discord, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		panic("NOOO")
	}

	discord.AddHandler(messageCreate)
	discord.AddHandler(ready)
	go discordConnnect(discord)
	defer discord.Close()

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		incTweets <- tweet
		fmt.Println(tweet.Text)
	}
	demux.Event = func(event *twitter.Event) {
		fmt.Printf("%#v\n", event)
	}

	fmt.Println("Starting Stream...")

	// users, _, err := client.Users.Lookup(&twitter.UserLookupParams{ScreenName: []string{"little_rust"}})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for _, u := range users {
	// 	fmt.Println(u.ID, u.ScreenName)
	// }
	// os.Exit(3)

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		Follow:        []string{"1016357953807400960"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal(err)
	}

	defer stream.Stop()

	// Receive messages until stopped or stream quits
	go demux.HandleChan(stream.Messages)

	go func(discord *discordgo.Session) {
	Connected:
		for {
			status := <-discordStatus
			if status {
				for {
					select {
					case t := <-incTweets:
						if strings.Contains(strings.ToLower(t.Text), "#almupdate") {
							_, err := discord.ChannelMessageSend(
								discordChannel,
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
							fmt.Println("Not posting tweet: ", t.Text)
						}

					case stop := <-discordStatus:
						if stop == true {
							break Connected
						}
					}

				}
			}
			time.Sleep(1 * time.Second)
		}
	}(discord)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Stopping Stream...")

	if err != nil {
		panic(err)
	}
}

func discordConnnect(s *discordgo.Session) {
	fmt.Println("⚪ DISCORD: Connecting")
	for {
		err := s.Open()
		discordStatus <- false
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
func disconnected(s *discordgo.Session, event *discordgo.Disconnect) {
	discordStatus <- false
	fmt.Println("🛑 DISCORD: Disconnected!")
	discordConnnect(s)
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	discordStatus <- true

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
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

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
