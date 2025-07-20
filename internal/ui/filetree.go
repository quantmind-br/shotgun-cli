package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// FileTreeModel represents the file tree component with exclusion capabilities
type FileTreeModel struct {
	root      *core.FileNode
	selection *core.SelectionState
	cursor    int
	expanded  map[string]bool
	viewport  []string               // Flattened view of the tree
	nodeMap   map[int]*core.FileNode // Maps viewport index to node
}

// NewFileTreeModel creates a new file tree model
func NewFileTreeModel(root *core.FileNode, selection *core.SelectionState) FileTreeModel {
	m := FileTreeModel{
		root:      root,
		selection: selection,
		cursor:    0,
		expanded:  make(map[string]bool),
		nodeMap:   make(map[int]*core.FileNode),
	}

	// Expand root by default
	if root != nil {
		m.expanded[root.Path] = true
	}

	m.rebuildViewport()
	return m
}

// Update handles file tree events
func (m FileTreeModel) Update(msg tea.Msg) (FileTreeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.viewport)-1 {
				m.cursor++
			}
		case "left", "h":
			// Collapse current node if it's a directory
			if node, exists := m.nodeMap[m.cursor]; exists && node.IsDir {
				m.expanded[node.Path] = false
				m.rebuildViewport()
			}
		case "right", "l":
			// Expand current node if it's a directory
			if node, exists := m.nodeMap[m.cursor]; exists && node.IsDir {
				m.expanded[node.Path] = true
				m.rebuildViewport()
			}
		case " ": // Space to toggle exclusion
			if node, exists := m.nodeMap[m.cursor]; exists {
				m.selection.ToggleFile(node.RelPath)
				// If it's a directory, toggle all children
				if node.IsDir {
					m.toggleDirectoryChildren(node)
				}
			}
		case "enter":
			// Toggle expansion for directories
			if node, exists := m.nodeMap[m.cursor]; exists && node.IsDir {
				m.expanded[node.Path] = !m.expanded[node.Path]
				m.rebuildViewport()
			}
		case "r":
			// Reset all exclusions
			m.selection.Reset()
		case "a":
			// Exclude all files
			m.excludeAll()
		case "A":
			// Include all files (clear exclusions)
			m.selection.Reset()
		}
	}

	return m, nil
}

// View renders the file tree
func (m FileTreeModel) View() string {
	if m.root == nil {
		return "No directory selected"
	}

	var lines []string

	for i, line := range m.viewport {
		style := lipgloss.NewStyle()

		// Highlight current cursor position
		if i == m.cursor {
			style = style.Background(lipgloss.Color("#343a40"))
		}

		// Get the node for this line
		if node, exists := m.nodeMap[i]; exists {
			// Determine the status indicator
			indicator := m.getStatusIndicator(node)
			line = indicator + " " + line

			// Color based on status
			if node.IsGitignored || node.IsCustomIgnored {
				style = style.Foreground(lipgloss.Color("#6c757d")) // Gray for filtered
			} else if m.selection.IsFileIncluded(node.RelPath) {
				style = style.Foreground(lipgloss.Color("#28a745")) // Green for included
			} else {
				style = style.Foreground(lipgloss.Color("#dc3545")) // Red for excluded
			}
		}

		lines = append(lines, style.Render(line))
	}

	// Add statistics
	stats := m.getStats()
	lines = append(lines, "")
	lines = append(lines, statusStyle.Render(stats))
	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("Space: toggle | hjkl: navigate | Enter: expand | r: reset | a/A: exclude/include all | c: continue"))

	return strings.Join(lines, "\n")
}

// rebuildViewport rebuilds the flattened viewport of the tree
func (m *FileTreeModel) rebuildViewport() {
	m.viewport = []string{}
	m.nodeMap = make(map[int]*core.FileNode)

	if m.root != nil {
		m.buildViewportRecursive(m.root, "", 0, true)
	}
}

// buildViewportRecursive builds the viewport recursively
func (m *FileTreeModel) buildViewportRecursive(node *core.FileNode, prefix string, depth int, isLast bool) {
	// Add current node to viewport
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	if depth == 0 {
		connector = ""
	}

	line := prefix + connector + node.Name
	if node.IsDir {
		if m.expanded[node.Path] {
			line += "/"
		} else {
			line += "/ (collapsed)"
		}
	}

	index := len(m.viewport)
	m.viewport = append(m.viewport, line)
	m.nodeMap[index] = node

	// Add children if directory is expanded
	if node.IsDir && m.expanded[node.Path] {
		for i, child := range node.Children {
			isChildLast := i == len(node.Children)-1
			newPrefix := prefix
			if depth > 0 {
				if isLast {
					newPrefix += "    "
				} else {
					newPrefix += "│   "
				}
			}
			m.buildViewportRecursive(child, newPrefix, depth+1, isChildLast)
		}
	}
}

// getStatusIndicator returns the status indicator for a node
func (m *FileTreeModel) getStatusIndicator(node *core.FileNode) string {
	if node.IsGitignored || node.IsCustomIgnored {
		return "[~]" // Filtered by patterns
	}

	if node.IsDir {
		// For directories, check if any children are included/excluded
		included, excluded := m.getDirectoryStats(node)
		if included > 0 && excluded > 0 {
			return "[-]" // Partially excluded
		} else if excluded > 0 {
			return "[x]" // Fully excluded
		} else {
			return "[ ]" // Fully included
		}
	} else {
		// For files, check individual status
		if m.selection.IsFileIncluded(node.RelPath) {
			return "[ ]" // Included
		} else {
			return "[x]" // Excluded
		}
	}
}

// getDirectoryStats gets inclusion/exclusion stats for a directory
func (m *FileTreeModel) getDirectoryStats(dir *core.FileNode) (included, excluded int) {
	if !dir.IsDir {
		return 0, 0
	}

	for _, child := range dir.Children {
		if child.IsDir {
			childIncluded, childExcluded := m.getDirectoryStats(child)
			included += childIncluded
			excluded += childExcluded
		} else {
			if !child.IsGitignored && !child.IsCustomIgnored {
				if m.selection.IsFileIncluded(child.RelPath) {
					included++
				} else {
					excluded++
				}
			}
		}
	}

	return included, excluded
}

// toggleDirectoryChildren toggles exclusion for all children of a directory
func (m *FileTreeModel) toggleDirectoryChildren(dir *core.FileNode) {
	if !dir.IsDir {
		return
	}

	// Determine if we should exclude or include based on current state
	shouldExclude := m.selection.IsFileIncluded(dir.RelPath)

	m.toggleDirectoryChildrenRecursive(dir, shouldExclude)
}

// toggleDirectoryChildrenRecursive recursively toggles children
func (m *FileTreeModel) toggleDirectoryChildrenRecursive(node *core.FileNode, shouldExclude bool) {
	for _, child := range node.Children {
		if child.IsDir {
			m.toggleDirectoryChildrenRecursive(child, shouldExclude)
		} else {
			if !child.IsGitignored && !child.IsCustomIgnored {
				if shouldExclude {
					m.selection.ExcludeFile(child.RelPath)
				} else {
					m.selection.IncludeFile(child.RelPath)
				}
			}
		}
	}
}

// excludeAll excludes all files in the tree
func (m *FileTreeModel) excludeAll() {
	if m.root != nil {
		m.excludeAllRecursive(m.root)
	}
}

// excludeAllRecursive recursively excludes all files
func (m *FileTreeModel) excludeAllRecursive(node *core.FileNode) {
	if node.IsDir {
		for _, child := range node.Children {
			m.excludeAllRecursive(child)
		}
	} else {
		if !node.IsGitignored && !node.IsCustomIgnored {
			m.selection.ExcludeFile(node.RelPath)
		}
	}
}

// getStats returns statistics about included/excluded files
func (m *FileTreeModel) getStats() string {
	if m.root == nil {
		return "No files"
	}

	total, included, excluded, filtered := m.getStatsRecursive(m.root)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#17a2b8")).Render("Total: "),
		lipgloss.NewStyle().Bold(true).Render(lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Render(fmt.Sprintf("%d", total)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(" | "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#28a745")).Render(fmt.Sprintf("Included: %d", included)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(" | "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#dc3545")).Render(fmt.Sprintf("Excluded: %d", excluded)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(" | "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(fmt.Sprintf("Filtered: %d", filtered)),
		)),
	)
}

// getStatsRecursive recursively counts file statistics
func (m *FileTreeModel) getStatsRecursive(node *core.FileNode) (total, included, excluded, filtered int) {
	if node.IsDir {
		for _, child := range node.Children {
			childTotal, childIncluded, childExcluded, childFiltered := m.getStatsRecursive(child)
			total += childTotal
			included += childIncluded
			excluded += childExcluded
			filtered += childFiltered
		}
	} else {
		total++
		if node.IsGitignored || node.IsCustomIgnored {
			filtered++
		} else if m.selection.IsFileIncluded(node.RelPath) {
			included++
		} else {
			excluded++
		}
	}

	return total, included, excluded, filtered
}
