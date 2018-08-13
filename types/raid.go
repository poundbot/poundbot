package types

import (
	"fmt"
	"strings"
	"time"
)

type EntityDeath struct {
	ServerKey string
	Name      string
	GridPos   string
	Owners    []uint64
	Timestamp `bson:",inline" json:",inline"`
}

type RaidInventory struct {
	Name  string
	Count int
}

type RaidNotification struct {
	SteamInfo     `bson:",inline"`
	ServerName    string
	GridPositions []string
	Items         map[string]int
	AlertAt       time.Time
}

func (rn RaidNotification) String() string {
	index := 0
	items := make([]string, len(rn.Items))
	for k, v := range rn.Items {
		items[index] = fmt.Sprintf("%s(%d)", k, v)
		index++
	}

	return fmt.Sprintf(`
	%s RAID ALERT! You are being raided!
	
	Locations: 
	  %s
	
	Destroyed:
	  %s
	`, rn.ServerName, strings.Join(rn.GridPositions, ", "), strings.Join(items, ", "))
}
