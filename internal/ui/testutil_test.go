package ui

import "regexp"

// ansiRe matches any ANSI escape sequence so tests can check plain text content
// independent of lipgloss styling.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}
