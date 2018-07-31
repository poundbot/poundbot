package mgo

import (
	"log"

	"github.com/globalsign/mgo"
	"mrpoundsign.com/poundbot/db"
)

const (
	usersCollection        = "users"
	discordAuthsCollection = "discord_auths"
	raidAlertsCollection   = "raid_alerts"
	chatsCollection        = "chats"
	clansCollection        = "clans"
)

type MongoConfig struct {
	DialAddress string
	Database    string
}

func NewMgo(mc MongoConfig) *Mgo {
	sess, err := mgo.Dial(mc.DialAddress)
	if err != nil {
		log.Fatal(err)
	}
	return &Mgo{session: sess, dbname: mc.Database}
}

type Mgo struct {
	dbname  string
	session *mgo.Session
}

func (m Mgo) Copy() db.DataAccessLayer {
	return Mgo{session: m.session.Copy()}
}

func (m Mgo) Close() {
	m.session.Close()
}

func (m Mgo) Users() db.UsersAccessLayer {
	return Users{collection: m.session.DB(m.dbname).C(usersCollection)}
}

func (m Mgo) DiscordAuths() db.DiscordAuthsAccessLayer {
	return DiscordAuths{collection: m.session.DB(m.dbname).C(discordAuthsCollection)}
}

func (m Mgo) RaidAlerts() db.RaidAlertsAccessLayer {
	return db.RaidAlertsAccessLayer(RaidAlerts{
		collection: m.session.DB(m.dbname).C(raidAlertsCollection),
		users:      m.Users(),
	})
}

func (m Mgo) Clans() db.ClansAccessLayer {
	return Clans{
		collection: m.session.DB(m.dbname).C(clansCollection),
		users:      m.Users(),
	}
}

func (m Mgo) Chats() db.ChatsAccessLayer {
	return Chats{collection: m.session.DB(m.dbname).C(chatsCollection)}
}

func (m Mgo) CreateIndexes() {
	log.Printf("Session database is %s\n", m.dbname)
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

var _ = db.DataAccessLayer(Mgo{})
