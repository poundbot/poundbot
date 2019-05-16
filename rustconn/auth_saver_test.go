package rustconn

import (
	"testing"

	"github.com/poundbot/poundbot/storage/mocks"
	"github.com/poundbot/poundbot/types"
)

func TestAuthSaver_Run(t *testing.T) {
	t.Parallel()

	var mockU *mocks.UsersStore
	var mockDA *mocks.DiscordAuthsStore
	done := make(chan struct{})

	tests := []struct {
		name string
		a    func() *AuthSaver
		with *types.DiscordAuth
		want *types.DiscordAuth
	}{
		{
			name: "With nothing",
			a: func() *AuthSaver {
				return newAuthSaver(mockDA, mockU, make(chan types.DiscordAuth), done)
			},
		},
		{
			name: "With AuthSuccess",

			a: func() *AuthSaver {
				result := types.DiscordAuth{PlayerID: "game:1001"}
				mockU = &mocks.UsersStore{}
				mockU.On("UpsertPlayer", result).Return(nil)

				mockDA = &mocks.DiscordAuthsStore{}
				mockDA.On("Remove", result).Return(nil)

				return newAuthSaver(mockDA, mockU, make(chan types.DiscordAuth), done)
			},
			with: &types.DiscordAuth{PlayerID: "game:1001"},
			want: &types.DiscordAuth{PlayerID: "game:1001"},
		},
	}
	for _, tt := range tests {
		mockU = nil
		mockDA = nil
		for len(done) > 0 {
			<-done
		}

		t.Run(tt.name, func(t *testing.T) {
			var server = tt.a()

			go func() {
				defer func() { done <- struct{}{} }()
				if tt.with != nil {
					server.authSuccess <- *tt.with
				}
			}()

			server.Run()
			if mockU != nil {
				mockU.AssertExpectations(t)
			}
			if mockDA != nil {
				mockDA.AssertExpectations(t)
			}
		})
	}
}
