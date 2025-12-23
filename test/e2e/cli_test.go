package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func repoRoot() string {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		panic(err)
	}
	return root
}

func TestCLIContextGenerateProducesFile(t *testing.T) {
	root := repoRoot()
	fixture := filepath.Join(root, "test", "fixtures", "sample-project")
	output := filepath.Join(t.TempDir(), "context-output.md")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, //nolint:gosec // test command with controlled args
		"go", "run", ".", "context", "generate",
		"--root", fixture, "--output", output, "--max-size", "5MB",
	)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "SHOTGUN_VERBOSE=false")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("context generate command failed: %v\n%s", err, out)
	}

	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file, got error: %v", err)
	}
}

func TestCLITemplateRenderCreatesFile(t *testing.T) {
	root := repoRoot()
	output := filepath.Join(t.TempDir(), "template.md")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, //nolint:gosec // test command with controlled args
		"go", "run", ".", "template", "render", "makePlan",
		"--var", "TASK=Document fixture",
		"--var", "RULES=Keep it short",
		"--var", "FILE_STRUCTURE=- main.go",
		"--output", output,
	)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "SHOTGUN_VERBOSE=false")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("template render failed: %v\n%s", err, out)
	}

	data, err := os.ReadFile(output) //nolint:gosec // test reading controlled file
	if err != nil || len(data) == 0 {
		t.Fatalf("expected rendered template file, err=%v", err)
	}
}
