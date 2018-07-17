package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/spf13/viper"
	"mrpoundsign.com/almbot/discord"
)

// var discordStatus = make(chan bool)
// var incTweets = make(chan *twitter.Tweet)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.AddConfigPath("/etc/poundbot/")
	viper.AddConfigPath("$HOME/.poundbot")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	twitterConsumerKey := viper.GetString("twitter.consumer.key")
	twitterConsumerSecret := viper.GetString("twitter.consumer.secret")
	twitterAccessToken := viper.GetString("twitter.access.token")
	twitterAccessSecret := viper.GetString("twitter.access.secret")
	discordToken := viper.GetString("discord.token")
	discordChannel := viper.GetString("discord.twitter-channel")

	fmt.Println("Discord channel is ", discordChannel)
	dr := discord.DiscordRunner(discordToken, discordChannel)
	err = dr.Start()
	if err != nil {
		panic(err)
	}
	defer dr.Close()

	// fmt.Printf(discordAccount + " " + discordToken)
	// os.Exit(3)

	config := oauth1.NewConfig(twitterConsumerKey, twitterConsumerSecret)
	token := oauth1.NewToken(twitterAccessToken, twitterAccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if tweet.User.ID == 1016357953807400960 {
			dr.TweetChan <- tweet
		}
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

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Stopping Stream...")

	if err != nil {
		panic(err)
	}
}
