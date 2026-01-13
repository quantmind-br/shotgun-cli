package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type ConfigSelectModel struct {
	metadata config.ConfigMetadata
	options  []string
	selected int
	focused  bool
	width    int
	modified bool
	expanded bool
}

func NewConfigSelect(metadata config.ConfigMetadata, currentValue string) *ConfigSelectModel {
	options := metadata.EnumOptions
	if len(options) == 0 {
		options = []string{currentValue}
	}

	selected := 0
	for i, opt := range options {
		if opt == currentValue {
			selected = i
			break
		}
	}

	return &ConfigSelectModel{
		metadata: metadata,
		options:  options,
		selected: selected,
	}
}

func (m *ConfigSelectModel) Init() tea.Cmd {
	return nil
}

func (m *ConfigSelectModel) Update(msg tea.Msg) (*ConfigSelectModel, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter", " ":
			m.expanded = !m.expanded
		case "up", "k":
			if m.expanded {
				m.selected--
				if m.selected < 0 {
					m.selected = len(m.options) - 1
				}
				m.modified = true
			}
		case "down", "j":
			if m.expanded {
				m.selected++
				if m.selected >= len(m.options) {
					m.selected = 0
				}
				m.modified = true
			}
		case "esc":
			if m.expanded {
				m.expanded = false
			}
		case "tab":
			if m.expanded {
				m.expanded = false
			} else {
				m.selected = (m.selected + 1) % len(m.options)
				m.modified = true
			}
		}
	}

	return m, nil
}

func (m *ConfigSelectModel) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *ConfigSelectModel) Blur() {
	m.focused = false
	m.expanded = false
}

func (m *ConfigSelectModel) IsFocused() bool {
	return m.focused
}

func (m *ConfigSelectModel) SetWidth(width int) {
	m.width = width
}

func (m *ConfigSelectModel) Value() string {
	if m.selected >= 0 && m.selected < len(m.options) {
		return m.options[m.selected]
	}
	return ""
}

func (m *ConfigSelectModel) SetValue(value string) {
	for i, opt := range m.options {
		if opt == value {
			m.selected = i
			m.modified = false
			return
		}
	}
	if len(m.options) > 0 {
		m.selected = 0
	}
}

func (m *ConfigSelectModel) IsModified() bool {
	return m.modified
}

func (m *ConfigSelectModel) Metadata() config.ConfigMetadata {
	return m.metadata
}

func (m *ConfigSelectModel) View() string {
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

	selectView := m.renderSelect()

	if m.focused {
		b.WriteString(styles.FocusedBorderStyle.
			Padding(0, 1).
			Render(selectView))
	} else {
		b.WriteString(styles.BlurredBorderStyle.
			Padding(0, 1).
			Render(selectView))
	}

	return b.String()
}

func (m *ConfigSelectModel) renderSelect() string {
	var b strings.Builder

	currentValue := m.Value()
	displayValue := currentValue
	if displayValue == "" {
		displayValue = "(empty)"
	}

	if m.expanded {
		for i, opt := range m.options {
			displayOpt := opt
			if displayOpt == "" {
				displayOpt = "(empty)"
			}

			if i == m.selected {
				cursor := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("▶ ")
				optStyle := lipgloss.NewStyle().
					Background(styles.Nord10).
					Foreground(styles.Nord6).
					Bold(true)
				b.WriteString(cursor + optStyle.Render(displayOpt))
			} else {
				optStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
				b.WriteString("  " + optStyle.Render(displayOpt))
			}
			if i < len(m.options)-1 {
				b.WriteString("\n")
			}
		}
	} else {
		valueStyle := lipgloss.NewStyle().Foreground(styles.TextColor)
		icon := lipgloss.NewStyle().Foreground(styles.MutedColor).Render(" ▼")
		b.WriteString(valueStyle.Render(displayValue) + icon)

		if m.focused {
			hint := lipgloss.NewStyle().Foreground(styles.MutedColor).Italic(true).
				Render("  (Enter to expand, Tab to cycle)")
			b.WriteString(hint)
		}
	}

	return b.String()
}
