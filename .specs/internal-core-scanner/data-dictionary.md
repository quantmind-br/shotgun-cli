# Dicionário de Dados — internal/core/scanner

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/core/scanner`
> **Nível de detalhe:** Detalhado

---

## 1. Interface — Scanner

| Método | Parâmetros | Retorno | Descrição |
|---|---|---|---|
| `Scan` | `rootPath string`, `config *ScanConfig` | `(*FileNode, error)` | Varredura síncrona simples do sistema de arquivos. |
| `ScanWithProgress` | `rootPath string`, `config *ScanConfig`, `progress chan<- Progress` | `(*FileNode, error)` | Varredura com canal de progresso. `progress = nil` = silencioso. |

---

## 2. Tipos de Dados

### FileNode
| Campo | Tipo | JSON tag | Descrição |
|---|---|---|---|
| `Name` | `string` | `"name"` | Nome base (basename) do arquivo ou diretório |
| `Path` | `string` | `"path"` | Caminho absoluto no sistema de arquivos |
| `RelPath` | `string` | `"rel_path"` | Caminho relativo ao root do scan (ex: `"src/main.go"`) |
| `IsDir` | `bool` | `"is_dir"` | Verdadeiro se o nó representa um diretório |
| `Children` | `[]*FileNode` | `"children,omitempty"` | Lista de filhos (só preenchida se `IsDir == true`) |
| `IsGitignored` | `bool` | `"is_gitignored"` | Verdadeiro se o caminho corresponde a regra `.gitignore` |
| `IsCustomIgnored` | `bool` | `"is_custom_ignored"` | Verdadeiro se o caminho corresponde a `.shotgunignore`, custom patterns, ou built-in |
| `Size` | `int64` | `"size"` | Tamanho em bytes (`0` para diretórios) |
| `Expanded` | `bool` | `"expanded"` | Estado de expansão na TUI |
| `Parent` | `*FileNode` | `"-"` | Referência ao pai (excluído da serialização JSON) |

**Métodos de FileNode:**

| Método | Retorno | Semântica |
|---|---|---|
| `IsIgnored()` | `bool` | `f.IsGitignored \|\| f.IsCustomIgnored` |
| `GetIgnoreReason()` | `string` | `"gitignored"`, `"custom ignored"`, `"gitignored and custom ignored"`, `"not ignored"` |
| `CountChildren()` | `int` | Total de todos os descendentes (filhos + netos + ...) — só para dirs |
| `CountFiles()` | `int` | Total de arquivos (nós não-dir) recursivamente |
| `CountDirectories()` | `int` | Total de diretórios recursivamente (inclui self) |
| `TotalSize()` | `int64` | Soma dos `Size` de todos os arquivos descendentes |

### Progress
| Campo | Tipo | JSON tag | Descrição |
|---|---|---|---|
| `Current` | `int64` | `"current"` | Quantidade de itens processados até agora |
| `Total` | `int64` | `"total"` | Total de itens. `-1` = streaming (desconhecido). `0` = zero items |
| `Stage` | `string` | `"stage"` | `"scanning"` ou `"complete"` |
| `Message` | `string` | `"message,omitempty"` | Contexto adicional (ex: `"Processing: src/main.go"`) |
| `Timestamp` | `time.Time` | `"timestamp"` | Instante em que o update foi gerado |

**Métodos de Progress:**

| Método | Retorno | Semântica |
|---|---|---|
| `Percentage()` | `float64` | `(Current/Total) × 100`. `-1.0` se streaming. `0.0` se total zero. |
| `IsStreaming()` | `bool` | `Total < 0` |
| `String()` | `string` | `%.1f%% (%d/%d) - stage: message` ou modo streaming `"%d items - stage: message"` |

### ScanConfig
| Campo | Tipo | JSON tag | Default | Semântica |
|---|---|---|---|---|
| `MaxFileSize` | `int64` | `"max_file_size"` | `0` | Limite de tamanho de arquivo em bytes. `0` = sem limite. |
| `MaxFiles` | `int64` | `"max_files"` | `0` | Máximo de arquivos a escanear. `0` = sem limite. |
| `MaxMemory` | `int64` | `"max_memory"` | `0` | Limite de memória. 🟡 **INFERIDO: não implementado** |
| `SkipBinary` | `bool` | `"skip_binary"` | `false` | Pular arquivos binários. 🟡 **INFERIDO: não implementado** |
| `IncludeHidden` | `bool` | `"include_hidden"` | `false` | Incluir arquivos/dirs que começam com `.` |
| `Workers` | `int` | `"workers"` | `1` | Nº de workers paralelos. 🟡 **INFERIDO: valor 1 hardcoded** |
| `IgnorePatterns` | `[]string` | `"ignore_patterns,omitempty"` | `[]` | Padrões custom estilo gitignore |
| `IncludePatterns` | `[]string` | `"include_patterns,omitempty"` | `[]` | Padrões de inclusão glob-style. Se vazio = tudo incluído. |
| `IncludeIgnored` | `bool` | `"include_ignored"` | `false` | Incluir ignorados com flags `IsGitignored`/`IsCustomIgnored` |
| `RespectGitignore` | `bool` | `"respect_gitignore"` | `true` | Carregar e respeitar `.gitignore` |
| `RespectShotgunignore` | `bool` | `"respect_shotgunignore"` | `true` | Carregar e respeitar `.shotgunignore` |

### IgnoreReason (do pacote `internal/core/ignore`)
| Valor | Int | String() | Semântica |
|---|---|---|---|
| `IgnoreReasonNone` | `0` | `"none"` | Caminho não ignorado |
| `IgnoreReasonBuiltIn` | `1` | `"built-in"` | Ignorado por padrões embutidos (node_modules, .git, *.png, etc.) |
| `IgnoreReasonGitignore` | `2` | `"gitignore"` | Ignorado por `.gitignore` |
| `IgnoreReasonCustom` | `3` | `"custom"` | Ignorado por `.shotgunignore` ou `IgnorePatterns` |
| `IgnoreReasonExplicit` | `4` | `"explicit"` | Ignorado por `AddExplicitExclude` |

---

## 3. Constantes

### StageConstants
| Constante | Valor | Uso |
|---|---|---|
| `StageScanning` | `"scanning"` | Stage do Progress durante varredura |
| `StageComplete` | `"complete"` | Stage do Progress após conclusão |

### Sentinel
| Variável | Valor | Uso |
|---|---|---|
| `ErrSkipDir` | `filepath.SkipDir` | Sentinel re-exportado — não agrega valor. |

---

## 4. Funções de Utilidade

### `DefaultScanConfig() *ScanConfig`
Retorna config com:
| Campo | Valor |
|---|---|
| `MaxFileSize` | `0` |
| `MaxFiles` | `0` |
| `MaxMemory` | `0` |
| `SkipBinary` | `false` |
| `IncludeHidden` | `false` |
| `Workers` | `1` |
| `RespectGitignore` | `true` |
| `RespectShotgunignore` | `true` |

### `FormatSize(bytes int64) string`
Formatação human-readável de bytes:
| Input | Output |
|---|---|
| `0` | `"0 B"` |
| `512` | `"512 B"` |
| `1536` | `"1.5 KB"` |
| `2097152` | `"2.0 MB"` |
| `3221225472` | `"3.0 GB"` |
| `1099511627776` | `"1.0 TB"` |

Escalas: `B → KB → MB → GB → TB → PB → EB` (base 1024, 1 casa decimal).

---

## 5. Funções Auxiliares

### `CollectSelections(node *FileNode, selections map[string]bool) map[string]bool`
- Recorre árvore de FileNode.
- Para cada nó **não ignorado**: adiciona `node.Path` ao mapa.
- Preserva entradas existentes.
- **Inclui diretórios** não-ignotos no mapa de seleções.
- `nil node` → retorna `nil` (não inicializa mapa).

### `NewSelectAll(root *FileNode) map[string]bool`
- Conveniência: `CollectSelections(root, make(map[string]bool))`.
- Sempre retorna mapa não-nil (mesmo para root nulo).

---

## 6. Estrutura da Árvore FileNode

A árvore resultante de um scan tem a seguinte estrutura:

```
FileNode (root) — IsDir=true, Expanded=true, Children=[...]
  │
  ├─ FileNode ("src", IsDir=true, Children=[main.go, handler.go])
  │   ├─ FileNode ("main.go", IsDir=false, Size=1234, Path="/project/src/main.go")
  │   └─ FileNode ("handler.go", IsDir=false, Size=5678, Path="/project/src/handler.go")
  │
  ├─ FileNode ("README.md", IsDir=false, Size=999)
  │
  ├─ FileNode ("node_modules", IsDir=true, IsGitignored=true, Children=[...])
  │   └─ (filhos presentes se IncludeIgnored=true, omitidos se IncludeIgnored=false)
  │
  └─ FileNode (".hidden", IsDir=false, IsCustomIgnored=true)
```

**Ordenação de children:**
1. Diretóros primeiro (ordem alfabética, case-insensitive)
2. Arquivos depois (ordem alfabética, case-insensitive)

---

## 7. Fluxo de Dados — Configuração do Scanner

```
NewFileSystemScanner() / NewFileSystemScannerWithIgnore(rules)
  │
  ├─ NewIgnoreEngine()
  │   ├─ builtInMatcher ← 42+ patterns
  │   └─ 4 matchers vazios
  │
  └─ [Scanner pronto — engine com built-in ativo]
  │
  ▼
ScanWithProgress(rootPath, config, progress)
  │
  ├─ config == nil → DefaultScanConfig()
  ├─ RespectGitignore → LoadGitignore(rootPath)
  ├─ RespectShotgunignore → LoadShotgunignore(rootPath)
  ├─ IgnorePatterns len > 0 → AddCustomRules(…)
  │
  ▼
walkAndBuild(rootPath, config, progress, total=-1)
  │
  ├─ filepath.WalkDir(rootPath, callback)
  │   ├─ Para cada path:
  │   │   ├─ shouldStopWalking(config, fileCount)
  │   │   ├─ shouldIgnore(relPath, isDir, config)
  │   │   │   ├─ matchesIncludePatterns(relPath, isDir, config) → false → ignorar
  │   │   │   ├─ ignoreEngine.ShouldIgnore(relPath) → true → !IncludeIgnored → ignorar
  │   │   │   └─ !IncludeHidden && startsWith(".") → ignorar
  │   │   ├─ getFileSize(d, config) → size, skip?
  │   │   ├─ createFileNode(…) → FileNode
  │   │   ├─ addNodeToTree(node, relPath, dirNodes)
  │   │   └─ reportProgress(progress, current, total, relPath) [cada 100 items]
  │   └─ return root, actualCount
  │
  ├─ sortChildren(root)
  │   └─ dirs first + alphabetical (case-insensitive), recursivo
  │
  └─ progress <- {Current: actualCount, Total: actualCount, Stage: "complete"}
  │
  ▼
return root, nil
```

---

## 8. Metadados do Módulo

| Propriedade | Valor |
|---|---|
| Package path | `github.com/quantmind-br/shotgun-cli/internal/core/scanner` |
| Go version | 1.24.0 |
| Licença | 🟡 INFERIDO: MIT (padrão Go) |
| Go doc level | `detalhado` |
| Arquivos de código | `scanner.go` (283 linhas), `filesystem.go` (354 linhas), `helpers.go` (28 linhas) |
| Arquivos de teste | `scanner_test.go` (595 linhas), `helpers_test.go` (199 linhas) |
| Dependências diretas | `internal/core/ignore` |
| Dependências indiretas | `github.com/sabhiram/go-gitignore` |
| Interface stability | `Scanner` — 2 métodos — API estável |
| Thread-safe | **Não** — nenhum lock interno |
| Workers implementados | **Não** — `Workers` campo existe mas valor 1 hardcoded |
