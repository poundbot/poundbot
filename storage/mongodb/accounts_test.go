// +build integration

package mongodb

import (
	"testing"
	"time"

	"bitbucket.org/mrpoundsign/poundbot/storage/mongodb/mongotest"
	"bitbucket.org/mrpoundsign/poundbot/types"
	"github.com/globalsign/mgo/bson"
	"github.com/stretchr/testify/assert"
)

var baseAccount = types.Account{
	BaseAccount: types.BaseAccount{GuildSnowflake: "snowflake"},
	Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
	Servers:     []types.Server{types.Server{Key: "floop2"}},
}

func NewAccounts(t *testing.T) (*Accounts, *mongotest.Collection) {
	coll, err := mongotest.NewCollection(accountsCollection)
	if err != nil {
		t.Fatal(err)
	}
	return &Accounts{collection: coll.C}, coll
}

func TestAccounts_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    []types.Account
		wantErr bool
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name: "some",
			want: []types.Account{
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "floop", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "floop3", Clans: []types.Clan{}}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			for _, account := range tt.want {
				coll.C.Insert(account)
			}

			var res []types.Account

			if err := accounts.All(&res); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.GetByDiscordGuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, res, "Did not get acounts.")
		})
	}
}

func TestAccounts_GetByDiscordGuild(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		want    types.Account
		wantErr bool
	}{
		{
			name: "exists",
			want: types.Account{
				Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				BaseAccount: types.BaseAccount{GuildSnowflake: "found"},
				Servers:     []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{}}},
			},
			key: "found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			make := []types.Account{
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "lost"},
					Servers:     []types.Server{types.Server{Key: "floop", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "found"},
					Servers:     []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "lost2"},
					Servers:     []types.Server{types.Server{Key: "floop3", Clans: []types.Clan{}}},
				},
			}

			for _, account := range make {
				coll.C.Insert(account)
			}

			got, err := accounts.GetByDiscordGuild(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accounts.GetByDiscordGuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "Account is not what we expected")
		})
	}
}

func TestAccounts_GetByServerKey(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    types.Account
		wantErr bool
	}{
		{
			name: "result",
			args: args{key: "floop2"},
			want: types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			make := []types.Account{
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "floop", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "floop3", Clans: []types.Clan{}}},
				},
			}

			for _, account := range make {
				coll.C.Insert(account)
			}

			got, err := accounts.GetByServerKey(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accounts.GetByDiscordGuild() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.want, got, "Account is not what we expected")
		})
	}
}

func TestAccounts_UpsertBase(t *testing.T) {
	t.Parallel()

	var baseAccount = types.Account{
		Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
		BaseAccount: types.BaseAccount{GuildSnowflake: "yarp1"},
	}

	tests := []struct {
		name      string
		account   types.Account
		wantCount int
		wantErr   bool
	}{
		{
			name:      "insert",
			account:   types.Account{BaseAccount: types.BaseAccount{GuildSnowflake: "yuss"}},
			wantCount: 2,
		},
		{
			name:      "upsert",
			account:   types.Account{BaseAccount: types.BaseAccount{GuildSnowflake: "yarp1"}},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			err := coll.C.Insert(baseAccount)
			if err != nil {
				t.Fatal(err)
			}

			if err = accounts.UpsertBase(tt.account.BaseAccount); (err != nil) != tt.wantErr {
				t.Fatalf("Accounts.UpsertBase() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestAccounts_Remove(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}
	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name:      "not found",
			args:      args{"yuss"},
			wantErr:   true,
			wantCount: 3,
		},
		{
			name:      "upsert",
			args:      args{"floop2"},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			make := []types.Account{
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "floop1"},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "floop2"},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "floop3"},
				},
			}

			for _, account := range make {
				err := coll.C.Insert(account)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := accounts.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.Count()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count, "Count is wrong")
		})
	}
}

func TestAccounts_AddClan(t *testing.T) {
	t.Parallel()

	type args struct {
		key  string
		clan types.Clan
	}
	tests := []struct {
		name    string
		args    args
		want    types.Account
		wantErr bool
	}{
		{
			name: "result",
			want: types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers: []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{
					types.Clan{Tag: "bloops", Members: []uint64{}, Moderators: []uint64{}, Invited: []uint64{}},
					types.Clan{Tag: "bloops2", Members: []uint64{}, Moderators: []uint64{}, Invited: []uint64{}},
				}}},
			},
			args: args{key: "floop2", clan: types.Clan{Tag: "bloops2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{types.Clan{Tag: "bloops"}}}},
			})

			if err := accounts.AddClan(tt.args.key, tt.args.clan); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.SetClans() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account types.Account
			coll.C.Find(bson.M{}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_RemoveClan(t *testing.T) {
	t.Parallel()

	type args struct {
		key     string
		clanTag string
	}
	tests := []struct {
		name    string
		want    types.Account
		args    args
		wantErr bool
	}{
		{
			name: "result",
			want: types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{types.Clan{Tag: "bloops2", Members: []uint64{}, Moderators: []uint64{}, Invited: []uint64{}}}}},
			},
			args: args{key: "floop2", clanTag: "bloops"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{types.Clan{Tag: "bloops"}, types.Clan{Tag: "bloops2"}}}},
			})

			if err := accounts.RemoveClan(tt.args.key, tt.args.clanTag); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.SetClans() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account types.Account
			coll.C.Find(bson.M{}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_SetClans(t *testing.T) {
	t.Parallel()

	type args struct {
		key   string
		clans []types.Clan
	}
	tests := []struct {
		name    string
		args    args
		want    types.Account
		wantErr bool
	}{
		{
			name: "result",
			want: types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "floop2", Clans: []types.Clan{types.Clan{Tag: "foo", Members: []uint64{}, Moderators: []uint64{}, Invited: []uint64{}}}}},
			},
			args: args{key: "floop2", clans: []types.Clan{types.Clan{Tag: "foo"}}},
		},
		{
			name:    "not found",
			args:    args{key: "floop", clans: []types.Clan{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(types.Account{
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "floop2"}},
			})

			if err := accounts.SetClans(tt.args.key, tt.args.clans); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.SetClans() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account types.Account
			coll.C.Find(bson.M{serverKeyField: tt.args.key}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_AddServer(t *testing.T) {
	type args struct {
		snowflake string
		server    types.Server
	}
	tests := []struct {
		name    string
		args    args
		want    types.Account
		wantErr bool
	}{
		{
			name: "add",
			args: args{server: types.Server{Key: "floop"}, snowflake: "snowflake"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.Insert(baseAccount)

			if err := accounts.AddServer(tt.args.snowflake, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}
			var account types.Account
			coll.C.Find(bson.M{serverKeyField: tt.args.snowflake}).One(&account)
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_RemoveServer(t *testing.T) {
	type args struct {
		snowflake string
		serverKey string
	}
	tests := []struct {
		name    string
		s       Accounts
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.RemoveServer(tt.args.snowflake, tt.args.serverKey); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.RemoveServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccounts_UpdateServer(t *testing.T) {
	type args struct {
		snowflake string
		oldKey    string
		server    types.Server
	}
	tests := []struct {
		name    string
		s       Accounts
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UpdateServer(tt.args.snowflake, tt.args.oldKey, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.UpdateServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccounts_RemoveNotInDiscordGuildList(t *testing.T) {
	type args struct {
		guildIDs []string
	}
	tests := []struct {
		name    string
		s       Accounts
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.RemoveNotInDiscordGuildList(tt.args.guildIDs); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.RemoveNotInDiscordGuildList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
