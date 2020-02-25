// +build integration

package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/poundbot/poundbot/storage/mongodb/mongotest"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	oid := primitive.NewObjectID()

	type args struct {
		alertIn time.Duration
		ed      types.EntityDeath
	}
	tests := []struct {
		name      string
		args      args
		want      types.RaidAlert
		atTimeNew bool
		wantCount int64
		wantErr   bool
	}{
		{
			name: "upsert",
			args: args{alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcd", Name: "thing", GridPos: "D7", OwnerIDs: []string{"1", "2"}}},
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
			args: args{alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcde", Name: "thing", GridPos: "D7", OwnerIDs: []string{"1", "3"}}},
			want: types.RaidAlert{
				GridPositions: []string{"D7"},
				PlayerID:      "3",
				Items:         map[string]int{"thing": 1},
				ServerKey:     "abcde",
			},
			atTimeNew: true,
			wantCount: 2,
		},
		{
			name: "noop",
			args: args{alertIn: time.Hour, ed: types.EntityDeath{ServerKey: "abcde", Name: "thing", GridPos: "D7", OwnerIDs: []string{"5"}}},
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

			coll.C.InsertOne(context.TODO(), types.RaidAlert{
				ID:            oid,
				GridPositions: []string{"D8"},
				PlayerID:      "2",
				Items:         map[string]int{"thing": 2},
				ServerKey:     "abcd",
			})

			usersColl.C.InsertOne(context.TODO(), types.BaseUser{GamesInfo: types.GamesInfo{PlayerIDs: []string{"2"}}})
			usersColl.C.InsertOne(context.TODO(), types.BaseUser{GamesInfo: types.GamesInfo{PlayerIDs: []string{"3"}}})

			if err := raidAlerts.AddInfo(tt.args.alertIn, tt.args.ed); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.AddInfo() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				count, err := coll.C.CountDocuments(context.TODO(), bson.M{})
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.wantCount, count)

				var rn types.RaidAlert
				err = coll.C.FindOne(context.TODO(), bson.M{"playerid": tt.want.PlayerID}).Decode(&rn)
				if err != nil {
					t.Fatal(err)
				}
				alertAt := rn.AlertAt
				rn.AlertAt = time.Time{}

				if tt.atTimeNew {
					assert.NotEqual(t, rn.AlertAt, alertAt)
					// since we don't know the object ID, we empty it.
					assert.NotEqual(t, rn.ID, "")
					rn.ID = primitive.NilObjectID
				} else {
					assert.Equal(t, rn.AlertAt, alertAt)
				}

				assert.Equal(t, tt.want, rn)
			}
		})
	}
}

func TestRaidAlerts_GetReady(t *testing.T) {
	oid := primitive.NewObjectID()
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
					ID:         oid,
					PlayerID:   "1001",
					AlertAt:    time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond),
					ServerName: "",
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
				coll.C.InsertOne(context.TODO(), alert)
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
	oid := primitive.NewObjectID()
	type args struct {
		alert types.RaidAlert
	}
	tests := []struct {
		name      string
		args      args
		alerts    []types.RaidAlert
		wantCount int64
		wantErr   bool
	}{
		{
			name: "one of two",
			args: args{alert: types.RaidAlert{ID: oid}},
			alerts: []types.RaidAlert{
				types.RaidAlert{PlayerID: "1001"},
				types.RaidAlert{ID: oid, PlayerID: "1002"},
			},
			wantCount: 1,
		},
		{
			name: "none",
			args: args{alert: types.RaidAlert{ID: oid, PlayerID: "1003"}},
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
				coll.C.InsertOne(context.TODO(), alert)
			}

			if err := raidAlerts.Remove(tt.args.alert); (err != nil) != tt.wantErr {
				t.Errorf("RaidAlerts.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.CountDocuments(context.TODO(), bson.M{})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
