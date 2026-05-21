# Análise de Código — `internal/core/diff`

| Campo | Valor |
|-------|-------|
| **Módulo** | `internal/core/diff` |
| **Package Path** | `github.com/quantmind-br/shotgun-cli/internal/core/diff` |
| **Tipo** | Domain utility (zero external dependencies) |
| **Linguagem** | Go |
| **Go Version** | 1.24.0 |
| **Nível de Detalhe** | Detalhado |

---

## 1. Visão Geral

O pacote `diff` fornece utilitários puros para processamento e divisão (**splitting**) de arquivos de diff no formato unified diff / git diff. A função central, `IntelligentSplit`, divide diffs grandes em pedaços (`Chunk`s) que respeitam os limites entre arquivos, evitando a fragmentação de hunks e garantindo a integridade estrutural de cada parte.

O pacote é utilizado exclusivamente pela CLI command `shotgun-cli diff split` (`cmd/diff.go`), que lê um arquivo de diff, invoca `IntelligentSplit` e grava os chunks em um diretório de saída.

**Dependências externas**: nenhuma. Usa apenas o pacote padrão `strings`.

**Dependentes**: `cmd` (`cmd/diff.go`, `cmd/diff_test.go`).

---

## 2. Arquivos do Pacote

### 2.1 `split.go` — Implementação principal

- **Linha de入口**: linha 1
- **Total de linhas**: ~180
- **Dependências internas**: `strings` (pacote padrão)
- **Tipos exportados**: `Chunk`, `SplitConfig`
- **Funções exportadas**: `DefaultSplitConfig`, `IntelligentSplit`, `IsDiffHeader`, `IsGitDiffHeader`, `CanSplitAt`, `CountFiles`, `TotalLines`, `TotalFiles`
- **Funções não exportadas**: nenhuma
- **Complexidade ciclomática**: moderada (função `IntelligentSplit` possui múltiplos caminhos)

### 2.2 `split_test.go` — Testes unitários

- **Total de linhas**: ~300
- **Cobertura**: testes para todas as funções exportadas
- **Framework**: `testing` + `testify/assert` + `testify/require`
- **Número de casos de teste**: ~15 casos nominais + subtestes

---

## 3. Tipos Detalhados

### 3.1 `Chunk` (struct)

```go
type Chunk struct {
    Lines     []string  // Linhas de conteúdo do diff nesta parte
    FileCount int       // Número de arquivos representados neste chunk
    StartLine int       // Número da linha original de início (1-indexed)
}
```

| Campo | Tipo | Acessibilidade | Descrição |
|-------|------|----------------|-----------|
| `Lines` | `[]string` | exported | Conjunto de linhas que compõem o chunk. Cada string é uma linha individual do diff original. |
| `FileCount` | `int` | exported | Quantidade de arquivos (`diff --git` ou `---`) contidos neste chunk. |
| `StartLine` | `int` | exported | Linha 1-indexada do diff original onde este chunk começa. Útil para rastreamento e referência. |

**Uso**: Retornado por `IntelligentSplit`. Cada `Chunk` é um pedaço independente e aplicável de um diff.

**Invariants**:
- `Lines` nunca é `nil` (sempre tem pelo menos 0 elementos).
- `FileCount` ≥ 0.
- `StartLine` ≥ 1.

---

### 3.2 `SplitConfig` (struct)

```go
type SplitConfig struct {
    ApproxLines int  // Número alvo de linhas por chunk
}
```

| Campo | Tipo | Acessibilidade | Descrição |
|-------|------|----------------|-----------|
| `ApproxLines` | `int` | exported | Limite alvo de linhas para cada chunk. O split real pode exceder este valor para respeitar limites de arquivos. |

**Invariants**:
- Se `ApproxLines ≤ 0`, o valor é corrigido internamente para `500` em `IntelligentSplit`.
- Não há limite máximo definido.

---

## 4. Funções Exportadas — Detalhamento

### 4.1 `DefaultSplitConfig() SplitConfig`

- **Pacote**: `diff`
- **Retorna**: `SplitConfig{ApproxLines: 500}`
- **Efeito colateral**: nenhum
- **Complexidade**: O(1)
- **Uso**: Ponto de entrada para obter configuração padrão.

---

### 4.2 `IntelligentSplit(lines []string, config SplitConfig) []Chunk`

- **Pacote**: `diff`
- **Parâmetros**:
  - `lines`: slice de strings representando o diff completo (uma linha por string)
  - `config`: configuração de split (deflata para `ApproxLines=500` se inválida)
- **Retorna**: `[]Chunk` — fatia de chunks resultantes
- **Efeito colateral**: nenhum

**Algoritmo (passo a passo)**:

1. **Caso vazio**: Se `len(lines) == 0`, retorna `[Chunk{Lines: nil, FileCount: 0, StartLine: 1}]`.
2. **Fallback de config**: Se `config.ApproxLines ≤ 0`, define para `500`.
3. **Inicialização**:
   - Cria `chunks = []`, `currentChunk = Chunk{}`, `currentFileLines = []`, `fileCount = 0`, `inFileSection = false`.
4. **Loop principal** (para cada linha `i` em `lines`):
   - Apenda a linha a `currentChunk.Lines` e `currentFileLines`.
   - Se a linha for um header de arquivo (`IsDiffHeader` ou `IsGitDiffHeader`):
     - Se já estávamos em uma seção de arquivo e `currentFileLines` tem mais de 1 linha, incremente `fileCount`.
     - Marque `inFileSection = true` e resete `currentFileLines = [line]`.
   - Calcule `shouldSplit`: `len(currentChunk.Lines) >= config.ApproxLines` **E** `CanSplitAt(lines, i, inFileSection)`.
   - Se `shouldSplit`:
     - Defina `currentChunk.FileCount` (incluindo arquivo atual se `inFileSection`).
     - Append `currentChunk` a `chunks`.
     - Reset: `currentChunk = Chunk{StartLine: i + 2}`, `fileCount = 0`, `inFileSection = false`, `currentFileLines = []`.
5. **Finalização**:
   - Se `currentChunk.Lines` não é vazio, conte o último arquivo e append ao resultado.
   - Se `chunks` está vazio, append um `Chunk` com todos os linhas e conte os arquivos.
6. **Retorna** `chunks`.

**Complexidade**: O(n) onde n = `len(lines)`. Cada linha é processada exatamente uma vez.

**Casos de canto**:
- Diff menor que `ApproxLines`: retornado como um único chunk.
- `ApproxLines = 0` ou negativo: comportamento como `500`.
- Diff sem headers: retornado como um único chunk sem splitting.

---

### 4.3 `IsDiffHeader(line string) bool`

- **Pacote**: `diff`
- **Retorna**: `true` se a linha começa com `"---"` ou `"+++"`
- **Efeito colateral**: nenhum
- **Complexidade**: O(m) onde m = comprimento do prefixo

**Descrição**: Detecta headers de diff no formato unified diff tradicional (`--- a/file.go` / `+++ b/file.go`).

---

### 4.4 `IsGitDiffHeader(line string) bool`

- **Pacote**: `diff`
- **Retorna**: `true` se a linha começa com `"diff --git"`
- **Efeito colateral**: nenhum
- **Complexidade**: O(m)

**Descrição**: Detecta headers de diff no formato git (`diff --git a/file.go b/file.go`).

---

### 4.5 `CanSplitAt(lines []string, index int, inFileSection bool) bool`

- **Pacote**: `diff`
- **Retorna**: `true` se a posição `index` é um ponto seguro para split
- **Efeito colateral**: nenhum

**Lógica**:

1. Se `index >= len(lines) - 1`: retorna `false` (última linha — não há próxima para validar).
2. Verifica a próxima linha (`lines[index+1]`):
   - Se for um novo header de arquivo (`IsGitDiffHeader` ou `IsDiffHeader`): `true`.
   - Se não estamos em seção de arquivo (`!inFileSection`): `true` (ponto seguro entre seções).
   - Se a próxima linha é um header de hunk (`@@`): `true` (seguro split antes de um novo hunk).
3. Caso contrário: `false` (dividir no meio do arquivo quebraria hunks).

**Casos de canto**:
- Index é a última linha: sempre `false`.
- Dentro de uma seção de arquivo, no meio de mudanças: `false`.

---

### 4.6 `CountFiles(lines []string) int`

- **Pacote**: `diff`
- **Retorna**: Contagem de headers de arquivo (git diff headers + unified `---` headers)
- **Efeito colateral**: nenhum

**Lógica**: Itera todas as linhas, incrementa para cada `diff --git` ou `---` encontrado.

**Nota**: Um único arquivo git diff conta 2 (header `diff --git` + header `---`), o que significa que `CountFiles` não retorna o número real de arquivos, mas sim o número de markers. O caller (`cmd/diff.go`) não depende diretamente da precisão do `FileCount`.

---

### 4.7 `TotalLines(chunks []Chunk) int`

- **Pacote**: `diff`
- **Retorna**: Soma dos `len(chunk.Lines)` para todos os chunks
- **Uso**: Verificação de integridade (total de linhas antes/depois do split deve ser igual)
- **Complexidade**: O(k) onde k = número de chunks

---

### 4.8 `TotalFiles(chunks []Chunk) int`

- **Pacote**: `diff`
- **Retorna**: Soma dos `chunk.FileCount` para todos os chunks
- **Uso**: Agregação de contagem de arquivos em chunks
- **Complexidade**: O(k)

---

## 5. Dependências

### 5.1 Dependências Diretas (imports do pacote)

| Import | Tipo | Finalidade |
|--------|------|------------|
| `strings` | stdlib | Prefixo detection (`HasPrefix`) |

### 5.2 Dependentes (quem importa este pacote)

| Pacote | Arquivo | Uso |
|--------|---------|-----|
| `cmd` | `cmd/diff.go` | Chama `diff.IntelligentSplit()`, `diff.SplitConfig{}` |
| `cmd` | `cmd/diff_test.go` | Testa todas as funções exportadas via `diff.` |

---

## 6. Métricas de Código

| Métrica | Valor |
|---------|-------|
| Arquivos Go | 2 |
| Linhas totais (implementação) | ~180 |
| Linhas totais (testes) | ~300 |
| Funções exportadas | 8 |
| Funções não exportadas | 0 |
| Tipos exportados | 2 |
| Funções não exportadas de teste | 0 |
| Testes unitários (nominais) | ~15 |
| Cobertura inferida | 100% das funções exportadas testadas |
| Complexidade ciclomática (IntelligentSplit) | ~8 caminhos |
| Cyclomatic Complexity (total) | ~12 |

---

## 7. Regras e Invariants Identificados

1. **Split respeita limites de arquivo**: O algoritmo nunca divide no meio de um bloco de diff de um arquivo. Só splita em fronteira de arquivo ou antes de um hunk header (`@@`).
2. **Chunk vazio é retornado como único chunk**: Se input vazio, retorna `[Chunk{}]`.
3. **Fallback de configuração**: `ApproxLines <= 0` é corrigido para `500` automaticamente.
4. **Integridade de linhas preservada**: `TotalLines(chunks)` deve ser igual a `len(lines)` após split.
5. **StartLine é 1-indexed**: Sempre ≥ 1, calculado como `i + 2` após um split (linha atual + 1 para a próxima linha após a atual).
6. **CountFiles não conta arquivos únicos corretamente**: Conta markers (`diff --git` + `---`), não arquivos. Um arquivo git diff gera contagem 2. **🟡 INFERIDO**: Este pode ser um bug ou intencional — `FileCount` não é usado de forma crítica.

---

## 8. Observações Arquiteturais

- **Responsabilidade única**: O pacote tem um único propósito claro — dividir diffs em chunks. Não há responsabilidade colateral.
- **Zero dependências externas**: Depende apenas de `strings` do stdlib, o que o torna facilmente portável.
- **Público vs. privado**: Todas as funções são exportadas. Não há funções internas privadas, o que sugere design minimalista — não havia necessidade de abstração interna.
- **Origem**: Conforme o `.beads/sync_base.jsonl`, este pacote foi extraído de `cmd/diff.go` como parte de uma refatoração para mover lógica pura para `internal/core/`, tornando-a testável sem overhead de CLI.
- **Padrão funcional options**: Usa struct `SplitConfig` que pode ser estendida no futuro com opções adicionais (ex: `ForceSplit`, `MinChunkSize`).

---

## 9. Fluxos de Execução Principais

### Fluxo 1: `IntelligentSplit` (split de diff)

```
Input: []string (linhas do diff) + SplitConfig
  → Valida input (vazio? config inválida?)
  → Loop sobre cada linha:
      → Append a currentChunk
      → Detecta headers de arquivo
      → Avalia CanSplitAt na linha atual
  → Se split necessário:
      → Finaliza chunk atual, salva, reinicia
  → Finaliza último chunk
  → Output: []Chunk
```

### Fluxo 2: `diff split` (CLI command)

```
Input: --input <arquivo>, --approx-lines <n>, --output-dir <dir>
  → Valida flags (input obrigatório, approx-lines > 0)
  → Abre arquivo de entrada
  → Lê todas as linhas com bufio.Scanner
  → Chama diff.IntelligentSplit(allLines, SplitConfig{ApproxLines})
  → Para cada chunk:
      → Escreve arquivo de chunk com cabeçalho opcional
  → Output: Arquivos de chunk em outputDir
```

---

## 10. Gaps / Perguntas para Revisão Humana

| # | Pergunta | Arquivo de Referência |
|---|----------|----------------------|
| G1 | `CountFiles` conta markers, não arquivos únicos. Isso é bug ou feature? O caller em `cmd/diff.go` exibe `chunk.FileCount` como "files in this chunk" — pode ser enganoso. | `split.go` linha ~155 |
| G2 | Não há testes de boundary para `StartLine`. O cálculo `i + 2` assume que a split ocorre após a linha atual — isso está documentado? | `split.go` linha ~90 |
| G3 | `CanSplitAt` permite split em qualquer ponto se `!inFileSection`. Isso significa que diffs com metadados antes do primeiro arquivo podem ser divididos arbitrariamente. É desejável? | `split.go` linha ~140 |
| G4 | O pacote não exporta a lógica de escrita de chunks (isso fica em `cmd/diff.go`). Deveria haver uma função `WriteChunk` neste pacote para melhor separação de responsabilidade? | `cmd/diff.go` função `writeDiffChunk` |
