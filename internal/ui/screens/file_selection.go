package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	keyEsc = "esc"
)

type RescanRequestMsg struct{}

type FileSelectionModel struct {
	tree       *components.FileTreeModel
	width      int
	height     int
	fileTree   *scanner.FileNode
	selections map[string]bool

	// Filter mode state
	filterMode   bool
	filterBuffer string
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

	if m.filterMode {
		return m.handleFilterMode(msg)
	}

	return m.handleNormalMode(msg, selections)
}

func (m *FileSelectionModel) handleFilterMode(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "enter":
		m.tree.SetFilter(m.filterBuffer)
		m.filterMode = false
	case keyEsc:
		m.filterMode = false
		m.filterBuffer = ""
	case "backspace":
		if len(m.filterBuffer) > 0 {
			m.filterBuffer = m.filterBuffer[:len(m.filterBuffer)-1]
		}
	default:
		if len(msg.String()) == 1 && msg.String() != " " {
			m.filterBuffer += msg.String()
		}
	}

	return nil
}

func (m *FileSelectionModel) handleNormalMode(msg tea.KeyMsg, selections map[string]bool) tea.Cmd {
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
	case "i":
		m.tree.ToggleShowIgnored()
	case "/":
		m.filterMode = true
		m.filterBuffer = m.tree.GetFilter()
	case "f5":
		return func() tea.Msg {
			return RescanRequestMsg{}
		}
	case "ctrl+c":
		m.tree.ClearFilter()
	}

	return nil
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
		estimatedTokens := tokens.EstimateFromBytes(totalSize)
		stats = styles.RenderTokenStats(selectedCount, formatSize(totalSize), tokens.FormatTokens(estimatedTokens))
	} else {
		stats = styles.RenderTokenStats(0, "0 B", "0")
	}

	// Add filter status with styling
	if m.tree != nil && m.tree.GetFilter() != "" {
		filterLabel := styles.StatsLabelStyle.Render("Filter:")
		filterValue := styles.StatsValueStyle.Render(m.tree.GetFilter())
		stats += " │ " + filterLabel + " " + filterValue
	}

	var treeView string
	if m.tree != nil {
		treeView = m.tree.View()
	} else {
		treeView = "Loading file tree..."
	}

	var footer string
	if m.filterMode {
		shortcuts := []string{
			"Type to filter",
			"Enter: Apply",
			"Esc: Cancel",
			"Backspace: Delete",
		}
		footer = styles.RenderFooter(shortcuts)
	} else {
		// Line 1: Navigation and selection
		line1 := []string{
			"↑/↓: Navigate",
			"←/→: Expand/Collapse",
			"Space: Select",
			"i: Ignored",
			"/: Filter",
		}
		// Line 2: Actions and commands
		line2 := []string{
			"F5: Rescan",
			"F7: Back",
			"F8: Next",
			"F1: Help",
			"Ctrl+Q: Quit",
		}
		footer = styles.RenderFooter(line1) + "\n" + styles.RenderFooter(line2)
	}

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")
	content.WriteString(stats)
	content.WriteString("\n")

	// Show filter input if in filter mode
	if m.filterMode {
		content.WriteString(fmt.Sprintf("Filter: %s_", m.filterBuffer))
		content.WriteString("\n")
	}

	content.WriteString("\n")
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
