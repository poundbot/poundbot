package types

import "strconv"

type Server struct {
	Key          string
	Address      string
	Clans        []Clan
	ChatChanID   string
	ServerChanID string
	RaidDelay    string
}

type BaseAccount struct {
	GuildSnowflake string
	OwnerSnowflake string
}

type Account struct {
	BaseAccount `bson:",inline" json:",inline"`
	Servers     []Server
	Timestamp   `bson:",inline" json:",inline"`
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
	clan.Members = *nuints

	nuints, err = convStringAToUnintA(sc.Moderators)
	if err != nil {
		return nil, err
	}
	clan.Moderators = *nuints

	nuints, err = convStringAToUnintA(sc.Invited)
	if err != nil {
		return nil, err
	}
	clan.Invited = *nuints

	return &clan, nil
}

func convStringAToUnintA(in []string) (*[]uint64, error) {
	var out []uint64
	var l = len(in)
	if l == 0 {
		return &out, nil
	}

	out = make([]uint64, len(in))
	for i, conv := range in {
		newuint, err := strconv.ParseUint(conv, 10, 64)
		if err != nil {
			return nil, err
		}
		out[i] = newuint
	}

	return &out, nil
}
