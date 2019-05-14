package types

import (
	"reflect"
	"testing"
)

func TestNewTimestamp(t *testing.T) {
	tests := []struct {
		name string
		want *Timestamp
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTimestamp(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimestamp_Touch(t *testing.T) {
	tests := []struct {
		name string
		t    *Timestamp
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.t.Touch()
		})
	}
}
