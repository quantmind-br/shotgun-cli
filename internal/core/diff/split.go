// Package diff provides utilities for processing and splitting git diff files.
package diff

import (
	"strings"
)

// Chunk represents a portion of a diff that has been split.
type Chunk struct {
	// Lines contains the diff content lines for this chunk.
	Lines []string

	// FileCount is the number of files represented in this chunk.
	FileCount int

	// StartLine is the original line number where this chunk starts (1-indexed).
	StartLine int
}

// SplitConfig configures how diffs are split.
type SplitConfig struct {
	// ApproxLines is the target number of lines per chunk.
	// The actual chunk size may vary to preserve file boundaries.
	ApproxLines int
}

// DefaultSplitConfig returns a default split configuration.
func DefaultSplitConfig() SplitConfig {
	return SplitConfig{
		ApproxLines: 500,
	}
}

// IntelligentSplit splits a diff into chunks while preserving file boundaries.
//
// The algorithm:
// 1. Tracks file boundaries by detecting diff headers
// 2. Accumulates lines until ApproxLines threshold is reached
// 3. Only splits at safe points (between files, before hunk headers)
// 4. Ensures each chunk maintains proper diff structure
//
// Parameters:
//   - lines: The diff content as a slice of strings (one per line)
//   - config: Configuration for the split operation
//
// Returns a slice of Chunk, each containing a valid portion of the diff.
func IntelligentSplit(lines []string, config SplitConfig) []Chunk {
	if len(lines) == 0 {
		return []Chunk{{Lines: lines, FileCount: 0, StartLine: 1}}
	}

	if config.ApproxLines <= 0 {
		config.ApproxLines = 500
	}

	var chunks []Chunk
	var currentChunk Chunk
	currentFileLines := make([]string, 0, config.ApproxLines)
	fileCount := 0
	inFileSection := false

	for i, line := range lines {
		currentChunk.Lines = append(currentChunk.Lines, line)
		currentFileLines = append(currentFileLines, line)

		if IsDiffHeader(line) || IsGitDiffHeader(line) {
			if inFileSection && len(currentFileLines) > 1 {
				fileCount++
			}
			inFileSection = true
			currentFileLines = []string{line}
		}

		shouldSplit := len(currentChunk.Lines) >= config.ApproxLines &&
			CanSplitAt(lines, i, inFileSection)

		if shouldSplit {
			currentChunk.FileCount = fileCount
			if inFileSection && len(currentFileLines) > 1 {
				currentChunk.FileCount++
			}
			chunks = append(chunks, currentChunk)

			currentChunk = Chunk{StartLine: i + 2}
			fileCount = 0
			inFileSection = false
			currentFileLines = []string{}
		}
	}

	if len(currentChunk.Lines) > 0 {
		if inFileSection && len(currentFileLines) > 1 {
			fileCount++
		}
		currentChunk.FileCount = fileCount
		chunks = append(chunks, currentChunk)
	}

	if len(chunks) == 0 {
		chunks = append(chunks, Chunk{
			Lines:     lines,
			FileCount: CountFiles(lines),
			StartLine: 1,
		})
	}

	return chunks
}

// IsDiffHeader returns true if the line is a unified diff header (--- or +++).
func IsDiffHeader(line string) bool {
	return strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++")
}

// IsGitDiffHeader returns true if the line is a git diff header.
func IsGitDiffHeader(line string) bool {
	return strings.HasPrefix(line, "diff --git")
}

// CanSplitAt returns true if the given index is a safe split point.
// Safe split points are:
// - Before a new file diff starts
// - At the end of a complete file section
// - Before hunk headers (@@)
func CanSplitAt(lines []string, index int, inFileSection bool) bool {
	if index >= len(lines)-1 {
		return false
	}

	nextLine := lines[index+1]

	// Good places to split:
	// 1. Before a new file diff starts
	if IsGitDiffHeader(nextLine) || IsDiffHeader(nextLine) {
		return true
	}

	// 2. At the end of a complete file section (after context lines)
	if !inFileSection {
		return true
	}

	// 3. Between files but not in the middle of a file change
	if strings.HasPrefix(nextLine, "@@") {
		// This is a hunk header, safe to split before it
		return true
	}

	return false
}

// CountFiles counts the number of files in a diff.
// It counts git diff headers and unified diff --- headers.
func CountFiles(lines []string) int {
	count := 0
	for _, line := range lines {
		if IsGitDiffHeader(line) || (IsDiffHeader(line) && strings.HasPrefix(line, "---")) {
			count++
		}
	}

	return count
}

// TotalLines returns the total number of lines across all chunks.
func TotalLines(chunks []Chunk) int {
	total := 0
	for _, chunk := range chunks {
		total += len(chunk.Lines)
	}
	return total
}

// TotalFiles returns the total number of files across all chunks.
func TotalFiles(chunks []Chunk) int {
	total := 0
	for _, chunk := range chunks {
		total += chunk.FileCount
	}
	return total
}
