// Package tokens provides token estimation utilities for LLM context optimization.
// It uses heuristic-based approximations to avoid dependencies on heavy tokenizer libraries.
package tokens

import (
	"fmt"
)

// Common token heuristic: approximately 1 token per 4 characters (bytes) for English text.
// This is a reasonable approximation for GPT-style tokenizers (cl100k_base, etc.)
const (
	// BytesPerToken is the average number of bytes per token (heuristic)
	BytesPerToken = 4

	// Common LLM context window sizes for reference
	Window4K   = 4096
	Window8K   = 8192
	Window16K  = 16384
	Window32K  = 32768
	Window64K  = 65536
	Window128K = 131072
)

// Estimate returns an estimated token count for the given text.
// Uses the heuristic: 1 token ~= 4 characters.
func Estimate(text string) int {
	return EstimateFromBytes(int64(len(text)))
}

// EstimateFromBytes returns an estimated token count from a byte size.
// Uses the heuristic: 1 token ~= 4 bytes.
func EstimateFromBytes(size int64) int {
	if size <= 0 {
		return 0
	}
	return int((size + BytesPerToken - 1) / BytesPerToken) // Round up
}

// BytesFromTokens converts a token count back to approximate byte size.
func BytesFromTokens(tokens int) int64 {
	return int64(tokens) * BytesPerToken
}

// Stats holds token and byte statistics for content.
type Stats struct {
	// Bytes is the raw byte count
	Bytes int64
	// Tokens is the estimated token count
	Tokens int
}

// NewStats creates Stats from a byte count.
func NewStats(bytes int64) Stats {
	return Stats{
		Bytes:  bytes,
		Tokens: EstimateFromBytes(bytes),
	}
}

// NewStatsFromText creates Stats from text content.
func NewStatsFromText(text string) Stats {
	return NewStats(int64(len(text)))
}

// Add combines two Stats, adding their values.
func (s Stats) Add(other Stats) Stats {
	return Stats{
		Bytes:  s.Bytes + other.Bytes,
		Tokens: s.Tokens + other.Tokens,
	}
}

// FormatTokens returns a human-readable token count string.
// Uses K/M suffixes for large numbers.
func FormatTokens(tokens int) string {
	switch {
	case tokens >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	case tokens >= 1_000:
		return fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	default:
		return fmt.Sprintf("%d", tokens)
	}
}

// ContextFit describes how content fits within a context window.
type ContextFit struct {
	// WindowSize is the target context window size in tokens
	WindowSize int
	// UsedTokens is the estimated tokens used
	UsedTokens int
	// Percentage is the percentage of window used (0-100+)
	Percentage float64
	// Fits indicates if content fits within the window
	Fits bool
}

// CheckContextFit evaluates how content fits within a given context window.
func CheckContextFit(tokens, windowSize int) ContextFit {
	percentage := 0.0
	if windowSize > 0 {
		percentage = float64(tokens) / float64(windowSize) * 100
	}
	return ContextFit{
		WindowSize: windowSize,
		UsedTokens: tokens,
		Percentage: percentage,
		Fits:       tokens <= windowSize,
	}
}
