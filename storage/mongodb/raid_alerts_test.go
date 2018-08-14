// +build integration

package mongodb

import (
	"testing"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage/mongodb/mongotest"
	"bitbucket.org/mrpoundsign/poundbot/types"
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

	type args struct {
		alertIn time.Duration
		ed      types.EntityDeath
	}
	tests := []struct {
		name      string
		args      args
		want      types.RaidNotification
		atTimeNew bool
		wantCount int
		wantErr   bool
	}{
		{
			name: "upsert",
			args: args{alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcd", Name: "thing", GridPos: "D7", Owners: []uint64{1, 2}}},
			want: types.RaidNotification{
				GridPositions: []string{"D8", "D7"},
				SteamInfo:     types.SteamInfo{SteamID: 2},
				Items:         map[string]int{"thing": 3},
			},
			wantCount: 1,
		},
		{
			name: "insert",
			args: args{alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcde", Name: "thing", GridPos: "D7", Owners: []uint64{1, 3}}},
			want: types.RaidNotification{
				GridPositions: []string{"D7"},
				SteamInfo:     types.SteamInfo{SteamID: 3},
				Items:         map[string]int{"thing": 1},
			},
			atTimeNew: true,
			wantCount: 2,
		},
		{
			name: "noop",
			args: args{alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcde", Name: "thing", GridPos: "D7", Owners: []uint64{5}}},
			want: types.RaidNotification{
				GridPositions: []string{"D8"},
				SteamInfo:     types.SteamInfo{SteamID: 2},
				Items:         map[string]int{"thing": 2},
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

			coll.C.Insert(types.RaidNotification{
				GridPositions: []string{"D8"},
				SteamInfo:     types.SteamInfo{SteamID: 2},
				Items:         map[string]int{"thing": 2},
			})

			usersColl.C.Insert(types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 2}})
			usersColl.C.Insert(types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 3}})

			if err := raidAlerts.AddInfo(tt.args.alertIn, tt.args.ed); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.AddInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				count, err := coll.C.Count()
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.wantCount, count)

				var rn types.RaidNotification
				err = coll.C.Find(tt.want.SteamInfo).One(&rn)
				if err != nil {
					t.Fatal(err)
				}
				alertAt := rn.AlertAt
				rn.AlertAt = time.Time{}
				assert.Equal(t, tt.want, rn)
				if tt.atTimeNew {
					assert.NotEqual(t, rn.AlertAt, alertAt)
				} else {
					assert.Equal(t, rn.AlertAt, alertAt)
				}
			}
		})
	}
}

func TestRaidAlerts_GetReady(t *testing.T) {
	type args struct {
		alerts *[]types.RaidNotification
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raidAlerts, coll := NewRaidAlerts(t)
			defer coll.Close()

			if err := raidAlerts.GetReady(tt.args.alerts); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.GetReady() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRaidAlerts_Remove(t *testing.T) {
	type args struct {
		alert types.RaidNotification
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raidAlerts, coll := NewRaidAlerts(t)
			defer coll.Close()

			if err := raidAlerts.Remove(tt.args.alert); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
