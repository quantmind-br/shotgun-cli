package storage

import (
	"errors"
	"testing"
)

func TestGetMissing(t *testing.T) {
	s := NewInMemory()
	if _, err := s.Get("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
