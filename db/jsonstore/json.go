package jsonstore

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
	scribble "github.com/nanobox-io/golang-scribble"
)

type s struct {
	driver     *scribble.Driver
	collection string
	resource   string
	mux        sync.Mutex
}

type discordAuthsStore struct {
	s
	d map[uint64]types.DiscordAuth
}

type usersStore struct {
	s
	d map[uint64]types.User
}

type raidAlertsStore struct {
	s
	d map[string]types.RaidNotification
}

var (
	discordAuths = discordAuthsStore{
		s: s{
			collection: "discord_auths",
		},
		d: map[uint64]types.DiscordAuth{},
	}
	users = usersStore{
		s: s{
			collection: "users",
		},
		d: map[uint64]types.User{},
	}

	raidAlerts = raidAlertsStore{
		s: s{
			collection: "raid_alerts",
		},
		d: map[string]types.RaidNotification{},
	}
)

type Json struct {
	dataPath string
}

func NewJson(dataPath string) *Json {
	return &Json{dataPath: dataPath}
}

func (j *Json) Init() {
	db, err := scribble.New(j.dataPath, nil)
	if err != nil {
		log.Panicln("Error", err)
	}

	for _, name := range []string{"users", "discord_auths", "raid_alerts"} {
		err := os.MkdirAll(filepath.Join(j.dataPath, name), os.ModePerm)
		if err != nil {
			log.Panicln("Error creating DB", err)
		}
	}
	users.s.driver = db
	records, err := users.s.driver.ReadAll(users.s.collection)
	if err != nil {
		log.Panicln("Users ReadAll Error", err)
	}

	// Slurp in users
	for _, f := range records {
		found := types.User{}
		if err := json.Unmarshal([]byte(f), &found); err != nil {
			fmt.Println("Error", err)
		}
		users.d[found.SteamID] = found
	}

	// Slurp in discord auths
	discordAuths.s.driver = db
	records, err = discordAuths.s.driver.ReadAll(discordAuths.s.collection)
	if err != nil {
		panic(err)
	} else {
		for _, f := range records {
			foundDa := types.DiscordAuth{}
			if err := json.Unmarshal([]byte(f), &foundDa); err != nil {
				fmt.Println("Error", err)
			}
			discordAuths.d[foundDa.SteamID] = foundDa
		}
	}

	// Slurp in raid alerts
	raidAlerts.s.driver = db
	records, err = raidAlerts.s.driver.ReadAll(raidAlerts.s.collection)
	if err != nil {
		panic(err)
	} else {
		for _, f := range records {
			foundRa := types.RaidNotification{}
			if err := json.Unmarshal([]byte(f), &foundRa); err != nil {
				fmt.Println("Error", err)
			}
			raidAlerts.d[foundRa.DiscordID] = foundRa
		}
	}
}

// Copy returns itself
func (j *Json) Copy() db.DataStore {
	return j
}

// Chats creates a dummy Chats
func (j *Json) Chats() db.ChatsStore {
	return Chats{}
}

func (j *Json) Clans() db.ClansStore {
	return Clans{}
}

func (j *Json) DiscordAuths() db.DiscordAuthsStore {
	return DiscordAuths{}
}

func (j *Json) RaidAlerts() db.RaidAlertsStore {
	return RaidAlerts{users: j.Users()}
}

func (j *Json) Users() db.UsersStore {
	return Users{}
}

// Close does nothing
func (j Json) Close() {}
