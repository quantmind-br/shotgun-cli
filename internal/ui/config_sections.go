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

// renderSectionTabs renders the section navigation tabs with responsive layout
func (m EnhancedConfigFormModel) renderSectionTabs() string {
	ctx := m.GetLayoutContext()
	vertical, useSymbols, maxTabs := ctx.GetTabLayout()

	var tabs []string

	// Determine how many tabs to show
	tabsToShow := len(m.sections)
	if maxTabs > 0 && tabsToShow > maxTabs {
		tabsToShow = maxTabs
	}

	for i, section := range m.sections[:tabsToShow] {
		// Use adaptive styles
		var style lipgloss.Style
		if i == m.activeSection {
			style = GetFieldActiveStyle(ctx)
		} else {
			style = GetFieldInactiveStyle(ctx)
		}

		// Add validation indicator
		hasErrors := m.sectionHasErrors(i)
		validationIcon := "✓"
		if hasErrors {
			validationIcon = "⚠"
		}

		// Create tab content based on layout
		var tabContent string
		if useSymbols {
			// Mobile mode - use only icons and validation indicators
			if i == m.activeSection {
				tabContent = fmt.Sprintf("▶ %s %s", section.Icon, validationIcon)
			} else {
				tabContent = fmt.Sprintf("%s %s", section.Icon, validationIcon)
			}
		} else {
			// Standard mode - include names
			activeIndicator := ""
			if i == m.activeSection && vertical {
				activeIndicator = "▶ " // Visual indicator for active tab in vertical mode
			}

			tabContent = fmt.Sprintf("%s%s %s %s", activeIndicator, section.Icon, section.Name, validationIcon)
		}

		tab := style.Render(tabContent)
		tabs = append(tabs, tab)
	}

	// Show overflow indicator if we're hiding tabs
	if maxTabs > 0 && len(m.sections) > maxTabs {
		overflowStyle := GetHelpStyle(ctx)
		overflowTab := overflowStyle.Render(fmt.Sprintf("... (+%d)", len(m.sections)-maxTabs))
		tabs = append(tabs, overflowTab)
	}

	// Layout tabs based on responsive settings
	if vertical {
		return lipgloss.JoinVertical(lipgloss.Left, tabs...)
	} else {
		return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	}
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

	// Responsive panel rendering with collapsible behavior
	ctx := m.GetLayoutContext()
	content := []string{header, fields}

	// Connection test panel (for OpenAI section) - Priority 1 (high)
	if section.Name == "OpenAI API" && m.showConnectionTest {
		if ctx.ShouldShowPanel("connection", 1) {
			// Full panel for wider screens
			connectionTest := m.renderConnectionTest()
			content = append(content, connectionTest)
		} else if m.connectionResult != "" || m.connectionTesting {
			// Collapsed state for narrow screens
			collapsedConnection := m.renderCollapsedConnectionTest()
			content = append(content, collapsedConnection)
		}
	}

	// Validation details panel - Priority 1 (high) when errors exist
	if m.showValidationDetails && len(m.errors) > 0 {
		if ctx.ShouldShowPanel("validation", 1) {
			// Full panel for wider screens
			validationPanel := m.renderValidationPanel()
			content = append(content, validationPanel)
		} else {
			// Collapsed state for narrow screens - always show error count
			collapsedValidation := m.renderCollapsedValidationPanel()
			content = append(content, collapsedValidation)
		}
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
	// Get layout context for responsive behavior
	ctx := m.GetLayoutContext()

	// Field label with required indicator
	label := field.Label
	if field.Required {
		label = label + " *"
	}
	labelStyle := GetFieldLabelStyle(ctx)
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

	// Apply styling based on state using adaptive styles
	var fieldStyle lipgloss.Style
	if isActive {
		fieldStyle = GetFieldActiveStyle(ctx)
	} else {
		fieldStyle = GetFieldInactiveStyle(ctx)
	}

	// Add validation error indicator
	if err, hasError := m.errors[field.Key]; hasError {
		fieldDisplay = fieldDisplay + " " + GetErrorStyle().Render("⚠ "+err)
	} else if m.realTimeValidation && isActive {
		fieldDisplay = fieldDisplay + " " + GetSuccessStyle().Render("✓")
	}

	fieldPart := fieldStyle.Render(fieldDisplay)

	// Get field layout configuration - vertical for narrow breakpoints, horizontal otherwise
	vertical, _, _ := ctx.GetFieldLayout()

	var line string
	if vertical {
		// Vertical layout: label above field for narrow screens
		line = lipgloss.JoinVertical(lipgloss.Left, labelStyle.Render(label), fieldPart)
	} else {
		// Horizontal layout: label and field side by side for wider screens
		line = lipgloss.JoinHorizontal(lipgloss.Top, labelStyle.Render(label), "  ", fieldPart)
	}

	// Help text with adaptive layout
	if field.HelpText != "" && isActive {
		// Get help text configuration from layout context
		showHelp, indentation, iconOnly := ctx.GetHelpTextLayout()

		if showHelp {
			helpStyle := GetHelpStyle(ctx)

			if iconOnly {
				// Mobile mode - icon only, no full text
				helpText := "💡"
				help := helpStyle.Render(helpText)
				return lipgloss.JoinVertical(lipgloss.Left, line, help)
			} else {
				// Standard mode with intelligent line wrapping
				wrappedHelp := m.renderWrappedHelpText(field.HelpText, ctx, indentation)
				return lipgloss.JoinVertical(lipgloss.Left, line, wrappedHelp)
			}
		}
	}

	return line
}

// renderWrappedHelpText renders help text with intelligent line wrapping
func (m EnhancedConfigFormModel) renderWrappedHelpText(helpText string, ctx *LayoutContext, indentation int) string {
	helpStyle := GetHelpStyle(ctx)

	// Calculate available width for help text
	availableWidth := ctx.Width - indentation - 4 // Reserve space for icon and margins
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// Prepare indentation string
	indentStr := strings.Repeat(" ", indentation)
	iconPrefix := "💡 "

	// If help text fits in one line, render it simply
	if len(helpText) <= availableWidth-len(iconPrefix) {
		helpLine := indentStr + iconPrefix + helpText
		return helpStyle.Render(helpLine)
	}

	// Wrap long help text intelligently
	words := strings.Fields(helpText)
	var lines []string
	var currentLine []string
	currentLength := 0

	for _, word := range words {
		// Check if adding this word would exceed the line limit
		wordLength := len(word)
		spaceIfNeeded := 0
		if len(currentLine) > 0 {
			spaceIfNeeded = 1 // For the space between words
		}

		if currentLength+spaceIfNeeded+wordLength > availableWidth-len(iconPrefix) && len(currentLine) > 0 {
			// Current line is full, start a new line
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentLength = wordLength
		} else {
			// Add word to current line
			currentLine = append(currentLine, word)
			currentLength += spaceIfNeeded + wordLength
		}
	}

	// Add any remaining words
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	// Render each line with proper indentation
	var renderedLines []string
	for i, line := range lines {
		var formattedLine string
		if i == 0 {
			// First line gets the icon
			formattedLine = indentStr + iconPrefix + line
		} else {
			// Subsequent lines get aligned indentation
			formattedLine = indentStr + strings.Repeat(" ", len(iconPrefix)) + line
		}
		renderedLines = append(renderedLines, helpStyle.Render(formattedLine))
	}

	return strings.Join(renderedLines, "\n")
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

// renderCollapsedConnectionTest renders a compact connection test status
func (m EnhancedConfigFormModel) renderCollapsedConnectionTest() string {
	ctx := m.GetLayoutContext()
	helpStyle := GetHelpStyle(ctx)

	var statusText string
	if m.connectionTesting {
		statusText = "🔄 Testing..."
	} else if m.connectionSuccess {
		statusText = "🔗✓ Connected"
	} else if m.connectionResult != "" {
		statusText = "🔗✗ Failed"
	} else {
		statusText = "🔗 Ready to test (Ctrl+T)"
	}

	return helpStyle.Render(statusText) + "\n"
}

// renderCollapsedValidationPanel renders a compact error summary
func (m EnhancedConfigFormModel) renderCollapsedValidationPanel() string {
	errorCount := len(m.errors)
	if errorCount == 0 {
		return ""
	}

	errorStyle := GetErrorStyle()
	var statusText string
	if errorCount == 1 {
		statusText = "⚠️ 1 error"
	} else {
		statusText = fmt.Sprintf("⚠️ %d errors", errorCount)
	}

	return errorStyle.Render(statusText) + "\n"
}

// renderStatusBar renders the status and actions bar with adaptive layout
func (m EnhancedConfigFormModel) renderStatusBar() string {
	// Get layout context for responsive behavior
	ctx := m.GetLayoutContext()
	var sections []string

	// Operation status with adaptive styling
	if m.lastOperationStatus != "" {
		var style lipgloss.Style
		if m.lastOperationSuccess {
			style = GetSuccessStyle()
		} else {
			style = GetErrorStyle()
		}

		status := style.Render(m.lastOperationStatus)
		sections = append(sections, status)
	}

	// Status indicators - group by priority for space management
	var indicators []string

	// High priority indicators (always show when present)
	if m.autoSave {
		indicators = append(indicators, "🔄 Auto-save enabled")
	}

	if m.realTimeValidation {
		indicators = append(indicators, "✓ Real-time validation")
	}

	// Render indicators if present
	if len(indicators) > 0 {
		helpStyle := GetHelpStyle(ctx)

		// Group indicators based on available space
		if ctx.ShowFullShortcuts && len(indicators) > 1 {
			// Wide screens - show all indicators on one line
			indicatorText := strings.Join(indicators, " | ")
			sections = append(sections, helpStyle.Render(indicatorText))
		} else {
			// Narrow screens - show each indicator separately for better wrapping
			for _, indicator := range indicators {
				sections = append(sections, helpStyle.Render(indicator))
			}
		}
	}

	// Adaptive keyboard shortcuts using layout context
	shortcutGroups := ctx.GetShortcutGroups()
	helpStyle := GetHelpStyle(ctx)

	for _, group := range shortcutGroups {
		groupText := strings.Join(group, " | ")
		shortcutLine := helpStyle.Render(groupText)
		sections = append(sections, shortcutLine)
	}

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

// SetSize sets the display size for the form with intelligent breakpoint detection
func (m *EnhancedConfigFormModel) SetSize(width, height int) {
	// Initialize layout context if not exists or update existing one
	if m.layoutContext == nil {
		m.layoutContext = NewLayoutContext(width, height)
	} else {
		// Check if breakpoint changed - this triggers re-render if needed
		breakpointChanged := m.layoutContext.Update(width, height)

		// If breakpoint changed, we might need to trigger a full re-render
		// This is handled by the calling code checking if the layout changed
		_ = breakpointChanged
	}

	// Update model's width/height for backward compatibility
	m.width = width
	m.height = height

	// Apply responsive field widths based on calculated layout
	fieldWidth := m.layoutContext.FieldWidth
	if fieldWidth < 10 {
		fieldWidth = 10 // Absolute minimum
	}

	// Update all field input widths
	for sectionIdx := range m.sections {
		for fieldIdx := range m.sections[sectionIdx].Fields {
			m.sections[sectionIdx].Fields[fieldIdx].Input.Width = fieldWidth
		}
	}
}

// GetLayoutContext returns the current layout context for responsive rendering
func (m *EnhancedConfigFormModel) GetLayoutContext() *LayoutContext {
	if m.layoutContext == nil {
		// Create default layout context if none exists
		m.layoutContext = NewLayoutContext(m.width, m.height)
	}
	return m.layoutContext
}

// IsBreakpointChange checks if changing to the given size would trigger a breakpoint change
func (m *EnhancedConfigFormModel) IsBreakpointChange(newWidth int) bool {
	if m.layoutContext == nil {
		return true // First layout always triggers change
	}
	return m.layoutContext.IsBreakpointChange(newWidth)
}
