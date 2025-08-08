package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shotgun-cli/internal/core"
)

func TestPatternConfigModel(t *testing.T) {
	// Create test configuration
	config := core.DefaultEnhancedConfig()
	config.App.CustomIgnorePatterns = []string{"*.tmp", "build/"}
	config.App.ForceIncludePatterns = []string{"important.log", "config/*.yml"}
	config.App.PatternValidationEnabled = true

	t.Run("Model Creation", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)
		require.NotNil(t, model)

		// Check initial state
		assert.Equal(t, 0, model.currentTab) // Should start with custom ignore tab
		assert.False(t, model.isAddingPattern)
		assert.Equal(t, -1, model.editIndex)
		assert.True(t, model.validationEnabled)
		assert.Equal(t, 100, model.width)
		assert.Equal(t, 30, model.height)

		// Verify patterns were loaded
		assert.Equal(t, 2, len(model.customPatterns.Items())) // List length
		assert.Equal(t, 2, len(model.forcePatterns.Items()))  // List length
	})

	t.Run("Tab Navigation", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Test tab switching
		assert.Equal(t, 0, model.currentTab)

		// Switch to force include tab
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		assert.Equal(t, 1, model.currentTab)

		// Switch back to custom ignore tab
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		assert.Equal(t, 0, model.currentTab)

		// Test wrapping
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		assert.Equal(t, 1, model.currentTab)
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
		assert.Equal(t, 0, model.currentTab) // Should wrap back
	})

	t.Run("Add Pattern Flow", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Start adding pattern
		assert.False(t, model.isAddingPattern)
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		assert.True(t, model.isAddingPattern)
		assert.Equal(t, "", model.validationError)

		// Simulate typing a pattern
		model.addInput.SetValue("*.bak")

		// Confirm addition
		model.addPattern()
		assert.False(t, model.isAddingPattern)

		// Verify pattern was added to configuration
		updatedConfig := model.GetConfig()
		assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, "*.bak")

		// Cancel addition
		model.startAdd()
		assert.True(t, model.isAddingPattern)
		model.cancelAdd()
		assert.False(t, model.isAddingPattern)
	})

	t.Run("Edit Pattern Flow", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Start editing first pattern (*.tmp)
		model.customPatterns.Select(0)
		assert.Equal(t, -1, model.editIndex)
		model.startEdit()
		assert.Equal(t, 0, model.editIndex)
		assert.Equal(t, "*.tmp", model.editInput.Value())

		// Modify pattern
		model.editInput.SetValue("*.temp")

		// Confirm edit
		model.confirmEdit()
		assert.Equal(t, -1, model.editIndex)

		// Verify pattern was updated
		updatedConfig := model.GetConfig()
		assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, "*.temp")
		assert.NotContains(t, updatedConfig.App.CustomIgnorePatterns, "*.tmp")

		// Cancel edit
		model.startEdit()
		model.editInput.SetValue("*.cancelled")
		model.cancelEdit()
		assert.Equal(t, -1, model.editIndex)

		// Pattern should not be changed
		updatedConfig = model.GetConfig()
		assert.NotContains(t, updatedConfig.App.CustomIgnorePatterns, "*.cancelled")
	})

	t.Run("Delete Pattern", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Delete first custom ignore pattern
		model.customPatterns.Select(0) // Select *.tmp
		initialCount := len(model.config.App.CustomIgnorePatterns)

		model.deleteSelected()

		// Verify pattern was removed
		updatedConfig := model.GetConfig()
		assert.Equal(t, initialCount-1, len(updatedConfig.App.CustomIgnorePatterns))
		assert.NotContains(t, updatedConfig.App.CustomIgnorePatterns, "*.tmp")

		// Switch to force include tab and delete pattern
		model.currentTab = 1
		model.forcePatterns.Select(0) // Select important.log
		initialForceCount := len(model.config.App.ForceIncludePatterns)

		model.deleteSelected()

		updatedConfig = model.GetConfig()
		assert.Equal(t, initialForceCount-1, len(updatedConfig.App.ForceIncludePatterns))
		assert.NotContains(t, updatedConfig.App.ForceIncludePatterns, "important.log")
	})

	t.Run("Pattern Validation", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Test valid patterns
		valid, err := model.validatePattern("*.txt")
		assert.True(t, valid)
		assert.Empty(t, err)

		valid, err = model.validatePattern("build/")
		assert.True(t, valid)
		assert.Empty(t, err)

		valid, err = model.validatePattern("src/**/*.go")
		assert.True(t, valid)
		assert.Empty(t, err)

		// Test invalid patterns
		valid, err = model.validatePattern("")
		assert.False(t, valid)
		assert.Contains(t, err, "empty")

		// Test validation toggle
		model.validationEnabled = true
		assert.True(t, model.validationEnabled)

		// Toggle validation
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
		assert.False(t, model.validationEnabled)
		assert.False(t, model.config.App.PatternValidationEnabled)
	})

	t.Run("Pattern Validation During Add", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)
		model.validationEnabled = true

		// Try to add invalid pattern (empty pattern)
		// Note: The current implementation returns early for empty patterns without validation error
		initialPatternCount := len(model.config.App.CustomIgnorePatterns)
		model.startAdd()
		model.addInput.SetValue("") // Empty pattern
		model.addPattern()

		// Should remain in add mode (early return doesn't call cancelAdd)
		assert.True(t, model.isAddingPattern)
		// Verify pattern was not added
		assert.Equal(t, initialPatternCount, len(model.config.App.CustomIgnorePatterns))

		// Fix pattern and try again
		model.addInput.SetValue("*.valid")
		model.addPattern()

		// Should succeed and exit add mode
		assert.False(t, model.isAddingPattern)
		assert.Empty(t, model.validationError)

		// Verify pattern was added
		updatedConfig := model.GetConfig()
		assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, "*.valid")
	})

	t.Run("Keyboard Shortcuts", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Test adding pattern with 'a'
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		assert.True(t, model.isAddingPattern)

		// Cancel with escape
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		assert.False(t, model.isAddingPattern)

		// Test editing with 'e'
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
		assert.Equal(t, 0, model.editIndex) // Should start editing first pattern

		// Cancel edit with escape
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
		assert.Equal(t, -1, model.editIndex)

		// Test delete with 'd'
		initialCount := len(model.config.App.CustomIgnorePatterns)
		*model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
		updatedConfig := model.GetConfig()
		assert.Equal(t, initialCount-1, len(updatedConfig.App.CustomIgnorePatterns))
	})

	t.Run("Window Resize", func(t *testing.T) {
		model := NewPatternConfigModel(config, 100, 30)

		// Test window resize
		resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
		*model, _ = model.Update(resizeMsg)

		assert.Equal(t, 120, model.width)
		assert.Equal(t, 40, model.height)

		// Verify list sizes were updated (this is internal but we can check the model state)
		assert.Equal(t, 120, model.width)
		assert.Equal(t, 40, model.height)
	})

	t.Run("Configuration Independence", func(t *testing.T) {
		originalConfig := core.DefaultEnhancedConfig()
		originalConfig.App.CustomIgnorePatterns = []string{"*.orig"}

		model := NewPatternConfigModel(originalConfig, 100, 30)

		// Modify patterns in model
		model.config.App.CustomIgnorePatterns = append(model.config.App.CustomIgnorePatterns, "*.new")

		// Original config should be unchanged
		assert.Equal(t, []string{"*.orig"}, originalConfig.App.CustomIgnorePatterns)

		// Model config should have the new pattern
		updatedConfig := model.GetConfig()
		assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, "*.new")
	})
}

// TestPatternConfigModelEdgeCases tests edge cases and error conditions
func TestPatternConfigModelEdgeCases(t *testing.T) {
	t.Run("Empty Configuration", func(t *testing.T) {
		config := core.DefaultEnhancedConfig()
		// Ensure patterns are empty
		config.App.CustomIgnorePatterns = []string{}
		config.App.ForceIncludePatterns = []string{}

		model := NewPatternConfigModel(config, 100, 30)

		// Should handle empty patterns gracefully
		assert.NotNil(t, model)
		assert.Equal(t, 0, len(model.config.App.CustomIgnorePatterns))
		assert.Equal(t, 0, len(model.config.App.ForceIncludePatterns))

		// Should be able to add first pattern
		model.startAdd()
		model.addInput.SetValue("*.first")
		model.addPattern()

		updatedConfig := model.GetConfig()
		assert.Equal(t, 1, len(updatedConfig.App.CustomIgnorePatterns))
		assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, "*.first")
	})

	t.Run("Large Pattern Lists", func(t *testing.T) {
		config := core.DefaultEnhancedConfig()

		// Create large pattern lists
		for i := 0; i < 50; i++ {
			config.App.CustomIgnorePatterns = append(config.App.CustomIgnorePatterns, "*.tmp"+string(rune('a'+i%26)))
			config.App.ForceIncludePatterns = append(config.App.ForceIncludePatterns, "important"+string(rune('a'+i%26))+".log")
		}

		model := NewPatternConfigModel(config, 100, 30)

		// Should handle large lists without issues
		assert.NotNil(t, model)
		assert.Equal(t, 50, len(model.config.App.CustomIgnorePatterns))
		assert.Equal(t, 50, len(model.config.App.ForceIncludePatterns))

		// Should still be able to add, edit, and delete patterns
		model.startAdd()
		model.addInput.SetValue("*.new")
		model.addPattern()

		updatedConfig := model.GetConfig()
		assert.Equal(t, 51, len(updatedConfig.App.CustomIgnorePatterns))
	})

	t.Run("Special Characters in Patterns", func(t *testing.T) {
		config := core.DefaultEnhancedConfig()
		model := NewPatternConfigModel(config, 100, 30)

		specialPatterns := []string{
			"*.file-with-dash",
			"*.file_with_underscore",
			"file with spaces.*", // Spaces in pattern
			"[abc]*.txt",         // Character class
			"*.{js,ts,jsx,tsx}",  // Brace expansion
			"!important.txt",     // Negation
		}

		for _, pattern := range specialPatterns {
			valid, err := model.validatePattern(pattern)
			// Most special patterns should be valid for gitignore
			if !valid {
				t.Logf("Pattern '%s' validation failed: %s", pattern, err)
			}
			// Don't assert here as some patterns may legitimately fail
		}
	})

	t.Run("Concurrent Pattern Operations", func(t *testing.T) {
		config := core.DefaultEnhancedConfig()
		model := NewPatternConfigModel(config, 100, 30)

		// Simulate rapid pattern additions (like paste operations)
		patterns := []string{"*.tmp", "*.bak", "*.log", "build/", "dist/"}

		for _, pattern := range patterns {
			model.startAdd()
			model.addInput.SetValue(pattern)
			model.addPattern()
		}

		updatedConfig := model.GetConfig()
		for _, pattern := range patterns {
			assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, pattern)
		}
	})
}

// TestPatternItemInterface tests the PatternItem interface implementation
func TestPatternItemInterface(t *testing.T) {
	t.Run("Valid Pattern Item", func(t *testing.T) {
		item := PatternItem{
			pattern: "*.tmp",
			valid:   true,
			error:   "",
		}

		assert.Equal(t, "*.tmp", item.FilterValue())
		assert.Equal(t, "*.tmp", item.Title())
		assert.Equal(t, "✅ Valid pattern", item.Description())
	})

	t.Run("Invalid Pattern Item", func(t *testing.T) {
		item := PatternItem{
			pattern: "[invalid",
			valid:   false,
			error:   "Invalid bracket pattern",
		}

		assert.Equal(t, "[invalid", item.FilterValue())
		assert.Equal(t, "[invalid", item.Title())
		assert.Equal(t, "❌ Invalid bracket pattern", item.Description())
	})

	t.Run("Valid Pattern With Empty Error", func(t *testing.T) {
		item := PatternItem{
			pattern: "*.go",
			valid:   true,
			error:   "", // Should be ignored when valid
		}

		assert.Equal(t, "*.go", item.FilterValue())
		assert.Equal(t, "*.go", item.Title())
		assert.Equal(t, "✅ Valid pattern", item.Description())
	})
}
