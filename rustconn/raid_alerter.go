package rustconn

import (
	"time"

	"github.com/poundbot/poundbot/types"
)

const raLogPrefix = "[RAIDALERT]"

// A RaidStore stores raid information
type RaidStore interface {
	GetReady() ([]types.RaidAlert, error)
	Remove(types.RaidAlert) error
}

// A RaidAlerter sends notifications on raids
type RaidAlerter struct {
	RaidStore  RaidStore
	RaidNotify chan types.RaidAlert
	SleepTime  time.Duration
	done       chan struct{}
}

// NewRaidAlerter constructs a RaidAlerter
func newRaidAlerter(ral RaidStore, rnc chan types.RaidAlert, done chan struct{}) *RaidAlerter {
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
	log.Println(raLogPrefix + " Starting RaidAlerter")
	for {
		select {
		case <-r.done:
			log.Println(raLogPrefix + "[WARN] Shutting down RaidAlerter")
			return
		case <-time.After(r.SleepTime):
			alerts, err := r.RaidStore.GetReady()
			if err != nil && err.Error() != "not found" {
				log.Printf("[ERROR] Get Raid Alerts Error: %s\n", err)
				continue
			}

			for _, result := range alerts {
				r.RaidNotify <- result
				r.RaidStore.Remove(result)
			}
		}
	}
}
