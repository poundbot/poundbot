package rustconn

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/gorilla/mux"
)

var logSymbol = "🕸️ "

// ServerConfig contains the base Server configuration
type ServerConfig struct {
	BindAddr  string
	Port      int
	Datastore db.DataStore
}

type ServerChannels struct {
	RaidNotify  chan types.RaidNotification
	DiscordAuth chan types.DiscordAuth
	AuthSuccess chan types.DiscordAuth
	ChatChan    chan string
	ChatOutChan chan types.ChatMessage
}

type ServerOptions struct {
	RaidAlerts bool
	ChatRelay  bool
}

// A Server runs the HTTP server, notification channels, and DB writing.
type Server struct {
	http.Server
	sc              *ServerConfig
	channels        ServerChannels
	options         ServerOptions
	shutdownRequest chan struct{}
}

// NewServer creates a Server
func NewServer(sc *ServerConfig, channels ServerChannels, options ServerOptions) *Server {
	s := Server{
		Server: http.Server{
			Addr: fmt.Sprintf("%s:%d", sc.BindAddr, sc.Port),
		},
		sc:       sc,
		channels: channels,
		options:  options,
	}

	r := mux.NewRouter()
	r.HandleFunc("/entity_death", s.entityDeathHandler)
	r.HandleFunc("/discord_auth", s.discordAuthHandler)
	r.HandleFunc("/chat", s.chatHandler)
	r.HandleFunc("/clans", s.clansHandler).Methods(http.MethodPut)
	r.HandleFunc("/clans/{tag}", s.clanHandler).Methods(http.MethodDelete, http.MethodPut)
	s.Handler = r

	s.shutdownRequest = make(chan struct{})

	return &s
}

// Serve starts the HTTP server, raid alerter, and Discord auth manager
func (s *Server) Serve() {
	// Start the AuthSaver
	go func() {
		var newConn = s.sc.Datastore.Copy()
		defer newConn.Close()

		var as = NewAuthSaver(newConn.DiscordAuths(), newConn.Users(), s.channels.AuthSuccess, s.shutdownRequest)
		as.Run()
	}()

	if s.options.RaidAlerts {
		// Start the RaidAlerter
		go func() {
			var newConn = s.sc.Datastore.Copy()
			defer newConn.Close()

			var ra = NewRaidAlerter(newConn.RaidAlerts(), s.channels.RaidNotify, s.shutdownRequest)
			ra.Run()
		}()
	}

	go func() {
		log.Printf(logSymbol+"Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
		if err := s.ListenAndServe(); err != nil {
			log.Printf(logSymbol+"HTTP server died with error %v\n", err)
		} else {
			log.Printf(logSymbol+"HTTP server graceful shutdown\n", err)
		}
	}()

}

func (s *Server) Stop() {
	log.Printf(logSymbol + "Stoping http server ...")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() { //Create shutdown context with 10 second timeout
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		//shutdown the server
		err := s.Shutdown(ctx)
		if err != nil {
			log.Printf(logSymbol+"Shutdown request error: %v", err)
		}
	}()
	s.shutdownRequest <- struct{}{} // AuthSaver
	if s.options.RaidAlerts {
		s.shutdownRequest <- struct{}{} // RaidAlerter
	}
	wg.Wait()
}

// clansHandler manages clans sync HTTP requests from the Rust server
// These requests are a complete refresh of all clans
func (s *Server) clansHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t []types.ServerClan
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(logSymbol + err.Error())
		return
	}

	clanCount := len(t)
	clans := make([]types.Clan, clanCount)
	tags := make([]string, clanCount)
	for i, sc := range t {
		c, err := types.ClanFromServerClan(sc)
		if err != nil {
			log.Printf(logSymbol+"clansHandler Error: %v\n", err)
			handleError(w, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}
		tags[i] = c.Tag
		clans[i] = *c
	}

	db := s.sc.Datastore.Copy()
	defer db.Close()

	for _, clan := range clans {
		db.Clans().Upsert(clan)
	}

	db.Clans().RemoveNotIn(tags)
	db.Users().RemoveClansNotIn(tags)
}

// clanHandler manages individual clan REST requests form the Rust server
func (s *Server) clanHandler(w http.ResponseWriter, r *http.Request) {
	db := s.sc.Datastore.Copy()
	defer db.Close()

	vars := mux.Vars(r)
	tag := vars["tag"]

	switch r.Method {
	case http.MethodDelete:
		log.Printf(logSymbol+"clanHandler: Removing clan %s\n", tag)
		db.Clans().Remove(tag)
		db.Users().RemoveClan(tag)
		return
	case http.MethodPut:
		log.Printf(logSymbol+"clanHandler: Updating clan %s\n", tag)
		decoder := json.NewDecoder(r.Body)
		var t types.ServerClan
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(logSymbol + err.Error())
			return
		}

		clan, err := types.ClanFromServerClan(t)
		if err != nil {
			handleError(w, types.RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}

		db.Clans().Upsert(*clan)
	}
}

// chatHandler manages Rust <-> discord chat requests and logging
// Discord -> Rust is through the ChatOutChan and Rust -> Discord is
// through ChatChan.
//
// HTTP POST requests are sent to ChatChan
//
// HTTP GET requests wait for messages and disconnect with http.StatusNoContent
// after 5 seconds.
func (s *Server) chatHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Make this less awful. Plugin must be updated to be smarter.
	if !s.options.ChatRelay {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			time.Sleep(10 * time.Second)
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	db := s.sc.Datastore.Copy()
	defer db.Close()

	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t types.ChatMessage
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(logSymbol + err.Error())
			return
		}

		t.Source = types.ChatSourceRust
		db.Chats().Log(t)
		go func(t types.ChatMessage, c chan string) {
			var clan = ""
			if t.ClanTag != "" {
				clan = fmt.Sprintf("[%s] ", t.ClanTag)
			}
			c <- fmt.Sprintf("☢️ **%s%s**: %s", clan, t.DisplayName, t.Message)
		}(t, s.channels.ChatChan)
	case http.MethodGet:
		select {
		case res := <-s.channels.ChatOutChan:
			b, err := json.Marshal(res)
			if err != nil {
				log.Println(logSymbol + err.Error())
				return
			}
			db.Chats().Log(res)

			w.Write(b)
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusNoContent)
		}

	default:
		handleError(w, types.RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("Method %s not allowed", r.Method),
		})
	}
}

// entityDeathHandler manages incoming Rust entity death notices and sends them
// to the RaidAlertsStore and RaidAlerts channel
func (s *Server) entityDeathHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ed types.EntityDeath
	err := decoder.Decode(&ed)
	if err != nil {
		log.Println(logSymbol + err.Error())
		return
	}

	db := s.sc.Datastore.Copy()
	defer db.Close()
	db.RaidAlerts().AddInfo(ed)
}

// discordAuthHandler takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsStore and DiscordAuth channel
func (s *Server) discordAuthHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t types.DiscordAuth
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(logSymbol + err.Error())
		return
	}

	log.Printf(logSymbol+"User Auth Request: %v from %v\n", t, r.Body)
	db := s.sc.Datastore.Copy()
	defer db.Close()

	u, err := db.Users().Get(t.SteamInfo)
	if err == nil {
		handleError(w, types.RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("%s is linked to you.", u.DiscordID),
		})
		return
	} else if t.DiscordID == "check" {
		handleError(w, types.RESTError{
			StatusCode: http.StatusNotFound,
			Error:      "Account is not linked to discord.",
		})
		return
	}

	err = db.DiscordAuths().Upsert(t)
	if err == nil {
		s.channels.DiscordAuth <- t
	} else {
		log.Println(logSymbol + err.Error())
	}
}

// handleError is a generic JSON HTTP error response
func handleError(w http.ResponseWriter, restError types.RESTError) {
	w.WriteHeader(restError.StatusCode)
	err := json.NewEncoder(w).Encode(restError)
	if err != nil {
		log.Printf(logSymbol+"Error encoding %v, %s\n", restError, err)
	}
}
