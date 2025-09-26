package components

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type FileTreeModel struct {
	tree        *scanner.FileNode
	cursor      int
	selections  map[string]bool
	showIgnored bool
	filter      string
	expanded    map[string]bool
	width       int
	height      int
	visibleItems []treeItem
	topIndex    int
}

type treeItem struct {
	node     *scanner.FileNode
	path     string
	depth    int
	isLast   bool
	hasNext  []bool
}

func NewFileTree(tree *scanner.FileNode, selections map[string]bool) *FileTreeModel {
	expanded := make(map[string]bool)
	if tree != nil {
		expanded[tree.Path] = true // Root is always expanded
	}

	model := &FileTreeModel{
		tree:        tree,
		selections:  make(map[string]bool),
		expanded:    expanded,
		showIgnored: false,
		filter:      "",
	}

	// Copy selections
	for k, v := range selections {
		model.selections[k] = v
	}

	model.rebuildVisibleItems()
	return model
}

func (m *FileTreeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *FileTreeModel) MoveUp() {
	if m.cursor > 0 {
		m.cursor--
		m.adjustScroll()
	}
}

func (m *FileTreeModel) MoveDown() {
	if m.cursor < len(m.visibleItems)-1 {
		m.cursor++
		m.adjustScroll()
	}
}

func (m *FileTreeModel) ExpandNode() {
	if m.cursor < len(m.visibleItems) {
		item := m.visibleItems[m.cursor]
		if item.node.IsDir {
			m.expanded[item.path] = true
			m.rebuildVisibleItems()
		}
	}
}

func (m *FileTreeModel) CollapseNode() {
	if m.cursor < len(m.visibleItems) {
		item := m.visibleItems[m.cursor]
		if item.node.IsDir && m.expanded[item.path] {
			m.expanded[item.path] = false
			m.rebuildVisibleItems()
		}
	}
}

func (m *FileTreeModel) ToggleSelection() {
	if m.cursor < len(m.visibleItems) {
		item := m.visibleItems[m.cursor]
		if !item.node.IsDir {
			m.selections[item.path] = !m.selections[item.path]
		}
	}
}

func (m *FileTreeModel) ToggleDirectorySelection() {
	if m.cursor < len(m.visibleItems) {
		item := m.visibleItems[m.cursor]
		if item.node.IsDir {
			// Toggle all files in directory
			allSelected := m.areAllFilesInDirSelected(item.node)
			m.setDirectorySelection(item.node, !allSelected)
		}
	}
}

func (m *FileTreeModel) ToggleShowIgnored() {
	m.showIgnored = !m.showIgnored
	m.rebuildVisibleItems()
}

func (m *FileTreeModel) SetFilter(filter string) {
	m.filter = filter
	m.rebuildVisibleItems()
}

func (m *FileTreeModel) GetFilter() string {
	return m.filter
}

func (m *FileTreeModel) ClearFilter() {
	m.filter = ""
	m.rebuildVisibleItems()
}

func (m *FileTreeModel) GetSelections() map[string]bool {
	return m.selections
}

func (m *FileTreeModel) View() string {
	if m.tree == nil {
		return "No files to display"
	}

	var content strings.Builder

	// Calculate visible range
	maxVisible := m.height
	if maxVisible <= 0 {
		maxVisible = 20 // Default height
	}

	start := m.topIndex
	end := start + maxVisible
	if end > len(m.visibleItems) {
		end = len(m.visibleItems)
	}

	for i := start; i < end; i++ {
		item := m.visibleItems[i]
		line := m.renderTreeItem(item, i == m.cursor)
		content.WriteString(line)
		if i < end-1 {
			content.WriteString("\n")
		}
	}

	return content.String()
}

func (m *FileTreeModel) renderTreeItem(item treeItem, isCursor bool) string {
	var prefix strings.Builder

	// Build tree structure prefix
	for d := 0; d < item.depth; d++ {
		if d < len(item.hasNext) && item.hasNext[d] {
			prefix.WriteString("│  ")
		} else {
			prefix.WriteString("   ")
		}
	}

	// Add tree connector
	if item.depth > 0 {
		if item.isLast {
			prefix.WriteString("└──")
		} else {
			prefix.WriteString("├──")
		}
	}

	// Selection checkbox (for files only)
	var checkbox string
	if !item.node.IsDir {
		if m.selections[item.path] {
			checkbox = "[✓] "
		} else {
			checkbox = "[ ] "
		}
	}

	// Directory indicator
	var dirIndicator string
	if item.node.IsDir {
		if m.expanded[item.path] {
			dirIndicator = "📂 "
		} else {
			dirIndicator = "📁 "
		}
	}

	// File name
	name := filepath.Base(item.path)
	if item.node.IsDir {
		name += "/"
	}

	// Ignore status
	var ignoreStatus string
	if item.node.IgnoreReason != "" {
		if strings.Contains(item.node.IgnoreReason, "gitignore") {
			ignoreStatus = " (g)"
		} else {
			ignoreStatus = " (c)"
		}
	}

	// File size (for files only)
	var sizeInfo string
	if !item.node.IsDir && item.node.Size > 0 {
		sizeInfo = fmt.Sprintf(" (%s)", formatFileSize(item.node.Size))
	}

	// Combine all parts
	line := prefix.String() + checkbox + dirIndicator + name + ignoreStatus + sizeInfo

	// Apply cursor highlighting
	if isCursor {
		line = styles.SelectedStyle.Render(line)
	}

	return line
}

func (m *FileTreeModel) rebuildVisibleItems() {
	m.visibleItems = nil
	if m.tree != nil {
		m.buildVisibleItems(m.tree, "", 0, true, nil)
	}

	// Adjust cursor if it's out of bounds
	if m.cursor >= len(m.visibleItems) {
		m.cursor = len(m.visibleItems) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *FileTreeModel) buildVisibleItems(node *scanner.FileNode, path string, depth int, isLast bool, hasNext []bool) {
	// Skip if filtered out
	if !m.shouldShowNode(node) {
		return
	}

	// Add current node
	currentHasNext := make([]bool, len(hasNext))
	copy(currentHasNext, hasNext)

	item := treeItem{
		node:     node,
		path:     node.Path,
		depth:    depth,
		isLast:   isLast,
		hasNext:  currentHasNext,
	}
	m.visibleItems = append(m.visibleItems, item)

	// Add children if expanded
	if node.IsDir && m.expanded[node.Path] && len(node.Children) > 0 {
		// Sort children: directories first, then files
		children := make([]*scanner.FileNode, len(node.Children))
		copy(children, node.Children)
		sort.Slice(children, func(i, j int) bool {
			if children[i].IsDir != children[j].IsDir {
				return children[i].IsDir
			}
			return children[i].Name < children[j].Name
		})

		for i, child := range children {
			childIsLast := i == len(children)-1
			childHasNext := append(currentHasNext, !isLast)
			m.buildVisibleItems(child, child.Path, depth+1, childIsLast, childHasNext)
		}
	}
}

func (m *FileTreeModel) shouldShowNode(node *scanner.FileNode) bool {
	// Check ignore status
	if !m.showIgnored && node.IgnoreReason != "" {
		return false
	}

	// Check filter
	if m.filter != "" && !strings.Contains(strings.ToLower(node.Name), strings.ToLower(m.filter)) {
		return false
	}

	return true
}

func (m *FileTreeModel) adjustScroll() {
	if m.height <= 0 {
		return
	}

	// Scroll up if cursor is above visible area
	if m.cursor < m.topIndex {
		m.topIndex = m.cursor
	}

	// Scroll down if cursor is below visible area
	if m.cursor >= m.topIndex+m.height {
		m.topIndex = m.cursor - m.height + 1
	}

	// Ensure topIndex is within bounds
	if m.topIndex < 0 {
		m.topIndex = 0
	}
	maxTop := len(m.visibleItems) - m.height
	if maxTop < 0 {
		maxTop = 0
	}
	if m.topIndex > maxTop {
		m.topIndex = maxTop
	}
}

func (m *FileTreeModel) areAllFilesInDirSelected(dir *scanner.FileNode) bool {
	hasFiles := false
	allSelected := true

	m.walkNode(dir, func(node *scanner.FileNode) {
		if !node.IsDir {
			hasFiles = true
			if !m.selections[node.Path] {
				allSelected = false
			}
		}
	})

	return hasFiles && allSelected
}

func (m *FileTreeModel) setDirectorySelection(dir *scanner.FileNode, selected bool) {
	m.walkNode(dir, func(node *scanner.FileNode) {
		if !node.IsDir {
			if selected {
				m.selections[node.Path] = true
			} else {
				delete(m.selections, node.Path)
			}
		}
	})
}

func (m *FileTreeModel) walkNode(node *scanner.FileNode, fn func(*scanner.FileNode)) {
	if !m.shouldShowNode(node) {
		return
	}

	fn(node)
	for _, child := range node.Children {
		m.walkNode(child, fn)
	}
}

func formatFileSize(bytes int64) string {
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