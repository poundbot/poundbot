package rustconn

import (
	"testing"

	"bitbucket.org/mrpoundsign/poundbot/db/mocks"
	"bitbucket.org/mrpoundsign/poundbot/types"
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
				return NewAuthSaver(mockDA, mockU, make(chan types.DiscordAuth), done)
			},
		},
		{
			name: "With AuthSuccess",
			a: func() *AuthSaver {
				result := types.DiscordAuth{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1001}}}
				mockU = &mocks.UsersStore{}
				mockU.On("UpsertBase", result.BaseUser).Return(nil)

				mockDA = &mocks.DiscordAuthsStore{}
				mockDA.On("Remove", result.SteamInfo).Return(nil)

				return NewAuthSaver(mockDA, mockU, make(chan types.DiscordAuth), done)
			},
			with: &types.DiscordAuth{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1001}}},
			want: &types.DiscordAuth{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1001}}},
		},
		// TODO: Add test cases.
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
					t.Logf("Auth %v", tt.with)
					server.AuthSuccess <- *tt.with
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
