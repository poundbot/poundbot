package gameapi

import (
	"testing"
	"time"

	"github.com/poundbot/poundbot/storage/mocks"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
)

type raidHandler struct {
	RaidAlert *types.RaidAlert
}

func (rh *raidHandler) RaidNotify(ra types.RaiAlertWithMessageChannel) {
	rh.RaidAlert = &ra.RaidAlert
}

func TestRaidAlerter_Run(t *testing.T) {
	t.Parallel()

	miu := func(ra types.RaiAlertWithMessageChannel, is messageIDSetter) {}

	var ra = types.RaidAlert{PlayerID: "1234"}

	tests := []struct {
		name       string
		raidAlerts []types.RaidAlert
		want       *types.RaidAlert
	}{
		{
			name: "With nothing",
		},
		{
			name:       "With RaidAlert",
			raidAlerts: []types.RaidAlert{ra},
			want:       &ra,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// var hit bool
			done := make(chan struct{}, 1)

			mockRH := &raidHandler{}

			mockRA := mocks.RaidAlertsStore{}

			mockRA.On("GetReady").
				Return(func() []types.RaidAlert {
					done <- struct{}{}
					return tt.raidAlerts
				}, nil)

			if len(tt.raidAlerts) != 0 {
				mockRA.On("IncrementNotifyCount", ra).Return(nil).Once()
				mockRA.On("Remove", ra).Return(nil).Once()
			}

			raidAlerter := newRaidAlerter(&mockRA, mockRH, done)
			raidAlerter.SleepTime = 1 * time.Microsecond
			raidAlerter.miu = miu
			raidAlerter.Run()
			mockRA.AssertExpectations(t)
			assert.EqualValues(t, tt.want, mockRH.RaidAlert, "They should be equal")
		})
	}
}
