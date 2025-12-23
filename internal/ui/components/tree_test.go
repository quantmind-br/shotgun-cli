package components

import (
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
	"github.com/stretchr/testify/assert"
)

func createTestNode(name, path string, isDir bool, children ...*scanner.FileNode) *scanner.FileNode {
	node := &scanner.FileNode{
		Name:     name,
		Path:     path,
		RelPath:  name, // Use name as relative path for tests
		IsDir:    isDir,
		Children: children,
	}
	// Set parent references
	for _, child := range children {
		child.Parent = node
	}

	return node
}

func TestNewFileTree(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		model := NewFileTree(nil, nil)

		assert.NotNil(t, model)
		assert.Nil(t, model.tree)
		assert.NotNil(t, model.selections)
		assert.NotNil(t, model.expanded)
	})

	t.Run("with tree", func(t *testing.T) {
		root := createTestNode("project", "/project", true)
		model := NewFileTree(root, nil)

		assert.NotNil(t, model)
		assert.Equal(t, root, model.tree)
		assert.True(t, model.expanded[root.Path]) // Root should be expanded
	})

	t.Run("with initial selections", func(t *testing.T) {
		file := createTestNode("main.go", "/project/main.go", false)
		root := createTestNode("project", "/project", true, file)

		selections := map[string]bool{
			"/project/main.go": true,
		}

		model := NewFileTree(root, selections)

		assert.True(t, model.selections["/project/main.go"])
	})
}

func TestFileTreeSetSize(t *testing.T) {
	model := NewFileTree(nil, nil)

	model.SetSize(100, 50)

	assert.Equal(t, 100, model.width)
	assert.Equal(t, 50, model.height)
}

func TestFileTreeNavigation(t *testing.T) {
	file1 := createTestNode("a.go", "/project/a.go", false)
	file2 := createTestNode("b.go", "/project/b.go", false)
	root := createTestNode("project", "/project", true, file1, file2)

	model := NewFileTree(root, nil)

	t.Run("move down", func(t *testing.T) {
		initialCursor := model.cursor
		model.MoveDown()
		assert.Equal(t, initialCursor+1, model.cursor)
	})

	t.Run("move up", func(t *testing.T) {
		model.cursor = 1
		model.MoveUp()
		assert.Equal(t, 0, model.cursor)
	})

	t.Run("move up at top", func(t *testing.T) {
		model.cursor = 0
		model.MoveUp()
		assert.Equal(t, 0, model.cursor) // Should stay at 0
	})

	t.Run("move down at bottom", func(t *testing.T) {
		model.cursor = len(model.visibleItems) - 1
		initialCursor := model.cursor
		model.MoveDown()
		assert.Equal(t, initialCursor, model.cursor) // Should stay at bottom
	})
}

func TestFileTreeExpandCollapse(t *testing.T) {
	subFile := createTestNode("sub.go", "/project/subdir/sub.go", false)
	subDir := createTestNode("subdir", "/project/subdir", true, subFile)
	file := createTestNode("main.go", "/project/main.go", false)
	root := createTestNode("project", "/project", true, subDir, file)

	model := NewFileTree(root, nil)

	// Find subdir position
	subdirIdx := -1
	for i, item := range model.visibleItems {
		if item.node == subDir {
			subdirIdx = i

			break
		}
	}

	t.Run("expand directory", func(t *testing.T) {
		// First collapse to ensure we can expand
		model.expanded[subDir.Path] = false
		model.rebuildVisibleItems()

		model.cursor = subdirIdx
		model.ExpandNode()

		assert.True(t, model.expanded[subDir.Path])
	})

	t.Run("collapse directory", func(t *testing.T) {
		model.expanded[subDir.Path] = true
		model.rebuildVisibleItems()

		// Find subdir position again
		for i, item := range model.visibleItems {
			if item.node == subDir {
				subdirIdx = i

				break
			}
		}

		model.cursor = subdirIdx
		model.CollapseNode()

		assert.False(t, model.expanded[subDir.Path])
	})

	t.Run("expand non-directory does nothing", func(t *testing.T) {
		fileIdx := -1
		for i, item := range model.visibleItems {
			if item.node == file {
				fileIdx = i

				break
			}
		}

		if fileIdx >= 0 {
			model.cursor = fileIdx
			model.ExpandNode() // Should not panic or change state
		}
	})
}

func TestFileTreeToggleSelection(t *testing.T) {
	file := createTestNode("main.go", "/project/main.go", false)
	root := createTestNode("project", "/project", true, file)

	model := NewFileTree(root, nil)

	// Find file position
	fileIdx := -1
	for i, item := range model.visibleItems {
		if item.node == file {
			fileIdx = i

			break
		}
	}

	t.Run("toggle file selection on", func(t *testing.T) {
		model.cursor = fileIdx
		model.ToggleSelection()

		assert.True(t, model.selections[file.Path])
	})

	t.Run("toggle file selection off", func(t *testing.T) {
		model.selections[file.Path] = true
		model.cursor = fileIdx
		model.ToggleSelection()

		assert.False(t, model.selections[file.Path])
	})

	t.Run("toggle on directory does nothing", func(t *testing.T) {
		model.cursor = 0 // Root directory
		initialSelections := len(model.selections)
		model.ToggleSelection()

		assert.Equal(t, initialSelections, len(model.selections))
	})
}

func TestFileTreeToggleSelectionOnDirectory(t *testing.T) {
	file1 := createTestNode("a.go", "/project/dir/a.go", false)
	file2 := createTestNode("b.go", "/project/dir/b.go", false)
	dir := createTestNode("dir", "/project/dir", true, file1, file2)
	root := createTestNode("project", "/project", true, dir)

	model := NewFileTree(root, nil)

	// Find dir position
	dirIdx := -1
	for i, item := range model.visibleItems {
		if item.node == dir {
			dirIdx = i

			break
		}
	}

	t.Run("select all files in directory", func(t *testing.T) {
		model.cursor = dirIdx
		model.ToggleSelection()

		assert.True(t, model.selections[file1.Path])
		assert.True(t, model.selections[file2.Path])
	})

	t.Run("deselect all files in directory", func(t *testing.T) {
		model.selections[file1.Path] = true
		model.selections[file2.Path] = true

		model.cursor = dirIdx
		model.ToggleSelection()

		assert.False(t, model.selections[file1.Path])
		assert.False(t, model.selections[file2.Path])
	})
}

func TestFileTreeToggleShowIgnored(t *testing.T) {
	model := NewFileTree(nil, nil)

	assert.False(t, model.showIgnored)

	model.ToggleShowIgnored()
	assert.True(t, model.showIgnored)

	model.ToggleShowIgnored()
	assert.False(t, model.showIgnored)
}

func TestFileTreeFilter(t *testing.T) {
	file1 := createTestNode("main.go", "/project/main.go", false)
	file2 := createTestNode("test.go", "/project/test.go", false)
	file3 := createTestNode("util.ts", "/project/util.ts", false)
	root := createTestNode("project", "/project", true, file1, file2, file3)

	model := NewFileTree(root, nil)

	t.Run("set filter", func(t *testing.T) {
		model.SetFilter("main")

		assert.Equal(t, "main", model.GetFilter())
	})

	t.Run("clear filter", func(t *testing.T) {
		model.SetFilter("test")
		model.ClearFilter()

		assert.Equal(t, "", model.GetFilter())
	})
}

func TestFileTreeGetSelections(t *testing.T) {
	file := createTestNode("main.go", "/project/main.go", false)
	root := createTestNode("project", "/project", true, file)

	selections := map[string]bool{
		"/project/main.go": true,
	}

	model := NewFileTree(root, selections)

	result := model.GetSelections()

	assert.Equal(t, selections, result)
}

func TestFileTreeView(t *testing.T) {
	t.Run("nil tree", func(t *testing.T) {
		model := NewFileTree(nil, nil)
		view := model.View()

		assert.Contains(t, view, "No files")
	})

	t.Run("with tree", func(t *testing.T) {
		file := createTestNode("main.go", "/project/main.go", false)
		root := createTestNode("project", "/project", true, file)

		model := NewFileTree(root, nil)
		model.SetSize(80, 20)
		view := model.View()

		assert.NotEmpty(t, view)
	})
}

func TestFileTreeRecomputeSelectionStates(t *testing.T) {
	file1 := createTestNode("a.go", "/project/dir/a.go", false)
	file2 := createTestNode("b.go", "/project/dir/b.go", false)
	dir := createTestNode("dir", "/project/dir", true, file1, file2)
	root := createTestNode("project", "/project", true, dir)

	t.Run("all files selected - directory is selected", func(t *testing.T) {
		selections := map[string]bool{
			file1.Path: true,
			file2.Path: true,
		}
		model := NewFileTree(root, selections)

		state := model.selectionStateFor(dir.Path)
		assert.Equal(t, styles.SelectionSelected, state)
	})

	t.Run("no files selected - directory is unselected", func(t *testing.T) {
		model := NewFileTree(root, nil)

		state := model.selectionStateFor(dir.Path)
		assert.Equal(t, styles.SelectionUnselected, state)
	})

	t.Run("some files selected - directory is partial", func(t *testing.T) {
		selections := map[string]bool{
			file1.Path: true,
		}
		model := NewFileTree(root, selections)

		state := model.selectionStateFor(dir.Path)
		assert.Equal(t, styles.SelectionPartial, state)
	})
}

func TestFileTreeAreAllFilesInDirSelected(t *testing.T) {
	file1 := createTestNode("a.go", "/project/dir/a.go", false)
	file2 := createTestNode("b.go", "/project/dir/b.go", false)
	dir := createTestNode("dir", "/project/dir", true, file1, file2)
	root := createTestNode("project", "/project", true, dir)

	t.Run("all selected", func(t *testing.T) {
		selections := map[string]bool{
			file1.Path: true,
			file2.Path: true,
		}
		model := NewFileTree(root, selections)

		assert.True(t, model.areAllFilesInDirSelected(dir))
	})

	t.Run("none selected", func(t *testing.T) {
		model := NewFileTree(root, nil)

		assert.False(t, model.areAllFilesInDirSelected(dir))
	})

	t.Run("partially selected", func(t *testing.T) {
		selections := map[string]bool{
			file1.Path: true,
		}
		model := NewFileTree(root, selections)

		assert.False(t, model.areAllFilesInDirSelected(dir))
	})
}

func TestFileTreeSetDirectorySelection(t *testing.T) {
	file1 := createTestNode("a.go", "/project/dir/a.go", false)
	file2 := createTestNode("b.go", "/project/dir/b.go", false)
	dir := createTestNode("dir", "/project/dir", true, file1, file2)
	root := createTestNode("project", "/project", true, dir)

	t.Run("select all", func(t *testing.T) {
		model := NewFileTree(root, nil)
		model.setDirectorySelection(dir, true)

		assert.True(t, model.selections[file1.Path])
		assert.True(t, model.selections[file2.Path])
	})

	t.Run("deselect all", func(t *testing.T) {
		selections := map[string]bool{
			file1.Path: true,
			file2.Path: true,
		}
		model := NewFileTree(root, selections)
		model.setDirectorySelection(dir, false)

		assert.False(t, model.selections[file1.Path])
		assert.False(t, model.selections[file2.Path])
	})
}

func TestFileTreeShouldShowNode(t *testing.T) {
	node := createTestNode("file.go", "/project/file.go", false)

	t.Run("normal node visible", func(t *testing.T) {
		model := NewFileTree(nil, nil)
		assert.True(t, model.shouldShowNode(node))
	})

	t.Run("gitignored node hidden by default", func(t *testing.T) {
		node.IsGitignored = true
		model := NewFileTree(nil, nil)
		assert.False(t, model.shouldShowNode(node))
		node.IsGitignored = false
	})

	t.Run("custom ignored node hidden by default", func(t *testing.T) {
		node.IsCustomIgnored = true
		model := NewFileTree(nil, nil)
		assert.False(t, model.shouldShowNode(node))
		node.IsCustomIgnored = false
	})

	t.Run("ignored node visible when showIgnored", func(t *testing.T) {
		node.IsGitignored = true
		model := NewFileTree(nil, nil)
		model.showIgnored = true
		assert.True(t, model.shouldShowNode(node))
		node.IsGitignored = false
	})

	t.Run("filter excludes non-matching", func(t *testing.T) {
		root := createTestNode("project", "/project", true, node)
		node.Parent = root
		model := NewFileTree(root, nil)
		model.SetFilter("test")
		assert.False(t, model.shouldShowNode(node)) // "file.go" doesn't match "test"
		node.Parent = nil
	})

	t.Run("filter includes matching", func(t *testing.T) {
		root := createTestNode("project", "/project", true, node)
		node.Parent = root
		model := NewFileTree(root, nil)
		model.SetFilter("file")
		assert.True(t, model.shouldShowNode(node))
		node.Parent = nil
	})
}

func TestFileTreeAdjustScroll(t *testing.T) {
	file1 := createTestNode("a.go", "/project/a.go", false)
	file2 := createTestNode("b.go", "/project/b.go", false)
	file3 := createTestNode("c.go", "/project/c.go", false)
	root := createTestNode("project", "/project", true, file1, file2, file3)

	model := NewFileTree(root, nil)
	model.SetSize(80, 2) // Only 2 visible at a time

	t.Run("scroll follows cursor down", func(t *testing.T) {
		model.cursor = 3 // Beyond visible area
		model.adjustScroll()

		assert.True(t, model.topIndex > 0)
	})

	t.Run("scroll follows cursor up", func(t *testing.T) {
		model.topIndex = 2
		model.cursor = 0
		model.adjustScroll()

		assert.Equal(t, 0, model.topIndex)
	})
}

func TestFormatFileSizeUI(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 100, "100 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFileSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTreeItemStruct(t *testing.T) {
	node := createTestNode("file.go", "/project/file.go", false)

	item := treeItem{
		node:    node,
		path:    node.Path,
		depth:   2,
		isLast:  true,
		hasNext: []bool{true, false},
	}

	assert.Equal(t, node, item.node)
	assert.Equal(t, "/project/file.go", item.path)
	assert.Equal(t, 2, item.depth)
	assert.True(t, item.isLast)
	assert.Len(t, item.hasNext, 2)
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected bool
	}{
		{"exact match", "file.go", "file.go", true},
		{"prefix match", "file.go", "file", true},
		{"suffix match", "file.go", "go", true},
		{"fuzzy match", "file_selection.go", "fsg", true},
		{"fuzzy match with path", "internal/ui/components/tree.go", "iuct", true},
		{"case insensitive", "FileSelection.go", "filesel", true},
		{"no match", "file.go", "xyz", false},
		{"empty pattern", "file.go", "", true},
		{"pattern longer than text", "a.go", "alongpattern", false},
		{"out of order", "abc", "bac", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fuzzyMatch(tt.text, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuzzyFilterShowsAncestors(t *testing.T) {
	// Create a nested structure
	file := createTestNode("target.go", "/project/src/pkg/target.go", false)
	file.RelPath = "src/pkg/target.go"

	pkg := createTestNode("pkg", "/project/src/pkg", true, file)
	pkg.RelPath = "src/pkg"

	src := createTestNode("src", "/project/src", true, pkg)
	src.RelPath = "src"

	root := createTestNode("project", "/project", true, src)
	root.RelPath = "."

	model := NewFileTree(root, nil)

	// Apply filter that matches the deepest file
	model.SetFilter("target")

	// All ancestors should be visible
	assert.True(t, model.filterMatches[root.Path], "root should be visible")
	assert.True(t, model.filterMatches[src.Path], "src should be visible")
	assert.True(t, model.filterMatches[pkg.Path], "pkg should be visible")
	assert.True(t, model.filterMatches[file.Path], "target.go should be visible")
}
