// +build integration

package mongodb

import (
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/poundbot/poundbot/storage/mongodb/mongotest"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
)

func NewRaidAlerts(t *testing.T) (*RaidAlerts, *mongotest.Collection) {
	coll, err := mongotest.NewCollection(raidAlertsCollection)
	if err != nil {
		t.Fatal(err)
	}
	return &RaidAlerts{collection: coll.C}, coll
}

func TestRaidAlerts_AddInfo(t *testing.T) {
	t.Parallel()

	oid := bson.NewObjectId()

	type args struct {
		alertIn    time.Duration
		validUntil time.Duration
		ed         types.EntityDeath
	}
	tests := []struct {
		name          string
		args          args
		want          types.RaidAlert
		atTimeNew     bool
		validUntilNew bool
		wantCount     int
		wantErr       bool
	}{
		{
			name: "upsert",
			args: args{
				alertIn:    time.Hour,
				validUntil: time.Hour,
				ed:         types.EntityDeath{ServerKey: "abcd", Name: "thing", GridPos: "D7", OwnerIDs: []string{"1", "2"}},
			},
			want: types.RaidAlert{
				ID:            oid,
				GridPositions: []string{"D8", "D7"},
				PlayerID:      "2",
				Items:         map[string]int{"thing": 3},
				ServerKey:     "abcd",
			},
			wantCount: 1,
		},
		{
			name: "insert",
			args: args{
				alertIn:    time.Hour,
				validUntil: time.Hour,
				ed:         types.EntityDeath{ServerKey: "abcde", Name: "thing", GridPos: "D7", OwnerIDs: []string{"1", "3"}},
			},
			want: types.RaidAlert{
				GridPositions: []string{"D7"},
				PlayerID:      "3",
				Items:         map[string]int{"thing": 1},
				ServerKey:     "abcde",
			},
			atTimeNew:     true,
			validUntilNew: true,
			wantCount:     2,
		},
		{
			name: "noop",
			args: args{
				alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcde", Name: "thing", GridPos: "D7", OwnerIDs: []string{"5"}},
			},
			want: types.RaidAlert{
				ID:            oid,
				GridPositions: []string{"D8"},
				PlayerID:      "2",
				Items:         map[string]int{"thing": 2},
				ServerKey:     "abcd",
			},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raidAlerts, coll := NewRaidAlerts(t)
			defer coll.Close()

			users, usersColl := NewUsers(t)
			defer usersColl.Close()

			raidAlerts.users = users

			coll.C.Insert(types.RaidAlert{
				ID:            oid,
				GridPositions: []string{"D8"},
				PlayerID:      "2",
				Items:         map[string]int{"thing": 2},
				ServerKey:     "abcd",
			})

			usersColl.C.Insert(types.BaseUser{GamesInfo: types.GamesInfo{PlayerIDs: []string{"2"}}})
			usersColl.C.Insert(types.BaseUser{GamesInfo: types.GamesInfo{PlayerIDs: []string{"3"}}})

			if err := raidAlerts.AddInfo(tt.args.alertIn, tt.args.validUntil, tt.args.ed); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.AddInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				count, err := coll.C.Count()
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.wantCount, count)

				var rn types.RaidAlert
				err = coll.C.Find(bson.M{"playerid": tt.want.PlayerID}).One(&rn)
				if err != nil {
					t.Fatal(err)
				}
				alertAt := rn.AlertAt
				validUntil := rn.ValidUntil

				rn.AlertAt = time.Time{}
				rn.ValidUntil = time.Time{}

				if tt.atTimeNew {
					assert.NotEqual(t, rn.AlertAt, alertAt)
					// since we don't know the object ID, we empty it.
					assert.NotEqual(t, rn.ID, "")
					rn.ID = ""
				} else {
					assert.Equal(t, rn.AlertAt, alertAt)
				}

				if tt.validUntilNew {
					assert.NotEqual(t, rn.ValidUntil, validUntil)
					// since we don't know the object ID, we empty it.
					assert.NotEqual(t, rn.ID, "")
					rn.ID = ""
				} else {
					assert.Equal(t, rn.ValidUntil, validUntil)
				}

				assert.Equal(t, tt.want, rn)
			}
		})
	}
}

func TestRaidAlerts_GetReady(t *testing.T) {
	oid := bson.NewObjectId()
	t.Parallel()

	tests := []struct {
		name    string
		alerts  []types.RaidAlert
		want    []types.RaidAlert
		wantErr bool
	}{
		{
			name: "one of two",
			alerts: []types.RaidAlert{
				types.RaidAlert{ID: oid, PlayerID: "1001", AlertAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC)},
				types.RaidAlert{AlertAt: time.Now().UTC().Add(time.Hour)},
			},
			want: []types.RaidAlert{
				types.RaidAlert{
					ID:            oid,
					PlayerID:      "1001",
					AlertAt:       time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond),
					ServerName:    "",
					GridPositions: []string{},
					Items:         map[string]int{},
				},
			},
		},
		{
			name: "none",
			alerts: []types.RaidAlert{
				types.RaidAlert{AlertAt: time.Now().UTC().Add(time.Hour)},
				types.RaidAlert{AlertAt: time.Now().UTC().Add(time.Hour)},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raidAlerts, coll := NewRaidAlerts(t)
			defer coll.Close()

			for _, alert := range tt.alerts {
				coll.C.Insert(alert)
			}

			got, err := raidAlerts.GetReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.GetReady() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRaidAlerts_Remove(t *testing.T) {
	// t.Parallel()

	oid := bson.NewObjectId()
	tests := []struct {
		name      string
		alert     types.RaidAlert
		alerts    []types.RaidAlert
		wantCount int
		wantErr   bool
	}{
		{
			name:  "one of two",
			alert: types.RaidAlert{ID: oid},
			alerts: []types.RaidAlert{
				types.RaidAlert{PlayerID: "1001"},
				types.RaidAlert{ID: oid, PlayerID: "1002"},
			},
			wantCount: 1,
		},
		{
			name:  "none",
			alert: types.RaidAlert{ID: oid, PlayerID: "1003"},
			alerts: []types.RaidAlert{
				types.RaidAlert{PlayerID: "1001"},
				types.RaidAlert{PlayerID: "1002"},
			},
			wantCount: 2,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raidAlerts, coll := NewRaidAlerts(t)
			defer coll.Close()

			for _, alert := range tt.alerts {
				err := coll.C.Insert(alert)
				if err != nil {
					t.Errorf("RaidAlerts.Remove() insert error = %v", err)
				}
			}

			if err := raidAlerts.Remove(tt.alert); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.Count()
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
