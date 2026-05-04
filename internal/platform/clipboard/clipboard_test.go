package clipboard

import (
	"strings"
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
	if unwrapped != originalErr { //nolint:errorlint // testing exact error identity
		t.Fatalf("Unwrap should return the original error")
	}
}

func TestIsAvailable(t *testing.T) {
	// Just verify it doesn't panic - actual availability depends on system
	_ = IsAvailable()
}

func TestCopySuccess(t *testing.T) {
	if !IsAvailable() {
		t.Skip("clipboard not available in this environment")
	}

	tests := []struct {
		name    string
		content string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"unicode", "こんにちは世界"},
		{"multiline", "line1\nline2\nline3"},
		{"special chars", "tab\there\nnewline"},
		{"long text", strings.Repeat("x", 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.content)
			if err != nil {
				t.Errorf("Copy(%q) failed: %v", tt.name, err)
			}
		})
	}
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
