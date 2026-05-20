package screens

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	rulesInputHeaderFooterHeight = 15
	rulesInputHorizontalPadding  = 6
	rulesInputMinHeight          = 5
	rulesInputMinWidth           = 20
)

type RulesInputModel struct {
	textarea textarea.Model
	width    int
	height   int
	focused  bool
}

func NewRulesInput(initialValue string) *RulesInputModel {
	ta := textarea.New()
	ta.Placeholder = "Add rules or constraints (optional)..."
	ta.Focus()
	ta.SetValue(initialValue)
	ta.ShowLineNumbers = false // Disable line numbers for cleaner display

	// Configure textarea styles with Nord colors for better visibility
	textColor := styles.TextColor
	cursorColor := styles.AccentColor
	placeholderColor := styles.DimText

	// Modify existing styles instead of replacing them
	ta.FocusedStyle.Text = ta.FocusedStyle.Text.Foreground(textColor).UnsetBackground()
	ta.FocusedStyle.Placeholder = ta.FocusedStyle.Placeholder.Foreground(placeholderColor).UnsetBackground()
	ta.FocusedStyle.Base = ta.FocusedStyle.Base.Foreground(textColor).UnsetBackground()
	ta.FocusedStyle.CursorLine = ta.FocusedStyle.CursorLine.UnsetBackground()
	ta.BlurredStyle.Text = ta.BlurredStyle.Text.Foreground(textColor).UnsetBackground()
	ta.BlurredStyle.Placeholder = ta.BlurredStyle.Placeholder.Foreground(placeholderColor).UnsetBackground()
	ta.BlurredStyle.Base = ta.BlurredStyle.Base.Foreground(textColor).UnsetBackground()
	ta.Cursor.Style = ta.Cursor.Style.Foreground(cursorColor)
	ta.Prompt = "" // Remove prompt to avoid clutter

	return &RulesInputModel{
		textarea: ta,
		focused:  true,
	}
}

func (m *RulesInputModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate available space for textarea
	availableHeight := height - rulesInputHeaderFooterHeight
	availableWidth := width - rulesInputHorizontalPadding

	if availableHeight < rulesInputMinHeight {
		availableHeight = rulesInputMinHeight
	}
	if availableWidth < rulesInputMinWidth {
		availableWidth = rulesInputMinWidth
	}

	m.textarea.SetWidth(availableWidth)
	m.textarea.SetHeight(availableHeight)
}

func (m *RulesInputModel) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	var cmd tea.Cmd

	switch keyMsg.String() {
	case "tab":
		if m.textarea.Focused() {
			m.textarea.Blur()
			m.focused = false
		} else {
			m.textarea.Focus()
			m.focused = true
		}
	case keyEsc:
		if m.textarea.Focused() {
			m.textarea.Blur()
			m.focused = false
		}
	default:
		m.textarea, cmd = m.textarea.Update(keyMsg)
	}

	return cmd
}

func (m *RulesInputModel) View() string {
	header := styles.RenderHeader(4, "Add Rules & Constraints")

	// Character count with styling
	currentLength := utf8.RuneCountInString(m.textarea.Value())
	var charCountStyle lipgloss.Style
	if currentLength == 0 {
		charCountStyle = lipgloss.NewStyle().Foreground(styles.MutedColor)
	} else {
		charCountStyle = lipgloss.NewStyle().Foreground(styles.TextColor)
	}
	charCount := charCountStyle.Render(fmt.Sprintf("Characters: %d", currentLength))

	instructions := styles.HelpStyle.Render(
		"Specify any coding standards, architectural constraints, or specific requirements. " +
			"This step is optional - you can leave it empty and proceed to the next step.")

	// Optional badge
	optionalBadge := lipgloss.NewStyle().
		Foreground(styles.Nord15).
		Bold(true).
		Render("OPTIONAL")

	optionalNote := styles.HelpStyle.Render("💡 This step is optional. Press F8 to skip or F7 to go back.")

	// Wrap textarea in a border that changes based on focus state
	var textareaView string
	if m.textarea.Focused() {
		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.PrimaryColor).
			Padding(0, 1)
		textareaView = borderStyle.Render(m.textarea.View())
	} else {
		borderStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.MutedColor).
			Padding(0, 1)
		textareaView = borderStyle.Render(m.textarea.View())
	}

	// Focus indicator
	var focusIndicator string
	if m.textarea.Focused() {
		focusIndicator = styles.StatusActiveStyle.Render("● Editing")
	} else {
		focusIndicator = styles.StatusInactiveStyle.Render("○ Press Tab to edit")
	}

	var body strings.Builder
	body.WriteString(instructions)
	body.WriteString("\n")
	body.WriteString(optionalNote)
	body.WriteString("\n\n")
	body.WriteString(focusIndicator)
	body.WriteString("\n\n")
	body.WriteString(textareaView)
	body.WriteString("\n\n")
	body.WriteString(charCount)

	bodyStr := body.String()
	footer := m.renderFooter()

	headerLine := header + "  " + optionalBadge
	headerHeight := strings.Count(headerLine, "\n") + 1
	bodyHeight := strings.Count(bodyStr, "\n") + 1
	footerHeight := strings.Count(footer, "\n") + 1
	// Always keep at least one blank line so the char-count line never glues
	// onto the footer hints.
	paddingLines := max(m.height-headerHeight-bodyHeight-footerHeight-2, 1)

	var content strings.Builder
	content.WriteString(headerLine)
	content.WriteString("\n\n")
	content.WriteString(bodyStr)
	content.WriteString(strings.Repeat("\n", paddingLines))
	content.WriteString(footer)

	return content.String()
}

func (m *RulesInputModel) renderFooter() string {
	return styles.RenderStatusBar(m.width, [][]string{
		{"Type: Enter text", "Tab: Edit/Done"},
		{"F7: Back", "F8: Next (Skip)", "F1: Help", "Ctrl+Q/Ctrl+C: Quit"},
	})
}

func (m *RulesInputModel) GetValue() string {
	return m.textarea.Value()
}

func (m *RulesInputModel) IsValid() bool {
	// Rules are always valid since they're optional
	return true
}

func (m *RulesInputModel) IsFocused() bool {
	return m.textarea.Focused()
}
