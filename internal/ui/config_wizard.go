package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	configSidebarWidth    = 24
	configHeaderHeight    = 3
	configFooterHeight    = 3
	configMessageDuration = 3 * time.Second
)

type ConfigSavedMsg struct{}

type ConfigSaveErrorMsg struct {
	Err error
}

type clearSavedMsgTick struct{}

type ConfigWizardModel struct {
	categories      []config.ConfigCategory
	categoryScreens map[config.ConfigCategory]*screens.ConfigCategoryModel
	activeCategory  int
	width           int
	height          int
	showHelp        bool
	confirmQuit     bool
	quitAfterSave   bool
	savedMessage    string
	errorMessage    string
}

func NewConfigWizard() *ConfigWizardModel {
	categories := config.AllCategories()
	categoryScreens := make(map[config.ConfigCategory]*screens.ConfigCategoryModel)

	for _, cat := range categories {
		categoryScreens[cat] = screens.NewConfigCategory(cat)
	}

	return &ConfigWizardModel{
		categories:      categories,
		categoryScreens: categoryScreens,
		activeCategory:  0,
	}
}

func (m *ConfigWizardModel) Init() tea.Cmd {
	return nil
}

func (m *ConfigWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case ConfigSavedMsg:
		m.savedMessage = "Configuration saved successfully!"
		m.errorMessage = ""
		if m.quitAfterSave {
			return m, tea.Quit
		}
		return m, m.scheduleClearMessage()

	case ConfigSaveErrorMsg:
		m.errorMessage = fmt.Sprintf("Error saving: %v", msg.Err)
		m.savedMessage = ""
		return m, m.scheduleClearMessage()

	case clearSavedMsgTick:
		m.savedMessage = ""
		m.errorMessage = ""
		return m, nil
	}

	currentScreen := m.currentScreen()
	if currentScreen != nil && currentScreen.IsEditing() {
		cmd := currentScreen.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *ConfigWizardModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	contentWidth := m.width - configSidebarWidth - 4
	contentHeight := m.height - configHeaderHeight - configFooterHeight - 2

	if contentWidth < 30 {
		contentWidth = 30
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	for _, screen := range m.categoryScreens {
		screen.SetSize(contentWidth, contentHeight)
	}

	return m, nil
}

func (m *ConfigWizardModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirmQuit {
		return m.handleConfirmQuit(msg)
	}

	if m.showHelp {
		if msg.String() == "f1" || msg.String() == "esc" || msg.String() == "q" {
			m.showHelp = false
		}
		return m, nil
	}

	currentScreen := m.currentScreen()
	if currentScreen != nil && currentScreen.IsEditing() {
		cmd := currentScreen.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c", "ctrl+q":
		if m.hasUnsavedChanges() {
			m.confirmQuit = true
			return m, nil
		}
		return m, tea.Quit

	case "q":
		if m.hasUnsavedChanges() {
			m.confirmQuit = true
			return m, nil
		}
		return m, tea.Quit

	case "tab":
		m.nextCategory()

	case "shift+tab":
		m.prevCategory()

	case "ctrl+s":
		return m, m.saveChanges()

	case "f1":
		m.showHelp = !m.showHelp

	default:
		if currentScreen != nil {
			cmd := currentScreen.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *ConfigWizardModel) handleConfirmQuit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		return m, tea.Quit
	case "n", "N", "esc":
		m.confirmQuit = false
	case "s", "S":
		m.confirmQuit = false
		m.quitAfterSave = true
		return m, m.saveChanges()
	}
	return m, nil
}

func (m *ConfigWizardModel) nextCategory() {
	if m.activeCategory < len(m.categories)-1 {
		m.activeCategory++
	} else {
		m.activeCategory = 0
	}
}

func (m *ConfigWizardModel) prevCategory() {
	if m.activeCategory > 0 {
		m.activeCategory--
	} else {
		m.activeCategory = len(m.categories) - 1
	}
}

func (m *ConfigWizardModel) currentScreen() *screens.ConfigCategoryModel {
	if m.activeCategory >= 0 && m.activeCategory < len(m.categories) {
		return m.categoryScreens[m.categories[m.activeCategory]]
	}
	return nil
}

func (m *ConfigWizardModel) saveChanges() tea.Cmd {
	return func() tea.Msg {
		for _, cat := range m.categories {
			screen := m.categoryScreens[cat]
			changes := screen.GetChanges()
			for key, value := range changes {
				viper.Set(key, value)
			}
		}

		if err := viper.WriteConfig(); err != nil {
			if err := viper.SafeWriteConfig(); err != nil {
				return ConfigSaveErrorMsg{Err: err}
			}
		}

		return ConfigSavedMsg{}
	}
}

func (m *ConfigWizardModel) hasUnsavedChanges() bool {
	for _, cat := range m.categories {
		if m.categoryScreens[cat].HasUnsavedChanges() {
			return true
		}
	}
	return false
}

func (m *ConfigWizardModel) scheduleClearMessage() tea.Cmd {
	return tea.Tick(configMessageDuration, func(t time.Time) tea.Msg {
		return clearSavedMsgTick{}
	})
}

func (m *ConfigWizardModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	if m.confirmQuit {
		return m.renderConfirmQuit()
	}

	return m.renderMain()
}

func (m *ConfigWizardModel) renderMain() string {
	var content strings.Builder

	header := m.renderHeader()
	content.WriteString(header)
	content.WriteString("\n")

	sidebar := m.renderSidebar()
	mainContent := ""
	if screen := m.currentScreen(); screen != nil {
		mainContent = screen.View()
	}

	sidebarStyle := lipgloss.NewStyle().
		Width(configSidebarWidth).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 1)

	contentStyle := lipgloss.NewStyle().
		Padding(1, 2)

	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarStyle.Render(sidebar),
		contentStyle.Render(mainContent),
	)
	content.WriteString(combined)
	content.WriteString("\n")

	if m.savedMessage != "" {
		content.WriteString(styles.RenderSuccess(m.savedMessage))
		content.WriteString("\n")
	}
	if m.errorMessage != "" {
		content.WriteString(styles.RenderError(m.errorMessage))
		content.WriteString("\n")
	}

	footer := m.renderFooter()
	content.WriteString(footer)

	return content.String()
}

func (m *ConfigWizardModel) renderHeader() string {
	title := styles.TitleStyle.Render("Configuration Settings")
	return title + "\n" + styles.RenderSeparator(m.width-2)
}

func (m *ConfigWizardModel) renderSidebar() string {
	var sidebar strings.Builder

	sidebar.WriteString(styles.SubtitleStyle.Render("Categories"))
	sidebar.WriteString("\n\n")

	for i, cat := range m.categories {
		var line string
		if i == m.activeCategory {
			cursor := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("▶ ")
			catStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.BrightText)
			line = cursor + catStyle.Render(string(cat))
		} else {
			catStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
			line = "  " + catStyle.Render(string(cat))
		}

		if m.categoryScreens[cat].HasUnsavedChanges() {
			indicator := lipgloss.NewStyle().Foreground(styles.WarningColor).Render(" •")
			line += indicator
		}

		sidebar.WriteString(line)
		if i < len(m.categories)-1 {
			sidebar.WriteString("\n")
		}
	}

	return sidebar.String()
}

func (m *ConfigWizardModel) renderFooter() string {
	shortcuts := []string{
		"Tab: Next category",
		"Enter: Edit",
		"Ctrl+S: Save",
		"F1: Help",
		"q: Quit",
	}
	return styles.RenderFooter(shortcuts)
}

func (m *ConfigWizardModel) renderHelp() string {
	var help strings.Builder

	help.WriteString(styles.TitleStyle.Render("Help - Configuration Settings"))
	help.WriteString("\n\n")

	sections := []struct {
		title string
		items []string
	}{
		{
			title: "Navigation",
			items: []string{
				"Tab / Shift+Tab  - Switch between categories",
				"↑/k, ↓/j         - Navigate within category",
				"Home/g, End/G    - Jump to first/last item",
			},
		},
		{
			title: "Editing",
			items: []string{
				"Enter            - Edit selected item",
				"Esc              - Exit edit mode",
				"Space            - Toggle boolean values",
				"r                - Reset to default value",
			},
		},
		{
			title: "Actions",
			items: []string{
				"Ctrl+S           - Save all changes",
				"F1               - Toggle this help",
				"q / Ctrl+Q       - Quit (prompts if unsaved)",
			},
		},
	}

	for _, section := range sections {
		help.WriteString(styles.SubtitleStyle.Render(section.title))
		help.WriteString("\n")
		for _, item := range section.items {
			help.WriteString("  " + styles.HelpStyle.Render(item) + "\n")
		}
		help.WriteString("\n")
	}

	help.WriteString(styles.HelpStyle.Render("\nPress F1, Esc, or q to close help"))

	return styles.RenderBox(help.String(), "")
}

func (m *ConfigWizardModel) renderConfirmQuit() string {
	var confirm strings.Builder

	confirm.WriteString(styles.WarningStyle.Render("Unsaved Changes"))
	confirm.WriteString("\n\n")
	confirm.WriteString("You have unsaved changes. What would you like to do?\n\n")
	confirm.WriteString("  [S] Save and quit\n")
	confirm.WriteString("  [Y] Quit without saving\n")
	confirm.WriteString("  [N] Cancel and continue editing\n")

	return styles.RenderBox(confirm.String(), "")
}

func (m *ConfigWizardModel) ActiveCategory() int {
	return m.activeCategory
}

func (m *ConfigWizardModel) ShowingHelp() bool {
	return m.showHelp
}

func (m *ConfigWizardModel) ConfirmingQuit() bool {
	return m.confirmQuit
}

func (m *ConfigWizardModel) SavedMessage() string {
	return m.savedMessage
}

func (m *ConfigWizardModel) ErrorMessage() string {
	return m.errorMessage
}
