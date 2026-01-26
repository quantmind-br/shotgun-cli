package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	keyEsc                          = "esc"
	fileSelectionHeaderFooterHeight = 6
)

type RescanRequestMsg struct{}

type FileSelectionModel struct {
	tree       *components.FileTreeModel
	width      int
	height     int
	fileTree   *scanner.FileNode
	selections map[string]bool

	filterMode   bool
	filterBuffer string

	spinner spinner.Model
	loading bool

	maxSizeBytes int64
	maxSizeStr   string
	totalTokens  int
}

func NewFileSelection(fileTree *scanner.FileNode, selections map[string]bool, maxSizeStr string) *FileSelectionModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	if selections == nil {
		selections = make(map[string]bool)
	}

	m := &FileSelectionModel{
		tree:       components.NewFileTree(fileTree, selections),
		fileTree:   fileTree,
		selections: selections,
		spinner:    s,
		loading:    fileTree == nil,
		maxSizeStr: maxSizeStr,
	}

	if maxSizeStr != "" {
		m.maxSizeBytes, _ = parseSize(maxSizeStr)
	}

	return m
}

func (m *FileSelectionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.tree != nil {
		m.tree.SetSize(width, height-fileSelectionHeaderFooterHeight)
	}
}

func (m *FileSelectionModel) Init() tea.Cmd {
	if m.loading {
		return m.spinner.Tick
	}
	return nil
}

func (m *FileSelectionModel) Update(msg tea.Msg) tea.Cmd {
	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return cmd
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || m.tree == nil {
		return nil
	}

	if m.filterMode {
		return m.handleFilterMode(keyMsg)
	}

	return m.handleNormalMode(keyMsg)
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

func (m *FileSelectionModel) handleNormalMode(msg tea.KeyMsg) tea.Cmd {
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
		m.syncSelections()
	case "a":
		m.tree.SelectAllVisible()
		m.syncSelections()
	case "A":
		m.tree.DeselectAllVisible()
		m.syncSelections()
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

func (m *FileSelectionModel) syncSelections() {
	// Clear existing selections
	for k := range m.selections {
		delete(m.selections, k)
	}

	// Copy new selections from tree
	treeSelections := m.tree.GetSelections()
	for k, v := range treeSelections {
		m.selections[k] = v
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

		bar := components.NewUsageBar(totalSize, m.maxSizeBytes, m.maxSizeStr, estimatedTokens, m.width-4)
		stats += "\n" + bar.View()
	} else {
		stats = styles.RenderTokenStats(0, "0 B", "0")
	}

	if m.tree != nil && m.tree.GetFilter() != "" {
		filterLabel := styles.StatsLabelStyle.Render("Filter:")
		filterValue := styles.StatsValueStyle.Render(m.tree.GetFilter())
		visibleCount := m.tree.GetVisibleFileCount()
		totalCount := m.tree.GetTotalFileCount()
		matchCount := styles.StatsValueStyle.Render(fmt.Sprintf("(%d/%d)", visibleCount, totalCount))
		stats += " │ " + filterLabel + " " + filterValue + " " + matchCount
	}

	var treeView string
	if m.loading {
		treeView = m.spinner.View() + " Scanning directory..."
	} else if m.tree != nil {
		treeView = m.tree.View()
	} else {
		treeView = "No files to display"
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
		line1 := []string{
			"↑/↓: Navigate",
			"←/→: Expand/Collapse",
			"Space: Select",
			"a/A: All/None",
			"i: Ignored",
			"/: Filter",
		}
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

func (m *FileSelectionModel) SetFileTree(tree *scanner.FileNode) {
	m.fileTree = tree
	m.tree = components.NewFileTree(tree, m.selections)
	m.loading = false
	if m.tree != nil {
		m.tree.SetSize(m.width, m.height-fileSelectionHeaderFooterHeight)
	}
}

func (m *FileSelectionModel) IsLoading() bool {
	return m.loading
}

// GetFileTree returns the current file tree
func (m *FileSelectionModel) GetFileTree() *scanner.FileNode {
	return m.fileTree
}

// GetSelections returns the current file selections map
func (m *FileSelectionModel) GetSelections() map[string]bool {
	return m.selections
}

// GetSelectedCount returns the number of selected files
func (m *FileSelectionModel) GetSelectedCount() int {
	if m.selections == nil {
		return 0
	}
	return len(m.selections)
}

// SetSelectionsForTest sets the selections map directly (for testing only)
func (m *FileSelectionModel) SetSelectionsForTest(selections map[string]bool) {
	m.selections = selections
	if m.tree != nil {
		m.tree = components.NewFileTree(m.fileTree, m.selections)
	}
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
