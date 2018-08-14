package rustconn

import (
	"log"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/types"
)

const raLogSymbol = "💬 "

// A RaidStore stores raid information
type RaidStore interface {
	GetReady(*[]types.RaidAlert) error
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
func NewRaidAlerter(ral RaidStore, rnc chan types.RaidAlert, done chan struct{}) *RaidAlerter {
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
	log.Println(raLogSymbol + "🛫 Starting RaidAlerter")
	for {
		select {
		case <-r.done:
			log.Println(raLogSymbol + "🛑 Shutting down RaidAlerter")
			return
		case <-time.After(r.SleepTime):
			var results []types.RaidAlert
			err := r.RaidStore.GetReady(&results)
			if err != nil && err.Error() != "not found" {
				log.Printf("Get Raid Alerts Error: %s\n", err)
				continue
			}

			// fmt.Println(len(results))

			for _, result := range results {
				r.RaidNotify <- result
				r.RaidStore.Remove(result)
			}
		}
	}
}
