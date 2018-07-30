package rustconn

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

const (
	SourceRust    = "rust"
	SourceDiscord = "discord"
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
	raidAlertColl   *mgo.Collection
	chatCollection  *mgo.Collection
	clanCollection  *mgo.Collection
	RaidNotify      chan RaidNotification
	DiscordAuth     chan DiscordAuth
	AuthSuccess     chan DiscordAuth
	ChatChan        chan string
	ChatOutChan     chan ChatMessage
}

func NewServer(sc *ServerConfig, rch chan RaidNotification, dach, asch chan DiscordAuth, cch chan string, coch chan ChatMessage) *Server {
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
	session, err := mgo.Dial(s.sc.MongoConfig.DialAddress)
	if err != nil {
		log.Fatal(err)
	}
	s.mongoDB = session.DB(s.sc.MongoConfig.Database)
	s.userColl = s.mongoDB.C("users")
	s.discordAuthColl = s.mongoDB.C("discord_auths")
	s.raidAlertColl = s.mongoDB.C("raid_alerts")
	s.chatCollection = s.mongoDB.C("chats")
	s.clanCollection = s.mongoDB.C("clans")

	index := mgo.Index{
		Key:        []string{"steam_id"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	s.userColl.EnsureIndex(index)
	s.discordAuthColl.EnsureIndex(index)

	index = mgo.Index{
		Key:        []string{"tag", "players"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}

	go s.authHandler()
	go s.raidAlerter()

	fmt.Printf("Starting HTTP Server on %s:%d\n", s.sc.BindAddr, s.sc.Port)
	r := mux.NewRouter()
	r.HandleFunc("/entity_death", s.entityDeathHandler)
	r.HandleFunc("/discord_auth", s.discordAuthHandler)
	r.HandleFunc("/chat", s.chatHandler)
	r.HandleFunc("/clans", s.clansHandler).Methods(http.MethodPut)
	r.HandleFunc("/clans/{tag}", s.clanHandler).Methods(http.MethodDelete, http.MethodPut)
	http.Handle("/", r)
	go log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", s.sc.BindAddr, s.sc.Port), r))
}

func (s Server) upsertClan(c Clan) {
	_, err := s.clanCollection.Upsert(
		bson.M{
			"tag": c.Tag,
		},
		bson.M{
			"$setOnInsert": bson.M{"created_at": time.Now().UTC().Add(5 * time.Minute)},
			"$set":         c.ClanBase,
		},
	)
	if err != nil {
		log.Printf("%s", err)
	}
	s.userColl.UpdateAll(
		bson.M{"steam_id": bson.M{"$in": c.Members}},
		bson.M{"$set": bson.M{"clan_tag": c.Tag}},
	)
}

func (s *Server) raidAlerter() {
	for {
		time.Sleep(10)
		var results []RaidNotification
		err := s.raidAlertColl.Find(bson.M{"alert_at": bson.M{"$lte": time.Now().UTC()}}).All(&results)
		if err != nil {
			log.Printf("Error getting raid notifications! %s\n", err)
		}
		if len(results) > 0 {
			for _, result := range results {
				s.RaidNotify <- result
				s.raidAlertColl.Remove(result.DiscordInfo)
			}
		}
	}
}

func (s *Server) authHandler() {
	for {
		as := <-s.AuthSuccess

		_, err := s.userColl.Upsert(
			as.BaseUser.SteamInfo,
			bson.M{
				"$setOnInsert": bson.M{"created_at": time.Now().UTC().Add(5 * time.Minute)},
				"$set":         as.BaseUser,
			},
		)
		if err == nil {
			s.discordAuthColl.Remove(as.SteamInfo)
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
	log.Printf("%v", r.Body)
	decoder := json.NewDecoder(r.Body)
	var t []ServerClan
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err)
		return
	}

	clanCount := len(t)
	clans := make([]Clan, clanCount)
	tags := make([]string, clanCount)
	for i, sc := range t {
		c, err := ClanFromServerClan(sc)
		if err != nil {
			log.Printf("%v\n", err)
			handleError(w, RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}
		tags[i] = c.Tag
		clans[i] = *c
	}

	for _, clan := range clans {
		s.upsertClan(clan)

	}

	s.clanCollection.RemoveAll(bson.M{"tag": bson.M{"$nin": tags}})
	s.userColl.UpdateAll(
		bson.M{"clan_tag": bson.M{"$nin": tags}},
		bson.M{"$unset": bson.M{"clan_tag": 1}},
	)
}

func (s *Server) clanHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tag := vars["tag"]
	switch r.Method {
	case http.MethodDelete:
		s.clanCollection.RemoveAll(bson.M{"tag": tag})
		s.userColl.UpdateAll(
			bson.M{"clan_tag": tag},
			bson.M{"$unset": bson.M{"clan_tag": 1}},
		)
		return
	case http.MethodPut:
		decoder := json.NewDecoder(r.Body)
		var t ServerClan
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("%v", t)
		clan, err := ClanFromServerClan(t)
		if err != nil {
			handleError(w, RESTError{
				StatusCode: http.StatusBadRequest,
				Error:      "Error processing clan data",
			})
			return
		}

		s.upsertClan(*clan)
	}
}

func (s *Server) chatHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var t ChatMessage
		err := decoder.Decode(&t)
		if err != nil {
			log.Println(err)
			return
		}

		t.Source = SourceRust
		s.chatCollection.Insert(t)
		go func(t ChatMessage, c chan string) {
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
				fmt.Println(err)
				return
			}
			s.chatCollection.Insert(res)

			w.Write(b)
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusNoContent)
		}

	default:
		handleError(w, RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("Method %s not allowed", r.Method),
		})
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
			s.raidAlertColl.Upsert(
				u.DiscordInfo,
				bson.M{
					"$setOnInsert": bson.M{
						"alert_at": time.Now().UTC().Add(5 * time.Minute),
					},
					"$inc": bson.M{
						fmt.Sprintf("items.%s", t.Name): 1,
					},
					"$addToSet": bson.M{
						"grid_positions": t.GridPos,
					},
				},
			)
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
		handleError(w, RESTError{
			StatusCode: http.StatusMethodNotAllowed,
			Error:      fmt.Sprintf("%s is linked to you.", u.DiscordID),
		})
		return
	} else if t.DiscordID == "check" {
		handleError(w, RESTError{
			StatusCode: http.StatusNotFound,
			Error:      "Account is not linked to discord.",
		})
		return
	}

	_, err = s.discordAuthColl.Upsert(t.SteamInfo, t)
	if err == nil {
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

func handleError(w http.ResponseWriter, restError RESTError) {
	w.WriteHeader(restError.StatusCode)
	err := json.NewEncoder(w).Encode(restError)
	if err != nil {
		// panic(err)
		log.Printf("Error encoding %v, %s\n", restError, err)
	}
}
