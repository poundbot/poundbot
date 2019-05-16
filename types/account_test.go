package types

import (
	"reflect"
	"testing"
)

func Benchmark_GetRegisteredPlayerIDs(b *testing.B) {
	account := Account{}
	account.RegisteredPlayerIDs = []string{
		"game:453645665675876547567545",
		"game:456756996657657645687765",
		"game:5647564658899764567657",
		"game:34566658765766546577567",
		"game:76574654526534567435466",
		"game:234575648455476345643565",
		"game:37354656664563345548",
		"game:456435645686587653455346",
		"game:54675675675647546756473456",
		"game:54675675467456745675467",
		"game:5467656475676576457657",
		"game:432553245455234554435",
	}
	for n := 0; n < b.N; n++ {
		account.GetRegisteredPlayerIDs("game")
	}
}

func TestServer_UsersClan(t *testing.T) {
	clans := []Clan{{Tag: "FoF", Members: []string{"one", "two"}}}
	type args struct {
		playerIDs []string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 Clan
	}{
		{
			name:  "User is in no clans",
			args:  args{playerIDs: []string{"three"}},
			want:  false,
			want1: Clan{},
		},
		{
			name:  "User is in clan",
			args:  args{playerIDs: []string{"two"}},
			want:  true,
			want1: clans[0],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Server{
				Clans: clans,
			}
			got, got1 := s.UsersClan(tt.args.playerIDs)
			if got != tt.want {
				t.Errorf("Server.UsersClan() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Server.UsersClan() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestAccount_ServerFromKey(t *testing.T) {
	servers := []Server{
		{Key: "one"},
		{Key: "two"},
	}
	type args struct {
		apiKey string
	}
	tests := []struct {
		name    string
		args    args
		want    Server
		wantErr bool
	}{
		{
			name:    "Server does not exist",
			args:    args{apiKey: "three"},
			want:    Server{},
			wantErr: true,
		},
		{
			name:    "Server exists",
			args:    args{apiKey: "two"},
			want:    servers[1],
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				Servers: servers,
			}
			got, err := a.ServerFromKey(tt.args.apiKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("Account.ServerFromKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.ServerFromKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_GetCommandPrefix(t *testing.T) {
	tests := []struct {
		name          string
		commandPrefix string
		want          string
	}{
		{
			name: "no command prefix",
			want: "!pb",
		},
		{
			name:          "command prefix",
			commandPrefix: "!awwwyeah",
			want:          "!awwwyeah",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: BaseAccount{CommandPrefix: tt.commandPrefix},
			}
			if got := a.GetCommandPrefix(); got != tt.want {
				t.Errorf("Account.GetCommandPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_GetAdminIDs(t *testing.T) {
	tests := []struct {
		name        string
		baseAccount BaseAccount
		want        []string
	}{
		{
			name:        "owner only",
			baseAccount: BaseAccount{OwnerSnowflake: "one"},
			want:        []string{"one"},
		},
		{
			name:        "owner and admins",
			baseAccount: BaseAccount{OwnerSnowflake: "one", AdminSnowflakes: []string{"two", "three"}},
			want:        []string{"two", "three", "one"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: tt.baseAccount,
			}
			if got := a.GetAdminIDs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.GetAdminIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClan_SetGame(t *testing.T) {
	type fields struct {
		Tag        string
		OwnerID    string
		Members    []string
		Moderators []string
	}
	type args struct {
		game string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Clan{
				Tag:        tt.fields.Tag,
				OwnerID:    tt.fields.OwnerID,
				Members:    tt.fields.Members,
				Moderators: tt.fields.Moderators,
			}
			c.SetGame(tt.args.game)
		})
	}
}

func TestAccount_GetRegisteredPlayerIDs(t *testing.T) {
	type fields struct {
		BaseAccount BaseAccount
	}
	type args struct {
		game string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name:   "empty",
			args:   args{"game"},
			fields: fields{BaseAccount: BaseAccount{}},
			want:   []string{},
		},
		{
			name: "different game",
			args: args{"game"},
			fields: fields{BaseAccount: BaseAccount{
				RegisteredPlayerIDs: []string{"rust:1234", "rust:2345"},
			}},
			want: []string{},
		},
		{
			name: "mixed game",
			args: args{"game"},
			fields: fields{BaseAccount: BaseAccount{
				RegisteredPlayerIDs: []string{
					"rust:1234", "rust:2345",
					"game:3456", "game:4567",
				},
			}},
			want: []string{"3456", "4567"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Account{
				BaseAccount: tt.fields.BaseAccount,
			}
			if got := a.GetRegisteredPlayerIDs(tt.args.game); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Account.GetRegisteredPlayerIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_ChannelIDForTag(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name        string
		s           Server
		args        args
		wantChannel string
		wantFound   bool
	}{
		{
			name: "channel exists",
			s: Server{
				Channels: []ServerChannel{{ChannelID: "1234", Tags: []string{"chat"}}},
			},
			args:        args{"chat"},
			wantChannel: "1234",
			wantFound:   true,
		},
		{
			name: "channel does not exist",
			s: Server{
				Channels: []ServerChannel{{ChannelID: "1234", Tags: []string{"chat"}}},
			},
			args:        args{"cha"},
			wantChannel: "",
			wantFound:   false,
		},
		{
			name:        "no channels exist",
			args:        args{"cha"},
			wantChannel: "",
			wantFound:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotChannel, gotFound := tt.s.ChannelIDForTag(tt.args.tag)
			if gotChannel != tt.wantChannel {
				t.Errorf("Server.ChannelIDForTag() gotChannel = %v, want %v", gotChannel, tt.wantChannel)
			}
			if gotFound != tt.wantFound {
				t.Errorf("Server.ChannelIDForTag() gotFound = %v, want %v", gotFound, tt.wantFound)
			}
		})
	}
}

func TestServer_SetChannelIDForTag(t *testing.T) {
	t.Parallel()
	type fields struct {
		Channels []ServerChannel
	}
	type args struct {
		channel string
		tag     string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantChannels []ServerChannel
		want         bool
	}{
		{
			name: "1. channel id and tag are new",
			args: args{channel: "9876", tag: "newchat"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat"}},
				{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
				{ChannelID: "9876", Tags: []string{"newchat"}},
			},
			want: true,
		},
		{
			name: "2. tag is new and channel exists",
			args: args{channel: "1234", tag: "newchat"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat", "newchat"}},
				{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
			},
			want: true,
		},
		{
			name: "3. channel is new and old channel has no other tags",
			args: args{channel: "9876", tag: "chat"},
			wantChannels: []ServerChannel{
				{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
				{ChannelID: "9876", Tags: []string{"chat"}},
			},
			want: true,
		},
		{
			name: "3.1. channel is new and old channel has no other tags, last channel in list",
			args: args{channel: "9876", tag: "lastchan"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat"}},
				{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
				{ChannelID: "9876", Tags: []string{"lastchan"}},
			},
			want: true,
		},
		{
			name: "4. channel is new and old has other tags",
			args: args{channel: "9876", tag: "serverchat"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat"}},
				{ChannelID: "5678", Tags: []string{"lasttag"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
				{ChannelID: "9876", Tags: []string{"serverchat"}},
			},
			want: true,
		},
		{
			name: "4.1. channel is new and old has other tags, last tag in list",
			args: args{channel: "9876", tag: "lasttag"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat"}},
				{ChannelID: "5678", Tags: []string{"serverchat"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
				{ChannelID: "9876", Tags: []string{"lasttag"}},
			},
			want: true,
		},
		{
			name: "5. nothing is new on single tag",
			args: args{channel: "1234", tag: "chat"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat"}},
				{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
			},
		},
		{
			name: "6. nothing is new on multiple tags",
			args: args{channel: "5678", tag: "lasttag"},
			wantChannels: []ServerChannel{
				{ChannelID: "1234", Tags: []string{"chat"}},
				{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
				{ChannelID: "7654", Tags: []string{"lastchan"}},
			},
		},
	}
	for _, tt := range tests {
		baseChannels := []ServerChannel{
			{ChannelID: "1234", Tags: []string{"chat"}},
			{ChannelID: "5678", Tags: []string{"serverchat", "lasttag"}},
			{ChannelID: "7654", Tags: []string{"lastchan"}},
		}
		t.Run(tt.name, func(t *testing.T) {
			s := Server{
				Channels: baseChannels,
			}

			if got := s.SetChannelIDForTag(tt.args.channel, tt.args.tag); got != tt.want {
				t.Errorf("Server.SetChannelIDForTag(%s) resulting channels = %v, want %v", tt.args, got, tt.want)
			}

			if !reflect.DeepEqual(s.Channels, tt.wantChannels) {
				t.Errorf("Server.SetChannelIDForTag(%s) resulting channels = %v, want %v", tt.args, s.Channels, tt.wantChannels)
			}
		})
	}
}
