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

// A MongoConfig is exactly what it sounds like.
type MongoConfig struct {
	DialAddress string // the mgo.Dial address
	Database    string // the database name
}

// NewMgo returns a connected Mgo
func NewMgo(mc MongoConfig) (*Mgo, error) {
	sess, err := mgo.Dial(mc.DialAddress)
	if err != nil {
		sess.Close()
		return nil, err
	}
	return &Mgo{session: sess, dbname: mc.Database}, nil
}

// An Mgo implements db.DataStore in MongoDB using Mgo
type Mgo struct {
	dbname  string
	session *mgo.Session
}

// Copy implements db.DataStore.Copy
func (m Mgo) Copy() db.DataStore {
	return Mgo{session: m.session.Copy()}
}

// Close implements db.DataStore.Close
func (m Mgo) Close() {
	m.session.Close()
}

// Users implements db.DataStore.Users
func (m Mgo) Users() db.UsersStore {
	return Users{collection: m.session.DB(m.dbname).C(usersCollection)}
}

// DiscordAuths implements db.DataStore.DiscordAuths
func (m Mgo) DiscordAuths() db.DiscordAuthsStore {
	return DiscordAuths{collection: m.session.DB(m.dbname).C(discordAuthsCollection)}
}

// RaidAlerts implements db.DataStore.RaidAlerts
func (m Mgo) RaidAlerts() db.RaidAlertsStore {
	return db.RaidAlertsStore(RaidAlerts{
		collection: m.session.DB(m.dbname).C(raidAlertsCollection),
		users:      m.Users(),
	})
}

// Clans implements db.DataStore.Clans
func (m Mgo) Clans() db.ClansStore {
	return Clans{
		collection: m.session.DB(m.dbname).C(clansCollection),
		users:      m.Users(),
	}
}

// Chats implements db.DataStore.Chats
func (m Mgo) Chats() db.ChatsStore {
	return Chats{collection: m.session.DB(m.dbname).C(chatsCollection)}
}

// CreateIndexes implements db.DataStore.CreateIndexes
func (m Mgo) CreateIndexes() {
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
