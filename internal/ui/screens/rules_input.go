package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type RulesInputModel struct {
	textarea textarea.Model
	width    int
	height   int
	focused  bool
}

func NewRulesInput(initialValue string) *RulesInputModel {
	ta := textarea.New()
	ta.Placeholder = "Add any rules, constraints, or guidelines (optional)...\n\n" +
		"Examples:\n" +
		"- Use TypeScript instead of JavaScript\n" +
		"- Follow the existing error handling patterns\n" +
		"- Ensure all functions have unit tests\n" +
		"- Use the company's coding standards\n" +
		"- Maintain backward compatibility"
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
	availableHeight := height - 12 // Reserve space for header, instructions, character count, footer, border
	availableWidth := width - 6    // Reserve space for margins and border

	if availableHeight < 5 {
		availableHeight = 5
	}
	if availableWidth < 20 {
		availableWidth = 20
	}

	m.textarea.SetWidth(availableWidth)
	m.textarea.SetHeight(availableHeight)
}

func (m *RulesInputModel) Update(msg tea.KeyMsg) (string, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case keyEsc:
		if m.textarea.Focused() {
			m.textarea.Blur()
			m.focused = false
		} else {
			m.textarea.Focus()
			m.focused = true
		}
	default:
		m.textarea, cmd = m.textarea.Update(msg)
	}

	return m.textarea.Value(), cmd
}

func (m *RulesInputModel) View() string {
	header := styles.RenderHeader(4, "Add Rules & Constraints")

	// Character count with styling
	currentLength := len(m.textarea.Value())
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
		focusIndicator = styles.StatusInactiveStyle.Render("○ Press Esc to edit")
	}

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("  ")
	content.WriteString(optionalBadge)
	content.WriteString("\n\n")
	content.WriteString(instructions)
	content.WriteString("\n")
	content.WriteString(optionalNote)
	content.WriteString("\n\n")
	content.WriteString(focusIndicator)
	content.WriteString("\n\n")
	content.WriteString(textareaView)
	content.WriteString("\n\n")
	content.WriteString(charCount)

	line1 := []string{
		"Type: Enter text",
		"Esc: Edit/Done",
	}
	line2 := []string{
		"F7: Back",
		"F8: Next (Skip)",
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(line1) + "\n" + styles.RenderFooter(line2)
	content.WriteString("\n\n")
	content.WriteString(footer)

	return content.String()
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
