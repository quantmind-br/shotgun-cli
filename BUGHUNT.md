# Caça a defeitos — shotgun-cli

Execução autônoma do ciclo gerar → executar → oráculo → minimizar → diagnosticar →
corrigir → revalidar, dentro de sandbox `bwrap` sem rede. A caça inteira — geração de
entrada hostil, execução e correção — aconteceu no clone confinado. A árvore real só foi
tocada no fim, para portar as correções e revalidar o entregável no lugar onde ele vai
viver; foi essa revalidação que expôs o BUG-11, e a correção dele (um arquivo `_test.go`)
é a única feita fora do sandbox, que já estava desmontado.

- **HEAD:** `63e8bf7`
- **Estado caçado:** commit + alterações não commitadas (`BUGHUNT_DIRTY=aplicar`, 154
  arquivos). Os dois arquivos Go não rastreados (`internal/core/selection/store.go` e
  `store_test.go`) foram copiados para dentro do sandbox — sem eles o projeto nem compila.
- **Toolchain:** go1.26.5 linux/amd64, golangci-lint 2.12.2, git.
- **Baseline (fase 2):** build OK, `go vet` limpo, suíte verde em duas execuções
  consecutivas, `-race` limpo, cobertura total **82,4%**. `gofmt -l` já apontava 4
  arquivos antes da caça (ver achados de baseline).

## Sumário

|#|Severidade|Módulo|Título|
|---|---|---|---|
|BUG-01|alta|`internal/core/diff`|Chunk cortado antes de `@@` sai sem cabeçalho e não aplica|
|BUG-02|alta|`internal/core/diff`|Linha de conteúdo `+++i;` classificada como cabeçalho de arquivo|
|BUG-03|média|`internal/core/diff`|`CountFiles` conta cabeçalhos, não arquivos (dobro em diff git)|
|BUG-04|baixa|`internal/core/diff`|`StartLine` do primeiro chunk é 0, contrariando o contrato 1-indexed|
|BUG-05|alta|`cmd`|`writeDiffChunk` engole erro de escrita e reporta sucesso com chunk truncado|
|BUG-06|alta|`internal/app` + `contextgen` + `scanner`|`--include-ignored` era inócuo no modo headless|
|BUG-07|alta|`internal/core/ignore`|`.gitignore` aninhado perde a recursividade e deixa vazar arquivo ignorado|
|BUG-08|média|`internal/core/ignore`|`dir/` de ignore aninhado vira `dir` e passa a excluir arquivo homônimo|
|BUG-09|média|`internal/core/ignore`|Um diretório ilegível aborta a carga de ignore do projeto inteiro|
|BUG-10|média|`internal/core/selection`|Gravações concorrentes perdem seleções e falham no `rename`|
|BUG-11|média|`internal/ui/styles`|Testes de acessibilidade quebram (ou viram vácuo) conforme `NO_COLOR`|

Total: **11 confirmados, corrigidos e protegidos por teste de regressão**.

---

### BUG-01 — Chunk cortado antes de `@@` sai sem cabeçalho e não aplica

- **Severidade:** alta
- **Módulo:** `internal/core/diff`
- **Invariante violado:** I1 (todo chunk emitido é um patch aplicável — o comando promete
  "each chunk maintains proper diff headers" e `--no-header` é anunciado como
  "for patch tool compatibility")
- **Oráculo que detectou:** I1, confirmado por O3 (`git apply --check`)
- **Iteração:** 4

**Reprodução mínima**

```go
lines := []string{
	"diff --git a/big.go b/big.go", "--- a/big.go", "+++ b/big.go",
	"@@ -1,3 +1,4 @@", " a", "-b", "+c", "+d",
	"@@ -20,3 +21,4 @@", " e", "-f", "+g", "+h",
}
IntelligentSplit(lines, SplitConfig{ApproxLines: 8})
```

**Comportamento observado:** o segundo chunk começa em `@@ -20,3 +21,4 @@`, sem
`diff --git`/`---`/`+++`. `git apply --check` rejeita com `exit status 128`.
**Comportamento esperado:** todo chunk começa por um cabeçalho de arquivo e é aplicável
isoladamente, na ordem em que foi emitido.
**Causa raiz:** `internal/core/diff/split.go:144` (código antigo) — `CanSplitAt` declarava
seguro qualquer ponto imediatamente antes de um `@@`, sem reemitir o cabeçalho do arquivo
que continua no próximo chunk.
**Correção:** `IntelligentSplit` foi reescrito sobre um parser que conhece a estrutura do
diff (seções de arquivo + hunks). O corte acontece entre arquivos; quando um arquivo
sozinho estoura `ApproxLines`, ele é cortado entre hunks e **o cabeçalho é reemitido** no
chunk de continuação. `CanSplitAt`, que era puramente léxico e ficou sem uso, foi removido.
**Teste de regressão:** `TestIntelligentSplit_ChunksApplyWithGit` em
`internal/core/diff/apply_test.go` — cria um repositório git de verdade, fatia o diff real
e exige `git apply --check` em cada chunk, além de conferir que aplicar todos em ordem
reproduz exatamente a modificação original.
**Evidência:**

```
=== COM O CÓDIGO ORIGINAL ===
--- FAIL: TestIntelligentSplit_ChunksApplyWithGit (0.01s)
    apply_test.go:84: chunk 1 não é um patch válido: exit status 128
FAIL	github.com/quantmind-br/shotgun-cli/internal/core/diff	0.007s
=== COM A CORREÇÃO ===
ok  	github.com/quantmind-br/shotgun-cli/internal/core/diff	0.010s
```

---

### BUG-02 — Linha de conteúdo `+++i;` classificada como cabeçalho de arquivo

- **Severidade:** alta
- **Módulo:** `internal/core/diff`
- **Invariante violado:** I2 (dentro de um hunk, toda linha é conteúdo; a estrutura do diff
  é definida pelos contadores do cabeçalho `@@`, não por prefixo de texto)
- **Oráculo que detectou:** I2
- **Iteração:** 3

**Reprodução mínima**

```go
lines := []string{
	"diff --git a/counter.c b/counter.c", "--- a/counter.c", "+++ b/counter.c",
	"@@ -1,4 +1,6 @@", "int main(void){", "int i=0;",
	"+++i;",        // é a ADIÇÃO da linha "++i;"
	"return i;", "}",
}
IntelligentSplit(lines, SplitConfig{ApproxLines: 6})
```

**Comportamento observado:** `IsDiffHeader("+++i;") == true`. O diff de um único arquivo era
contabilizado como 3 arquivos e o hunk era partido em `+++i;`, gerando um chunk
inaplicável. O mesmo vale para `----` (remoção da linha `---`, comum em front matter YAML e
régua Markdown) e `+++++`.
**Comportamento esperado:** `++i;` é código C/C++/Java corriqueiro; a linha é conteúdo do
hunk e não pode virar fronteira de arquivo.
**Causa raiz:** `internal/core/diff/split.go:112` (código antigo) — `IsDiffHeader` testava
apenas `strings.HasPrefix(line, "---")`/`"+++"`, sem exigir o espaço que separa o marcador
do caminho e sem qualquer noção de estar dentro de um hunk.
**Correção:** dois níveis. `IsDiffHeader` passou a exigir `"--- "`/`"+++ "` (mata `+++i;`,
`----`, `+++++`); e o parser novo debita cada linha do orçamento declarado pelo cabeçalho
`@@ -a,b +c,d @@`, de modo que enquanto o hunk tem linhas a consumir nenhuma linha é
tratada como estrutura — o que resolve inclusive os casos que continuam ambíguos no exame
léxico isolado (ver Suspeitas).
**Teste de regressão:** `TestIntelligentSplit_ChunkSelfConsistent` e
`TestIntelligentSplit_NoLineLost` em `internal/core/diff/split_invariants_test.go` (diffs
aleatórios que injetam `+++i;` como conteúdo), mais os casos `{"+++i;", false}` e
`{"----", false}` em `TestIsDiffHeader`.
**Evidência:**

```
ANTES: P4 VIOLADO: IsDiffHeader("+++i;") = true, mas é conteúdo (linha adicionada '++i;')
       chunk 1 lines=["+++i;" "return i;" "}"]   <- hunk partido no meio
ANTES: --- FAIL: TestIntelligentSplit_ChunkSelfConsistent
           chunk 0 declara 8 arquivos, tem 4
DEPOIS: ok  github.com/quantmind-br/shotgun-cli/internal/core/diff  0.014s
```

---

### BUG-03 — `CountFiles` conta cabeçalhos, não arquivos

- **Severidade:** média
- **Módulo:** `internal/core/diff`
- **Invariante violado:** I3 (o número de arquivos informado ao usuário é o número de
  arquivos do diff)
- **Oráculo que detectou:** I3
- **Iteração:** 2

**Reprodução mínima**

```go
lines := []string{
	"diff --git a/a.go b/a.go", "--- a/a.go", "+++ b/a.go", "@@ -1 +1 @@", "-x", "+y",
	"diff --git a/b.go b/b.go", "--- a/b.go", "+++ b/b.go", "@@ -1 +1 @@", "-x", "+y",
}
CountFiles(lines) // 4
```

**Comportamento observado:** 4 para um diff de 2 arquivos. Num diff git todo arquivo tem
`diff --git` **e** `--- a/...`, e ambos eram somados. O número dobrado chegava ao usuário em
`cmd/diff.go:137` (`📄 Chunk %d: ... (%d files)`) e no cabeçalho de cada chunk gravado.
**Comportamento esperado:** 2.
**Causa raiz:** `internal/core/diff/split.go:157` (código antigo) — a condição
`IsGitDiffHeader(line) || (IsDiffHeader(line) && HasPrefix(line, "---"))` conta as duas
linhas do mesmo arquivo.
**Correção:** `CountFiles` passou a delegar ao parser (`len(files)` das seções detectadas).
**Testes existentes que codificavam o defeito:** `TestCountFiles` em
`internal/core/diff/split_test.go` (casos "counts both headers" → esperava 2, e "counts all
git and --- headers" → esperava 6 para 3 arquivos) e `TestCountFiles` em `cmd/diff_test.go`
(esperava 2 e 4). Os nomes dos casos assumiam explicitamente a contagem de cabeçalhos; como
a função se chama `CountFiles`, documenta "counts the number of files" e alimenta uma saída
rotulada "files", as expectativas foram corrigidas para 1, 3, 1 e 2.
**Teste de regressão:** `TestCountFiles` (ambos os arquivos, expectativas corrigidas) e
`TestProbe_CountFilesReportsFiles` incorporado a `TestIntelligentSplit_ChunkSelfConsistent`.
**Evidência:**

```
ANTES: --- FAIL: TestCountFiles/single_file_with_git_header_counts_one_file
           expected: 1   actual: 2
       --- FAIL: TestCountFiles/multiple_files_counts_each_file_once
           expected: 3   actual: 6
DEPOIS: ok  github.com/quantmind-br/shotgun-cli/internal/core/diff  0.014s
```

---

### BUG-04 — `StartLine` do primeiro chunk é 0

- **Severidade:** baixa
- **Módulo:** `internal/core/diff`
- **Invariante violado:** I4 (`StartLine is the original line number where this chunk
  starts (1-indexed)`, `split.go:16`)
- **Oráculo que detectou:** I4
- **Iteração:** 1

**Reprodução mínima**

```go
chunks := IntelligentSplit([]string{
	"diff --git a/file.go b/file.go", "--- a/file.go", "+++ b/file.go",
	"@@ -1 +1 @@", "-a", "+b",
}, DefaultSplitConfig())
chunks[0].StartLine // 0
```

**Comportamento observado:** 0 para todo diff não vazio — o caso `len(lines) == 0` e o
fallback definiam `StartLine: 1` explicitamente, mas o caminho normal não.
**Comportamento esperado:** 1. Os chunks seguintes já vinham corretos (`i + 2`), então o
campo era inconsistente entre o primeiro e os demais.
**Causa raiz:** `internal/core/diff/split.go:57` (código antigo) — `var currentChunk Chunk`
deixava o campo no zero-value e nada o inicializava antes do primeiro `append`.
**Correção:** o `chunkBuilder` grava `StartLine` com a posição original da primeira linha
de cada chunk, inclusive do primeiro; em chunk de continuação aponta para a primeira linha
original (não para o cabeçalho reemitido).
**Teste de regressão:** `TestIntelligentSplit_StartLineIsOneIndexed` e a checagem de offsets
em `TestIntelligentSplit_PreambleRespectsBudget`, em
`internal/core/diff/split_invariants_test.go`.
**Evidência:**

```
ANTES: --- FAIL: TestIntelligentSplit_StartLineIsOneIndexed
           split_invariants_test.go:23: VIOLADO: chunks[0].StartLine = 0
DEPOIS: ok  github.com/quantmind-br/shotgun-cli/internal/core/diff  0.014s
```

---

### BUG-05 — `writeDiffChunk` engole erro de escrita

- **Severidade:** alta
- **Módulo:** `cmd`
- **Invariante violado:** I5 (falha de E/S vira erro; o comando nunca anuncia sucesso com
  arquivo truncado)
- **Oráculo que detectou:** O5 (erro engolido)
- **Iteração:** 5

**Reprodução mínima**

```go
// /dev/full aceita open e devolve ENOSPC em toda escrita
writeDiffChunk("/dev/full", chunk, 1, 1, false) // → nil
```

**Comportamento observado:** retorna `nil`. Com o disco cheio, `shotgun-cli diff split`
imprime `✅ Diff file split successfully!` e deixa chunks truncados — perda silenciosa de
partes do patch.
**Comportamento esperado:** erro propagado, comando falha.
**Causa raiz:** `cmd/diff.go:152-162` (código antigo) — todos os `fmt.Fprintf`/`Fprintln`
tinham o retorno descartado com `_, _ =` e o `Close` ia para um `defer` que também
descartava o erro.
**Correção:** escrita passa por `bufio.Writer`; o erro é colhido no `Flush` e o `Close`
explícito também é verificado (nesta ordem, para não mascarar o erro de escrita).
**Teste de regressão:** `TestWriteDiffChunk_ReportsWriteError` em `cmd/diff_test.go`
(pula automaticamente se `/dev/full` não existir).
**Evidência:**

```
ANTES: --- FAIL: TestWriteDiffChunk_ReportsWriteError
           Error: An error is expected but got nil.
           Messages: erro de escrita não pode ser engolido
DEPOIS: writeDiffChunk(/dev/full) err = failed to write chunk file /dev/full:
        write /dev/full: no space left on device
        ok  github.com/quantmind-br/shotgun-cli/cmd  0.013s
```

---

### BUG-06 — `--include-ignored` era inócuo no modo headless

- **Severidade:** alta
- **Módulo:** `internal/app`, `internal/core/contextgen`, `internal/core/scanner`
- **Invariante violado:** I6 (uma flag documentada — "Include ignored files" — muda o
  resultado)
- **Oráculo que detectou:** M4 (metamórfico: mudar a flag tem de mudar a saída)
- **Iteração:** 6

**Reprodução mínima**

```bash
mkdir -p /tmp/p && cd /tmp/p
printf 'ignored.txt\n' > .gitignore
printf 'MARCADOR\n' > ignored.txt
printf 'normal\n' > normal.txt
shotgun-cli context generate --root /tmp/p --output /tmp/out.md --task t --rules r --include-ignored
grep -c MARCADOR /tmp/out.md   # 0
```

**Comportamento observado:** saída byte a byte idêntica com e sem a flag; o arquivo ignorado
não aparecia nem na árvore nem no conteúdo. O scanner fazia a parte dele (com
`IncludeIgnored=true` o nó entra na árvore, marcado), mas três filtros posteriores o
descartavam.
**Comportamento esperado:** com a flag, o arquivo ignorado aparece na árvore e seu conteúdo
entra no contexto; sem a flag, nada muda.
**Causa raiz:** três pontos, todos verificando `IsIgnored()` sem saber da intenção do scan —
`internal/core/scanner/helpers.go:52` (`selectAllExcept` nunca seleciona nó ignorado),
`internal/core/contextgen/content.go:37` (`collectFileContents` descarta nó ignorado mesmo
quando explicitamente selecionado) e `internal/core/contextgen/tree.go:66`
(`shouldSkipNode`, com `WithShowIgnored` existente porém nunca chamado em produção).
**Correção:** `SelectAllExcept` recebeu o parâmetro `includeIgnored`, repassado por
`internal/app/service.go` a partir de `scanConfig.IncludeIgnored` e pelo wizard a partir do
próprio `scanConfig`; `collectFileContents` passou a tratar o mapa de seleção como
autoridade (mapa `nil` continua significando "todos os não ignorados"); e
`contextgen.GenerateConfig` ganhou `IncludeIgnored`, usado para configurar o renderizador da
árvore.
**Teste de regressão:** `TestCLIContextGenerateIncludeIgnored` em `test/e2e/cli_test.go`
(contrato completo pela CLI) e `TestSelectAllExcept_IncludeIgnored` em
`internal/core/scanner/helpers_test.go`.
**Evidência:**

```
ANTES: sem: conteudo-ignorado=0 normal=2 arvore=0
       com: conteudo-ignorado=0 normal=2 arvore=0     <- flag sem efeito algum
DEPOIS: sem: conteudo-ignorado=0 normal=2 arvore=0
        com: conteudo-ignorado=2 normal=2 arvore=3
        ok  github.com/quantmind-br/shotgun-cli/test/e2e  0.255s
```

---

### BUG-07 — `.gitignore` aninhado perde a recursividade e deixa vazar arquivo ignorado

- **Severidade:** alta
- **Módulo:** `internal/core/ignore`
- **Invariante violado:** I7 (as regras de `.gitignore` seguem a semântica do git — o
  README e a arquitetura descrevem o motor como camada que respeita `.gitignore`)
- **Oráculo que detectou:** I7
- **Iteração:** 7

**Reprodução mínima**

```
raiz/sub/.gitignore   contendo  *.secretx
raiz/sub/deep/a.secretx
```

```go
e := NewIgnoreEngine(); e.LoadGitignore("raiz")
e.ShouldIgnore("sub/deep/a.secretx") // false
```

**Comportamento observado:** `false` — o arquivo entra no contexto enviado ao LLM. No git,
um padrão sem barra vale para todos os níveis abaixo do diretório do `.gitignore`. Um
`.gitignore` de subprojeto com `*.key`, `*.pem` ou `secrets*` protege apenas o primeiro
nível; tudo mais fundo vaza.
**Comportamento esperado:** `true` para qualquer profundidade abaixo de `sub/`.
**Causa raiz:** `internal/core/ignore/engine.go:283` (código antigo) —
`filepath.Join(relDir, line)` produz `sub/*.secretx`, que ancora o padrão num único nível.
**Correção:** o rebase virou `rebasePattern`, que distingue padrão ancorado (contém barra ou
começa com `/` → `relDir/padrão`) de padrão livre (→ `relDir/**/padrão`), preservando a
negação `!`. `LoadGitignore` e `LoadShotgunignore`, que duplicavam a lógica, passaram a
compartilhar `collectIgnorePatterns`.
**Teste de regressão:** `TestLoadGitignore_NestedPatternSemantics` e
`TestLoadShotgunignore_NestedPatternSemantics` em `internal/core/ignore/engine_test.go`
(cobrem recursividade, ancoragem, negação e escopo).
**Evidência:**

```
ANTES: --- FAIL: TestLoadShotgunignore_NestedPatternSemantics
           Should be true — padrão de .shotgunignore aninhado vale recursivamente
       --- FAIL: TestLoadGitignore_NestedPatternSemantics
DEPOIS: ok  github.com/quantmind-br/shotgun-cli/internal/core/ignore  0.017s
```

---

### BUG-08 — `dir/` de ignore aninhado vira `dir`

- **Severidade:** média
- **Módulo:** `internal/core/ignore`
- **Invariante violado:** I8 (a barra final restringe o padrão a diretórios)
- **Oráculo que detectou:** I8
- **Iteração:** 7

**Reprodução mínima**

```
raiz/sub/.gitignore  contendo  build/
```

```go
e.ShouldIgnore("sub/build") // true, mesmo sendo um ARQUIVO chamado build
```

**Comportamento observado:** um arquivo `sub/build` (sem extensão, como o binário de saída
que muitos projetos versionam sob outro nome) é excluído do contexto silenciosamente.
**Comportamento esperado:** `build/` casa apenas o diretório e seu conteúdo.
**Causa raiz:** mesma linha do BUG-07 — `filepath.Join` normaliza o caminho e **remove a
barra final**, transformando `build/` em `sub/build`.
**Correção:** `rebasePattern` concatena strings em vez de usar `filepath.Join`, preservando
a barra final.
**Teste de regressão:** casos `sub/build/artifact.o` (ignorado) e `sub/build` (não
ignorado) em `TestLoadGitignore_NestedPatternSemantics`.
**Evidência:** incluída na saída do BUG-07 (o caso `sub/build` falhava com
`ShouldIgnore("sub/build")=true(gitignore) esperado false`).

---

### BUG-09 — Um diretório ilegível aborta a carga de ignore do projeto inteiro

- **Severidade:** média
- **Módulo:** `internal/core/ignore`
- **Invariante violado:** I9 (um subdiretório inacessível degrada o resultado localmente,
  não derruba a operação)
- **Oráculo que detectou:** O5/O1 (falha propagada onde havia caminho de degradação)
- **Iteração:** 7

**Reprodução mínima**

```go
// raiz/.gitignore existe e é legível; raiz/locked tem modo 0o000
e := NewIgnoreEngine()
e.LoadGitignore("raiz") // erro: permission denied
```

**Comportamento observado:** `failed to walk directory for gitignore files: ... permission
denied`. O erro sobe pelo scanner e derruba a geração inteira, embora as regras legíveis
estivessem disponíveis. Já havia tratamento tolerante para arquivos ilegíveis
(`continue // Skip files we can't read`), mas não para diretórios.
**Comportamento esperado:** pular a subárvore inacessível e manter as regras que deram para
ler.
**Causa raiz:** `internal/core/ignore/engine.go:230` e `:396` (código antigo) — a
`filepath.WalkFunc` fazia `return err`, o que interrompe a caminhada inteira.
**Correção:** a `WalkFunc` compartilhada ignora o erro da entrada e segue a caminhada.
**Teste de regressão:** `TestLoadIgnoreFiles_UnreadableDirectory` em
`internal/core/ignore/engine_test.go` (pula quando executado como root).
**Evidência:**

```
ANTES: --- FAIL: TestLoadIgnoreFiles_UnreadableDirectory
           Received unexpected error: failed to walk directory for gitignore files:
           open .../locked: permission denied
DEPOIS: ok  github.com/quantmind-br/shotgun-cli/internal/core/ignore  0.017s
```

---

### BUG-10 — Gravações concorrentes perdem seleções e falham no `rename`

- **Severidade:** média
- **Módulo:** `internal/core/selection`
- **Invariante violado:** I10 (o arquivo do store guarda todos os projetos; salvar um
  projeto não pode apagar o de outro)
- **Oráculo que detectou:** O6/M3 (repetição concorrente + estado residual)
- **Iteração:** 8

**Reprodução mínima**

```go
s := NewStore(filepath.Join(t.TempDir(), "sel.json"))
for i := range 8 { go s.Save(fmt.Sprintf("/proj/%d", i), []string{"f.go"}) }
// depois do Wait: metade dos projetos não tem seleção salva
```

**Comportamento observado:** 4 a 6 dos 8 projetos perdem a seleção, e parte dos `Save`
falha com `commit selection store: rename .../sel.json.tmp .../sel.json: no such file or
directory`.
**Comportamento esperado:** as oito entradas presentes ao final; nenhum erro.
**Causa raiz:** `internal/core/selection/store.go:85` (código antigo) — o arquivo temporário
tinha nome fixo (`s.path + ".tmp"`), então um escritor renomeia o temporário do outro; além
disso o ciclo ler-modificar-gravar não tinha exclusão mútua, e o store é um arquivo único
compartilhado por todos os projetos.
**Correção:** `os.CreateTemp` para nome único (mais `Chmod 0600`, já que `CreateTemp` cria
com 0600 mas o `Chmod` explícito documenta a intenção e cobre umask exótico), `defer
os.Remove` do temporário e um `sync.Mutex` serializando o ciclo ler-modificar-gravar.
**Teste de regressão:** `TestStore_ConcurrentSavesKeepEveryProject` e
`TestStore_Save_UnwritableDirReportsError` em `internal/core/selection/store_test.go`
(executados também sob `-race`).
**Evidência:**

```
ANTES: --- FAIL: TestStore_ConcurrentSavesKeepEveryProject
           Save(2) err = commit selection store: rename .../sel.json.tmp .../sel.json:
             no such file or directory
           VIOLADO: 6/8 projetos perderam a seleção salva (race no read-modify-write + tmp fixo)
DEPOIS: ok  github.com/quantmind-br/shotgun-cli/internal/core/selection  1.007s   (com -race)
```

### BUG-11 — Testes de acessibilidade dependem de `NO_COLOR`

- **Severidade:** média
- **Módulo:** `internal/ui/styles`
- **Invariante violado:** I11 (o resultado da suíte não depende do ambiente de quem a roda;
  um teste ou verifica a propriedade ou se declara inaplicável)
- **Oráculo que detectou:** flakiness determinística por ambiente (fase 2, regra "rode a
  suíte duas vezes" estendida a dois ambientes)
- **Iteração:** 8 (detectado na validação final, fora do sandbox)

**Reprodução mínima**

```bash
NO_COLOR=1 go test -run Accessibility ./internal/ui/styles/   # FAIL
env -u NO_COLOR go test -run Accessibility ./internal/ui/styles/  # ok
```

**Comportamento observado:** com `NO_COLOR` definido, `TestMutedColorAccessibility` e
`TestDimTextAccessibility` falham com "insufficient contrast ratio for accessibility".
Foi assim que apareceu: a suíte passou dentro do sandbox (que roda com `--clearenv`, sem
`NO_COLOR`) e falhou na validação final no shell real, onde `NO_COLOR=1` está definido.
**Comportamento esperado:** verde nos dois ambientes. E mais: sob `NO_COLOR` o teste não
testa nada — `TierColor` devolve `lipgloss.Color("")` para *toda* cor
(`internal/ui/styles/theme.go:14-19`), então `MutedColor == Nord3` é trivialmente
verdadeiro. O teste não distingue "a paleta tem contraste ruim" de "não há paleta".
**Causa raiz:** `internal/ui/styles/theme_test.go:440` e `:448` — a asserção compara os
valores já degradados da paleta, sem checar se a paleta está ativa
(`noColor = os.Getenv("NO_COLOR") != ""`, `theme.go:12`).
**Correção:** helper `requireColorPalette` que pula os dois testes quando `noColor` está
ativo, com a razão explícita. A propriedade continua verificada em ambiente colorido, e a
suíte deixa de depender do ambiente.
**Observação de escopo:** este achado é anterior à caça (arquivos idênticos ao HEAD,
`git diff HEAD -- internal/ui/styles/` vazio antes da correção) e não tem relação com as
demais mudanças. Foi o único corrigido e validado fora do sandbox, que já havia sido
desmontado; a mudança é restrita a um arquivo `_test.go` e não executa entrada hostil.
**Teste de regressão:** os próprios `TestMutedColorAccessibility` e
`TestDimTextAccessibility`, agora determinísticos.
**Evidência:**

```
ANTES (NO_COLOR=1):
--- FAIL: TestDimTextAccessibility (0.00s)
    theme_test.go:449: DimText should not use Nord3 - insufficient contrast ratio
--- FAIL: TestMutedColorAccessibility (0.00s)
    theme_test.go:441: MutedColor should not use Nord3 - insufficient contrast ratio
FAIL	github.com/quantmind-br/shotgun-cli/internal/ui/styles	0.003s
ANTES (sem NO_COLOR): ok

DEPOIS: verde nos dois ambientes
  COM NO_COLOR=1: ok  github.com/quantmind-br/shotgun-cli/internal/ui/styles  0.004s
  SEM NO_COLOR:   ok  github.com/quantmind-br/shotgun-cli/internal/ui/styles  0.003s
```

---

## Achados de baseline (não corrigidos — exigem decisão do usuário)

- **BASE-01 (média): o job de lint do CI não roda.** `.golangci.yml` está no formato v1 e o
  `golangci-lint` atual (2.12.2, o mesmo que `golangci/golangci-lint-action@v6` com
  `version: latest` instala) recusa o arquivo:
  `can't load config: unsupported version of the configuration: ""`. Migrar a configuração
  para o schema v2 é decisão de projeto (muda linters ativos), por isso não foi feito aqui.
  Os pacotes alterados foram verificados com `golangci-lint run --no-config
  --default=standard`: sem apontamentos novos (o único achado, `content.go:300` QF1012, é
  pré-existente e de estilo).
- **BASE-02 (baixa): 4 arquivos fora do `gofmt`** já no baseline —
  `internal/app/config_test.go`, `internal/core/llm/registry_test.go`,
  `internal/core/template/loader.go`, `internal/ui/styles/theme.go`. Não foram tocados para
  não misturar ruído com as correções; os arquivos que a caça alterou estão formatados.

## Suspeitas sem reprodução (hipóteses para investigação futura)

- **Concorrência entre processos no store de seleção.** O `sync.Mutex` do BUG-10 resolve o
  caso dentro de um processo (o do TUI e o da CLI). Duas instâncias simultâneas de
  `shotgun-cli` em projetos diferentes ainda podem perder uma atualização, porque o ciclo
  ler-modificar-gravar não é atômico entre processos. Um lock de arquivo portátil é uma
  decisão de design (o pacote `core/` é stdlib-only e o projeto compila para Windows), por
  isso ficou de fora. Impacto: perder a lista de arquivos desmarcados de um projeto.
- **`IsDiffHeader` continua ambíguo por natureza.** `"--- end of section"` é ao mesmo tempo
  um cabeçalho válido (arquivo com espaço no nome) e a remoção da linha
  `"-- end of section"`. O splitter não depende mais disso — quem decide estrutura é o
  orçamento do hunk — mas o helper exportado, usado isoladamente, permanece heurístico.
- **Precedência entre `.gitignore` de níveis diferentes.** Todos os padrões são compilados
  num matcher único, na ordem da caminhada. O git avalia o `.gitignore` mais profundo com
  prioridade sobre o do pai (inclusive para negações). Não encontrei um caso reprodutível
  que quebrasse na prática, mas a modelagem não é equivalente.

## Alvos de fuzz/propriedade criados

|Alvo|Arquivo|Como rodar|
|---|---|---|
|`TestIntelligentSplit_NoLineLost`|`internal/core/diff/split_invariants_test.go`|`go test -run TestIntelligentSplit_NoLineLost ./internal/core/diff/`|
|`TestIntelligentSplit_ChunkSelfConsistent`|`internal/core/diff/split_invariants_test.go`|`go test -run TestIntelligentSplit_ChunkSelfConsistent ./internal/core/diff/`|
|`TestIntelligentSplit_ChunksApplyWithGit`|`internal/core/diff/apply_test.go`|`go test -run TestIntelligentSplit_ChunksApplyWithGit ./internal/core/diff/`|

Os dois primeiros geram diffs aleatórios determinísticos (`rand.NewPCG(2026, 727)` e
`rand.NewPCG(99, 7)`, 400 iterações cada, `ApproxLines` de 1 a 30) e injetam de propósito a
linha `+++i;` como conteúdo — foi esse corpus que expôs BUG-02, BUG-03 e BUG-04. O terceiro
é o oráculo forte: `git apply --check` em cada chunk emitido.

## Limite conhecido e documentado do splitter

Um hunk isolado maior que `ApproxLines` **não** é partido: fatiar um hunk exigiria
recalcular os contadores de `@@ -a,b +c,d @@` de cada fragmento, e um chunk com contadores
errados é rejeitado por qualquer ferramenta de patch. O comportamento está fixado em
`TestIntelligentSplit_SingleHunkExceedsBudget` e é coerente com a documentação do campo
("The actual chunk size may vary to preserve file boundaries").

## Orçamento e critério de parada

- **Iterações do laço:** 8.
- **Alvos cobertos:** `internal/core/diff`, `cmd` (comando diff), `internal/core/ignore`,
  `internal/core/scanner` (helpers de seleção), `internal/core/contextgen`, `internal/app`,
  `internal/core/selection`, `internal/ui/styles`.
- **Motivo da parada:** meta atingida — 11 bugs confirmados, corrigidos e com teste de
  regressão verde (o critério de parada da skill é 10).
- **Não investigados a fundo:** `internal/platform/*` (clientes LLM; `llmbase` sem testes e
  `platform/http` em 50% de cobertura continuam sendo o alvo de maior retorno esperado numa
  próxima rodada) e as máquinas de estado do TUI.

## Validação final

Fase 2 repetida dentro do sandbox, sobre exatamente os arquivos entregues:

```
go build ./...                     OK
go vet ./...                       OK
go test -skip "TestScanCoordinator|TestGenerateCoordinator|TestWizardClipboardCopyCmd" ./...
  23 pacotes ok, 0 FAIL
go test -race -skip "<mesmos>" ./...
  RACE_EXIT=0, 23 pacotes ok, nenhum DATA RACE
gofmt -l   → apenas os 4 arquivos já fora de formato no baseline (BASE-02)
golangci-lint run --no-config --default=standard  (pacotes alterados) → nenhum achado novo
```

E repetida na árvore real, já com as correções aplicadas, para confirmar o entregável no
lugar onde ele vai viver — foi essa execução que revelou o BUG-11:

```
go build ./...   OK
go vet ./...     OK
go test -skip "<mesmos>" ./...   verde com NO_COLOR=1 e sem NO_COLOR
```

### Delta de cobertura por módulo (baseline → final)

|Pacote|Antes|Depois|Delta|
|---|---|---|---|
|`internal/core/ignore`|89,6%|95,4%|**+5,8**|
|`internal/ui`|70,5%|70,6%|+0,1|
|`internal/core/selection`|79,5%|78,8%|−0,7|
|`internal/core/contextgen`|91,2%|90,6%|−0,6|
|`internal/core/diff`|98,4%|96,9%|−1,5|
|**Total**|**82,4%**|**82,7%**|**+0,3**|

As quedas em `diff`, `contextgen` e `selection` são de denominador: o parser novo, o ramo
de seleção explícita e os caminhos de erro da escrita atômica acrescentam instruções, e as
poucas linhas descobertas são ramos de erro de E/S difíceis de forçar de forma
determinística (`os.CreateTemp` falhando após `MkdirAll` ter sucedido, por exemplo). Os
pacotes seguem acima do piso de 80% exigido pelo CI.
