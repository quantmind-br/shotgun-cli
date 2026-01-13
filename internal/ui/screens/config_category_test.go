package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		category      config.ConfigCategory
		expectedCount int
	}{
		{"Scanner category", config.CategoryScanner, 9},
		{"Context category", config.CategoryContext, 3},
		{"Template category", config.CategoryTemplate, 1},
		{"Output category", config.CategoryOutput, 2},
		{"LLM category", config.CategoryLLM, 5},
		{"Gemini category", config.CategoryGemini, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := NewConfigCategory(tt.category)

			require.NotNil(t, model)
			assert.Equal(t, tt.category, model.Category())
			assert.Equal(t, tt.expectedCount, model.ItemCount())
			assert.Equal(t, 0, model.Cursor())
			assert.False(t, model.IsEditing())
		})
	}
}

func TestConfigCategory_SetSize(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 24)

	view := model.View()
	assert.NotEmpty(t, view)
}

func TestConfigCategory_Init(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	cmd := model.Init()
	assert.Nil(t, cmd)
}

func TestConfigCategory_Navigation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		key            string
		startCursor    int
		expectedCursor int
	}{
		{"down moves cursor", "down", 0, 1},
		{"j moves cursor down", "j", 0, 1},
		{"up moves cursor", "up", 1, 0},
		{"k moves cursor up", "k", 1, 0},
		{"down at end stays", "down", 8, 8},
		{"up at start stays", "up", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := NewConfigCategory(config.CategoryScanner)
			model.SetSize(80, 100)
			model.cursor = tt.startCursor

			model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})

			assert.Equal(t, tt.expectedCursor, model.Cursor())
		})
	}
}

func TestConfigCategory_HomeEnd(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)
	model.cursor = 4

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	assert.Equal(t, 0, model.Cursor())

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.Equal(t, 8, model.Cursor())

	model.cursor = 4
	model.Update(tea.KeyMsg{Type: tea.KeyHome})
	assert.Equal(t, 0, model.Cursor())

	model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	assert.Equal(t, 8, model.Cursor())
}

func TestConfigCategory_EnterEditMode(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	assert.False(t, model.IsEditing())

	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, model.IsEditing())
}

func TestConfigCategory_ExitEditMode(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, model.IsEditing())

	model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, model.IsEditing())
}

func TestConfigCategory_SpaceToggle(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	for i, item := range model.items {
		if item.fieldToggle != nil {
			model.cursor = i
			break
		}
	}

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
}

func TestConfigCategory_ResetToDefault(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
}

func TestConfigCategory_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		category config.ConfigCategory
	}{
		{"Scanner view", config.CategoryScanner},
		{"Context view", config.CategoryContext},
		{"Output view", config.CategoryOutput},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := NewConfigCategory(tt.category)
			model.SetSize(80, 50)

			view := model.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, string(tt.category))
			assert.Contains(t, view, "Settings")
		})
	}
}

func TestConfigCategory_ViewWithScroll(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 20)

	view := model.View()
	assert.NotEmpty(t, view)

	for i := 0; i < 8; i++ {
		model.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	view = model.View()
	assert.NotEmpty(t, view)
}

func TestConfigCategory_HasUnsavedChanges(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	assert.False(t, model.HasUnsavedChanges())
}

func TestConfigCategory_GetChanges_Empty(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	changes := model.GetChanges()
	assert.Empty(t, changes)
}

func TestConfigCategory_EmptyCategory(t *testing.T) {
	t.Parallel()

	model := &ConfigCategoryModel{
		category: "Empty",
		items:    []configItem{},
	}

	view := model.View()
	assert.Contains(t, view, "No configuration items")
}

func TestConfigCategory_UpdateNonKeyMsg(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	cmd := model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	assert.Nil(t, cmd)
}

func TestConfigCategory_EditModeUpdate(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 100)

	model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, model.IsEditing())

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
}

func TestConfigCategory_EnsureVisible(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 30)

	for i := 0; i < 8; i++ {
		model.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	assert.True(t, model.scrollY >= 0)
}

func TestConfigCategory_EnsureVisible_SmallHeight(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.SetSize(80, 10)

	model.ensureVisible()

	model.cursor = 5
	model.ensureVisible()
}

func TestConfigCategory_EnsureVisible_ZeroHeight(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.height = 0
	model.itemHeight = 0

	model.ensureVisible()
}

func TestConfigCategory_CurrentItem_OutOfBounds(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.cursor = -1
	assert.Nil(t, model.currentItem())

	model.cursor = 100
	assert.Nil(t, model.currentItem())
}

func TestConfigCategory_ResetCurrentToDefault_NilItem(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.cursor = -1
	model.resetCurrentToDefault()
}

func TestConfigCategory_ExitEditMode_NilItem(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.editMode = true
	model.cursor = -1
	model.exitEditMode()
	assert.False(t, model.editMode)
}

func TestConfigCategory_EnterEditMode_NilItem(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.cursor = -1
	cmd := model.enterEditMode()
	assert.Nil(t, cmd)
}

func TestConfigCategory_HandleEditMode_NilItem(t *testing.T) {
	t.Parallel()

	model := NewConfigCategory(config.CategoryScanner)
	model.editMode = true
	model.cursor = -1
	cmd := model.handleEditMode(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	assert.Nil(t, cmd)
}

func TestConfigCategory_HomeEnd_EmptyItems(t *testing.T) {
	t.Parallel()

	model := &ConfigCategoryModel{
		category: "Empty",
		items:    []configItem{},
	}
	model.SetSize(80, 100)

	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	assert.Equal(t, 0, model.Cursor())
}

func TestCreateFieldForItem_AllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		meta      config.ConfigMetadata
		hasField  bool
		hasToggle bool
		hasSelect bool
	}{
		{
			name: "bool type creates toggle",
			meta: config.ConfigMetadata{
				Key:  config.KeyScannerSkipBinary,
				Type: config.TypeBool,
			},
			hasToggle: true,
		},
		{
			name: "enum type creates select",
			meta: config.ConfigMetadata{
				Key:         config.KeyOutputFormat,
				Type:        config.TypeEnum,
				EnumOptions: []string{"markdown", "text"},
			},
			hasSelect: true,
		},
		{
			name: "int type creates field",
			meta: config.ConfigMetadata{
				Key:  config.KeyScannerMaxFiles,
				Type: config.TypeInt,
			},
			hasField: true,
		},
		{
			name: "string type creates field",
			meta: config.ConfigMetadata{
				Key:  config.KeyLLMAPIKey,
				Type: config.TypeString,
			},
			hasField: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			item := configItem{metadata: tt.meta}
			createFieldForItem(&item, tt.meta)

			if tt.hasField {
				assert.NotNil(t, item.fieldField)
			}
			if tt.hasToggle {
				assert.NotNil(t, item.fieldToggle)
			}
			if tt.hasSelect {
				assert.NotNil(t, item.fieldSelect)
			}
		})
	}
}
