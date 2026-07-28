package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// git roda comandos no diretório dado e falha o teste em caso de erro.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...) //nolint:gosec // argumentos controlados pelo teste
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}

	return string(out)
}

// TestIntelligentSplit_ChunksApplyWithGit é o oráculo forte do splitter: cada
// chunk gerado precisa ser um patch que o git aceita, e aplicá-los em ordem
// tem de reproduzir exatamente a modificação original. É esse contrato que a
// flag --no-header promete ("for patch tool compatibility").
func TestIntelligentSplit_ChunksApplyWithGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git indisponível")
	}

	repo := t.TempDir()
	git(t, repo, "init", "-q")

	// Arquivo com hunks bem separados e conteúdo que parece cabeçalho de diff.
	original := make([]string, 0, 48)
	modified := make([]string, 0, 48)
	for i := range 12 {
		original = append(original, "linha "+strconv.Itoa(i), "estavel a", "estavel b", "estavel c")
		if i%2 == 0 {
			modified = append(modified, "++i;", "estavel a", "estavel b", "estavel c")
		} else {
			modified = append(modified, "linha "+strconv.Itoa(i), "estavel a", "estavel b", "estavel c")
		}
	}

	file := filepath.Join(repo, "src.txt")
	if err := os.WriteFile(file, []byte(strings.Join(original, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "add", "src.txt")
	git(t, repo, "commit", "-qm", "base")

	want := strings.Join(modified, "\n") + "\n"
	if err := os.WriteFile(file, []byte(want), 0o600); err != nil {
		t.Fatal(err)
	}

	raw := git(t, repo, "diff", "--unified=1")
	lines := strings.Split(strings.TrimSuffix(raw, "\n"), "\n")

	chunks := IntelligentSplit(lines, SplitConfig{ApproxLines: 12})
	if len(chunks) < 2 {
		t.Fatalf("esperado mais de um chunk para exercitar o oráculo, veio %d", len(chunks))
	}

	// Volta ao estado base e aplica os chunks em ordem.
	git(t, repo, "checkout", "--", "src.txt")

	for i, c := range chunks {
		patch := filepath.Join(t.TempDir(), "chunk.patch")
		if err := os.WriteFile(patch, []byte(strings.Join(c.Lines, "\n")+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "apply", "--check", patch) //nolint:gosec // caminho do teste
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("chunk %d não é um patch válido: %v\n%s\n--- chunk ---\n%s",
				i, err, out, strings.Join(c.Lines, "\n"))
		}
		git(t, repo, "apply", patch)
	}

	got, err := os.ReadFile(file) //nolint:gosec // caminho do teste
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("aplicar os chunks não reproduziu a modificação original\n--- obtido ---\n%s\n--- esperado ---\n%s", got, want)
	}
}

// TestIntelligentSplit_SingleHunkExceedsBudget documenta o limite conhecido:
// um hunk isolado maior que ApproxLines não é partido, porque fatiar um hunk
// exigiria recalcular os contadores do cabeçalho "@@" — e um chunk com
// contadores errados seria rejeitado por qualquer ferramenta de patch.
func TestIntelligentSplit_SingleHunkExceedsBudget(t *testing.T) {
	lines := make([]string, 0, 54)
	lines = append(lines,
		"diff --git a/big.txt b/big.txt",
		"--- a/big.txt",
		"+++ b/big.txt",
		"@@ -1,50 +1,50 @@",
	)
	for range 50 {
		lines = append(lines, " ctx")
	}

	chunks := IntelligentSplit(lines, SplitConfig{ApproxLines: 10})
	if len(chunks) != 1 {
		t.Fatalf("um hunk único não pode ser partido; veio %d chunks", len(chunks))
	}
	if len(chunks[0].Lines) != len(lines) {
		t.Errorf("chunk devia conter o hunk inteiro: %d de %d linhas", len(chunks[0].Lines), len(lines))
	}
}
