package gameapi

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
)

type raidNotifier interface {
	RaidNotify(types.RaidAlert)
}

// A raidStore stores raid information
type raidStore interface {
	GetReady() ([]types.RaidAlert, error)
	Remove(types.RaidAlert) error
}

// A RaidAlerter sends notifications on raids
type RaidAlerter struct {
	rs        raidStore
	rn        raidNotifier
	SleepTime time.Duration
	done      <-chan struct{}
}

// NewRaidAlerter constructs a RaidAlerter
func newRaidAlerter(ral raidStore, rn raidNotifier, done <-chan struct{}) *RaidAlerter {
	return &RaidAlerter{
		rs:        ral,
		rn:        rn,
		done:      done,
		SleepTime: 1 * time.Second,
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
			alerts, err := r.rs.GetReady()
			if err != nil && err != mgo.ErrNotFound {
				raLog.WithError(err).Error("could not get raid alert")
				continue
			}

			for _, result := range alerts {
				// We only want to send the raid notification when we are the instance
				// that can remove it. This is a "simple" way of allowing multiple
				// instances to run against the same DB.
				if err := r.rs.Remove(result); err != nil {
					raLog.WithError(err).Error("storage: Could not remove alert")
					continue
				}
				r.rn.RaidNotify(result)
			}
		}
	}
}
