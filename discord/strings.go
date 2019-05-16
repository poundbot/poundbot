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

// getQuotedParts finds a "string which spans multiple spaces" in a message.
// Then takes that and replaces the Quote string with a single string value of the quote contents
func getQuotedParts(str string) []string {
	inQuote := false
	f := func(c rune) bool {
		switch {
		case c == '"':
			inQuote = !inQuote
			return true
		case inQuote:
			return false
		default:
			return c == ' '
		}
	}
	return strings.FieldsFunc(str, f)
}
