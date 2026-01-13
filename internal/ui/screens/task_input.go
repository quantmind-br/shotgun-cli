package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	taskInputHeaderFooterHeight = 10
	taskInputHorizontalPadding  = 6
	taskInputMinHeight          = 5
	taskInputMinWidth           = 20
)

type TaskInputModel struct {
	textarea       textarea.Model
	width          int
	height         int
	focused        bool
	willSkipToNext bool // true if F8 will skip Rules and go directly to Review
}

func NewTaskInput(initialValue string) *TaskInputModel {
	ta := textarea.New()
	ta.Placeholder = "Describe the task you want to accomplish...\n\n" +
		"Example:\n" +
		"- Add a new user authentication feature\n" +
		"- Fix the memory leak in the data processor\n" +
		"- Refactor the payment handling code to use new API"
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

	return &TaskInputModel{
		textarea: ta,
		focused:  true,
	}
}

func (m *TaskInputModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate available space for textarea
	availableHeight := height - taskInputHeaderFooterHeight
	availableWidth := width - taskInputHorizontalPadding

	if availableHeight < taskInputMinHeight {
		availableHeight = taskInputMinHeight
	}
	if availableWidth < taskInputMinWidth {
		availableWidth = taskInputMinWidth
	}

	m.textarea.SetWidth(availableWidth)
	m.textarea.SetHeight(availableHeight)
}

func (m *TaskInputModel) Update(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	var cmd tea.Cmd

	switch keyMsg.String() {
	case keyEsc:
		if m.textarea.Focused() {
			m.textarea.Blur()
			m.focused = false
		} else {
			m.textarea.Focus()
			m.focused = true
		}
	default:
		m.textarea, cmd = m.textarea.Update(keyMsg)
	}

	return cmd
}

func (m *TaskInputModel) View() string {
	header := styles.RenderHeader(3, "Describe Your Task")

	// Character count with color based on length
	currentLength := len(m.textarea.Value())
	var charCountStyle lipgloss.Style
	if currentLength == 0 {
		charCountStyle = lipgloss.NewStyle().Foreground(styles.WarningColor)
	} else if currentLength < 50 {
		charCountStyle = lipgloss.NewStyle().Foreground(styles.TextColor)
	} else {
		charCountStyle = lipgloss.NewStyle().Foreground(styles.SuccessColor)
	}
	charCount := charCountStyle.Render(fmt.Sprintf("Characters: %d", currentLength))

	instructions := styles.HelpStyle.Render(
		"Enter a detailed description of what you want to accomplish. " +
			"Be specific about requirements, constraints, and expected outcomes.")

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
	content.WriteString("\n\n")
	content.WriteString(instructions)
	content.WriteString("\n\n")
	content.WriteString(focusIndicator)
	content.WriteString("\n\n")
	content.WriteString(textareaView)
	content.WriteString("\n\n")
	content.WriteString(charCount)

	// Validation message
	if currentLength == 0 {
		content.WriteString("\n")
		content.WriteString(styles.RenderWarning("Task description is required to continue"))
	}

	line1 := []string{
		"Type: Enter text",
		"Esc: Edit/Done",
	}

	nextAction := "F8: Next"
	if m.willSkipToNext {
		nextAction = "F8: Skip to Review"
	}

	line2 := []string{
		"F7: Back",
		nextAction,
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(line1) + "\n" + styles.RenderFooter(line2)
	content.WriteString("\n\n")
	content.WriteString(footer)

	return content.String()
}

func (m *TaskInputModel) GetValue() string {
	return m.textarea.Value()
}

func (m *TaskInputModel) SetValueForTest(value string) {
	m.textarea.SetValue(value)
}

func (m *TaskInputModel) IsValid() bool {
	return len(strings.TrimSpace(m.textarea.Value())) > 0
}

func (m *TaskInputModel) IsFocused() bool {
	return m.textarea.Focused()
}

// SetWillSkipToReview sets whether F8 will skip rules and go directly to Review
func (m *TaskInputModel) SetWillSkipToReview(skip bool) {
	m.willSkipToNext = skip
}
