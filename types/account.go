package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/globalsign/mgo/bson"
)

type Server struct {
	Name      string
	Key       string
	Address   string
	Clans     []Clan
	RaidDelay string
	Timestamp `bson:",inline"`
	Channels  []ServerChannel `bson:",omitempty" json:"channels"`
}

func (s Server) ChannelIDForTag(tag string) (channel string, found bool) {
	for i := range s.Channels {
		for _, cTag := range s.Channels[i].Tags {
			if tag == cTag {
				return s.Channels[i].ChannelID, true
			}
		}
	}
	return "", false
}

func (s *Server) SetChannelIDForTag(channel string, tag string) {
	var channelFound bool

	for i := range s.Channels {
		var sChan = s.Channels[i].ChannelID
		var sTags = s.Channels[i].Tags
		for j, cTag := range sTags {
			if tag == cTag {
				if sChan == channel {
					return
				}
				if len(s.Channels[i].Tags) < 2 {
					s.Channels[i].ChannelID = channel
					return
				}

				s.Channels[i].Tags = append(sTags[:j], sTags[j+1:]...)
				continue
			}
		} // for tags
		if sChan == channel {
			s.Channels[i].Tags = append(sTags, tag)
			channelFound = true
			continue
		}
	} // for channels
	if channelFound {
		return
	}
	s.Channels = append(s.Channels, ServerChannel{ChannelID: channel, Tags: []string{tag}})
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

type ServerChannel struct {
	ChannelID string `bson:"channel_id" json:"channel_id"`
	Tags      []string
}

type BaseAccount struct {
	GuildSnowflake      string
	OwnerSnowflake      string
	CommandPrefix       string
	AdminSnowflakes     []string `bson:",omitempty"`
	RegisteredPlayerIDs []string `bson:",omitempty"`
}

type Account struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
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

// GetAdminIDs returns the Discord IDs considered "admins"
func (a Account) GetAdminIDs() []string {
	return append(a.AdminSnowflakes, a.OwnerSnowflake)
}

// GetRegisteredPlayerIDs returns a list of player IDs
// for the game requested. These ids are stripped of their
// prefix (e.g. "rust:1001" would be "1001")
func (a Account) GetRegisteredPlayerIDs(game string) []string {
	ids := []string{}
	gamePrefix := game + ":"
	for _, id := range a.RegisteredPlayerIDs {
		if strings.HasPrefix(id, gamePrefix) {
			ids = append(ids, id[len(gamePrefix):])
		}
	}
	return ids
}

// Clan is a clan from the game
type Clan struct {
	Tag        string
	OwnerID    string
	Members    []string `bson:",omitempty"`
	Moderators []string `bson:",omitempty"`
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
