package ui

import (
	"runtime"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// NumberedTextArea is a textarea component (numbering functionality removed)
// Now enhanced with Windows compatibility
type NumberedTextArea struct {
	textarea textarea.Model
	// Windows compatibility fields
	windowsCompat  *WindowsCompatibleTextArea
	useWindowsMode bool
}

// NewNumberedTextArea creates a new textarea component with Windows compatibility
func NewNumberedTextArea() NumberedTextArea {
	ta := textarea.New()
	ta.ShowLineNumbers = false

	// Enable Unicode/UTF-8 support for accented characters
	ta.CharLimit = 0 // No character limit to avoid Unicode issues

	// Initialize Windows compatibility if needed
	var windowsCompat *WindowsCompatibleTextArea
	useWindowsMode := runtime.GOOS == "windows"

	if useWindowsMode {
		winCompat := NewWindowsCompatibleTextArea("Enter your text here...")
		windowsCompat = &winCompat
	}

	return NumberedTextArea{
		textarea:       ta,
		windowsCompat:  windowsCompat,
		useWindowsMode: useWindowsMode,
	}
}

// Update handles events with Windows compatibility
func (nta NumberedTextArea) Update(msg tea.Msg) (NumberedTextArea, tea.Cmd) {
	var cmd tea.Cmd

	// Use Windows-compatible mode on Windows systems
	if nta.useWindowsMode && nta.windowsCompat != nil {
		updatedCompat, compatCmd := nta.windowsCompat.Update(msg)
		nta.windowsCompat = &updatedCompat
		return nta, compatCmd
	}

	// Standard BubbleTea handling for other platforms
	nta.textarea, cmd = nta.textarea.Update(msg)
	return nta, cmd
}

// View renders the textarea with Windows compatibility
func (nta NumberedTextArea) View() string {
	// Use Windows-compatible rendering on Windows systems
	if nta.useWindowsMode && nta.windowsCompat != nil {
		return nta.windowsCompat.View()
	}

	// Standard BubbleTea rendering for other platforms
	return nta.textarea.View()
}

// Value returns the current textarea value with Windows compatibility
func (nta NumberedTextArea) Value() string {
	// Use Windows-compatible value retrieval on Windows systems
	if nta.useWindowsMode && nta.windowsCompat != nil {
		return nta.windowsCompat.Value()
	}

	// Standard BubbleTea value retrieval for other platforms
	return nta.textarea.Value()
}

// SetValue sets the textarea value with Windows compatibility
func (nta *NumberedTextArea) SetValue(value string) {
	// Use Windows-compatible value setting on Windows systems
	if nta.useWindowsMode && nta.windowsCompat != nil {
		nta.windowsCompat.SetValue(value)
	}

	// Always update the standard textarea as well for compatibility
	nta.textarea.SetValue(value)
}

// SetWidth sets the textarea width with Windows compatibility
func (nta *NumberedTextArea) SetWidth(width int) {
	if nta.useWindowsMode && nta.windowsCompat != nil {
		nta.windowsCompat.SetSize(width, nta.windowsCompat.height)
	}
	nta.textarea.SetWidth(width)
}

// SetHeight sets the textarea height with Windows compatibility
func (nta *NumberedTextArea) SetHeight(height int) {
	if nta.useWindowsMode && nta.windowsCompat != nil {
		nta.windowsCompat.SetSize(nta.windowsCompat.width, height)
	}
	nta.textarea.SetHeight(height)
}

// Focus focuses the textarea with Windows compatibility
func (nta *NumberedTextArea) Focus() tea.Cmd {
	if nta.useWindowsMode && nta.windowsCompat != nil {
		return nta.windowsCompat.Focus()
	}
	return nta.textarea.Focus()
}

// Blur removes focus from the textarea with Windows compatibility
func (nta NumberedTextArea) Blur() {
	if nta.useWindowsMode && nta.windowsCompat != nil {
		nta.windowsCompat.Blur()
	}
	nta.textarea.Blur()
}

// Focused returns whether the textarea is focused with Windows compatibility
func (nta NumberedTextArea) Focused() bool {
	if nta.useWindowsMode && nta.windowsCompat != nil {
		return nta.windowsCompat.Focused()
	}
	return nta.textarea.Focused()
}

// SetPlaceholder sets the placeholder text with Windows compatibility
func (nta *NumberedTextArea) SetPlaceholder(placeholder string) {
	if nta.useWindowsMode && nta.windowsCompat != nil {
		nta.windowsCompat.placeholder = placeholder
	}
	nta.textarea.Placeholder = placeholder
}

// SetFullScreenMode adjusts the textarea for full-screen usage with Windows compatibility
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

// Windows-specific methods for enhanced functionality

// IsWindowsMode returns whether Windows compatibility mode is active
func (nta NumberedTextArea) IsWindowsMode() bool {
	return nta.useWindowsMode
}

// GetWindowsInstructions returns Windows-specific usage instructions
func (nta NumberedTextArea) GetWindowsInstructions() string {
	if !nta.useWindowsMode {
		return ""
	}
	return "Windows Mode: Press Tab to toggle edit mode for better paste support. Press Esc to cancel edits."
}

// ForceWindowsMode enables Windows compatibility mode regardless of OS
func (nta *NumberedTextArea) ForceWindowsMode(enable bool) {
	nta.useWindowsMode = enable
	if enable && nta.windowsCompat == nil {
		winCompat := NewWindowsCompatibleTextArea("Enter your text here...")
		nta.windowsCompat = &winCompat
	}
}
