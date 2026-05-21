# Fluxograma — `IntelligentSplit`

| Campo | Valor |
|-------|-------|
| **Módulo** | `internal/core/diff` |
| **Função** | `IntelligentSplit(lines []string, config SplitConfig) []Chunk` |
| **Arquivo** | `split.go` |
| **Nível de Detalhe** | Detalhado |

---

## 1. Diagrama de Fluxo (Mermaid)

```mermaid
flowchart TD
    Start([Início: IntelligentSplit]) --> CheckEmpty{lines\nvazio?}
    CheckEmpty -->|Sim| RetEmpty[retorna Chunk{\n  Lines: nil,\n  FileCount: 0,\n  StartLine: 1\n}]
    RetEmpty --> End([Fim: retorna []Chunk])

    CheckEmpty -->|Não| CheckConfig{ApproxLines\n≤ 0?}
    CheckConfig -->|Sim| SetDefault[config.ApproxLines = 500]
    CheckConfig -->|Não| InitVars
    SetDefault --> InitVars

    InitVars[Inicialização:\n  chunks = []\n  currentChunk = Chunk{}\n  currentFileLines = []\n  fileCount = 0\n  inFileSection = false\n  i = 0] --> LoopStart{Linha i <\nlen(lines)?}

    LoopStart -->|Sim| AppendLines[Append linha\n  currentChunk.Lines\n  currentFileLines]

    AppendLines --> CheckHeader{Linha é\nheader de\narquivo?\n(IsDiffHeader ou\nIsGitDiffHeader)}

    CheckHeader -->|Sim| HandleHeader[Se inFileSection\n  && currentFileLines\n     length > 1:\n    fileCount++\ninFileSection = true\ncurrentFileLines = [linha]]

    HandleHeader --> EvalSplit
    CheckHeader -->|Não| EvalSplit

    EvalSplit{shouldSplit:\n  len(currentChunk.Lines)\n  ≥ ApproxLines\n  AND CanSplitAt?\n  (linha i)} -->|Não| NextLine[i++] --> LoopStart

    EvalSplit -->|Sim| FinalizeChunk[Finaliza chunk atual:\n  currentChunk.FileCount\n  = fileCount\n  (se inFileSection\n    && length > 1:\n    FileCount++)\n  chunks.append(\n    currentChunk)\n\n  currentChunk = Chunk{\n    StartLine: i + 2}\n  fileCount = 0\n  inFileSection = false\n  currentFileLines = []]

    FinalizeChunk --> NextLine

    LoopStart -->|Não| PostLoop{currentChunk\nLines vazio?}
    PostLoop -->|Sim| CheckEmptyChunks{chunks\nvazio?}
    PostLoop -->|Não| FinalLastChunk[Se inFileSection\n  && length > 1:\n  fileCount++\ncurrentChunk.FileCount\n  = fileCount\n  chunks.append(\n    currentChunk)]

    FinalLastChunk --> CheckEmptyChunks

    CheckEmptyChunks -->|Sim| ForceSingle[chunks.append(\n  Chunk{\n    Lines: lines,\n    FileCount: CountFiles(\n      lines),\n    StartLine: 1\n  })]
    CheckEmptyChunks -->|Não| ReturnChunks
    ForceSingle --> ReturnChunks

    ReturnChunks([retorna chunks]) --> End

    NextLine --> LoopStart
```

---

## 2. Fluxograma Detalhado — Decisão de Split

O ponto mais crítico da função é a decisão de quando splitar. Abaixo está o fluxograma da função `CanSplitAt`, usada dentro de `IntelligentSplit`.

```mermaid
flowchart TD
    StartCanSplit([CanSplitAt(lines, index, inFileSection)]) --> CheckIndex{index ≥\nlen(lines) - 1?}
    CheckIndex -->|Sim| RetFalse([retorna false])
    CheckIndex -->|Não| GetNextLine[Obtém\nnextLine =\nlines[index + 1]]

    GetNextLine --> CheckGitHeader{nextLine é\nIsGitDiffHeader?\n(diff --git)}
    CheckGitHeader -->|Sim| RetGit([retorna true])

    CheckGitHeader -->|Não| CheckDiffHeader{nextLine é\nIsDiffHeader?\n(--- ou +++)}
    CheckDiffHeader -->|Sim| RetDiff([retorna true])

    CheckDiffHeader -->|Não| CheckNotInSection{!inFileSection?}
    CheckNotInSection -->|Sim| RetNotIn([retorna true])

    CheckNotInSection -->|Não| CheckHunk{nextLine é\n'@@'}
    CheckHunk -->|Sim| RetHunk([retorna true])

    CheckHunk -->|Não| RetMiddle([retorna false])

    RetGit --> End
    RetDiff --> End
    RetNotIn --> End
    RetHunk --> End
    RetMiddle --> End
```

---

## 3. Fluxograma do Caller: CLI `diff split`

```mermaid
flowchart TD
    StartCLI([Início: CLI command\ndiff split]) --> ValidateInput{input flag\né fornecido?}
    ValidateInput -->|Não| ErrInput([erro: input obrigatório])
    ValidateInput -->|Sim| CheckExists{arquivo\nexiste?}
    CheckExists -->|Não| ErrNotFound([erro: arquivo não existe])
    CheckExists -->|Sim| CheckReadable{arquivo\nlegível?}
    CheckReadable -->|Não| ErrRead([erro: não pode ler])
    CheckReadable -->|Sim| ValidateApprox{approx-lines\n> 0?}
    ValidateApprox -->|Não| ErrApprox([erro: approx-lines\ninválido])
    ValidateApprox -->|Sim| OpenFile[abre arquivo]
    OpenFile --> ReadLines[bufio.Scanner\nlê todas as linhas\n→ []string allLines]
    ReadLines --> CheckLinesEmpty{allLines\nvazio?}
    CheckLinesEmpty -->|Sim| ErrEmpty([erro: arquivo vazio])
    CheckLinesEmpty -->|Não| CallSplit[Chama:\n  diff.IntelligentSplit(\n    allLines,\n    SplitConfig{ApproxLines})\n  → []Chunk chunks]

    CallSplit --> LoopChunks{para cada\nchunk em chunks}
    LoopChunks -->|Sim| BuildFilename[Constrói nome:\n  input-chunk-NN.diff]
    BuildFilename --> WriteChunk[writeDiffChunk:\n  1. Cria arquivo\n  2. Header opcional\n  3. Escreve lines]
    WriteChunk --> PrintInfo[Imprime:\n  Chunk N: filename\n  (L lines, F files)]
    PrintInfo --> NextChunk

    LoopChunks -->|Não| SuccessMsg[Imprime sucesso]
    SuccessMsg --> EndCLI([Fim CLI])

    ErrInput --> EndCLI
    ErrNotFound --> EndCLI
    ErrRead --> EndCLI
    ErrApprox --> EndCLI
    ErrEmpty --> EndCLI

    NextChunk --> LoopChunks
```

---

## 4. Descrição Passo a Passo — `IntelligentSplit`

### Fase 1: Validação Inicial

| Passo | Ação | Condição | Resultado |
|-------|------|----------|-----------|
| 1 | Verifica se `lines` é vazio | `len(lines) == 0` | Retorna `[Chunk{Lines: nil, FileCount: 0, StartLine: 1}]` |
| 2 | Valida `ApproxLines` | `config.ApproxLines <= 0` | Corrige para `500` |

### Fase 2: Inicialização

| Variável | Valor Inicial | Finalidade |
|----------|---------------|------------|
| `chunks` | `[]` (vazio) | Acumulador de chunks resultantes |
| `currentChunk` | `Chunk{}` | Chunk atual sendo construído |
| `currentFileLines` | `[]string{}` | Buffer de linhas desde o último header de arquivo |
| `fileCount` | `0` | Contador de arquivos na chunk atual |
| `inFileSection` | `false` | Flag de estado: dentro de seção de arquivo? |

### Fase 3: Loop Principal (percorre cada linha)

Para cada linha `i` do `0` a `len(lines) - 1`:

1. **Acumula linha**: Adiciona linha a `currentChunk.Lines` e `currentFileLines`.
2. **Detecta header**: Se `IsDiffHeader(line)` ou `IsGitDiffHeader(line)`:
   - Se já estava em seção (`inFileSection == true`) e há mais de 1 linha em `currentFileLines`, incrementa `fileCount`.
   - Reinicia: `inFileSection = true`, `currentFileLines = [linha]`.
3. **Avalia split**: `shouldSplit = (len(currentChunk.Lines) >= ApproxLines) AND CanSplitAt(lines, i, inFileSection)`
4. **Se deve splitar**:
   - Finaliza `currentChunk.FileCount`.
   - Salva `currentChunk` em `chunks`.
   - Reinicia: `currentChunk = {StartLine: i + 2}`, `fileCount = 0`, `inFileSection = false`, `currentFileLines = []`.

### Fase 4: Finalização Pós-Loop

1. Se `currentChunk.Lines` tem conteúdo:
   - Conta último arquivo se `inFileSection`.
   - Salva `currentChunk` em `chunks`.
2. Se `chunks` está vazio (nenhum split ocorreu):
   - Cria um único chunk com todas as linhas e `FileCount = CountFiles(lines)`.
3. Retorna `chunks`.

---

## 5. Fluxograma de Decisão — `CanSplitAt`

| Condição Avaliada | Resultado | Justificativa |
|-------------------|-----------|---------------|
| `index >= len(lines) - 1` | **NO split** | Última linha — não há próxima para validar. Split aqui quebraria o final. |
| `nextLine` começa com `diff --git` | **SPLIT** | Início de um novo arquivo. Fronteira clara. |
| `nextLine` começa com `---` ou `+++` | **SPLIT** | Header unified diff. Fronteira clara de arquivo. |
| `!inFileSection` | **SPLIT** | Não estamos em uma seção de arquivo. Qualquer ponto é seguro. |
| `nextLine` começa com `@@` | **SPLIT** | Início de um hunk. Dividir antes do hunk preserva a estrutura. |
| Nenhuma das acima | **NO split** | Estamos no meio de um arquivo (linhas de contexto/adicionadas/removidas). Dividir quebraria hunks. |

---

## 6. Casos de Teste de Fluxo

### Caso A: Diff pequeno (sem split)

```
Input:  7 linhas, ApproxLines = 500
Fase 1: Não vazio, ApproxLines ok (500)
Fase 2: Inicializa variáveis
Fase 3: Loop de 0 a 6 — nunca atinge 500 linhas, never shouldSplit=true
Fase 4: Salva chunk final com todas as 7 linhas
Output: [Chunk{Lines: 7 linhas, FileCount: 2, StartLine: 1}]
```

### Caso B: Diff grande com múltiplos arquivos

```
Input:  1500 linhas, 3 arquivos, ApproxLines = 500
Fase 1-2: Igual
Fase 3:
  - Linha 0-499: Acumula, atinge 500, CanSplitAt? Sim (entre arquivos) → Split!
  - Linha 500-999: Acumula, atinge 500, CanSplitAt? Sim → Split!
  - Linha 1000-1499: Acumula, loop termina
Fase 4: Salva último chunk
Output: [Chunk{~500 linhas}, Chunk{~500 linhas}, Chunk{~500 linhas}]
```

### Caso C: Diff com arquivo que excede ApproxLines

```
Input:  Arquivo único com 600 linhas, ApproxLines = 500
Fase 3:
  - Linha 0-499: Atinge 500, CanSplitAt? NÃO (dentro do arquivo, nenhuma fronteira)
  - Linha 500-599: Continua acumulando
Fase 4: Salva chunk único com 600 linhas (excede ApproxLines mas preserva integridade)
Output: [Chunk{600 linhas, FileCount: 2, StartLine: 1}]
```

**Nota importante**: O Caso C demonstra o comportamento "inteligente" — o algoritmo **prioriza a integridade do diff sobre o limite de linhas**. Se não há ponto seguro para split dentro de um arquivo grande, o chunk excede `ApproxLines` sem quebrar o diff.
