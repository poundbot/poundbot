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

	var rn = types.RaidNotification{
		SteamInfo: types.SteamInfo{
			SteamID: 1234,
		},
	}
	var rnResult types.RaidNotification

	tests := []struct {
		name string
		r    func() *RaidAlerter
		want types.RaidNotification
	}{
		{
			name: "With nothing",
			r: func() *RaidAlerter {
				ch := make(chan types.RaidNotification)
				done := make(chan struct{})

				mockRA = &mocks.RaidAlertsStore{}

				go func() { done <- struct{}{} }()

				return NewRaidAlerter(mockRA, ch, done)
			},
			want: types.RaidNotification{},
		},
		{
			name: "With RaidAlert",
			r: func() *RaidAlerter {
				ch := make(chan types.RaidNotification)
				first := true // Track first run of GetReady
				done := make(chan struct{})

				mockRA = &mocks.RaidAlertsStore{}

				mockRA.On("GetReady", mock.AnythingOfType("*[]types.RaidNotification")).
					Return(func(args *[]types.RaidNotification) error {
						if first {
							first = false
							*args = []types.RaidNotification{rn}
						} else {
							*args = []types.RaidNotification{}
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
		rnResult = types.RaidNotification{}
		mockRA = nil

		t.Run(tt.name, func(t *testing.T) {
			tt.r().Run()
			mockRA.AssertExpectations(t)
			assert.Equal(t, tt.want, rnResult, "They should be equal")
		})
	}
}
