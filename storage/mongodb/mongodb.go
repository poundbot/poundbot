package mongodb

import (
	"context"
	"time"

	pblog "github.com/poundbot/poundbot/log"
	"github.com/poundbot/poundbot/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var log = pblog.Log.WithField("sys", "MONGO")

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
	DialAddress string // the mongo dial address
	Database    string // the database name
}

// NewMongoDB returns a connected Mgo
func NewMongoDB(mc Config) (*MongoDB, error) {
	// var sErr error
	var client *mongo.Client

	client, err := mongo.NewClient(options.Client().ApplyURI(mc.DialAddress))
	if err != nil {
		return nil, err
	}

	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return &MongoDB{client: client, dbname: mc.Database, address: mc.DialAddress}, nil
}

// An MongoDB implements storage for MongoDB
type MongoDB struct {
	dbname  string
	address string
	client  *mongo.Client
}

// Copy implements storage.Storage.Copy
func (m *MongoDB) Copy() storage.Storage {
	return m
}

func (m *MongoDB) ChatQueue() storage.ChatQueueStore {
	return ChatQueue{collection: m.client.Database(m.dbname).Collection(chatQueueCollection)}
}

// MessageLocks implements MessageLocks
func (m *MongoDB) MessageLocks() storage.MessageLocksStore {
	return MessageLocks{collection: m.client.Database(m.dbname).Collection(messageLocksCollection)}
}

// Close implements Close
func (m *MongoDB) Close() {
	// m.client.Disconnect()
}

// Users implements storage.Storage.Users
func (m *MongoDB) Users() storage.UsersStore {
	return Users{collection: m.client.Database(m.dbname).Collection(usersCollection)}
}

// DiscordAuths implements storage.Storage.DiscordAuths
func (m *MongoDB) DiscordAuths() storage.DiscordAuthsStore {
	return DiscordAuths{collection: m.client.Database(m.dbname).Collection(discordAuthsCollection)}
}

// RaidAlerts implements storage.Storage.RaidAlerts
func (m *MongoDB) RaidAlerts() storage.RaidAlertsStore {
	return RaidAlerts{
		collection: m.client.Database(m.dbname).Collection(raidAlertsCollection),
		users:      m.Users(),
	}
}

// Accounts implements storage.Storage.ServerAccounts
func (m *MongoDB) Accounts() storage.AccountsStore {
	return Accounts{collection: m.client.Database(m.dbname).Collection(accountsCollection)}
}

// Init implements Init
func (m *MongoDB) Init() {
	log.Printf("Database is %s", m.dbname)
	// mongoDB := m.client.Database(m.dbname)
	// userColl := mongoDB.Collection(usersCollection)
	// discordAuthColl := mongoDB.Collection(discordAuthsCollection)
	// accountColl := mongoDB.Collection(accountsCollection)
	// messageLocksColl := mongoDB.Collection(messageLocksCollection)
	// chatQueueColl := mongoDB.Collection(chatQueueCollection)

	// chatQueueColl.Create(&mgo.CollectionInfo{
	// 	Capped:   true,
	// 	MaxBytes: 16384,
	// 	MaxDocs:  1000,
	// })

	// messageLocksColl.Create(&mgo.CollectionInfo{
	// 	Capped:   true,
	// 	MaxBytes: 16384,
	// 	MaxDocs:  1000,
	// })

	// messageLocksColl.EnsureIndex(mgo.Index{
	// 	Key:      []string{"messageid"},
	// 	Unique:   true,
	// 	DropDups: true,
	// })

	// userColl.EnsureIndex(mgo.Index{
	// 	Key:      []string{"playerids"},
	// 	Unique:   true,
	// 	DropDups: true,
	// })

	// discordAuthColl.EnsureIndex(mgo.Index{
	// 	Key:      []string{"playerid"},
	// 	Unique:   true,
	// 	DropDups: true,
	// })

	// accountColl.EnsureIndex(mgo.Index{
	// 	Key:    []string{"servers.key"},
	// 	Unique: false,
	// })
}
