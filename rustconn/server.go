package rustconn

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/poundbot/poundbot/storage"
	"github.com/poundbot/poundbot/types"
)

const upgradeURL = "https://umod.org/plugins/pound-bot"

var logPrefix = "[RC]"

// ServerConfig contains the base Server configuration
type ServerConfig struct {
	BindAddr string
	Port     int
	Storage  storage.Storage
}

type ServerChannels struct {
	RaidNotify  chan types.RaidAlert
	DiscordAuth chan types.DiscordAuth
	AuthSuccess chan types.DiscordAuth
	ChatChan    chan types.ChatMessage
	ChatQueue   storage.ChatQueueStore
}

// A Server runs the HTTP server, notification channels, and DB writing.
type Server struct {
	http.Server
	sc              *ServerConfig
	channels        ServerChannels
	shutdownRequest chan struct{}
}

// NewServer creates a Server
func NewServer(sc *ServerConfig, channels ServerChannels) *Server {
	s := Server{
		Server: http.Server{
			Addr: fmt.Sprintf("%s:%d", sc.BindAddr, sc.Port),
		},
		sc:       sc,
		channels: channels,
	}

	requestUUID := RequestUUID{}
	serverAuth := ServerAuth{as: sc.Storage.Accounts()}
	r := mux.NewRouter()

	// Handles all /api requests, and sets the server auth handler
	api := r.PathPrefix("/api").Subrouter()
	api.Use(serverAuth.Handle)
	api.Use(requestUUID.Handle)
	api.HandleFunc(
		"/entity_death",
		newEntityDeath(logPrefix, sc.Storage.RaidAlerts()),
	)
	api.HandleFunc(
		"/discord_auth",
		newDiscordAuth(logPrefix, sc.Storage.DiscordAuths(), sc.Storage.Users(), channels.DiscordAuth),
	)
	api.HandleFunc(
		"/chat",
		newChat(logPrefix, channels.ChatQueue, channels.ChatChan),
	).Methods(http.MethodGet, http.MethodPost)
	api.HandleFunc(
		"/clans",
		newClans(logPrefix, sc.Storage.Accounts()),
	).Methods(http.MethodPut)
	api.HandleFunc(
		"/clans/{tag}",
		newClan(logPrefix, sc.Storage.Accounts(), sc.Storage.Users()),
	).Methods(http.MethodDelete, http.MethodPut)
	api.HandleFunc("/players/registered", newRegisteredPlayers()).Methods(http.MethodGet)

	s.Handler = r

	s.shutdownRequest = make(chan struct{})

	return &s
}

// Start starts the HTTP server, raid alerter, and Discord auth manager
func (s *Server) Start() error {
	// Start the AuthSaver
	go func() {
		var newConn = s.sc.Storage.Copy()
		defer newConn.Close()

		var as = newAuthSaver(newConn.DiscordAuths(), newConn.Users(), s.channels.AuthSuccess, s.shutdownRequest)
		as.Run()
	}()

	// Start the RaidAlerter
	go func() {
		var newConn = s.sc.Storage.Copy()
		defer newConn.Close()

		var ra = newRaidAlerter(newConn.RaidAlerts(), s.channels.RaidNotify, s.shutdownRequest)
		ra.Run()
	}()

	go func() {
		log.Printf(logPrefix+" Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
		if err := s.ListenAndServe(); err != nil {
			log.Printf(logPrefix+" HTTP server died with error %v\n", err)
		} else {
			log.Printf(logPrefix+" HTTP server graceful shutdown\n", err)
		}
	}()

	return nil
}

// Stop stops the http server
func (s *Server) Stop() {
	log.Printf(logPrefix + "[WARN] Shutting down HTTP server ...")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() { //Create shutdown context with 10 second timeout
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//shutdown the server
		err := s.Shutdown(ctx)
		if err != nil {
			log.Printf(logPrefix+"[WARN] Shutdown request error: %v", err)
		}
	}()
	s.shutdownRequest <- struct{}{} // AuthSaver
	s.shutdownRequest <- struct{}{} // RaidAlerter
	wg.Wait()
}
