package types

import (
	"fmt"
	"time"

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
	PlayerID      string
	ServerName    string
	ServerKey     string
	GridPositions []string
	Items         map[string]int
	AlertAt       time.Time
}

func (rn RaidAlert) String() string {
	index := 0
	items := make([]string, len(rn.Items))
	for k, v := range rn.Items {
		items[index] = fmt.Sprintf("%s(%d)", k, v)
		index++
	}

	return messages.RaidAlert(rn.ServerName, rn.GridPositions, items)
}
