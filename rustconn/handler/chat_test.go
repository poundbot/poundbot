package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"bitbucket.org/mrpoundsign/poundbot/db/mocks"
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
		in     string
		logged *types.ChatMessage
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
			},
			body: "{\"SteamID\":1234,\"ClanTag\":\"FoO\",\"DisplayName\":\"player\",\"Message\":\"hello there!\",\"Source\":\"discord\"}",
			logged: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "discord",
			},
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
			logged: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "rust",
			},
			in: "☢️ **[FoO] player**: hello there!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in string

			tt.s.in = make(chan string)
			tt.s.out = make(chan types.ChatMessage)

			cs := mocks.ChatsStore{}
			tt.s.cs = &cs
			var log *types.ChatMessage

			cs.On("Log", mock.AnythingOfType("types.ChatMessage")).
				Return(func(m types.ChatMessage) error {
					log = &m
					return nil
				})

			req, err := http.NewRequest(tt.method, "/chat", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			var wg sync.WaitGroup

			// Collect any incoming messages
			if tt.in != "" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case in = <-tt.s.in:
						break
					}
				}()
			}

			if tt.out != nil {
				tt.s.sleep = 10 * time.Minute
				wg.Add(1)
				go func() {
					defer wg.Done()
					tt.s.out <- *tt.out
				}()
			}

			handler := http.HandlerFunc(tt.s.Handle)
			handler.ServeHTTP(rr, req)

			wg.Wait()

			assert.Equal(t, tt.body, rr.Body.String(), "handler returned bad body")
			assert.Equal(t, tt.status, rr.Code, "handler returned wrong status code")
			assert.Equal(t, tt.in, in, "handler got wrong in message")
			assert.Equal(t, tt.logged, log, "Chat log should be equal")
		})
	}
}
