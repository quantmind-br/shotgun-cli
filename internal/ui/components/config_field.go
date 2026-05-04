package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type ConfigFieldModel struct {
	metadata    config.ConfigMetadata
	value       string
	input       textinput.Model
	focused     bool
	width       int
	err         error
	modified    bool
	placeholder string
}

func NewConfigField(metadata config.ConfigMetadata, currentValue string) *ConfigFieldModel {
	ti := textinput.New()
	ti.Placeholder = getPlaceholder(metadata)
	ti.CharLimit = 256
	ti.Width = 40

	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.PrimaryColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.TextColor)
	ti.PlaceholderStyle = styles.InputPlaceholderStyle
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	ti.SetValue(currentValue)

	return &ConfigFieldModel{
		metadata:    metadata,
		value:       currentValue,
		input:       ti,
		placeholder: ti.Placeholder,
	}
}

func getPlaceholder(m config.ConfigMetadata) string {
	switch m.Type {
	case config.TypeSize:
		return "e.g., 10MB, 500KB"
	case config.TypePath:
		return "/path/to/file"
	case config.TypeURL:
		return "https://api.example.com"
	case config.TypeInt:
		return "Enter a number"
	case config.TypeTimeout:
		return "Timeout in seconds"
	default:
		return ""
	}
}

func (m *ConfigFieldModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ConfigFieldModel) Update(msg tea.Msg) (*ConfigFieldModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	newValue := m.input.Value()
	if newValue != m.value {
		m.modified = true
		m.err = config.ValidateValue(m.metadata.Key, newValue)
	}

	return m, cmd
}

func (m *ConfigFieldModel) Focus() tea.Cmd {
	m.focused = true
	return m.input.Focus()
}

func (m *ConfigFieldModel) Blur() {
	m.focused = false
	m.input.Blur()
}

func (m *ConfigFieldModel) IsFocused() bool {
	return m.focused
}

func (m *ConfigFieldModel) SetWidth(width int) {
	m.width = width
	inputWidth := width - 4
	if inputWidth < 20 {
		inputWidth = 20
	}
	if inputWidth > 60 {
		inputWidth = 60
	}
	m.input.Width = inputWidth
}

func (m *ConfigFieldModel) Value() string {
	return m.input.Value()
}

func (m *ConfigFieldModel) SetValue(value string) {
	m.input.SetValue(value)
	m.value = value
	m.modified = false
	m.err = nil
}

func (m *ConfigFieldModel) Error() error {
	return m.err
}

func (m *ConfigFieldModel) IsModified() bool {
	return m.modified
}

func (m *ConfigFieldModel) IsValid() bool {
	return m.err == nil
}

func (m *ConfigFieldModel) Metadata() config.ConfigMetadata {
	return m.metadata
}

func (m *ConfigFieldModel) View() string {
	var b strings.Builder

	labelStyle := styles.InputLabelStyle
	if m.focused {
		labelStyle = labelStyle.Foreground(styles.PrimaryColor)
	}
	b.WriteString(labelStyle.Render(m.metadata.Key))
	b.WriteString("\n")

	descStyle := lipgloss.NewStyle().Foreground(styles.MutedColor).Italic(true)
	b.WriteString(descStyle.Render(m.metadata.Description))
	b.WriteString("\n")

	if m.focused {
		b.WriteString(styles.FocusedBorderStyle.
			Width(m.input.Width + 2).
			Padding(0).
			Render(m.input.View()))
	} else {
		b.WriteString(styles.BlurredBorderStyle.
			Width(m.input.Width + 2).
			Padding(0).
			Render(m.input.View()))
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(styles.RenderError(m.err.Error()))
	} else if m.modified {
		b.WriteString("\n")
		b.WriteString(styles.RenderSuccess("Valid"))
	}

	return b.String()
}
