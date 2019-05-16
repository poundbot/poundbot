package rustconn

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
)

// A RaidStore stores raid information
type RaidStore interface {
	GetReady() ([]types.RaidAlert, error)
	Remove(types.RaidAlert) error
}

// A RaidAlerter sends notifications on raids
type RaidAlerter struct {
	RaidStore  RaidStore
	RaidNotify chan<- types.RaidAlert
	SleepTime  time.Duration
	done       <-chan struct{}
}

// NewRaidAlerter constructs a RaidAlerter
func newRaidAlerter(ral RaidStore, rnc chan<- types.RaidAlert, done <-chan struct{}) *RaidAlerter {
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
	raLog := log.WithField("sys", "RALERT")
	raLog.Info("Starting")
	for {
		select {
		case <-r.done:
			raLog.Warn("Shutting down")
			return
		case <-time.After(r.SleepTime):
			alerts, err := r.RaidStore.GetReady()
			if err != nil && err != mgo.ErrNotFound {
				raLog.WithError(err).Error("could not get raid alert")
				continue
			}

			for _, result := range alerts {
				if err := r.RaidStore.Remove(result); err != nil {
					raLog.WithError(err).Error("storage: Could not remove alert")
				}
				r.RaidNotify <- result
			}
		}
	}
}
