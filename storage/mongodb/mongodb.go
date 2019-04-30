package mongodb

import (
	"context"
	"log"
	"time"

	"github.com/poundbot/poundbot/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	accountsCollection     = "accounts"
	chatsCollection        = "chats"
	discordAuthsCollection = "discord_auths"
	raidAlertsCollection   = "raid_alerts"
	usersCollection        = "users"
)

func upsertOptions() *options.UpdateOptions {
	u := true
	return &options.UpdateOptions{Upsert: &u}
}

// A Config is exactly what it sounds like.
type Config struct {
	DialAddress string // the mgo.Dial address
	Database    string // the database name
}

// NewMongoDB returns a connected Mgo
func NewMongoDB(mc Config) (*MongoDb, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	sess, err := mongo.Connect(ctx, options.Client().ApplyURI(mc.DialAddress))

	if err != nil {
		return nil, err
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = sess.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return &MongoDb{session: *sess, dbname: mc.Database}, nil
}

// An MongoDb implements storage.Storage for MongoDB
type MongoDb struct {
	dbname  string
	session mongo.Client
}

// Copy implements storage.Storage.Copy
func (m MongoDb) Copy() storage.Storage {
	return m
}

// Close implements storage.Storage.Close
func (m MongoDb) Close() {}

// Users implements storage.Storage.Users
func (m MongoDb) Users() storage.UsersStore {
	return Users{collection: *m.session.Database(m.dbname).Collection(usersCollection)}
}

// DiscordAuths implements storage.Storage.DiscordAuths
func (m MongoDb) DiscordAuths() storage.DiscordAuthsStore {
	return DiscordAuths{collection: *m.session.Database(m.dbname).Collection(discordAuthsCollection)}
}

// RaidAlerts implements storage.Storage.RaidAlerts
func (m MongoDb) RaidAlerts() storage.RaidAlertsStore {
	return storage.RaidAlertsStore(RaidAlerts{
		collection: *m.session.Database(m.dbname).Collection(raidAlertsCollection),
		users:      m.Users(),
	})
}

// ServerAccounts implements storage.Storage.ServerAccounts
func (m MongoDb) Accounts() storage.AccountsStore {
	return Accounts{collection: *m.session.Database(m.dbname).Collection(accountsCollection)}
}

// Init implements storage.Storage.Init
func (m MongoDb) Init() {
	log.Printf("Database is %s\n", m.dbname)
	// mongoDB := m.session.Database(m.dbname)
	// userIndexes := mongoDB.Collection(usersCollection).Indexes()
	// discordAuthIndexes := mongoDB.Collection(discordAuthsCollection).Indexes()
	// accountsIndexes := mongoDB.Collection(accountsCollection).Indexes()

	// userIndexes.CreateOne(
	// 	nil,
	// 	mongo.IndexModel{
	// 		Keys: }
	// 	)(mgo.Index{
	// 	Key:      []string{"playerids"},
	// 	Unique:   true,
	// 	DropDups: true,
	// })

	// discordAuthIndexes.EnsureIndex(mgo.Index{
	// 	Key:      []string{"playerid"},
	// 	Unique:   true,
	// 	DropDups: true,
	// })

	// accountsIndexes.EnsureIndex(mgo.Index{
	// 	Key:    []string{"servers.key"},
	// 	Unique: false,
	// })
}
