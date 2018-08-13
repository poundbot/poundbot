// +build integration

package mongodb

import (
	"testing"

	"bitbucket.org/mrpoundsign/poundbot/storage/mongodb/mongotest"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/stretchr/testify/assert"
)

var baseUser = types.BaseUser{
	SteamInfo:   types.SteamInfo{SteamID: 1000},
	DisplayName: "Player 1",
	DiscordInfo: types.DiscordInfo{
		DiscordName: "Da Player 1",
		Snowflake:   "9879438734974398",
	},
}

func NewUsers(t *testing.T) (*Users, *mongotest.Collection) {
	coll, err := mongotest.NewCollection(usersCollection)
	if err != nil {
		t.Fatal(err)
	}
	return &Users{collection: coll.C}, coll
}

func TestUsers_Get(t *testing.T) {
	t.Parallel()
	type args struct {
		steamID uint64
	}
	tests := []struct {
		name    string
		args    args
		want    *types.User
		wantErr bool
	}{
		{
			name: "found",
			args: args{steamID: 1000},
			want: &types.User{BaseUser: baseUser},
		},
		{
			name:    "not found",
			args:    args{steamID: 1001},
			want:    &types.User{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, coll := NewUsers(t)
			defer coll.Close()

			users.collection.Insert(baseUser)

			var got types.User
			err := users.Get(tt.args.steamID, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Users.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, &got) {
				t.Errorf("Users.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUsers_UpsertBase(t *testing.T) {
	t.Parallel()
	type args struct {
		user types.BaseUser
	}
	tests := []struct {
		name      string
		user      types.User
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name:      "insert",
			user:      types.User{BaseUser: types.BaseUser{SteamInfo: types.SteamInfo{SteamID: 1002}}},
			wantCount: 2,
		},
		{
			name:      "upsert",
			user:      types.User{BaseUser: baseUser},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, coll := NewUsers(t)
			defer coll.Close()

			err := users.collection.Insert(baseUser)
			if err != nil {
				t.Fatal(err)
			}

			err = users.UpsertBase(tt.user.BaseUser)
			if err != nil {
				t.Fatal(err)
			}

			count, err := users.collection.Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestUsers_RemoveClan(t *testing.T) {
	t.Parallel()
	type args struct {
		serverKey string
		tag       string
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
			users, coll := NewUsers(t)
			defer coll.Close()
			if err := users.RemoveClan(tt.args.serverKey, tt.args.tag); (err != nil) != tt.wantErr {
				t.Errorf("Users.RemoveClan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsers_RemoveClansNotIn(t *testing.T) {
	t.Parallel()
	type args struct {
		serverKey string
		tags      []string
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
			users, coll := NewUsers(t)
			defer coll.Close()
			if err := users.RemoveClansNotIn(tt.args.serverKey, tt.args.tags); (err != nil) != tt.wantErr {
				t.Errorf("Users.RemoveClansNotIn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUsers_SetClanIn(t *testing.T) {
	t.Parallel()
	type args struct {
		serverKey string
		tag       string
		steamIds  []uint64
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
			users, coll := NewUsers(t)
			defer coll.Close()
			if err := users.SetClanIn(tt.args.serverKey, tt.args.tag, tt.args.steamIds); (err != nil) != tt.wantErr {
				t.Errorf("Users.SetClanIn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}