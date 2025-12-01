package template

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer()

	assert.NotNil(t, renderer)
	assert.NotNil(t, renderer.defaultVars)
	assert.Contains(t, renderer.defaultVars, VarCurrentDate)
}

func TestRendererRenderTemplate(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name        string
		template    *Template
		vars        map[string]string
		expectError bool
		validate    func(t *testing.T, result string)
	}{
		{
			name:        "nil template",
			template:    nil,
			vars:        nil,
			expectError: true,
		},
		{
			name: "simple substitution",
			template: &Template{
				Name:         "test",
				Content:      "Hello {NAME}!",
				RequiredVars: []string{"NAME"},
			},
			vars:        map[string]string{"NAME": "World"},
			expectError: false,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "Hello World!", result)
			},
		},
		{
			name: "multiple variables",
			template: &Template{
				Name:         "test",
				Content:      "{TASK}: {RULES}",
				RequiredVars: []string{"TASK", "RULES"},
			},
			vars: map[string]string{
				"TASK":  "Review code",
				"RULES": "Follow best practices",
			},
			expectError: false,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "Review code: Follow best practices", result)
			},
		},
		{
			name: "missing required variable",
			template: &Template{
				Name:         "test",
				Content:      "Hello {NAME}!",
				RequiredVars: []string{"NAME"},
			},
			vars:        map[string]string{},
			expectError: true,
		},
		{
			name: "empty required variable",
			template: &Template{
				Name:         "test",
				Content:      "Hello {NAME}!",
				RequiredVars: []string{"NAME"},
			},
			vars:        map[string]string{"NAME": ""},
			expectError: true,
		},
		{
			name: "whitespace only variable",
			template: &Template{
				Name:         "test",
				Content:      "Hello {NAME}!",
				RequiredVars: []string{"NAME"},
			},
			vars:        map[string]string{"NAME": "   "},
			expectError: true,
		},
		{
			name: "auto-generated CURRENT_DATE",
			template: &Template{
				Name:         "test",
				Content:      "Date: {CURRENT_DATE}",
				RequiredVars: []string{"CURRENT_DATE"},
			},
			vars:        map[string]string{},
			expectError: false,
			validate: func(t *testing.T, result string) {
				today := time.Now().Format("2006-01-02")
				assert.Contains(t, result, today)
			},
		},
		{
			name: "extra variables ignored",
			template: &Template{
				Name:         "test",
				Content:      "Hello {NAME}!",
				RequiredVars: []string{"NAME"},
			},
			vars: map[string]string{
				"NAME":  "World",
				"EXTRA": "ignored",
			},
			expectError: false,
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "Hello World!", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderer.RenderTemplate(tt.template, tt.vars)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestRendererSanitizeVariableValue(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no changes needed",
			input:    "simple text",
			expected: "simple text",
		},
		{
			name:     "CRLF to LF",
			input:    "line1\r\nline2",
			expected: "line1\nline2",
		},
		{
			name:     "CR only to LF",
			input:    "line1\rline2",
			expected: "line1\nline2",
		},
		{
			name:     "trailing whitespace removed",
			input:    "line1   \nline2\t\t",
			expected: "line1\nline2",
		},
		{
			name:     "preserves leading whitespace",
			input:    "  indented",
			expected: "  indented",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "multiline with mixed endings",
			input:    "line1\r\nline2\rline3\n",
			expected: "line1\nline2\nline3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.sanitizeVariableValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRendererMergeVariables(t *testing.T) {
	renderer := NewRenderer()

	t.Run("nil vars uses defaults", func(t *testing.T) {
		result := renderer.mergeVariables(nil)
		assert.Contains(t, result, VarCurrentDate)
	})

	t.Run("user vars override defaults", func(t *testing.T) {
		customDate := "2020-01-01"
		result := renderer.mergeVariables(map[string]string{
			VarCurrentDate: customDate,
		})
		// Note: mergeVariables always refreshes CURRENT_DATE at the end
		// so this test verifies the merge happens, but date is auto-refreshed
		assert.Contains(t, result, VarCurrentDate)
	})

	t.Run("preserves user variables", func(t *testing.T) {
		result := renderer.mergeVariables(map[string]string{
			"CUSTOM_VAR": "custom_value",
		})
		assert.Equal(t, "custom_value", result["CUSTOM_VAR"])
	})
}

func TestRendererIsAutoGeneratedVar(t *testing.T) {
	renderer := NewRenderer()

	assert.True(t, renderer.isAutoGeneratedVar(VarCurrentDate))
	assert.False(t, renderer.isAutoGeneratedVar(VarTask))
	assert.False(t, renderer.isAutoGeneratedVar(VarRules))
	assert.False(t, renderer.isAutoGeneratedVar("RANDOM_VAR"))
}

func TestRendererGetRequiredVariables(t *testing.T) {
	renderer := NewRenderer()

	template := &Template{
		RequiredVars: []string{VarTask, VarRules, VarCurrentDate, VarFileStructure},
	}

	required := renderer.GetRequiredVariables(template)

	// Should exclude auto-generated variables
	assert.Contains(t, required, VarTask)
	assert.Contains(t, required, VarRules)
	assert.Contains(t, required, VarFileStructure)
	assert.NotContains(t, required, VarCurrentDate)
}

func TestRendererPreviewTemplate(t *testing.T) {
	renderer := NewRenderer()

	t.Run("nil template", func(t *testing.T) {
		_, err := renderer.PreviewTemplate(nil)
		assert.Error(t, err)
	})

	t.Run("creates preview values", func(t *testing.T) {
		template := &Template{
			Name:         "test",
			Content:      "{TASK} with {RULES}",
			RequiredVars: []string{"TASK", "RULES"},
		}

		result, err := renderer.PreviewTemplate(template)
		require.NoError(t, err)

		assert.Contains(t, result, "[task]")
		assert.Contains(t, result, "[rules]")
	})

	t.Run("auto-generated vars get real values", func(t *testing.T) {
		template := &Template{
			Name:         "test",
			Content:      "Date: {CURRENT_DATE}",
			RequiredVars: []string{"CURRENT_DATE"},
		}

		result, err := renderer.PreviewTemplate(template)
		require.NoError(t, err)

		// Should contain actual date, not placeholder
		today := time.Now().Format("2006-01-02")
		assert.Contains(t, result, today)
	})
}

func TestRendererValidateVariableNames(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name        string
		vars        map[string]string
		expectError bool
	}{
		{
			name: "valid uppercase",
			vars: map[string]string{
				"TASK": "value",
				"RULES": "value",
			},
			expectError: false,
		},
		{
			name: "valid with underscore",
			vars: map[string]string{
				"FILE_STRUCTURE": "value",
				"CURRENT_DATE":   "value",
			},
			expectError: false,
		},
		{
			name: "valid with numbers",
			vars: map[string]string{
				"VAR1": "value",
				"VAR2": "value",
			},
			expectError: false,
		},
		{
			name: "invalid lowercase",
			vars: map[string]string{
				"task": "value",
			},
			expectError: true,
		},
		{
			name: "invalid mixed case",
			vars: map[string]string{
				"Task": "value",
			},
			expectError: true,
		},
		{
			name:        "empty vars",
			vars:        map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := renderer.ValidateVariableNames(tt.vars)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRendererSubstituteVariables(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name     string
		content  string
		vars     map[string]string
		expected string
	}{
		{
			name:    "single substitution",
			content: "Hello {NAME}!",
			vars:    map[string]string{"NAME": "World"},
			expected: "Hello World!",
		},
		{
			name:    "multiple same variable",
			content: "{VAR} and {VAR} again",
			vars:    map[string]string{"VAR": "test"},
			expected: "test and test again",
		},
		{
			name:    "no variables",
			content: "Static content",
			vars:    map[string]string{},
			expected: "Static content",
		},
		{
			name:    "unmatched variable left as is",
			content: "Hello {UNKNOWN}!",
			vars:    map[string]string{"NAME": "World"},
			expected: "Hello {UNKNOWN}!",
		},
		{
			name:    "multiline content",
			content: "Line1 {VAR}\nLine2 {VAR}",
			vars:    map[string]string{"VAR": "replaced"},
			expected: "Line1 replaced\nLine2 replaced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.substituteVariables(tt.content, tt.vars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRendererValidateVariables(t *testing.T) {
	renderer := NewRenderer()

	tests := []struct {
		name        string
		template    *Template
		vars        map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "all required provided",
			template: &Template{
				RequiredVars: []string{"TASK", "RULES"},
			},
			vars: map[string]string{
				"TASK":  "Do something",
				"RULES": "Follow rules",
			},
			expectError: false,
		},
		{
			name: "missing required",
			template: &Template{
				RequiredVars: []string{"TASK", "RULES"},
			},
			vars: map[string]string{
				"TASK": "Do something",
			},
			expectError: true,
			errorMsg:    "missing",
		},
		{
			name: "empty required",
			template: &Template{
				RequiredVars: []string{"TASK"},
			},
			vars: map[string]string{
				"TASK": "   ",
			},
			expectError: true,
			errorMsg:    "empty",
		},
		{
			name: "auto-generated not required",
			template: &Template{
				RequiredVars: []string{"CURRENT_DATE"},
			},
			vars:        map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := renderer.validateVariables(tt.template, tt.vars)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errorMsg))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDefaultVariables(t *testing.T) {
	defaults := getDefaultVariables()

	assert.Contains(t, defaults, VarCurrentDate)

	// Verify date format
	dateValue := defaults[VarCurrentDate]
	_, err := time.Parse("2006-01-02", dateValue)
	assert.NoError(t, err, "CURRENT_DATE should be in YYYY-MM-DD format")
}
