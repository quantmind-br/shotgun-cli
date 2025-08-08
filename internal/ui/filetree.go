package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// DirectoryStats holds cached statistics for a directory
type DirectoryStats struct {
	Total    int
	Included int
	Excluded int
	Filtered int
}

// PatternConfigRequestMsg is sent when user requests pattern configuration
type PatternConfigRequestMsg struct{}

// FileTreeModel represents the file tree component with exclusion capabilities
type FileTreeModel struct {
	root            *core.FileNode
	selection       *core.SelectionState
	cursor          int
	expanded        map[string]bool
	viewport        []string                  // Flattened view of the tree
	nodeMap         map[int]*core.FileNode    // Maps viewport index to node
	cachedStats     map[string]DirectoryStats // NEW: Cache stats by path
	statsCacheValid bool                      // NEW: Cache validity flag
}

// NewFileTreeModel creates a new file tree model
func NewFileTreeModel(root *core.FileNode, selection *core.SelectionState) FileTreeModel {
	m := FileTreeModel{
		root:            root,
		selection:       selection,
		cursor:          0,
		expanded:        make(map[string]bool),
		nodeMap:         make(map[int]*core.FileNode),
		cachedStats:     make(map[string]DirectoryStats),
		statsCacheValid: false,
	}

	// Expand root by default
	if root != nil {
		m.expanded[root.Path] = true
	}

	m.rebuildViewport()
	return m
}

// invalidateStatsCache invalidates the statistics cache
func (m *FileTreeModel) invalidateStatsCache() {
	m.statsCacheValid = false
	m.cachedStats = make(map[string]DirectoryStats)
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
				// Invalidate cache since selection changed
				m.invalidateStatsCache()
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
			m.invalidateStatsCache()
		case "a":
			// Exclude all files
			m.excludeAll()
			m.invalidateStatsCache()
		case "A":
			// Include all files (clear exclusions)
			m.selection.Reset()
			m.invalidateStatsCache()
		case "p", "P":
			// Open pattern configuration - this will be handled by the main app
			return m, tea.Cmd(func() tea.Msg {
				return PatternConfigRequestMsg{}
			})
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

	// Define styles locally
	statusStyleLocal := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")).Bold(true)
	helpStyleLocal := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

	lines = append(lines, statusStyleLocal.Render(stats))
	lines = append(lines, "")
	lines = append(lines, helpStyleLocal.Render("Space: toggle | hjkl: navigate | Enter: expand | r: reset | a/A: exclude/include all | p: patterns | F5: continue | o: options"))

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
		// Use hierarchical logic instead of recursive stats
		if m.selection.IsPathExcluded(node.RelPath) {
			return "[x]" // Directory explicitly excluded
		} else {
			return "[ ]" // Directory included (hierarchical system handles children)
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

// excludeAll excludes all files in the tree
func (m *FileTreeModel) excludeAll() {
	if m.root != nil {
		m.excludeAllRecursive(m.root)
		m.invalidateStatsCache()
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

	stats := m.getCachedGlobalStats()

	return lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#17a2b8")).Render("Total: "),
		lipgloss.NewStyle().Bold(true).Render(lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Render(fmt.Sprintf("%d", stats.Total)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(" | "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#28a745")).Render(fmt.Sprintf("Included: %d", stats.Included)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(" | "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#dc3545")).Render(fmt.Sprintf("Excluded: %d", stats.Excluded)),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(" | "),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6c757d")).Render(fmt.Sprintf("Filtered: %d", stats.Filtered)),
		)),
	)
}

// getCachedGlobalStats returns cached global statistics, calculating if needed
func (m *FileTreeModel) getCachedGlobalStats() DirectoryStats {
	const globalKey = "__global__"

	if m.statsCacheValid {
		if stats, exists := m.cachedStats[globalKey]; exists {
			return stats
		}
	}

	// Calculate fresh statistics
	stats := m.calculateStatsRecursive(m.root)

	// Cache the result
	if m.cachedStats == nil {
		m.cachedStats = make(map[string]DirectoryStats)
	}
	m.cachedStats[globalKey] = stats
	m.statsCacheValid = true

	return stats
}

// calculateStatsRecursive calculates file statistics (only called when cache invalid)
func (m *FileTreeModel) calculateStatsRecursive(node *core.FileNode) DirectoryStats {
	stats := DirectoryStats{}

	if node.IsDir {
		for _, child := range node.Children {
			childStats := m.calculateStatsRecursive(child)
			stats.Total += childStats.Total
			stats.Included += childStats.Included
			stats.Excluded += childStats.Excluded
			stats.Filtered += childStats.Filtered
		}
	} else {
		stats.Total++
		if node.IsGitignored || node.IsCustomIgnored {
			stats.Filtered++
		} else if m.selection.IsFileIncluded(node.RelPath) {
			stats.Included++
		} else {
			stats.Excluded++
		}
	}

	return stats
}
