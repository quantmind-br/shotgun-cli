package handlers

import "testing"

func TestServeShutsDown(t *testing.T) {
	if err := Serve(); err != nil {
		t.Fatalf("Serve returned error: %v", err)
	}
}
