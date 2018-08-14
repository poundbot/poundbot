package rustconn

import (
	"testing"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage/mocks"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRaidAlerter_Run(t *testing.T) {
	t.Parallel()

	var mockRA *mocks.RaidAlertsStore

	var rn = types.RaidAlert{
		SteamInfo: types.SteamInfo{
			SteamID: 1234,
		},
	}
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

				return NewRaidAlerter(mockRA, ch, done)
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

				mockRA.On("GetReady", mock.AnythingOfType("*[]types.RaidAlert")).
					Return(func(args *[]types.RaidAlert) error {
						if first {
							first = false
							*args = []types.RaidAlert{rn}
						} else {
							*args = []types.RaidAlert{}
						}

						return nil
					})

				mockRA.On("Remove", rn).Return(nil).Once()

				go func() {
					rnResult = <-ch
					done <- struct{}{}
				}()

				r := NewRaidAlerter(mockRA, ch, done)
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
