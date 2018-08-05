package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/db/mongodb"
	"bitbucket.org/mrpoundsign/poundbot/discord"
	"bitbucket.org/mrpoundsign/poundbot/rust"
	"bitbucket.org/mrpoundsign/poundbot/rustconn"
	"bitbucket.org/mrpoundsign/poundbot/twitter"
	"github.com/spf13/viper"
)

func newDiscordConfig(cfg *viper.Viper) *discord.RunnerConfig {
	return &discord.RunnerConfig{
		Token:       cfg.GetString("token"),
		LinkChan:    cfg.GetString("channels.link"),
		StatusChan:  cfg.GetString("channels.status"),
		GeneralChan: cfg.GetString("channels.general"),
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

func newServerConfig(cfg *viper.Viper) *rustconn.ServerConfig {
	return &rustconn.ServerConfig{
		BindAddr: cfg.GetString("bind_address"),
		Port:     cfg.GetInt("port"),
	}
}

func newRustServerConfig(cfg *viper.Viper) *rust.ServerConfig {
	return &rust.ServerConfig{Hostname: cfg.GetString("hostname"), Port: cfg.GetInt("port")}
}

func main() {
	var wg sync.WaitGroup
	killChan := make(chan struct{})
	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetDefault("player-delta-frequency", 30)
	viper.SetDefault("rust.api-server.bind_addr", "")
	viper.SetDefault("rust.api-server.port", 9090)
	viper.SetDefault("mongo.dial-addr", "mongodb://localhost")
	viper.SetDefault("mongo.database", "poundbot")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Panicf("fatal error config file: %s\n", err)
	}

	dConfig := newDiscordConfig(viper.Sub("discord"))
	tConfig := newTwitterConfig(viper.Sub("twitter"))
	rConfig := newRustServerConfig(viper.Sub("rust.server"))
	pDeltaFreq := viper.GetInt("player-delta-frequency")
	asConfig := newServerConfig(viper.Sub("rust.api-server"))

	datastore, err := mongodb.NewMgo(mongodb.MongoConfig{
		DialAddress: viper.GetString("mongo.dial-addr"),
		Database:    viper.GetString("mongo.database"),
	})
	if err != nil {
		log.Panicf("Could not connect to DB: %v\n", err)
	}
	datastore.CreateIndexes()

	asConfig.Datastore = *datastore

	rs, err := rust.NewServer(rConfig)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = rs.Update()
	if err != nil {
		log.Println("🤖 ⚠️ Error contacting Rust server: " + err.Error())
	}

	log.Printf("🤖 Starting discord, linkChan %s, statusChan %s", dConfig.LinkChan, dConfig.StatusChan)
	dr := discord.Runner(dConfig)
	wg.Add(1)
	err = dr.Start()
	if err != nil {
		log.Println("🤖 ⚠️ Could not start Discord")
	}
	go func() {
		<-killChan
		log.Println("🤖 Shutting down Discord...")
		dr.Close()
		wg.Done()
	}()

	server := rustconn.NewServer(asConfig, dr.RaidAlertChan, dr.DiscordAuth, dr.AuthSuccess, dr.GeneralChan, dr.GeneralOutChan)
	server.Serve()
	wg.Add(1)
	go func() {
		<-killChan
		log.Println("🤖 Shutting down HTTP Server...")
		server.Stop()
		wg.Done()
	}()

	t := twitter.NewTwitter(tConfig, dr.LinkChan)
	t.Start()
	wg.Add(1)
	go func() {
		<-killChan
		log.Println("🤖 Shutting down Twitter...")
		t.Stop()
		wg.Done()
	}()

	go func(statusChan chan string) {
		defer wg.Done()
		var lastCheck = time.Now().UTC()

		var serverDown = true
		var downChecks uint
		var playerDelta int8

		var lowestPlayers uint8

		var waitOrKill = func(t time.Duration) (kill bool) {
			select {
			case <-killChan:
				log.Println("🤖 Shutting down Rust Monitor")
				kill = true
			case <-time.After(t):
				kill = false
			}
			return
		}

		for {
			err := rs.Update()
			if err != nil {
				playerDelta = 0
				serverDown = true
				downChecks++
				if downChecks%3 == 0 {
					log.Println("🤖 🏃 ⚠️ Server is down!")
					if waitOrKill(20 * time.Second) {
						return
					}
				}
				if waitOrKill(5 * time.Second) {
					return
				}
			} else {
				if downChecks > 0 {
					downChecks = 0
					log.Println("🤖 🏃 Server is back!")
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
						log.Printf("🤖 🏃 Sending notice of %d new players\n", playerDelta)
						statusChan <- message
						playerDelta = 0
					}
				}
			}

			if waitOrKill(30 * time.Second) {
				return
			}
		}
	}(dr.StatusChan)

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
	killChan <- struct{}{} // HTTP Server
	killChan <- struct{}{} // Twitter
	killChan <- struct{}{} // Discord
	killChan <- struct{}{} // Rust server monitor

	wg.Wait()

	if err != nil {
		panic(err)
	}
}
