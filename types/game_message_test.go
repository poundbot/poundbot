package types

import "testing"

func TestGameMessageEmbedStyle_ColorInt(t *testing.T) {
	type fields struct {
		Color string
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "blue",
			fields: fields{Color: "blue"},
			want:   255,
		},
		{
			name:   "red",
			fields: fields{Color: "red"},
			want:   16711680,
		},
		{
			name:   "hex 6 #f0f0f0",
			fields: fields{Color: "#f0f0f0"},
			want:   15790320,
		},
		{
			name:   "hex 3 #f0f",
			fields: fields{Color: "#f0f"},
			want:   16711935,
		},
		{
			name:   "empty",
			fields: fields{Color: ""},
			want:   255,
		},
		{
			name:   "invalid string",
			fields: fields{Color: "RONG COLOR"},
			want:   255,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := GameMessageEmbedStyle{
				Color: tt.fields.Color,
			}
			if got := es.ColorInt(); got != tt.want {
				t.Errorf("GameMessageEmbedStyle.ColorInt() = %v, want %v", got, tt.want)
			}
		})
	}
}
