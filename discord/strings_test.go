package discord

import (
	"testing"
)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pinString(tt.args.pin); got != tt.want {
				t.Errorf("pinString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_escapeDiscordString(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "@everyone", want: "@\u200Beveryone"},
		{name: "@here", want: "@\u200Bhere"},
		{name: "\\backslash\\", want: "\\\\backslash\\\\"},
		{name: "`code`", want: "\\`code\\`"},
		{name: "||spoiler||", want: "\\||spoiler\\||"},
		{name: "~~strikethrough~~", want: "\\~~strikethrough\\~~"},
		{name: "*italics*", want: "\\*italics\\*"},
		{name: "**bold**", want: "\\*\\*bold\\*\\*"},
		{name: "__underline__", want: "\\_\\_underline\\_\\_"},
		{name: "\\__***underline bold italics***__\\", want: "\\\\\\_\\_\\*\\*\\*underline bold italics\\*\\*\\*\\_\\_\\\\"},
		{name: "<@123456>", want: "\\<@123456>"},
		{name: "<@!123456>", want: "\\<@!123456>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeDiscordString(tt.name); got != tt.want {
				t.Errorf("escapeDiscordString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_truncateString(t *testing.T) {
	type args struct {
		str string
		num int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "No Truncation",
			args: args{str: "This should not truncate", num: 50},
			want: "This should not truncate",
		},
		{
			name: "Truncation on word",
			args: args{str: "123 56789012 4567", num: 13},
			want: "123 56789012…",
		},
		{
			name: "Truncation is end of word",
			args: args{str: "123 56789012 4567", num: 12},
			want: "123…",
		},
		{
			name: "Exact size",
			args: args{str: "123 56789012 4567", num: 17},
			want: "123 56789012 4567",
		},
		{
			name: "Truncate big string",
			args: args{str: "bigstring", num: 4},
			want: "big…",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateString(tt.args.str, tt.args.num); got != tt.want {
				t.Errorf("truncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}
