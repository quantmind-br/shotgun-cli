package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigField(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:         config.KeyScannerMaxFiles,
		Type:        config.TypeInt,
		Description: "Maximum files to scan",
	}

	field := NewConfigField(metadata, "1000")

	assert.Equal(t, "1000", field.Value())
	assert.Equal(t, metadata.Key, field.Metadata().Key)
	assert.False(t, field.IsFocused())
	assert.False(t, field.IsModified())
	assert.True(t, field.IsValid())
}

func TestConfigField_FocusBlur(t *testing.T) {
	t.Parallel()

	field := NewConfigField(config.ConfigMetadata{
		Key:  config.KeyScannerMaxFiles,
		Type: config.TypeInt,
	}, "100")

	assert.False(t, field.IsFocused())

	field.Focus()
	assert.True(t, field.IsFocused())

	field.Blur()
	assert.False(t, field.IsFocused())
}

func TestConfigField_SetValue(t *testing.T) {
	t.Parallel()

	field := NewConfigField(config.ConfigMetadata{
		Key:  config.KeyScannerMaxFiles,
		Type: config.TypeInt,
	}, "100")

	field.SetValue("500")
	assert.Equal(t, "500", field.Value())
	assert.False(t, field.IsModified())
}

func TestConfigField_SetWidth(t *testing.T) {
	t.Parallel()

	field := NewConfigField(config.ConfigMetadata{
		Key:  config.KeyScannerMaxFiles,
		Type: config.TypeInt,
	}, "100")

	field.SetWidth(80)
	assert.NotEmpty(t, field.View())
}

func TestConfigField_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata config.ConfigMetadata
		value    string
		focused  bool
	}{
		{
			name: "int field unfocused",
			metadata: config.ConfigMetadata{
				Key:         config.KeyScannerMaxFiles,
				Type:        config.TypeInt,
				Description: "Max files",
			},
			value:   "1000",
			focused: false,
		},
		{
			name: "size field focused",
			metadata: config.ConfigMetadata{
				Key:         config.KeyScannerMaxFileSize,
				Type:        config.TypeSize,
				Description: "Max file size",
			},
			value:   "10MB",
			focused: true,
		},
		{
			name: "path field",
			metadata: config.ConfigMetadata{
				Key:         config.KeyTemplateCustomPath,
				Type:        config.TypePath,
				Description: "Custom path",
			},
			value:   "/tmp/test",
			focused: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			field := NewConfigField(tt.metadata, tt.value)
			field.SetWidth(60)
			if tt.focused {
				field.Focus()
			}
			view := field.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, tt.metadata.Key)
			assert.Contains(t, view, tt.metadata.Description)
		})
	}
}

func TestConfigField_Update_WhenNotFocused(t *testing.T) {
	t.Parallel()

	field := NewConfigField(config.ConfigMetadata{
		Key:  config.KeyScannerMaxFiles,
		Type: config.TypeInt,
	}, "100")

	updated, cmd := field.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

	assert.Equal(t, "100", updated.Value())
	assert.Nil(t, cmd)
}

func TestConfigField_Init(t *testing.T) {
	t.Parallel()

	field := NewConfigField(config.ConfigMetadata{
		Key:  config.KeyScannerMaxFiles,
		Type: config.TypeInt,
	}, "100")

	cmd := field.Init()
	assert.NotNil(t, cmd)
}

func TestConfigField_Error(t *testing.T) {
	t.Parallel()

	field := NewConfigField(config.ConfigMetadata{
		Key:  config.KeyScannerMaxFiles,
		Type: config.TypeInt,
	}, "100")

	assert.Nil(t, field.Error())
}

func TestGetPlaceholder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cfgType  config.ConfigType
		expected string
	}{
		{config.TypeSize, "e.g., 10MB, 500KB"},
		{config.TypePath, "/path/to/file"},
		{config.TypeURL, "https://api.example.com"},
		{config.TypeInt, "Enter a number"},
		{config.TypeTimeout, "Timeout in seconds"},
		{config.TypeString, ""},
		{config.TypeBool, ""},
		{config.TypeEnum, ""},
	}

	for _, tt := range tests {
		t.Run(tt.cfgType.String(), func(t *testing.T) {
			t.Parallel()
			m := config.ConfigMetadata{Type: tt.cfgType}
			placeholder := getPlaceholder(m)
			assert.Equal(t, tt.expected, placeholder)
		})
	}
}

func TestConfigField_Metadata(t *testing.T) {
	t.Parallel()

	metadata := config.ConfigMetadata{
		Key:          config.KeyScannerMaxFiles,
		Type:         config.TypeInt,
		Description:  "Test desc",
		DefaultValue: 1000,
	}

	field := NewConfigField(metadata, "500")
	require.Equal(t, metadata.Key, field.Metadata().Key)
	require.Equal(t, metadata.Type, field.Metadata().Type)
}
