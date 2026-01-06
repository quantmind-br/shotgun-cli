package diff

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDiffHeader(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"--- a/file.go", true},
		{"+++ b/file.go", true},
		{"--- /dev/null", true},
		{"+++ /dev/null", true},
		{"---", true},
		{"+++", true},
		{"diff --git a/file.go b/file.go", false},
		{"@@ -1,5 +1,5 @@", false},
		{"+added line", false},
		{"-removed line", false},
		{" context line", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsDiffHeader(tt.line))
		})
	}
}

func TestIsGitDiffHeader(t *testing.T) {
	tests := []struct {
		line     string
		expected bool
	}{
		{"diff --git a/file.go b/file.go", true},
		{"diff --git a/path/to/file.txt b/path/to/file.txt", true},
		{"--- a/file.go", false},
		{"+++ b/file.go", false},
		{"@@ -1,5 +1,5 @@", false},
		{"diff", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsGitDiffHeader(tt.line))
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
			name:     "empty diff",
			lines:    []string{},
			expected: 0,
		},
		{
			name: "single file with git header counts both headers",
			lines: []string{
				"diff --git a/file.go b/file.go",
				"--- a/file.go",
				"+++ b/file.go",
				"@@ -1,3 +1,3 @@",
				" line1",
				"-old",
				"+new",
			},
			expected: 2,
		},
		{
			name: "single file without git header",
			lines: []string{
				"--- a/file.go",
				"+++ b/file.go",
				"@@ -1,3 +1,3 @@",
				" line1",
			},
			expected: 1,
		},
		{
			name: "multiple files counts all git and --- headers",
			lines: []string{
				"diff --git a/file1.go b/file1.go",
				"--- a/file1.go",
				"+++ b/file1.go",
				"@@ -1,3 +1,3 @@",
				" line1",
				"diff --git a/file2.go b/file2.go",
				"--- a/file2.go",
				"+++ b/file2.go",
				"@@ -1,3 +1,3 @@",
				" line2",
				"diff --git a/file3.go b/file3.go",
				"--- a/file3.go",
				"+++ b/file3.go",
				"@@ -1,3 +1,3 @@",
				" line3",
			},
			expected: 6,
		},
		{
			name: "only git headers",
			lines: []string{
				"diff --git a/file1.go b/file1.go",
				"diff --git a/file2.go b/file2.go",
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, CountFiles(tt.lines))
		})
	}
}

func TestCanSplitAt(t *testing.T) {
	tests := []struct {
		name          string
		lines         []string
		index         int
		inFileSection bool
		expected      bool
	}{
		{
			name:          "at end of lines",
			lines:         []string{"line1", "line2"},
			index:         1,
			inFileSection: false,
			expected:      false,
		},
		{
			name:          "before git diff header",
			lines:         []string{"line1", "diff --git a/file.go b/file.go"},
			index:         0,
			inFileSection: true,
			expected:      true,
		},
		{
			name:          "before --- header",
			lines:         []string{"line1", "--- a/file.go"},
			index:         0,
			inFileSection: true,
			expected:      true,
		},
		{
			name:          "before hunk header",
			lines:         []string{"line1", "@@ -1,3 +1,3 @@"},
			index:         0,
			inFileSection: true,
			expected:      true,
		},
		{
			name:          "not in file section",
			lines:         []string{"line1", "line2", "line3"},
			index:         1,
			inFileSection: false,
			expected:      true,
		},
		{
			name:          "in file section middle of changes",
			lines:         []string{"+added", "-removed", " context"},
			index:         0,
			inFileSection: true,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, CanSplitAt(tt.lines, tt.index, tt.inFileSection))
		})
	}
}

func TestDefaultSplitConfig(t *testing.T) {
	config := DefaultSplitConfig()
	assert.Equal(t, 500, config.ApproxLines)
}

func TestIntelligentSplit_EmptyInput(t *testing.T) {
	chunks := IntelligentSplit([]string{}, DefaultSplitConfig())
	require.Len(t, chunks, 1)
	assert.Equal(t, 0, len(chunks[0].Lines))
	assert.Equal(t, 0, chunks[0].FileCount)
	assert.Equal(t, 1, chunks[0].StartLine)
}

func TestIntelligentSplit_SingleSmallFile(t *testing.T) {
	lines := []string{
		"diff --git a/file.go b/file.go",
		"--- a/file.go",
		"+++ b/file.go",
		"@@ -1,3 +1,3 @@",
		" line1",
		"-old",
		"+new",
	}

	chunks := IntelligentSplit(lines, DefaultSplitConfig())
	require.Len(t, chunks, 1)
	assert.Equal(t, 7, len(chunks[0].Lines))
	assert.GreaterOrEqual(t, chunks[0].FileCount, 1, "should have at least 1 file")
}

func TestIntelligentSplit_MultipleFilesNoSplit(t *testing.T) {
	lines := []string{
		"diff --git a/file1.go b/file1.go",
		"--- a/file1.go",
		"+++ b/file1.go",
		"@@ -1,3 +1,3 @@",
		" line1",
		"diff --git a/file2.go b/file2.go",
		"--- a/file2.go",
		"+++ b/file2.go",
		"@@ -1,3 +1,3 @@",
		" line2",
	}

	chunks := IntelligentSplit(lines, DefaultSplitConfig())
	require.Len(t, chunks, 1)
	assert.GreaterOrEqual(t, chunks[0].FileCount, 2, "should have at least 2 files")
}

func TestIntelligentSplit_ForceSplit(t *testing.T) {
	var lines []string
	for i := 0; i < 3; i++ {
		lines = append(lines,
			"diff --git a/file"+string(rune('0'+i))+".go b/file"+string(rune('0'+i))+".go",
			"--- a/file"+string(rune('0'+i))+".go",
			"+++ b/file"+string(rune('0'+i))+".go",
			"@@ -1,100 +1,100 @@",
		)
		for j := 0; j < 100; j++ {
			lines = append(lines, " context line")
		}
	}

	config := SplitConfig{ApproxLines: 50}
	chunks := IntelligentSplit(lines, config)

	assert.Greater(t, len(chunks), 1, "should split into multiple chunks")

	totalLines := TotalLines(chunks)
	assert.Equal(t, len(lines), totalLines, "total lines should match")
}

func TestIntelligentSplit_ZeroApproxLines(t *testing.T) {
	lines := []string{
		"diff --git a/file.go b/file.go",
		"--- a/file.go",
		"+++ b/file.go",
	}

	config := SplitConfig{ApproxLines: 0}
	chunks := IntelligentSplit(lines, config)

	require.Len(t, chunks, 1)
	assert.Equal(t, 3, len(chunks[0].Lines))
}

func TestIntelligentSplit_NegativeApproxLines(t *testing.T) {
	lines := []string{
		"diff --git a/file.go b/file.go",
		"--- a/file.go",
		"+++ b/file.go",
	}

	config := SplitConfig{ApproxLines: -10}
	chunks := IntelligentSplit(lines, config)

	require.Len(t, chunks, 1)
}

func TestIntelligentSplit_PreservesFileBoundaries(t *testing.T) {
	lines := []string{
		"diff --git a/file1.go b/file1.go",
		"--- a/file1.go",
		"+++ b/file1.go",
		"@@ -1,5 +1,5 @@",
		" line1",
		" line2",
		" line3",
		" line4",
		" line5",
		"diff --git a/file2.go b/file2.go",
		"--- a/file2.go",
		"+++ b/file2.go",
		"@@ -1,5 +1,5 @@",
		" line1",
		" line2",
	}

	config := SplitConfig{ApproxLines: 10}
	chunks := IntelligentSplit(lines, config)

	for i, chunk := range chunks {
		content := strings.Join(chunk.Lines, "\n")
		if strings.Contains(content, "diff --git") {
			assert.True(t,
				strings.HasPrefix(chunk.Lines[0], "diff --git") || i == 0,
				"chunk should start with diff header or be first chunk")
		}
	}
}

func TestTotalLines(t *testing.T) {
	chunks := []Chunk{
		{Lines: []string{"a", "b", "c"}},
		{Lines: []string{"d", "e"}},
		{Lines: []string{"f"}},
	}

	assert.Equal(t, 6, TotalLines(chunks))
}

func TestTotalLines_Empty(t *testing.T) {
	assert.Equal(t, 0, TotalLines([]Chunk{}))
}

func TestTotalFiles(t *testing.T) {
	chunks := []Chunk{
		{FileCount: 2},
		{FileCount: 3},
		{FileCount: 1},
	}

	assert.Equal(t, 6, TotalFiles(chunks))
}

func TestTotalFiles_Empty(t *testing.T) {
	assert.Equal(t, 0, TotalFiles([]Chunk{}))
}

func TestIntelligentSplit_RealWorldDiff(t *testing.T) {
	diff := `diff --git a/cmd/root.go b/cmd/root.go
index abc123..def456 100644
--- a/cmd/root.go
+++ b/cmd/root.go
@@ -10,6 +10,7 @@ import (
 	"fmt"
 	"os"
+	"strings"
 )
 
 func main() {
@@ -20,3 +21,10 @@ func main() {
 	fmt.Println("Hello")
+	// New code
+	for i := 0; i < 10; i++ {
+		fmt.Println(i)
+	}
 }
diff --git a/internal/pkg/util.go b/internal/pkg/util.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/internal/pkg/util.go
@@ -0,0 +1,15 @@
+package pkg
+
+func Helper() string {
+	return "helper"
+}
`
	lines := strings.Split(diff, "\n")
	chunks := IntelligentSplit(lines, DefaultSplitConfig())

	require.Len(t, chunks, 1, "small diff should not be split")
	assert.GreaterOrEqual(t, chunks[0].FileCount, 2, "should detect at least 2 file sections")
}

func TestChunk_Structure(t *testing.T) {
	chunk := Chunk{
		Lines:     []string{"line1", "line2"},
		FileCount: 3,
		StartLine: 42,
	}

	assert.Equal(t, 2, len(chunk.Lines))
	assert.Equal(t, 3, chunk.FileCount)
	assert.Equal(t, 42, chunk.StartLine)
}
