package mongodb

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/globalsign/mgo"
	pblog "github.com/poundbot/poundbot/log"
	"github.com/poundbot/poundbot/storage"
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
	DialAddress string // the mgo.Dial address
	Database    string // the database name
	SSL         bool
	InsecureSSL bool
}

// NewMongoDB returns a connected Mgo
func NewMongoDB(dial, db string) (*MongoDB, error) {
	var sErr error
	var sess *mgo.Session

	mc, err := parseDialURL(dial)
	if err != nil {
		return nil, fmt.Errorf("invalid dial url: %w", err)
	}

	mc.Database = db

	if mc.SSL {
		dialInfo, err := mgo.ParseURL(mc.DialAddress)
		if err != nil {
			log.Println(err)
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: mc.InsecureSSL,
			}
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			if err != nil {
				log.Println(err)
			}
			return conn, err
		}
		sess, sErr = mgo.DialWithInfo(dialInfo)
	} else {
		sess, sErr = mgo.Dial(mc.DialAddress)
	}
	if sErr != nil {
		return nil, sErr
	}
	return &MongoDB{session: sess, dbname: mc.Database}, nil
}

// An MongoDB implements storage for MongoDB
type MongoDB struct {
	dbname  string
	session *mgo.Session
}

// Copy implements storage.Storage.Copy
func (m *MongoDB) Copy() storage.Storage {
	return &MongoDB{dbname: m.dbname, session: m.session.Copy()}
}

func (m *MongoDB) ChatQueue() storage.ChatQueueStore {
	return ChatQueue{collection: m.session.DB(m.dbname).C(chatQueueCollection)}
}

// MessageLocks implements MessageLocks
func (m *MongoDB) MessageLocks() storage.MessageLocksStore {
	return MessageLocks{collection: m.session.DB(m.dbname).C(messageLocksCollection)}
}

// Close implements Close
func (m *MongoDB) Close() {
	m.session.Close()
}

// Users implements storage.Storage.Users
func (m *MongoDB) Users() storage.UsersStore {
	return Users{collection: m.session.DB(m.dbname).C(usersCollection)}
}

// DiscordAuths implements storage.Storage.DiscordAuths
func (m *MongoDB) DiscordAuths() storage.DiscordAuthsStore {
	return DiscordAuths{collection: m.session.DB(m.dbname).C(discordAuthsCollection)}
}

// RaidAlerts implements storage.Storage.RaidAlerts
func (m *MongoDB) RaidAlerts() storage.RaidAlertsStore {
	return RaidAlerts{
		collection: m.session.DB(m.dbname).C(raidAlertsCollection),
		users:      m.Users(),
	}
}

// Accounts implements storage.Storage.ServerAccounts
func (m *MongoDB) Accounts() storage.AccountsStore {
	return Accounts{collection: m.session.DB(m.dbname).C(accountsCollection)}
}

// Init implements Init
func (m *MongoDB) Init() {
	log.Printf("Database is %s", m.dbname)
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

func parseDialURL(dialURL string) (*Config, error) {
	u, err := url.Parse(dialURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "mongodb" {
		return nil, errors.New("scheme for mongodb dial should be mongodb://")
	}

	c := Config{DialAddress: fmt.Sprintf("mongodb://%s/", u.Host)}

	q := u.Query()

	if q.Get("ssl") == "true" || q.Get("tls") == "true" {
		q.Del("ssl")
		q.Del("tls")

		c.SSL = true
		if q.Get("tlsInsecure") == "true" {
			q.Del("tlsInsecure")
			c.InsecureSSL = true
		}
	}

	u.RawQuery = q.Encode()

	// u.RawQuery = u.Query().Encode()

	c.DialAddress = u.String()
	return &c, nil
}
