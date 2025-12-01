package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTemplateName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{"simple markdown", "template.md", "template"},
		{"with prompt prefix", "prompt_review.md", "review"},
		{"no extension", "template", "template"},
		{"nested name", "my_template.md", "my_template"},
		{"uppercase", "TEMPLATE.md", "TEMPLATE"},
		{"double extension", "template.test.md", "template.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTemplateName(tt.fileName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		fileName string
		expected string
	}{
		{
			name:     "markdown header",
			content:  "# Code Review Template\n\nThis is the content.",
			fileName: "review.md",
			expected: "Code Review Template",
		},
		{
			name:     "html comment",
			content:  "<!-- Template for code analysis -->\nContent here",
			fileName: "analysis.md",
			expected: "Template for code analysis",
		},
		{
			name:     "multiple headers uses first",
			content:  "# First Header\n## Second Header\nContent",
			fileName: "test.md",
			expected: "First Header",
		},
		{
			name:     "no header fallback",
			content:  "Just some content without headers",
			fileName: "custom.md",
			expected: "Template for custom",
		},
		{
			name:     "empty header line",
			content:  "#\n# Actual Title\nContent",
			fileName: "test.md",
			expected: "Actual Title",
		},
		{
			name:     "h2 header",
			content:  "## Secondary Title\nContent",
			fileName: "test.md",
			expected: "Secondary Title",
		},
		{
			name:     "empty content",
			content:  "",
			fileName: "empty.md",
			expected: "Template for empty",
		},
		{
			name:     "with prompt prefix in filename",
			content:  "No header here",
			fileName: "prompt_test.md",
			expected: "Template for test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDescription(tt.content, tt.fileName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRequiredVars(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single variable",
			content:  "Hello {NAME}!",
			expected: []string{"NAME"},
		},
		{
			name:     "multiple variables",
			content:  "{TASK} should follow {RULES}",
			expected: []string{"TASK", "RULES"},
		},
		{
			name:     "repeated variable counted once",
			content:  "{VAR} and {VAR} again",
			expected: []string{"VAR"},
		},
		{
			name:     "underscore in name",
			content:  "{FILE_STRUCTURE} and {CURRENT_DATE}",
			expected: []string{"FILE_STRUCTURE", "CURRENT_DATE"},
		},
		{
			name:     "no variables",
			content:  "Plain text without variables",
			expected: []string{},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "numbers in variable name",
			content:  "{VAR1} and {VAR2}",
			expected: []string{"VAR1", "VAR2"},
		},
		{
			name:     "lowercase not matched",
			content:  "{lowercase} should not match",
			expected: []string{},
		},
		{
			name:     "mixed case not matched",
			content:  "{MixedCase} should not match",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractRequiredVars(tt.content)
			require.NoError(t, err)

			// Convert to map for easier comparison (order independent)
			resultMap := make(map[string]bool)
			for _, v := range result {
				resultMap[v] = true
			}

			expectedMap := make(map[string]bool)
			for _, v := range tt.expected {
				expectedMap[v] = true
			}

			assert.Equal(t, expectedMap, resultMap)
		})
	}
}

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		fileName    string
		filePath    string
		expectError bool
		validate    func(t *testing.T, tmpl *Template)
	}{
		{
			name:     "valid template",
			content:  "# Review Template\n\n{TASK}\n{RULES}",
			fileName: "review.md",
			filePath: "/templates/review.md",
			validate: func(t *testing.T, tmpl *Template) {
				assert.Equal(t, "review", tmpl.Name)
				assert.Equal(t, "Review Template", tmpl.Description)
				assert.Contains(t, tmpl.RequiredVars, "TASK")
				assert.Contains(t, tmpl.RequiredVars, "RULES")
				assert.Equal(t, "/templates/review.md", tmpl.FilePath)
				assert.True(t, tmpl.IsEmbedded)
			},
		},
		{
			name:        "empty content",
			content:     "",
			fileName:    "empty.md",
			filePath:    "/templates/empty.md",
			expectError: true,
		},
		{
			name:     "template without variables",
			content:  "# Simple Template\n\nJust static content.",
			fileName: "simple.md",
			filePath: "/templates/simple.md",
			validate: func(t *testing.T, tmpl *Template) {
				assert.Equal(t, "simple", tmpl.Name)
				assert.Len(t, tmpl.RequiredVars, 0)
			},
		},
		{
			name:     "with prompt prefix",
			content:  "# Analysis\n{FILE_STRUCTURE}",
			fileName: "prompt_analysis.md",
			filePath: "/templates/prompt_analysis.md",
			validate: func(t *testing.T, tmpl *Template) {
				assert.Equal(t, "analysis", tmpl.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := parseTemplate(tt.content, tt.fileName, tt.filePath)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, tmpl)

			if tt.validate != nil {
				tt.validate(t, tmpl)
			}
		})
	}
}

func TestValidateTemplateContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "empty content",
			content:     "",
			expectError: true,
		},
		{
			name:        "valid content",
			content:     "Hello {WORLD}!",
			expectError: false,
		},
		{
			name:        "unmatched brace open",
			content:     "Hello {WORLD",
			expectError: true,
		},
		{
			name:        "unmatched brace close",
			content:     "Hello WORLD}",
			expectError: true,
		},
		{
			name:        "balanced braces in code block",
			content:     "```\nif (x) { y }\n```",
			expectError: false,
		},
		{
			name:        "unbalanced in regular line",
			content:     "Text with {unbalanced",
			expectError: true,
		},
		{
			name:        "json-like content",
			content:     `{"key": "value"}`,
			expectError: false,
		},
		{
			name:        "multiline with code block",
			content:     "# Template\n\n```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n\n{TASK}",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTemplateContent(tt.content)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateGetVariableNames(t *testing.T) {
	tmpl := &Template{
		Content: "Hello {NAME}, your task is {TASK}. Date: {CURRENT_DATE}",
	}

	vars := tmpl.GetVariableNames()

	assert.Len(t, vars, 3)
	// Check that all expected variables are present
	varMap := make(map[string]bool)
	for _, v := range vars {
		varMap[v] = true
	}
	assert.True(t, varMap["NAME"])
	assert.True(t, varMap["TASK"])
	assert.True(t, varMap["CURRENT_DATE"])
}

func TestTemplateHasVariable(t *testing.T) {
	tmpl := &Template{
		Content: "Task: {TASK}\nRules: {RULES}",
	}

	assert.True(t, tmpl.HasVariable("TASK"))
	assert.True(t, tmpl.HasVariable("RULES"))
	assert.False(t, tmpl.HasVariable("NOTEXIST"))
	assert.False(t, tmpl.HasVariable("task")) // case sensitive
}

func TestTemplateGetVariableCount(t *testing.T) {
	tmpl := &Template{
		Content: "{VAR} appears once, {OTHER} appears {OTHER} twice",
	}

	assert.Equal(t, 1, tmpl.GetVariableCount("VAR"))
	assert.Equal(t, 2, tmpl.GetVariableCount("OTHER"))
	assert.Equal(t, 0, tmpl.GetVariableCount("NOTEXIST"))
}

func TestTemplateIsValid(t *testing.T) {
	tests := []struct {
		name        string
		template    Template
		expectError bool
	}{
		{
			name: "valid template",
			template: Template{
				Name:    "test",
				Content: "Hello {WORLD}",
			},
			expectError: false,
		},
		{
			name: "empty name",
			template: Template{
				Name:    "",
				Content: "Hello {WORLD}",
			},
			expectError: true,
		},
		{
			name: "empty content",
			template: Template{
				Name:    "test",
				Content: "",
			},
			expectError: true,
		},
		{
			name: "invalid content",
			template: Template{
				Name:    "test",
				Content: "Unbalanced {brace",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.IsValid()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVariablePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"uppercase only", "{TASK}", []string{"{TASK}"}},
		{"with underscore", "{FILE_STRUCTURE}", []string{"{FILE_STRUCTURE}"}},
		{"with numbers", "{VAR123}", []string{"{VAR123}"}},
		{"multiple", "{A} and {B}", []string{"{A}", "{B}"}},
		{"lowercase no match", "{task}", nil},
		{"mixed case no match", "{Task}", nil},
		{"starts with number no match", "{1VAR}", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := variablePattern.FindAllString(tt.input, -1)
			assert.Equal(t, tt.expected, matches)
		})
	}
}
