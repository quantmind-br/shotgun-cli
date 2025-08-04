package ui

import (
	"runtime"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestWindowsCompatibleTextAreaCreation(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")

	assert.Equal(t, "Test placeholder", ta.placeholder)
	assert.False(t, ta.isEditing)
	assert.False(t, ta.focused)
	assert.Equal(t, 80, ta.width)
	assert.Equal(t, 10, ta.height)
	assert.Equal(t, "", ta.Value())
}

func TestWindowsCompatibleTextAreaSetSize(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")

	ta.SetSize(100, 20)

	assert.Equal(t, 100, ta.width)
	assert.Equal(t, 20, ta.height)
}

func TestWindowsCompatibleTextAreaSetValue(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")

	ta.SetValue("Test content")

	assert.Equal(t, "Test content", ta.Value())
	assert.Equal(t, "Test content", ta.manualText)
}

func TestWindowsCompatibleTextAreaFocusAndBlur(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")

	// Initially not focused
	assert.False(t, ta.Focused())

	// Focus
	cmd := ta.Focus()
	assert.NotNil(t, cmd)
	assert.True(t, ta.Focused())

	// Blur
	ta.Blur()
	assert.False(t, ta.Focused())
}

func TestWindowsCompatibleTextAreaTabToggleEditing(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.SetValue("Initial content")

	// Tab to enter editing mode
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedTA, cmd := ta.Update(msg)

	assert.True(t, updatedTA.isEditing)
	assert.Nil(t, cmd)

	// Tab again to exit editing mode
	updatedTA, cmd = updatedTA.Update(msg)

	assert.False(t, updatedTA.isEditing)
	assert.Nil(t, cmd)
}

func TestWindowsCompatibleTextAreaManualInput(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.isEditing = true

	// Test character input
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}}
	updatedTA, cmd := ta.Update(msg)
	assert.Equal(t, "H", updatedTA.manualText)
	assert.Nil(t, cmd)

	// Test more characters
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedTA, cmd = updatedTA.Update(msg)
	assert.Equal(t, "He", updatedTA.manualText)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updatedTA, cmd = updatedTA.Update(msg)
	assert.Equal(t, "Hel", updatedTA.manualText)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	updatedTA, cmd = updatedTA.Update(msg)
	assert.Equal(t, "Hell", updatedTA.manualText)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	updatedTA, cmd = updatedTA.Update(msg)
	assert.Equal(t, "Hello", updatedTA.manualText)
}

func TestWindowsCompatibleTextAreaBackspace(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.isEditing = true
	ta.manualText = "Hello"

	// Test backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedTA, cmd := ta.Update(msg)

	assert.Equal(t, "Hell", updatedTA.manualText)
	assert.Nil(t, cmd)

	// Test multiple backspaces
	updatedTA, _ = updatedTA.Update(msg)
	assert.Equal(t, "Hel", updatedTA.manualText)

	updatedTA, _ = updatedTA.Update(msg)
	assert.Equal(t, "He", updatedTA.manualText)

	updatedTA, _ = updatedTA.Update(msg)
	assert.Equal(t, "H", updatedTA.manualText)

	updatedTA, _ = updatedTA.Update(msg)
	assert.Equal(t, "", updatedTA.manualText)

	// Backspace on empty string should not crash
	updatedTA, _ = updatedTA.Update(msg)
	assert.Equal(t, "", updatedTA.manualText)
}

func TestWindowsCompatibleTextAreaEnterKey(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.isEditing = true
	ta.manualText = "Line 1"

	// Test enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedTA, cmd := ta.Update(msg)

	assert.Equal(t, "Line 1\n", updatedTA.manualText)
	assert.Nil(t, cmd)

	// Add more content after newline
	for _, r := range "Line 2" {
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedTA, cmd = updatedTA.Update(msg)
	}

	assert.Equal(t, "Line 1\nLine 2", updatedTA.manualText)
}

func TestWindowsCompatibleTextAreaEscapeKey(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.SetValue("Original content")
	ta.isEditing = true
	ta.manualText = "Modified content"

	// Test escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedTA, cmd := ta.Update(msg)

	assert.False(t, updatedTA.isEditing)
	assert.Equal(t, "Original content", updatedTA.manualText) // Should revert to textarea value
	assert.Nil(t, cmd)
}

func TestWindowsCompatibleTextAreaViewInEditingMode(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.isEditing = true
	ta.manualText = "Test content"

	view := ta.View()

	// Should show manual text with cursor
	assert.Contains(t, view, "Test content█")

	// On Windows, should show editing mode indicator
	if runtime.GOOS == "windows" {
		assert.Contains(t, view, "[Windows Edit Mode - Tab to exit]")
	}
}

func TestWindowsCompatibleTextAreaViewNormalMode(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.isEditing = false
	ta.SetValue("Test content")

	view := ta.View()

	// Should not contain cursor or editing indicators
	assert.NotContains(t, view, "█")
	assert.NotContains(t, view, "[Windows Edit Mode")
}

func TestWindowsCompatibleTextInputCreation(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")

	assert.Equal(t, "Test placeholder", ti.placeholder)
	assert.False(t, ti.isEditing)
	assert.False(t, ti.focused)
	assert.Equal(t, 0, ti.resetCount)
	assert.Equal(t, "", ti.Value())
}

func TestWindowsCompatibleTextInputSetWidth(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")

	ti.SetWidth(60)

	assert.Equal(t, 60, ti.input.Width)
}

func TestWindowsCompatibleTextInputSetValue(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")

	ti.SetValue("Test content")

	assert.Equal(t, "Test content", ti.Value())
	assert.Equal(t, "Test content", ti.manualText)
}

func TestWindowsCompatibleTextInputSetValueWindowsReset(t *testing.T) {
	// Skip this test on non-Windows systems
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	ti := NewWindowsCompatibleTextInput("Test placeholder")

	// First set value
	ti.SetValue("First value")
	assert.Equal(t, 1, ti.resetCount)
	assert.Equal(t, "First value", ti.lastResetValue)

	// Set same value again (should not increment reset count)
	ti.SetValue("First value")
	assert.Equal(t, 1, ti.resetCount)

	// Set different value (should increment reset count)
	ti.SetValue("Second value")
	assert.Equal(t, 2, ti.resetCount)
	assert.Equal(t, "Second value", ti.lastResetValue)
}

func TestWindowsCompatibleTextInputFocusAndBlur(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")

	// Initially not focused
	assert.False(t, ti.Focused())

	// Focus
	cmd := ti.Focus()
	assert.NotNil(t, cmd)
	assert.True(t, ti.Focused())

	// Blur
	ti.Blur()
	assert.False(t, ti.Focused())
}

func TestWindowsCompatibleTextInputTabToggleEditing(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.SetValue("Initial content")

	// Tab to enter editing mode
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedTI, cmd := ti.Update(msg)

	assert.True(t, updatedTI.isEditing)
	assert.Nil(t, cmd)

	// Tab again to exit editing mode
	updatedTI, cmd = updatedTI.Update(msg)

	assert.False(t, updatedTI.isEditing)
	assert.Nil(t, cmd)
}

func TestWindowsCompatibleTextInputManualInput(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.isEditing = true

	// Test character input
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}}
	updatedTI, cmd := ti.Update(msg)
	assert.Equal(t, "T", updatedTI.manualText)
	assert.Nil(t, cmd)

	// Test more characters
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	updatedTI, cmd = updatedTI.Update(msg)
	assert.Equal(t, "Te", updatedTI.manualText)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updatedTI, cmd = updatedTI.Update(msg)
	assert.Equal(t, "Tes", updatedTI.manualText)

	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}
	updatedTI, cmd = updatedTI.Update(msg)
	assert.Equal(t, "Test", updatedTI.manualText)
}

func TestWindowsCompatibleTextInputBackspace(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.isEditing = true
	ti.manualText = "Test"

	// Test backspace
	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	updatedTI, cmd := ti.Update(msg)

	assert.Equal(t, "Tes", updatedTI.manualText)
	assert.Nil(t, cmd)

	// Test multiple backspaces
	updatedTI, _ = updatedTI.Update(msg)
	assert.Equal(t, "Te", updatedTI.manualText)

	updatedTI, _ = updatedTI.Update(msg)
	assert.Equal(t, "T", updatedTI.manualText)

	updatedTI, _ = updatedTI.Update(msg)
	assert.Equal(t, "", updatedTI.manualText)

	// Backspace on empty string should not crash
	updatedTI, _ = updatedTI.Update(msg)
	assert.Equal(t, "", updatedTI.manualText)
}

func TestWindowsCompatibleTextInputEnterKeyExitsEditing(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.isEditing = true
	ti.manualText = "Test content"

	// Test enter key (should exit editing mode for single-line input)
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	updatedTI, cmd := ti.Update(msg)

	assert.False(t, updatedTI.isEditing)
	assert.Equal(t, "Test content", updatedTI.Value()) // Should be applied to input
	assert.Nil(t, cmd)
}

func TestWindowsCompatibleTextInputEscapeKey(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.SetValue("Original content")
	ti.isEditing = true
	ti.manualText = "Modified content"

	// Test escape key
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	updatedTI, cmd := ti.Update(msg)

	assert.False(t, updatedTI.isEditing)
	assert.Equal(t, "Original content", updatedTI.manualText) // Should revert to input value
	assert.Nil(t, cmd)
}

func TestWindowsCompatibleTextInputViewInEditingMode(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.isEditing = true
	ti.manualText = "Test content"

	view := ti.View()

	// Should show manual text with cursor
	assert.Contains(t, view, "Test content█")

	// On Windows, should show editing mode indicator
	if runtime.GOOS == "windows" {
		assert.Contains(t, view, "[Edit Mode]")
	}
}

func TestWindowsCompatibleTextInputViewNormalMode(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.isEditing = false
	ti.SetValue("Test content")

	view := ti.View()

	// Should not contain cursor or editing indicators
	assert.NotContains(t, view, "█")
	assert.NotContains(t, view, "[Edit Mode")
}

func TestIsWindowsInputWorkaroundNeeded(t *testing.T) {
	result := IsWindowsInputWorkaroundNeeded()

	// Should return true only on Windows
	expected := runtime.GOOS == "windows"
	assert.Equal(t, expected, result)
}

func TestCreateSafeTextInput(t *testing.T) {
	input := CreateSafeTextInput("Test placeholder")

	assert.NotNil(t, input)
	assert.Equal(t, "Test placeholder", input.Placeholder)

	// On Windows, should be blurred initially
	if runtime.GOOS == "windows" {
		assert.False(t, input.Focused())
	}
}

func TestApplyWindowsInputFix(t *testing.T) {
	// Create a regular textinput first
	input := CreateSafeTextInput("Test")
	input.SetValue("Some content")

	// Apply Windows fix
	ApplyWindowsInputFix(&input)

	// On Windows, value should be reset and input should be blurred
	if runtime.GOOS == "windows" {
		assert.Equal(t, "", input.Value())
		assert.False(t, input.Focused())
	}
}

func TestValidateInputCharacter(t *testing.T) {
	tests := []struct {
		char     string
		expected bool
		name     string
	}{
		{"a", true, "lowercase letter"},
		{"A", true, "uppercase letter"},
		{"1", true, "digit"},
		{" ", true, "space"},
		{"!", true, "exclamation"},
		{"@", true, "at symbol"},
		{"#", true, "hash"},
		{"$", true, "dollar"},
		{"%", true, "percent"},
		{"^", true, "caret"},
		{"&", true, "ampersand"},
		{"*", true, "asterisk"},
		{"(", true, "open paren"},
		{")", true, "close paren"},
		{"-", true, "dash"},
		{"_", true, "underscore"},
		{"=", true, "equals"},
		{"+", true, "plus"},
		{"[", true, "open bracket"},
		{"]", true, "close bracket"},
		{"{", true, "open brace"},
		{"}", true, "close brace"},
		{"|", true, "pipe"},
		{"\\", true, "backslash"},
		{":", true, "colon"},
		{";", true, "semicolon"},
		{"\"", true, "double quote"},
		{"'", true, "single quote"},
		{"<", true, "less than"},
		{">", true, "greater than"},
		{",", true, "comma"},
		{".", true, "period"},
		{"?", true, "question mark"},
		{"/", true, "forward slash"},
		{"~", true, "tilde"},
		{"`", true, "backtick"},
		{"\n", false, "newline"},
		{"\t", false, "tab"},
		{"\r", false, "carriage return"},
		{"\x00", false, "null character"},
		{"\x01", false, "control character"},
		{"\x1F", false, "unit separator"},
		{"", false, "empty string"},
		{"ab", false, "multi character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateInputCharacter(tt.char)
			assert.Equal(t, tt.expected, result, "Character: %q", tt.char)
		})
	}
}

func TestSanitizeInputText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		name     string
	}{
		{"Hello World", "Hello World", "normal text"},
		{"Hello\nWorld", "HelloWorld", "text with newline"},
		{"Hello\tWorld", "HelloWorld", "text with tab"},
		{"Hello\x00World", "HelloWorld", "text with null character"},
		{"Test123!@#", "Test123!@#", "text with symbols"},
		{"", "", "empty string"},
		{"\n\t\r", "", "only control characters"},
		{"Good\x01Bad\x02Text", "GoodBadText", "text with control characters"},
		{"Normal text with spaces", "Normal text with spaces", "normal text with spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInputText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWindowsCompatibleTextAreaMultilineHandling(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.isEditing = true

	// Add multiple lines
	ta.manualText = "Line 1\nLine 2\nLine 3"

	view := ta.View()

	// Should handle multiple lines properly in view
	lines := strings.Split(view, "\n")

	// Check that cursor is added to the last line
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Line 3█") {
			found = true
			break
		}
	}
	assert.True(t, found, "Cursor should be on the last line")
}

func TestWindowsCompatibleInputCharacterFiltering(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.isEditing = true

	// Test that control characters are filtered out
	invalidChars := []rune{'\x00', '\x01', '\x02', '\x1F', '\n', '\t', '\r'}

	for _, char := range invalidChars {
		initialText := ti.manualText
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		updatedTI, _ := ti.Update(msg)

		// Text should not change for invalid characters
		assert.Equal(t, initialText, updatedTI.manualText, "Control character %q should be filtered", char)
		ti = updatedTI
	}

	// Test that valid characters are accepted
	validChars := []rune{'a', 'A', '1', ' ', '!', '@', '#'}

	for _, char := range validChars {
		initialText := ti.manualText
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}}
		updatedTI, _ := ti.Update(msg)

		// Text should include the valid character
		assert.Equal(t, initialText+string(char), updatedTI.manualText, "Valid character %q should be accepted", char)
		ti = updatedTI
	}
}

func TestWindowsCompatibleTextAreaEditingModeValue(t *testing.T) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.SetValue("Textarea content")

	// In normal mode, should return textarea value
	assert.Equal(t, "Textarea content", ta.Value())

	// Switch to editing mode
	ta.isEditing = true
	ta.manualText = "Manual content"

	// In editing mode, should return manual text
	assert.Equal(t, "Manual content", ta.Value())

	// Switch back to normal mode
	ta.isEditing = false

	// Should return textarea value again
	assert.Equal(t, "Textarea content", ta.Value())
}

func TestWindowsCompatibleTextInputEditingModeValue(t *testing.T) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.SetValue("Input content")

	// In normal mode, should return input value
	assert.Equal(t, "Input content", ti.Value())

	// Switch to editing mode
	ti.isEditing = true
	ti.manualText = "Manual content"

	// In editing mode, should return manual text
	assert.Equal(t, "Manual content", ti.Value())

	// Switch back to normal mode
	ti.isEditing = false

	// Should return input value again
	assert.Equal(t, "Input content", ti.Value())
}

// Integration tests for Windows compatibility
func TestWindowsCompatibilityIntegration(t *testing.T) {
	// This test verifies that Windows compatibility components work together

	// Test that Windows workaround functions don't crash on any platform
	assert.NotPanics(t, func() {
		IsWindowsInputWorkaroundNeeded()
	})

	assert.NotPanics(t, func() {
		input := CreateSafeTextInput("test")
		ApplyWindowsInputFix(&input)
	})

	assert.NotPanics(t, func() {
		ValidateInputCharacter("a")
		ValidateInputCharacter("\n")
		ValidateInputCharacter("")
	})

	assert.NotPanics(t, func() {
		SanitizeInputText("Hello\nWorld\t!")
	})

	// Test that Windows components can be created and used safely
	assert.NotPanics(t, func() {
		ta := NewWindowsCompatibleTextArea("test")
		ta.SetSize(100, 20)
		ta.SetValue("test content")
		ta.Focus()
		ta.Blur()
		ta.View()
	})

	assert.NotPanics(t, func() {
		ti := NewWindowsCompatibleTextInput("test")
		ti.SetWidth(60)
		ti.SetValue("test content")
		ti.Focus()
		ti.Blur()
		ti.View()
	})
}

func TestSlashCharacterHandling(t *testing.T) {
	// Test that slash character "/" is properly handled in Windows input
	textInput := NewWindowsCompatibleTextInput("URL")
	textInput.Focus()

	// Enter edit mode
	textInput, _ = textInput.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, textInput.isEditing)

	// Test typing slash character directly
	textInput, _ = textInput.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'/'},
	})

	// Verify slash was added
	assert.Contains(t, textInput.Value(), "/")

	// Test URL-like text with multiple slashes
	testText := "https://api.openai.com/v1"
	textInput.SetValue("")

	// Type each character including slashes
	for _, r := range testText {
		textInput, _ = textInput.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{r},
		})
	}

	// Verify complete URL was accepted
	assert.Equal(t, testText, textInput.Value())

	// Test ValidateInputCharacter function specifically for slash
	assert.True(t, ValidateInputCharacter("/"), "Slash character should be valid")
	assert.True(t, ValidateInputCharacter(":"), "Colon character should be valid")
	assert.True(t, ValidateInputCharacter("."), "Dot character should be valid")
}

func TestClipboardIntegration(t *testing.T) {
	// Test clipboard integration (mock test since we can't control real clipboard in tests)
	textInput := NewWindowsCompatibleTextInput("URL")
	textInput.Focus()

	// Enter edit mode
	textInput, _ = textInput.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.True(t, textInput.isEditing)

	// Test SanitizeInputText function with URL-like content
	testClipboard := "https://api.openai.com/v1/chat/completions"
	sanitized := SanitizeInputText(testClipboard)

	// Verify all URL characters are preserved
	assert.Equal(t, testClipboard, sanitized, "URL should be fully preserved after sanitization")
	assert.Contains(t, sanitized, "/", "Slashes should be preserved")
	assert.Contains(t, sanitized, ":", "Colons should be preserved")
	assert.Contains(t, sanitized, ".", "Dots should be preserved")
}

// Benchmark tests for performance validation
func BenchmarkWindowsCompatibleTextAreaUpdate(b *testing.B) {
	ta := NewWindowsCompatibleTextArea("Test placeholder")
	ta.Focus()
	ta.isEditing = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ta, _ = ta.Update(msg)
	}
}

func BenchmarkWindowsCompatibleTextInputUpdate(b *testing.B) {
	ti := NewWindowsCompatibleTextInput("Test placeholder")
	ti.Focus()
	ti.isEditing = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ti, _ = ti.Update(msg)
	}
}

func BenchmarkValidateInputCharacter(b *testing.B) {
	chars := []string{"a", "A", "1", " ", "!", "\n", "\t", "\x00"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		char := chars[i%len(chars)]
		ValidateInputCharacter(char)
	}
}

func BenchmarkSanitizeInputText(b *testing.B) {
	text := "Hello\nWorld\tTest\x00Content\x01More"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeInputText(text)
	}
}
