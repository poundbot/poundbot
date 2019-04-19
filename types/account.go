package types

import (
	"errors"

	"github.com/globalsign/mgo/bson"
)

type Server struct {
	Name         string
	Key          string
	Address      string
	Clans        []Clan
	ChatChanID   string
	ServerChanID string
	RaidDelay    string
	Timestamp    `bson:",inline"`
}

type BaseAccount struct {
	GuildSnowflake string
	OwnerSnowflake string
}

type Account struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	BaseAccount `bson:",inline" json:",inline"`
	Servers     []Server
	Timestamp   `bson:",inline" json:",inline"`
	Disabled    bool
}

// A ServerClan is a clan a Rust server sends
type ServerClan struct {
	Tag         string
	Owner       string
	Description string
	Members     []string
	Moderators  []string
	Invited     []string
}

type Clan struct {
	Tag         string
	OwnerID     string
	Description string
	Members     []string
	Moderators  []string
	Invited     []string
}

func (a Account) ServerFromKey(apiKey string) (Server, error) {
	for i := range a.Servers {
		if a.Servers[i].Key == apiKey {
			return a.Servers[i], nil
		}
	}
	return Server{}, errors.New("server not found")
}

func (s Server) UsersClan(GameUserID string) *Clan {
	for _, serverClan := range s.Clans {
		for _, member := range serverClan.Members {
			if member == GameUserID {
				return &serverClan
			}
		}
	}
	return nil
}

// ClanFromServerClan Converts strings to uints
func ClanFromServerClan(sc ServerClan) (*Clan, error) {
	var clan = Clan{}
	clan.Tag = sc.Tag
	clan.Description = sc.Description
	clan.OwnerID = sc.Owner

	clan.Members = sc.Members

	clan.Moderators = sc.Moderators

	clan.Invited = sc.Invited

	return &clan, nil
}
