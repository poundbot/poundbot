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

type ServerConfig struct {
	BindAddr string
	Port     int
	Database db.DataAccessLayer
}

type Server struct {
	sc          *ServerConfig
	RaidNotify  chan types.RaidNotification
	DiscordAuth chan types.DiscordAuth
	AuthSuccess chan types.DiscordAuth
	ChatChan    chan string
	ChatOutChan chan types.ChatMessage
}

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

func (s *Server) Serve() {

	go s.authHandler()
	go s.raidAlerter()

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

func (s *Server) raidAlerter() {
	db := s.sc.Database.Copy()
	defer db.Close()

	for {
		time.Sleep(10)
		var results []types.RaidNotification
		db.RaidAlerts().GetReady(&results)

		if len(results) > 0 {
			for _, result := range results {
				s.RaidNotify <- result
				db.RaidAlerts().Remove(result)
			}
		}
	}
}

func (s *Server) authHandler() {
	db := s.sc.Database.Copy()
	defer db.Close()

	for {
		as := <-s.AuthSuccess

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
	}
}

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

func handleError(w http.ResponseWriter, restError types.RESTError) {
	w.WriteHeader(restError.StatusCode)
	err := json.NewEncoder(w).Encode(restError)
	if err != nil {
		// panic(err)
		log.Printf(logSymbol+"Error encoding %v, %s\n", restError, err)
	}
}
