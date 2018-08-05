package rustconn

import (
	"net/http"
	"reflect"
	"testing"

	"mrpoundsign.com/poundbot/types"
)

// func MockDAL() *mocks.DataStore {
// 	var mockDAL = mocks.DataStore{}
// 	mockDAL.On("Copy").Return(&mockDAL)
// 	mockDAL.On("Close")
// 	return &mockDAL
// }

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
