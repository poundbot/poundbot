package main

import (
	"flag"
	"fmt"
	"strings"

	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"github.com/poundbot/poundbot/discord"
	"github.com/poundbot/poundbot/gameapi"
	pblog "github.com/poundbot/poundbot/log"
	"github.com/poundbot/poundbot/messages"
	"github.com/poundbot/poundbot/storage/mongodb"
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
	log              = pblog.Log
)

type service interface {
	Start() error
	Stop()
}

func newServerConfig(cfg *viper.Viper, storage *mongodb.MongoDB) *gameapi.ServerConfig {
	return &gameapi.ServerConfig{
		BindAddr: cfg.GetString("http.bind_address"),
		Port:     cfg.GetInt("http.port"),
		Storage:  storage,
	}
}

func start(s service, name string) error {
	if err := s.Start(); err != nil {
		log.Warnf("Failed to start %s: %s", name, err)
		return fmt.Errorf("failed to start service %s: %w", name, err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-killChan
		log.Printf("Requesting %s shutdown...", name)
		s.Stop()
	}()

	return nil
}

func versionString() string {
	return fmt.Sprintf("PoundBot %s (%s @ %s)", version, buildstamp, githash)
}

func main() {
	messages.Init()
	flag.Parse()
	// If the version flag is set, print the version and quit.

	log.Println(versionString())
	if *versionFlag {
		return
	}

	servicesCount := 2 // Always at least 1 for discord, but should always be >1

	runtime.GOMAXPROCS(runtime.NumCPU())

	viper.SetConfigFile(fmt.Sprintf("%s/config.json", filepath.Clean(*configLocation)))
	viper.SetDefault("mongo.dial", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "poundbot")
	viper.SetDefault("mongo.ssl.enabled", false)
	viper.SetDefault("mongo.ssl.insecure", false)
	viper.SetDefault("http.bind_addr", "")
	viper.SetDefault("http.port", 9090)
	viper.SetDefault("discord.token", "YOUR DISCORD BOT AUTH TOKEN")
	viper.SetDefault("profiler.port", 6061)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if *writeConfigForce {
		*writeConfig = true
	} else {
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {
			log.Println(err)
			flag.Usage()
			os.Exit(1)
		}
	}

	if *writeConfig {
		err := viper.WriteConfig()
		if err != nil {
			log.Fatalf("Could not write config: %s", err)
		}
		log.Printf("Wrote new config file to %s", viper.ConfigFileUsed())
		os.Exit(0)
	}

	go func() {
		log.Fatal(http.ListenAndServe("localhost:"+viper.GetString("profiler.port"), nil))
	}()

	dialAddr := viper.GetString("mongo.dial-addr")
	if len(dialAddr) != 0 {
		log.Warn("DEPRICATION WARNING: mongo.dial-addr has been renamed to mongo.dial.")
	} else {
		dialAddr = viper.GetString("mongo.dial")
	}

	store, err := mongodb.NewMongoDB(mongodb.Config{
		DialAddress: dialAddr,
		Database:    viper.GetString("mongo.database"),
		SSL:         viper.GetBool("mongo.ssl.enabled"),
		InsecureSSL: viper.GetBool("mongo.ssl.insecure"),
	})

	if err != nil {
		log.Panicf("Could not connect to DB: %v", err)
	}

	store.Init()

	webConfig := newServerConfig(viper.GetViper(), store)

	// Discord server
	dr := discord.NewRunner(viper.GetString("discord.token"), store.Accounts(), store.DiscordAuths(),
		store.Users(), store.MessageLocks(), store.ChatQueue())
	if err := start(dr, "Discord"); err != nil {
		log.Fatalf("Could not start Discord, %v", err)
		os.Exit(1)
	}

	// HTTP API server
	server := gameapi.NewServer(
		webConfig,
		dr,
		gameapi.ServerChannels{
			AuthSuccess: dr.AuthSuccess,
			ChatQueue:   store.ChatQueue(),
		},
	)

	if err := start(server, "HTTP Server"); err != nil {
		log.Fatalf("Could not start HTTP server, %v\n", err)
		os.Exit(1)
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

	log.Warn("Stopping...")
	for i := 0; i < servicesCount; i++ {
		go func() { killChan <- struct{}{} }()
	}

	wg.Wait()
}
