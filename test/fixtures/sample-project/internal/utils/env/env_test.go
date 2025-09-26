package env

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	t.Setenv("SAMPLE_KEY", "value")
	if Get("SAMPLE_KEY", "fallback") != "value" {
		t.Fatal("expected value from env")
	}
}
