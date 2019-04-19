package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClanFromServerClan(t *testing.T) {
	t.Parallel()
	type args struct {
		sc ServerClan
	}
	tests := []struct {
		name    string
		args    args
		want    *Clan
		wantErr bool
	}{
		{
			"ok",
			args{ServerClan{Tag: "FOO", Owner: "1001", Description: "Foo Clan"}},
			&Clan{Tag: "FOO", OwnerID: "1001", Description: "Foo Clan"},
			false,
		},
		{
			"smol",
			args{ServerClan{Tag: "FOO", Owner: "1001"}},
			&Clan{Tag: "FOO", OwnerID: "1001"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ClanFromServerClan(tt.args.sc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClanFromServerClan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, got, tt.want, "they should be equal")
		})
	}
}
