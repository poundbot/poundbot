package rustconn

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/storage/mocks"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEntityDeath_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		e      *entityDeath
		method string
		status int
		rBody  string
		ed     *types.EntityDeath
		log    string
	}{
		{
			name:   "POST empty request",
			e:      &entityDeath{},
			method: http.MethodPost,
			status: http.StatusBadRequest,
			log:    "[C] [request-1](5cafadc080e1a9498fea8f03:server1) Invalid JSON: EOF\n",
		},
		{
			name:   "POST entity death",
			e:      &entityDeath{},
			method: http.MethodPost,
			status: http.StatusOK,
			rBody: `
			{
				"Name": "foo",
				"GridPos": "A10",
				"OwnerIDs": ["1", "2", "3"],
				"CreatedAt": "2001-02-03T04:05:06Z"
			}
			`,
			ed: &types.EntityDeath{
				ServerName: "server1",
				Name:       "foo",
				GridPos:    "A10",
				OwnerIDs:   []string{"game:1", "game:2", "game:3"},
				Timestamp:  types.Timestamp{CreatedAt: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)},
			},
		},
		{
			name:   "old API POST empty request",
			e:      &entityDeath{},
			method: http.MethodPost,
			status: http.StatusBadRequest,
			log:    "[C] [request-1](5cafadc080e1a9498fea8f03:server1) Invalid JSON: EOF\n",
		},
		{
			name:   "old API POST entity death",
			e:      &entityDeath{},
			method: http.MethodPost,
			status: http.StatusOK,
			rBody: `
			{
				"Name": "foo",
				"GridPos": "A10",
				"Owners": [1, 2, 3],
				"CreatedAt": "2001-02-03T04:05:06Z"
			}
			`,
			ed: &types.EntityDeath{
				ServerName: "server1",
				Name:       "foo",
				GridPos:    "A10",
				OwnerIDs:   []string{"game:1", "game:2", "game:3"},
				Timestamp:  types.Timestamp{CreatedAt: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC)},
			},
		},
	}

	for _, tt := range tests {
		logBuffer := bytes.NewBuffer([]byte{})
		tt.e.logger = &log.Logger{}
		tt.e.logger.SetOutput(logBuffer)
		tt.e.logger.SetPrefix("[C] ")

		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/entity_death", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}
			var added *types.EntityDeath
			ras := mocks.RaidAlertsStore{}
			tt.e.ras = &ras
			// ras.On("AddInfo", mock.AnythingOfType("types.EntityDeath")).
			ras.On("AddInfo", mock.AnythingOfType("time.Duration"), mock.AnythingOfType("types.EntityDeath")).
				Return(func(t time.Duration, ed types.EntityDeath) error {
					added = &ed
					return nil
				})

			rr := httptest.NewRecorder()

			ctx := context.WithValue(context.Background(), contextKeyServerKey, "bloop")
			ctx = context.WithValue(ctx, contextKeyRequestUUID, "request-1")
			ctx = context.WithValue(ctx, contextKeyGame, "game")
			ctx = context.WithValue(ctx, contextKeyAccount, types.Account{
				ID: bson.ObjectIdHex("5cafadc080e1a9498fea8f03"), //bson.NewObjectId(),
				Servers: []types.Server{
					{ChatChanID: "1234", Key: "bloop", Name: "server1"},
				},
			})

			req = req.WithContext(ctx)

			handler := http.HandlerFunc(tt.e.Handle)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.status, rr.Code)
			assert.Equal(t, tt.ed, added)
			assert.Equal(t, tt.log, logBuffer.String(), "log was incorrect")
		})
	}
}
