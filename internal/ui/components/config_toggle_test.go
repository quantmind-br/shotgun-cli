package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigToggle(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:         config.KeyScannerSkipBinary,
		Type:        config.TypeBool,
		Description: "Skip binary files",
	}

	toggle := NewConfigToggle(metadata, true)

	assert.True(t, toggle.Value())
	assert.Equal(t, metadata.Key, toggle.Metadata().Key)
	assert.False(t, toggle.IsFocused())
	assert.False(t, toggle.IsModified())
}

func TestConfigToggle_FocusBlur(t *testing.T) {
	t.Parallel()

	toggle := NewConfigToggle(config.ConfigMetadata{
		Key:  config.KeyScannerSkipBinary,
		Type: config.TypeBool,
	}, true)

	assert.False(t, toggle.IsFocused())

	toggle.Focus()
	assert.True(t, toggle.IsFocused())

	toggle.Blur()
	assert.False(t, toggle.IsFocused())
}

func TestConfigToggle_SetValue(t *testing.T) {
	t.Parallel()

	toggle := NewConfigToggle(config.ConfigMetadata{
		Key:  config.KeyScannerSkipBinary,
		Type: config.TypeBool,
	}, true)

	toggle.SetValue(false)
	assert.False(t, toggle.Value())
	assert.False(t, toggle.IsModified())
}

func TestConfigToggle_Update_Toggle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		key         string
		startValue  bool
		expectValue bool
	}{
		{"enter toggles on to off", "enter", true, false},
		{"enter toggles off to on", "enter", false, true},
		{"space toggles on to off", " ", true, false},
		{"space toggles off to on", " ", false, true},
		{"tab toggles on to off", "tab", true, false},
		{"tab toggles off to on", "tab", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			toggle := NewConfigToggle(config.ConfigMetadata{
				Key:  config.KeyScannerSkipBinary,
				Type: config.TypeBool,
			}, tt.startValue)

			toggle.Focus()
			updated, _ := toggle.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})

			assert.Equal(t, tt.expectValue, updated.Value())
			assert.True(t, updated.IsModified())
		})
	}
}

func TestConfigToggle_Update_YN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		key         string
		startValue  bool
		expectValue bool
		modified    bool
	}{
		{"y sets true (was false)", "y", false, true, true},
		{"Y sets true (was false)", "Y", false, true, true},
		{"y stays true (was true)", "y", true, true, false},
		{"n sets false (was true)", "n", true, false, true},
		{"N sets false (was true)", "N", true, false, true},
		{"n stays false (was false)", "n", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			toggle := NewConfigToggle(config.ConfigMetadata{
				Key:  config.KeyScannerSkipBinary,
				Type: config.TypeBool,
			}, tt.startValue)

			toggle.Focus()
			updated, _ := toggle.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})

			assert.Equal(t, tt.expectValue, updated.Value())
			assert.Equal(t, tt.modified, updated.IsModified())
		})
	}
}

func TestConfigToggle_Update_WhenNotFocused(t *testing.T) {
	t.Parallel()

	toggle := NewConfigToggle(config.ConfigMetadata{
		Key:  config.KeyScannerSkipBinary,
		Type: config.TypeBool,
	}, true)

	updated, cmd := toggle.Update(tea.KeyMsg{Type: tea.KeyEnter})

	assert.True(t, updated.Value())
	assert.False(t, updated.IsModified())
	assert.Nil(t, cmd)
}

func TestConfigToggle_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   bool
		focused bool
	}{
		{"true unfocused", true, false},
		{"false unfocused", false, false},
		{"true focused", true, true},
		{"false focused", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			toggle := NewConfigToggle(config.ConfigMetadata{
				Key:         config.KeyScannerSkipBinary,
				Type:        config.TypeBool,
				Description: "Test description",
			}, tt.value)

			toggle.SetWidth(60)
			if tt.focused {
				toggle.Focus()
			}

			view := toggle.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, config.KeyScannerSkipBinary)
			assert.Contains(t, view, "Test description")
			assert.Contains(t, view, "ON")
			assert.Contains(t, view, "OFF")
		})
	}
}

func TestConfigToggle_Init(t *testing.T) {
	t.Parallel()

	toggle := NewConfigToggle(config.ConfigMetadata{
		Key:  config.KeyScannerSkipBinary,
		Type: config.TypeBool,
	}, true)

	cmd := toggle.Init()
	assert.Nil(t, cmd)
}

func TestConfigToggle_Metadata(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:          config.KeyScannerSkipBinary,
		Type:         config.TypeBool,
		Description:  "Skip binary",
		DefaultValue: true,
	}

	toggle := NewConfigToggle(metadata, false)
	require.Equal(t, metadata.Key, toggle.Metadata().Key)
	require.Equal(t, metadata.Type, toggle.Metadata().Type)
}
