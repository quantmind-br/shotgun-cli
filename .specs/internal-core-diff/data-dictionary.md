# Dicionário de Dados — `internal/core/diff`

| Campo | Valor |
|-------|-------|
| **Módulo** | `internal/core/diff` |
| **Package Path** | `github.com/quantmind-br/shotgun-cli/internal/core/diff` |
| **Nível de Detalhe** | Detalhado |
| **Idioma** | pt-br |

---

## 1. Tipos de Dados

### 1.1 `diff.Chunk`

**Descrição**: Representa um pedaço (chunk) de diff resultante de uma operação de splitting. Contém as linhas de texto do diff, a contagem de arquivos no chunk e a linha original de início.

**Declaração**:
```go
type Chunk struct {
    Lines     []string
    FileCount int
    StartLine int
}
```

| Campo | Tipo | Cardinalidade | Obrigatório | Nulo | Descrição | Restrições |
|-------|------|---------------|-------------|------|-----------|------------|
| `Lines` | `[]string` | 0..N | Sim | Não | Linhas individuais do diff neste chunk. Cada string é uma linha do diff original. | Sempre não-nil (pode ser 0 elementos em input vazio). Cada string contém uma linha sem quebra de linha. |
| `FileCount` | `int` | 1 | Sim | Não (valor 0) | Número de markers de arquivo (`diff --git` ou `---`) neste chunk. | ≥ 0. Nota: conta markers, não arquivos únicos (ver Gap G1). |
| `StartLine` | `int` | 1 | Sim | Não (valor 1) | Número 1-indexado da linha original onde este chunk começa. | ≥ 1. Primeiro chunk = 1. Subsequent chunks = índice da linha de split + 2. |

**Relações**:
- **Composição de**: `diff.IntelligentSplit()`
- **Consumido por**: `cmd.writeDiffChunk()` (via `cmd/diff.go`)
- **Agregado por**: `diff.TotalLines()`, `diff.TotalFiles()`

**Exemplo de instância**:
```go
Chunk{
    Lines:     []string{"diff --git a/cmd/diff.go b/cmd/diff.go", "--- a/cmd/diff.go", "+++ b/cmd/diff.go"},
    FileCount: 2,
    StartLine: 1,
}
```

---

### 1.2 `diff.SplitConfig`

**Descrição**: Configuração paramétrica para a operação de splitting. Define o tamanho alvo dos chunks.

**Declaração**:
```go
type SplitConfig struct {
    ApproxLines int
}
```

| Campo | Tipo | Cardinalidade | Obrigatório | Nulo | Descrição | Restrições |
|-------|------|---------------|-------------|------|-----------|------------|
| `ApproxLines` | `int` | 1 | Sim | Não (valor 500) | Número alvo de linhas por chunk. O split real pode exceder para respeitar limites de arquivo. | Se ≤ 0, corrigido internamente para 500. Sem limite superior. |

**Valores padrão**:
- `ApproxLines = 500` (via `DefaultSplitConfig()`)

**Relações**:
- **Criado por**: `diff.DefaultSplitConfig()` (padrão), caller (configuração customizada)
- **Consumido por**: `diff.IntelligentSplit()`

**Exemplo de instância**:
```go
SplitConfig{ApproxLines: 1000}  // Chunk maior que o padrão
SplitConfig{ApproxLines: 500}   // Valor padrão
```

---

## 2. Valores Constantes

| Valor | Tipo | Localização | Descrição |
|-------|------|-------------|-----------|
| `500` | `int` | `DefaultSplitConfig()`, `IntelligentSplit()` | Tamanho padrão alvo de linhas por chunk. Usado como fallback quando `ApproxLines ≤ 0`. |

---

## 3. Tipos de Dados de Entrada/Saída

### 3.1 Entrada: `[]string` (linhas de diff)

**Descrição**: Fatia de strings onde cada string é uma linha individual de um diff em formato unified diff / git diff.

**Formato das linhas**:
```
diff --git <path_a> <path_b>       ← Header git diff
index <hash_a>..<hash_b> <mode>    ← Info de índice
--- <path_a>                       ← Unified diff: arquivo antigo
+++ <path_b>                       ← Unified diff: arquivo novo
@@ <old_range> <new_range> @@      ← Header de hunk
 <context>                          ← Linha de contexto (prefixo espaço)
-<removed>                          ← Linha removida
+<added>                            ← Linha adicionada
```

**Invariantes**:
- Cada string não contém caractere de nova-linha (`\n`).
- A ordem das strings corresponde à ordem no arquivo original.
- A entrada pode estar vazia (`[]string{}`).

---

### 3.2 Saída: `[]Chunk`

**Descrição**: Fatia de `Chunk` resultante do split. Cada chunk é um subconjunto válido e auto-contido do diff original.

**Invariantes**:
- `TotalLines(chunks) == len(lines)` (integridade de linhas preservada).
- Nenhum chunk contém uma linha incompleta (nunca divide no meio de um hunk).
- O primeiro chunk sempre tem `StartLine = 1`.
- Chunks subsequentes têm `StartLine > anterior.StartLine`.
- Se input é vazio, retorna `[Chunk{Lines: nil, FileCount: 0, StartLine: 1}]`.

---

## 4. Valores Internos Temporários (não exportados)

Estes são valores usados internamente pelo algoritmo `IntelligentSplit`, visíveis no código-fonte mas não expostos via interface pública.

| Variável | Tipo | Finalidade |
|----------|------|------------|
| `currentChunk` | `Chunk` | Acumulador do chunk atual em construção. |
| `currentFileLines` | `[]string` | Buffer das linhas desde o último header de arquivo detectado. Usado para contar arquivos. |
| `fileCount` | `int` | Contador de arquivos na chunk atual. Incrementado ao detectar um novo header. |
| `inFileSection` | `bool` | Flag que indica se estamos dentro de uma seção de diff de arquivo (após detectar header). |
| `shouldSplit` | `bool` | Resultado da avaliação de split: `len(currentChunk.Lines) >= ApproxLines AND CanSplitAt(...)`. |

---

## 5. Enumerações / Tipos de Sentinela

Nenhum tipo enum ou sentinela é definido neste pacote. O pacote usa apenas:

| Conceito | Tipo | Valores |
|----------|------|---------|
| Flag booleano | `bool` | `true` = dentro de seção de arquivo, `false` = fora |
| Contador | `int` | ≥ 0 |

---

## 6. Fluxo de Dados — Mapeamento Completo

```
┌──────────────────────────────────────────────────────────────────┐
│                        cmd/diff.go                               │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  splitDiffFile()                                           │  │
│  │    1. Lê arquivo com bufio.Scanner                         │  │
│  │    2. armazena em []string (allLines)                      │  │
│  │    3. Chama: diff.IntelligentSplit(allLines, config)       │  │
│  └───────────────────────┬────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │              internal/core/diff                             │  │
│  │    ┌───────────────────┐   ┌───────────────────┐           │  │
│  │    │ SplitConfig       │   │ []string (input)  │           │  │
│  │    │ ApproxLines: 500  │   │ Lines[0..N]       │           │  │
│  │    └────────┬──────────┘   └────────┬──────────┘           │  │
│  │             │                       │                       │  │
│  │             │      IntelligentSplit │                       │  │
│  │             │◄──────────────────────┤                       │  │
│  │             │                       │                       │  │
│  │             └──────────────────────►│                       │  │
│  │                                     │                       │  │
│  │                                     ▼                       │  │
│  │                              []Chunk                         │  │
│  │                              ├── Chunk{Lines:...,           │  │
│  │                              │     FileCount: N,            │  │
│  │                              │     StartLine: 1}            │  │
│  │                              ├── Chunk{Lines:...,           │  │
│  │                              │     FileCount: M,            │  │
│  │                              │     StartLine: K}            │  │
│  │                              └── ...                          │  │
│  └────────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  writeDiffChunk()                                           │  │
│  │    1. Cria arquivo em outputDir                            │  │
│  │    2. Escreve cabeçalho opcional                           │  │
│  │    3. Escreve chunk.Lines como conteúdo                    │  │
│  └────────────────────────────────────────────────────────────┘  │
│                          │                                        │
│                          ▼                                        │
│                 Arquivo de chunk (.diff)                          │
└──────────────────────────────────────────────────────────────────┘
```

---

## 7. Glossário de Termos Técnicos

| Termo | Definição |
|-------|-----------|
| **Diff** | Diferença entre duas versões de arquivos, representada no formato unified diff. |
| **Git Diff Header** | Linha que inicia a descrição de um arquivo alterado: `diff --git a/<path> b/<path>`. |
| **Unified Diff Header** | Linhas `---` (arquivo antigo) e `+++` (arquivo novo) em formato unified diff. |
| **Hunk** | Bloco de mudanças adjacentes em um diff, delimitado por headers `@@`. |
| **Chunk** | Pedaço resultante do split de um diff. Contém múltiplas linhas e possivelmente múltiplos arquivos. |
| **Split Point** | Posição segura no diff onde o split pode ocorrer sem quebrar a estrutura. |
| **File Section** | Conjunto de linhas pertencentes a um único arquivo dentro de um diff. |
| **Marker** | Linha que indica a presença de um arquivo no diff (`diff --git` ou `---`). |
| **InFileSection** | Estado booleano: `true` = após ter detectado um header de arquivo; `false` = antes do primeiro header ou entre seções. |
| **ApproxLines** | Número aproximado de linhas alvo para cada chunk. O split real pode exceder este valor. |
| **StartLine** | Número da linha 1-indexada do diff original onde o chunk começa. |
