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
	IncrementNotifyCount(types.RaidAlert) error
	SetMessageID(types.RaidAlert, string) error
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

			for _, alert := range alerts {
				// Increment notify count should ensure we're the node that should notify for this action.
				if err := r.rs.IncrementNotifyCount(alert); err != nil {
					continue
				}

				if alert.ValidUntil.Before(time.Now()) {
					if err := r.rs.Remove(alert); err != nil {
						raLog.WithError(err).Error("storage: Could not remove alert")
						continue
					}
				}

				r.rn.RaidNotify(alert)
				go func(ra types.RaidAlert) {
					newMessageID := <-ra.MessageIDChannel
					if newMessageID != ra.MessageID {
						err := r.rs.SetMessageID(ra, newMessageID)
						if err != nil {
							raLog.WithError(err).Error("storage: Could not set message ID")
						}
					}
				}(alert)
			}
		}
	}
}
