package types

import (
	"time"

	"github.com/poundbot/poundbot/pbclock"
)

var iclock = pbclock.Clock

type Timestamp struct {
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTimestamp creates a Timestamp with the CreatedAt and UpdatedAt set
// to current time UTC
func NewTimestamp() *Timestamp {
	return &Timestamp{CreatedAt: iclock().Now().UTC(), UpdatedAt: iclock().Now().UTC()}
}

// Touch sets UpdatedAt to the current time.
func (t *Timestamp) Touch() {
	t.UpdatedAt = iclock().Now().UTC()
}
