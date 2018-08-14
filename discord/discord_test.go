package discord

import "testing"

func Test_pinString(t *testing.T) {
	t.Parallel()

	type args struct {
		pin int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"one", args{pin: 1}, "0001"},
		{"nine hundred", args{pin: 900}, "0900"},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pinString(tt.args.pin); got != tt.want {
				t.Errorf("pinString() = %v, want %v", got, tt.want)
			}
		})
	}
}
