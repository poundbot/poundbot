package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	ChatSourceRust    = "rust"
	ChatSourceDiscord = "discord"
)

type Ack func(bool)

type ChatMessage struct {
	SteamInfo   `bson:",inline"`
	ClanTag     string
	DisplayName string
	Message     string
	Source      string
}

type EntityDeath struct {
	Name      string
	GridPos   string
	Owners    []uint64
	CreatedAt time.Time
}

type MongoID struct {
	ID bson.ObjectId `bson:"_id,omitempty"`
}

type DiscordInfo struct {
	DiscordID string `bson:"discord_id"`
}

type SteamInfo struct {
	SteamID uint64 `bson:"steam_id"`
}

type BaseUser struct {
	SteamInfo   `bson:",inline"`
	DiscordInfo `bson:",inline"`
	DisplayName string `bson:"display_name"`
	ClanTag     string `bson:"clan_tag"`
}

type User struct {
	MongoID   `bson:",inline"`
	BaseUser  `bson:",inline"`
	CreatedAt time.Time `bson:"created_at"`
}

type ServerClan struct {
	Tag         string   `json:"tag"`
	Owner       string   `json:"owner"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
	Moderators  []string `json:"moderators"`
	Invited     []string `json:"invited"`
}

type ClanBase struct {
	Tag         string   `bson:"tag"`
	OwnerID     uint64   `bson:"owner_id"`
	Description string   `bson:"description"`
	Members     []uint64 `bson:"members"`
	Moderators  []uint64 `bson:"moderators"`
	Invited     []uint64 `bson:"invited"`
}

type Clan struct {
	MongoID   `bson:",inline"`
	ClanBase  `bson:",inline"`
	CreatedAt time.Time `bson:"created_at"`
}

type DiscordAuth struct {
	MongoID    `bson:",inline"`
	BaseUser   `bson:",inline"`
	Pin        int
	SentToUser bool
	Ack        Ack `bson:"-"`
}

type RaidNotification struct {
	MongoID       `bson:",inline"`
	DiscordInfo   `bson:",inline"`
	GridPositions []string       `bson:"grid_positions"`
	Items         map[string]int `bson:"items"`
	AlertAt       time.Time      `bson:"alert_at"`
}

type RaidInventory struct {
	Name  string
	Count int
}

type RESTError struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

func (rn RaidNotification) String() string {
	index := 0
	items := make([]string, len(rn.Items))
	for k, v := range rn.Items {
		items[index] = fmt.Sprintf("%s(%d)", k, v)
		index++
	}

	return fmt.Sprintf(`
	RAID ALERT! You are being raided!
	
	Locations: 
	  %s
	
	Destroyed:
	  %s
	`, strings.Join(rn.GridPositions, ", "), strings.Join(items, ", "))
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
	var out = make([]uint64, len(in))
	for i, conv := range in {
		newuint, err := strconv.ParseUint(conv, 10, 64)
		if err != nil {
			return nil, err
		}
		out[i] = newuint
	}

	return out, nil
}
