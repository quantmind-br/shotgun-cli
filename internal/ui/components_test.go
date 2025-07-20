package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewNumberedTextArea(t *testing.T) {
	nta := NewNumberedTextArea()
	
	if !nta.IsAutoNumberingEnabled() {
		t.Error("NewNumberedTextArea() should have auto-numbering enabled by default")
	}
	
	if nta.Value() != "" {
		t.Error("NewNumberedTextArea() should start with empty value")
	}
}

func TestNumberedTextAreaBasicMethods(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test SetPlaceholder
	nta.SetPlaceholder("Test placeholder")
	
	// Test SetValue and Value
	testValue := "Test content"
	nta.SetValue(testValue)
	if nta.Value() != testValue {
		t.Errorf("Value() = %q, want %q", nta.Value(), testValue)
	}
	
	// Test dimensions
	nta.SetWidth(80)
	nta.SetHeight(10)
	
	// Test auto-numbering toggle
	nta.EnableAutoNumbering(false)
	if nta.IsAutoNumberingEnabled() {
		t.Error("EnableAutoNumbering(false) should disable auto-numbering")
	}
	
	nta.EnableAutoNumbering(true)
	if !nta.IsAutoNumberingEnabled() {
		t.Error("EnableAutoNumbering(true) should enable auto-numbering")
	}
}

func TestNumberedTextAreaGetNumberedValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			expected: "",
		},
		{
			name:     "single line unnumbered",
			input:    "First task",
			expected: "1. First task",
		},
		{
			name:     "multiple lines unnumbered",
			input:    "First task\nSecond task\nThird task",
			expected: "1. First task\n2. Second task\n3. Third task",
		},
		{
			name:     "already numbered content",
			input:    "1. First task\n2. Second task",
			expected: "1. First task\n2. Second task",
		},
		{
			name:     "inconsistent numbering",
			input:    "1. First task\n3. Second task\n2. Third task",
			expected: "1. First task\n2. Second task\n3. Third task",
		},
		{
			name:     "mixed numbered and unnumbered",
			input:    "1. First task\nUnnumbered task\n2. Another task",
			expected: "1. First task\n2. Unnumbered task\n3. Another task",
		},
		{
			name:     "content with empty lines",
			input:    "First task\n\nSecond task\n\nThird task",
			expected: "1. First task\n\n2. Second task\n\n3. Third task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nta := NewNumberedTextArea()
			nta.SetValue(tt.input)
			
			result := nta.GetNumberedValue()
			if result != tt.expected {
				t.Errorf("GetNumberedValue() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNumberedTextAreaAutoNumberingBehavior(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test Enter key handling with empty content
	nta.SetValue("")
	nta, _ = nta.Update(tea.KeyMsg{Type: tea.KeyEnter})
	
	// Value should be "1. " after pressing Enter on empty content
	expectedAfterEnter := "1. "
	if nta.Value() != expectedAfterEnter {
		t.Errorf("After Enter on empty content: Value() = %q, want %q", nta.Value(), expectedAfterEnter)
	}
}

func TestNumberedTextAreaHandleAutoNumbering(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
		desc     string
	}{
		{
			name:     "first line numbering",
			initial:  "",
			expected: "1. ",
			desc:     "Empty content should get first number",
		},
		{
			name:     "continue numbering sequence",
			initial:  "1. First task\n",
			expected: "1. First task\n2. ",
			desc:     "Should continue numbering after existing numbered line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nta := NewNumberedTextArea()
			nta.SetValue(tt.initial)
			
			// Simulate the auto-numbering behavior
			nta = nta.handleAutoNumbering()
			
			if nta.Value() != tt.expected {
				t.Errorf("handleAutoNumbering() = %q, want %q (test: %s)", nta.Value(), tt.expected, tt.desc)
			}
		})
	}
}

func TestNumberedTextAreaView(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test that View() returns a string (basic rendering test)
	view := nta.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}
	
	// Test with auto-numbering disabled
	nta.EnableAutoNumbering(false)
	viewDisabled := nta.View()
	if viewDisabled == "" {
		t.Error("View() should return non-empty string even with auto-numbering disabled")
	}
}

func TestNumberedTextAreaUpdate(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test that Update returns the model and a command
	_, cmd := nta.Update(tea.KeyMsg{Type: tea.KeyEnter})
	
	// cmd can be nil, that's fine
	_ = cmd
	
	// Test various key inputs
	testKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("test")},
		{Type: tea.KeyBackspace},
		{Type: tea.KeyLeft},
		{Type: tea.KeyRight},
	}
	
	for _, key := range testKeys {
		nta, _ = nta.Update(key)
		// Just ensure it doesn't panic
	}
}

func TestNumberedTextAreaFocusManagement(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test focus methods
	if nta.Focused() {
		t.Error("NewNumberedTextArea() should not be focused initially")
	}
	
	// Focus the component
	nta.Focus()
	
	// Note: We can't easily test if it's actually focused without 
	// more complex setup, but we can ensure the methods don't panic
	
	// Blur the component
	nta.Blur()
}

func TestNumberedTextAreaEdgeCases(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test with only whitespace
	nta.SetValue("   \n  \t  \n   ")
	result := nta.GetNumberedValue()
	// Should preserve empty lines but not add numbers to whitespace-only lines
	expected := "\n\n"
	if result != expected {
		t.Errorf("GetNumberedValue() with only whitespace = %q, want %q", result, expected)
	}
	
	// Test with very long content
	longContent := strings.Repeat("This is a very long line that might cause issues. ", 100)
	nta.SetValue(longContent)
	numberedLong := nta.GetNumberedValue()
	if !strings.HasPrefix(numberedLong, "1. ") {
		t.Error("GetNumberedValue() should number even very long content")
	}
	
	// Test with special characters
	specialContent := "Special chars: !@#$%^&*()_+{}|:\"<>?[]\\;',./"
	nta.SetValue(specialContent)
	numberedSpecial := nta.GetNumberedValue()
	expectedSpecial := "1. " + specialContent
	if numberedSpecial != expectedSpecial {
		t.Errorf("GetNumberedValue() with special chars = %q, want %q", numberedSpecial, expectedSpecial)
	}
}

func TestNumberedTextAreaIntegration(t *testing.T) {
	// Test that NumberedTextArea can be used in a larger context
	nta1 := NewNumberedTextArea()
	nta2 := NewNumberedTextArea()
	
	// Set different content in each
	nta1.SetValue("Task 1\nTask 2")
	nta2.SetValue("Rule 1\nRule 2")
	
	// Get numbered values
	tasks := nta1.GetNumberedValue()
	rules := nta2.GetNumberedValue()
	
	expectedTasks := "1. Task 1\n2. Task 2"
	expectedRules := "1. Rule 1\n2. Rule 2"
	
	if tasks != expectedTasks {
		t.Errorf("Integration test tasks = %q, want %q", tasks, expectedTasks)
	}
	
	if rules != expectedRules {
		t.Errorf("Integration test rules = %q, want %q", rules, expectedRules)
	}
}

func TestNumberedTextAreaErrorHandling(t *testing.T) {
	nta := NewNumberedTextArea()
	
	// Test that methods handle invalid input gracefully
	nta.SetValue("")
	nta.SetWidth(0)
	nta.SetHeight(0)
	
	// These shouldn't panic
	_ = nta.View()
	_ = nta.GetNumberedValue()
	nta, _ = nta.Update(tea.KeyMsg{})
}

func BenchmarkNumberedTextAreaGetNumberedValue(b *testing.B) {
	nta := NewNumberedTextArea()
	content := "Task 1\nTask 2\nTask 3\nTask 4\nTask 5"
	nta.SetValue(content)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nta.GetNumberedValue()
	}
}

func BenchmarkNumberedTextAreaUpdate(b *testing.B) {
	nta := NewNumberedTextArea()
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nta, _ = nta.Update(key)
	}
}