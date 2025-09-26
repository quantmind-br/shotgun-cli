package template

import (
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil manager")
	}

	if manager.templates == nil {
		t.Fatal("Manager templates map is nil")
	}

	if manager.renderer == nil {
		t.Fatal("Manager renderer is nil")
	}
}

func TestManager_ListTemplates(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	templates, err := manager.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}

	if len(templates) == 0 {
		t.Error("Expected at least one template, got none")
	}

	// Verify each template has required fields
	for _, template := range templates {
		if template.Name == "" {
			t.Error("Template name is empty")
		}
		if template.Content == "" {
			t.Error("Template content is empty")
		}
		if !template.IsEmbedded {
			t.Error("Template should be marked as embedded")
		}
	}
}

func TestManager_GetTemplate(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Get list of available templates first
	templates, err := manager.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() failed: %v", err)
	}

	if len(templates) == 0 {
		t.Skip("No templates available for testing")
	}

	// Test getting an existing template
	templateName := templates[0].Name
	template, err := manager.GetTemplate(templateName)
	if err != nil {
		t.Fatalf("GetTemplate(%s) failed: %v", templateName, err)
	}

	if template == nil {
		t.Fatal("GetTemplate() returned nil template")
	}

	if template.Name != templateName {
		t.Errorf("Expected template name %s, got %s", templateName, template.Name)
	}

	// Test getting a non-existent template
	_, err = manager.GetTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent template, got nil")
	}
}

func TestManager_RenderTemplate(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Create a test template for rendering
	testTemplate := &Template{
		Name:         "test",
		Content:      "Hello {NAME}, today is {CURRENT_DATE}. Task: {TASK}",
		RequiredVars: []string{"NAME", "TASK"},
		IsEmbedded:   false,
	}

	manager.templates["test"] = testTemplate

	tests := []struct {
		name      string
		template  string
		vars      map[string]string
		wantError bool
		contains  []string
	}{
		{
			name:     "valid render",
			template: "test",
			vars: map[string]string{
				"NAME": "World",
				"TASK": "Testing",
			},
			wantError: false,
			contains:  []string{"Hello World", "Task: Testing"},
		},
		{
			name:      "missing required variable",
			template:  "test",
			vars:      map[string]string{"NAME": "World"},
			wantError: true,
		},
		{
			name:      "empty required variable",
			template:  "test",
			vars:      map[string]string{"NAME": "World", "TASK": ""},
			wantError: true,
		},
		{
			name:      "non-existent template",
			template:  "nonexistent",
			vars:      map[string]string{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.RenderTemplate(tt.template, tt.vars)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}

			// Verify CURRENT_DATE is substituted
			if !strings.Contains(result, "today is 20") { // Should contain year starting with 20
				t.Errorf("CURRENT_DATE not properly substituted in: %s", result)
			}
		})
	}
}

func TestManager_ValidateTemplate(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Add test templates
	validTemplate := &Template{
		Name:    "valid",
		Content: "Valid template with {VAR1} and {VAR2}",
	}
	invalidTemplate := &Template{
		Name:    "invalid",
		Content: "Invalid template with {UNCLOSED",
	}

	manager.templates["valid"] = validTemplate
	manager.templates["invalid"] = invalidTemplate

	// Test valid template
	err = manager.ValidateTemplate("valid")
	if err != nil {
		t.Errorf("ValidateTemplate() failed for valid template: %v", err)
	}

	// Test invalid template
	err = manager.ValidateTemplate("invalid")
	if err == nil {
		t.Error("Expected validation error for invalid template, got nil")
	}

	// Test non-existent template
	err = manager.ValidateTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent template, got nil")
	}
}

func TestManager_GetTemplateNames(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	names := manager.GetTemplateNames()
	if len(names) == 0 {
		t.Error("Expected at least one template name, got none")
	}

	// Verify no duplicate names
	nameSet := make(map[string]bool)
	for _, name := range names {
		if nameSet[name] {
			t.Errorf("Duplicate template name found: %s", name)
		}
		nameSet[name] = true
	}
}

func TestManager_HasTemplate(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Get a template name to test with
	names := manager.GetTemplateNames()
	if len(names) == 0 {
		t.Skip("No templates available for testing")
	}

	// Test existing template
	if !manager.HasTemplate(names[0]) {
		t.Errorf("HasTemplate() returned false for existing template: %s", names[0])
	}

	// Test non-existing template
	if manager.HasTemplate("definitely-not-a-template") {
		t.Error("HasTemplate() returned true for non-existing template")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	names := manager.GetTemplateNames()
	if len(names) == 0 {
		t.Skip("No templates available for testing")
	}

	templateName := names[0]

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Test concurrent reads
			_, err := manager.GetTemplate(templateName)
			if err != nil {
				t.Errorf("Concurrent GetTemplate() failed: %v", err)
			}

			_, err = manager.ListTemplates()
			if err != nil {
				t.Errorf("Concurrent ListTemplates() failed: %v", err)
			}

			if !manager.HasTemplate(templateName) {
				t.Errorf("Concurrent HasTemplate() failed for: %s", templateName)
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func BenchmarkManager_GetTemplate(b *testing.B) {
	manager, err := NewManager()
	if err != nil {
		b.Fatalf("NewManager() failed: %v", err)
	}

	names := manager.GetTemplateNames()
	if len(names) == 0 {
		b.Skip("No templates available for benchmarking")
	}

	templateName := names[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetTemplate(templateName)
		if err != nil {
			b.Fatalf("GetTemplate() failed: %v", err)
		}
	}
}

func BenchmarkManager_RenderTemplate(b *testing.B) {
	manager, err := NewManager()
	if err != nil {
		b.Fatalf("NewManager() failed: %v", err)
	}

	// Add a test template for benchmarking
	testTemplate := &Template{
		Name:         "bench",
		Content:      "Benchmark template with {VAR1}, {VAR2}, and {CURRENT_DATE}",
		RequiredVars: []string{"VAR1", "VAR2"},
	}
	manager.templates["bench"] = testTemplate

	vars := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.RenderTemplate("bench", vars)
		if err != nil {
			b.Fatalf("RenderTemplate() failed: %v", err)
		}
	}
}

func TestManager_ValidateAllEmbeddedTemplates(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	templateNames := manager.GetTemplateNames()
	if len(templateNames) == 0 {
		t.Skip("No embedded templates found to validate")
	}

	for _, name := range templateNames {
		t.Run(name, func(t *testing.T) {
			err := manager.ValidateTemplate(name)
			if err != nil {
				t.Errorf("Template %s failed validation: %v", name, err)
			}
		})
	}
}

func TestManager_GetRequiredVariablesForKnownTemplate(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Test with makePlan template if it exists
	if !manager.HasTemplate("makePlan") {
		t.Skip("makePlan template not found, skipping this test")
	}

	vars, err := manager.GetRequiredVariables("makePlan")
	if err != nil {
		t.Fatalf("GetRequiredVariables() failed: %v", err)
	}

	// Verify that the template has expected variables
	expectedVars := []string{"TASK", "RULES", "FILE_STRUCTURE"}
	for _, expectedVar := range expectedVars {
		found := false
		for _, actualVar := range vars {
			if actualVar == expectedVar {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected variable %s not found in required variables: %v", expectedVar, vars)
		}
	}
}

func TestManager_GetRequiredVariables(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Test with non-existent template
	_, err = manager.GetRequiredVariables("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	// Test with existing template
	templateNames := manager.GetTemplateNames()
	if len(templateNames) > 0 {
		vars, err := manager.GetRequiredVariables(templateNames[0])
		if err != nil {
			t.Errorf("GetRequiredVariables() failed for template %s: %v", templateNames[0], err)
		}
		
		// Variables should be a slice (may be empty)
		if vars == nil {
			t.Error("GetRequiredVariables() returned nil instead of empty slice")
		}
	}
}

func TestTemplateDescriptionFallback(t *testing.T) {
	// Test template without header or comment - should use filename as fallback
	content := "This is a template with {VAR1} variable but no header comment."
	fileName := "test_template.md"
	filePath := "test_template.md"

	template, err := parseTemplate(content, fileName, filePath)
	if err != nil {
		t.Fatalf("parseTemplate() failed: %v", err)
	}

	expectedDescription := "Template for test_template"
	if template.Description != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, template.Description)
	}

	// Test with prompt_ prefix - should be removed
	fileName2 := "prompt_makePlan.md"
	template2, err := parseTemplate(content, fileName2, fileName2)
	if err != nil {
		t.Fatalf("parseTemplate() failed: %v", err)
	}

	expectedDescription2 := "Template for makePlan"
	if template2.Description != expectedDescription2 {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription2, template2.Description)
	}
}
