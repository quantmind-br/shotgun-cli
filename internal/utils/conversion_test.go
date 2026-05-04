package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		// Valid cases
		{"pure bytes", "1024", 1024, false},
		{"kilobytes", "1KB", 1024, false},
		{"megabytes", "1MB", 1024 * 1024, false},
		{"gigabytes", "1GB", 1024 * 1024 * 1024, false},
		{"bytes suffix", "512B", 512, false},
		{"lowercase kb", "500kb", 500 * 1024, false},
		{"mixed case MB", "5Mb", 5 * 1024 * 1024, false},
		{"with spaces", "  10MB  ", 10 * 1024 * 1024, false},
		{"decimal value", "1.5MB", int64(1.5 * 1024 * 1024), false},
		{"decimal kb", "2.5KB", int64(2.5 * 1024), false},
		{"large value", "100GB", 100 * 1024 * 1024 * 1024, false},
		{"zero", "0", 0, false},
		{"zero mb", "0MB", 0, false},

		// Invalid cases
		{"empty string", "", 0, true},
		{"invalid suffix", "10XB", 0, true},
		{"no number", "MB", 0, true},
		{"invalid format", "abc", 0, true},
		{"special chars", "10@MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSize(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseSizeWithDefault(t *testing.T) {
	defaultVal := int64(1024 * 1024) // 1MB

	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"valid input", "5MB", 5 * 1024 * 1024},
		{"invalid input uses default", "invalid", defaultVal},
		{"empty input uses default", "", defaultVal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSizeWithDefault(tt.input, defaultVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"kilobytes decimal", 1536, "1.5 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"megabytes decimal", int64(1.5 * 1024 * 1024), "1.5 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{"zero", 0, "0 B"},
		{"small value", 100, "100 B"},
		{"large megabytes", 100 * 1024 * 1024, "100.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkParseSize(b *testing.B) {
	inputs := []string{"1024", "10MB", "1.5GB", "500KB"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_, _ = ParseSize(input)
		}
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	inputs := []int64{512, 1024 * 1024, 1024 * 1024 * 1024}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, input := range inputs {
			_ = FormatBytes(input)
		}
	}
}
