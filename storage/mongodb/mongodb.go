package mongodb

import (
	"log"

	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/storage"
)

const (
	accountsCollection     = "accounts"
	chatsCollection        = "chats"
	discordAuthsCollection = "discord_auths"
	raidAlertsCollection   = "raid_alerts"
	usersCollection        = "users"
	messageLocksCollection = "message_locks"
	chatQueueCollection    = "chat_queue"
)

// A Config is exactly what it sounds like.
type Config struct {
	DialAddress string // the mgo.Dial address
	Database    string // the database name
}

// NewMongoDB returns a connected Mgo
func NewMongoDB(mc Config) (*MongoDB, error) {
	sess, err := mgo.Dial(mc.DialAddress)
	if err != nil {
		return nil, err
	}
	return &MongoDB{session: sess, dbname: mc.Database}, nil
}

// An MongoDB implements storage for MongoDB
type MongoDB struct {
	dbname  string
	session *mgo.Session
}

// Copy implements storage.Storage.Copy
func (m MongoDB) Copy() storage.Storage {
	return MongoDB{dbname: m.dbname, session: m.session.Copy()}
}

func (m MongoDB) ChatQueue() storage.ChatQueueStore {
	return ChatQueue{collection: m.session.DB(m.dbname).C(chatQueueCollection)}
}

// MessageLocks implements MessageLocks
func (m MongoDB) MessageLocks() storage.MessageLocksStore {
	return MessageLocks{collection: m.session.DB(m.dbname).C(messageLocksCollection)}
}

// Close implements Close
func (m MongoDB) Close() {
	m.session.Close()
}

// Users implements storage.Storage.Users
func (m MongoDB) Users() storage.UsersStore {
	return Users{collection: m.session.DB(m.dbname).C(usersCollection)}
}

// DiscordAuths implements storage.Storage.DiscordAuths
func (m MongoDB) DiscordAuths() storage.DiscordAuthsStore {
	return DiscordAuths{collection: m.session.DB(m.dbname).C(discordAuthsCollection)}
}

// RaidAlerts implements storage.Storage.RaidAlerts
func (m MongoDB) RaidAlerts() storage.RaidAlertsStore {
	return RaidAlerts{
		collection: m.session.DB(m.dbname).C(raidAlertsCollection),
		users:      m.Users(),
	}
}

// ServerAccounts implements storage.Storage.ServerAccounts
func (m MongoDB) Accounts() storage.AccountsStore {
	return Accounts{collection: m.session.DB(m.dbname).C(accountsCollection)}
}

// Init implements Init
func (m MongoDB) Init() {
	log.Printf("Database is %s\n", m.dbname)
	mongoDB := m.session.DB(m.dbname)
	userColl := mongoDB.C(usersCollection)
	discordAuthColl := mongoDB.C(discordAuthsCollection)
	accountColl := mongoDB.C(accountsCollection)
	messageLocksColl := mongoDB.C(messageLocksCollection)
	chatQueueColl := mongoDB.C(chatQueueCollection)

	chatQueueColl.Create(&mgo.CollectionInfo{
		Capped:   true,
		MaxBytes: 16384,
		MaxDocs:  1000,
	})

	messageLocksColl.Create(&mgo.CollectionInfo{
		Capped:   true,
		MaxBytes: 16384,
		MaxDocs:  1000,
	})

	messageLocksColl.EnsureIndex(mgo.Index{
		Key:      []string{"messageid"},
		Unique:   true,
		DropDups: true,
	})

	userColl.EnsureIndex(mgo.Index{
		Key:      []string{"playerids"},
		Unique:   true,
		DropDups: true,
	})

	discordAuthColl.EnsureIndex(mgo.Index{
		Key:      []string{"playerid"},
		Unique:   true,
		DropDups: true,
	})

	accountColl.EnsureIndex(mgo.Index{
		Key:    []string{"servers.key"},
		Unique: false,
	})
}
