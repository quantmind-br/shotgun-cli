package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gitignore "github.com/sabhiram/go-gitignore"

	"shotgun-cli/internal/core"
)

// PatternConfigModel represents the pattern configuration UI component
type PatternConfigModel struct {
	// Configuration
	config *core.EnhancedConfig

	// UI State
	currentTab      int             // 0: Custom Ignore, 1: Force Include
	customPatterns  list.Model      // List of custom ignore patterns
	forcePatterns   list.Model      // List of force include patterns
	addInput        textinput.Model // Input for adding new patterns
	editIndex       int             // Index being edited (-1 if none)
	editInput       textinput.Model // Input for editing patterns
	isAddingPattern bool            // Whether we're in add mode

	// Validation
	validationEnabled bool
	validationError   string

	// Dimensions
	width  int
	height int
}

// PatternItem represents a pattern list item
type PatternItem struct {
	pattern string
	valid   bool
	error   string
}

// FilterValue implements list.Item interface
func (p PatternItem) FilterValue() string { return p.pattern }

// Title implements list.Item interface
func (p PatternItem) Title() string { return p.pattern }

// Description implements list.Item interface
func (p PatternItem) Description() string {
	if !p.valid && p.error != "" {
		return "❌ " + p.error
	}
	return "✅ Valid pattern"
}

// NewPatternConfigModel creates a new pattern configuration model
func NewPatternConfigModel(config *core.EnhancedConfig, width, height int) *PatternConfigModel {
	// Create text inputs
	addInput := textinput.New()
	addInput.Placeholder = "Enter new pattern (e.g., *.tmp, temp/, *.log)"
	addInput.CharLimit = 256

	editInput := textinput.New()
	editInput.CharLimit = 256

	// Create lists
	customList := list.New([]list.Item{}, newPatternDelegate(), 0, 0)
	customList.Title = "🚫 Custom Ignore Patterns"
	customList.SetShowHelp(false)
	customList.SetShowStatusBar(false)
	customList.SetFilteringEnabled(false)

	forceList := list.New([]list.Item{}, newPatternDelegate(), 0, 0)
	forceList.Title = "✅ Force Include Patterns"
	forceList.SetShowHelp(false)
	forceList.SetShowStatusBar(false)
	forceList.SetFilteringEnabled(false)

	m := &PatternConfigModel{
		config:            config.Clone(),
		currentTab:        0,
		customPatterns:    customList,
		forcePatterns:     forceList,
		addInput:          addInput,
		editInput:         editInput,
		editIndex:         -1,
		validationEnabled: config.App.PatternValidationEnabled,
		width:             width,
		height:            height,
	}

	// Load existing patterns
	m.loadPatterns()
	m.updateSizes()

	return m
}

// loadPatterns loads existing patterns into the lists
func (m *PatternConfigModel) loadPatterns() {
	// Load custom ignore patterns
	customItems := make([]list.Item, 0, len(m.config.App.CustomIgnorePatterns))
	for _, pattern := range m.config.App.CustomIgnorePatterns {
		item := PatternItem{
			pattern: pattern,
			valid:   true,
		}
		if m.validationEnabled {
			item.valid, item.error = m.validatePattern(pattern)
		}
		customItems = append(customItems, item)
	}
	m.customPatterns.SetItems(customItems)

	// Load force include patterns
	forceItems := make([]list.Item, 0, len(m.config.App.ForceIncludePatterns))
	for _, pattern := range m.config.App.ForceIncludePatterns {
		item := PatternItem{
			pattern: pattern,
			valid:   true,
		}
		if m.validationEnabled {
			item.valid, item.error = m.validatePattern(pattern)
		}
		forceItems = append(forceItems, item)
	}
	m.forcePatterns.SetItems(forceItems)
}

// validatePattern validates a gitignore pattern
func (m *PatternConfigModel) validatePattern(pattern string) (bool, string) {
	if pattern == "" {
		return false, "Pattern cannot be empty"
	}

	// Test pattern by creating a temporary gitignore
	testIgnore := gitignore.CompileIgnoreLines(pattern)
	if testIgnore == nil {
		return false, "Invalid gitignore pattern syntax"
	}

	return true, ""
}

// updateSizes updates the sizes of UI components
func (m *PatternConfigModel) updateSizes() {
	listHeight := (m.height - 8) / 2 // Account for tabs, input, and padding
	listWidth := m.width - 4

	m.customPatterns.SetSize(listWidth, listHeight)
	m.forcePatterns.SetSize(listWidth, listHeight)

	m.addInput.Width = listWidth - 20
	m.editInput.Width = listWidth - 20
}

// Update handles input events
func (m PatternConfigModel) Update(msg tea.Msg) (PatternConfigModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle editing mode
		if m.editIndex != -1 {
			switch msg.String() {
			case "enter":
				m.confirmEdit()
				return m, nil
			case "esc":
				m.cancelEdit()
				return m, nil
			default:
				var cmd tea.Cmd
				m.editInput, cmd = m.editInput.Update(msg)
				return m, cmd
			}
		}

		// Handle adding mode
		if m.isAddingPattern {
			switch msg.String() {
			case "enter":
				m.addPattern()
				return m, nil
			case "esc":
				m.cancelAdd()
				return m, nil
			default:
				var cmd tea.Cmd
				m.addInput, cmd = m.addInput.Update(msg)
				return m, cmd
			}
		}

		// Handle normal mode
		switch msg.String() {
		case "tab", "shift+tab":
			if msg.String() == "tab" {
				m.currentTab = (m.currentTab + 1) % 2
			} else {
				m.currentTab = (m.currentTab - 1 + 2) % 2
			}
			return m, nil

		case "a":
			m.startAdd()
			return m, nil

		case "d":
			m.deleteSelected()
			return m, nil

		case "e":
			m.startEdit()
			return m, nil

		case "v":
			m.validationEnabled = !m.validationEnabled
			m.config.App.PatternValidationEnabled = m.validationEnabled
			m.loadPatterns() // Refresh validation status
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()
		return m, nil
	}

	// Update active list
	if m.currentTab == 0 {
		var cmd tea.Cmd
		m.customPatterns, cmd = m.customPatterns.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.forcePatterns, cmd = m.forcePatterns.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// startAdd begins adding a new pattern
func (m *PatternConfigModel) startAdd() {
	m.isAddingPattern = true
	m.addInput.SetValue("")
	m.addInput.Focus()
	m.validationError = ""
}

// cancelAdd cancels adding a pattern
func (m *PatternConfigModel) cancelAdd() {
	m.isAddingPattern = false
	m.addInput.Blur()
}

// addPattern adds a new pattern to the current list
func (m *PatternConfigModel) addPattern() {
	pattern := strings.TrimSpace(m.addInput.Value())
	if pattern == "" {
		return
	}

	// Validate pattern if enabled
	if m.validationEnabled {
		valid, err := m.validatePattern(pattern)
		if !valid {
			m.validationError = err
			return
		}
	}

	// Add to appropriate list and config
	if m.currentTab == 0 {
		// Custom ignore patterns
		m.config.App.CustomIgnorePatterns = append(m.config.App.CustomIgnorePatterns, pattern)
	} else {
		// Force include patterns
		m.config.App.ForceIncludePatterns = append(m.config.App.ForceIncludePatterns, pattern)
	}

	m.loadPatterns()
	m.cancelAdd()
}

// startEdit begins editing the selected pattern
func (m *PatternConfigModel) startEdit() {
	var selectedIndex int
	var pattern string

	if m.currentTab == 0 {
		selectedIndex = m.customPatterns.Index()
		if selectedIndex >= 0 && selectedIndex < len(m.config.App.CustomIgnorePatterns) {
			pattern = m.config.App.CustomIgnorePatterns[selectedIndex]
		}
	} else {
		selectedIndex = m.forcePatterns.Index()
		if selectedIndex >= 0 && selectedIndex < len(m.config.App.ForceIncludePatterns) {
			pattern = m.config.App.ForceIncludePatterns[selectedIndex]
		}
	}

	if pattern != "" {
		m.editIndex = selectedIndex
		m.editInput.SetValue(pattern)
		m.editInput.Focus()
		m.validationError = ""
	}
}

// confirmEdit confirms the edit and updates the pattern
func (m *PatternConfigModel) confirmEdit() {
	pattern := strings.TrimSpace(m.editInput.Value())
	if pattern == "" {
		return
	}

	// Validate pattern if enabled
	if m.validationEnabled {
		valid, err := m.validatePattern(pattern)
		if !valid {
			m.validationError = err
			return
		}
	}

	// Update pattern in config
	if m.currentTab == 0 && m.editIndex >= 0 && m.editIndex < len(m.config.App.CustomIgnorePatterns) {
		m.config.App.CustomIgnorePatterns[m.editIndex] = pattern
	} else if m.currentTab == 1 && m.editIndex >= 0 && m.editIndex < len(m.config.App.ForceIncludePatterns) {
		m.config.App.ForceIncludePatterns[m.editIndex] = pattern
	}

	m.loadPatterns()
	m.cancelEdit()
}

// cancelEdit cancels editing
func (m *PatternConfigModel) cancelEdit() {
	m.editIndex = -1
	m.editInput.Blur()
}

// deleteSelected deletes the selected pattern
func (m *PatternConfigModel) deleteSelected() {
	if m.currentTab == 0 {
		selectedIndex := m.customPatterns.Index()
		if selectedIndex >= 0 && selectedIndex < len(m.config.App.CustomIgnorePatterns) {
			// Remove pattern
			patterns := m.config.App.CustomIgnorePatterns
			m.config.App.CustomIgnorePatterns = append(patterns[:selectedIndex], patterns[selectedIndex+1:]...)
		}
	} else {
		selectedIndex := m.forcePatterns.Index()
		if selectedIndex >= 0 && selectedIndex < len(m.config.App.ForceIncludePatterns) {
			// Remove pattern
			patterns := m.config.App.ForceIncludePatterns
			m.config.App.ForceIncludePatterns = append(patterns[:selectedIndex], patterns[selectedIndex+1:]...)
		}
	}

	m.loadPatterns()
}

// GetConfig returns the updated configuration
func (m *PatternConfigModel) GetConfig() *core.EnhancedConfig {
	return m.config
}

// View renders the pattern configuration interface
func (m PatternConfigModel) View() string {
	var sections []string

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1).
		Render("🎯 File Pattern Configuration")

	sections = append(sections, title)

	// Tab headers
	tabStyle := lipgloss.NewStyle().Padding(0, 2)
	activeTabStyle := tabStyle.Copy().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Background(lipgloss.Color("#3C3C3C"))

	tab1 := "🚫 Custom Ignore"
	tab2 := "✅ Force Include"

	if m.currentTab == 0 {
		tab1 = activeTabStyle.Render(tab1)
		tab2 = tabStyle.Render(tab2)
	} else {
		tab1 = tabStyle.Render(tab1)
		tab2 = activeTabStyle.Render(tab2)
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, tab1, tab2)
	sections = append(sections, tabs)

	// Current list
	var currentList string
	if m.currentTab == 0 {
		currentList = m.customPatterns.View()
	} else {
		currentList = m.forcePatterns.View()
	}
	sections = append(sections, currentList)

	// Input area
	if m.isAddingPattern {
		inputTitle := lipgloss.NewStyle().Bold(true).Render("Add New Pattern:")
		inputArea := lipgloss.JoinVertical(lipgloss.Left,
			inputTitle,
			m.addInput.View(),
			"Press Enter to add, Esc to cancel",
		)
		sections = append(sections, inputArea)
	} else if m.editIndex != -1 {
		inputTitle := lipgloss.NewStyle().Bold(true).Render("Edit Pattern:")
		inputArea := lipgloss.JoinVertical(lipgloss.Left,
			inputTitle,
			m.editInput.View(),
			"Press Enter to save, Esc to cancel",
		)
		sections = append(sections, inputArea)
	}

	// Validation error
	if m.validationError != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)
		sections = append(sections, errorStyle.Render("❌ "+m.validationError))
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginTop(1)

	var helpText string
	if m.isAddingPattern || m.editIndex != -1 {
		helpText = "Enter: confirm • Esc: cancel"
	} else {
		validationStatus := "disabled"
		if m.validationEnabled {
			validationStatus = "enabled"
		}
		helpText = fmt.Sprintf("Tab: switch tabs • A: add • E: edit • D: delete • V: toggle validation (%s) • Q: quit", validationStatus)
	}

	sections = append(sections, helpStyle.Render(helpText))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// newPatternDelegate creates a new list delegate for pattern items
func newPatternDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	return delegate
}
