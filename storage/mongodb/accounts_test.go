// +build integration

package mongodb

import (
	"testing"
	"time"
	"context"

	"github.com/poundbot/poundbot/pbclock"
	"github.com/poundbot/poundbot/storage/mongodb/mongotest"
	"github.com/poundbot/poundbot/types"
	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var baseAccount = types.Account{
	BaseAccount: types.BaseAccount{GuildSnowflake: "snowflake"},
	Timestamp:   types.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
	Servers:     []types.Server{types.Server{Name: "base", Key: "serverkey"}},
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
					ID:        primitive.NewObjectID(),
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "guildsnowflake", Clans: []types.Clan{}}},
				},
				types.Account{
					ID:        primitive.NewObjectID(),
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{}}},
				},
				types.Account{
					ID:        primitive.NewObjectID(),
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "guildsnowflake3", Clans: []types.Clan{}}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			for _, account := range tt.want {
				coll.C.InsertOne(context.Background(), account)
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

	id := primitive.NewObjectID()

	tests := []struct {
		name    string
		key     string
		want    types.Account
		wantErr bool
	}{
		{
			name: "exists",
			want: types.Account{
				ID:          id,
				Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				BaseAccount: types.BaseAccount{GuildSnowflake: "found"},
				Servers:     []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{}}},
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
					Servers:     []types.Server{types.Server{Key: "guildsnowflake", Clans: []types.Clan{}}},
				},
				types.Account{
					ID:          id,
					Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "found"},
					Servers:     []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "lost2"},
					Servers:     []types.Server{types.Server{Key: "guildsnowflake3", Clans: []types.Clan{}}},
				},
			}

			for _, account := range make {
				coll.C.InsertOne(context.Background(),account)
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

	id := primitive.NewObjectID()

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
			args: args{key: "guildsnowflake2"},
			want: types.Account{
				ID:        id,
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			docs := []types.Account{
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "guildsnowflake", Clans: []types.Clan{}}},
				},
				types.Account{
					ID:        id,
					Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{}}},
				},
				types.Account{
					Timestamp: types.Timestamp{CreatedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					Servers:   []types.Server{types.Server{Key: "guildsnowflake3", Clans: []types.Clan{}}},
				},
			}

			for _, account := range docs {
				coll.C.InsertOne(context.Background(), account)
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
		wantCount int64
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

			_, err := coll.C.InsertOne(context.Background(), baseAccount)
			if err != nil {
				t.Fatal(err)
			}

			if err = accounts.UpsertBase(tt.account.BaseAccount); (err != nil) != tt.wantErr {
				t.Fatalf("Accounts.UpsertBase() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.CountDocuments(context.Background(), bson.M{})
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
		wantCount int64
		wantErr   bool
	}{
		{
			name:      "not found",
			args:      args{"nonexistant"},
			wantErr:   true,
			wantCount: 3,
		},
		{
			name:      "upsert",
			args:      args{"guildsnowflake2"},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			make := []types.Account{
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake1"},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake2"},
				},
				types.Account{
					Timestamp:   types.Timestamp{CreatedAt: iclock().Now().UTC().Add(-2 * time.Hour).Truncate(time.Millisecond)},
					BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake3"},
				},
			}

			for _, account := range make {
				_, err := coll.C.InsertOne(context.Background(), account)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err := accounts.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			count, err := coll.C.CountDocuments(context.Background(), bson.M{})
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantCount, count, "Count is wrong")
		})
	}
}

func TestAccounts_RemoveNotInDiscordGuildList(t *testing.T) {
	pbclock.Mock()
	t.Parallel()

	accounts, coll := NewAccounts(t)
	defer coll.Close()

	docs := []types.Account{
		types.Account{
			ID:          primitive.NewObjectID(),
			Timestamp:   types.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
			BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake1"},
			Disabled:    true,
		},
		types.Account{
			ID:          primitive.NewObjectID(),
			Timestamp:   types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
			BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake2"},
			Disabled:    false,
		},
	}

	for _, doc := range docs {
		_, err := coll.C.InsertOne(context.Background(), doc)
		if err != nil {
			t.Fatal(err)
		}
	}

	wantDocs := []types.Account{
		types.Account{
			ID:          docs[0].ID,
			Timestamp:   types.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
			BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake1"},
			Disabled:    true,
		},
		types.Account{
			ID: docs[1].ID,
			Timestamp: types.Timestamp{
				CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond),
				UpdatedAt: iclock().Now().UTC().Truncate(time.Millisecond),
			},
			BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake2"},
			Disabled:    false,
		},
		types.Account{
			Timestamp: types.Timestamp{
				CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond),
				UpdatedAt: iclock().Now().UTC().Truncate(time.Millisecond),
			},
			BaseAccount: types.BaseAccount{GuildSnowflake: "guildsnowflake3"},
			Disabled:    false,
		},
	}

	args := []types.BaseAccount{
		wantDocs[1].BaseAccount,
		wantDocs[2].BaseAccount,
	}

	err := accounts.RemoveNotInDiscordGuildList(args)
	if err != nil {
		t.Errorf("Accounts.RemoveNotInDiscordGuildList() error %v", err)
		return
	}

	count, err := coll.C.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(3), count, "Count is wrong")

	cur, err := coll.C.Find(context.Background(), bson.M{}, &options.FindOptions{Sort: bson.M{accountKeyField: 1}})
	//.Sort(accountsKeyField)
	if err != nil {
		t.Fatal(err)
	}
	defer cur.Close(context.Background())

	docs = []types.Account{}
	for cur.Next(context.Background()) {
		var doc types.Account
		err := cur.Decode(&doc)
		if err != nil {
			t.Fatal(err)
		}

		docs = append(docs, doc)
	}

	// Since we don't know the inserted ID, we'll set it ourselves.
	wantDocs[2].ID = docs[2].ID

	assert.Equal(t, wantDocs, docs, "Docs are wrong: %v", docs)
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
				Servers: []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{
					types.Clan{Tag: "bloops"},
					types.Clan{Tag: "bloops2"},
				}}},
			},
			args: args{key: "guildsnowflake2", clan: types.Clan{Tag: "bloops2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			id := primitive.NewObjectID()
			tt.want.ID = id

			coll.C.InsertOne(context.Background(), types.Account{
				ID:        id,
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{types.Clan{Tag: "bloops"}}}},
			})

			if err := accounts.AddClan(tt.args.key, tt.args.clan); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.AddClan() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account types.Account
			result := coll.C.FindOne(context.Background(), bson.M{})
			err := result.Decode(&account)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_RemoveClan(t *testing.T) {
	t.Parallel()

	id := primitive.NewObjectID()

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
				ID:        id,
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{types.Clan{Tag: "bloops2"}}}},
			},
			args: args{key: "guildsnowflake2", clanTag: "bloops"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.InsertOne(context.Background(), types.Account{
				ID:        id,
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{types.Clan{Tag: "bloops"}, types.Clan{Tag: "bloops2"}}}},
			})

			if err := accounts.RemoveClan(tt.args.key, tt.args.clanTag); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.RemoveClan() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account types.Account
			result := coll.C.FindOne(context.Background(), bson.M{})
			err := result.Decode(&account)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_SetClans(t *testing.T) {
	t.Parallel()

	id := primitive.NewObjectID()

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
				ID:        id,
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "guildsnowflake2", Clans: []types.Clan{types.Clan{Tag: "foo"}}}},
			},
			args: args{key: "guildsnowflake2", clans: []types.Clan{types.Clan{Tag: "foo"}}},
		},
		{
			name:    "not found",
			args:    args{key: "guildsnowflake", clans: []types.Clan{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.InsertOne(context.Background(), types.Account{
				ID:        id,
				Timestamp: types.Timestamp{CreatedAt: time.Date(2014, 1, 31, 14, 50, 20, 720408938, time.UTC).Truncate(time.Millisecond)},
				Servers:   []types.Server{types.Server{Key: "guildsnowflake2"}},
			})

			if err := accounts.SetClans(tt.args.key, tt.args.clans); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.SetClans() error = %v, wantErr %v", err, tt.wantErr)
			}

			var account types.Account
			result := coll.C.FindOne(context.Background(), bson.M{serverKeyField: tt.args.key})
			result.Decode(&account)
			
			assert.Equal(t, tt.want, account)
		})
	}
}

func TestAccounts_AddServer(t *testing.T) {
	pbclock.Mock()
	type args struct {
		snowflake string
		server    types.Server
	}
	tests := []struct {
		name    string
		args    args
		want    []types.Server
		wantErr bool
	}{
		{
			name: "add",
			args: args{server: types.Server{Key: "serverkey"}, snowflake: "snowflake"},
			want: []types.Server{baseAccount.Servers[0], types.Server{
				Timestamp: types.Timestamp{CreatedAt: iclock().Now().UTC().Truncate(time.Millisecond)},
				Key: "serverkey"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts, coll := NewAccounts(t)
			defer coll.Close()

			coll.C.InsertOne(context.Background(), baseAccount)

			if err := accounts.AddServer(tt.args.snowflake, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("Accounts.AddServer() error = %v, wantErr %v", err, tt.wantErr)
			}
			var account types.Account
			result := coll.C.FindOne(context.Background(), bson.M{accountKeyField: tt.args.snowflake})
			err := result.Decode(&account)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.want, account.Servers)
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
