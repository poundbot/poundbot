package types

import (
	"errors"
	"fmt"

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

func (s Server) UsersClan(playerIDs []string) (bool, Clan) {
	for _, clan := range s.Clans {

		for _, member := range clan.Members {
			for _, id := range playerIDs {
				if member == id {
					return true, clan
				}
			}
		}
	}
	return false, Clan{}
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

func (a Account) ServerFromKey(apiKey string) (Server, error) {
	for i := range a.Servers {
		if a.Servers[i].Key == apiKey {
			return a.Servers[i], nil
		}
	}
	return Server{}, errors.New("server not found")
}

// Clan is a clan from the game
type Clan struct {
	Tag         string
	Owner       string
	Description string
	Members     []string
	Moderators  []string
	Invited     []string
}

// SetGame adds game name to all IDs
func (c *Clan) SetGame(game string) {
	c.Owner = fmt.Sprintf("%s:%s", game, c.Owner)
	for i := range c.Members {
		c.Members[i] = fmt.Sprintf("%s:%s", game, c.Members[i])
	}

	for i := range c.Moderators {
		c.Moderators[i] = fmt.Sprintf("%s:%s", game, c.Moderators[i])
	}

	for i := range c.Invited {
		c.Invited[i] = fmt.Sprintf("%s:%s", game, c.Invited[i])
	}
}
