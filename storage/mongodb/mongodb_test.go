package mongodb

import (
	"reflect"
	"testing"
)

func Test_parseDialURL(t *testing.T) {
	type args struct {
		dialURL string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name:    "wrong scheme",
			args:    args{dialURL: "http://localhost"},
			wantErr: true,
		},
		{
			name:    "normal",
			args:    args{dialURL: "mongodb://localhost/"},
			want:    &Config{DialAddress: "mongodb://localhost/"},
			wantErr: false,
		},
		{
			name:    "ssl",
			args:    args{dialURL: "mongodb://localhost/?ssl=true"},
			want:    &Config{DialAddress: "mongodb://localhost/", SSL: true},
			wantErr: false,
		},
		{
			name:    "ssl insecure",
			args:    args{dialURL: "mongodb://localhost/?ssl=true&tlsInsecure=true"},
			want:    &Config{DialAddress: "mongodb://localhost/", SSL: true, InsecureSSL: true},
			wantErr: false,
		},
		{
			name:    "tls",
			args:    args{dialURL: "mongodb://localhost/?tls=true"},
			want:    &Config{DialAddress: "mongodb://localhost/", SSL: true},
			wantErr: false,
		},
		{
			name:    "tls insecure",
			args:    args{dialURL: "mongodb://localhost/?tls=true&tlsInsecure=true"},
			want:    &Config{DialAddress: "mongodb://localhost/", SSL: true, InsecureSSL: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDialURL(tt.args.dialURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDialURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDialURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
