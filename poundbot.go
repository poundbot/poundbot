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

func newDiscordConfig(cfg *viper.Viper) *discord.RunnerConfig {
	return &discord.RunnerConfig{
		Token:      cfg.GetString("token"),
		LinkChan:   cfg.GetString("channels.link"),
		StatusChan: cfg.GetString("channels.status"),
	}
}

func newTwitterConfig(cfg *viper.Viper) *twitter.Config {
	return &twitter.Config{
		ConsumerKey:    cfg.GetString("consumer.key"),
		ConsumerSecret: cfg.GetString("consumer.secret"),
		AccessToken:    cfg.GetString("access.token"),
		AccessSecret:   cfg.GetString("access.secret"),
		UserID:         cfg.GetInt64("userid"),
		Filters:        cfg.GetStringSlice("filters"),
	}
}

func newRustServer(cfg *viper.Viper) *rust.Server {
	return &rust.Server{Hostname: cfg.GetString("hostname"), Port: cfg.GetInt("port")}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.AddConfigPath("/etc/poundbot/")
	viper.AddConfigPath("$HOME/.poundbot")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetDefault("player-delta-frequency", 30)
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Panicf("fatal error config file: %s\n", err)
	}

	dConfig := newDiscordConfig(viper.Sub("discord"))
	tConfig := newTwitterConfig(viper.Sub("twitter"))
	rConfig := newRustServer(viper.Sub("rust.server"))
	pDeltaFreq := viper.GetInt("player-delta-frequency")

	rs, err := rust.NewServerInfo(rConfig)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = rs.Update()
	if err != nil {
		log.Fatalf(err.Error())
	}
	// log.Printf("%v", rs)
	// os.Exit(3)

	log.Printf("🤖 Starting discord, linkChan %s, statusChan %s", dConfig.LinkChan, dConfig.StatusChan)
	dr := discord.DiscordRunner(dConfig)
	err = dr.Start()
	if err != nil {
		log.Println("🤖⚠️ Could not start Discord")
	}
	defer func() {
		log.Println("🤖 Shutting down Discord...")
		dr.Close()
	}()

	t := twitter.NewTwitter(tConfig, dr.LinkChan)
	t.Start()
	defer func() {
		log.Println("🤖 Shutting down Twitter...")
		t.Stop()
	}()

	go func(statusChan *chan string) {
		defer func() {
			log.Println("🤖 Shutting down Rust Monitor")
		}()
		var lastCheck = time.Now().UTC()

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
					fmt.Println("🤖 Server is down!")
				}
				time.Sleep(30 * time.Second)
				continue
			}

			if serverDown {
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
			if playerDelta > 3 || duration >= pDeltaFreq {
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
