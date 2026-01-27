package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
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

	showingFullPreview bool
	previewScrollY     int
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
		manager, err := template.NewManager(template.ManagerConfig{
			CustomPath: viper.GetString(config.KeyTemplateCustomPath),
		})
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

func (m *TemplateSelectionModel) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || m.loading || len(m.templates) == 0 {
		return nil
	}

	if m.showingFullPreview {
		return m.handleModalKeyPress(keyMsg)
	}

	switch keyMsg.String() {
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
		}
	case "home":
		m.cursor = 0
	case "end":
		m.cursor = len(m.templates) - 1
	case "v":
		if m.cursor >= 0 && m.cursor < len(m.templates) {
			m.showingFullPreview = true
			m.previewScrollY = 0
		}
	}

	return nil
}

func (m *TemplateSelectionModel) GetSelected() *template.Template {
	return m.selectedTemplate
}

func (m *TemplateSelectionModel) SetSelectedForTest(tmpl *template.Template) {
	m.selectedTemplate = tmpl
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
	if m.showingFullPreview {
		return m.renderFullPreviewModal()
	}

	header := styles.RenderHeader(2, "Choose Template")

	if earlyReturn := m.checkEarlyReturns(header); earlyReturn != "" {
		return earlyReturn
	}

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")

	// Render template list directly
	content.WriteString(m.renderTemplateList())

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
		noTemplatesMsg := styles.RenderWarning("No templates found.") + "\n\n" +
			styles.HelpStyle.Render("To add templates:\n") +
			styles.HelpStyle.Render("  1. Create a template file in ~/.config/shotgun/templates/\n") +
			styles.HelpStyle.Render("  2. Or run: shotgun template create <name>\n") +
			styles.HelpStyle.Render("\nTemplates use {{TASK}}, {{RULES}}, and {{FILES}} variables.")
		return header + "\n\n" + noTemplatesMsg
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

func (m *TemplateSelectionModel) renderFooter() string {
	line1 := []string{
		"↑/↓: Navigate",
		"Enter/Space: Select",
		"v: View full",
	}

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

func (m *TemplateSelectionModel) handleModalKeyPress(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "q":
		m.showingFullPreview = false
		m.previewScrollY = 0
	case "j", "down":
		m.scrollPreviewDown(1)
	case "k", "up":
		m.scrollPreviewUp(1)
	case "pgdown", "ctrl+d":
		m.scrollPreviewDown(m.getVisibleHeight())
	case "pgup", "ctrl+u":
		m.scrollPreviewUp(m.getVisibleHeight())
	case "g", "home":
		m.previewScrollY = 0
	case "G", "end":
		m.scrollToEnd()
	}
	return nil
}

func (m *TemplateSelectionModel) scrollPreviewDown(lines int) {
	if m.cursor < 0 || m.cursor >= len(m.templates) {
		return
	}
	tmpl := m.templates[m.cursor]
	totalLines := len(strings.Split(tmpl.Content, "\n"))
	maxScroll := totalLines - m.getVisibleHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}

	m.previewScrollY += lines
	if m.previewScrollY > maxScroll {
		m.previewScrollY = maxScroll
	}
}

func (m *TemplateSelectionModel) scrollPreviewUp(lines int) {
	m.previewScrollY -= lines
	if m.previewScrollY < 0 {
		m.previewScrollY = 0
	}
}

func (m *TemplateSelectionModel) getVisibleHeight() int {
	height := m.height - 8
	if height < 5 {
		height = 5
	}
	return height
}

func (m *TemplateSelectionModel) scrollToEnd() {
	if m.cursor < 0 || m.cursor >= len(m.templates) {
		return
	}
	tmpl := m.templates[m.cursor]
	totalLines := len(strings.Split(tmpl.Content, "\n"))
	maxScroll := totalLines - m.getVisibleHeight()
	if maxScroll < 0 {
		maxScroll = 0
	}
	m.previewScrollY = maxScroll
}

func (m *TemplateSelectionModel) renderFullPreviewModal() string {
	if m.cursor < 0 || m.cursor >= len(m.templates) {
		return "No template selected"
	}

	tmpl := m.templates[m.cursor]
	var content strings.Builder

	header := styles.RenderHeader(0, "Template: "+tmpl.Name)
	content.WriteString(header)
	content.WriteString("\n\n")

	lines := strings.Split(tmpl.Content, "\n")
	visibleHeight := m.getVisibleHeight()

	startLine := m.previewScrollY
	endLine := startLine + visibleHeight
	if endLine > len(lines) {
		endLine = len(lines)
	}

	for i := startLine; i < endLine; i++ {
		lineContent := lines[i]
		if len(lineContent) > m.width-4 {
			lineContent = lineContent[:m.width-7] + "..."
		}
		content.WriteString(styles.CodeStyle.Render(lineContent))
		content.WriteString("\n")
	}

	if len(lines) > visibleHeight {
		scrollInfo := fmt.Sprintf("\n[Lines %d-%d of %d]",
			startLine+1, endLine, len(lines))
		content.WriteString(lipgloss.NewStyle().Foreground(styles.MutedColor).Render(scrollInfo))
	}

	content.WriteString("\n\n")
	footer := styles.RenderFooter([]string{
		"j/k: Scroll",
		"PgUp/PgDn: Page",
		"g/G: Top/Bottom",
		"Esc/q: Close",
	})
	content.WriteString(footer)

	return content.String()
}
