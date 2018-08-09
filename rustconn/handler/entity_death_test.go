package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"bitbucket.org/mrpoundsign/poundbot/db/mocks"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/stretchr/testify/assert"
)

func TestEntityDeath_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		e      *EntityDeath
		method string
		status int
		rBody  string
		ed     *types.EntityDeath
	}{
		{
			name:   "empty request",
			e:      &EntityDeath{},
			method: http.MethodPost,
			status: http.StatusBadRequest,
		},
		{
			name:   "entity death",
			e:      &EntityDeath{},
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
				Name:      "foo",
				GridPos:   "A10",
				Owners:    []uint64{1, 2, 3},
				CreatedAt: time.Date(2001, 2, 3, 4, 5, 6, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/entity_death", strings.NewReader(tt.rBody))
			if err != nil {
				t.Fatal(err)
			}
			var added *types.EntityDeath
			ras := mocks.RaidAlertsStore{}
			tt.e.ras = &ras
			ras.On("AddInfo", mock.AnythingOfType("types.EntityDeath")).
				Return(func(ed types.EntityDeath) error {
					added = &ed
					return nil
				})

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(tt.e.Handle)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.status, rr.Code)
			assert.Equal(t, tt.ed, added)
		})
	}
}