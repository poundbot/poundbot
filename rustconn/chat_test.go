package rustconn

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/stretchr/testify/assert"
)

type ChatCache struct {
	channel chan types.ChatMessage
}

func (c ChatCache) GetOutChannel(name string) chan types.ChatMessage {
	return c.channel
}

func TestChat_Handle(t *testing.T) {
	pbclock.Mock()
	t.Parallel()

	tests := []struct {
		name     string
		method   string
		s        *chat
		body     string
		rBody    string
		status   int
		rMessage *types.ChatMessage
		dMessage *types.ChatMessage
		log      string
	}{
		{
			name:   "chat GET",
			method: http.MethodGet,
			s:      &chat{},
			status: http.StatusOK,
			dMessage: &types.ChatMessage{
				PlayerID:    "1234",
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
			},
			body: "{\"ClanTag\":\"FoO\",\"DisplayName\":\"player\",\"Message\":\"hello there!\"}",
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
		logBuffer := bytes.NewBuffer([]byte{})
		tt.s.logger = &log.Logger{}
		tt.s.logger.SetOutput(logBuffer)
		tt.s.logger.SetPrefix("[C] ")

		t.Run(tt.name, func(t *testing.T) {
			var messageFromRust *types.ChatMessage

			var wg sync.WaitGroup

			tt.s.sleep = time.Second

			if tt.dMessage != nil {
				tt.s.ccache = ChatCache{channel: make(chan types.ChatMessage, 1)}
				defer close(tt.s.ccache.GetOutChannel("bloop"))
				tt.s.ccache.GetOutChannel("bloop") <- *tt.dMessage
			}

			// Collect any incoming messages
			if tt.rMessage != nil {
				tt.s.in = make(chan types.ChatMessage, 1)
				defer close(tt.s.in)

				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case result := <-tt.s.in:
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
			oid, err := primitive.ObjectIDFromHex("5cafadc080e1a9498fea8f03")
			if err != nil {
				t.Fatal("could not create ObjectID")
			}

			ctx := context.WithValue(context.Background(), contextKeyRequestUUID, "request-1")
			ctx = context.WithValue(ctx, contextKeyServerKey, "bloop")
			ctx = context.WithValue(ctx, contextKeyGame, "game")
			ctx = context.WithValue(ctx, contextKeyAccount, types.Account{
				ID: oid,
				Servers: []types.Server{
					{ChatChanID: "1234", Key: "bloop", Name: "server-name"},
				},
			})

			req = req.WithContext(ctx)
			handler := http.HandlerFunc(tt.s.Handle)
			handler.ServeHTTP(rr, req)

			wg.Wait()

			assert.Equal(t, tt.body, rr.Body.String(), "handler returned bad body")
			assert.Equal(t, tt.status, rr.Code, "handler returned wrong status code")
			assert.Equal(t, tt.rMessage, messageFromRust, "handler got wrong message from rust")
			assert.Equal(t, tt.log, logBuffer.String(), "log was incorrect")
		})
	}
}
