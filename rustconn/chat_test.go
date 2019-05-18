package rustconn

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"

	"context"

	"github.com/stretchr/testify/assert"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"
)

type chatQueueMock struct {
	message bool
}

func (cqm chatQueueMock) GetGameServerMessage(sk, tag string, to time.Duration) (types.ChatMessage, bool) {
	if !cqm.message {
		return types.ChatMessage{}, false
	}
	cm := types.ChatMessage{
		PlayerID:    "1234",
		ClanTag:     "FoO",
		DisplayName: "player",
		Message:     "hello there!",
	}
	return cm, true
}

func TestChat_Handle(t *testing.T) {
	pbclock.Mock()
	t.Parallel()

	tests := []struct {
		name     string
		s        *chat
		method   string             // http method
		body     string             // response body
		status   int                // response status
		dMessage bool               // true if there is a discord message in queue
		rBody    string             // request body
		rMessage *types.ChatMessage // message from Rust
		log      string
	}{
		{
			name:     "chat GET",
			method:   http.MethodGet,
			s:        &chat{},
			status:   http.StatusOK,
			dMessage: true,
			body:     "{\"ClanTag\":\"FoO\",\"DisplayName\":\"player\",\"Message\":\"hello there!\"}",
		},
		{
			name:   "chat GET no message",
			method: http.MethodGet,
			s:      &chat{},
			status: http.StatusNoContent,
			body:   "",
		},
		{
			name:   "chat POST",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusOK,
			rBody: `
			{
				"PlayerID":"1234",
				"ClanTag":"FoO",
				"DisplayName":"player",
				"Message":"hello there!"
			}
			`,
			rMessage: &types.ChatMessage{
				PlayerID:    "game:1234",
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				ChannelID:   "1234",
			},
		},
		{
			name:   "old chat POST",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusOK,
			rBody: `
			{
				"SteamID":1234,
				"ClanTag":"FoO",
				"DisplayName":"player",
				"Message":"hello there!"
			}
			`,
			rMessage: &types.ChatMessage{
				PlayerID:    "game:1234",
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				ChannelID:   "1234",
			},
		},
		{
			name:   "chat POST bad json",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusBadRequest,
			body:   "{\"StatusCode\":400,\"Error\":\"Invalid request\"}\n",
			rBody:  "not JSON",
			log:    "[C] [request-1](5cafadc080e1a9498fea8f03:server-name) Invalid JSON: invalid character 'o' in literal null (expecting 'u')\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var messageFromRust *types.ChatMessage

			var wg sync.WaitGroup

			tt.s.cqs = &chatQueueMock{message: tt.dMessage}

			// Collect any incoming messages
			if tt.rMessage != nil {
				in := make(chan types.ChatMessage, 1)
				defer close(in)
				tt.s.in = in

				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case result := <-in:
							messageFromRust = &result
							return
						case <-time.After(10 * time.Second):
							return
						}
					}
				}()
			}

			req, err := http.NewRequest(tt.method, "/chat", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			ctx := context.WithValue(context.Background(), contextKeyRequestUUID, "request-1")
			ctx = context.WithValue(ctx, contextKeyServerKey, "bloop")
			ctx = context.WithValue(ctx, contextKeyGame, "game")
			ctx = context.WithValue(ctx, contextKeyAccount, types.Account{
				ID: bson.ObjectIdHex("5cafadc080e1a9498fea8f03"),
				Servers: []types.Server{
					{
						Key:  "bloop",
						Name: "server-name",
						Channels: []types.ServerChannel{
							{ChannelID: "1234", Tags: []string{"chat"}},
						},
					},
				},
			})

			req = req.WithContext(ctx)
			handler := http.HandlerFunc(tt.s.Handle)
			handler.ServeHTTP(rr, req)

			wg.Wait()

			assert.Equal(t, tt.body, rr.Body.String(), "handler returned bad body")
			assert.Equal(t, tt.status, rr.Code, "handler returned wrong status code")
			assert.Equal(t, tt.rMessage, messageFromRust, "handler got wrong message from rust")
			// assert.Equal(t, tt.log, hook.LastEntry().Message, "log was incorrect")
		})
	}
}
