package template

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/assets"
)

var _ TemplateManager = (*Manager)(nil)

func newTestManager(tb testing.TB) *Manager {
	tb.Helper()
	mgr, err := NewManager(ManagerConfig{})
	if err != nil {
		tb.Fatalf("NewManager failed: %v", err)
	}

	return mgr
}

func TestManager_ListTemplates_IsDeterministic(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	templates, err := mgr.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}
	if len(templates) == 0 {
		t.Fatalf("expected embedded templates to be present")
	}

	names := make([]string, len(templates))
	for i, tmpl := range templates {
		if tmpl.Name == "" {
			t.Fatalf("template %d missing name", i)
		}
		if tmpl.Content == "" {
			t.Fatalf("template %s missing content", tmpl.Name)
		}
		names[i] = tmpl.Name
	}

	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	if !equalSlices(names, sorted) {
		t.Fatalf("template listing should be sorted: got %v want %v", names, sorted)
	}
}

func equalSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestManager_GetTemplate_Scenarios(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)

	custom := &Template{Name: "custom", Content: "Custom", RequiredVars: []string{"TASK"}}
	mgr.templates[custom.Name] = custom

	cases := []struct {
		name      string
		template  string
		wantError bool
	}{
		{name: "existing", template: custom.Name},
		{name: "missing", template: "not-found", wantError: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmpl, err := mgr.GetTemplate(tc.template)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error for template %s", tc.template)
				}

				return
			}
			if err != nil {
				t.Fatalf("GetTemplate failed: %v", err)
			}
			if tmpl.Name != tc.template {
				t.Fatalf("unexpected template: %s", tmpl.Name)
			}
		})
	}
}

func TestManager_RenderTemplate_Table(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	mgr.templates["greeting"] = &Template{
		Name:         "greeting",
		Content:      "Hello {TASK}! Today is {CURRENT_DATE}.",
		RequiredVars: []string{"TASK", "CURRENT_DATE"},
	}

	cases := []struct {
		name      string
		vars      map[string]string
		expect    []string
		wantError bool
	}{
		{
			name:   "valid substitution",
			vars:   map[string]string{"TASK": "world"},
			expect: []string{"Hello world!", time.Now().Format("2006-01-02")[:4]},
		},
		{
			name:      "missing variable",
			vars:      map[string]string{},
			wantError: true,
		},
		{
			name:      "empty trimmed variable",
			vars:      map[string]string{"TASK": "   "},
			wantError: true,
		},
		{
			name:   "special characters",
			vars:   map[string]string{"TASK": "release v1.0 ✔"},
			expect: []string{"release v1.0 ✔"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := mgr.RenderTemplate("greeting", tc.vars)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error for %+v", tc.vars)
				}

				return
			}
			if err != nil {
				t.Fatalf("RenderTemplate failed: %v", err)
			}
			for _, snippet := range tc.expect {
				if !strings.Contains(out, snippet) {
					t.Fatalf("expected output to contain %q, got %q", snippet, out)
				}
			}
		})
	}
}

func TestManager_ValidateTemplate_Errors(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	mgr.templates["broken"] = &Template{Name: "broken", Content: "{UNCLOSED"}

	if err := mgr.ValidateTemplate("broken"); err == nil {
		t.Fatalf("expected validation error for malformed template")
	}

	if err := mgr.ValidateTemplate("not-present"); err == nil {
		t.Fatalf("expected error for missing template")
	}
}

func TestManager_GetRequiredVariables_FiltersAutoVars(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	mgr.templates["vars"] = &Template{
		Name:         "vars",
		Content:      "{TASK} {CURRENT_DATE} {RULES}",
		RequiredVars: []string{"TASK", "CURRENT_DATE", "RULES"},
	}

	vars, err := mgr.GetRequiredVariables("vars")
	if err != nil {
		t.Fatalf("GetRequiredVariables failed: %v", err)
	}

	if len(vars) != 2 {
		t.Fatalf("expected auto-generated vars to be filtered, got %v", vars)
	}
	sort.Strings(vars)
	if !equalSlices(vars, []string{"RULES", "TASK"}) {
		t.Fatalf("unexpected vars: %v", vars)
	}
}

func TestManager_RenderTemplate_AllowsCustomTemplate(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	mgr.templates["custom"] = &Template{
		Name:         "custom",
		Content:      "Ticket: {TASK}\nRules: {RULES}",
		RequiredVars: []string{"TASK", "RULES"},
	}

	out, err := mgr.RenderTemplate("custom", map[string]string{"TASK": "T123", "RULES": "none"})
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}
	if !strings.Contains(out, "Ticket: T123") || !strings.Contains(out, "Rules: none") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestManager_ListTemplates_LoadsEmbedded(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	names := mgr.GetTemplateNames()
	if len(names) == 0 {
		t.Fatalf("expected embedded templates to be available")
	}

	if !mgr.HasTemplate("makePlan") {
		t.Log("makePlan template not present; continuing")
	}
}

func BenchmarkManager_RenderTemplate(b *testing.B) {
	mgr := newTestManager(b)
	mgr.templates["bench"] = &Template{
		Name:         "bench",
		Content:      "Result: {TASK} {RULES} {CURRENT_DATE}",
		RequiredVars: []string{"TASK", "RULES", "CURRENT_DATE"},
	}

	vars := map[string]string{"TASK": "bench", "RULES": "all"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mgr.RenderTemplate("bench", vars); err != nil {
			b.Fatalf("RenderTemplate failed: %v", err)
		}
	}
}

func TestManager_MultiSource_EmbeddedTemplatesLoad(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	templates, err := mgr.ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	// Should have loaded embedded templates
	if len(templates) == 0 {
		t.Fatal("Expected embedded templates to be loaded")
	}

	// Check that embedded templates have correct source metadata
	foundEmbedded := false
	for _, tmpl := range templates {
		if tmpl.IsEmbedded {
			foundEmbedded = true
			if tmpl.Source != "embedded" {
				t.Errorf("Embedded template %s should have source 'embedded', got '%s'", tmpl.Name, tmpl.Source)
			}
		}
	}

	if !foundEmbedded {
		t.Error("Expected at least one embedded template")
	}
}

func TestManager_MultiSource_PriorityOverride(t *testing.T) {
	t.Parallel()

	// Test that later sources override earlier ones
	mgr := &Manager{
		templates: make(map[string]*Template),
		renderer:  NewRenderer(),
	}

	// Create mock sources
	source1 := &mockTemplateSource{
		templates: map[string]*Template{
			"test":  {Name: "test", Content: "Source 1", Source: "source1", IsEmbedded: false},
			"only1": {Name: "only1", Content: "Only in 1", Source: "source1", IsEmbedded: false},
		},
	}

	source2 := &mockTemplateSource{
		templates: map[string]*Template{
			"test":  {Name: "test", Content: "Source 2", Source: "source2", IsEmbedded: false},
			"only2": {Name: "only2", Content: "Only in 2", Source: "source2", IsEmbedded: false},
		},
	}

	// Load from sources (source2 should override source1)
	sources := []TemplateSource{source1, source2}
	if err := mgr.loadFromSources(sources); err != nil {
		t.Fatalf("loadFromSources failed: %v", err)
	}

	// Check that "test" template is from source2 (later source)
	testTmpl, err := mgr.GetTemplate("test")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if testTmpl.Content != "Source 2" {
		t.Errorf("Expected template from source2, got content: %s", testTmpl.Content)
	}

	if testTmpl.Source != "source2" {
		t.Errorf("Expected source 'source2', got '%s'", testTmpl.Source)
	}

	// Check that both unique templates are present
	if !mgr.HasTemplate("only1") {
		t.Error("Expected 'only1' template from source1")
	}

	if !mgr.HasTemplate("only2") {
		t.Error("Expected 'only2' template from source2")
	}
}

func TestManager_MultiSource_GracefulDegradation(t *testing.T) {
	t.Parallel()

	mgr := &Manager{
		templates: make(map[string]*Template),
		renderer:  NewRenderer(),
	}

	// Create sources with one failing
	goodSource := &mockTemplateSource{
		templates: map[string]*Template{
			"good": {Name: "good", Content: "Good", Source: "good", IsEmbedded: false},
		},
	}

	failingSource := &mockTemplateSource{
		shouldError: true,
	}

	// Load from sources (should continue despite failure)
	sources := []TemplateSource{goodSource, failingSource}
	if err := mgr.loadFromSources(sources); err != nil {
		t.Fatalf("loadFromSources should not fail if individual sources fail: %v", err)
	}

	// Should have loaded templates from good source
	if !mgr.HasTemplate("good") {
		t.Error("Expected templates from good source to be loaded")
	}
}

func TestManager_MultiSource_SourceMetadata(t *testing.T) {
	t.Parallel()

	mgr := &Manager{
		templates: make(map[string]*Template),
		renderer:  NewRenderer(),
	}

	source := &mockTemplateSource{
		templates: map[string]*Template{
			"test": {
				Name:       "test",
				Content:    "Test content",
				Source:     "custom-source",
				IsEmbedded: false,
			},
		},
	}

	if err := mgr.loadFromSources([]TemplateSource{source}); err != nil {
		t.Fatalf("loadFromSources failed: %v", err)
	}

	tmpl, err := mgr.GetTemplate("test")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if tmpl.IsEmbedded {
		t.Error("Template should not be marked as embedded")
	}

	if tmpl.Source != "custom-source" {
		t.Errorf("Expected source 'custom-source', got '%s'", tmpl.Source)
	}
}

// mockTemplateSource is a mock implementation for testing
type mockTemplateSource struct {
	templates   map[string]*Template
	shouldError bool
}

func (m *mockTemplateSource) LoadTemplates() (map[string]*Template, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}

	return m.templates, nil
}

func (m *mockTemplateSource) GetSourceName() string {
	return "mock"
}

func TestCustomTemplateWorkflow(t *testing.T) {
	// Integration test for full custom template workflow
	t.Parallel()

	// Create a temporary config directory
	tmpDir := t.TempDir()
	userTemplatesDir := filepath.Join(tmpDir, "templates")

	// Create user templates directory
	if err := os.MkdirAll(userTemplatesDir, 0o750); err != nil {
		t.Fatalf("Failed to create user templates directory: %v", err)
	}

	// Place a custom template file
	customTemplateContent := `# Custom Test Template
This is a custom template for testing.
Task: {TASK}
Rules: {RULES}
`
	customTemplatePath := filepath.Join(userTemplatesDir, "custom_test.md")
	if err := os.WriteFile(customTemplatePath, []byte(customTemplateContent), 0o600); err != nil {
		t.Fatalf("Failed to write custom template: %v", err)
	}

	// Create an embedded template with the same name to test override
	embeddedFS, err := fs.Sub(assets.Templates, "templates")
	if err != nil {
		t.Fatalf("Failed to create embedded FS: %v", err)
	}

	// Initialize manager with both sources
	mgr := &Manager{
		templates: make(map[string]*Template),
		renderer:  NewRenderer(),
	}

	sources := []TemplateSource{
		NewEmbeddedSource(embeddedFS),
		NewFilesystemSource(userTemplatesDir, "user"),
	}

	if err := mgr.loadFromSources(sources); err != nil {
		t.Fatalf("Failed to load from sources: %v", err)
	}

	// Verify custom template loads
	customTmpl, err := mgr.GetTemplate("custom_test")
	if err != nil {
		t.Fatalf("Custom template should be loaded: %v", err)
	}

	if customTmpl.Source != "user" {
		t.Errorf("Expected source 'user', got '%s'", customTmpl.Source)
	}

	if customTmpl.IsEmbedded {
		t.Error("Custom template should not be marked as embedded")
	}

	// Verify template can be rendered
	rendered, err := mgr.RenderTemplate("custom_test", map[string]string{
		"TASK":  "test task",
		"RULES": "test rules",
	})
	if err != nil {
		t.Fatalf("Failed to render custom template: %v", err)
	}

	if !strings.Contains(rendered, "test task") {
		t.Errorf("Rendered template should contain 'test task', got: %s", rendered)
	}

	if !strings.Contains(rendered, "test rules") {
		t.Errorf("Rendered template should contain 'test rules', got: %s", rendered)
	}

	// Verify embedded templates still work
	if !mgr.HasTemplate("analyzeBug") {
		t.Error("Embedded templates should still be available")
	}
}
