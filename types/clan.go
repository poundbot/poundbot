package types

import (
	"strconv"
	"time"
)

type ServerClan struct {
	Tag         string   `json:"tag"`
	Owner       string   `json:"owner"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
	Moderators  []string `json:"moderators"`
	Invited     []string `json:"invited"`
}

type BaseClan struct {
	Tag         string   `bson:"tag"`
	OwnerID     uint64   `bson:"owner_id"`
	Description string   `bson:"description"`
	Members     []uint64 `bson:"members"`
	Moderators  []uint64 `bson:"moderators"`
	Invited     []uint64 `bson:"invited"`
}

type Clan struct {
	BaseClan  `bson:",inline"`
	CreatedAt time.Time `bson:"created_at"`
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
