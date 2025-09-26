package math

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 2) != 4 {
		t.Fatal("expected 4")
	}
}

func TestMultiply(t *testing.T) {
	if Multiply(3, 3) != 9 {
		t.Fatal("expected 9")
	}
}
