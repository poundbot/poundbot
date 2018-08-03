package rustconn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"mrpoundsign.com/poundbot/db"
	"mrpoundsign.com/poundbot/types"
)

var logSymbol = "🕸️ "

// ServerConfig contains the base Server configuration
type ServerConfig struct {
	BindAddr string
	Port     int
	Database db.DataAccessLayer
}

// A Server runs the HTTP server, notification channels, and DB writing.
type Server struct {
	sc          *ServerConfig
	RaidNotify  chan types.RaidNotification
	DiscordAuth chan types.DiscordAuth
	AuthSuccess chan types.DiscordAuth
	ChatChan    chan string
	ChatOutChan chan types.ChatMessage
}

// NewServer creates a Server
func NewServer(sc *ServerConfig, rch chan types.RaidNotification, dach, asch chan types.DiscordAuth, cch chan string, coch chan types.ChatMessage) *Server {
	return &Server{
		sc:          sc,
		RaidNotify:  rch,
		DiscordAuth: dach,
		AuthSuccess: asch,
		ChatChan:    cch,
		ChatOutChan: coch,
	}
}

// Serve starts the HTTP server, raid alerter, and Discord auth manager
func (s *Server) Serve() {
	done := make(chan struct{})
	defer func() {
		// One done channel per runner
		done <- struct{}{}
		done <- struct{}{}
	}()

	go s.authHandler(done)
	go s.raidAlerter(done)

	log.Printf(logSymbol+"Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
	r := mux.NewRouter()
	r.HandleFunc("/entity_death", s.entityDeathHandler)
	r.HandleFunc("/discord_auth", s.discordAuthHandler)
	r.HandleFunc("/chat", s.chatHandler)
	r.HandleFunc("/clans", s.clansHandler).Methods(http.MethodPut)
	r.HandleFunc("/clans/{tag}", s.clanHandler).Methods(http.MethodDelete, http.MethodPut)
	http.Handle("/", r)
	go log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.sc.BindAddr, s.sc.Port), r))
}

// raidAlerter checks for raids that need to be alerted and sends them
// out through the RaidNotify channel
func (s *Server) raidAlerter(done chan struct{}) {
	db := s.sc.Database.Copy()
	defer db.Close()

	for {
		var results []types.RaidNotification
		db.RaidAlerts().GetReady(&results)

		for _, result := range results {
			s.RaidNotify <- result
			db.RaidAlerts().Remove(result)
		}

	ExitCheck:
		select {
		case <-done:
			return
		default:
			break ExitCheck
		}
		time.Sleep(1)
	}
}

// authHandler writes users sent in through the AuthSuccess channel
func (s *Server) authHandler(done chan struct{}) {
	db := s.sc.Database.Copy()
	defer db.Close()
ExitCheck:
	for {
		select {
		case as := <-s.AuthSuccess:

			err := db.Users().BaseUpsert(as.BaseUser)

			if err == nil {
				db.DiscordAuths().Remove(as.SteamInfo)
				if as.Ack != nil {
					as.Ack(true)
				}
			} else {
				if as.Ack != nil {
					as.Ack(false)
				}
			}
		case <-done:
			break ExitCheck
		}
	}
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

	db := s.sc.Database.Copy()
	defer db.Close()

	for _, clan := range clans {
		db.Clans().Upsert(clan)
	}

	db.Clans().RemoveNotIn(tags)
	db.Users().RemoveClansNotIn(tags)
}

// clanHandler manages individual clan REST requests form the Rust server
func (s *Server) clanHandler(w http.ResponseWriter, r *http.Request) {
	db := s.sc.Database.Copy()
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
	db := s.sc.Database.Copy()
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
		}(t, s.ChatChan)
	case http.MethodGet:
		select {
		case res := <-s.ChatOutChan:
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
// to the RaidAlertsAccessLayer and RaidAlerts channel
func (s *Server) entityDeathHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var ed types.EntityDeath
	err := decoder.Decode(&ed)
	if err != nil {
		log.Println(logSymbol + err.Error())
		return
	}

	db := s.sc.Database.Copy()
	defer db.Close()
	db.RaidAlerts().AddInfo(ed)
}

// discordAuthHandler takes Discord verification requests from the Rust server
// and sends them to the DiscordAuthsAccessLayer and DiscordAuth channel
func (s *Server) discordAuthHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t types.DiscordAuth
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(logSymbol + err.Error())
		return
	}

	log.Printf(logSymbol+"User Auth Request: %v from %v\n", t, r.Body)
	db := s.sc.Database.Copy()
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
		s.DiscordAuth <- t
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
