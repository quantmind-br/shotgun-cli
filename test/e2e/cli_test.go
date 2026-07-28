package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestCLIContextGenerateIncludeIgnored trava o contrato da flag
// --include-ignored: sem ela o arquivo ignorado não aparece; com ela, o
// conteúdo dele entra no contexto. Antes da correção a flag era inócua no
// modo headless (a seleção padrão descartava todo nó marcado como ignorado).
func TestCLIContextGenerateIncludeIgnored(t *testing.T) {
	root := repoRoot()
	fixture := t.TempDir()
	outDir := t.TempDir()

	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(fixture, name), []byte(content), 0o600); err != nil {
			t.Fatalf("fixture %s: %v", name, err)
		}
	}
	write(".gitignore", "ignored.txt\n")
	write("ignored.txt", "MARCADOR_IGNORADO\n")
	write("normal.txt", "MARCADOR_NORMAL\n")

	run := func(output string, extra ...string) string {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		args := append([]string{"run", ".", "context", "generate",
			"--root", fixture, "--output", output, "--task", "t", "--rules", "r"}, extra...)
		cmd := exec.CommandContext(ctx, "go", args...) //nolint:gosec // test command with controlled args
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "SHOTGUN_VERBOSE=false")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("context generate failed: %v\n%s", err, out)
		}
		data, err := os.ReadFile(output) //nolint:gosec // path built by the test
		if err != nil {
			t.Fatalf("read output: %v", err)
		}

		return string(data)
	}

	without := run(filepath.Join(outDir, "sem.md"))
	if strings.Contains(without, "MARCADOR_IGNORADO") {
		t.Error("sem --include-ignored o arquivo ignorado não pode entrar no contexto")
	}
	if !strings.Contains(without, "MARCADOR_NORMAL") {
		t.Error("o arquivo normal precisa estar no contexto")
	}

	with := run(filepath.Join(outDir, "com.md"), "--include-ignored")
	if !strings.Contains(with, "MARCADOR_IGNORADO") {
		t.Error("com --include-ignored o conteúdo do arquivo ignorado precisa entrar no contexto")
	}
	if !strings.Contains(with, "ignored.txt") {
		t.Error("com --include-ignored o arquivo ignorado precisa aparecer na árvore")
	}
}

// TestCLIVersionReportsInjectedBuildInfo trava o contrato de build-info: o
// binário de build/ é produzido por `make build`, que injeta os quatro campos
// via -ldflags. Se algum -X deixar de resolver (prefixo de pacote errado, alvo
// sem LDFLAGS), o campo cai no sentinela e este teste falha.
//
// Precisa do binário compilado: `go run .` nunca recebe ldflags.
func TestCLIVersionReportsInjectedBuildInfo(t *testing.T) {
	binaryPath := filepath.Join("..", "..", "build", "shotgun-cli")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, binaryPath, "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("--version falhou: %v\n%s", err, out)
	}

	got := string(out)
	for _, field := range []string{"commit:", "built:", "built by:"} {
		if !strings.Contains(got, field) {
			t.Errorf("--version não reportou %q:\n%s", field, got)
		}
	}
	for _, sentinel := range []string{"version dev", "unknown"} {
		if strings.Contains(got, sentinel) {
			t.Errorf("--version reportou o sentinela %q — os -ldflags não chegaram ao binário:\n%s",
				sentinel, got)
		}
	}
}
