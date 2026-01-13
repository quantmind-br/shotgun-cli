package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type ConfigToggleModel struct {
	metadata config.ConfigMetadata
	value    bool
	focused  bool
	width    int
	modified bool
}

func NewConfigToggle(metadata config.ConfigMetadata, currentValue bool) *ConfigToggleModel {
	return &ConfigToggleModel{
		metadata: metadata,
		value:    currentValue,
	}
}

func (m *ConfigToggleModel) Init() tea.Cmd {
	return nil
}

func (m *ConfigToggleModel) Update(msg tea.Msg) (*ConfigToggleModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter", " ", "tab":
			m.value = !m.value
			m.modified = true
		case "y", "Y":
			if !m.value {
				m.value = true
				m.modified = true
			}
		case "n", "N":
			if m.value {
				m.value = false
				m.modified = true
			}
		}
	}

	return m, nil
}

func (m *ConfigToggleModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *ConfigToggleModel) Blur() {
	m.focused = false
}

func (m *ConfigToggleModel) IsFocused() bool {
	return m.focused
}

func (m *ConfigToggleModel) SetWidth(width int) {
	m.width = width
}

func (m *ConfigToggleModel) Value() bool {
	return m.value
}

func (m *ConfigToggleModel) SetValue(value bool) {
	m.value = value
	m.modified = false
}

func (m *ConfigToggleModel) IsModified() bool {
	return m.modified
}

func (m *ConfigToggleModel) Metadata() config.ConfigMetadata {
	return m.metadata
}

func (m *ConfigToggleModel) View() string {
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

	toggleView := m.renderToggle()

	if m.focused {
		b.WriteString(styles.FocusedBorderStyle.
			Padding(0, 1).
			Render(toggleView))
	} else {
		b.WriteString(styles.BlurredBorderStyle.
			Padding(0, 1).
			Render(toggleView))
	}

	return b.String()
}

func (m *ConfigToggleModel) renderToggle() string {
	var toggle string

	onStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)
	offStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)

	if m.value {
		onStyle = lipgloss.NewStyle().Foreground(styles.SuccessColor).Bold(true)
		toggle = "  " + offStyle.Render("OFF") + "  " +
			lipgloss.NewStyle().Foreground(styles.SuccessColor).Render("◉━━") +
			"  " + onStyle.Render("ON")
	} else {
		offStyle = lipgloss.NewStyle().Foreground(styles.ErrorColor).Bold(true)
		toggle = "  " + offStyle.Render("OFF") + "  " +
			lipgloss.NewStyle().Foreground(styles.ErrorColor).Render("━━◉") +
			"  " + onStyle.Render("ON")
	}

	hint := ""
	if m.focused {
		hint = lipgloss.NewStyle().Foreground(styles.MutedColor).Italic(true).
			Render("  (Space/Enter to toggle)")
	}

	return toggle + hint
}
