package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// NumberedTextArea is a textarea component (numbering functionality removed)
type NumberedTextArea struct {
	textarea textarea.Model
}

// NewNumberedTextArea creates a new textarea component
func NewNumberedTextArea() NumberedTextArea {
	ta := textarea.New()
	ta.ShowLineNumbers = false

	// Enable Unicode/UTF-8 support for accented characters
	ta.CharLimit = 0 // No character limit to avoid Unicode issues

	return NumberedTextArea{
		textarea: ta,
	}
}

// Update handles events
func (nta NumberedTextArea) Update(msg tea.Msg) (NumberedTextArea, tea.Cmd) {
	var cmd tea.Cmd
	nta.textarea, cmd = nta.textarea.Update(msg)
	return nta, cmd
}

// View renders the textarea
func (nta NumberedTextArea) View() string {
	return nta.textarea.View()
}

// Value returns the current textarea value
func (nta NumberedTextArea) Value() string {
	return nta.textarea.Value()
}

// SetValue sets the textarea value
func (nta *NumberedTextArea) SetValue(value string) {
	nta.textarea.SetValue(value)
}

// SetWidth sets the textarea width
func (nta *NumberedTextArea) SetWidth(width int) {
	nta.textarea.SetWidth(width)
}

// SetHeight sets the textarea height
func (nta *NumberedTextArea) SetHeight(height int) {
	nta.textarea.SetHeight(height)
}

// Focus focuses the textarea
func (nta *NumberedTextArea) Focus() tea.Cmd {
	return nta.textarea.Focus()
}

// Blur removes focus from the textarea
func (nta NumberedTextArea) Blur() {
	nta.textarea.Blur()
}

// Focused returns whether the textarea is focused
func (nta NumberedTextArea) Focused() bool {
	return nta.textarea.Focused()
}

// SetPlaceholder sets the placeholder text
func (nta *NumberedTextArea) SetPlaceholder(placeholder string) {
	nta.textarea.Placeholder = placeholder
}

// SetFullScreenMode adjusts the textarea for full-screen usage
func (nta *NumberedTextArea) SetFullScreenMode(terminalWidth, terminalHeight int) {
	// Use most of the terminal width, leaving some margin
	width := terminalWidth - 10
	if width < 40 {
		width = 40
	}
	if width > 120 {
		width = 120
	}

	// Use a good portion of terminal height for input
	height := terminalHeight / 3
	if height < 6 {
		height = 6
	}
	if height > 15 {
		height = 15
	}

	nta.SetWidth(width)
	nta.SetHeight(height)
}
