package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"

	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	configItemHeight      = 5
	configHeaderHeight    = 4
	configScrollIndicator = 1
)

type ConfigFieldInterface interface {
	Focus() tea.Cmd
	Blur()
	IsFocused() bool
	Update(tea.Msg) (*ConfigFieldInterface, tea.Cmd)
	View() string
	IsModified() bool
	Metadata() config.ConfigMetadata
	SetWidth(int)
}

type configItem struct {
	metadata    config.ConfigMetadata
	fieldField  *components.ConfigFieldModel
	fieldToggle *components.ConfigToggleModel
	fieldSelect *components.ConfigSelectModel
}

type ConfigCategoryModel struct {
	category   config.ConfigCategory
	items      []configItem
	cursor     int
	width      int
	height     int
	scrollY    int
	editMode   bool
	itemHeight int
}

func NewConfigCategory(category config.ConfigCategory) *ConfigCategoryModel {
	metas := config.GetByCategory(category)
	items := make([]configItem, len(metas))

	for i, meta := range metas {
		items[i] = configItem{
			metadata: meta,
		}
		createFieldForItem(&items[i], meta)
	}

	return &ConfigCategoryModel{
		category:   category,
		items:      items,
		itemHeight: configItemHeight,
	}
}

func createFieldForItem(item *configItem, meta config.ConfigMetadata) {
	currentVal := viper.Get(meta.Key)

	switch meta.Type {
	case config.TypeBool:
		boolVal, _ := currentVal.(bool)
		item.fieldToggle = components.NewConfigToggle(meta, boolVal)

	case config.TypeEnum:
		strVal := fmt.Sprintf("%v", currentVal)
		if currentVal == nil {
			strVal = ""
		}
		item.fieldSelect = components.NewConfigSelect(meta, strVal)

	default:
		strVal := fmt.Sprintf("%v", currentVal)
		if currentVal == nil {
			strVal = ""
		}
		item.fieldField = components.NewConfigField(meta, strVal)
	}
}

func (m *ConfigCategoryModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	fieldWidth := width - 8
	if fieldWidth < 30 {
		fieldWidth = 30
	}

	for i := range m.items {
		if m.items[i].fieldField != nil {
			m.items[i].fieldField.SetWidth(fieldWidth)
		}
		if m.items[i].fieldToggle != nil {
			m.items[i].fieldToggle.SetWidth(fieldWidth)
		}
		if m.items[i].fieldSelect != nil {
			m.items[i].fieldSelect.SetWidth(fieldWidth)
		}
	}
}

func (m *ConfigCategoryModel) Init() tea.Cmd {
	return nil
}

func (m *ConfigCategoryModel) Update(msg tea.Msg) tea.Cmd {
	if m.editMode {
		return m.handleEditMode(msg)
	}
	return m.handleNavigationMode(msg)
}

func (m *ConfigCategoryModel) handleNavigationMode(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch keyMsg.String() {
	case "up", "k":
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case "enter":
		return m.enterEditMode()
	case " ":
		item := m.currentItem()
		if item != nil && item.fieldToggle != nil {
			item.fieldToggle.Focus()
			item.fieldToggle.Update(msg)
			item.fieldToggle.Blur()
		}
	case "r":
		m.resetCurrentToDefault()
	case "home", "g":
		m.cursor = 0
		m.ensureVisible()
	case "end", "G":
		if len(m.items) > 0 {
			m.cursor = len(m.items) - 1
			m.ensureVisible()
		}
	}

	return nil
}

func (m *ConfigCategoryModel) handleEditMode(msg tea.Msg) tea.Cmd {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok && keyMsg.String() == "esc" {
		m.exitEditMode()
		return nil
	}

	item := m.currentItem()
	if item == nil {
		return nil
	}

	if item.fieldField != nil {
		item.fieldField.Update(msg)
	} else if item.fieldToggle != nil {
		item.fieldToggle.Update(msg)
	} else if item.fieldSelect != nil {
		item.fieldSelect.Update(msg)
	}

	return nil
}

func (m *ConfigCategoryModel) moveCursor(delta int) {
	newPos := m.cursor + delta
	if newPos >= 0 && newPos < len(m.items) {
		m.cursor = newPos
		m.ensureVisible()
	}
}

func (m *ConfigCategoryModel) ensureVisible() {
	if m.height <= 0 || m.itemHeight <= 0 {
		return
	}

	visibleItems := (m.height - configHeaderHeight - configScrollIndicator*2) / m.itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	if m.cursor < m.scrollY {
		m.scrollY = m.cursor
	} else if m.cursor >= m.scrollY+visibleItems {
		m.scrollY = m.cursor - visibleItems + 1
	}
}

func (m *ConfigCategoryModel) enterEditMode() tea.Cmd {
	m.editMode = true
	item := m.currentItem()
	if item == nil {
		return nil
	}

	if item.fieldField != nil {
		return item.fieldField.Focus()
	} else if item.fieldToggle != nil {
		return item.fieldToggle.Focus()
	} else if item.fieldSelect != nil {
		return item.fieldSelect.Focus()
	}

	return nil
}

func (m *ConfigCategoryModel) exitEditMode() {
	m.editMode = false
	item := m.currentItem()
	if item == nil {
		return
	}

	if item.fieldField != nil {
		item.fieldField.Blur()
	} else if item.fieldToggle != nil {
		item.fieldToggle.Blur()
	} else if item.fieldSelect != nil {
		item.fieldSelect.Blur()
	}
}

func (m *ConfigCategoryModel) currentItem() *configItem {
	if m.cursor >= 0 && m.cursor < len(m.items) {
		return &m.items[m.cursor]
	}
	return nil
}

func (m *ConfigCategoryModel) resetCurrentToDefault() {
	item := m.currentItem()
	if item == nil {
		return
	}

	defaultVal := item.metadata.DefaultValue

	if item.fieldField != nil {
		item.fieldField.SetValue(fmt.Sprintf("%v", defaultVal))
	} else if item.fieldToggle != nil {
		if boolVal, ok := defaultVal.(bool); ok {
			item.fieldToggle.SetValue(boolVal)
		}
	} else if item.fieldSelect != nil {
		item.fieldSelect.SetValue(fmt.Sprintf("%v", defaultVal))
	}
}

func (m *ConfigCategoryModel) View() string {
	if len(m.items) == 0 {
		return styles.HelpStyle.Render("No configuration items in this category")
	}

	var content strings.Builder

	header := styles.SubtitleStyle.Render(string(m.category) + " Settings")
	content.WriteString(header)
	content.WriteString("\n")
	content.WriteString(styles.RenderSeparator(m.width - 4))
	content.WriteString("\n\n")

	visibleItems := (m.height - configHeaderHeight - configScrollIndicator*2) / m.itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	endIdx := m.scrollY + visibleItems
	if endIdx > len(m.items) {
		endIdx = len(m.items)
	}

	if m.scrollY > 0 {
		content.WriteString(styles.HelpStyle.Render("  ↑ more above"))
		content.WriteString("\n")
	}

	for i := m.scrollY; i < endIdx; i++ {
		item := m.items[i]

		cursorStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		if i == m.cursor {
			content.WriteString(cursorStyle.Render("▶ "))
		} else {
			content.WriteString("  ")
		}

		content.WriteString(m.renderItem(&item))
		content.WriteString("\n")
	}

	if endIdx < len(m.items) {
		content.WriteString(styles.HelpStyle.Render("  ↓ more below"))
	}

	return content.String()
}

func (m *ConfigCategoryModel) renderItem(item *configItem) string {
	if item.fieldField != nil {
		return item.fieldField.View()
	} else if item.fieldToggle != nil {
		return item.fieldToggle.View()
	} else if item.fieldSelect != nil {
		return item.fieldSelect.View()
	}
	return ""
}

func (m *ConfigCategoryModel) HasUnsavedChanges() bool {
	for _, item := range m.items {
		if m.isItemModified(&item) {
			return true
		}
	}
	return false
}

func (m *ConfigCategoryModel) isItemModified(item *configItem) bool {
	if item.fieldField != nil {
		return item.fieldField.IsModified()
	} else if item.fieldToggle != nil {
		return item.fieldToggle.IsModified()
	} else if item.fieldSelect != nil {
		return item.fieldSelect.IsModified()
	}
	return false
}

func (m *ConfigCategoryModel) GetChanges() map[string]interface{} {
	changes := make(map[string]interface{})

	for _, item := range m.items {
		if !m.isItemModified(&item) {
			continue
		}

		key := item.metadata.Key

		if item.fieldField != nil {
			changes[key] = item.fieldField.Value()
		} else if item.fieldToggle != nil {
			changes[key] = item.fieldToggle.Value()
		} else if item.fieldSelect != nil {
			changes[key] = item.fieldSelect.Value()
		}
	}

	return changes
}

func (m *ConfigCategoryModel) Category() config.ConfigCategory {
	return m.category
}

func (m *ConfigCategoryModel) IsEditing() bool {
	return m.editMode
}

func (m *ConfigCategoryModel) ItemCount() int {
	return len(m.items)
}

func (m *ConfigCategoryModel) Cursor() int {
	return m.cursor
}
