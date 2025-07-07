package ui

import (
	"fmt"
	"strings"

	"shotgun-cli/internal/file"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FileTreeModel struct {
	root         *file.FileNode
	cursor       int
	flatList     []*file.FileNode
	expandedDirs map[string]bool
	width        int
	height       int
}

func NewFileTreeModel(root *file.FileNode) *FileTreeModel {
	model := &FileTreeModel{
		root:         root,
		cursor:       0,
		expandedDirs: make(map[string]bool),
	}
	
	model.expandedDirs[root.Path] = true
	model.rebuildFlatList()
	
	return model
}

func (m *FileTreeModel) Init() tea.Cmd {
	return nil
}

func (m *FileTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.flatList)-1 {
				m.cursor++
			}
		case " ":
			if m.cursor < len(m.flatList) {
				current := m.flatList[m.cursor]
				current.IsSelected = !current.IsSelected
				
				if current.IsDir {
					m.toggleDirSelection(current, current.IsSelected)
				}
			}
		case "enter":
			if m.cursor < len(m.flatList) {
				current := m.flatList[m.cursor]
				if current.IsDir {
					m.expandedDirs[current.Path] = !m.expandedDirs[current.Path]
					m.rebuildFlatList()
				}
			}
		}
	}
	return m, nil
}

func (m *FileTreeModel) toggleDirSelection(dir *file.FileNode, selected bool) {
	for _, child := range dir.Children {
		child.IsSelected = selected
		if child.IsDir {
			m.toggleDirSelection(child, selected)
		}
	}
}

func (m *FileTreeModel) rebuildFlatList() {
	m.flatList = []*file.FileNode{}
	m.addToFlatList(m.root, 0)
}

func (m *FileTreeModel) addToFlatList(node *file.FileNode, depth int) {
	if node.Path != "." {
		m.flatList = append(m.flatList, node)
	}
	
	if node.IsDir && m.expandedDirs[node.Path] {
		for _, child := range node.Children {
			m.addToFlatList(child, depth+1)
		}
	}
}

func (m *FileTreeModel) View() string {
	if len(m.flatList) == 0 {
		return "No files to display"
	}
	
	var lines []string
	
	for i, node := range m.flatList {
		prefix := strings.Repeat("  ", m.getDepth(node))
		
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}
		
		selection := "✗"  // Padrão é desselecionado (mostra o X)
		if node.IsSelected {
			selection = "✓"  // Selecionado mostra check
		}
		
		expansion := " "
		if node.IsDir {
			if m.expandedDirs[node.Path] {
				expansion = "▼"
			} else {
				expansion = "▶"
			}
		}
		
		name := node.Name
		if node.IsDir {
			name = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render(name + "/")
		}
		
		line := fmt.Sprintf("%s %s %s %s%s", cursor, selection, expansion, prefix, name)
		
		if i == m.cursor {
			line = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(line)
		}
		
		lines = append(lines, line)
	}
	
	content := strings.Join(lines, "\n")
	
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Height(m.height - 4)
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Use ↑/↓ to navigate, Space to deselect/select, Enter to expand/collapse, Alt+D to continue")
	
	return fmt.Sprintf("%s\n\n%s", style.Render(content), instructions)
}

func (m *FileTreeModel) getDepth(node *file.FileNode) int {
	return strings.Count(node.Path, "/")
}

func (m *FileTreeModel) GetSelectedFiles() []*file.FileNode {
	var selected []*file.FileNode
	m.collectSelected(m.root, &selected)
	return selected
}

func (m *FileTreeModel) collectSelected(node *file.FileNode, selected *[]*file.FileNode) {
	if node.IsSelected && !node.IsDir {
		*selected = append(*selected, node)
	}
	
	for _, child := range node.Children {
		m.collectSelected(child, selected)
	}
}