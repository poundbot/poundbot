package gameapi

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/poundbot/poundbot/types"
)

type raidNotifier interface {
	RaidNotify(types.RaiAlertWithMessageChannel)
}

// A raidStore stores raid information
type raidStore interface {
	GetReady() ([]types.RaidAlert, error)
	IncrementNotifyCount(types.RaidAlert) error
	Remove(types.RaidAlert) error
	messageIDSetter
}

type messageIDSetter interface {
	SetMessageID(types.RaidAlert, string) error
}

// A RaidAlerter sends notifications on raids
type RaidAlerter struct {
	rs        raidStore
	rn        raidNotifier
	SleepTime time.Duration
	done      <-chan struct{}
	miu       func(ra types.RaiAlertWithMessageChannel, is messageIDSetter)
}

// NewRaidAlerter constructs a RaidAlerter
func newRaidAlerter(ral raidStore, rn raidNotifier, done <-chan struct{}) *RaidAlerter {
	return &RaidAlerter{
		rs:        ral,
		rn:        rn,
		done:      done,
		SleepTime: 1 * time.Second,
		miu:       messageIDUpdate,
	}
}

func messageIDUpdate(ra types.RaiAlertWithMessageChannel, is messageIDSetter) {
	raLog := log.WithField("sys", "RALERT")
	newMessageID, ok := <-ra.MessageIDChannel
	if !ok {
		raLog.Trace("messageID channel close")
	}
	raLog.Tracef("New message ID is %s", newMessageID)
	if newMessageID != ra.MessageID {
		err := is.SetMessageID(ra.RaidAlert, newMessageID)
		if err != nil {
			raLog.WithError(err).Error("storage: Could not set message ID")
		}
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
				shouldNotify := true
				raLog.Tracef("Processing alert %s, %d", alert.ID, alert.NotifyCount)

				// Increment notify count should ensure we're the node that should notify for this action.
				if err := r.rs.IncrementNotifyCount(alert); err != nil {
					raLog.WithError(err).Trace("could not increment")
					shouldNotify = false
				}

				if alert.ValidUntil.Before(time.Now()) {
					raLog.Trace("removing")
					if err := r.rs.Remove(alert); err != nil {
						raLog.Trace("coul not remove")
						raLog.WithError(err).Error("storage: Could not remove alert")
						continue
					}
				}

				if shouldNotify {
					message := types.RaiAlertWithMessageChannel{
						RaidAlert:        alert,
						MessageIDChannel: make(chan string),
					}
					log.Trace("notifying")
					r.rn.RaidNotify(message)
					go r.miu(message, r.rs)
				}
			}
		}
	}
}
