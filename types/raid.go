package types

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/messages"
)

type EntityDeath struct {
	ServerName string
	ServerKey  string
	Name       string
	GridPos    string
	OwnerIDs   []string
	Timestamp  `bson:",inline" json:",inline"`
}

type RaidInventory struct {
	Name  string
	Count int
}

type RaidAlert struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	PlayerID      string
	ServerName    string
	ServerKey     string
	GridPositions []string
	Items         map[string]int
	AlertAt       time.Time
	ValidUntil    time.Time
	MessageID     string // The private message ID in discord
	NotifyCount   int
}

type RaiAlertWithMessageChannel struct {
	RaidAlert
	MessageIDChannel chan string
}

func (ra RaidAlert) ItemCount() int {
	count := 0
	for _, v := range ra.Items {
		count += v
	}
	return count
}

func (ra RaidAlert) String() string {
	index := 0
	items := make([]string, len(ra.Items))
	for k, v := range ra.Items {
		items[index] = fmt.Sprintf("%s(%d)", k, v)
		index++
	}

	return messages.RaidAlert(ra.ServerName, ra.GridPositions, items)
}
