package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/stretchr/testify/assert"
)

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

	cmd := model.Update(tea.KeyMsg{}, nil)
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
	model.filterBuffer = "test"

	// Test Enter applies filter
	cmd := model.handleFilterMode(tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'x'}})
	assert.Nil(t, cmd)
	assert.False(t, model.filterMode)
	assert.Equal(t, "test", model.tree.GetFilter())

	// Test Esc cancels filter
	model.filterMode = true
	model.filterBuffer = "test"
	cmd = model.handleFilterMode(tea.KeyMsg{Type: tea.KeyEsc, Runes: []rune{'x'}})
	assert.Nil(t, cmd)
	assert.False(t, model.filterMode)
	assert.Equal(t, "", model.filterBuffer)

	// Test Backspace removes character
	model.filterMode = true
	model.filterBuffer = "test"
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
	model := NewFileSelection(fileTree, nil)
	model.SetSize(100, 50)

	// Create a test selections map
	testSelections := make(map[string]bool)

	// Test Up arrow moves cursor up
	cmd := model.handleNormalMode(tea.KeyMsg{Type: tea.KeyUp}, testSelections)
	assert.Nil(t, cmd)

	// Test Down arrow moves cursor down
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyDown}, testSelections)
	assert.Nil(t, cmd)

	// Test Left arrow collapses node
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyLeft}, testSelections)
	assert.Nil(t, cmd)

	// Test Right arrow expands node
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRight}, testSelections)
	assert.Nil(t, cmd)

	// Test Space toggles selection
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}, testSelections)
	assert.Nil(t, cmd)
	// Selection count should have changed
	// (can't verify exact count without knowing initial state)

	// Test 'd' toggles directory selection
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, testSelections)
	assert.Nil(t, cmd)

	// Test 'i' toggles show ignored
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}, testSelections)
	assert.Nil(t, cmd)

	// Test '/' enters filter mode
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}, testSelections)
	assert.Nil(t, cmd)
	assert.True(t, model.filterMode)
	assert.Equal(t, "", model.filterBuffer)

	// Test F5 returns RescanRequestMsg
	model.filterMode = false // Reset filter mode
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyF5}, testSelections)
	assert.NotNil(t, cmd)
	// Cmd should return a RescanRequestMsg
	if msg := cmd(); msg != nil {
		assert.IsType(t, RescanRequestMsg{}, msg)
	}

	// Test Ctrl+C clears filter
	model.filterMode = false
	model.tree.SetFilter("test")
	cmd = model.handleNormalMode(tea.KeyMsg{Type: tea.KeyCtrlC}, testSelections)
	assert.Nil(t, cmd)
	assert.Equal(t, "", model.tree.GetFilter())
}

func TestFileSelectionUpdate(t *testing.T) {
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
	}
	model := NewFileSelection(fileTree, nil)

	// Test in filter mode calls handleFilterMode
	model.filterMode = true
	cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter}, nil)
	assert.Nil(t, cmd)
	assert.False(t, model.filterMode)

	// Test in normal mode calls handleNormalMode
	model.filterMode = false
	testSelections := make(map[string]bool)
	cmd = model.Update(tea.KeyMsg{Type: tea.KeyUp}, testSelections)
	assert.Nil(t, cmd)
}

func TestFileSelectionUpdateSelections(t *testing.T) {
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
	model := NewFileSelection(fileTree, nil)
	model.SetSize(100, 50)

	// Move cursor down to the file (root is expanded by default, so file is visible)
	model.tree.MoveDown()

	// Toggle selection on the file (ToggleSelection only works on files, not directories)
	model.tree.ToggleSelection()

	// Create a selections map to update
	testSelections := make(map[string]bool)

	// Call updateSelections
	model.updateSelections(testSelections)

	// Verify selections were copied from tree to map
	assert.NotEmpty(t, testSelections)
	assert.True(t, testSelections["/root/file1.go"])
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
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "main.go",
				Path:  "/root/main.go",
				IsDir: false,
				Size:  100,
			},
		},
	}
	model := NewFileSelection(fileTree, nil)
	model.SetSize(100, 50)
	model.tree.SetFilter("main")

	view := model.View()

	// Should show filter in stats
	assert.Contains(t, view, "Filter: main")
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
