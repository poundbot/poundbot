package types

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	GuildSnowflake  string
	OwnerSnowflake  string
	CommandPrefix   string
	AdminSnowflakes []string `bson:",omitempty"`
}

type Account struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	BaseAccount `bson:",inline" json:",inline"`
	Servers     []Server `bson:",omitempty"`
	Timestamp   `bson:",inline" json:",inline"`
	Disabled    bool
}

// ServerFromKey finds a server for a given key. Errors if not found.
func (a Account) ServerFromKey(apiKey string) (Server, error) {
	for i := range a.Servers {
		if a.Servers[i].Key == apiKey {
			return a.Servers[i], nil
		}
	}
	return Server{}, errors.New("server not found")
}

// GetCommandPrefix the Discord command prefix. Defaults to "!pb"
func (a Account) GetCommandPrefix() string {
	if a.CommandPrefix == "" {
		return "!pb"
	}
	return a.CommandPrefix
}

func (a Account) GetAdminIDs() []string {
	return append(a.AdminSnowflakes, a.OwnerSnowflake)
}

// Clan is a clan from the game
type Clan struct {
	Tag        string
	OwnerID    string
	Members    []string
	Moderators []string
}

// SetGame adds game name to all IDs
func (c *Clan) SetGame(game string) {
	c.OwnerID = fmt.Sprintf("%s:%s", game, c.OwnerID)
	for i := range c.Members {
		c.Members[i] = fmt.Sprintf("%s:%s", game, c.Members[i])
	}

	for i := range c.Moderators {
		c.Moderators[i] = fmt.Sprintf("%s:%s", game, c.Moderators[i])
	}
}
