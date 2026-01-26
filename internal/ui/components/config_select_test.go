package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigSelect(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		Description: "Output format",
		EnumOptions: []string{"markdown", "text"},
	}

	sel := NewConfigSelect(metadata, "markdown")

	assert.Equal(t, "markdown", sel.Value())
	assert.Equal(t, metadata.Key, sel.Metadata().Key)
	assert.False(t, sel.IsFocused())
	assert.False(t, sel.IsModified())
}

func TestNewConfigSelect_WithEmptyOptions(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:         config.KeyLLMModel,
		Type:        config.TypeEnum,
		Description: "Model name",
		EnumOptions: nil,
	}

	sel := NewConfigSelect(metadata, "gpt-4")

	assert.Equal(t, "gpt-4", sel.Value())
}

func TestNewConfigSelect_ValueNotInOptions(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		Description: "Output format",
		EnumOptions: []string{"markdown", "text"},
	}

	sel := NewConfigSelect(metadata, "unknown")

	assert.Equal(t, "markdown", sel.Value())
}

func TestConfigSelect_FocusBlur(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	assert.False(t, sel.IsFocused())

	sel.Focus()
	assert.True(t, sel.IsFocused())

	sel.Blur()
	assert.False(t, sel.IsFocused())
}

func TestConfigSelect_SetValue(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	sel.SetValue("text")
	assert.Equal(t, "text", sel.Value())
	assert.False(t, sel.IsModified())

	sel.SetValue("unknown")
	assert.Equal(t, "markdown", sel.Value())
}

func TestConfigSelect_Update_ExpandCollapse(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	sel.Focus()

	sel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, sel.expanded)

	sel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, sel.expanded)

	sel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	assert.True(t, sel.expanded)

	sel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, sel.expanded)
}

func TestConfigSelect_Update_Navigation(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text", "json"},
	}, "markdown")

	sel.Focus()
	sel.Update(tea.KeyMsg{Type: tea.KeyEnter})

	sel.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, "text", sel.Value())

	sel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	assert.Equal(t, "json", sel.Value())

	sel.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, "markdown", sel.Value())

	sel.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, "json", sel.Value())

	sel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	assert.Equal(t, "text", sel.Value())
}

func TestConfigSelect_Update_TabCycle(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	sel.Focus()

	sel.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, "text", sel.Value())

	sel.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, "markdown", sel.Value())
}

func TestConfigSelect_Update_TabClosesExpanded(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	sel.Focus()
	sel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, sel.expanded)

	sel.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.False(t, sel.expanded)
}

func TestConfigSelect_Update_WhenNotFocused(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	updated, cmd := sel.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.Equal(t, "markdown", updated.Value())
	assert.False(t, updated.IsModified())
	assert.Nil(t, cmd)
}

func TestConfigSelect_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		focused  bool
		expanded bool
	}{
		{"collapsed unfocused", "markdown", false, false},
		{"collapsed focused", "markdown", true, false},
		{"expanded focused", "text", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sel := NewConfigSelect(config.ConfigMetadata{
				Key:         config.KeyOutputFormat,
				Type:        config.TypeEnum,
				Description: "Test description",
				EnumOptions: []string{"markdown", "text"},
			}, tt.value)

			sel.SetWidth(60)
			if tt.focused {
				sel.Focus()
			}
			if tt.expanded {
				sel.Update(tea.KeyMsg{Type: tea.KeyEnter})
			}

			view := sel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, config.KeyOutputFormat)
			assert.Contains(t, view, "Test description")
		})
	}
}

func TestConfigSelect_View_WithEmptyOption(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         "test_empty_option",
		Type:        config.TypeEnum,
		Description: "Test Empty Option",
		EnumOptions: []string{"", "option1", "option2"},
	}, "")

	sel.SetWidth(60)
	view := sel.View()
	assert.Contains(t, view, "(empty)")
}

func TestConfigSelect_Init(t *testing.T) {
	t.Parallel()

	sel := NewConfigSelect(config.ConfigMetadata{
		Key:         config.KeyOutputFormat,
		Type:        config.TypeEnum,
		EnumOptions: []string{"markdown", "text"},
	}, "markdown")

	cmd := sel.Init()
	assert.Nil(t, cmd)
}

func TestConfigSelect_Metadata(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:          config.KeyOutputFormat,
		Type:         config.TypeEnum,
		Description:  "Output format",
		EnumOptions:  []string{"markdown", "text"},
		DefaultValue: "markdown",
	}

	sel := NewConfigSelect(metadata, "text")
	require.Equal(t, metadata.Key, sel.Metadata().Key)
	require.Equal(t, metadata.Type, sel.Metadata().Type)
}

func TestConfigSelect_Value_EmptyOptions(t *testing.T) {
	t.Parallel()

	sel := &ConfigSelectModel{
		options:  []string{},
		selected: 0,
	}

	assert.Equal(t, "", sel.Value())
}

func TestConfigSelect_Value_OutOfBounds(t *testing.T) {
	t.Parallel()

	sel := &ConfigSelectModel{
		options:  []string{"a", "b"},
		selected: 5,
	}

	assert.Equal(t, "", sel.Value())
}
