package messages

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRaidAlert(t *testing.T) {
	type args struct {
		serverName    string
		gridPositions []string
		items         []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				serverName:    "I am a server",
				gridPositions: []string{"A1", "D10"},
				items:         []string{"foo(8)", "bar(10)"},
			},
			want: `
I am a server RAID ALERT! You are being raided!

  Locations:
    A1, D10

  Destroyed:
    foo(8), bar(10)
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RaidAlert(tt.args.serverName, tt.args.gridPositions, tt.args.items)
			assert.Equal(t, tt.want, got)
		})
	}
}
