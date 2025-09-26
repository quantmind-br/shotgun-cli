package template

import (
	"sort"
	"strings"
	"testing"
	"time"
)

var _ TemplateManager = (*Manager)(nil)

func newTestManager(tb testing.TB) *Manager {
	tb.Helper()
	mgr, err := NewManager()
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
