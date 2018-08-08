package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"bitbucket.org/mrpoundsign/poundbot/db"

	"bitbucket.org/mrpoundsign/poundbot/db/jsonstore"
	"bitbucket.org/mrpoundsign/poundbot/db/mongodb"
	"bitbucket.org/mrpoundsign/poundbot/discord"
	"bitbucket.org/mrpoundsign/poundbot/rust"
	"bitbucket.org/mrpoundsign/poundbot/rustconn"
	"bitbucket.org/mrpoundsign/poundbot/twitter"
	"github.com/spf13/viper"
)

var (
	version          = "DEVEL"
	buildstamp       = "NOWISH, I GUESS"
	githash          = "GIT HASHY WITH IT"
	versionFlag      = flag.Bool("v", false, "Displays the version and then quits")
	configLocation   = flag.String("c", ".", "The config.json location")
	writeConfig      = flag.Bool("w", false, "Writes a config and exits")
	writeConfigForce = flag.Bool("init", false, "Forces writing of config and exits\nWARNING! This will destroy your config file")
	wg               sync.WaitGroup
	killChan         = make(chan struct{})
)

type service interface {
	Start() error
	Stop()
}

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

func start(s service, name string) error {
	if err := s.Start(); err != nil {
		log.Printf("🤖 ⚠️ Failed to start %s: %s\n", name, err)
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-killChan
		log.Printf("🤖 Requesting %s shutdown...\n", name)
		s.Stop()
	}()

	return nil
}

func main() {
	flag.Parse()
	// If the version flag is set, print the version and quit.
	if *versionFlag {
		fmt.Printf("PoundBot %s (%s @ %s)\n", version, buildstamp, githash)
		return
	}

	servicesCount := 2 // ALways at least 1 for discord, but should always be >1

	// var thingsToKill = 0
	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.SetConfigFile(fmt.Sprintf("%s/config.json", filepath.Clean(*configLocation)))
	viper.SetDefault("datastore", "json")
	viper.SetDefault("json-store.path", "./json-store")
	viper.SetDefault("mongo.dial-addr", "mongodb://localhost")
	viper.SetDefault("mongo.database", "poundbot")
	viper.SetDefault("features.twitter", false)
	viper.SetDefault("features.players-joined", false)
	viper.SetDefault("features.raid-alerts", true)
	viper.SetDefault("features.chat-relay", true)
	viper.SetDefault("players-joined-frequency", 30)
	viper.SetDefault("http.bind_addr", "")
	viper.SetDefault("http.port", 9090)
	viper.SetDefault("discord.token", "YOUR DISCORD BOT AUTH TOKEN")
	viper.SetDefault("discord.channels.link", "CHANNEL ID FOR TWITTER LINKS")
	viper.SetDefault("discord.channels.status", "CHANNEL ID FOR SERVER STATUS (players joined)")
	viper.SetDefault("discord.channels.general", "CHANNEL ID FOR CHAT RELAY")
	viper.SetDefault("twitter.consumer.key", "CONSUMER KEY")
	viper.SetDefault("twitter.consumer.secret", "SECRET KEY")
	viper.SetDefault("twitter.access.token", "ACCESS TOKEN")
	viper.SetDefault("twitter.access.secret", "ACCESS SECRET")
	viper.SetDefault("twitter.userid", int64(0))
	viper.SetDefault("twitter.filters", "#ServerUpdate")

	var loaded = false

	if *writeConfigForce {
		*writeConfig = true
	} else {
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {
			log.Println(err)
			flag.Usage()
			os.Exit(1)
		}
		loaded = true
	}

	if loaded {
		if viper.IsSet("rust.api-server") {
			log.Println("Deprecated config option: /rust.api-server. Please remove.")
			log.Println("  copying to /http")
			viper.RegisterAlias("http", "rust.api-server")
		}

		if viper.IsSet("player-delta-frequency") {
			log.Println("Deprecated config option: /player-delta-frequency. Please remove.")
			log.Println("  copying to /players-joined-frequency")
			viper.RegisterAlias("players-joined-frequency", "player-delta-frequency")
		}
	}

	if *writeConfig {
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf("Could not write config: %s", err)
		}
		log.Printf("Wrote new config file to %s\n", viper.ConfigFileUsed())
		os.Exit(0)
	}

	dConfig := newDiscordConfig(viper.Sub("discord"))
	asConfig := newServerConfig(viper.Sub("http"))

	var datastore db.DataStore

	switch viper.GetString("datastore") {
	case "mongodb":
		mongo, err := mongodb.NewMongoDB(mongodb.Config{
			DialAddress: viper.GetString("mongo.dial-addr"),
			Database:    viper.GetString("mongo.database"),
		})
		if err != nil {
			log.Panicf("Could not connect to DB: %v\n", err)
		}
		datastore = mongo
	case "json":
		path := filepath.Clean(viper.GetString("json-store.path"))
		os.MkdirAll(path, os.ModePerm)
		datastore = jsonstore.NewJson(path)
	}

	datastore.Init()

	asConfig.Datastore = datastore

	// Discoed server
	dr := discord.Runner(dConfig)
	if err := start(dr, "Discord"); err != nil {
		log.Fatalf("Could not start Discord, %v\n", err)
	}

	// HTTP API server
	server := rustconn.NewServer(asConfig,
		rustconn.ServerChannels{
			RaidNotify:  dr.RaidAlertChan,
			DiscordAuth: dr.DiscordAuth,
			AuthSuccess: dr.AuthSuccess,
			ChatChan:    dr.GeneralChan,
			ChatOutChan: dr.GeneralOutChan,
		},
		rustconn.ServerOptions{
			RaidAlerts: viper.GetBool("features.raid-alerts"),
			ChatRelay:  viper.GetBool("features.chat-relay"),
		},
	)

	if err := start(server, "HTTP Server"); err != nil {
		log.Fatalf("Could not start HTTP server, %v\n", err)
	}

	if viper.GetBool("features.twitter") {
		servicesCount++
		tConfig := newTwitterConfig(viper.Sub("twitter"))
		t := twitter.NewTwitter(tConfig, dr.LinkChan)
		start(t, "Twitter")
	}

	// Rust Server Watcher
	if viper.GetBool("features.players-joined") {
		servicesCount++
		rConfig := newRustServerConfig(viper.Sub("rust.server"))
		pDeltaFreq := viper.GetInt("players-joined-frequency")
		rw, err := rust.NewWatcher(*rConfig, pDeltaFreq, dr.StatusChan)
		if err != nil {
			log.Fatalf("Can't start rust watcher, %v\n", err)
		}
		start(rw, "RustWatcher")
	}

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
	for i := 0; i < servicesCount; i++ {
		go func() { killChan <- struct{}{} }()
	}

	wg.Wait()
}
