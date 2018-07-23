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
}

func NewServer(sc *ServerConfig, ch chan RaidNotification) *Server {
	return &Server{sc: sc, RaidNotify: ch}
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

	u := User{DiscordInfo: DiscordInfo{DiscordID: "MrPoundsign#2364"}, SteamInfo: SteamInfo{SteamID: 76561197960794006}, CreatedAt: time.Now().UTC()}
	s.userColl.Upsert(u.UpsertID(), u)
	// s.RaidNotify = make(chan RaidNotification)
	fmt.Printf("Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
	http.HandleFunc("/entity_death", s.entityDeathHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.sc.BindAddr, s.sc.Port), nil))
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
		u, err := s.FindUser(steamID)
		if err == nil {
			// log.Printf("Found user %v", u)
			func(rn RaidNotification) {
				log.Printf("Sending raid notification! %v\n", rn)
				s.RaidNotify <- rn
				log.Printf("Sent notification! %v\n", rn)
			}(RaidNotification{DiscordID: u.DiscordID, Items: []RaidInventory{RaidInventory{Name: t.Name, Count: 1}}})
		} else {
			log.Printf("Could not find user for id %d: %v\n", steamID, err)
		}
	}
}

func (s *Server) FindUser(steamid uint64) (*User, error) {
	u := User{}
	err := s.userColl.Find(SteamInfo{SteamID: steamid}).One(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
