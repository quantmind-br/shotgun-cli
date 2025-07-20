package ui

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// NumberedTextArea is a textarea component with automatic numbering functionality
type NumberedTextArea struct {
	textarea         textarea.Model
	autoNumbering    bool
	lastLineCount    int
	numberingPattern *regexp.Regexp
}

// NewNumberedTextArea creates a new numbered textarea component
func NewNumberedTextArea() NumberedTextArea {
	ta := textarea.New()
	ta.ShowLineNumbers = false // We'll handle our own numbering
	
	// Compile regex for detecting numbered lines
	numberingPattern := regexp.MustCompile(`^\s*(\d+)\.\s*(.*)$`)
	
	return NumberedTextArea{
		textarea:         ta,
		autoNumbering:    true,
		lastLineCount:    0,
		numberingPattern: numberingPattern,
	}
}

// Update handles events and implements auto-numbering logic
func (nta NumberedTextArea) Update(msg tea.Msg) (NumberedTextArea, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if nta.autoNumbering {
				// Handle auto-numbering on Enter
				nta = nta.handleAutoNumbering()
			}
		}
	}
	
	// Update the underlying textarea
	nta.textarea, cmd = nta.textarea.Update(msg)
	
	// Track line count changes for numbering adjustments
	currentLines := len(strings.Split(nta.textarea.Value(), "\n"))
	if currentLines != nta.lastLineCount {
		nta.lastLineCount = currentLines
	}
	
	return nta, cmd
}

// handleAutoNumbering implements the auto-numbering logic when Enter is pressed
func (nta NumberedTextArea) handleAutoNumbering() NumberedTextArea {
	content := nta.textarea.Value()
	lines := strings.Split(content, "\n")
	
	if len(lines) == 0 {
		return nta
	}
	
	// Get the current cursor position info
	currentLine := len(lines) - 1
	if currentLine < 0 {
		currentLine = 0
	}
	
	// Check if we're continuing a numbered list
	if currentLine > 0 {
		previousLine := strings.TrimSpace(lines[currentLine-1])
		
		// Check if previous line was numbered
		if matches := nta.numberingPattern.FindStringSubmatch(previousLine); matches != nil {
			// Extract the number from previous line
			if num, err := strconv.Atoi(matches[1]); err == nil {
				nextNumber := num + 1
				
				// If current line is empty, add the next number
				currentLineContent := strings.TrimSpace(lines[currentLine])
				if currentLineContent == "" {
					lines[currentLine] = strconv.Itoa(nextNumber) + ". "
					newContent := strings.Join(lines, "\n")
					nta.textarea.SetValue(newContent)
				}
			}
		} else if strings.TrimSpace(lines[currentLine-1]) != "" {
			// Previous line wasn't numbered but had content
			// Check if user wants to start numbering
			currentLineContent := strings.TrimSpace(lines[currentLine])
			if currentLineContent == "" {
				// Start with number 1
				lines[currentLine] = "1. "
				newContent := strings.Join(lines, "\n")
				nta.textarea.SetValue(newContent)
			}
		}
	} else {
		// First line - start numbering if it's empty
		if strings.TrimSpace(lines[0]) == "" {
			lines[0] = "1. "
			newContent := strings.Join(lines, "\n")
			nta.textarea.SetValue(newContent)
		}
	}
	
	return nta
}

// GetNumberedValue returns the current value with proper numbering
func (nta *NumberedTextArea) GetNumberedValue() string {
	content := nta.textarea.Value()
	lines := strings.Split(content, "\n")
	
	var numberedLines []string
	currentNumber := 1
	
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Skip empty lines
		if trimmedLine == "" {
			numberedLines = append(numberedLines, "")
			continue
		}
		
		// Check if line is already numbered
		if matches := nta.numberingPattern.FindStringSubmatch(line); matches != nil {
			// Line is already numbered, use as-is but ensure sequential numbering
			numberedLines = append(numberedLines, strconv.Itoa(currentNumber)+". "+matches[2])
			currentNumber++
		} else {
			// Line is not numbered, add numbering if it has content
			if trimmedLine != "" {
				numberedLines = append(numberedLines, strconv.Itoa(currentNumber)+". "+trimmedLine)
				currentNumber++
			} else {
				numberedLines = append(numberedLines, line)
			}
		}
	}
	
	return strings.Join(numberedLines, "\n")
}

// View renders the textarea with visual numbering indicators
func (nta NumberedTextArea) View() string {
	// Add a subtle border to indicate this is a numbered textarea
	baseView := nta.textarea.View()
	
	if nta.autoNumbering {
		// Add a small indicator that this textarea supports auto-numbering
		style := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)
		
		return style.Render(baseView)
	}
	
	return baseView
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

// EnableAutoNumbering enables or disables auto-numbering
func (nta *NumberedTextArea) EnableAutoNumbering(enabled bool) {
	nta.autoNumbering = enabled
}

// IsAutoNumberingEnabled returns whether auto-numbering is enabled
func (nta NumberedTextArea) IsAutoNumberingEnabled() bool {
	return nta.autoNumbering
}