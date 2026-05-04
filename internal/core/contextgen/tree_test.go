package contextgen

import (
	"strings"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestFileNode(name, path string, isDir bool, size int64, children ...*scanner.FileNode) *scanner.FileNode {
	return &scanner.FileNode{
		Name:     name,
		Path:     path,
		IsDir:    isDir,
		Size:     size,
		Children: children,
	}
}

func TestNewTreeRenderer(t *testing.T) {
	renderer := NewTreeRenderer()

	assert.NotNil(t, renderer)
	assert.False(t, renderer.showIgnored)
	assert.Equal(t, -1, renderer.maxDepth)
}

func TestTreeRendererWithShowIgnored(t *testing.T) {
	renderer := NewTreeRenderer().WithShowIgnored(true)

	assert.True(t, renderer.showIgnored)
}

func TestTreeRendererWithMaxDepth(t *testing.T) {
	renderer := NewTreeRenderer().WithMaxDepth(3)

	assert.Equal(t, 3, renderer.maxDepth)
}

func TestTreeRendererChaining(t *testing.T) {
	renderer := NewTreeRenderer().
		WithShowIgnored(true).
		WithMaxDepth(5)

	assert.True(t, renderer.showIgnored)
	assert.Equal(t, 5, renderer.maxDepth)
}

func TestRenderTreeNilRoot(t *testing.T) {
	renderer := NewTreeRenderer()

	_, err := renderer.RenderTree(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestRenderTreeSingleFile(t *testing.T) {
	renderer := NewTreeRenderer()
	root := createTestFileNode("file.go", "/project/file.go", false, 1024)

	result, err := renderer.RenderTree(root)
	require.NoError(t, err)

	assert.Contains(t, result, "file.go")
	assert.Contains(t, result, "[1.0KB]") // size info for file
}

func TestRenderTreeSingleDirectory(t *testing.T) {
	renderer := NewTreeRenderer()
	root := createTestFileNode("project", "/project", true, 0)

	result, err := renderer.RenderTree(root)
	require.NoError(t, err)

	assert.Contains(t, result, "project/") // directory has trailing slash
}

func TestRenderTreeWithChildren(t *testing.T) {
	renderer := NewTreeRenderer()

	file1 := createTestFileNode("main.go", "/project/main.go", false, 512)
	file2 := createTestFileNode("util.go", "/project/util.go", false, 256)
	root := createTestFileNode("project", "/project", true, 0, file1, file2)

	result, err := renderer.RenderTree(root)
	require.NoError(t, err)

	assert.Contains(t, result, "project/")
	assert.Contains(t, result, "main.go")
	assert.Contains(t, result, "util.go")
	// Check for tree structure characters
	assert.True(t, strings.Contains(result, "├──") || strings.Contains(result, "└──"))
}

func TestRenderTreeNestedDirectories(t *testing.T) {
	renderer := NewTreeRenderer()

	deepFile := createTestFileNode("deep.go", "/project/src/pkg/deep.go", false, 100)
	pkgDir := createTestFileNode("pkg", "/project/src/pkg", true, 0, deepFile)
	srcDir := createTestFileNode("src", "/project/src", true, 0, pkgDir)
	root := createTestFileNode("project", "/project", true, 0, srcDir)

	result, err := renderer.RenderTree(root)
	require.NoError(t, err)

	assert.Contains(t, result, "project/")
	assert.Contains(t, result, "src/")
	assert.Contains(t, result, "pkg/")
	assert.Contains(t, result, "deep.go")
}

func TestRenderTreeMaxDepth(t *testing.T) {
	renderer := NewTreeRenderer().WithMaxDepth(1)

	deepFile := createTestFileNode("deep.go", "/project/src/pkg/deep.go", false, 100)
	pkgDir := createTestFileNode("pkg", "/project/src/pkg", true, 0, deepFile)
	srcDir := createTestFileNode("src", "/project/src", true, 0, pkgDir)
	root := createTestFileNode("project", "/project", true, 0, srcDir)

	result, err := renderer.RenderTree(root)
	require.NoError(t, err)

	assert.Contains(t, result, "project/")
	assert.Contains(t, result, "src/")
	// Files beyond depth 1 should not appear
	assert.NotContains(t, result, "pkg/")
	assert.NotContains(t, result, "deep.go")
}

func TestRenderTreeIgnoredFiles(t *testing.T) {
	t.Run("hide ignored by default", func(t *testing.T) {
		renderer := NewTreeRenderer()

		regularFile := createTestFileNode("main.go", "/project/main.go", false, 100)
		ignoredFile := createTestFileNode("ignored.go", "/project/ignored.go", false, 100)
		ignoredFile.IsGitignored = true
		root := createTestFileNode("project", "/project", true, 0, regularFile, ignoredFile)

		result, err := renderer.RenderTree(root)
		require.NoError(t, err)

		assert.Contains(t, result, "main.go")
		assert.NotContains(t, result, "ignored.go")
	})

	t.Run("show ignored when enabled", func(t *testing.T) {
		renderer := NewTreeRenderer().WithShowIgnored(true)

		regularFile := createTestFileNode("main.go", "/project/main.go", false, 100)
		ignoredFile := createTestFileNode("ignored.go", "/project/ignored.go", false, 100)
		ignoredFile.IsGitignored = true
		root := createTestFileNode("project", "/project", true, 0, regularFile, ignoredFile)

		result, err := renderer.RenderTree(root)
		require.NoError(t, err)

		assert.Contains(t, result, "main.go")
		assert.Contains(t, result, "ignored.go")
		assert.Contains(t, result, "(g)") // gitignored indicator
	})

	t.Run("custom ignored indicator", func(t *testing.T) {
		renderer := NewTreeRenderer().WithShowIgnored(true)

		file := createTestFileNode("custom.go", "/project/custom.go", false, 100)
		file.IsCustomIgnored = true
		root := createTestFileNode("project", "/project", true, 0, file)

		result, err := renderer.RenderTree(root)
		require.NoError(t, err)

		assert.Contains(t, result, "(c)") // custom ignored indicator
	})
}

func TestRenderTreeSorting(t *testing.T) {
	renderer := NewTreeRenderer()

	// Create files and directories in non-sorted order
	fileB := createTestFileNode("b.go", "/project/b.go", false, 100)
	fileA := createTestFileNode("a.go", "/project/a.go", false, 100)
	dirZ := createTestFileNode("z_dir", "/project/z_dir", true, 0)
	dirA := createTestFileNode("a_dir", "/project/a_dir", true, 0)

	root := createTestFileNode("project", "/project", true, 0, fileB, fileA, dirZ, dirA)

	result, err := renderer.RenderTree(root)
	require.NoError(t, err)

	// Directories should come before files
	dirAPos := strings.Index(result, "a_dir/")
	dirZPos := strings.Index(result, "z_dir/")
	fileAPos := strings.Index(result, "a.go")
	fileBPos := strings.Index(result, "b.go")

	assert.True(t, dirAPos < dirZPos, "directories should be sorted alphabetically")
	assert.True(t, dirZPos < fileAPos, "directories should come before files")
	assert.True(t, fileAPos < fileBPos, "files should be sorted alphabetically")
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0B"},
		{"small bytes", 100, "100B"},
		{"exactly 1KB", 1024, "1.0KB"},
		{"kilobytes", 1536, "1.5KB"},
		{"exactly 1MB", 1024 * 1024, "1.0MB"},
		{"megabytes", 1536 * 1024, "1.5MB"},
		{"exactly 1GB", 1024 * 1024 * 1024, "1.0GB"},
		{"gigabytes", 1536 * 1024 * 1024, "1.5GB"},
		{"large file", 10 * 1024 * 1024, "10.0MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIgnoreIndicator(t *testing.T) {
	renderer := NewTreeRenderer()

	t.Run("not ignored", func(t *testing.T) {
		node := createTestFileNode("file.go", "/project/file.go", false, 100)
		result := renderer.getIgnoreIndicator(node)
		assert.Equal(t, "", result)
	})

	t.Run("gitignored", func(t *testing.T) {
		node := createTestFileNode("file.go", "/project/file.go", false, 100)
		node.IsGitignored = true
		result := renderer.getIgnoreIndicator(node)
		assert.Equal(t, " (g)", result)
	})

	t.Run("custom ignored", func(t *testing.T) {
		node := createTestFileNode("file.go", "/project/file.go", false, 100)
		node.IsCustomIgnored = true
		result := renderer.getIgnoreIndicator(node)
		assert.Equal(t, " (c)", result)
	})
}

func TestGetSizeInfo(t *testing.T) {
	renderer := NewTreeRenderer()

	t.Run("directory no size", func(t *testing.T) {
		node := createTestFileNode("dir", "/project/dir", true, 0)
		result := renderer.getSizeInfo(node)
		assert.Equal(t, "", result)
	})

	t.Run("file with zero size", func(t *testing.T) {
		node := createTestFileNode("empty.txt", "/project/empty.txt", false, 0)
		result := renderer.getSizeInfo(node)
		assert.Equal(t, "", result)
	})

	t.Run("file with size", func(t *testing.T) {
		node := createTestFileNode("file.go", "/project/file.go", false, 1024)
		result := renderer.getSizeInfo(node)
		assert.Equal(t, " [1.0KB]", result)
	})
}

func TestShouldSkipNode(t *testing.T) {
	t.Run("skip beyond max depth", func(t *testing.T) {
		renderer := NewTreeRenderer().WithMaxDepth(2)
		node := createTestFileNode("file.go", "/project/file.go", false, 100)

		assert.False(t, renderer.shouldSkipNode(node, 1))
		assert.False(t, renderer.shouldSkipNode(node, 2))
		assert.True(t, renderer.shouldSkipNode(node, 3))
	})

	t.Run("skip ignored when not showing", func(t *testing.T) {
		renderer := NewTreeRenderer()
		node := createTestFileNode("file.go", "/project/file.go", false, 100)
		node.IsGitignored = true

		assert.True(t, renderer.shouldSkipNode(node, 0))
	})

	t.Run("show ignored when enabled", func(t *testing.T) {
		renderer := NewTreeRenderer().WithShowIgnored(true)
		node := createTestFileNode("file.go", "/project/file.go", false, 100)
		node.IsGitignored = true

		assert.False(t, renderer.shouldSkipNode(node, 0))
	})
}
