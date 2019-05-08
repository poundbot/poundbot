package rustconn

import (
	"testing"
	"time"

	"github.com/poundbot/poundbot/storage/mocks"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
)

func TestRaidAlerter_Run(t *testing.T) {
	t.Parallel()

	var mockRA *mocks.RaidAlertsStore
	var rn = types.RaidAlert{PlayerID: "1234"}
	var rnResult types.RaidAlert

	tests := []struct {
		name string
		r    func() *RaidAlerter
		want types.RaidAlert
	}{
		{
			name: "With nothing",
			r: func() *RaidAlerter {
				ch := make(chan types.RaidAlert)
				done := make(chan struct{})

				mockRA = &mocks.RaidAlertsStore{}

				go func() { done <- struct{}{} }()

				return newRaidAlerter(mockRA, ch, done)
			},
			want: types.RaidAlert{},
		},
		{
			name: "With RaidAlert",
			r: func() *RaidAlerter {
				ch := make(chan types.RaidAlert)
				first := true // Track first run of GetReady
				done := make(chan struct{})

				mockRA = &mocks.RaidAlertsStore{}

				mockRA.On("GetReady").
					Return(func() []types.RaidAlert {
						if first {
							first = false
							return []types.RaidAlert{rn}
						}

						return []types.RaidAlert{}
					}, nil)

				mockRA.On("Remove", rn).Return(nil).Once()

				go func() {
					rnResult = <-ch
					done <- struct{}{}
				}()

				r := newRaidAlerter(mockRA, ch, done)
				r.SleepTime = 1 * time.Microsecond
				return r
			},
			want: rn,
		},
	}
	for _, tt := range tests {

		// Reset rnTesult
		rnResult = types.RaidAlert{}
		mockRA = nil

		t.Run(tt.name, func(t *testing.T) {
			tt.r().Run()
			mockRA.AssertExpectations(t)
			assert.Equal(t, tt.want, rnResult, "They should be equal")
		})
	}
}
