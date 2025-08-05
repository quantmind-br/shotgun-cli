package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test parseFrontmatter function
func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectedKey string
		expectedName string
		expectedDesc string
	}{
		{
			name: "Valid YAML frontmatter",
			content: `---
key: "test-template"
name: "Test Template"
description: "A test template for unit testing"
---
# Test Template Content
This is a test template with {TASK} and {RULES}.

{FILE_STRUCTURE}`,
			expectError:  false,
			expectedKey:  "test-template",
			expectedName: "Test Template",
			expectedDesc: "A test template for unit testing",
		},
		{
			name: "Valid YAML with special characters",
			content: `---
key: "special-chars"
name: "Template with Special Chars: @#$%"
description: "Description with newlines\nand special chars & symbols"
---
Template content here.`,
			expectError:  false,
			expectedKey:  "special-chars",
			expectedName: "Template with Special Chars: @#$%",
			expectedDesc: "Description with newlines\nand special chars & symbols",
		},
		{
			name: "Missing frontmatter delimiter",
			content: `key: "missing-delimiter"
name: "Missing Delimiter"
description: "No frontmatter delimiters"
---
Content here.`,
			expectError: true,
		},
		{
			name: "Unclosed frontmatter",
			content: `---
key: "unclosed"
name: "Unclosed Frontmatter"
description: "Missing closing delimiter"
Content without closing delimiter.`,
			expectError: true,
		},
		{
			name: "Invalid YAML syntax",
			content: `---
key: "invalid-yaml"
name: "Valid Name"
description: "Invalid YAML syntax"
invalid_yaml_here: [unclosed bracket
---
Content here.`,
			expectError: true,
		},
		{
			name: "Empty required fields",
			content: `---
key: ""
name: "Empty Key"
description: "Key is empty"
---
Content here.`,
			expectError: true,
		},
		{
			name: "Missing required fields",
			content: `---
key: "missing-fields"
name: "Missing Description"
---
Content here.`,
			expectError: true,
		},
		{
			name: "Empty content after frontmatter",
			content: `---
key: "empty-content"
name: "Empty Content"
description: "Template with no content"
---
`,
			expectError:  false,
			expectedKey:  "empty-content",
			expectedName: "Empty Content",
			expectedDesc: "Template with no content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseFrontmatter(tt.content)

			if tt.expectError {
				assert.Error(t, err, "Expected error for test case: %s", tt.name)
				assert.Nil(t, result, "Result should be nil when error occurs")
			} else {
				assert.NoError(t, err, "Unexpected error for test case: %s", tt.name)
				require.NotNil(t, result, "Result should not be nil for valid input")

				assert.Equal(t, tt.expectedKey, result.Metadata.Key, "Key mismatch")
				assert.Equal(t, tt.expectedName, result.Metadata.Name, "Name mismatch")
				assert.Equal(t, tt.expectedDesc, result.Metadata.Description, "Description mismatch")
				
				// Special case: empty content is valid for templates with no body
				if tt.name != "Empty content after frontmatter" {
					assert.NotEmpty(t, result.Content, "Content should not be empty")
				}
			}
		})
	}
}

// Test validateCustomTemplateMetadata function
func TestValidateCustomTemplateMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata CustomTemplateMetadata
		wantErr  bool
	}{
		{
			name: "Valid metadata",
			metadata: CustomTemplateMetadata{
				Key:         "valid-key",
				Name:        "Valid Name",
				Description: "Valid description",
			},
			wantErr: false,
		},
		{
			name: "Empty key",
			metadata: CustomTemplateMetadata{
				Key:         "",
				Name:        "Valid Name",
				Description: "Valid description",
			},
			wantErr: true,
		},
		{
			name: "Whitespace-only key",
			metadata: CustomTemplateMetadata{
				Key:         "   ",
				Name:        "Valid Name",
				Description: "Valid description",
			},
			wantErr: true,
		},
		{
			name: "Empty name",
			metadata: CustomTemplateMetadata{
				Key:         "valid-key",
				Name:        "",
				Description: "Valid description",
			},
			wantErr: true,
		},
		{
			name: "Empty description",
			metadata: CustomTemplateMetadata{
				Key:         "valid-key",
				Name:        "Valid Name",
				Description: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCustomTemplateMetadata(tt.metadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test loadCustomTemplate function
func TestLoadCustomTemplate(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "shotgun-test-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name         string
		filename     string
		content      string
		expectError  bool
		expectedKey  string
		expectedName string
	}{
		{
			name:     "Valid template file",
			filename: "valid_template.md",
			content: `---
key: "valid-template"
name: "Valid Template"
description: "A valid test template"
---
# Valid Template
This template has {TASK} and {RULES}.
Content: {FILE_STRUCTURE}`,
			expectError:  false,
			expectedKey:  "valid-template",
			expectedName: "Valid Template",
		},
		{
			name:     "Invalid YAML frontmatter",
			filename: "invalid_yaml.md",
			content: `---
key: "invalid-yaml"
name: "Invalid YAML"
description: "Invalid YAML syntax"
invalid_yaml: [unclosed_bracket
---
Content here.`,
			expectError: true,
		},
		{
			name:     "Missing required fields",
			filename: "missing_fields.md",
			content: `---
key: "missing-description"
name: "Missing Description Template"
---
Content here.`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(tempDir, tt.filename)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Test loadCustomTemplate
			templateInfo, content, err := loadCustomTemplate(filePath)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, templateInfo)
				assert.Empty(t, content)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, templateInfo)
				assert.Equal(t, tt.expectedKey, templateInfo.Key)
				assert.Equal(t, tt.expectedName, templateInfo.Name)
				assert.Equal(t, TemplateSourceCustom, templateInfo.Source)
				assert.Equal(t, filePath, templateInfo.FilePath)
				assert.NotEmpty(t, content)
			}
		})
	}

	// Test with non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "does_not_exist.md")
		templateInfo, content, err := loadCustomTemplate(nonExistentPath)
		assert.Error(t, err)
		assert.Nil(t, templateInfo)
		assert.Empty(t, content)
	})
}

// Test loadCustomTemplatesFromDirectory function
func TestLoadCustomTemplatesFromDirectory(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir, err := os.MkdirTemp("", "shotgun-test-templates-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test template files
	testTemplates := map[string]string{
		"template1.md": `---
key: "template-1"
name: "Template One"
description: "First test template"
---
Template 1 content with {TASK}.`,
		"template2.md": `---
key: "template-2"
name: "Template Two"
description: "Second test template"
---
Template 2 content with {RULES}.`,
		"invalid.md": `---
key: missing-quotes
name: "Invalid Template"
description: "Template with invalid YAML"
---
Invalid template content.`,
		"not_a_template.txt": "This is not a markdown file.",
		"duplicate_key.md": `---
key: "template-1"
name: "Duplicate Key Template"
description: "Template with duplicate key"
---
Duplicate template content.`,
	}

	// Write test files
	for filename, content := range testTemplates {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test loading templates
	templates, content, err := loadCustomTemplatesFromDirectory(tempDir)
	assert.NoError(t, err) // Should not error even with invalid templates

	// Should load 3 valid templates (template1.md, template2.md, and duplicate_key.md)
	// The invalid.md template is now valid YAML, non-.md files should be skipped
	assert.Len(t, templates, 3, "Should load exactly 3 valid templates")
	assert.Len(t, content, 3, "Should have content for 3 templates")

	// Check that valid templates are loaded
	templateKeys := make([]string, len(templates))
	for i, tmpl := range templates {
		templateKeys[i] = tmpl.Key
		assert.Equal(t, TemplateSourceCustom, tmpl.Source)
		assert.Contains(t, content, tmpl.Key)
	}

	assert.Contains(t, templateKeys, "template-1")
	assert.Contains(t, templateKeys, "template-2")
	assert.Contains(t, templateKeys, "missing-quotes") // The previously "invalid" template is now valid
}

// Test validateTemplateKeyConflicts function
func TestValidateTemplateKeyConflicts(t *testing.T) {
	// Mock builtin templates
	builtinTemplates := []TemplateInfo{
		{Key: "dev", Name: "Dev Template", Source: TemplateSourceBuiltin},
		{Key: "architect", Name: "Architect Template", Source: TemplateSourceBuiltin},
	}

	// Mock custom templates with some conflicts
	customTemplates := []TemplateInfo{
		{Key: "dev", Name: "Custom Dev Template", Source: TemplateSourceCustom},           // Conflict
		{Key: "architect", Name: "Custom Architect Template", Source: TemplateSourceCustom}, // Conflict
		{Key: "custom-1", Name: "Custom Template 1", Source: TemplateSourceCustom},         // No conflict
		{Key: "custom-2", Name: "Custom Template 2", Source: TemplateSourceCustom},         // No conflict
	}

	validCustomTemplates := validateTemplateKeyConflicts(builtinTemplates, customTemplates)

	// Should only return templates without conflicts
	assert.Len(t, validCustomTemplates, 2, "Should filter out conflicting templates")

	validKeys := make([]string, len(validCustomTemplates))
	for i, tmpl := range validCustomTemplates {
		validKeys[i] = tmpl.Key
	}

	assert.Contains(t, validKeys, "custom-1")
	assert.Contains(t, validKeys, "custom-2")
	assert.NotContains(t, validKeys, "dev")
	assert.NotContains(t, validKeys, "architect")
}

// Test GetTemplateInfoByKey helper function
func TestGetTemplateInfoByKey(t *testing.T) {
	templates := []TemplateInfo{
		{Key: "template-1", Name: "Template One"},
		{Key: "template-2", Name: "Template Two"},
	}

	// Test existing key
	tmpl, found := GetTemplateInfoByKey(templates, "template-1")
	assert.True(t, found)
	assert.Equal(t, "Template One", tmpl.Name)

	// Test non-existing key
	_, found = GetTemplateInfoByKey(templates, "non-existent")
	assert.False(t, found)
}

// Integration test for SimpleTemplateProcessor with custom templates
func TestSimpleTemplateProcessorIntegration(t *testing.T) {
	// Create temporary directories
	builtinDir, err := os.MkdirTemp("", "shotgun-builtin-*")
	require.NoError(t, err)
	defer os.RemoveAll(builtinDir)

	customDir, err := os.MkdirTemp("", "shotgun-custom-*")
	require.NoError(t, err)
	defer os.RemoveAll(customDir)

	// Create all required builtin template files (from AvailableTemplates)
	builtinContent := `Test builtin template with {TASK} and {RULES}.
File structure: {FILE_STRUCTURE}`
	
	builtinFiles := []string{
		"prompt_makeDiffGitFormat.md",
		"prompt_makePlan.md", 
		"prompt_analyzeBug.md",
		"prompt_projectManager.md",
	}
	
	for _, filename := range builtinFiles {
		err = os.WriteFile(filepath.Join(builtinDir, filename), []byte(builtinContent), 0644)
		require.NoError(t, err)
	}

	// Create custom template
	customContent := `---
key: "custom-test"
name: "Custom Test Template"
description: "A custom template for testing"
---
Custom template with {TASK} and {RULES}.
Files: {FILE_STRUCTURE}`
	err = os.WriteFile(filepath.Join(customDir, "custom_test.md"), []byte(customContent), 0644)
	require.NoError(t, err)

	// Test SimpleTemplateProcessor
	processor := NewSimpleTemplateProcessor()
	err = processor.LoadTemplates(builtinDir, customDir)
	assert.NoError(t, err)

	// Test that both builtin and custom templates are loaded
	allTemplates := processor.GetAllTemplateInfos()
	builtinTemplates := processor.GetBuiltinTemplateInfos()
	customTemplates := processor.GetCustomTemplateInfos()
	
	// We expect 4 builtin templates (from AvailableTemplates) and 1 custom template
	assert.Len(t, builtinTemplates, 4, "Should have 4 builtin templates")
	assert.Len(t, customTemplates, 1, "Should have 1 custom template")
	assert.Len(t, allTemplates, 5, "Should load 4 builtin + 1 custom template")

	// Test template content retrieval
	content, err := processor.GetTemplateContent("custom-test")
	assert.NoError(t, err)
	assert.Contains(t, content, "Custom template with")

	// Test prompt generation with custom template
	data := TemplateData{
		Task:          "Test task",
		Rules:         "Test rules",
		FileStructure: "Test file structure",
	}
	
	prompt, err := processor.GeneratePrompt("custom-test", data)
	assert.NoError(t, err)
	assert.Contains(t, prompt, "Test task")
	assert.Contains(t, prompt, "Test rules")
	assert.Contains(t, prompt, "Test file structure")
	assert.NotContains(t, prompt, "{TASK}") // Should be replaced
}

// Benchmark tests
func BenchmarkParseFrontmatter(b *testing.B) {
	content := `---
key: "benchmark-template"
name: "Benchmark Template"
description: "Template for benchmarking frontmatter parsing"
---
# Benchmark Template
This is benchmark content with {TASK} and {RULES}.
File structure: {FILE_STRUCTURE}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseFrontmatter(content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoadCustomTemplate(b *testing.B) {
	// Create temporary file
	tempDir, err := os.MkdirTemp("", "shotgun-benchmark-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	content := `---
key: "benchmark-template"
name: "Benchmark Template"
description: "Template for benchmarking"
---
Benchmark template content with {TASK}, {RULES}, and {FILE_STRUCTURE}.`

	filePath := filepath.Join(tempDir, "benchmark.md")
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := loadCustomTemplate(filePath)
		if err != nil {
			b.Fatal(err)
		}
	}
}