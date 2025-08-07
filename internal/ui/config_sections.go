package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// View renders the enhanced configuration form
func (m EnhancedConfigFormModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading configuration..."
	}

	var sections []string

	// Header with application info
	header := m.renderEnhancedHeader()
	sections = append(sections, header)

	// Section navigation tabs
	tabs := m.renderSectionTabs()
	sections = append(sections, tabs)

	// Current section content
	sectionContent := m.renderCurrentSection()
	sections = append(sections, sectionContent)

	// Status and actions bar
	statusBar := m.renderStatusBar()
	sections = append(sections, statusBar)

	// Help panel (if shown)
	if m.showHelp {
		help := m.renderEnhancedHelp()
		sections = append(sections, help)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderEnhancedHeader renders the enhanced configuration header
func (m EnhancedConfigFormModel) renderEnhancedHeader() string {
	title := enhancedSectionTitleStyle.Render("🔧 Enhanced Configuration")
	subtitle := enhancedSectionDescStyle.Render("Advanced configuration with real-time validation and Windows compatibility")

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "")
}

// renderSectionTabs renders the section navigation tabs
func (m EnhancedConfigFormModel) renderSectionTabs() string {
	var tabs []string

	for i, section := range m.sections {
		style := enhancedFieldInactiveStyle
		if i == m.activeSection {
			style = enhancedFieldActiveStyle
		}

		// Add validation indicator
		hasErrors := m.sectionHasErrors(i)
		validationIcon := "✓"
		if hasErrors {
			validationIcon = "⚠"
		}

		tabContent := fmt.Sprintf("%s %s %s", section.Icon, section.Name, validationIcon)
		tab := style.Render(tabContent)
		tabs = append(tabs, tab)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderCurrentSection renders the currently active section
func (m EnhancedConfigFormModel) renderCurrentSection() string {
	if m.activeSection < 0 || m.activeSection >= len(m.sections) {
		return "Invalid section"
	}

	section := m.sections[m.activeSection]

	// Section header
	header := m.renderSectionHeader(section)

	// Fields
	var fieldViews []string
	for i, field := range section.Fields {
		fieldView := m.renderEnhancedField(field, i == m.activeField)
		fieldViews = append(fieldViews, fieldView)
	}

	fields := lipgloss.JoinVertical(lipgloss.Left, fieldViews...)

	// Connection test panel (for OpenAI section)
	var connectionTest string
	if section.Name == "OpenAI API" && m.showConnectionTest {
		connectionTest = m.renderConnectionTest()
	}

	// Validation details panel
	var validationPanel string
	if m.showValidationDetails && len(m.errors) > 0 {
		validationPanel = m.renderValidationPanel()
	}

	content := []string{header, fields}
	if connectionTest != "" {
		content = append(content, connectionTest)
	}
	if validationPanel != "" {
		content = append(content, validationPanel)
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderSectionHeader renders the header for a configuration section
func (m EnhancedConfigFormModel) renderSectionHeader(section EnhancedConfigSection) string {
	title := enhancedSectionTitleStyle.Render(fmt.Sprintf("%s %s", section.Icon, section.Name))
	description := enhancedSectionDescStyle.Render(section.Description)

	return lipgloss.JoinVertical(lipgloss.Left, title, description, "")
}

// renderEnhancedField renders a configuration field with enhanced features
func (m EnhancedConfigFormModel) renderEnhancedField(field EnhancedConfigField, isActive bool) string {
	// Field label with required indicator
	label := field.Label
	if field.Required {
		label = label + " *"
	}
	labelStyle := enhancedFieldLabelStyle
	if isActive {
		labelStyle = labelStyle.Foreground(lipgloss.Color("12"))
	}
	// Field input/display
	var fieldDisplay string

	switch field.Type {
	case FieldToggle:
		// Render toggle field
		value := "✗ Disabled"
		if field.ManualValue == "true" || field.Value == "true" {
			value = "✓ Enabled"
		}
		fieldDisplay = value

	case FieldSelect:
		// Render select field
		currentValue := field.ManualValue
		if currentValue == "" {
			currentValue = field.Value
		}

		if len(field.Options) > 0 {
			options := strings.Join(field.Options, ", ")
			fieldDisplay = fmt.Sprintf("%s (options: %s)", currentValue, options)
		} else {
			fieldDisplay = currentValue
		}

	case FieldPassword:
		// Render password field
		if field.ManualValue != "" || field.Value != "" {
			fieldDisplay = strings.Repeat("*", 8)
		} else {
			fieldDisplay = "(not set)"
		}

	default:
		// Render text/number field
		currentValue := field.ManualValue
		if currentValue == "" {
			currentValue = field.Value
		}

		if m.editing && isActive {
			// Show the actual input widget when editing
			fieldDisplay = field.Input.View()
		} else {
			// Show the current value
			if currentValue == "" {
				fieldDisplay = enhancedHelpStyle.Render("(" + field.Placeholder + ")")
			} else {
				fieldDisplay = currentValue
			}
		}
	}

	// Apply styling based on state
	fieldStyle := enhancedFieldInactiveStyle
	if isActive {
		fieldStyle = enhancedFieldActiveStyle
	}

	// Add validation error indicator
	if err, hasError := m.errors[field.Key]; hasError {
		fieldDisplay = fieldDisplay + " " + enhancedErrorStyle.Render("⚠ "+err)
	} else if m.realTimeValidation && isActive {
		fieldDisplay = fieldDisplay + " " + enhancedSuccessStyle.Render("✓")
	}

	fieldPart := fieldStyle.Render(fieldDisplay)

	// Combine label and field on one line for a more compact view
	line := lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(label), "  ", fieldPart)

	// Help text
	if field.HelpText != "" && isActive {
		// Indent help text to align with the field
		help := enhancedHelpStyle.Render(strings.Repeat(" ", 30) + "💡 " + field.HelpText)
		return lipgloss.JoinVertical(lipgloss.Left, line, help)
	}

	return line
}

// renderConnectionTest renders the connection test panel
func (m EnhancedConfigFormModel) renderConnectionTest() string {
	var content []string

	title := enhancedSectionTitleStyle.Render("🔗 Connection Test")
	content = append(content, title)

	if m.connectionTesting {
		status := "🔄 Testing connection..."
		content = append(content, status)
	} else if m.connectionResult != "" {
		var style lipgloss.Style
		if m.connectionSuccess {
			style = enhancedSuccessStyle
		} else {
			style = enhancedErrorStyle
		}

		result := style.Render(m.connectionResult)
		content = append(content, result)
	} else {
		instruction := enhancedHelpStyle.Render("Press Ctrl+T to test API connection")
		content = append(content, instruction)
	}

	panel := lipgloss.JoinVertical(lipgloss.Left, content...)
	return enhancedFieldInactiveStyle.Render(panel) + "\n"
}

// renderValidationPanel renders the validation details panel
func (m EnhancedConfigFormModel) renderValidationPanel() string {
	if len(m.errors) == 0 {
		return ""
	}

	var content []string

	title := enhancedErrorStyle.Render("⚠️ Validation Errors")
	content = append(content, title)

	for field, error := range m.errors {
		errorLine := fmt.Sprintf("• %s: %s", field, error)
		content = append(content, enhancedErrorStyle.Render(errorLine))
	}

	panel := lipgloss.JoinVertical(lipgloss.Left, content...)
	return enhancedFieldInactiveStyle.Render(panel) + "\n"
}

// renderStatusBar renders the status and actions bar
func (m EnhancedConfigFormModel) renderStatusBar() string {
	var sections []string

	// Operation status
	if m.lastOperationStatus != "" {
		style := enhancedSuccessStyle
		if !m.lastOperationSuccess {
			style = enhancedErrorStyle
		}

		status := style.Render(m.lastOperationStatus)
		sections = append(sections, status)
	}

	// Auto-save indicator
	if m.autoSave {
		autoSaveStatus := enhancedHelpStyle.Render("🔄 Auto-save enabled")
		sections = append(sections, autoSaveStatus)
	}

	// Real-time validation indicator
	if m.realTimeValidation {
		validationStatus := enhancedHelpStyle.Render("✓ Real-time validation")
		sections = append(sections, validationStatus)
	}

	// Keyboard shortcuts
	shortcuts := []string{
		"Ctrl+S: Save",
		"Ctrl+T: Test",
		"Ctrl+R: Reset",
		"?: Help",
		"Tab: Navigate",
		"Enter: Edit",
	}

	shortcutText := strings.Join(shortcuts, " | ")
	shortcutPanel := enhancedHelpStyle.Render(shortcutText)
	sections = append(sections, shortcutPanel)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderEnhancedHelp renders the enhanced help panel
func (m EnhancedConfigFormModel) renderEnhancedHelp() string {
	var content []string

	title := enhancedSectionTitleStyle.Render("🆘 Enhanced Configuration Help")
	content = append(content, title)

	helpSections := []struct {
		title string
		items []string
	}{
		{
			title: "Navigation",
			items: []string{
				"←/→ or h/l: Switch sections",
				"↑/↓ or j/k: Navigate fields",
				"Tab/Shift+Tab: Next/previous field",
			},
		},
		{
			title: "Editing",
			items: []string{
				"Enter/Space: Edit text field or toggle checkbox",
				"Escape: Cancel editing",
				"Ctrl+V: Paste (Windows compatible)",
			},
		},
		{
			title: "Actions",
			items: []string{
				"Ctrl+S: Save configuration",
				"Ctrl+T: Test API connection",
				"Ctrl+R: Reset to defaults",
				"Ctrl+C: Quit application",
			},
		},
		{
			title: "Features",
			items: []string{
				"✓ Real-time validation feedback",
				"🔄 Auto-save on field changes",
				"🔗 Connection testing",
				"⚠️ Enhanced error reporting",
				"🖥️ Windows input compatibility",
			},
		},
	}

	for _, section := range helpSections {
		sectionTitle := enhancedFieldLabelStyle.Render(section.title + ":")
		content = append(content, sectionTitle)

		for _, item := range section.items {
			content = append(content, "  "+enhancedHelpStyle.Render(item))
		}
		content = append(content, "")
	}

	// Windows-specific notes
	windowsNotes := enhancedSectionTitleStyle.Render("🖥️ Windows Compatibility Notes")
	content = append(content, windowsNotes)

	windowsHelp := []string{
		"• Enhanced paste handling with Ctrl+V",
		"• Manual text tracking to prevent character loss",
		"• Multiple input reset cycles for stability",
		"• Bracket paste mode support",
		"• Compatible with cmd.exe, PowerShell, and Windows Terminal",
	}

	for _, note := range windowsHelp {
		content = append(content, enhancedHelpStyle.Render(note))
	}

	helpPanel := lipgloss.JoinVertical(lipgloss.Left, content...)
	return enhancedFieldInactiveStyle.Render(helpPanel)
}

// Helper methods for view rendering

// sectionHasErrors checks if a section has validation errors
func (m EnhancedConfigFormModel) sectionHasErrors(sectionIndex int) bool {
	if sectionIndex < 0 || sectionIndex >= len(m.sections) {
		return false
	}

	section := m.sections[sectionIndex]
	for _, field := range section.Fields {
		if _, hasError := m.errors[field.Key]; hasError {
			return true
		}
	}

	return false
}

// GetConfigurationData returns the current configuration data
func (m EnhancedConfigFormModel) GetConfigurationData() *core.EnhancedConfig {
	return m.extractEnhancedConfigurationData()
}

// ValidateConfiguration validates the current configuration
func (m *EnhancedConfigFormModel) ValidateConfiguration() *core.ConfigValidationResult {
	config := m.extractEnhancedConfigurationData()
	validator := core.SetupEnhancedValidator()
	return config.Validate(validator)
}

// SetFieldValue sets a field value programmatically
func (m *EnhancedConfigFormModel) SetFieldValue(key, value string) error {
	for sectionIdx := range m.sections {
		for fieldIdx := range m.sections[sectionIdx].Fields {
			field := &m.sections[sectionIdx].Fields[fieldIdx]
			if field.Key == key {
				field.Value = value
				field.ManualValue = value
				field.Input.SetValue(value)

				// Validate the new value
				if field.Validator != nil {
					if err := field.Validator(value); err != nil {
						m.errors[key] = err.Error()
						return err
					} else {
						delete(m.errors, key)
					}
				}

				return nil
			}
		}
	}

	return fmt.Errorf("field not found: %s", key)
}

// GetFieldValue gets a field value
func (m EnhancedConfigFormModel) GetFieldValue(key string) (string, error) {
	for _, section := range m.sections {
		for _, field := range section.Fields {
			if field.Key == key {
				if field.ManualValue != "" {
					return field.ManualValue, nil
				}
				return field.Value, nil
			}
		}
	}

	return "", fmt.Errorf("field not found: %s", key)
}

// SetSize sets the display size for the form
func (m *EnhancedConfigFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Adjust field widths based on available space
	fieldWidth := width - 20 // Leave margin for styling
	if fieldWidth < 20 {
		fieldWidth = 20
	}

	for sectionIdx := range m.sections {
		for fieldIdx := range m.sections[sectionIdx].Fields {
			m.sections[sectionIdx].Fields[fieldIdx].Input.Width = fieldWidth
		}
	}
}
