package types

import "testing"

func TestRaidAlert_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rn   RaidAlert
		want string
	}{
		{
			name: "test",
			rn: RaidAlert{
				ServerName:    "I am a server",
				GridPositions: []string{"A1, D10"},
				Items:         map[string]int{"foo": 8, "bar": 10},
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
			if got := tt.rn.String(); got != tt.want {
				t.Errorf("RaidAlert.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
