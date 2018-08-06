package rustconn

import (
	"log"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/db"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

// A RaidAlerter sends notifications on raids
type RaidAlerter struct {
	RaidStore  db.RaidAlertsStore
	RaidNotify chan types.RaidNotification
	SleepTime  time.Duration
	done       chan struct{}
}

// NewRaidAlerter constructs a RaidAlerter
func NewRaidAlerter(ral db.RaidAlertsStore, rnc chan types.RaidNotification, done chan struct{}) *RaidAlerter {
	return &RaidAlerter{
		RaidStore:  ral,
		RaidNotify: rnc,
		done:       done,
		SleepTime:  1 * time.Second,
	}
}

// Run checks for raids that need to be alerted and sends them
// out through the RaidNotify channel. It runs in a loop.
func (r *RaidAlerter) Run() {
	for {
		select {
		case <-r.done:
			log.Println(logSymbol + "Shutting down RaidAlerter")
			return
		case <-time.After(r.SleepTime):
			var results []types.RaidNotification
			err := r.RaidStore.GetReady(&results)
			if err != nil {
				log.Println(err)
			}

			for _, result := range results {
				r.RaidNotify <- result
				r.RaidStore.Remove(result)
			}
		}
	}
}
