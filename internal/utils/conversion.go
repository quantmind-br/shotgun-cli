package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseSize parses a size string (e.g., "1MB", "5GB", "500KB") and returns the size in bytes.
// Supports GB, MB, KB, and B suffixes. Pure numeric values are treated as bytes.
func ParseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Handle numeric-only input (assume bytes)
	if val, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return val, nil
	}

	// Parse with units
	var multiplier int64
	var numStr string

	switch {
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		numStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		numStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "B"):
		multiplier = 1
		numStr = strings.TrimSuffix(sizeStr, "B")
	default:
		return 0, fmt.Errorf("invalid size format, use KB, MB, GB, or B")
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value: %w", err)
	}

	return int64(val * float64(multiplier)), nil
}

// ParseSizeWithDefault parses a size string and returns a default value on error.
// Useful for configuration parsing where a fallback is acceptable.
func ParseSizeWithDefault(sizeStr string, defaultSize int64) int64 {
	size, err := ParseSize(sizeStr)
	if err != nil {
		return defaultSize
	}

	return size
}

// FormatBytes formats a byte count into a human-readable string (e.g., "1.5 MB").
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
