package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bitbucket.org/mrpoundsign/poundbot/storage/mocks"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

func TestChat_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		s      *Chat
		body   string
		rBody  string
		status int
		in     *types.ChatMessage
		out    *types.ChatMessage
	}{
		{
			name:   "chat disabled GET",
			method: http.MethodGet,
			s:      &Chat{d: true},
			status: http.StatusNoContent,
		},
		{
			name:   "chat disabled POST",
			method: http.MethodPost,
			s:      &Chat{d: true},
			status: http.StatusOK,
			rBody:  `{"steam_id": 1234}`,
		},
		{
			name:   "chat enabled GET",
			method: http.MethodGet,
			s:      &Chat{},
			status: http.StatusOK,
			out: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "discord",
				Timestamp:   types.Timestamp{CreatedAt: time.Time{}},
			},
			body: "{\"SteamID\":1234,\"ClanTag\":\"FoO\",\"DisplayName\":\"player\",\"Message\":\"hello there!\",\"Source\":\"discord\",\"CreatedAt\":\"0001-01-01T00:00:00Z\"}",
		},
		{
			name:   "chat enabled POST",
			method: http.MethodPost,
			s:      &Chat{},
			status: http.StatusOK,
			rBody: `
			{
				"SteamId":1234,
				"ClanTag":"FoO",
				"DisplayName":"player",
				"Message":"hello there!"
			}
			`,
			in: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "rust",
				ChannelID:   "1234",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in *types.ChatMessage

			tt.s.in = make(chan types.ChatMessage)

			cs := mocks.ChatsStore{}
			tt.s.cs = &cs

			if tt.out != nil {
				cs.On("GetNext", "bloop", mock.AnythingOfType("*types.ChatMessage")).
					Return(func(serverKey string, m *types.ChatMessage) error {
						tmp := *tt.out
						*m = tmp
						return nil
					})
			}

			req, err := http.NewRequest(tt.method, "/chat", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			var wg sync.WaitGroup

			// Collect any incoming messages
			if tt.in != nil {
				wg.Add(1)
				var thing types.ChatMessage
				go func() {
					defer wg.Done()
					select {
					case thing = <-tt.s.in:
						in = &thing
						break
					}
				}()
			}

			context.Set(req, "serverKey", "bloop")
			context.Set(req, "account", types.Account{BaseAccount: types.BaseAccount{Servers: []types.Server{
				{ChatChanID: "1234", Key: "bloop"},
			}}})
			handler := http.HandlerFunc(tt.s.Handle)
			handler.ServeHTTP(rr, req)

			wg.Wait()

			assert.Equal(t, tt.body, rr.Body.String(), "handler returned bad body")
			assert.Equal(t, tt.status, rr.Code, "handler returned wrong status code")
			assert.Equal(t, tt.in, in, "handler got wrong in message")
		})
	}
}
