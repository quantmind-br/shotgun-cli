// Package clipboard provides cross-platform clipboard operations.
// It uses the atotto/clipboard library for platform-independent clipboard access.
package clipboard

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// ClipboardError represents an error during clipboard operations
type ClipboardError struct {
	Err error
}

func (e *ClipboardError) Error() string {
	return fmt.Sprintf("clipboard error: %v", e.Err)
}

func (e *ClipboardError) Unwrap() error {
	return e.Err
}

// Copy copies the given content to the system clipboard.
// It uses the atotto/clipboard library which handles platform detection internally.
func Copy(content string) error {
	if err := clipboard.WriteAll(content); err != nil {
		return &ClipboardError{Err: err}
	}

	return nil
}

// IsAvailable checks if clipboard operations are supported on this system.
func IsAvailable() bool {
	return !clipboard.Unsupported
}
