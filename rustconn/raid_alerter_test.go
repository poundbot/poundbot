package rustconn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"bitbucket.org/mrpoundsign/poundbot/db/mocks"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

func TestRaidAlerter_Run(t *testing.T) {
	var mockRA *mocks.RaidAlertsStore
	var done chan struct{}

	var rn = types.RaidNotification{
		DiscordInfo: types.DiscordInfo{
			DiscordID: "Foo#1234",
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
				done = make(chan struct{})

				mockRA = &mocks.RaidAlertsStore{}

				mockRA.On("GetReady", mock.AnythingOfType("*[]types.RaidNotification")).
					Return(func(args *[]types.RaidNotification) error {
						*args = []types.RaidNotification{}
						go func() { done <- struct{}{} }()

						return nil
					})

				go func() {
					rnResult = <-ch
				}()

				return NewRaidAlerter(mockRA, ch, done)
			},
			want: types.RaidNotification{},
		},
		{
			name: "With RaidAlert",
			r: func() *RaidAlerter {
				ch := make(chan types.RaidNotification)
				done = make(chan struct{})
				var first = true // Track first run of GetReady

				mockRA = &mocks.RaidAlertsStore{}

				mockRA.On("GetReady", mock.AnythingOfType("*[]types.RaidNotification")).
					Return(func(args *[]types.RaidNotification) error {
						if first {
							first = false
							*args = []types.RaidNotification{rn}
						} else {
							*args = []types.RaidNotification{}
							go func() { done <- struct{}{} }()
						}

						return nil
					})

				mockRA.On("Remove", rn).Return(nil).Once()

				go func() {
					rnResult = <-ch
				}()

				return NewRaidAlerter(mockRA, ch, done)
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
