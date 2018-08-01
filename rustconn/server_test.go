package rustconn

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"mrpoundsign.com/poundbot/db"
	"mrpoundsign.com/poundbot/db/mocks"
	"mrpoundsign.com/poundbot/types"
)

func MockDAL() *mocks.DataAccessLayer {
	var mockDAL = mocks.DataAccessLayer{}
	mockDAL.On("Copy").Return(&mockDAL)
	mockDAL.On("Close")
	return &mockDAL
}

func TestNewServer(t *testing.T) {
	type args struct {
		sc   *ServerConfig
		rch  chan types.RaidNotification
		dach chan types.DiscordAuth
		asch chan types.DiscordAuth
		cch  chan string
		coch chan types.ChatMessage
	}
	tests := []struct {
		name string
		args args
		want *Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(tt.args.sc, tt.args.rch, tt.args.dach, tt.args.asch, tt.args.cch, tt.args.coch); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Serve(t *testing.T) {
	tests := []struct {
		name string
		s    *Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Serve()
		})
	}
}

func TestServer_raidAlerter(t *testing.T) {
	var mockDAL *mocks.DataAccessLayer
	var mockRA *mocks.RaidAlertsAccessLayer
	var done chan struct{}

	var rn = types.RaidNotification{
		DiscordInfo: types.DiscordInfo{
			DiscordID: "Foo#1234",
		},
	}
	var rnResult types.RaidNotification

	tests := []struct {
		name string
		s    func() *Server
		want types.RaidNotification
	}{
		{
			name: "With nothing",
			s: func() *Server {
				mockDAL = MockDAL()
				ch := make(chan types.RaidNotification)
				done = make(chan struct{})

				mockDAL.On("RaidAlerts").Return(func() db.RaidAlertsAccessLayer {

					mockRA = &mocks.RaidAlertsAccessLayer{}

					mockRA.On("GetReady", mock.AnythingOfType("*[]types.RaidNotification")).
						Return(func(args *[]types.RaidNotification) error {
							*args = []types.RaidNotification{}
							go func() { done <- struct{}{} }()

							return nil
						})

					return mockRA
				}())

				go func() {
					rnResult = <-ch
				}()

				return &Server{sc: &ServerConfig{Database: mockDAL}, RaidNotify: ch}
			},
			want: types.RaidNotification{},
		},
		{
			name: "With RaidAlert",
			s: func() *Server {
				mockDAL = MockDAL()
				ch := make(chan types.RaidNotification)
				done = make(chan struct{})
				var first = true // Track first run of GetReady

				mockDAL.On("RaidAlerts").Return(func() db.RaidAlertsAccessLayer {

					mockRA = &mocks.RaidAlertsAccessLayer{}

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
					return mockRA
				}())

				go func() {
					rnResult = <-ch
				}()

				return &Server{sc: &ServerConfig{Database: mockDAL}, RaidNotify: ch}
			},
			want: rn,
		},
	}
	for _, tt := range tests {
		// Reset rnTesult
		rnResult = types.RaidNotification{}
		mockDAL = nil
		mockRA = nil

		t.Run(tt.name, func(t *testing.T) {
			tt.s().raidAlerter(done)
			mockDAL.AssertExpectations(t)
			mockRA.AssertExpectations(t)
			assert.Equal(t, tt.want, rnResult, "They should be equal")
		})
	}
}

func TestServer_authHandler(t *testing.T) {
	var mockDAL *mocks.DataAccessLayer
	var mockU *mocks.UsersAccessLayer
	var mockDA *mocks.DiscordAuthsAccessLayer
	var done chan struct{}

	tests := []struct {
		name string
		s    func() *Server
		with *types.DiscordAuth
		want *types.DiscordAuth
	}{
		{
			name: "With nothing",
			s: func() *Server {
				mockDAL = MockDAL()
				go func() { done <- struct{}{} }()
				return &Server{sc: &ServerConfig{Database: mockDAL}, AuthSuccess: make(chan types.DiscordAuth)}
			},
		},
		{
			name: "With AuthSuccess",
			s: func() *Server {
				mockDAL = MockDAL()
				result := types.DiscordAuth{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1001}}}

				mockDAL.On("Users").Return(func() db.UsersAccessLayer {
					mockU = &mocks.UsersAccessLayer{}
					mockU.On("BaseUpsert", result.BaseUser).Return(nil)
					return mockU
				}())

				mockDAL.On("DiscordAuths").Return(func() db.DiscordAuthsAccessLayer {
					mockDA = &mocks.DiscordAuthsAccessLayer{}
					mockDA.On("Remove", result.SteamInfo).Return(nil)
					return mockDA
				}())

				return &Server{sc: &ServerConfig{Database: mockDAL}, AuthSuccess: make(chan types.DiscordAuth)}
			},
			with: &types.DiscordAuth{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1001}}},
			want: &types.DiscordAuth{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1001}}},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {

		mockDAL = nil
		mockU = nil
		mockDA = nil
		done = make(chan struct{})

		t.Run(tt.name, func(t *testing.T) {
			var server = tt.s()

			go func() {
				if tt.with != nil {
					t.Logf("Auth %v", tt.with)
					server.AuthSuccess <- *tt.with
				}
				done <- struct{}{}
			}()

			server.authHandler(done)
			mockDAL.AssertExpectations(t)
			if mockU != nil {
				mockU.AssertExpectations(t)
			}
			if mockDA != nil {
				mockDA.AssertExpectations(t)
			}
		})
	}
}

func TestServer_clansHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.clansHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_clanHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.clanHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_chatHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.chatHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_entityDeathHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.entityDeathHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestServer_discordAuthHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		s    *Server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.discordAuthHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_handleError(t *testing.T) {
	type args struct {
		w         http.ResponseWriter
		restError types.RESTError
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleError(tt.args.w, tt.args.restError)
		})
	}
}
