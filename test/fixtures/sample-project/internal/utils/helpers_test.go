package utils

import "testing"

func TestNormalizeWhitespace(t *testing.T) {
	if got := NormalizeWhitespace("hello\tworld"); got != "hello world" {
		t.Fatalf("unexpected result: %s", got)
	}
}
