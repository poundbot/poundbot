package db

import (
	"errors"
	"log"

	mgo "github.com/globalsign/mgo"
)

var (
	session  *mgo.Session
	database string
)

const (
	userCollection        = "users"
	discordAuthCollection = "discord_auths"
	raidAlertCollection   = "raid_alerts"
	chatCollection        = "chats"
	clanCollection        = "clans"
)

type MongoConfig struct {
	DialAddress string
	Database    string
}

type Session struct {
	session  *mgo.Session
	database string
}

func Init(mc MongoConfig) {
	sess, err := mgo.Dial(mc.DialAddress)
	if err != nil {
		log.Fatal(err)
	}
	session = sess
	database = mc.Database
	log.Printf("Session database is %s\n", database)
	mongoDB := session.DB(database)
	userColl := mongoDB.C("users")
	discordAuthColl := mongoDB.C("discord_auths")
	clanCollection := mongoDB.C("clans")

	index := mgo.Index{
		Key:        []string{"steam_id"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	userColl.EnsureIndex(index)
	discordAuthColl.EnsureIndex(index)

	index = mgo.Index{
		Key:        []string{"tag", "members"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	clanCollection.EnsureIndex(index)
}

func NewSession() (*Session, error) {
	if session == nil {
		return nil, errors.New("Not Connected")
	}
	return &Session{session: session.Copy(), database: database}, nil
}

func (s *Session) Close() {
	if s.session == nil {
		return
	}
	s.session.Close()
	s.session = nil
}

func (s *Session) UserCollection() *mgo.Collection {
	return s.collection(userCollection)
}

func (s *Session) DiscordAuthCollection() *mgo.Collection {
	return s.collection(discordAuthCollection)
}

func (s *Session) RaidAlertCollection() *mgo.Collection {
	return s.collection(raidAlertCollection)
}

func (s *Session) ChatCollection() *mgo.Collection {
	return s.collection(chatCollection)
}

func (s *Session) ClanCollection() *mgo.Collection {
	return s.collection(clanCollection)
}

func (s *Session) collection(collection string) *mgo.Collection {
	// log.Printf("Database is %s\n", s.database)
	return s.session.DB(s.database).C(collection)
}
