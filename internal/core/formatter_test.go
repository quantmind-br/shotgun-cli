package core

import (
	"testing"
)

func TestParseNumberedLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NumberedLine
	}{
		{
			name:  "empty string",
			input: "",
			expected: []NumberedLine{},
		},
		{
			name:  "single numbered line",
			input: "1. First task",
			expected: []NumberedLine{
				{Number: 1, Content: "First task", IsEmpty: false},
			},
		},
		{
			name:  "multiple numbered lines",
			input: "1. First task\n2. Second task\n3. Third task",
			expected: []NumberedLine{
				{Number: 1, Content: "First task", IsEmpty: false},
				{Number: 2, Content: "Second task", IsEmpty: false},
				{Number: 3, Content: "Third task", IsEmpty: false},
			},
		},
		{
			name:  "mixed numbered and unnumbered content",
			input: "1. First task\nUnnumbered content\n2. Second task",
			expected: []NumberedLine{
				{Number: 1, Content: "First task", IsEmpty: false},
				{Number: 0, Content: "Unnumbered content", IsEmpty: false},
				{Number: 2, Content: "Second task", IsEmpty: false},
			},
		},
		{
			name:  "empty lines",
			input: "1. First task\n\n2. Second task",
			expected: []NumberedLine{
				{Number: 1, Content: "First task", IsEmpty: false},
				{Number: 0, Content: "", IsEmpty: true},
				{Number: 2, Content: "Second task", IsEmpty: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNumberedLines(tt.input)
			if err != nil {
				t.Fatalf("ParseNumberedLines() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("ParseNumberedLines() length = %v, want %v", len(result), len(tt.expected))
			}

			for i, line := range result {
				expected := tt.expected[i]
				if line.Number != expected.Number || line.Content != expected.Content || line.IsEmpty != expected.IsEmpty {
					t.Errorf("ParseNumberedLines() line %d = %+v, want %+v", i, line, expected)
				}
			}
		})
	}
}

func TestFormatNumberedText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  FormattingOptions
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			options:  DefaultFormattingOptions(),
			expected: "",
		},
		{
			name:     "single unnumbered line",
			input:    "First task",
			options:  DefaultFormattingOptions(),
			expected: "1. First task",
		},
		{
			name:     "multiple unnumbered lines",
			input:    "First task\nSecond task\nThird task",
			options:  DefaultFormattingOptions(),
			expected: "1. First task\n2. Second task\n3. Third task",
		},
		{
			name:     "already numbered text",
			input:    "1. First task\n2. Second task",
			options:  DefaultFormattingOptions(),
			expected: "1. First task\n2. Second task",
		},
		{
			name:     "inconsistent numbering",
			input:    "1. First task\n3. Second task\n2. Third task",
			options:  DefaultFormattingOptions(),
			expected: "1. First task\n2. Second task\n3. Third task",
		},
		{
			name:    "numbering disabled",
			input:   "1. First task\n2. Second task",
			options: FormattingOptions{EnableNumbering: false},
			expected: "First task\nSecond task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatNumberedText(tt.input, tt.options)
			if err != nil {
				t.Fatalf("FormatNumberedText() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("FormatNumberedText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestReformatWithNumbers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unnumbered tasks",
			input:    "Create login form\nAdd validation\nTest the feature",
			expected: "1. Create login form\n2. Add validation\n3. Test the feature",
		},
		{
			name:     "mixed content",
			input:    "1. First task\nUnnumbered task\n3. Another task",
			expected: "1. First task\n2. Unnumbered task\n3. Another task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ReformatWithNumbers(tt.input)
			if err != nil {
				t.Fatalf("ReformatWithNumbers() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("ReformatWithNumbers() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateNumberedFormat(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValid bool
		expectedIssues int
	}{
		{
			name:           "valid numbered text",
			input:          "1. First task\n2. Second task\n3. Third task",
			expectedValid:  true,
			expectedIssues: 0,
		},
		{
			name:           "invalid numbering sequence",
			input:          "1. First task\n3. Second task\n2. Third task",
			expectedValid:  false,
			expectedIssues: 2,
		},
		{
			name:           "mixed numbered and unnumbered",
			input:          "1. First task\nUnnumbered content\n2. Second task",
			expectedValid:  false,
			expectedIssues: 1,
		},
		{
			name:           "unnumbered content only",
			input:          "First task\nSecond task\nThird task",
			expectedValid:  true,
			expectedIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, issues, err := ValidateNumberedFormat(tt.input)
			if err != nil {
				t.Fatalf("ValidateNumberedFormat() error = %v", err)
			}

			if valid != tt.expectedValid {
				t.Errorf("ValidateNumberedFormat() valid = %v, want %v", valid, tt.expectedValid)
			}

			if len(issues) != tt.expectedIssues {
				t.Errorf("ValidateNumberedFormat() issues count = %v, want %v", len(issues), tt.expectedIssues)
			}
		})
	}
}

func TestHasNumberedContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "has numbered content",
			input:    "1. First task\n2. Second task",
			expected: true,
		},
		{
			name:     "no numbered content",
			input:    "First task\nSecond task",
			expected: false,
		},
		{
			name:     "mixed content",
			input:    "Unnumbered\n1. Numbered\nMore unnumbered",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasNumberedContent(tt.input)
			if result != tt.expected {
				t.Errorf("HasNumberedContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSplitIntoTasks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "numbered tasks",
			input:    "1. First task\n2. Second task\n3. Third task",
			expected: []string{"First task", "Second task", "Third task"},
		},
		{
			name:     "mixed content with empty lines",
			input:    "1. First task\n\n2. Second task\n\n3. Third task",
			expected: []string{"First task", "Second task", "Third task"},
		},
		{
			name:     "unnumbered content",
			input:    "First task\nSecond task\nThird task",
			expected: []string{"First task", "Second task", "Third task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SplitIntoTasks(tt.input)
			if err != nil {
				t.Fatalf("SplitIntoTasks() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("SplitIntoTasks() length = %v, want %v", len(result), len(tt.expected))
			}

			for i, task := range result {
				if task != tt.expected[i] {
					t.Errorf("SplitIntoTasks() task %d = %q, want %q", i, task, tt.expected[i])
				}
			}
		})
	}
}