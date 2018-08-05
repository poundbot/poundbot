package rustconn

import (
	"time"

	"mrpoundsign.com/poundbot/db"
	"mrpoundsign.com/poundbot/types"
)

// A RaidAlerter sends notifications on raids
type RaidAlerter struct {
	RaidStore  db.RaidAlertsStore
	RaidNotify chan types.RaidNotification
	done       chan struct{}
}

// NewRaidAlerter constructs a RaidAlerter
func NewRaidAlerter(ral db.RaidAlertsStore, rnc chan types.RaidNotification, done chan struct{}) *RaidAlerter {
	return &RaidAlerter{
		RaidStore:  ral,
		RaidNotify: rnc,
		done:       done,
	}
}

// Run checks for raids that need to be alerted and sends them
// out through the RaidNotify channel. It runs in a loop.
func (r *RaidAlerter) Run() {
	for {
		var results []types.RaidNotification
		r.RaidStore.GetReady(&results)

		for _, result := range results {
			r.RaidNotify <- result
			r.RaidStore.Remove(result)
		}

	ExitCheck:
		select {
		case <-r.done:
			return
		default:
			break ExitCheck
		}
		time.Sleep(1)
	}
}
