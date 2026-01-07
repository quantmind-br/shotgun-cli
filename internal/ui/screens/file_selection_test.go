package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/stretchr/testify/assert"
)

const testFilterValue = "test"

func TestNewFileSelection(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	selections := map[string]bool{
		"/root/file1.go": true,
	}

	model := NewFileSelection(fileTree, selections)

	assert.NotNil(t, model)
	assert.Equal(t, fileTree, model.fileTree)
	assert.Equal(t, selections, model.selections)
	assert.NotNil(t, model.tree)
	assert.False(t, model.filterMode)
	assert.Equal(t, "", model.filterBuffer)
}

func TestFileSelectionSetSize(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(fileTree, nil)

	model.SetSize(100, 50)

	assert.Equal(t, 100, model.width)
	assert.Equal(t, 50, model.height)
	assert.NotNil(t, model.tree)
	// Tree should have received size-6
	// (we can't directly verify this without exposing tree.width)
}

func TestFileSelectionUpdateWithNilTree(t *testing.T) {
	model := &FileSelectionModel{
		tree: nil,
	}

	cmd := model.Update(tea.KeyMsg{})
	assert.Nil(t, cmd)
}

func TestFileSelectionHandleFilterMode(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(fileTree, nil)
	model.filterMode = true
	model.filterBuffer = testFilterValue

	// Test Enter applies filter
	cmd := model.handleFilterMode(tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'x'}})
	assert.Nil(t, cmd)
	assert.False(t, model.filterMode)
	assert.Equal(t, testFilterValue, model.tree.GetFilter())

	// Test Esc cancels filter
	model.filterMode = true
	model.filterBuffer = testFilterValue
	cmd = model.handleFilterMode(tea.KeyMsg{Type: tea.KeyEsc, Runes: []rune{'x'}})
	assert.Nil(t, cmd)
	assert.False(t, model.filterMode)
	assert.Equal(t, "", model.filterBuffer)

	// Test Backspace removes character
	model.filterMode = true
	model.filterBuffer = testFilterValue
	cmd = model.handleFilterMode(tea.KeyMsg{Type: tea.KeyBackspace, Runes: []rune{'x'}})
	assert.Nil(t, cmd)
	assert.Equal(t, "tes", model.filterBuffer)

	// Test Backspace on empty buffer does nothing
	model.filterMode = true
	model.filterBuffer = ""
	cmd = model.handleFilterMode(tea.KeyMsg{Type: tea.KeyBackspace, Runes: []rune{'x'}})
	assert.Nil(t, cmd)
	assert.Equal(t, "", model.filterBuffer)

	// Test regular character adds to buffer
	model.filterMode = true
	model.filterBuffer = "te"
	cmd = model.handleFilterMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	assert.Nil(t, cmd)
	assert.Equal(t, "tes", model.filterBuffer)

	// Test space does not add to buffer (filtering restriction)
	model.filterMode = true
	model.filterBuffer = "test"
	cmd = model.handleFilterMode(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	assert.Nil(t, cmd)
	assert.Equal(t, "test", model.filterBuffer)
}

func TestFileSelectionHandleNormalMode(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/root/file1.go",
				IsDir: false,
				Size:  100,
			},
		},
	}
	selections := make(map[string]bool)
	model := NewFileSelection(fileTree, selections)
	model.SetSize(100, 50)

	// Test Up arrow moves cursor up
	cmd := model.handleNormalMode(tea.KeyMsg{Type: tea.KeyUp})
	assert.Nil(t, cmd)

	// Test Down arrow moves cursor down
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyDown})
	assert.Nil(t, cmd)

	// Test Left arrow collapses node
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Nil(t, cmd)

	// Test Right arrow expands node
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRight})
	assert.Nil(t, cmd)

	// Test Space toggles selection
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	assert.Nil(t, cmd)

	// Test 'd' toggles directory selection
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	assert.Nil(t, cmd)

	// Test 'i' toggles show ignored
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	assert.Nil(t, cmd)

	// Test '/' enters filter mode
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	assert.Nil(t, cmd)
	assert.True(t, model.filterMode)
	assert.Equal(t, "", model.filterBuffer)

	// Test F5 returns RescanRequestMsg
	model.filterMode = false // Reset filter mode
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyF5})
	assert.NotNil(t, cmd)
	// Cmd should return a RescanRequestMsg
	if msg := cmd(); msg != nil {
		assert.IsType(t, RescanRequestMsg{}, msg)
	}

	// Test Ctrl+C clears filter
	model.filterMode = false
	model.tree.SetFilter("test")
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.Nil(t, cmd)
	assert.Equal(t, "", model.tree.GetFilter())
}

func TestFileSelectionUpdate(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	selections := make(map[string]bool)
	model := NewFileSelection(fileTree, selections)

	// Test in filter mode calls handleFilterMode
	model.filterMode = true
	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.False(t, model.filterMode)

	// Test in normal mode calls handleNormalMode
	model.filterMode = false
	cmd = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Nil(t, cmd)
}

func TestFileSelectionSyncSelections(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/root/file1.go",
				IsDir: false,
				Size:  100,
			},
		},
	}
	selections := make(map[string]bool)
	model := NewFileSelection(fileTree, selections)
	model.SetSize(100, 50)

	// Move cursor down to the file (root is expanded by default, so file is visible)
	model.tree.MoveDown()

	// Toggle selection on the file (ToggleSelection only works on files, not directories)
	model.tree.ToggleSelection()

	// Call syncSelections to sync tree selections to model.selections
	model.syncSelections()

	// Verify selections were copied from tree to model's selections map
	assert.NotEmpty(t, model.selections)
	assert.True(t, model.selections["/root/file1.go"])
}

func TestFileSelectionView(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/root/file1.go",
				IsDir: false,
				Size:  100,
			},
		},
	}
	selections := map[string]bool{
		"/root/file1.go": true,
	}
	model := NewFileSelection(fileTree, selections)
	model.SetSize(100, 50)

	view := model.View()

	// Verify view contains expected elements
	assert.Contains(t, view, "Select Files")
	assert.Contains(t, view, "Selected: 1 files")
	assert.Contains(t, view, "Navigate")
	assert.Contains(t, view, "Expand/Collapse")
}

func TestFileSelectionViewWithNilSelections(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(fileTree, nil)

	view := model.View()

	// Should handle nil selections gracefully
	assert.Contains(t, view, "Select Files")
	assert.Contains(t, view, "Selected: 0 files")
}

func TestFileSelectionViewWithFilter(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:    "root",
		Path:    "/root",
		RelPath: ".",
		IsDir:   true,
		Children: []*scanner.FileNode{
			{
				Name:    "main.go",
				Path:    "/root/main.go",
				RelPath: "main.go",
				IsDir:   false,
				Size:    100,
			},
		},
	}
	fileTree.Children[0].Parent = fileTree

	model := NewFileSelection(fileTree, nil)
	model.SetSize(100, 50)
	model.tree.SetFilter("main")

	view := model.View()

	assert.Contains(t, view, "Filter: main")
}

func TestFileSelectionViewWithFilterShowsMatchCount(t *testing.T) {
	t.Parallel()

	file1 := &scanner.FileNode{
		Name:    "main.go",
		Path:    "/root/main.go",
		RelPath: "main.go",
		IsDir:   false,
		Size:    100,
	}
	file2 := &scanner.FileNode{
		Name:    "readme.md",
		Path:    "/root/readme.md",
		RelPath: "readme.md",
		IsDir:   false,
		Size:    200,
	}
	file3 := &scanner.FileNode{
		Name:    "test.go",
		Path:    "/root/test.go",
		RelPath: "test.go",
		IsDir:   false,
		Size:    150,
	}
	fileTree := &scanner.FileNode{
		Name:     "root",
		Path:     "/root",
		RelPath:  ".",
		IsDir:    true,
		Children: []*scanner.FileNode{file1, file2, file3},
	}
	file1.Parent = fileTree
	file2.Parent = fileTree
	file3.Parent = fileTree

	model := NewFileSelection(fileTree, nil)
	model.SetSize(100, 50)
	model.tree.SetFilter(".go")

	view := model.View()

	assert.Contains(t, view, "Filter:")
	assert.Contains(t, view, "(2/3)")
}

func TestFileSelectionViewInFilterMode(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(fileTree, nil)
	model.filterMode = true
	model.filterBuffer = "test"

	view := model.View()

	// Should show filter input UI
	assert.Contains(t, view, "Filter: test_")
	assert.Contains(t, view, "Type to filter")
	assert.Contains(t, view, "Enter: Apply")
}

func TestFileSelectionCalculateSelectedSize(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/root/file1.go",
				IsDir: false,
				Size:  100,
			},
			{
				Name:  "file2.go",
				Path:  "/root/file2.go",
				IsDir: false,
				Size:  200,
			},
		},
	}
	selections := map[string]bool{
		"/root/file1.go": true,
		"/root/file2.go": true,
	}
	model := NewFileSelection(fileTree, selections)

	size := model.calculateSelectedSize()

	// Should calculate total size of selected files
	assert.Equal(t, int64(300), size)
}

func TestFileSelectionCalculateSelectedSizeWithNil(t *testing.T) {
	// Test with nil fileTree
	model1 := &FileSelectionModel{
		fileTree:   nil,
		selections: map[string]bool{"/root/file1.go": true},
	}
	size := model1.calculateSelectedSize()
	assert.Equal(t, int64(0), size)

	// Test with nil selections
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model2 := &FileSelectionModel{
		fileTree:   fileTree,
		selections: nil,
	}
	size = model2.calculateSelectedSize()
	assert.Equal(t, int64(0), size)
}

func TestFileSelectionWalkTree(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "child1",
				Path:  "/root/child1",
				IsDir: false,
			},
		},
	}

	model := &FileSelectionModel{}

	visitedPaths := make([]string, 0)
	model.walkTree(fileTree, func(node *scanner.FileNode, path string) {
		visitedPaths = append(visitedPaths, path)
	})

	// Should visit all nodes
	assert.Contains(t, visitedPaths, "/root")
	assert.Contains(t, visitedPaths, "/root/child1")
	assert.Len(t, visitedPaths, 2)
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatSize(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFileSelection_LoadingState_ShowsSpinner(t *testing.T) {
	t.Parallel()

	model := NewFileSelection(nil, make(map[string]bool))
	model.SetSize(80, 24)

	view := model.View()

	assert.NotContains(t, view, "Loading file tree...")
	assert.Contains(t, view, "Scanning directory...")
	assert.True(t, model.loading)
}

func TestFileSelection_LoadedState_HidesSpinner(t *testing.T) {
	t.Parallel()

	tree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(tree, make(map[string]bool))
	model.SetSize(80, 24)

	view := model.View()

	assert.NotContains(t, view, "Scanning directory...")
	assert.False(t, model.loading)
}

func TestFileSelection_Init_LoadingState_ReturnsSpinnerTick(t *testing.T) {
	t.Parallel()

	model := NewFileSelection(nil, make(map[string]bool))

	cmd := model.Init()

	assert.NotNil(t, cmd)
}

func TestFileSelection_Init_LoadedState_ReturnsNil(t *testing.T) {
	t.Parallel()

	tree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(tree, make(map[string]bool))

	cmd := model.Init()

	assert.Nil(t, cmd)
}

func TestFileSelection_Update_LoadingState_HandlesSpinner(t *testing.T) {
	t.Parallel()

	model := NewFileSelection(nil, make(map[string]bool))

	tickMsg := model.spinner.Tick()
	cmd := model.Update(tickMsg)

	assert.NotNil(t, cmd)
}

func TestFileSelection_SetFileTree_TransitionsToLoaded(t *testing.T) {
	t.Parallel()

	model := NewFileSelection(nil, make(map[string]bool))
	model.SetSize(80, 24)

	assert.True(t, model.loading)
	assert.Nil(t, model.fileTree)

	tree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{Name: "file.go", Path: "/root/file.go", IsDir: false},
		},
	}
	model.SetFileTree(tree)

	assert.False(t, model.loading)
	assert.NotNil(t, model.fileTree)
}

func TestFileSelection_View_AfterSetFileTree_ShowsTree(t *testing.T) {
	t.Parallel()

	model := NewFileSelection(nil, make(map[string]bool))
	model.SetSize(80, 24)

	view1 := model.View()
	assert.Contains(t, view1, "Scanning")

	tree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{Name: "file.go", Path: "/root/file.go", IsDir: false},
		},
	}
	model.SetFileTree(tree)

	view2 := model.View()
	assert.NotContains(t, view2, "Scanning")
	assert.Contains(t, view2, "file.go")
}

func TestFileSelection_IsLoading(t *testing.T) {
	t.Parallel()

	t.Run("returns true when loading", func(t *testing.T) {
		t.Parallel()
		model := NewFileSelection(nil, make(map[string]bool))
		assert.True(t, model.IsLoading())
	})

	t.Run("returns false when loaded", func(t *testing.T) {
		t.Parallel()
		tree := &scanner.FileNode{Name: "root", Path: "/root", IsDir: true}
		model := NewFileSelection(tree, make(map[string]bool))
		assert.False(t, model.IsLoading())
	})
}

func TestFileSelectionSelectAllKey(t *testing.T) {
	t.Run("a key selects all visible files", func(t *testing.T) {
		fileTree := &scanner.FileNode{
			Name:  "root",
			Path:  "/root",
			IsDir: true,
			Children: []*scanner.FileNode{
				{Name: "a.go", Path: "/root/a.go", IsDir: false, Size: 100},
				{Name: "b.go", Path: "/root/b.go", IsDir: false, Size: 200},
			},
		}
		selections := make(map[string]bool)
		model := NewFileSelection(fileTree, selections)
		model.SetSize(100, 50)

		model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		assert.Len(t, model.selections, 2)
		assert.True(t, model.selections["/root/a.go"])
		assert.True(t, model.selections["/root/b.go"])
	})

	t.Run("A key deselects all visible files", func(t *testing.T) {
		fileTree := &scanner.FileNode{
			Name:  "root",
			Path:  "/root",
			IsDir: true,
			Children: []*scanner.FileNode{
				{Name: "a.go", Path: "/root/a.go", IsDir: false, Size: 100},
				{Name: "b.go", Path: "/root/b.go", IsDir: false, Size: 200},
			},
		}
		selections := map[string]bool{
			"/root/a.go": true,
			"/root/b.go": true,
		}
		model := NewFileSelection(fileTree, selections)
		model.SetSize(100, 50)

		model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})

		assert.Empty(t, model.selections)
	})

	t.Run("a key does not trigger in filter mode", func(t *testing.T) {
		fileTree := &scanner.FileNode{
			Name:  "root",
			Path:  "/root",
			IsDir: true,
			Children: []*scanner.FileNode{
				{Name: "a.go", Path: "/root/a.go", IsDir: false, Size: 100},
			},
		}
		selections := make(map[string]bool)
		model := NewFileSelection(fileTree, selections)
		model.SetSize(100, 50)
		model.filterMode = true

		model.handleFilterMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

		assert.Equal(t, "a", model.filterBuffer)
		assert.Empty(t, model.selections)
	})
}
