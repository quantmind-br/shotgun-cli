package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTemplateWithVariables(t *testing.T) {
	vars := map[string]string{
		"TASK":           "Implement feature",
		"RULES":          "Follow guidelines",
		"FILE_STRUCTURE": "- main.go",
	}

	out := captureStdout(t, func() {
		if err := renderTemplate("makePlan", vars, ""); err != nil {
			t.Fatalf("renderTemplate error: %v", err)
		}
	})

	if !strings.Contains(out, "Implement feature") {
		t.Fatalf("expected task text in output")
	}
}

func TestRenderTemplateMissingVariable(t *testing.T) {
	if err := renderTemplate("makePlan", map[string]string{}, ""); err == nil {
		t.Fatal("expected error for missing required variables")
	}
}

func TestRenderTemplateWritesFile(t *testing.T) {
	vars := map[string]string{
		"TASK":           "Integration",
		"RULES":          "None",
		"FILE_STRUCTURE": "- README.md",
	}

	tmp := filepath.Join(t.TempDir(), "out.md")
	if err := renderTemplate("makePlan", vars, tmp); err != nil {
		t.Fatalf("renderTemplate error: %v", err)
	}

	data, err := os.ReadFile(tmp) //nolint:gosec // test reading controlled file
	if err != nil || len(data) == 0 {
		t.Fatalf("expected file to be written, err=%v", err)
	}
}

func TestTemplateListCommand(t *testing.T) {
	output := captureStdout(t, func() {
		if err := templateListCmd.RunE(templateListCmd, nil); err != nil {
			t.Fatalf("templateListCmd error: %v", err)
		}
	})

	if !strings.Contains(output, "Available Templates") {
		t.Fatalf("expected header in list output")
	}
}

func TestTemplateRenderCommandErrorsForUnknown(t *testing.T) {
	err := templateRenderCmd.PreRunE(templateRenderCmd, []string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
}
