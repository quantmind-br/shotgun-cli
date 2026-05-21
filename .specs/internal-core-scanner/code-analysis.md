# Análise de Código — internal/core/scanner

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/core/scanner`
> **Tipo:** Domain layer (depende de `internal/core/ignore`)
> **Nível de detalhe:** Detalhado
> **Arquivos analisados:** 5 (scanner.go, filesystem.go, helpers.go, scanner_test.go, helpers_test.go)

---

## 1. Visão Geral

Este módulo implementa a funcionalidade central de **varredura do sistema de arquivos** para a CLI Shotgun. Ele percorre diretórios recursivamente, respeitando regras de `.gitignore`, `.shotgunignore`, padrões personalizados, limites de arquivo/diretório, e constrói uma árvore de `FileNode` para consumo pelas camadas superiores (TUI, geração de contexto, diffs).

O módulo expõe uma **interface `Scanner`** abstraindo a varredura simples (`Scan`) e a varredura com relatório de progresso (`ScanWithProgress`), além de **funções auxiliares** (`CollectSelections`, `NewSelectAll`) para construção de seleções de arquivos a partir da árvore resultante.

A **engine de ignorância** (`internal/core/ignore.IgnoreEngine`) é injetada (via `NewFileSystemScanner`) ou criada internamente, e o scanner depende exclusivamente dela para filtragem por padrões.

---

## 2. Estrutura de Arquivos

### scanner.go (283 linhas, arquivo de tipos e interface)
- **Package:** `scanner`
- **Imports do stdlib:** `fmt`, `time`
- **Tipos públicos:** `Scanner` (interface), `FileNode` (struct), `Progress` (struct), `ScanConfig` (struct)
- **Funções/funções de tipo:** `DefaultScanConfig()`, `FormatSize(int64) string`
- **Métodos de FileNode:** `IsIgnored()`, `GetIgnoreReason()`, `CountChildren()`, `CountFiles()`, `CountDirectories()`, `TotalSize()`
- **Métodos de Progress:** `Percentage()`, `IsStreaming()`, `String()`
- **Constantes:** `StageScanning = "scanning"`, `StageComplete = "complete"`

### filesystem.go (354 linhas, implementação concreta)
- **Package:** `scanner`
- **Imports do stdlib:** `fmt`, `os`, `path/filepath`, `sort`, `strings`, `time`
- **Import do projeto:** `internal/core/ignore`
- **Tipos:** `FileSystemScanner` (struct privada, implementa `Scanner`)
- **Constructors:** `NewFileSystemScanner()`, `NewFileSystemScannerWithIgnore([]string) (*FileSystemScanner, error)`
- **Métodos de Scanner:** `Scan()`, `ScanWithProgress()`
- **Métodos internos de filesystem.go:** `countItems()`, `handleCountError()`, `shouldStopCounting()`, `skipIfDirectory()`, `shouldStopWalking()`, `shouldSkipLargeFile()`, `getFilesize()`, `createFileNode()`, `addNodeToTree()`, `findParentNode()`, `sortChildren()`, `matchesIncludePatterns()`, `shouldIgnore()`, `getIgnoreStatus()`, `getIgnoreStatusWithEngine()`, `classifyIgnoreReason()`, `isHiddenFile()`, `shouldIncludeIgnored()`, `normRel()`
- **Variáveis:** `ErrSkipDir = filepath.SkipDir` (sentinela)
- **Lints:** `//nolint:nilerr` em 3 locais (intencional: continue walking)

### helpers.go (28 linhas, funções auxiliares)
- **Package:** `scanner`
- **Funções públicas:** `CollectSelections(*FileNode, map[string]bool) map[string]bool`, `NewSelectAll(*FileNode) map[string]bool`
- **Sem imports além do stdlib.**
- **Sem dependências externas.**

### scanner_test.go (595 linhas, testes de scanner)
- **Cobertura de teste:** 14 testes principais + 2 benchmarks
- **Testes unitários:**
  - `TestFileNodeBasic` — serialização JSON de FileNode (2 sub-cases)
  - `TestFileNodeMethods` — CountChildren, CountFiles, CountDirectories, TotalSize, IsIgnored (14 sub-cases)
  - `TestIgnoreReasons` — GetIgnoreReason (4 sub-cases)
  - `TestFormatSize` — Formatação de bytes (6 sub-cases)
  - `TestProgress` — Percentage, String (3 sub-cases)
  - `TestDefaultScanConfig` — Validação de defaults (5 assertions)
  - `TestFileSystemScanner` — Teste integrado com dir temporário (5 sub-scenarios: básico, progress, hidden, tamanho, gitignore)
  - `TestNewFileSystemScannerWithIgnore` — Construtor com regras (2 sub-cases)
  - `TestHiddenFileConsistencyWithIgnoreEngine` — 3 cenários de consistência hidden + ignore
  - `TestScannerInterface` — Verificação de implementação da interface
  - `TestTreeSorting` — Ordenação: dirs primeiro, depois arquivos, ambos alfabeticamente
  - `TestShotgunignoreIntegration` — 3 cenários: sem shotgunignore, com shotgunignore, IncludeIgnored
  - `TestIncludePatterns` — 8 cenários de padrões de inclusão
  - `TestIncludePatternsWithIgnoreRules` — Interação include + ignore
  - `TestScannerHandlesPermissionError` — Tolerância a erro de permissão
- **Benchmarks:** `BenchmarkFileSystemScanner` (10 dirs × 50 files = 500 files)

### helpers_test.go (199 linhas, testes de helpers)
- **Testes:**
  - `TestCollectSelections_NilNode` — Retorno nil para nó nulo
  - `TestCollectSelections_NilSelections` — Mapa nil inicializado
  - `TestCollectSelections_SingleFile` — Mapa com 1 entrada
  - `TestCollectSelections_IgnoredFile` — Arquivo gitignored não entra no mapa
  - `TestCollectSelections_CustomIgnoredFile` — Arquivo custom ignorado não entra
  - `TestCollectSelections_DirectoryWithChildren` — Dir com 2 filhos
  - `TestCollectSelections_NestedDirectories` — Árvore 2 níveis
  - `TestCollectSelections_MixedIgnored` — 3 filhos: 1 visível, 1 gitignored, 1 custom-ignored
  - `TestCollectSelections_IgnoredDirectory` — Dir ignorado mas filhos não
  - `TestCollectSelections_PreserveExisting` — Mapa existente preservado
  - `TestNewSelectAll_NilRoot` — Root nulo → mapa vazio
  - `TestNewSelectAll_SingleFile` — 1 arquivo
  - `TestNewSelectAll_DirectoryTree` — Árvore 3 níveis com vendor gitignored
  - `TestNewSelectAll_EmptyDirectory` — Dir vazio
- **Diretório temporário:** Nenhum — usa árvores em memória.

---

## 3. Interface e Tipos Detalhados

### Scanner (interface)
```go
type Scanner interface {
    Scan(rootPath string, config *ScanConfig) (*FileNode, error)
    ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}
```
- **2 métodos públicos** — o contrato mínimo do scanner.
- **`Scan`:** delega para `ScanWithProgress` com `progress = nil`.
- **`ScanWithProgress`:** método real; retorna árvore + canal de progresso (ou nil para silencioso).

### FileNode (struct — nó da árvore de arquivos)
| Campo | Tipo | JSON tag | Descrição |
|---|---|---|---|
| `Name` | `string` | `"name"` | Nome base do arquivo/diretório |
| `Path` | `string` | `"path"` | Caminho absoluto |
| `RelPath` | `string` | `"rel_path"` | Caminho relativo ao root do scan |
| `IsDir` | `bool` | `"is_dir"` | É diretório? |
| `Children` | `[]*FileNode` | `"children,omitempty"` | Filhos (só para dirs) |
| `IsGitignored` | `bool` | `"is_gitignored"` | Ignorado por .gitignore? |
| `IsCustomIgnored` | `bool` | `"is_custom_ignored"` | Ignorado por regras custom? |
| `Size` | `int64` | `"size"` | Tamanho em bytes (0 para dirs) |
| `Expanded` | `bool` | `"expanded"` | Expandido na TUI? |
| `Parent` | `*FileNode` | `"-"` | Referência ao pai (não serializa) |

**Métodos de FileNode:**
| Método | Retorno | Semântica |
|---|---|---|
| `IsIgnored()` | `bool` | `IsGitignored \|\| IsCustomIgnored` |
| `GetIgnoreReason()` | `string` | Texto legível: "gitignored", "custom ignored", "both", "not ignored" |
| `CountChildren()` | `int` | Total de nós filhos recursivamente (só conta de dirs) |
| `CountFiles()` | `int` | Total de arquivos (não-dirs) recursivamente |
| `CountDirectories()` | `int` | Total de dirs recursivamente (inclui self) |
| `TotalSize()` | `int64` | Soma de todos os tamanhos de arquivos abaixo deste nó |

### Progress (struct — relatório de progresso)
| Campo | Tipo | JSON tag | Descrição |
|---|---|---|---|
| `Current` | `int64` | `"current"` | Itens processados |
| `Total` | `int64` | `"total"` | Total de itens. `-1` = modo streaming (desconhecido) |
| `Stage` | `string` | `"stage"` | Etapa atual ("scanning" ou "complete") |
| `Message` | `string` | `"message,omitempty"` | Mensagem contextual |
| `Timestamp` | `time.Time` | `"timestamp"` | Quando o update foi criado |

**Métodos de Progress:**
| Método | Retorno | Semântica |
|---|---|---|
| `Percentage()` | `float64` | `(Current / Total) × 100`. Retorna `-1.0` se streaming, `0.0` se total zero |
| `IsStreaming()` | `bool` | `Total < 0` |
| `String()` | `string` | Formatação legível: `%.1f%% (%d/%d) - stage: message` ou modo streaming |

### ScanConfig (struct — configurações do scan)
| Campo | Tipo | JSON tag | Default | Descrição |
|---|---|---|---|---|
| `MaxFileSize` | `int64` | `"max_file_size"` | `0` (sem limite) | Tamanho máximo de arquivo em bytes |
| `MaxFiles` | `int64` | `"max_files"` | `0` (sem limite) | Número máximo de arquivos |
| `MaxMemory` | `int64` | `"max_memory"` | `0` (sem limite) | Limite de memória 🟡 **INFERIDO: não implementado** |
| `SkipBinary` | `bool` | `"skip_binary"` | `false` | Pular arquivos binários 🟡 **INFERIDO: campo presente mas não verificado** |
| `IncludeHidden` | `bool` | `"include_hidden"` | `false` | Incluir arquivos ocultos (`.`) |
| `Workers` | `int` | `"workers"` | `1` | Threads de scan 🟡 **INFERIDO: valor hardcoded a 1, multithreading não implementado** |
| `IgnorePatterns` | `[]string` | `"ignore_patterns,omitempty"` | `[]` | Padrões custom (semântica gitignore) |
| `IncludePatterns` | `[]string` | `"include_patterns,omitempty"` | `[]` | Padrões de inclusão (glob-style) |
| `IncludeIgnored` | `bool` | `"include_ignored"` | `false` | Incluir ignorados com flags |
| `RespectGitignore` | `bool` | `"respect_gitignore"` | `true` | Carregar .gitignore |
| `RespectShotgunignore` | `bool` | `"respect_shotgunignore"` | `true` | Carregar .shotgunignore |

---

## 4. FileSystemScanner — Construtores

### `NewFileSystemScanner() *FileSystemScanner`
- Cria um scanner com uma nova instância de `ignore.NewIgnoreEngine()`.
- Engine inicializada com **42+ padrões built-in** (node_modules, .git, *.png, etc.).
- **Não recebe parâmetros.**

### `NewFileSystemScannerWithIgnore(ignoreRules []string) (*FileSystemScanner, error)`
- Cria scanner + adiciona regras custom via `ignoreEngine.AddCustomRules()`.
- Se `ignoreRules` não vazio e `AddCustomRules` retorna erro → erro propagado.
- Útil quando se conhece regras de ignorância antecipadamente (ex: `*.tmp`).

---

## 5. Motor de Scan — `ScanWithProgress()`

### Passo 1: Validação e Inicialização
```
ScanWithProgress(rootPath, config, progress)
  │
  ├─ config == nil → config = DefaultScanConfig()
  │
  ├─ os.Stat(rootPath) → valida existência
  │   └─ não é dir → erro
  │
  ├─ config.RespectGitignore == true
  │   └─ ignoreEngine.LoadGitignore(rootPath) → .gitignore + nested
  │
  ├─ config.RespectShotgunignore == true
  │   └─ ignoreEngine.LoadShotgunignore(rootPath) → .shotgunignore → custom rules
  │
  └─ config.IgnorePatterns len > 0
      └─ ignoreEngine.AddCustomRules(config.IgnorePatterns)
```

### Passo 2: Contagem ou Scan Direto (Stream Mode)
O scanner **não faz contagem prévia** — vai direto para `walkAndBuild` com `total = -1` (streaming).
O progresso inicial envia `Total: -1` para indicar modo desconhecido ao consumidor.

### Passo 3: `walkAndBuild()` — Percursos da Árvore
```
filepath.WalkDir(rootPath, callback)
  │
  ├─ Para cada path:
  │   ├─ err != nil → handleWalkError(d)
  │   │   ├─ dir → filepath.SkipDir
  │   │   └─ file → nil (suppress)
  │   │
  │   ├─ shouldStopWalking(config, d, fileCount)
  │   │   └─ MaxFiles > 0 && fileCount >= MaxFiles → filepath.SkipDir
  │   │
  │   ├─ relPath = filepath.Rel(rootPath, path)
  │   │   └─ erro ou "." → nil (continue)
  │   │
  │   ├─ shouldIgnore(relPath, d.IsDir(), config)
  │   │   ├─ matchesIncludePatterns() → false → true (ignore)
  │   │   ├─ ignoreEngine.ShouldIgnore(relPath) → true → !IncludeIgnored → skip
  │   │   └─ !IncludeHidden && startsWith(".") → true (ignore)
  │   │   └─ else → false (não ignora)
  │   │
  │   ├─ getFileSize(d, config) → size + skipFile?
  │   │   ├─ dir → size=0
  │   │   ├─ MaxFileSize > 0 && size > limit → skipFile=true
  │   │   └─ else → size, false
  │   │
  │   ├─ createFileNode(path, relPath, d, size, config)
  │   │   ├─ isGitignored, isCustomIgnored = getIgnoreStatus()
  │   │   └─ FileNode{Name, Path, RelPath, IsDir, IsGitignored, IsCustomIgnored, Size}
  │   │
  │   ├─ addNodeToTree(node, relPath, dirNodes)
  │   │   ├─ parentNode = findParentNode(relPath, dirNodes)
  │   │   ├─ parentNode.Children.append(node)
  │   │   ├─ node.Parent = parentNode
  │   │   └─ if dir: dirNodes[normRel(relPath)] = node
  │   │
  │   ├─ current++ (fileCount++ if not dir)
  │   │
  │   └─ reportProgress(progress, current, total, relPath)
  │       └─ progress != nil && current % 100 == 0 → enviar update
  │
  └─ return root, current, err
```

### Passo 4: Ordenação
```
sortChildren(root)
  │
  ├─ Para cada nó de directory:
  │   ├─ sort.Slice(children, comparator)
  │   │   ├─ dirs antes de files
  │   │   └─ same type → alphabetical (case-insensitive)
  │   └─ Recursivamente: sortChildren(child)
```

### Passo 5: Progresso Final
```
progress <- Progress{Current: actualCount, Total: actualCount, Stage: "complete", ...}
return root, nil
```

---

## 6. Funções Auxiliares — `helpers.go`

### `CollectSelections(node, selections) map[string]bool`
- Recorre a árvero de FileNode.
- Para cada nó **não ignorado** (`!node.IsIgnored()`): adiciona `node.Path` ao mapa.
- Para diretórios: recursivamente chama em todos os filhos.
- **Preserva entradas existentes** no mapa.
- **Nota:** inclui nós de diretório não-ignotos no mapa de seleções.

### `NewSelectAll(root) map[string]bool`
- Conveniência: `CollectSelections(root, make(map[string]bool))`.
- Retorna mapa com **todas** as entradas não-ignotas (arquivos + dirs).

---

## 7. Dependências

### Diretas
| Dependência | Uso |
|---|---|
| `internal/core/ignore` | `IgnoreEngine` — motor de padrões; injetado via constructor |
| `path/filepath` | `WalkDir`, `Rel`, `Base`, `ToSlash`, `Dir` |
| `os` | `Stat`, `DirEntry`, `FileInfo` |
| `sort` | `Slice` para ordenação de children |
| `strings` | `ToLower`, `HasPrefix` |
| `fmt` | `Errorf`, `Sprintf` |
| `time` | `Timestamp` em Progress |

### Indiretas (via `internal/core/ignore`)
| Dependência | Uso |
|---|---|
| `github.com/sabhiram/go-gitignore` | Matchers de padrões gitignore |

### Consumidores do módulo
| Consumidor | Uso |
|---|---|
| `internal/core/contextgen` | `Scan()`, `ScanWithProgress()` — gera contexto de arquivos |
| `internal/core/diff` | `Scan()` — compara estruturas de árvore |
| `internal/ui` | Progress channel — TUI spinner |

---

## 8. Observações de Qualidade

### ✅ Pontos fortes
- **Interface limpa:** `Scanner` abstrai progresso sem expor complexidade.
- **Arquitetura de tree em map:** `dirNodes map[string]*FileNode` com `normRel()` para lookup O(1) de pais.
- **Progress throttling:** updates a cada 100 itens — evita flooding de canal.
- **Streaming mode:** `Total = -1` sinaliza progresso indeterminado para UI.
- **Hierarquia de ignores:** engine de 5 camadas com `classifyIgnoreReason()` mapeando para `IsGitignored`/`IsCustomIgnored`.
- **Ordenação consistente:** dirs primeiro, depois arquivos, ambos case-insensitive.
- **Tolerância a erros:** `handleWalkError` — dirs pulados, arquivos suprimidos, scan continua.
- **Benchmarks presentes:** `BenchmarkFileSystemScanner` com 500 files.

### ⚠️ Riscos identificados
1. **`MaxMemory` não implementado:** Campo existe em `ScanConfig` mas nunca é verificado em nenhum método.
2. **`SkipBinary` não implementado:** Campo existe mas nunca é consultado em `shouldIgnore` ou `getFileSize`.
3. **`Workers = 1` hardcoded:** Multithreading configurável mas funcionalidade de workers não implementada — scan é inteiramente sequencial (`filepath.WalkDir`).
4. **`findParentNode` tem fallback O(n):** Se o parent não existe no map (caso raro), faz loop `strings.Split` + busca linear — funciona mas não é O(1).
5. **`ShouldIgnore` reavalia include patterns:** Para cada arquivo, `matchesIncludePatterns` é chamado — O(m) onde m = número de padrões de inclusão.
6. **Árvore completa em memória:** `walkAndBuild` constrói TODA a árvore antes de retornar — para projetos gigantes (100k+ files), consumo de memória é linear.
7. **`countItems` duplica lógica:** Existe `countItems()` mas não é usado no caminho principal (o scan usa `walkAndBuild` direto). Pode ser legado ou fallback.
8. **`normRel()` usa `filepath.ToSlash()`:** Funciona mas `filepath.ToSlash` é call de stdlib — não há otimização de string manual.
9. **`ErrSkipDir` é exportado desnecessariamente:** `var ErrSkipDir = filepath.SkipDir` expõe uma constante que é do stdlib — não agrega valor.

### 🔍 Nulos/Inferidos
- 🟡 **INFERIDO:** `MaxMemory` em `ScanConfig` — campo presente mas sem implementação. Provavelmente planejado para limites de memória (GC pressure, max nodes).
- 🟡 **INFERIDO:** `SkipBinary` em `ScanConfig` — campo presente mas sem verificação de tipo binário. Provavelmente planeado para heurística de magic bytes.
- 🟡 **INFERIDO:** `Workers` em `ScanConfig` — valor 1 hardcoded, sem canal de workers ou goroutines paralelas.
- 🟡 **INFERIDO:** Licença do módulo (inferido: MIT, padrão Go).
