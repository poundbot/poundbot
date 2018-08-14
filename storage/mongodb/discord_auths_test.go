package mongodb

import (
	"testing"

	"bitbucket.org/mrpoundsign/poundbot/storage/mongodb/mongotest"
	"bitbucket.org/mrpoundsign/poundbot/types"
)

func NewDiscordAuths(t *testing.T) (*DiscordAuths, *mongotest.Collection) {
	coll, err := mongotest.NewCollection(discordAuthsCollection)
	if err != nil {
		t.Fatal(err)
	}
	return &DiscordAuths{collection: coll.C}, coll
}

func TestDiscordAuths_Get(t *testing.T) {
	type args struct {
		discordName string
		da          *types.DiscordAuth
	}
	tests := []struct {
		name    string
		d       DiscordAuths
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Get(tt.args.discordName, tt.args.da); (err != nil) != tt.wantErr {
				t.Errorf("DiscordAuths.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscordAuths_GetSnowflake(t *testing.T) {
	type args struct {
		snowflake string
		da        *types.DiscordAuth
	}
	tests := []struct {
		name    string
		d       DiscordAuths
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.GetSnowflake(tt.args.snowflake, tt.args.da); (err != nil) != tt.wantErr {
				t.Errorf("DiscordAuths.GetSnowflake() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscordAuths_Remove(t *testing.T) {
	type args struct {
		si types.SteamInfo
	}
	tests := []struct {
		name    string
		d       DiscordAuths
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Remove(tt.args.si); (err != nil) != tt.wantErr {
				t.Errorf("DiscordAuths.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiscordAuths_Upsert(t *testing.T) {
	type args struct {
		da types.DiscordAuth
	}
	tests := []struct {
		name    string
		d       DiscordAuths
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.d.Upsert(tt.args.da); (err != nil) != tt.wantErr {
				t.Errorf("DiscordAuths.Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
