package rustconn

import (
	"reflect"
	"testing"

	"github.com/poundbot/poundbot/types"
)

func Test_serverClan_ToClan(t *testing.T) {
	type fields struct {
		Tag        string
		ClanTag    string
		Owner      string
		OwnerID    string
		Members    []string
		Moderators []string
	}
	tests := []struct {
		name   string
		fields fields
		want   types.Clan
	}{
		{
			name:   "RustIO Clan",
			fields: fields{Tag: "PS", Owner: "1", Members: []string{"1", "2", "3"}, Moderators: []string{"2"}},
			want:   types.Clan{Tag: "PS", OwnerID: "1", Members: []string{"1", "2", "3"}, Moderators: []string{"2"}},
		},
		{
			name:   "Clans Clan",
			fields: fields{ClanTag: "PS", OwnerID: "1", Members: []string{"1", "2", "3"}, Moderators: []string{"2"}},
			want:   types.Clan{Tag: "PS", OwnerID: "1", Members: []string{"1", "2", "3"}, Moderators: []string{"2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := serverClan{
				Tag:        tt.fields.Tag,
				ClanTag:    tt.fields.ClanTag,
				Owner:      tt.fields.Owner,
				OwnerID:    tt.fields.OwnerID,
				Members:    tt.fields.Members,
				Moderators: tt.fields.Moderators,
			}
			if got := s.ToClan(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverClan.ToClan() = %v, want %v", got, tt.want)
			}
		})
	}
}
