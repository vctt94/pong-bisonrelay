package main

import "strings"

func countLines(s string) int {
	return strings.Count(s, "\n") + 1 // +1 because the last line might not have a newline character
}
