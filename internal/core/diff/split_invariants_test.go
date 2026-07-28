package diff

// Testes de regressão do caçador de bugs: invariantes estruturais do splitter
// verificados contra diffs gerados aleatoriamente, incluindo linhas de
// conteúdo que se parecem com cabeçalho ("+++i;").

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

// StartLine documentado como 1-indexed.
func TestIntelligentSplit_StartLineIsOneIndexed(t *testing.T) {
	lines := []string{
		"diff --git a/file.go b/file.go", "--- a/file.go", "+++ b/file.go",
		"@@ -1 +1 @@", "-a", "+b",
	}
	chunks := IntelligentSplit(lines, DefaultSplitConfig())
	if chunks[0].StartLine != 1 {
		t.Errorf("VIOLADO: chunks[0].StartLine = %d", chunks[0].StartLine)
	}
}

// Nenhuma linha original pode ser perdida nem reordenada: a entrada precisa ser
// subsequência da concatenação dos chunks (os únicos extras são cabeçalhos
// repetidos em continuações).
func TestIntelligentSplit_NoLineLost(t *testing.T) {
	rng := rand.New(rand.NewPCG(2026, 727))
	for iter := range 400 {
		lines := randomDiff(rng)
		approx := 1 + rng.IntN(30)
		chunks := IntelligentSplit(lines, SplitConfig{ApproxLines: approx})

		var got []string
		for _, c := range chunks {
			got = append(got, c.Lines...)
		}
		j := 0
		for _, want := range lines {
			for j < len(got) && got[j] != want {
				j++
			}
			if j == len(got) {
				t.Fatalf("VIOLADO (iter=%d approx=%d): linha %q sumiu\nin=%q\nout=%q",
					iter, approx, want, lines, got)
			}
			j++
		}
	}
}

// Cada chunk precisa ser autoconsistente: contado isoladamente, tem exatamente
// os arquivos que declara — o que só vale se todo chunk começa com cabeçalho.
func TestIntelligentSplit_ChunkSelfConsistent(t *testing.T) {
	rng := rand.New(rand.NewPCG(99, 7))
	for iter := range 400 {
		lines := randomDiff(rng)
		approx := 1 + rng.IntN(30)
		chunks := IntelligentSplit(lines, SplitConfig{ApproxLines: approx})
		for i, c := range chunks {
			if got := CountFiles(c.Lines); got != c.FileCount {
				t.Fatalf("VIOLADO (iter=%d approx=%d): chunk %d declara %d arquivos, tem %d\n%q",
					iter, approx, i, c.FileCount, got, c.Lines)
			}
			if !strings.HasPrefix(c.Lines[0], "diff --git") {
				t.Fatalf("VIOLADO (iter=%d approx=%d): chunk %d começa em %q",
					iter, approx, i, c.Lines[0])
			}
		}
	}
}

func randomDiff(rng *rand.Rand) []string {
	// 4 arquivos × (3 linhas de cabeçalho + 3 hunks × 7 linhas) no pior caso.
	lines := make([]string, 0, 96)
	for f := range 1 + rng.IntN(4) {
		lines = append(lines,
			fmt.Sprintf("diff --git a/f%d.txt b/f%d.txt", f, f),
			fmt.Sprintf("--- a/f%d.txt", f),
			fmt.Sprintf("+++ b/f%d.txt", f),
		)
		for h := range 1 + rng.IntN(3) {
			body := make([]string, 0, 6)
			for range 1 + rng.IntN(6) {
				switch rng.IntN(4) {
				case 0:
					body = append(body, " ctx")
				case 1:
					body = append(body, "-old")
				case 2:
					body = append(body, "+++i;") // conteúdo que parece cabeçalho
				default:
					body = append(body, "+new")
				}
			}
			old, new := 0, 0
			for _, l := range body {
				switch l[0] {
				case ' ':
					old, new = old+1, new+1
				case '-':
					old++
				default:
					new++
				}
			}
			lines = append(lines, fmt.Sprintf("@@ -%d,%d +%d,%d @@", h*10+1, old, h*10+1, new))
			lines = append(lines, body...)
		}
	}
	return lines
}

// TestParseSections_EdgeCases cobre as formas de diff que o parser precisa
// tolerar sem inventar arquivos nem perder linhas.
func TestParseSections_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		lines     []string
		wantFiles int
	}{
		{
			name: "preâmbulo do format-patch antes do primeiro arquivo",
			lines: []string{
				"From 0123456 Mon Sep 17 00:00:00 2001",
				"Subject: [PATCH] muda coisas",
				"",
				"diff --git a/a.txt b/a.txt",
				"--- a/a.txt",
				"+++ b/a.txt",
				"@@ -1 +1 @@",
				"-a",
				"+b",
			},
			wantFiles: 1,
		},
		{
			name:      "hunk sem cabeçalho de arquivo vira seção anônima",
			lines:     []string{"@@ -1 +1 @@", "-a", "+b"},
			wantFiles: 1,
		},
		{
			name:      "cabeçalho @@ malformado é conteúdo, não estrutura",
			lines:     []string{"--- a/a.txt", "+++ b/a.txt", "@@ nao e um hunk @@", "@@ -x,y +z,w @@"},
			wantFiles: 1,
		},
		{
			name: "assinatura de e-mail depois do último hunk",
			lines: []string{
				"diff --git a/a.txt b/a.txt", "--- a/a.txt", "+++ b/a.txt",
				"@@ -1 +1 @@", "-a", "+b",
				"-- ", "2.39.0",
			},
			wantFiles: 1,
		},
		{
			name: "sem contador de linhas o hunk vale 1 linha de cada lado",
			lines: []string{
				"diff --git a/a.txt b/a.txt", "--- a/a.txt", "+++ b/a.txt",
				"@@ -1 +1 @@", "-a", "+b",
				"diff --git a/b.txt b/b.txt", "--- a/b.txt", "+++ b/b.txt",
				"@@ -1 +1 @@", "-c", "+d",
			},
			wantFiles: 2,
		},
		{
			name: "marcador de arquivo sem newline final pertence ao hunk",
			lines: []string{
				"diff --git a/a.txt b/a.txt", "--- a/a.txt", "+++ b/a.txt",
				"@@ -1 +1 @@", "-a", "\\ No newline at end of file", "+b",
			},
			wantFiles: 1,
		},
		{
			name:      "texto que não é diff nenhum",
			lines:     []string{"apenas", "texto", "solto"},
			wantFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountFiles(tt.lines); got != tt.wantFiles {
				t.Errorf("CountFiles = %d, esperado %d", got, tt.wantFiles)
			}

			// Nenhuma linha pode sumir, qualquer que seja o orçamento.
			for _, approx := range []int{1, 3, 500} {
				chunks := IntelligentSplit(tt.lines, SplitConfig{ApproxLines: approx})
				var got []string
				for _, c := range chunks {
					got = append(got, c.Lines...)
				}
				j := 0
				for _, want := range tt.lines {
					for j < len(got) && got[j] != want {
						j++
					}
					if j == len(got) {
						t.Fatalf("approx=%d: linha %q sumiu (saída %q)", approx, want, got)
					}
					j++
				}
			}
		})
	}
}

// TestIntelligentSplit_PreambleRespectsBudget garante que um preâmbulo longo
// (ou um arquivo que não é diff) também é fatiado.
func TestIntelligentSplit_PreambleRespectsBudget(t *testing.T) {
	lines := make([]string, 0, 20)
	for i := range 20 {
		lines = append(lines, "texto "+strconv.Itoa(i))
	}

	chunks := IntelligentSplit(lines, SplitConfig{ApproxLines: 5})
	if len(chunks) != 4 {
		t.Fatalf("esperado 4 chunks de 5 linhas, veio %d", len(chunks))
	}
	if chunks[0].StartLine != 1 || chunks[1].StartLine != 6 {
		t.Errorf("StartLine incorreto: %d, %d", chunks[0].StartLine, chunks[1].StartLine)
	}
	if TotalLines(chunks) != len(lines) {
		t.Errorf("preâmbulo não pode duplicar linhas: %d de %d", TotalLines(chunks), len(lines))
	}
}
