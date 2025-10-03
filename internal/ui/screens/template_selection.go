package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

	m.renderTemplateList(&content)
	m.renderTemplateDetails(&content)

	footer := m.renderFooter()
	content.WriteString("\n")
	content.WriteString(footer)

	return content.String()
}

func (m *TemplateSelectionModel) checkEarlyReturns(header string) string {
	if m.loading {
		return header + "\n\nLoading templates..."
	}
	if m.err != nil {
		return header + "\n\n" + styles.RenderError(fmt.Sprintf("Error loading templates: %v", m.err))
	}
	if len(m.templates) == 0 {
		return header + "\n\nNo templates found."
	}
	return ""
}

func (m *TemplateSelectionModel) renderTemplateList(content *strings.Builder) {
	startIdx, endIdx := m.calculateScrollBounds()

	for i := startIdx; i < endIdx && i < len(m.templates); i++ {
		line := m.formatTemplateLine(i)
		content.WriteString(line)
		content.WriteString("\n")
	}
}

func (m *TemplateSelectionModel) calculateScrollBounds() (int, int) {
	availableHeight := m.height - 8
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
	template := m.templates[i]
	isSelected := m.selectedTemplate != nil && m.selectedTemplate.Name == template.Name

	if i == m.cursor {
		prefix := "▶ "
		suffix := ""
		if isSelected {
			suffix = " " + styles.SuccessStyle.Render("✓")
		}
		return styles.SelectedStyle.Render(prefix+template.Name) + suffix
	}

	if isSelected {
		return "  " + template.Name + " " + styles.SuccessStyle.Render("✓")
	}

	return "  " + template.Name
}

func (m *TemplateSelectionModel) renderTemplateDetails(content *strings.Builder) {
	if m.cursor < 0 || m.cursor >= len(m.templates) {
		return
	}

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

func (m *TemplateSelectionModel) renderFooter() string {
	shortcuts := []string{
		"↑/↓: Navigate",
		"Enter/Space: Select",
		"F8: Next",
		"F10: Back",
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	return styles.RenderFooter(shortcuts)
}