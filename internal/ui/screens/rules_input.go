package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const maxRulesLength = 1500

type RulesInputModel struct {
	textarea   textarea.Model
	width      int
	height     int
	focused    bool
}

func NewRulesInput(initialValue string) *RulesInputModel {
	ta := textarea.New()
	ta.Placeholder = "Add any rules, constraints, or guidelines (optional)...\n\nExamples:\n- Use TypeScript instead of JavaScript\n- Follow the existing error handling patterns\n- Ensure all functions have unit tests\n- Use the company's coding standards\n- Maintain backward compatibility"
	ta.Focus()
	ta.CharLimit = maxRulesLength
	ta.SetValue(initialValue)

	return &RulesInputModel{
		textarea: ta,
		focused:  true,
	}
}

func (m *RulesInputModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate available space for textarea
	availableHeight := height - 9 // Reserve space for header, instructions, character count, footer
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

func (m *RulesInputModel) Update(msg tea.KeyMsg) (string, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m.textarea.Value(), tea.Quit
	case "esc":
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
	header := styles.RenderHeader(4, "Add Rules & Constraints (Optional)")

	// Character count with color coding
	currentLength := len(m.textarea.Value())
	charCountColor := styles.HelpStyle
	if currentLength > maxRulesLength*8/10 { // 80% of limit
		charCountColor = styles.ErrorStyle
	} else if currentLength > maxRulesLength*6/10 { // 60% of limit
		charCountColor = charCountColor.Copy().Foreground(styles.WarningColor)
	}

	charCount := charCountColor.Render(fmt.Sprintf("Characters: %d/%d", currentLength, maxRulesLength))

	instructions := styles.HelpStyle.Render("Specify any coding standards, architectural constraints, or specific requirements. This step is optional - you can leave it empty and proceed to the next step.")

	optionalNote := styles.HelpStyle.Copy().Italic(true).Render("💡 This step is optional. Press F8 to skip or F10 to go back.")

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")
	content.WriteString(instructions)
	content.WriteString("\n")
	content.WriteString(optionalNote)
	content.WriteString("\n\n")
	content.WriteString(m.textarea.View())
	content.WriteString("\n\n")
	content.WriteString(charCount)

	shortcuts := []string{
		"Esc: Focus/Unfocus",
		"F8: Next (Skip)",
		"F10: Back",
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(shortcuts)
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