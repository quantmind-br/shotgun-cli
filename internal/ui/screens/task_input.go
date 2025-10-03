package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const maxTaskLength = 2000

type TaskInputModel struct {
	textarea textarea.Model
	width    int
	height   int
	focused  bool
}

func NewTaskInput(initialValue string) *TaskInputModel {
	ta := textarea.New()
	ta.Placeholder = "Describe the task you want to accomplish...\n\n" +
		"Example:\n" +
		"- Add a new user authentication feature\n" +
		"- Fix the memory leak in the data processor\n" +
		"- Refactor the payment handling code to use new API"
	ta.Focus()
	ta.CharLimit = maxTaskLength
	ta.SetValue(initialValue)
	ta.ShowLineNumbers = false // Disable line numbers for cleaner display

	// Configure textarea styles for better visibility on dark backgrounds
	textColor := lipgloss.Color("#ECEFF4")        // Light gray/white for text
	cursorColor := lipgloss.Color("#A3BE8C")      // Green for cursor
	placeholderColor := lipgloss.Color("#5C7E8C") // Muted color for placeholder

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

	return &TaskInputModel{
		textarea: ta,
		focused:  true,
	}
}

func (m *TaskInputModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate available space for textarea
	availableHeight := height - 8 // Reserve space for header, character count, footer
	availableWidth := width - 4   // Reserve space for margins

	if availableHeight < 5 {
		availableHeight = 5
	}
	if availableWidth < 20 {
		availableWidth = 20
	}

	m.textarea.SetWidth(availableWidth)
	m.textarea.SetHeight(availableHeight)
}

func (m *TaskInputModel) Update(msg tea.KeyMsg) (string, tea.Cmd) {
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

func (m *TaskInputModel) View() string {
	header := styles.RenderHeader(3, "Describe Your Task")

	// Character count with color coding
	currentLength := len(m.textarea.Value())
	charCountColor := styles.HelpStyle
	if currentLength > maxTaskLength*8/10 { // 80% of limit
		charCountColor = styles.ErrorStyle
	} else if currentLength > maxTaskLength*6/10 { // 60% of limit
		charCountColor = charCountColor.Foreground(styles.WarningColor)
	}

	charCount := charCountColor.Render(fmt.Sprintf("Characters: %d/%d", currentLength, maxTaskLength))

	instructions := styles.HelpStyle.Render(
		"Enter a detailed description of what you want to accomplish. " +
			"Be specific about requirements, constraints, and expected outcomes.")

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")
	content.WriteString(instructions)
	content.WriteString("\n\n")
	content.WriteString(m.textarea.View())
	content.WriteString("\n\n")
	content.WriteString(charCount)

	// Validation message
	if currentLength == 0 {
		content.WriteString("\n")
		content.WriteString(styles.ErrorStyle.Render("⚠ Task description is required to continue"))
	}

	shortcuts := []string{
		"Esc: Focus/Unfocus",
		"F8: Next",
		"F10: Back",
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(shortcuts)
	content.WriteString("\n\n")
	content.WriteString(footer)

	return content.String()
}

func (m *TaskInputModel) GetValue() string {
	return m.textarea.Value()
}

func (m *TaskInputModel) IsValid() bool {
	return len(strings.TrimSpace(m.textarea.Value())) > 0
}

func (m *TaskInputModel) IsFocused() bool {
	return m.textarea.Focused()
}
