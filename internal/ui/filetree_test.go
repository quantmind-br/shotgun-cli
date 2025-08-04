package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"shotgun-cli/internal/core"
)

// Helper function to create a test file tree
func createTestFileTree() *core.FileNode {
	return &core.FileNode{
		Name:    "project",
		Path:    "/project",
		RelPath: "",
		IsDir:   true,
		Children: []*core.FileNode{
			{
				Name:    "src",
				Path:    "/project/src",
				RelPath: "src",
				IsDir:   true,
				Children: []*core.FileNode{
					{
						Name:    "components",
						Path:    "/project/src/components",
						RelPath: "src/components",
						IsDir:   true,
						Children: []*core.FileNode{
							{
								Name:    "Button.tsx",
								Path:    "/project/src/components/Button.tsx",
								RelPath: "src/components/Button.tsx",
								IsDir:   false,
							},
							{
								Name:    "Input.tsx",
								Path:    "/project/src/components/Input.tsx",
								RelPath: "src/components/Input.tsx",
								IsDir:   false,
							},
						},
					},
					{
						Name:    "utils",
						Path:    "/project/src/utils",
						RelPath: "src/utils",
						IsDir:   true,
						Children: []*core.FileNode{
							{
								Name:    "helpers.ts",
								Path:    "/project/src/utils/helpers.ts",
								RelPath: "src/utils/helpers.ts",
								IsDir:   false,
							},
						},
					},
				},
			},
			{
				Name:    "docs",
				Path:    "/project/docs",
				RelPath: "docs",
				IsDir:   true,
				Children: []*core.FileNode{
					{
						Name:    "README.md",
						Path:    "/project/docs/README.md",
						RelPath: "docs/README.md",
						IsDir:   false,
					},
				},
			},
		},
	}
}

func TestFileTreeModelCreation(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	if model.root != root {
		t.Error("Root should be set correctly")
	}
	if model.selection != selection {
		t.Error("Selection should be set correctly")
	}
	if model.cursor != 0 {
		t.Error("Cursor should start at 0")
	}
	if !model.expanded[root.Path] {
		t.Error("Root should be expanded by default")
	}
}

func TestViewportBuilding(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Check that viewport is built
	if len(model.viewport) == 0 {
		t.Error("Viewport should be built on creation")
	}

	// Check that nodeMap is populated
	if len(model.nodeMap) == 0 {
		t.Error("NodeMap should be populated")
	}

	// First entry should be the root
	if model.nodeMap[0] != root {
		t.Error("First viewport entry should be root node")
	}
}

func TestKeyboardNavigation(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)
	initialCursor := model.cursor

	// Test down navigation
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	if model.cursor != initialCursor+1 {
		t.Error("Cursor should move down")
	}

	// Test up navigation
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})

	if model.cursor != initialCursor {
		t.Error("Cursor should move back up")
	}
}

func TestToggleExclusion(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Navigate to a specific node (src directory)
	model.cursor = 1 // Assuming src is at index 1

	// Toggle exclusion with space key
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})

	// Check that the node was toggled
	if node, exists := model.nodeMap[1]; exists {
		if selection.IsFileIncluded(node.RelPath) {
			t.Error("Node should be excluded after toggle")
		}
	}

	// Toggle again
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})

	// Check that the node was toggled back
	if node, exists := model.nodeMap[1]; exists {
		if !selection.IsFileIncluded(node.RelPath) {
			t.Error("Node should be included after second toggle")
		}
	}
}

func TestDirectoryExpansion(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Find src directory
	srcIndex := -1
	for i, node := range model.nodeMap {
		if node.RelPath == "src" {
			srcIndex = i
			break
		}
	}

	if srcIndex == -1 {
		t.Fatal("Could not find src directory in viewport")
	}

	model.cursor = srcIndex
	initialExpanded := model.expanded[model.nodeMap[srcIndex].Path]

	// Toggle expansion with enter
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	newExpanded := model.expanded[model.nodeMap[srcIndex].Path]
	if newExpanded == initialExpanded {
		t.Error("Directory expansion should be toggled")
	}
}

func TestStatusIndicators(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Test included file indicator
	buttonNode := &core.FileNode{
		RelPath: "src/components/Button.tsx",
		IsDir:   false,
	}

	indicator := model.getStatusIndicator(buttonNode)
	if indicator != "[ ]" {
		t.Errorf("Included file should show '[ ]', got '%s'", indicator)
	}

	// Exclude the file
	selection.ExcludeFile(buttonNode.RelPath)

	indicator = model.getStatusIndicator(buttonNode)
	if indicator != "[x]" {
		t.Errorf("Excluded file should show '[x]', got '%s'", indicator)
	}

	// Test directory indicators
	dirNode := &core.FileNode{
		RelPath: "src/components",
		IsDir:   true,
	}

	// Directory should show [ ] when not explicitly excluded
	indicator = model.getStatusIndicator(dirNode)
	if indicator != "[ ]" {
		t.Errorf("Included directory should show '[ ]', got '%s'", indicator)
	}

	// Exclude directory
	selection.ExcludeFile(dirNode.RelPath)

	indicator = model.getStatusIndicator(dirNode)
	if indicator != "[x]" {
		t.Errorf("Excluded directory should show '[x]', got '%s'", indicator)
	}
}

func TestFilteredFileIndicators(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Test gitignored file
	gitIgnoredNode := &core.FileNode{
		RelPath:      "src/temp.tmp",
		IsDir:        false,
		IsGitignored: true,
	}

	indicator := model.getStatusIndicator(gitIgnoredNode)
	if indicator != "[~]" {
		t.Errorf("Gitignored file should show '[~]', got '%s'", indicator)
	}

	// Test custom ignored file
	customIgnoredNode := &core.FileNode{
		RelPath:         "src/cache.cache",
		IsDir:           false,
		IsCustomIgnored: true,
	}

	indicator = model.getStatusIndicator(customIgnoredNode)
	if indicator != "[~]" {
		t.Errorf("Custom ignored file should show '[~]', got '%s'", indicator)
	}
}

func TestCacheInvalidation(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Initially cache should be invalid
	if model.statsCacheValid {
		t.Error("Stats cache should start invalid")
	}

	// Get stats to populate cache
	_ = model.getCachedGlobalStats()

	if !model.statsCacheValid {
		t.Error("Stats cache should be valid after calculation")
	}

	// Invalidate cache
	model.invalidateStatsCache()

	if model.statsCacheValid {
		t.Error("Stats cache should be invalid after invalidation")
	}

	if len(model.cachedStats) != 0 {
		t.Error("Cached stats should be cleared after invalidation")
	}
}

func TestResetAllExclusions(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Exclude some files
	selection.ExcludeFile("src/components/Button.tsx")
	selection.ExcludeFile("docs/README.md")

	// Reset with 'r' key
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	// Check that all files are included
	if !selection.IsFileIncluded("src/components/Button.tsx") {
		t.Error("File should be included after reset")
	}
	if !selection.IsFileIncluded("docs/README.md") {
		t.Error("File should be included after reset")
	}
}

func TestExcludeAllFiles(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Exclude all with 'a' key
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Check that files are excluded (but directories might not be, depending on implementation)
	// The excludeAll function excludes individual files, not directories
	if selection.IsFileIncluded("src/components/Button.tsx") {
		t.Error("File should be excluded after exclude all")
	}
}

// Performance benchmarks

func BenchmarkViewRendering(b *testing.B) {
	root := createLargeTestTree(100) // Create tree with 100 files
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func BenchmarkViewRenderingWithExclusions(b *testing.B) {
	root := createLargeTestTree(100)
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Exclude some directories to test hierarchical performance
	selection.ExcludeFile("dir1")
	selection.ExcludeFile("dir2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func BenchmarkCachedStatsVsRecalculation(b *testing.B) {
	root := createLargeTestTree(1000)
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	b.Run("CachedStats", func(b *testing.B) {
		// Populate cache first
		_ = model.getCachedGlobalStats()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.getCachedGlobalStats()
		}
	})

	b.Run("RecalculatedStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Force recalculation by invalidating cache
			model.invalidateStatsCache()
			_ = model.getCachedGlobalStats()
		}
	})
}

func BenchmarkToggleOperations(b *testing.B) {
	root := createLargeTestTree(100)
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Navigate to different nodes and toggle
		model.cursor = i % len(model.nodeMap)
		if node, exists := model.nodeMap[model.cursor]; exists {
			selection.ToggleFile(node.RelPath)
			model.invalidateStatsCache()
		}
	}
}

// Helper function to create a large test tree for performance testing
func createLargeTestTree(numFiles int) *core.FileNode {
	root := &core.FileNode{
		Name:     "project",
		Path:     "/project",
		RelPath:  "",
		IsDir:    true,
		Children: []*core.FileNode{},
	}

	// Create multiple directories with files
	for i := 0; i < numFiles/10; i++ {
		dir := &core.FileNode{
			Name:     fmt.Sprintf("dir%d", i),
			Path:     fmt.Sprintf("/project/dir%d", i),
			RelPath:  fmt.Sprintf("dir%d", i),
			IsDir:    true,
			Children: []*core.FileNode{},
		}

		// Add files to each directory
		for j := 0; j < 10; j++ {
			file := &core.FileNode{
				Name:    fmt.Sprintf("file%d.txt", j),
				Path:    fmt.Sprintf("/project/dir%d/file%d.txt", i, j),
				RelPath: fmt.Sprintf("dir%d/file%d.txt", i, j),
				IsDir:   false,
			}
			dir.Children = append(dir.Children, file)
		}

		root.Children = append(root.Children, dir)
	}

	return root
}

func TestViewContainsExpectedElements(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)
	view := model.View()

	// Check that view contains expected elements
	if !strings.Contains(view, "project") {
		t.Error("View should contain root project name")
	}

	if !strings.Contains(view, "Total:") {
		t.Error("View should contain statistics")
	}

	if !strings.Contains(view, "Space: toggle") {
		t.Error("View should contain help text")
	}

	// Check that status indicators are present
	if !strings.Contains(view, "[ ]") && !strings.Contains(view, "[x]") {
		t.Error("View should contain status indicators")
	}
}

func TestHierarchicalStatusInheritance(t *testing.T) {
	root := createTestFileTree()
	selection := core.NewSelectionState()

	model := NewFileTreeModel(root, selection)

	// Exclude parent directory
	selection.ExcludeFile("src/components")

	// Check that child files show as excluded in the view
	view := model.View()

	// The view should reflect hierarchical exclusion
	// This is more of an integration test to ensure the UI properly shows
	// the hierarchical state calculated by the core selection system

	// Verify the view contains content
	if len(view) == 0 {
		t.Error("View should not be empty")
	}

	if !selection.IsPathExcluded("src/components/Button.tsx") {
		t.Error("Child file should be excluded due to parent exclusion")
	}
}
