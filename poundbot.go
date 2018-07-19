package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"mrpoundsign.com/poundbot/discord"
	"mrpoundsign.com/poundbot/rust"
	"mrpoundsign.com/poundbot/twitter"
)

// var rustServer = net.TCPAddr{IP: "rust.alittlemercy.com", Port: 28015}

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

	rs, err := rust.NewServerInfo(rust.Server{Hostname: "rust.alittlemercy.com", Port: 28015})
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = rs.Update()
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Printf("%v", rs)
	// os.Exit(3)

	tConsumerKey := viper.GetString("twitter.consumer.key")
	tConsumerSecret := viper.GetString("twitter.consumer.secret")
	tAccessToken := viper.GetString("twitter.access.token")
	tAccessSecret := viper.GetString("twitter.access.secret")
	discordToken := viper.GetString("discord.token")
	linkChan := viper.GetString("discord.link-channel")
	statusChan := viper.GetString("discord.status-channel")

	log.Printf("🤖 Starting discord, linkChan %s, statusChan %s", linkChan, statusChan)
	dr := discord.DiscordRunner(discordToken, linkChan, statusChan)
	err = dr.Start()
	if err != nil {
		log.Println("🤖⚠️ Could not start Discord")
	}
	defer func() {
		log.Println("🤖 Shutting down Discord...")
		dr.Close()
	}()

	tcreds := twitter.Config{
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

	go func(statusChan *chan string) {
		var lastCheck = time.Now().UTC()
		// var lastUp = time.Now().UTC()

		var serverDown = true
		var downChecks uint
		var playerDelta int8

		var lowestPlayers uint8

		for {
			err := rs.Update()
			if err != nil {
				playerDelta = 0
				serverDown = true
				downChecks++
				if downChecks == 3 {
					fmt.Println("Server is down!")
				}
				continue
			}

			if serverDown {
				// lastUp = time.Now().UTC()
				lastCheck = time.Now().UTC()
				playerDelta = 0
				lowestPlayers = rs.PlayerInfo.Players
			}
			serverDown = false
			playerDelta += rs.PlayerInfo.PlayersDelta
			if playerDelta < 0 && rs.PlayerInfo.Players < lowestPlayers {
				playerDelta = 0
				lowestPlayers = rs.PlayerInfo.Players
			}
			// lastUp = time.Now().UTC()
			var now = time.Now().UTC()
			var duration = int(now.Sub(lastCheck).Minutes())
			if playerDelta > 3 || duration >= 10 {
				lastCheck = time.Now().UTC()
				if playerDelta > 0 {
					lowestPlayers = rs.PlayerInfo.Players
					var playerString = "player has"
					if playerDelta > 1 {
						playerString = "players have"
					}
					message := fmt.Sprintf("@here %d new %s connected, %d of %d playing now!", playerDelta, playerString, rs.PlayerInfo.Players, rs.PlayerInfo.MaxPlayers)
					log.Printf("🤖 Sending notice of %d new players\n", playerDelta)
					*statusChan <- message
					playerDelta = 0
				}
			}
			// log.Printf("%d, %d", duration, playerDelta)
			time.Sleep(30 * time.Second)
		}
	}(&dr.StatusChan)

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
