package rustconn

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"context"

	"github.com/stretchr/testify/assert"

	ptime "bitbucket.org/mrpoundsign/poundbot/time"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

type ChatCache struct {
	channel chan types.ChatMessage
}

func (c ChatCache) GetOutChannel(name string) chan types.ChatMessage {
	return c.channel
}

func TestChat_Handle(t *testing.T) {
	ptime.Mock()
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
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "discord",
				Timestamp:   types.Timestamp{CreatedAt: ptime.Clock().Now().UTC()},
			},
			body: "{\"SteamID\":1234,\"ClanTag\":\"FoO\",\"DisplayName\":\"player\",\"Message\":\"hello there!\",\"Source\":\"discord\",\"CreatedAt\":\"1970-01-01T00:00:00Z\"}",
		},
		{
			name:   "chat POST",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusOK,
			rBody: `
			{
				"SteamId":1234,
				"ClanTag":"FoO",
				"DisplayName":"player",
				"Message":"hello there!"
			}
			`,
			rMessage: &types.ChatMessage{
				SteamInfo:   types.SteamInfo{SteamID: 1234},
				ClanTag:     "FoO",
				DisplayName: "player",
				Message:     "hello there!",
				Source:      "rust",
				ChannelID:   "1234",
				Timestamp:   types.Timestamp{CreatedAt: ptime.Clock().Now().UTC()},
			},
		},
		{
			name:   "chat POST bad json",
			method: http.MethodPost,
			s:      &chat{},
			status: http.StatusOK,
			rBody:  "not JSON",
			log:    "[C] [request-1] Invalid JSON: invalid character 'o' in literal null (expecting 'u')\n",
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

			ctx := context.WithValue(context.Background(), contextKeyRequestUUID, "request-1")
			ctx = context.WithValue(ctx, contextKeyServerKey, "bloop")
			ctx = context.WithValue(ctx, contextKeyAccount, types.Account{Servers: []types.Server{
				{ChatChanID: "1234", Key: "bloop"},
			}})

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
