package mongodb

import (
	"log"

	"bitbucket.org/mrpoundsign/poundbot/storage"
	"github.com/globalsign/mgo"
)

const (
	accountsCollection     = "accounts"
	chatsCollection        = "chats"
	discordAuthsCollection = "discord_auths"
	raidAlertsCollection   = "raid_alerts"
	usersCollection        = "users"
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

// An MongoDb implements storage.Storage for MongoDB
type MongoDb struct {
	dbname  string
	session *mgo.Session
}

// Copy implements storage.Storage.Copy
func (m MongoDb) Copy() storage.Storage {
	return MongoDb{dbname: m.dbname, session: m.session.Copy()}
}

// Close implements storage.Storage.Close
func (m MongoDb) Close() {
	m.session.Close()
}

// Users implements storage.Storage.Users
func (m MongoDb) Users() storage.UsersStore {
	return Users{collection: m.session.DB(m.dbname).C(usersCollection)}
}

// DiscordAuths implements storage.Storage.DiscordAuths
func (m MongoDb) DiscordAuths() storage.DiscordAuthsStore {
	return DiscordAuths{collection: m.session.DB(m.dbname).C(discordAuthsCollection)}
}

// RaidAlerts implements storage.Storage.RaidAlerts
func (m MongoDb) RaidAlerts() storage.RaidAlertsStore {
	return storage.RaidAlertsStore(RaidAlerts{
		collection: m.session.DB(m.dbname).C(raidAlertsCollection),
		users:      m.Users(),
	})
}

// ServerAccounts implements storage.Storage.ServerAccounts
func (m MongoDb) Accounts() storage.AccountsStore {
	return Accounts{collection: m.session.DB(m.dbname).C(accountsCollection)}
}

// Init implements storage.Storage.Init
func (m MongoDb) Init() {
	log.Printf("Database is %s\n", m.dbname)
	mongoDB := m.session.DB(m.dbname)
	userColl := mongoDB.C(usersCollection)
	discordAuthColl := mongoDB.C(discordAuthsCollection)
	// accountsColl := mongoDB.C(accountsCollection)

	index := mgo.Index{
		Key:        []string{"steamid"},
		Unique:     true,
		DropDups:   true,
		Background: false,
		Sparse:     false,
	}
	userColl.EnsureIndex(index)
	discordAuthColl.EnsureIndex(index)
}
