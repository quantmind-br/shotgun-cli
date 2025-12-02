package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type TemplateSelectionModel struct {
	templates        []*template.Template
	cursor           int
	selectedTemplate *template.Template
	width            int
	height           int
	loading          bool
	err              error
}

type TemplatesLoadedMsg struct {
	Templates []*template.Template
}

type TemplatesErrorMsg struct {
	Err error
}

func NewTemplateSelection() *TemplateSelectionModel {
	return &TemplateSelectionModel{
		loading: true,
	}
}

func (m *TemplateSelectionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *TemplateSelectionModel) LoadTemplates() tea.Cmd {
	return func() tea.Msg {
		manager, err := template.NewManager()
		if err != nil {
			return TemplatesErrorMsg{Err: err}
		}
		templateList, err := manager.ListTemplates()
		if err != nil {
			return TemplatesErrorMsg{Err: err}
		}

		// Convert []template.Template to []*template.Template
		templates := make([]*template.Template, len(templateList))
		for i := range templateList {
			templates[i] = &templateList[i]
		}
		return TemplatesLoadedMsg{Templates: templates}
	}
}

func (m *TemplateSelectionModel) Update(msg tea.KeyMsg) (*template.Template, tea.Cmd) {
	if m.loading || len(m.templates) == 0 {
		return nil, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.templates)-1 {
			m.cursor++
		}
	case "enter", " ":
		if m.cursor >= 0 && m.cursor < len(m.templates) {
			m.selectedTemplate = m.templates[m.cursor]
			return m.templates[m.cursor], nil
		}
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.templates) - 1
	}

	return nil, nil
}

func (m *TemplateSelectionModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case TemplatesLoadedMsg:
		m.templates = msg.Templates
		m.loading = false
		m.err = nil
		if len(m.templates) > 0 {
			m.cursor = 0
		}
	case TemplatesErrorMsg:
		m.err = msg.Err
		m.loading = false
	}
	return nil
}

func (m *TemplateSelectionModel) View() string {
	header := styles.RenderHeader(2, "Choose Template")

	if earlyReturn := m.checkEarlyReturns(header); earlyReturn != "" {
		return earlyReturn
	}

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")

	// Create two-column layout
	listWidth := 35
	detailsWidth := m.width - listWidth - 5
	if detailsWidth < 30 {
		detailsWidth = 30
	}

	// Render template list
	listContent := m.renderTemplateList()
	detailsContent := m.renderTemplateDetails()

	// Side by side layout
	listStyle := lipgloss.NewStyle().
		Width(listWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(0, 1)

	detailsStyle := lipgloss.NewStyle().
		Width(detailsWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.SecondaryColor).
		Padding(0, 1)

	listBox := listStyle.Render(listContent)
	detailsBox := detailsStyle.Render(detailsContent)

	// Join horizontally
	combined := lipgloss.JoinHorizontal(lipgloss.Top, listBox, "  ", detailsBox)
	content.WriteString(combined)

	footer := m.renderFooter()
	content.WriteString("\n\n")
	content.WriteString(footer)

	return content.String()
}

func (m *TemplateSelectionModel) checkEarlyReturns(header string) string {
	if m.loading {
		loadingStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		return header + "\n\n" + loadingStyle.Render("⏳ Loading templates...")
	}
	if m.err != nil {
		return header + "\n\n" + styles.RenderError(fmt.Sprintf("Error loading templates: %v", m.err))
	}
	if len(m.templates) == 0 {
		return header + "\n\n" + styles.RenderWarning("No templates found.")
	}
	return ""
}

func (m *TemplateSelectionModel) renderTemplateList() string {
	var content strings.Builder

	// Title
	title := styles.SubtitleStyle.Render("Templates")
	content.WriteString(title)
	content.WriteString("\n")
	content.WriteString(styles.RenderSeparator(30))
	content.WriteString("\n")

	startIdx, endIdx := m.calculateScrollBounds()

	// Show scroll indicator if needed
	if startIdx > 0 {
		scrollUp := lipgloss.NewStyle().Foreground(styles.MutedColor).Render("  ↑ more above")
		content.WriteString(scrollUp)
		content.WriteString("\n")
	}

	for i := startIdx; i < endIdx && i < len(m.templates); i++ {
		line := m.formatTemplateLine(i)
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show scroll indicator if needed
	if endIdx < len(m.templates) {
		scrollDown := lipgloss.NewStyle().Foreground(styles.MutedColor).Render("  ↓ more below")
		content.WriteString(scrollDown)
	}

	return content.String()
}

func (m *TemplateSelectionModel) calculateScrollBounds() (int, int) {
	availableHeight := m.height - 12
	if availableHeight < 5 {
		availableHeight = 5
	}

	startIdx := 0
	endIdx := len(m.templates)

	if len(m.templates) > availableHeight {
		if m.cursor < availableHeight/2 {
			endIdx = availableHeight
		} else if m.cursor >= len(m.templates)-availableHeight/2 {
			startIdx = len(m.templates) - availableHeight
		} else {
			startIdx = m.cursor - availableHeight/2
			endIdx = startIdx + availableHeight
		}
	}

	return startIdx, endIdx
}

func (m *TemplateSelectionModel) formatTemplateLine(i int) string {
	tmpl := m.templates[i]
	isSelected := m.selectedTemplate != nil && m.selectedTemplate.Name == tmpl.Name
	isCursor := i == m.cursor

	var line string
	name := tmpl.Name

	if isCursor {
		cursor := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("▶")
		nameStyled := styles.SelectedStyle.Render(" " + name)
		line = cursor + nameStyled
	} else {
		nameStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
		line = "  " + nameStyle.Render(name)
	}

	if isSelected {
		checkmark := styles.SuccessStyle.Render(" ✓")
		line += checkmark
	}

	return line
}

func (m *TemplateSelectionModel) renderTemplateDetails() string {
	if m.cursor < 0 || m.cursor >= len(m.templates) {
		return "Select a template to see details"
	}

	selectedTemplate := m.templates[m.cursor]
	var content strings.Builder

	// Template name
	nameLabel := styles.SubtitleStyle.Render("Template: ")
	nameValue := styles.StatsValueStyle.Render(selectedTemplate.Name)
	content.WriteString(nameLabel)
	content.WriteString(nameValue)
	content.WriteString("\n")
	content.WriteString(styles.RenderSeparator(40))
	content.WriteString("\n\n")

	// Description
	descLabel := styles.TitleStyle.Render("📋 Description")
	content.WriteString(descLabel)
	content.WriteString("\n")
	descText := lipgloss.NewStyle().Foreground(styles.TextColor).Render(selectedTemplate.Description)
	content.WriteString(descText)
	content.WriteString("\n\n")

	// Required variables
	if len(selectedTemplate.RequiredVars) > 0 {
		varsLabel := styles.TitleStyle.Render("🔧 Required Variables")
		content.WriteString(varsLabel)
		content.WriteString("\n")
		for _, variable := range selectedTemplate.RequiredVars {
			varStyle := lipgloss.NewStyle().Foreground(styles.Nord15)
			bullet := lipgloss.NewStyle().Foreground(styles.MutedColor).Render("  • ")
			content.WriteString(bullet)
			content.WriteString(varStyle.Render(variable))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Template preview (first few lines)
	if selectedTemplate.Content != "" {
		previewLabel := styles.TitleStyle.Render("👁 Preview")
		content.WriteString(previewLabel)
		content.WriteString("\n")

		// Get first 5 non-empty lines
		lines := strings.Split(selectedTemplate.Content, "\n")
		previewLines := 0
		maxPreviewLines := 5
		for _, line := range lines {
			if previewLines >= maxPreviewLines {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				// Truncate long lines
				if len(trimmed) > 50 {
					trimmed = trimmed[:47] + "..."
				}
				lineStyled := styles.CodeStyle.Render("  " + trimmed)
				content.WriteString(lineStyled)
				content.WriteString("\n")
				previewLines++
			}
		}
		if len(lines) > maxPreviewLines {
			moreLines := lipgloss.NewStyle().Foreground(styles.MutedColor).Italic(true)
			content.WriteString(moreLines.Render(fmt.Sprintf("  ... (%d more lines)", len(lines)-maxPreviewLines)))
		}
	}

	return content.String()
}

func (m *TemplateSelectionModel) renderFooter() string {
	line1 := []string{
		"↑/↓: Navigate",
		"Enter/Space: Select",
	}

	// Determine what F8 will do based on selected template's requirements
	nextAction := "F8: Next"
	if m.selectedTemplate != nil {
		requiresTask := m.selectedTemplate.HasVariable(template.VarTask)
		requiresRules := m.selectedTemplate.HasVariable(template.VarRules)

		if !requiresTask && !requiresRules {
			nextAction = "F8: Skip to Review"
		} else if !requiresTask {
			nextAction = "F8: Skip to Rules"
		}
	}

	line2 := []string{
		"F7: Back",
		nextAction,
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	return styles.RenderFooter(line1) + "\n" + styles.RenderFooter(line2)
}
