package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDiffHeader(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"triple minus", "--- a/file.go", true},
		{"triple plus", "+++ b/file.go", true},
		{"minus only", "--- ", true},
		{"plus only", "+++ ", true},
		{"single minus", "- removed line", false},
		{"single plus", "+ added line", false},
		{"double minus", "-- comment", false},
		{"empty line", "", false},
		{"regular text", "some text", false},
		{"hunk header", "@@ -1,5 +1,6 @@", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDiffHeader(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGitDiffHeader(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"standard git diff", "diff --git a/file.go b/file.go", true},
		{"git diff with path", "diff --git a/path/to/file.go b/path/to/file.go", true},
		{"diff without git", "diff a/file.go b/file.go", false},
		{"regular diff header", "--- a/file.go", false},
		{"empty line", "", false},
		{"random text", "some random text", false},
		{"partial match", "diff --gi", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGitDiffHeader(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCanSplitHere(t *testing.T) {
	tests := []struct {
		name          string
		lines         []string
		index         int
		inFileSection bool
		expected      bool
	}{
		{
			name:          "before git diff header",
			lines:         []string{"context line", "diff --git a/file.go b/file.go"},
			index:         0,
			inFileSection: true,
			expected:      true,
		},
		{
			name:          "before diff header",
			lines:         []string{"context line", "--- a/file.go"},
			index:         0,
			inFileSection: true,
			expected:      true,
		},
		{
			name:          "before hunk header",
			lines:         []string{"context line", "@@ -1,5 +1,6 @@"},
			index:         0,
			inFileSection: true,
			expected:      true,
		},
		{
			name:          "at last line",
			lines:         []string{"last line"},
			index:         0,
			inFileSection: true,
			expected:      false,
		},
		{
			name:          "not in file section",
			lines:         []string{"line1", "line2"},
			index:         0,
			inFileSection: false,
			expected:      true,
		},
		{
			name:          "in middle of file section",
			lines:         []string{"line1", "line2"},
			index:         0,
			inFileSection: true,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canSplitHere(tt.lines, tt.index, tt.inFileSection)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountFiles(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "single file git diff",
			lines: []string{
				"diff --git a/file.go b/file.go",
				"--- a/file.go",
				"+++ b/file.go",
				"@@ -1,5 +1,6 @@",
				" context",
			},
			expected: 2, // git diff header and --- header
		},
		{
			name: "multiple files",
			lines: []string{
				"diff --git a/file1.go b/file1.go",
				"--- a/file1.go",
				"+++ b/file1.go",
				"diff --git a/file2.go b/file2.go",
				"--- a/file2.go",
				"+++ b/file2.go",
			},
			expected: 4,
		},
		{
			name:     "no files",
			lines:    []string{"some text", "more text"},
			expected: 0,
		},
		{
			name:     "empty",
			lines:    []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countFiles(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntelligentSplitDiff(t *testing.T) {
	tests := []struct {
		name           string
		lines          []string
		approxLines    int
		expectedChunks int
		validate       func(t *testing.T, chunks []DiffChunk)
	}{
		{
			name:           "small diff single chunk",
			lines:          []string{"line1", "line2", "line3"},
			approxLines:    10,
			expectedChunks: 1,
			validate: func(t *testing.T, chunks []DiffChunk) {
				assert.Len(t, chunks[0].Lines, 3)
			},
		},
		{
			name: "split at file boundary",
			lines: []string{
				"diff --git a/file1.go b/file1.go",
				"--- a/file1.go",
				"+++ b/file1.go",
				"@@ -1,3 +1,3 @@",
				" line1",
				" line2",
				"diff --git a/file2.go b/file2.go",
				"--- a/file2.go",
				"+++ b/file2.go",
				"@@ -1,3 +1,3 @@",
				" line1",
				" line2",
			},
			approxLines:    6,
			expectedChunks: 2,
			validate: func(t *testing.T, chunks []DiffChunk) {
				// First chunk should contain file1
				assert.True(t, strings.Contains(chunks[0].Lines[0], "file1"))
				// Second chunk should contain file2
				assert.True(t, strings.Contains(chunks[1].Lines[0], "file2"))
			},
		},
		{
			name:           "empty lines",
			lines:          []string{},
			approxLines:    10,
			expectedChunks: 1,
			validate: func(t *testing.T, chunks []DiffChunk) {
				assert.Len(t, chunks[0].Lines, 0)
			},
		},
		{
			name: "preserves hunk integrity",
			lines: []string{
				"diff --git a/file.go b/file.go",
				"--- a/file.go",
				"+++ b/file.go",
				"@@ -1,5 +1,6 @@",
				" context1",
				"-removed",
				"+added",
				" context2",
				"@@ -10,5 +11,6 @@",
				" another",
			},
			approxLines:    5,
			expectedChunks: 2,
			validate: func(t *testing.T, chunks []DiffChunk) {
				// Should split at hunk boundary, not in middle of changes
				for _, chunk := range chunks {
					assert.NotEmpty(t, chunk.Lines)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := intelligentSplitDiff(tt.lines, tt.approxLines)
			assert.Len(t, chunks, tt.expectedChunks)
			if tt.validate != nil {
				tt.validate(t, chunks)
			}
		})
	}
}

func TestSplitDiffFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test diff file
	diffContent := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {}
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,4 @@
 package util
+import "strings"
 func helper() {}
`
	inputPath := filepath.Join(tmpDir, "test.diff")
	err := os.WriteFile(inputPath, []byte(diffContent), 0o600)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "chunks")

	t.Run("split with header", func(t *testing.T) {
		err := splitDiffFile(inputPath, outputDir, 5, false)
		require.NoError(t, err)

		// Check that output files were created
		entries, err := os.ReadDir(outputDir)
		require.NoError(t, err)
		assert.NotEmpty(t, entries)

		// Verify first chunk has header
		chunkPath := filepath.Join(outputDir, "test-chunk-01.diff")
		firstChunk, err := os.ReadFile(chunkPath) //nolint:gosec // test reading controlled file
		require.NoError(t, err)
		assert.Contains(t, string(firstChunk), "# Diff Chunk")
	})

	t.Run("split without header", func(t *testing.T) {
		outputDirNoHeader := filepath.Join(tmpDir, "chunks-no-header")
		err := splitDiffFile(inputPath, outputDirNoHeader, 5, true)
		require.NoError(t, err)

		// Verify first chunk does NOT have header
		chunkPath := filepath.Join(outputDirNoHeader, "test-chunk-01.diff")
		firstChunk, err := os.ReadFile(chunkPath) //nolint:gosec // test reading controlled file
		require.NoError(t, err)
		assert.NotContains(t, string(firstChunk), "# Diff Chunk")
	})
}

func TestSplitDiffFileErrors(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.diff")
		err := os.WriteFile(emptyFile, []byte{}, 0o600)
		require.NoError(t, err)

		err = splitDiffFile(emptyFile, filepath.Join(tmpDir, "out"), 10, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("non-existent file", func(t *testing.T) {
		err := splitDiffFile(filepath.Join(tmpDir, "nonexistent.diff"), filepath.Join(tmpDir, "out"), 10, false)
		assert.Error(t, err)
	})
}

func TestWriteChunk(t *testing.T) {
	tmpDir := t.TempDir()

	chunk := DiffChunk{
		Lines:     []string{"line1", "line2", "line3"},
		FileCount: 1,
		StartLine: 1,
	}

	t.Run("with header", func(t *testing.T) {
		path := filepath.Join(tmpDir, "chunk-with-header.diff")
		err := writeChunk(path, chunk, 1, 3, false)
		require.NoError(t, err)

		content, err := os.ReadFile(path) //nolint:gosec // test reading controlled file
		require.NoError(t, err)

		assert.Contains(t, string(content), "# Diff Chunk 1 of 3")
		assert.Contains(t, string(content), "# Files in this chunk: 1")
		assert.Contains(t, string(content), "# Lines in this chunk: 3")
		assert.Contains(t, string(content), "line1")
		assert.Contains(t, string(content), "line2")
		assert.Contains(t, string(content), "line3")
	})

	t.Run("without header", func(t *testing.T) {
		path := filepath.Join(tmpDir, "chunk-no-header.diff")
		err := writeChunk(path, chunk, 1, 3, true)
		require.NoError(t, err)

		content, err := os.ReadFile(path) //nolint:gosec // test reading controlled file
		require.NoError(t, err)

		assert.NotContains(t, string(content), "# Diff Chunk")
		assert.Contains(t, string(content), "line1")
	})
}
