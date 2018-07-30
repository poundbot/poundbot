package types

import (
	"reflect"
	"testing"
)

func Test_convStringAToUnintA(t *testing.T) {
	type args struct {
		in []string
	}
	tests := []struct {
		name    string
		args    args
		want    []uint64
		wantErr bool
	}{
		{"all uint64s", args{in: []string{"1001", "2801"}}, []uint64{1001, 2801}, false},
		{"empty array", args{in: []string{}}, []uint64{}, false},
		{"negatives", args{in: []string{"1001", "-2801"}}, nil, true},
		{"empty strings", args{in: []string{"", "2801"}}, nil, true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convStringAToUnintA(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("convStringAToUnintA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convStringAToUnintA() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestClanFromServerClan(t *testing.T) {
// 	type args struct {
// 		sc ServerClan
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *Clan
// 		wantErr bool
// 	}{
// 		{
// 			"ok",
// 			args{ServerClan{Tag: "FOO", Owner: "1001", Description: "Foo Clan"}},
// 			&Clan{Tag: "FOO", OwnerID: 1001, Description: "Foo Clan"},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := ClanFromServerClan(tt.args.sc)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ClanFromServerClan() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ClanFromServerClan() = \n%v\n, want \n%v", *got, tt.want)
// 			}
// 		})
// 	}
// }
