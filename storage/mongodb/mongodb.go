package mongodb

import (
	"context"
	"time"

	pblog "github.com/poundbot/poundbot/log"
	"github.com/poundbot/poundbot/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

type collection struct {
	options map[string]interface{}
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

func (m MongoDB) newConnection() (*mongo.Database, error) {
	var client *mongo.Client

	client, err := mongo.NewClient(options.Client().ApplyURI(m.address))
	if err != nil {
		return nil, err
	}

	err = client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}

	return client.Database(m.dbname), nil
}

// Copy implements storage.Storage.Copy
func (m *MongoDB) Copy() storage.Storage {
	return m
}

func (m *MongoDB) ChatQueue() storage.ChatQueueStore {
	return ChatQueue{collection: m.client.Database(m.dbname).Collection(chatQueueCollection), cloner: m}
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

func (m *MongoDB) ensureCapped(name string) {
	ecLog := log.WithField("sys", "enureCapped").WithField("coll", name)
	mongoDB := m.client.Database(m.dbname)
	c := mongoDB.RunCommand(context.Background(), //bson.D{
		bson.D{
			primitive.E{Key: "create", Value: name},
			primitive.E{Key: "capped", Value: true},
			primitive.E{Key: "size", Value: 16384},
			primitive.E{Key: "max", Value: 100},
		},
	)

	if c.Err() != nil {
		switch c.Err().(type) {
		case mongo.CommandError:
			if c.Err().(mongo.CommandError).Name == "NamespaceExists" {
				c = mongoDB.RunCommand(context.Background(), //bson.D{
					bson.D{
						primitive.E{Key: "convertToCapped", Value: name},
						primitive.E{Key: "capped", Value: true},
						primitive.E{Key: "size", Value: 16384},
						primitive.E{Key: "max", Value: 100},
					},
				)
				if c.Err() != nil {
					ecLog.WithError(c.Err()).Fatal("could not convert to capped")
				}
			}
		default:
			ecLog.WithError(c.Err()).Fatal("error running command")
		}
	}
}

// Init implements Init
func (m *MongoDB) Init() {
	log.Printf("Database is %s", m.dbname)
	mongoDB := m.client.Database(m.dbname)
	// discordAuthColl :=
	// accountColl := mongoDB.Collection(accountsCollection)
	// messageLocksColl := mongoDB.Collection(messageLocksCollection)
	// chatQueueColl := mongoDB.Collection(chatQueueCollection)

	m.ensureCapped(chatQueueCollection)
	m.ensureCapped(messageLocksCollection)

	mongoDB.Collection(messageLocksCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				primitive.E{Key: "messageid", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)

	mongoDB.Collection(usersCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				primitive.E{Key: "playerids", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)

	mongoDB.Collection(discordAuthsCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				primitive.E{Key: "playerid", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)

	mongoDB.Collection(accountsCollection).Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				primitive.E{Key: "servers.key", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
}
