package ui

import (
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

// WindowsCompatibleTextArea provides Windows-compatible textarea with paste support
type WindowsCompatibleTextArea struct {
	textarea    textarea.Model
	manualText  string
	isEditing   bool
	placeholder string
	focused     bool
	width       int
	height      int
}

// WindowsCompatibleTextInput provides Windows-compatible textinput with enhanced handling
type WindowsCompatibleTextInput struct {
	input          textinput.Model
	manualText     string
	isEditing      bool
	placeholder    string
	focused        bool
	resetCount     int
	lastResetValue string
}

// NewWindowsCompatibleTextArea creates a new Windows-compatible textarea
func NewWindowsCompatibleTextArea(placeholder string) WindowsCompatibleTextArea {
	ta := textarea.New()
	ta.CharLimit = 0 // Enable UTF-8
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.Placeholder = placeholder

	return WindowsCompatibleTextArea{
		textarea:    ta,
		manualText:  "",
		isEditing:   false,
		placeholder: placeholder,
		focused:     false,
		width:       80,
		height:      10,
	}
}

// NewWindowsCompatibleTextInput creates a new Windows-compatible textinput
func NewWindowsCompatibleTextInput(placeholder string) WindowsCompatibleTextInput {
	input := textinput.New()

	// Critical Windows workaround sequence
	input.Reset()
	input.SetValue("")
	input.Reset()
	input.Blur() // Start unfocused
	input.Placeholder = placeholder
	input.Width = 40

	return WindowsCompatibleTextInput{
		input:          input,
		manualText:     "",
		isEditing:      false,
		placeholder:    placeholder,
		focused:        false,
		resetCount:     0,
		lastResetValue: "",
	}
}

// WindowsCompatibleTextArea methods

// SetSize sets the dimensions of the textarea
func (w *WindowsCompatibleTextArea) SetSize(width, height int) {
	w.width = width
	w.height = height
	w.textarea.SetWidth(width)
	w.textarea.SetHeight(height)
}

// SetValue sets the textarea value
func (w *WindowsCompatibleTextArea) SetValue(value string) {
	w.textarea.SetValue(value)
	w.manualText = value
}

// Value returns the current textarea value
func (w WindowsCompatibleTextArea) Value() string {
	if w.isEditing {
		return w.manualText
	}
	return w.textarea.Value()
}

// Focus sets the textarea to focused state
func (w *WindowsCompatibleTextArea) Focus() tea.Cmd {
	w.focused = true
	return w.textarea.Focus()
}

// Blur sets the textarea to unfocused state
func (w *WindowsCompatibleTextArea) Blur() {
	w.focused = false
	w.textarea.Blur()
}

// Focused returns whether the textarea is focused
func (w WindowsCompatibleTextArea) Focused() bool {
	return w.focused
}

// Update handles textarea updates with Windows compatibility
func (w WindowsCompatibleTextArea) Update(msg tea.Msg) (WindowsCompatibleTextArea, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if w.focused {
				// Toggle editing mode for Windows workaround
				w.isEditing = !w.isEditing
				if !w.isEditing {
					// Apply manual text to textarea
					w.textarea.SetValue(w.manualText)
				} else {
					// Start manual tracking
					w.manualText = w.textarea.Value()
				}
			}
			return w, nil

		case "esc":
			if w.isEditing {
				// Cancel manual editing and revert
				w.isEditing = false
				w.manualText = w.textarea.Value()
			}
			return w, nil
		}

		// Handle manual input for Windows compatibility
		if w.focused && w.isEditing {
			return w.handleManualInput(msg)
		}
	}

	// Standard BubbleTea handling for non-editing mode or non-Windows
	if w.focused && !w.isEditing {
		var cmd tea.Cmd
		w.textarea, cmd = w.textarea.Update(msg)
		return w, cmd
	}

	return w, nil
}

// handleManualInput processes manual text input for Windows compatibility
func (w WindowsCompatibleTextArea) handleManualInput(msg tea.KeyMsg) (WindowsCompatibleTextArea, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		if len(w.manualText) > 0 {
			w.manualText = w.manualText[:len(w.manualText)-1]
		}
	case "enter":
		w.manualText += "\n"
	case "ctrl+c":
		// Allow copy operations to pass through
		return w, nil
	case "ctrl+v":
		// Handle paste in manual mode - Windows compatibility with real clipboard
		if w.isEditing {
			// Initialize clipboard if not already done
			if err := clipboard.Init(); err == nil {
				// Read from clipboard
				clipboardText := clipboard.Read(clipboard.FmtText)
				if len(clipboardText) > 0 {
					// Sanitize clipboard content for safe terminal usage
					pastedText := SanitizeInputText(string(clipboardText))
					w.manualText += pastedText
				}
			} else {
				// Fallback message if clipboard access fails
				w.manualText += "[Clipboard access failed - use Right-click paste]"
			}
		}
		return w, nil
	default:
		// Filter for safe characters - accept printable ASCII and common symbols
		char := msg.String()
		if len(char) == 1 {
			// Accept printable ASCII range and common extended characters
			r := rune(char[0])
			if r >= 32 && r <= 126 || r >= 160 && r <= 255 {
				w.manualText += char
			}
		}
	}
	return w, nil
}

// View renders the textarea with Windows compatibility indicators
func (w WindowsCompatibleTextArea) View() string {
	if w.isEditing {
		// Show manual text with cursor indicator and editing mode notice
		lines := strings.Split(w.manualText, "\n")
		if len(lines) == 0 {
			lines = []string{""}
		}

		// Add cursor to last line
		lastLineIdx := len(lines) - 1
		lines[lastLineIdx] += "█"

		content := strings.Join(lines, "\n")

		// Add editing mode indicator
		if runtime.GOOS == "windows" {
			indicator := lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Render(" [Windows Edit Mode - Tab to exit] ")
			content = indicator + "\n" + content
		}

		return content
	}

	return w.textarea.View()
}

// WindowsCompatibleTextInput methods

// SetWidth sets the input width
func (w *WindowsCompatibleTextInput) SetWidth(width int) {
	w.input.Width = width
}

// SetValue sets the input value with Windows compatibility
func (w *WindowsCompatibleTextInput) SetValue(value string) {
	// Enhanced Windows reset sequence
	if runtime.GOOS == "windows" && value != w.lastResetValue {
		w.input.Reset()
		w.input.SetValue("")
		w.input.Reset()
		w.resetCount++
		w.lastResetValue = value
	}

	w.input.SetValue(value)
	w.manualText = value
}

// Value returns the current input value
func (w WindowsCompatibleTextInput) Value() string {
	if w.isEditing {
		return w.manualText
	}
	return w.input.Value()
}

// Focus sets the input to focused state
func (w *WindowsCompatibleTextInput) Focus() tea.Cmd {
	w.focused = true
	return w.input.Focus()
}

// Blur sets the input to unfocused state
func (w *WindowsCompatibleTextInput) Blur() {
	w.focused = false
	w.input.Blur()
}

// Focused returns whether the input is focused
func (w WindowsCompatibleTextInput) Focused() bool {
	return w.focused
}

// Update handles input updates with Windows compatibility
func (w WindowsCompatibleTextInput) Update(msg tea.Msg) (WindowsCompatibleTextInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if w.focused {
				// Toggle editing mode for Windows workaround
				w.isEditing = !w.isEditing
				if !w.isEditing {
					// Apply manual text to input
					w.input.SetValue(w.manualText)
				} else {
					// Start manual tracking
					w.manualText = w.input.Value()
				}
			}
			return w, nil

		case "esc":
			if w.isEditing {
				// Cancel manual editing and revert
				w.isEditing = false
				w.manualText = w.input.Value()
			}
			return w, nil
		}

		// Handle manual input for Windows compatibility
		if w.focused && w.isEditing {
			return w.handleManualTextInput(msg)
		}
	}

	// Standard BubbleTea handling for non-editing mode
	if w.focused && !w.isEditing {
		var cmd tea.Cmd
		w.input, cmd = w.input.Update(msg)
		return w, cmd
	}

	return w, nil
}

// handleManualTextInput processes manual text input for single-line input
func (w WindowsCompatibleTextInput) handleManualTextInput(msg tea.KeyMsg) (WindowsCompatibleTextInput, tea.Cmd) {
	switch msg.String() {
	case "backspace":
		if len(w.manualText) > 0 {
			w.manualText = w.manualText[:len(w.manualText)-1]
		}
	case "enter":
		// For single-line input, enter should exit editing mode
		w.isEditing = false
		w.input.SetValue(w.manualText)
		return w, nil
	case "ctrl+c":
		// Allow copy operations to pass through
		return w, nil
	case "ctrl+v":
		// Handle paste in manual mode - Windows compatibility with real clipboard
		if w.isEditing {
			// Initialize clipboard if not already done
			if err := clipboard.Init(); err == nil {
				// Read from clipboard
				clipboardText := clipboard.Read(clipboard.FmtText)
				if len(clipboardText) > 0 {
					// Sanitize clipboard content for safe terminal usage (single line)
					pastedText := SanitizeInputText(string(clipboardText))
					// Remove newlines for single-line input
					pastedText = strings.ReplaceAll(pastedText, "\n", " ")
					pastedText = strings.ReplaceAll(pastedText, "\r", " ")
					w.manualText += pastedText
				}
			} else {
				// Fallback message if clipboard access fails
				w.manualText += "[Clipboard access failed - use Right-click paste]"
			}
		}
		return w, nil
	default:
		// Filter for safe characters - single line, no newlines
		char := msg.String()
		if len(char) == 1 {
			// Accept printable ASCII range and common extended characters, but no newlines
			r := rune(char[0])
			if r >= 32 && r <= 126 || r >= 160 && r <= 255 {
				w.manualText += char
			}
		}
	}
	return w, nil
}

// View renders the input with Windows compatibility indicators
func (w WindowsCompatibleTextInput) View() string {
	if w.isEditing {
		// Show manual text with cursor indicator
		content := w.manualText + "█"

		// Add editing mode indicator for Windows
		if runtime.GOOS == "windows" {
			indicator := lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Render(" [Edit Mode] ")
			return content + indicator
		}

		return content
	}

	return w.input.View()
}

// Utility functions for Windows input handling

// IsWindowsInputWorkaroundNeeded checks if Windows-specific workarounds are needed
func IsWindowsInputWorkaroundNeeded() bool {
	return runtime.GOOS == "windows"
}

// CreateSafeTextInput creates a textinput with Windows compatibility
func CreateSafeTextInput(placeholder string) textinput.Model {
	input := textinput.New()

	if IsWindowsInputWorkaroundNeeded() {
		// Enhanced Windows workaround sequence
		input.Reset()
		input.SetValue("")
		input.Reset()
		input.Blur() // Start unfocused
	}

	input.Placeholder = placeholder
	return input
}

// ApplyWindowsInputFix applies Windows-specific fixes to an existing textinput
func ApplyWindowsInputFix(input *textinput.Model) {
	if IsWindowsInputWorkaroundNeeded() {
		input.Reset()
		input.SetValue("")
		input.Reset()
		input.Blur()
	}
}

// ValidateInputCharacter validates if a character is safe for Windows terminals
func ValidateInputCharacter(char string) bool {
	if len(char) != 1 {
		return false
	}

	r := rune(char[0])

	// Accept printable ASCII and common extended characters
	// Avoid control characters that can cause issues on Windows
	return (r >= 32 && r <= 126) || (r >= 160 && r <= 255)
}

// SanitizeInputText removes problematic characters for Windows compatibility
func SanitizeInputText(text string) string {
	var result strings.Builder
	for _, r := range text {
		char := string(r)
		if ValidateInputCharacter(char) {
			result.WriteString(char)
		}
	}
	return result.String()
}
