package main

import (
	"flag"
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"bitbucket.org/mrpoundsign/poundbot/discord"
	"bitbucket.org/mrpoundsign/poundbot/rust"
	"bitbucket.org/mrpoundsign/poundbot/rustconn"
	"bitbucket.org/mrpoundsign/poundbot/storage"
	"bitbucket.org/mrpoundsign/poundbot/storage/jsonstore"
	"bitbucket.org/mrpoundsign/poundbot/storage/mongodb"
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
		Token: cfg.GetString("token"),
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
		log.Printf("[MAIN][WARN] Failed to start %s: %s\n", name, err)
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-killChan
		log.Printf("[MAIN] Requesting %s shutdown...\n", name)
		s.Stop()
	}()

	return nil
}

func versionString() string {
	return fmt.Sprintf("PoundBot %s (%s @ %s)\n", version, buildstamp, githash)
}

func main() {
	flag.Parse()
	// If the version flag is set, print the version and quit.

	fmt.Println(versionString())
	if *versionFlag {
		return
	}

	servicesCount := 2 // ALways at least 1 for discord, but should always be >1

	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.SetConfigFile(fmt.Sprintf("%s/config.json", filepath.Clean(*configLocation)))
	viper.SetDefault("storage", "mongodb")
	viper.SetDefault("json-store.path", "./json-store")
	viper.SetDefault("mongo.dial-addr", "mongodb://localhost")
	viper.SetDefault("mongo.database", "poundbot")
	viper.SetDefault("features.players-joined", false)
	viper.SetDefault("features.raid-alerts", true)
	viper.SetDefault("features.chat-relay", true)
	viper.SetDefault("players-joined-frequency", 30)
	viper.SetDefault("http.bind_addr", "")
	viper.SetDefault("http.port", 9090)
	viper.SetDefault("discord.token", "YOUR DISCORD BOT AUTH TOKEN")
	viper.SetDefault("discord.channels.status", "CHANNEL ID FOR SERVER STATUS (players joined)")
	viper.SetDefault("discord.channels.general", "CHANNEL ID FOR CHAT RELAY")

	go func() {
		log.Fatal(http.ListenAndServe("localhost:6061", nil))
	}()

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

	var store storage.Storage

	switch viper.GetString("storage") {
	case "mongodb":
		mongo, err := mongodb.NewMongoDB(mongodb.Config{
			DialAddress: viper.GetString("mongo.dial-addr"),
			Database:    viper.GetString("mongo.database"),
		})
		if err != nil {
			log.Panicf("Could not connect to DB: %v\n", err)
		}
		store = mongo
	case "json":
		path := filepath.Clean(viper.GetString("json-store.path"))
		os.MkdirAll(path, os.ModePerm)
		store = jsonstore.NewJson(path)
	}

	store.Init()

	asConfig.Storage = store

	// Discord server
	dr := discord.Runner(dConfig.Token, store.Accounts(), store.Chats(), store.DiscordAuths(), store.Users())
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
		},
		rustconn.ServerOptions{
			RaidAlerts: viper.GetBool("features.raid-alerts"),
			ChatRelay:  !viper.GetBool("features.chat-relay"),
		},
	)

	if err := start(server, "HTTP Server"); err != nil {
		log.Fatalf("Could not start HTTP server, %v\n", err)
	}

	// Rust Server Watcher
	// if viper.GetBool("features.players-joined") {
	// 	servicesCount++
	// 	rConfig := newRustServerConfig(viper.Sub("rust.server"))
	// 	pDeltaFreq := viper.GetInt("players-joined-frequency")
	// 	rw, err := rust.NewWatcher(*rConfig, pDeltaFreq, dr.StatusChan)
	// 	if err != nil {
	// 		log.Fatalf("Can't start rust watcher, %v\n", err)
	// 	}
	// 	start(rw, "RustWatcher")
	// }

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

	log.Println("[MAIN] Stopping...")
	for i := 0; i < servicesCount; i++ {
		go func() { killChan <- struct{}{} }()
	}

	wg.Wait()
}
