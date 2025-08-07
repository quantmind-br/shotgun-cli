package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the configuration form
func (m ConfigFormModel) View() string {
	if m.showHelp {
		return m.renderConfigHelp()
	}

	var content []string

	// Header
	title := sectionTitleStyle.Render("⚙️  shotgun-cli Configuration")
	content = append(content, title, "")

	// Navigation breadcrumb
	breadcrumb := m.renderBreadcrumb()
	content = append(content, breadcrumb, "")

	// Current section
	sectionContent := m.renderCurrentSection()
	content = append(content, sectionContent)

	// Operation status messages
	if m.lastOperationStatus != "" {
		statusContent := m.renderOperationStatus()
		content = append(content, "", statusContent)
	}

	// Status/error messages
	if len(m.errors) > 0 {
		content = append(content, "", m.renderErrors())
	}

	// Footer with controls
	footer := m.renderFooter()
	content = append(content, "", footer)

	return strings.Join(content, "\n")
}

// renderBreadcrumb shows navigation between sections
func (m ConfigFormModel) renderBreadcrumb() string {
	var items []string

	for i, section := range m.sections {
		style := fieldInactiveStyle
		if i == m.activeSection {
			style = fieldActiveStyle
		}

		title := section.Title
		if len(title) > 25 {
			title = title[:22] + "..."
		}

		items = append(items, style.Render(title))
	}

	return strings.Join(items, " → ")
}

// renderCurrentSection renders the active configuration section
func (m ConfigFormModel) renderCurrentSection() string {
	section := m.sections[m.activeSection]
	var content []string

	// Section header
	content = append(content, sectionTitleStyle.Render(section.Title))
	if section.Description != "" {
		content = append(content, sectionDescStyle.Render(section.Description))
	}
	content = append(content, "")

	// Fields
	for i, field := range section.Fields {
		fieldContent := m.renderField(field, i == m.activeField)
		content = append(content, fieldContent)

		// Add spacing between fields
		if i < len(section.Fields)-1 {
			content = append(content, "")
		}
	}

	return strings.Join(content, "\n")
}

// renderField renders a single configuration field
func (m ConfigFormModel) renderField(field ConfigField, isActive bool) string {
	var parts []string

	// Field label
	labelStyle := fieldLabelStyle
	if isActive {
		labelStyle = labelStyle.Foreground(lipgloss.Color("33")) // Yellow for active
	}

	label := labelStyle.Render(field.Label + ":")
	parts = append(parts, label)

	// Field value/input
	var valueContent string

	if m.editing && isActive {
		// Show manual text when editing to avoid textinput bugs
		if field.Type == FieldPassword {
			// Show asterisks for password fields
			valueContent = strings.Repeat("*", len(m.editingText))
		} else {
			// Show actual text for other fields
			valueContent = m.editingText
		}
		// Add cursor indicator
		valueContent += "|"
	} else {
		// Show current value
		switch field.Type {
		case FieldPassword:
			if field.Value != "" {
				valueContent = "••••••••••••••••"
			} else {
				valueContent = fieldInactiveStyle.Render("(not set)")
			}
		case FieldToggle:
			if field.Value == "true" {
				valueContent = "✅ Enabled"
			} else {
				valueContent = "❌ Disabled"
			}
		case FieldSelect:
			valueContent = fmt.Sprintf("📋 %s", field.Value)
		default:
			if field.Value != "" {
				valueContent = field.Value
			} else {
				valueContent = fieldInactiveStyle.Render("(empty)")
			}
		}

		// Apply active/inactive styling
		if isActive {
			valueContent = fieldActiveStyle.Render(" " + valueContent + " ")
		} else {
			valueContent = " " + valueContent
		}
	}

	// Combine label and value
	fieldLine := label + " " + valueContent
	parts = []string{fieldLine}

	// Add field help if active
	if isActive && field.Help != "" {
		helpText := configHelpStyle.Render("💡 " + field.Help)
		parts = append(parts, "   "+helpText)
	}

	// Add field error if present
	if errorMsg, hasError := m.errors[field.Label]; hasError {
		errorText := configErrorStyle.Render("❌ " + errorMsg)
		parts = append(parts, "   "+errorText)
	}

	return strings.Join(parts, "\n")
}

// renderErrors shows validation errors
func (m ConfigFormModel) renderErrors() string {
	if len(m.errors) == 0 {
		return ""
	}

	var errors []string
	errors = append(errors, configErrorStyle.Render("⚠️  Configuration Errors:"))

	for field, msg := range m.errors {
		errors = append(errors, fmt.Sprintf("  • %s: %s", field, msg))
	}

	return strings.Join(errors, "\n")
}

// renderOperationStatus shows the result of the last configuration operation
func (m ConfigFormModel) renderOperationStatus() string {
	if m.lastOperationStatus == "" {
		return ""
	}

	var style lipgloss.Style
	var icon string

	if m.lastOperationSuccess {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")). // Green
			Bold(true)
		icon = "✅"
	} else {
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true)
		icon = "❌"
	}

	var content []string

	// Main status message
	statusLine := style.Render(fmt.Sprintf("%s %s", icon, m.lastOperationStatus))
	content = append(content, statusLine)

	// Additional details if available
	if m.lastOperationDetails != "" {
		detailStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")). // Gray
			Italic(true)
		detailLine := detailStyle.Render("   " + m.lastOperationDetails)
		content = append(content, detailLine)
	}

	return strings.Join(content, "\n")
}

// renderFooter shows available actions and keyboard shortcuts
func (m ConfigFormModel) renderFooter() string {
	var controls []string

	if m.editing {
		controls = append(controls,
			"Enter: Save field",
			"Esc: Cancel editing",
		)
	} else {
		controls = append(controls,
			"↑↓/jk: Navigate fields",
			"←→/hl: Switch sections",
			"Enter/Space: Edit field",
			"F2: Save config",
			"F3: Reset",
			"F4: Test connection",
			"F1: Help",
			"Esc: Exit",
		)
	}

	helpText := configHelpStyle.Render(strings.Join(controls, " • "))

	// Add status indicators
	var statusItems []string

	// Connection status
	if m.activeSection == 0 { // OpenAI section
		if m.config.OpenAI.APIKey != "" {
			statusItems = append(statusItems, "🔑 API Key: Set")
		} else {
			statusItems = append(statusItems, "⚠️  API Key: Not Set")
		}
	}

	// Translation status - show actual availability, not just config setting
	if m.config.Translation.Enabled {
		// Check if translator is actually available and configured
		translatorWorking := false
		if m.config.OpenAI.APIKey != "" {
			// We have an API key, so translation should work
			translatorWorking = true
		}

		if translatorWorking {
			statusItems = append(statusItems, "🌐 Translation: Ready")
		} else {
			statusItems = append(statusItems, "⚠️  Translation: Enabled but API key missing")
		}
	} else {
		statusItems = append(statusItems, "🌐 Translation: Disabled")
	}

	var result []string
	if len(statusItems) > 0 {
		statusLine := configHelpStyle.Render(strings.Join(statusItems, " • "))
		result = append(result, statusLine)
	}
	result = append(result, helpText)

	return strings.Join(result, "\n")
}

// renderConfigHelp shows detailed help information
func (m ConfigFormModel) renderConfigHelp() string {
	helpContent := []string{
		sectionTitleStyle.Render("📖 Configuration Help"),
		"",
		"This configuration interface allows you to set up OpenAI API integration",
		"for automatic translation of tasks and rules from your native language to English.",
		"",
		sectionTitleStyle.Render("🔧 Configuration Sections:"),
		"",
		configHelpStyle.Render("OpenAI API Configuration:"),
		"  • API Key: Your OpenAI API key (stored directly in config file)",
		"  • Base URL: API endpoint (use default for OpenAI, custom for compatible services)",
		"  • Model: The AI model to use (gpt-4o recommended for best translation quality)",
		"  • Timeout: Request timeout in seconds",
		"  • Max Tokens: Maximum response length",
		"  • Temperature: Controls creativity (0.0=focused, 2.0=creative)",
		"",
		configHelpStyle.Render("Translation Settings:"),
		"  • Enable Translation: Turn automatic translation on/off",
		"  • Target Language: Language code for translations (usually 'en' for English)",
		"  • Custom Prompt: Override default translation instructions",
		"",
		configHelpStyle.Render("Application Settings:"),
		"  • Auto Save: Automatically save configuration changes",
		"  • Line Numbers: Show line numbers in text areas",
		"  • Default Template: Prompt template to select on startup",
		"",
		sectionTitleStyle.Render("🎮 Controls:"),
		"",
		configHelpStyle.Render("Navigation:"),
		"  • ↑↓ or j/k: Move between fields",
		"  • ←→ or h/l: Switch between sections",
		"  • Tab/Shift+Tab: Navigate fields",
		"",
		configHelpStyle.Render("Editing:"),
		"  • Enter/Space: Edit text field or toggle checkbox",
		"  • Enter: Save changes while editing",
		"  • Esc: Cancel editing or exit configuration",
		"",
		configHelpStyle.Render("Actions:"),
		"  • F2: Save all configuration changes",
		"  • F3: Reset to default values",
		"  • F4: Test API connection",
		"  • F1: Toggle this help screen",
		"",
		sectionTitleStyle.Render("🔐 Security:"),
		"",
		"Your API key is stored directly in the configuration file:",
		"  • Location: ~/.config/shotgun-cli/config.json",
		"  • Format: Plain text (ensure file permissions are secure)",
		"  • Recommendation: Use appropriate file permissions (600) to secure the config file",
		"",
		configHelpStyle.Render("Press 'F1' again to return to configuration"),
	}

	return strings.Join(helpContent, "\n")
}

// renderTestResult shows the result of connection testing
func (m ConfigFormModel) renderTestResult(success bool, message string) string {
	var style lipgloss.Style
	var icon string

	if success {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green
		icon = "✅"
	} else {
		style = configErrorStyle
		icon = "❌"
	}

	return style.Render(fmt.Sprintf("%s %s", icon, message))
}

// renderSaveStatus shows configuration save status
func (m ConfigFormModel) renderSaveStatus(saved bool, message string) string {
	if saved {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green
		return style.Render("💾 Configuration saved successfully")
	} else {
		return configErrorStyle.Render("❌ Failed to save configuration: " + message)
	}
}

// Utility methods for field operations

// GetConfigurationData extracts configuration data from the form
func (m ConfigFormModel) GetConfigurationData() map[string]interface{} {
	data := make(map[string]interface{})

	for _, section := range m.sections {
		sectionData := make(map[string]interface{})
		for _, field := range section.Fields {
			sectionData[field.Label] = field.Value
		}
		data[section.Title] = sectionData
	}

	return data
}

// ValidateConfiguration checks if all required fields are filled
func (m *ConfigFormModel) ValidateConfiguration() bool {
	m.errors = make(map[string]string) // Clear previous errors

	for _, section := range m.sections {
		for _, field := range section.Fields {
			if field.Required && strings.TrimSpace(field.Value) == "" {
				m.errors[field.Label] = "This field is required"
			}

			// Type-specific validation
			switch field.Type {
			case FieldNumber:
				if field.Value != "" {
					// Could add number validation here
				}
			case FieldText:
				if field.Label == "Temperature" && field.Value != "" {
					// Could add temperature range validation here
				}
			}
		}
	}

	return len(m.errors) == 0
}

// SetFieldValue updates a field value by label
func (m *ConfigFormModel) SetFieldValue(label, value string) {
	for sectionIdx := range m.sections {
		for fieldIdx := range m.sections[sectionIdx].Fields {
			if m.sections[sectionIdx].Fields[fieldIdx].Label == label {
				m.sections[sectionIdx].Fields[fieldIdx].Value = value
				m.sections[sectionIdx].Fields[fieldIdx].Input.SetValue(value)
				return
			}
		}
	}
}

// GetFieldValue retrieves a field value by label
func (m ConfigFormModel) GetFieldValue(label string) string {
	for _, section := range m.sections {
		for _, field := range section.Fields {
			if field.Label == label {
				return field.Value
			}
		}
	}
	return ""
}
