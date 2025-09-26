package utils

import "strings"

func NormalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
