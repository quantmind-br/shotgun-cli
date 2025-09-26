package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo464/shotgun-cli/internal/core/scanner"
	"github.com/diogo464/shotgun-cli/internal/ui/components"
	"github.com/diogo464/shotgun-cli/internal/ui/styles"
)

type FileSelectionModel struct {
	tree      *components.FileTreeModel
	width     int
	height    int
	fileTree  *scanner.FileNode
	selections map[string]bool
}

func NewFileSelection(fileTree *scanner.FileNode, selections map[string]bool) *FileSelectionModel {
	return &FileSelectionModel{
		tree:       components.NewFileTree(fileTree, selections),
		fileTree:   fileTree,
		selections: selections,
	}
}

func (m *FileSelectionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.tree != nil {
		m.tree.SetSize(width, height-6) // Reserve space for header and footer
	}
}

func (m *FileSelectionModel) Update(msg tea.KeyMsg, selections map[string]bool) tea.Cmd {
	if m.tree == nil {
		return nil
	}

	var cmd tea.Cmd

	switch msg.String() {
	case "up", "k":
		m.tree.MoveUp()
	case "down", "j":
		m.tree.MoveDown()
	case "left", "h":
		m.tree.CollapseNode()
	case "right", "l":
		m.tree.ExpandNode()
	case " ":
		m.tree.ToggleSelection()
		m.updateSelections(selections)
	case "d":
		m.tree.ToggleDirectorySelection()
		m.updateSelections(selections)
	case "i":
		m.tree.ToggleShowIgnored()
	case "/":
		// TODO: Implement filter mode
	case "f5":
		// TODO: Implement rescan
	default:
		// Handle other keys if needed
	}

	return cmd
}

func (m *FileSelectionModel) updateSelections(selections map[string]bool) {
	// Clear existing selections
	for k := range selections {
		delete(selections, k)
	}

	// Copy new selections from tree
	treeSelections := m.tree.GetSelections()
	for k, v := range treeSelections {
		selections[k] = v
	}
}

func (m *FileSelectionModel) View() string {
	header := styles.RenderHeader(1, "Select Files")

	var stats string
	if m.selections != nil {
		selectedCount := len(m.selections)
		totalSize := m.calculateSelectedSize()
		stats = fmt.Sprintf("Selected: %d files (%s)", selectedCount, formatSize(totalSize))
	} else {
		stats = "Selected: 0 files (0 B)"
	}

	var treeView string
	if m.tree != nil {
		treeView = m.tree.View()
	} else {
		treeView = "Loading file tree..."
	}

	shortcuts := []string{
		"↑/↓: Navigate",
		"←/→: Expand/Collapse",
		"Space: Select File",
		"d: Select Directory",
		"i: Toggle Ignored",
		"F8: Next",
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(shortcuts)

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")
	content.WriteString(stats)
	content.WriteString("\n\n")
	content.WriteString(treeView)
	content.WriteString("\n")
	content.WriteString(footer)

	return content.String()
}

func (m *FileSelectionModel) calculateSelectedSize() int64 {
	if m.fileTree == nil || m.selections == nil {
		return 0
	}

	var totalSize int64
	m.walkTree(m.fileTree, func(node *scanner.FileNode, path string) {
		if !node.IsDir && m.selections[path] {
			totalSize += node.Size
		}
	})

	return totalSize
}

func (m *FileSelectionModel) walkTree(node *scanner.FileNode, fn func(*scanner.FileNode, string)) {
	fn(node, node.Path)
	for _, child := range node.Children {
		m.walkTree(child, fn)
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}