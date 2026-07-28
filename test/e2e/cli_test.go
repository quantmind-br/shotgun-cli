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

// TestCLILLMDoctorExitCode trava o contrato de exit code do `llm doctor`: com
// provider configurado ele sai 0, sem API key sai non-zero. Antes ele sempre
// saía 0, então não servia para gatear script nenhum.
//
// A configuração é isolada num --config temporário e o ambiente é limpo de
// SHOTGUN_*: sem isso, um ~/.config/shotgun-cli/config.yaml da máquina pode
// fornecer uma API key e o caso de falha deixa de falhar.
func TestCLILLMDoctorExitCode(t *testing.T) {
	root := repoRoot()

	// Env sem nenhuma variável SHOTGUN_*, que o viper leria via AutomaticEnv.
	cleanEnv := make([]string, 0, len(os.Environ()))
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "SHOTGUN_") {
			cleanEnv = append(cleanEnv, kv)
		}
	}

	tests := []struct {
		name       string
		configYAML string
		wantErr    bool
	}{
		{
			name:       "sem api key sai non-zero",
			configYAML: "llm:\n  provider: openai\n  model: gpt-4o\n",
			wantErr:    true,
		},
		{
			name:       "provider completo sai zero",
			configYAML: "llm:\n  provider: openai\n  api-key: sk-test-key\n  model: gpt-4o\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgPath := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(cfgPath, []byte(tt.configYAML), 0o600); err != nil {
				t.Fatalf("não foi possível escrever o config: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, //nolint:gosec // test command with controlled args
				"go", "run", ".", "llm", "doctor", "--config", cfgPath,
			)
			cmd.Dir = root
			cmd.Env = cleanEnv

			out, err := cmd.CombinedOutput()
			if tt.wantErr && err == nil {
				t.Fatalf("esperava exit non-zero, obteve 0:\n%s", out)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("esperava exit 0, obteve %v:\n%s", err, out)
			}

			// A lista de issues aparece uma única vez: o erro devolvido é curto e
			// não a repete.
			if got := strings.Count(string(out), "API key not configured"); tt.wantErr && got != 1 {
				t.Errorf("esperava a issue de API key exatamente 1 vez, apareceu %d:\n%s", got, out)
			}
			// O usage não deve ser despejado por cima das instruções de correção.
			if tt.wantErr && strings.Contains(string(out), "Usage:") {
				t.Errorf("doctor despejou o usage no caminho de falha:\n%s", out)
			}
		})
	}
}
