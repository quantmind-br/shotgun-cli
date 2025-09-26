package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type TemplateSelectionModel struct {
	templates     []*template.Template
	cursor        int
	width         int
	height        int
	loading       bool
	err           error
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

	if m.loading {
		return header + "\n\nLoading templates..."
	}

	if m.err != nil {
		return header + "\n\n" + styles.RenderError(fmt.Sprintf("Error loading templates: %v", m.err))
	}

	if len(m.templates) == 0 {
		return header + "\n\nNo templates found."
	}

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")

	// Calculate available height for list
	availableHeight := m.height - 8 // Reserve space for header, footer, and description
	startIdx := 0
	endIdx := len(m.templates)

	// Implement scrolling if needed
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

	// Render template list
	for i := startIdx; i < endIdx && i < len(m.templates); i++ {
		template := m.templates[i]
		prefix := "  "
		if i == m.cursor {
			prefix = styles.SelectedStyle.Render("▶ ")
		}

		line := fmt.Sprintf("%s%s", prefix, template.Name)
		if i == m.cursor {
			line = styles.SelectedStyle.Render(line)
		}
		content.WriteString(line)
		content.WriteString("\n")
	}

	// Show selected template description
	if m.cursor >= 0 && m.cursor < len(m.templates) {
		selectedTemplate := m.templates[m.cursor]
		content.WriteString("\n")
		content.WriteString(styles.TitleStyle.Render("Description:"))
		content.WriteString("\n")
		content.WriteString(selectedTemplate.Description)
		content.WriteString("\n")

		if len(selectedTemplate.RequiredVars) > 0 {
			content.WriteString("\n")
			content.WriteString(styles.TitleStyle.Render("Required Variables:"))
			content.WriteString("\n")
			for _, variable := range selectedTemplate.RequiredVars {
				content.WriteString(fmt.Sprintf("  • %s", variable))
				content.WriteString("\n")
			}
		}
	}

	shortcuts := []string{
		"↑/↓: Navigate",
		"Enter/Space: Select",
		"F8: Next",
		"F10: Back",
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(shortcuts)
	content.WriteString("\n")
	content.WriteString(footer)

	return content.String()
}