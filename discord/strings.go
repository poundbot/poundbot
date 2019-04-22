package discord

import (
	"fmt"
	"strings"
)

func pinString(pin int) string {
	return fmt.Sprintf("%04d", pin)
}

// truncateString truncates a string, adding an elipsis
// used for sending messages to games.
func truncateString(str string, num int) string {
	if len(str) <= num {
		return str
	}

	words := strings.Fields(str)
	firstWord := true
	var out string

	for i := range words {
		if len(out)+len(words[i]) >= num-1 {
			if firstWord {
				out = str[0 : num-1]
				break
			}
			break
		}
		if firstWord {
			firstWord = false
			out = words[i]
			continue
		}
		out = fmt.Sprintf("%s %s", out, words[i])
	}
	return out + "…"
}

func escapeDiscordString(s string) string {
	r := strings.NewReplacer(
		"@everyone", "@\u200Beveryone",
		"@here", "@\u200Bhere",
		"\\", "\\\\",
		"`", "\\`",
		"||", "\\||",
		"*", "\\*",
		"~~", "\\~~",
		"_", "\\_",
		"<@", "\\<@",
	)
	return r.Replace(s)
}
