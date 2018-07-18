package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/spf13/viper"
	"mrpoundsign.com/poundbot/discord"
	"mrpoundsign.com/poundbot/twitter"
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
		log.Panicf("fatal error config file: %s\n", err)
	}

	tConsumerKey := viper.GetString("twitter.consumer.key")
	tConsumerSecret := viper.GetString("twitter.consumer.secret")
	tAccessToken := viper.GetString("twitter.access.token")
	tAccessSecret := viper.GetString("twitter.access.secret")
	discordToken := viper.GetString("discord.token")
	discordChannel := viper.GetString("discord.twitter-channel")

	log.Println("🤖 Starting discord on channel", discordChannel)
	dr := discord.DiscordRunner(discordToken, discordChannel)
	err = dr.Start()
	if err != nil {
		log.Println("🤖⚠️ Could not start Discord")
	}
	defer func() {
		log.Println("🤖 Shutting down Discord...")
		dr.Close()
	}()

	tcreds := twitter.TwitterConfig{
		ConsumerKey:    tConsumerKey,
		ConsumerSecret: tConsumerSecret,
		AccessToken:    tAccessToken,
		AccessSecret:   tAccessSecret,
		UserID:         1016357953807400960,
		Filters:        []string{"#almupdate"},
	}

	t := twitter.NewTwitter(tcreds, dr.LinkChan)
	t.Start()
	defer func() {
		log.Println("🤖 Shutting down Twitter...")
		t.Stop()
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(
		sc,
		syscall.SIGTERM, // "the normal way to politely ask a program to terminate"
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGQUIT, // Ctrl-\
		syscall.SIGKILL, // "always fatal", "SIGKILL and SIGSTOP may not be caught by a program"
		syscall.SIGHUP,  // "terminal is disconnected"
		os.Kill,
		os.Interrupt,
	)
	<-sc

	log.Println("🤖 Stopping...")

	if err != nil {
		panic(err)
	}
}
