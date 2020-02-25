package gameapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type chatQueueMock struct {
	message bool
}

type discordMessageHandler struct {
	message *types.ChatMessage
}

func (dmh *discordMessageHandler) SendChatMessage(cm types.ChatMessage) {
	dmh.message = &cm
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.cqs = &chatQueueMock{message: tt.dMessage}

			req, err := http.NewRequest(tt.method, "/chat", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			id, err := primitive.ObjectIDFromHex("5cafadc080e1a9498fea8f03")
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(context.Background(), contextKeyRequestUUID, "request-1")
			ctx = context.WithValue(ctx, contextKeyServerKey, "bloop")
			ctx = context.WithValue(ctx, contextKeyGame, "game")
			ctx = context.WithValue(ctx, contextKeyAccount, types.Account{
				ID: id,
				Servers: []types.AccountServer{
					{
						Key:  "bloop",
						Name: "server-name",
						Channels: []types.AccountServerChannel{
							{ChannelID: "1234", Tags: []string{"chat"}},
						},
					},
				},
			})

			req = req.WithContext(ctx)
			handler := http.HandlerFunc(tt.s.handle)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.body, rr.Body.String(), "handler returned bad body")
			assert.Equal(t, tt.status, rr.Code, "handler returned wrong status code")
			// assert.Equal(t, tt.log, hook.LastEntry().Message, "log was incorrect")
		})
	}
}
