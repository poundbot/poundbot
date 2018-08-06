package mongodb

import (
	"log"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"github.com/globalsign/mgo"
)

const (
	usersCollection        = "users"
	discordAuthsCollection = "discord_auths"
	raidAlertsCollection   = "raid_alerts"
	chatsCollection        = "chats"
	clansCollection        = "clans"
)

// A Config is exactly what it sounds like.
type Config struct {
	DialAddress string // the mgo.Dial address
	Database    string // the database name
}

// NewMongoDB returns a connected Mgo
func NewMongoDB(mc Config) (*MongoDb, error) {
	sess, err := mgo.Dial(mc.DialAddress)
	if err != nil {
		sess.Close()
		return nil, err
	}
	return &MongoDb{session: sess, dbname: mc.Database}, nil
}

// An MongoDb implements db.DataStore for MongoDB
type MongoDb struct {
	dbname  string
	session *mgo.Session
}

// Copy implements db.DataStore.Copy
func (m MongoDb) Copy() db.DataStore {
	return MongoDb{session: m.session.Copy()}
}

// Close implements db.DataStore.Close
func (m MongoDb) Close() {
	m.session.Close()
}

// Users implements db.DataStore.Users
func (m MongoDb) Users() db.UsersStore {
	return Users{collection: m.session.DB(m.dbname).C(usersCollection)}
}

// DiscordAuths implements db.DataStore.DiscordAuths
func (m MongoDb) DiscordAuths() db.DiscordAuthsStore {
	return DiscordAuths{collection: m.session.DB(m.dbname).C(discordAuthsCollection)}
}

// RaidAlerts implements db.DataStore.RaidAlerts
func (m MongoDb) RaidAlerts() db.RaidAlertsStore {
	return db.RaidAlertsStore(RaidAlerts{
		collection: m.session.DB(m.dbname).C(raidAlertsCollection),
		users:      m.Users(),
	})
}

// Clans implements db.DataStore.Clans
func (m MongoDb) Clans() db.ClansStore {
	return Clans{
		collection: m.session.DB(m.dbname).C(clansCollection),
		users:      m.Users(),
	}
}

// Chats implements db.DataStore.Chats
func (m MongoDb) Chats() db.ChatsStore {
	return Chats{collection: m.session.DB(m.dbname).C(chatsCollection)}
}

// Init implements db.DataStore.Init
func (m MongoDb) Init() {
	log.Printf("Database is %s\n", m.dbname)
	mongoDB := m.session.DB(m.dbname)
	userColl := mongoDB.C(usersCollection)
	discordAuthColl := mongoDB.C(discordAuthsCollection)
	clanCollection := mongoDB.C(clansCollection)

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
