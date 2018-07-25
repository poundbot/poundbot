package rustconn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/globalsign/mgo"
)

type MongoConfig struct {
	DialAddress string
	Database    string
}

type ServerConfig struct {
	BindAddr    string
	Port        int
	MongoConfig MongoConfig
}

type Server struct {
	sc              *ServerConfig
	mongoDB         *mgo.Database
	userColl        *mgo.Collection
	discordAuthColl *mgo.Collection
	RaidNotify      chan RaidNotification
	DiscordAuth     chan DiscordAuth
	AuthSuccess     chan DiscordAuth
}

func NewServer(sc *ServerConfig, rch chan RaidNotification, dach chan DiscordAuth, asch chan DiscordAuth) *Server {
	return &Server{
		sc:          sc,
		RaidNotify:  rch,
		DiscordAuth: dach,
		AuthSuccess: asch,
	}
}

func (s *Server) Serve() {
	session, err := mgo.Dial(s.sc.MongoConfig.DialAddress)
	if err != nil {
		log.Fatal(err)
	}
	s.mongoDB = session.DB(s.sc.MongoConfig.Database)
	s.userColl = s.mongoDB.C("users")
	s.discordAuthColl = s.mongoDB.C("discord_auth")

	index := mgo.Index{
		Key:        []string{"steam_id"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	s.userColl.EnsureIndex(index)
	s.discordAuthColl.EnsureIndex(index)

	// u := User{DiscordInfo: DiscordInfo{DiscordID: "MrPoundsign#2364"}, SteamInfo: SteamInfo{SteamID: 76561197960794006}, CreatedAt: time.Now().UTC()}
	// s.userColl.Upsert(u.UpsertID(), u)
	go s.authHandler()
	fmt.Printf("Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
	http.HandleFunc("/entity_death", s.entityDeathHandler)
	http.HandleFunc("/discord_auth", s.discordAuthHandler)
	go log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.sc.BindAddr, s.sc.Port), nil))
}

func (s *Server) authHandler() {
	for {
		as := <-s.AuthSuccess
		u := User{DiscordInfo: as.DiscordInfo, SteamInfo: as.SteamInfo, CreatedAt: time.Now().UTC()}
		_, err := s.userColl.Upsert(u.UpsertID(), u)
		if err == nil {
			s.discordAuthColl.Remove(as.SteamInfo)
		}
	}
}

func (s *Server) entityDeathHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t EntityDeath
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err)
		return
	}
	for _, steamID := range t.Owners {
		u, err := s.findUser(steamID)
		if err == nil {
			func(rn RaidNotification) {
				s.RaidNotify <- rn
			}(RaidNotification{DiscordID: u.DiscordID, Items: []RaidInventory{RaidInventory{Name: t.Name, Count: 1}}})
		}
	}
}

func (s *Server) discordAuthHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t DiscordAuth
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("User Auth Request: %v from %v\n", t, r.Body)
	u, err := s.findUser(t.SteamID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"%s is already linked to the bot.\"}", u.DiscordID)))
		log.Println("User already exists")
		return
	}

	fmt.Printf("%v", t.UpsertID())
	_, err = s.discordAuthColl.Upsert(t.UpsertID(), t)
	if err == nil {
		fmt.Printf("Sending to channel: %v\n", t)
		s.DiscordAuth <- t
	} else {
		log.Println(err)
	}
}

func (s *Server) findUser(steamid uint64) (u *User, err error) {
	err = s.userColl.Find(SteamInfo{SteamID: steamid}).One(&u)
	if err != nil {
		return nil, err
	}
	return
}

func (s *Server) removeUser(steamid uint64) (err error) {
	err = s.userColl.Remove(SteamInfo{SteamID: steamid})
	return
}
