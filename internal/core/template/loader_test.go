package template

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/assets"
)

func TestEmbeddedSource_LoadTemplates(t *testing.T) {
	// Create a test embedded source using the actual embedded FS
	templatesFS, err := fs.Sub(assets.Templates, "templates")
	if err != nil {
		t.Fatalf("Failed to create templates filesystem: %v", err)
	}

	source := NewEmbeddedSource(templatesFS)

	templates, err := source.LoadTemplates()
	if err != nil {
		t.Fatalf("Failed to load embedded templates: %v", err)
	}

	// Should have loaded at least the embedded templates
	if len(templates) == 0 {
		t.Error("Expected at least one embedded template, got none")
	}

	// Check that all templates have the correct metadata
	for name, tmpl := range templates {
		if tmpl.IsEmbedded != true {
			t.Errorf("Template %s should be marked as embedded", name)
		}

		if tmpl.Source != sourceEmbedded {
			t.Errorf("Template %s should have source 'embedded', got '%s'", name, tmpl.Source)
		}

		if tmpl.Name == "" {
			t.Errorf("Template %s has empty name", name)
		}

		if tmpl.Content == "" {
			t.Errorf("Template %s has empty content", name)
		}
	}
}

func TestFilesystemSource_LoadTemplates(t *testing.T) {
	// Create a temporary directory for test templates
	tmpDir := t.TempDir()

	// Create test template files
	testTemplates := map[string]string{
		"test1.md": `# Test Template 1
This is a test template.
{TASK}
`,
		"prompt_test2.md": `# Test Template 2
Another test template.
{TASK} and {RULES}
`,
		"invalid.txt": "not a markdown file",
	}

	for filename, content := range testTemplates {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Load templates from filesystem
	source := NewFilesystemSource(tmpDir, "test")
	templates, err := source.LoadTemplates()
	if err != nil {
		t.Fatalf("Failed to load filesystem templates: %v", err)
	}

	// Should have loaded 2 .md files (not the .txt file)
	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(templates))
	}

	// Check test1 template
	if tmpl, ok := templates["test1"]; ok {
		if tmpl.IsEmbedded != false {
			t.Error("Filesystem template should not be marked as embedded")
		}

		if tmpl.Source != "test" {
			t.Errorf("Expected source 'test', got '%s'", tmpl.Source)
		}

		if tmpl.Name != "test1" {
			t.Errorf("Expected name 'test1', got '%s'", tmpl.Name)
		}
	} else {
		t.Error("Expected template 'test1' to be loaded")
	}

	// Check test2 template (should have prompt_ prefix stripped)
	if tmpl, ok := templates["test2"]; ok {
		if tmpl.Name != "test2" {
			t.Errorf("Expected name 'test2' (prefix stripped), got '%s'", tmpl.Name)
		}
	} else {
		t.Error("Expected template 'test2' to be loaded (with prefix stripped)")
	}
}

func TestFilesystemSource_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	source := NewFilesystemSource(tmpDir, "empty")
	templates, err := source.LoadTemplates()
	if err != nil {
		t.Fatalf("Should not error on empty directory: %v", err)
	}

	if len(templates) != 0 {
		t.Errorf("Expected 0 templates from empty directory, got %d", len(templates))
	}
}

func TestFilesystemSource_InvalidDirectory(t *testing.T) {
	source := NewFilesystemSource("/nonexistent/path", "invalid")
	_, err := source.LoadTemplates()
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestFilesystemSource_MalformedTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malformed template (empty content)
	malformedPath := filepath.Join(tmpDir, "malformed.md")
	if err := os.WriteFile(malformedPath, []byte(""), 0o600); err != nil {
		t.Fatalf("Failed to create malformed template: %v", err)
	}

	// Create a valid template
	validPath := filepath.Join(tmpDir, "valid.md")
	validContent := "# Valid Template\n{TASK}\n"
	if err := os.WriteFile(validPath, []byte(validContent), 0o600); err != nil {
		t.Fatalf("Failed to create valid template: %v", err)
	}

	// Load templates - should skip malformed but load valid
	source := NewFilesystemSource(tmpDir, "test")
	templates, err := source.LoadTemplates()
	if err != nil {
		t.Fatalf("Should not error when some templates are malformed: %v", err)
	}

	// Should have loaded only the valid template
	if len(templates) != 1 {
		t.Errorf("Expected 1 valid template, got %d", len(templates))
	}

	if _, ok := templates["valid"]; !ok {
		t.Error("Expected 'valid' template to be loaded")
	}

	if _, ok := templates["malformed"]; ok {
		t.Error("Malformed template should not be loaded")
	}
}

func TestLoadTemplatesFromFS_NameExtraction(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		filename     string
		expectedName string
	}{
		{"simple.md", "simple"},
		{"prompt_prefixed.md", "prefixed"},
		{"prompt_with_underscore.md", "with_underscore"},
		{"no_prefix.md", "no_prefix"},
	}

	for _, tc := range testCases {
		content := "# Test\n{TASK}\n"
		path := filepath.Join(tmpDir, tc.filename)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	source := NewFilesystemSource(tmpDir, "test")
	templates, err := source.LoadTemplates()
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	for _, tc := range testCases {
		if _, ok := templates[tc.expectedName]; !ok {
			t.Errorf("Expected template with name '%s' from file '%s'", tc.expectedName, tc.filename)
		}
	}
}

func TestFilesystemSource_GetSourceName(t *testing.T) {
	source := NewFilesystemSource("/some/path", "custom-source")
	if name := source.GetSourceName(); name != "custom-source" {
		t.Errorf("Expected source name 'custom-source', got '%s'", name)
	}
}

func TestEmbeddedSource_GetSourceName(t *testing.T) {
	templatesFS, err := fs.Sub(assets.Templates, "templates")
	if err != nil {
		t.Fatalf("Failed to create templates filesystem: %v", err)
	}

	source := NewEmbeddedSource(templatesFS)
	if name := source.GetSourceName(); name != "embedded" {
		t.Errorf("Expected source name 'embedded', got '%s'", name)
	}
}
