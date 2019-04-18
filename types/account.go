package types

import (
	"errors"
	"strconv"

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
	OwnerID     uint64
	Description string
	Members     []uint64
	Moderators  []uint64
	Invited     []uint64
}

func (a Account) ServerFromKey(key string) (Server, error) {
	for i := range a.Servers {
		if a.Servers[i].Key == key {
			return a.Servers[i], nil
		}
	}
	return Server{}, errors.New("server not found")
}

func (s Server) UsersClan(steamID uint64) *Clan {
	for _, serverClan := range s.Clans {
		for _, member := range serverClan.Members {
			if member == steamID {
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
	i, err := strconv.ParseUint(sc.Owner, 10, 64)
	if err != nil {
		return nil, err
	}
	clan.OwnerID = i

	nuints, err := convStringAToUnintA(sc.Members)
	if err != nil {
		return nil, err
	}
	clan.Members = nuints

	nuints, err = convStringAToUnintA(sc.Moderators)
	if err != nil {
		return nil, err
	}
	clan.Moderators = nuints

	nuints, err = convStringAToUnintA(sc.Invited)
	if err != nil {
		return nil, err
	}
	clan.Invited = nuints

	return &clan, nil
}

func convStringAToUnintA(in []string) ([]uint64, error) {
	var out []uint64
	var l = len(in)
	if l == 0 {
		return out, nil
	}

	out = make([]uint64, len(in))
	for i, conv := range in {
		newuint, err := strconv.ParseUint(conv, 10, 64)
		if err != nil {
			return nil, err
		}
		out[i] = newuint
	}

	return out, nil
}
