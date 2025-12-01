package clipboard

import (
	"testing"
)

func TestClipboardErrorFormat(t *testing.T) {
	t.Parallel()

	err := &ClipboardError{Err: nil}
	if err.Error() != "clipboard error: <nil>" {
		t.Fatalf("unexpected error format: %s", err.Error())
	}
}

func TestClipboardErrorUnwrap(t *testing.T) {
	t.Parallel()

	originalErr := &ClipboardError{Err: nil}
	wrapped := &ClipboardError{Err: originalErr}

	unwrapped := wrapped.Unwrap()
	if unwrapped != originalErr {
		t.Fatalf("Unwrap should return the original error")
	}
}

func TestIsAvailable(t *testing.T) {
	// Just verify it doesn't panic - actual availability depends on system
	_ = IsAvailable()
}

// Note: Copy function cannot be easily unit tested without mocking the system clipboard.
// The atotto/clipboard library handles platform-specific operations internally.
// Integration tests should verify clipboard functionality on actual systems.

func BenchmarkCopy(b *testing.B) {
	// Skip if clipboard is not available (CI environments)
	if !IsAvailable() {
		b.Skip("clipboard not available in this environment")
	}

	content := "benchmark content"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Copy(content); err != nil {
			b.Fatalf("Copy failed: %v", err)
		}
	}
}
