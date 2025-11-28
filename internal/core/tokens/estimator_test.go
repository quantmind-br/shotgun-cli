package tokens

import (
	"testing"
)

func TestEstimate(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty", "", 0},
		{"single char", "a", 1},
		{"four chars", "abcd", 1},
		{"five chars", "abcde", 2},
		{"eight chars", "abcdefgh", 2},
		{"typical sentence", "Hello, world!", 4}, // 13 chars / 4 = 3.25 -> 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Estimate(tt.text)
			if got != tt.expected {
				t.Errorf("Estimate(%q) = %d, want %d", tt.text, got, tt.expected)
			}
		})
	}
}

func TestEstimateFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected int
	}{
		{"zero", 0, 0},
		{"negative", -10, 0},
		{"one byte", 1, 1},
		{"four bytes", 4, 1},
		{"five bytes", 5, 2},
		{"1KB", 1024, 256},
		{"1MB", 1024 * 1024, 262144},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateFromBytes(tt.size)
			if got != tt.expected {
				t.Errorf("EstimateFromBytes(%d) = %d, want %d", tt.size, got, tt.expected)
			}
		})
	}
}

func TestBytesFromTokens(t *testing.T) {
	tests := []struct {
		tokens   int
		expected int64
	}{
		{0, 0},
		{1, 4},
		{100, 400},
		{1000, 4000},
	}

	for _, tt := range tests {
		got := BytesFromTokens(tt.tokens)
		if got != tt.expected {
			t.Errorf("BytesFromTokens(%d) = %d, want %d", tt.tokens, got, tt.expected)
		}
	}
}

func TestStats(t *testing.T) {
	t.Run("NewStats", func(t *testing.T) {
		stats := NewStats(1024)
		if stats.Bytes != 1024 {
			t.Errorf("Bytes = %d, want 1024", stats.Bytes)
		}
		if stats.Tokens != 256 {
			t.Errorf("Tokens = %d, want 256", stats.Tokens)
		}
	})

	t.Run("NewStatsFromText", func(t *testing.T) {
		stats := NewStatsFromText("Hello")
		if stats.Bytes != 5 {
			t.Errorf("Bytes = %d, want 5", stats.Bytes)
		}
		if stats.Tokens != 2 {
			t.Errorf("Tokens = %d, want 2", stats.Tokens)
		}
	})

	t.Run("Add", func(t *testing.T) {
		s1 := NewStats(100)
		s2 := NewStats(200)
		combined := s1.Add(s2)
		if combined.Bytes != 300 {
			t.Errorf("Combined Bytes = %d, want 300", combined.Bytes)
		}
		if combined.Tokens != 75 { // 100/4 + 200/4 = 25 + 50 = 75
			t.Errorf("Combined Tokens = %d, want 75", combined.Tokens)
		}
	})
}

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		tokens   int
		expected string
	}{
		{0, "0"},
		{100, "100"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{10000, "10.0K"},
		{100000, "100.0K"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
	}

	for _, tt := range tests {
		got := FormatTokens(tt.tokens)
		if got != tt.expected {
			t.Errorf("FormatTokens(%d) = %q, want %q", tt.tokens, got, tt.expected)
		}
	}
}

func TestCheckContextFit(t *testing.T) {
	tests := []struct {
		name       string
		tokens     int
		windowSize int
		wantFits   bool
		wantPct    float64
	}{
		{"empty in 8K", 0, Window8K, true, 0},
		{"half filled", 4096, Window8K, true, 50},
		{"exactly full", 8192, Window8K, true, 100},
		{"overflow", 10000, Window8K, false, 122.0703125},
		{"zero window", 100, 0, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckContextFit(tt.tokens, tt.windowSize)
			if got.Fits != tt.wantFits {
				t.Errorf("Fits = %v, want %v", got.Fits, tt.wantFits)
			}
			if got.Percentage != tt.wantPct {
				t.Errorf("Percentage = %v, want %v", got.Percentage, tt.wantPct)
			}
		})
	}
}
