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

	"bitbucket.org/mrpoundsign/poundbot/chatcache"
	"bitbucket.org/mrpoundsign/poundbot/discord"
	"bitbucket.org/mrpoundsign/poundbot/messages"
	"bitbucket.org/mrpoundsign/poundbot/rustconn"
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

func newServerConfig(cfg *viper.Viper, storage *mongodb.MongoDb) *rustconn.ServerConfig {
	return &rustconn.ServerConfig{
		BindAddr: cfg.GetString("bind_address"),
		Port:     cfg.GetInt("port"),
		Storage:  storage,
	}
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
	messages.Init()
	flag.Parse()
	// If the version flag is set, print the version and quit.

	fmt.Println(versionString())
	if *versionFlag {
		return
	}

	servicesCount := 2 // ALways at least 1 for discord, but should always be >1

	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.SetConfigFile(fmt.Sprintf("%s/config.json", filepath.Clean(*configLocation)))
	viper.SetDefault("mongo.dial-addr", "mongodb://localhost")
	viper.SetDefault("mongo.database", "poundbot")
	viper.SetDefault("http.bind_addr", "")
	viper.SetDefault("http.port", 9090)
	viper.SetDefault("discord.token", "YOUR DISCORD BOT AUTH TOKEN")

	go func() {
		log.Fatal(http.ListenAndServe("localhost:6061", nil))
	}()

	// var loaded = false

	if *writeConfigForce {
		*writeConfig = true
	} else {
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {
			log.Println(err)
			flag.Usage()
			os.Exit(1)
		}
		// loaded = true
	}

	if *writeConfig {
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf("Could not write config: %s", err)
		}
		log.Printf("Wrote new config file to %s\n", viper.ConfigFileUsed())
		os.Exit(0)
	}

	store, err := mongodb.NewMongoDB(mongodb.Config{
		DialAddress: viper.GetString("mongo.dial-addr"),
		Database:    viper.GetString("mongo.database"),
	})

	if err != nil {
		log.Panicf("Could not connect to DB: %v\n", err)
	}

	store.Init()

	dConfig := newDiscordConfig(viper.Sub("discord"))
	webConfig := newServerConfig(viper.Sub("http"), store)

	ccache := chatcache.NewChatCache()

	// Discord server
	dr := discord.Runner(dConfig.Token, ccache, store.Accounts(), store.DiscordAuths(), store.Users())
	if err := start(dr, "Discord"); err != nil {
		log.Fatalf("Could not start Discord, %v\n", err)
	}

	// HTTP API server
	server := rustconn.NewServer(
		webConfig,
		rustconn.ServerChannels{
			RaidNotify:  dr.RaidAlertChan,
			DiscordAuth: dr.DiscordAuth,
			AuthSuccess: dr.AuthSuccess,
			ChatChan:    dr.ChatChan,
			ChatCache:   *ccache,
		},
	)

	if err := start(server, "HTTP Server"); err != nil {
		log.Fatalf("Could not start HTTP server, %v\n", err)
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

	log.Println("[MAIN] Stopping...")
	for i := 0; i < servicesCount; i++ {
		go func() { killChan <- struct{}{} }()
	}

	wg.Wait()
}
