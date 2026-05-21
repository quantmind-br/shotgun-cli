# Análise de Código — internal/core/ignore

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/core/ignore`
> **Tipo:** Domain layer (zero dependências internas)
> **Nível de detalhe:** Detalhado
> **Arquivos analisados:** 2 (engine.go, engine_test.go)

---

## 1. Visão Geral

Este módulo implementa um **motor de ignição em camadas (layered ignore engine)** para filtragem de arquivos e diretórios durante o scan do sistema de arquivos. As regras são organizadas em 5 camadas com prioridades bem definidas:

1. **Explicit Excludes** (exclusões explícitas) — maior prioridade
2. **Explicit Includes** (inclusões explícitas) — sobrescreve tudo exceto excludes explícitos
3. **Built-in** (padrões internos embutidos)
4. **Gitignore** (regras de `.gitignore`)
5. **Custom** (padrões personalizados) — menor prioridade

O motor depende da biblioteca externa `github.com/sabhiram/go-gitignore` para parsing e matching de padrões estilo `.gitignore`.

---

## 2. Estrutura de Arquivos

### engine.go (326 linhas, 1 arquivo de código)
- **Package:** `ignore`
- **Imports do stdlib:** `fmt`, `os`, `path/filepath`, `strings`
- **Import de terceiros:** `github.com/sabhiram/go-gitignore`
- **Tipos públicos:** `IgnoreEngine` (interface), `LayeredIgnoreEngine` (struct), `IgnoreReason` (enum)
- **Funções/Constructors públicos:** `NewIgnoreEngine()`, `Reason` (alias deprecated)
- **Métodos da interface:** 10 métodos
- **Lints:** `nolint:revive` no alias `Reason`, `nolint:gosec` no `os.ReadFile` (path controlado)

### engine_test.go (456 linhas)
- **Cobertura de teste:** 12 testes + 2 benchmarks
- **Testes unitários:**
  - `TestIgnoreReason_String` — 6 sub-cases
  - `TestNewIgnoreEngine` — validação de inicialização
  - `TestLayeredIgnoreEngine_BuiltInPatterns` — 15 sub-cases (ignores + não-ignores)
  - `TestLayeredIgnoreEngine_LoadGitignore` — 3 cenários (ausência, válido, vazio)
  - `TestLayeredIgnoreEngine_CustomRules` — 3 cenários (individual, múltiplas, vazias)
  - `TestLayeredIgnoreEngine_ExplicitRules` — 3 cenários (exclude, include override, vazias)
  - `TestLayeredIgnoreEngine_RulePrecedence` — 5 sub-cases de prioridade
  - `TestLayeredIgnoreEngine_IsGitignored` — 4 sub-cases
  - `TestLayeredIgnoreEngine_IsCustomIgnored` — 4 sub-cases
  - `TestLayeredIgnoreEngine_PathNormalization` — 2 sub-cases (cross-platform)
  - `TestLayeredIgnoreEngine_LoadShotgunignore` — 4 cenários (ausência, válido, nested, vazio)
- **Benchmarks:** `BenchmarkLayeredIgnoreEngine_ShouldIgnore`, `BenchmarkLayeredIgnoreEngine_NewIgnoreEngine`
- **Diretório temporário:** `os.MkdirTemp("", "ignore_test")` / `os.MkdirTemp("", "shotgunignore_test")`

---

## 3. Tipos e Interfaces Detalhados

### IgnoreReason (enum)
```
type IgnoreReason int

Constantes:
  IgnoreReasonNone        → 0 → "none"
  IgnoreReasonBuiltIn     → 1 → "built-in"
  IgnoreReasonGitignore   → 2 → "gitignore"
  IgnoreReasonCustom      → 3 → "custom"
  IgnoreReasonExplicit    → 4 → "explicit"
```
- **String() method:** switch-case com fallback "unknown" para valores fora do range.
- **Reason alias:** `Reason = IgnoreReason` mantido por compatibilidade retroativa com `//nolint:revive`.

### IgnoreEngine (interface)
```go
type IgnoreEngine interface {
    ShouldIgnore(relPath string) (bool, IgnoreReason)
    LoadGitignore(rootDir string) error
    AddCustomRule(pattern string) error
    AddCustomRules(patterns []string) error
    AddExplicitExclude(pattern string) error
    AddExplicitInclude(pattern string) error
    IsGitignored(relPath string) bool
    IsCustomIgnored(relPath string) bool
    LoadShotgunignore(rootDir string) error
}
```
- **9 métodos públicos** — contrato completo de engine de ignorância.
- **`ShouldIgnore`:** único método que retorna par `(bool, IgnoreReason)`.

### LayeredIgnoreEngine (struct)
```go
type LayeredIgnoreEngine struct {
    builtInMatcher   *gitignore.GitIgnore
    gitignoreMatcher *gitignore.GitIgnore
    customMatcher    *gitignore.GitIgnore
    explicitExcludes *gitignore.GitIgnore
    explicitIncludes *gitignore.GitIgnore

    customPatterns          []string
    explicitExcludePatterns []string
    explicitIncludePatterns []string
}
```
- **5 matchers** do tipo `*gitignore.GitIgnore` (do pacote `sabhiram/go-gitignore`).
- **3 slices de rastreamento** para acúmulo de padrões entre chamadas.
- **Não é thread-safe** — não há mutex ou lock interno.

---

## 4. Construtor — `NewIgnoreEngine()`

O construtor inicializa o motor com:

1. **Lista de 42+ padrões built-in** em `builtInPatterns`, agrupados em:
   - Shotgun-specific: `shotgun-prompt*.md`
   - VCS: `.git/`, `.svn/`, `.hg/`, `.bzr/`
   - IDE/Editor: `.vscode/`, `.idea/`, `*.swp`, `*.swo`, `*~`, `.DS_Store`, `Thumbs.db`
   - Build/Dependencies: `node_modules/`, `vendor/`, `target/`, `build/`, `dist/`, etc.
   - Cache/Temp: `__pycache__/`, `*.pyc`, `*.pyo`, `.cache/`, `*.pytest_cache/`, etc.
   - Imagens/Mídia: `*.png`, `*.jpg`, `*.mp3`, `*.mp4`, etc.
   - Fontes: `*.ttf`, `*.otf`, etc.
   - Documentos: `*.pdf`, `*.doc*`, `*.xls*`, `*.ppt*`
   - Binários: `*.exe`, `*.dll`, `*.so`, `*.dylib`
   - Bancos: `*.sqlite`, `*.db`
   - Logs: `*.log`, `logs/`
   - Pacotes: `*.jar`, `*.war`, `*.zip`, `*.tar.gz`, etc.
   - OS files: `.DS_Store?`, `._*`, `.Spotlight-V100`, etc.

2. **5 matchers vazios** compilados com `gitignore.CompileIgnoreLines()`.

---

## 5. Motor de Prioridade — `ShouldIgnore()`

```
ShouldIgnore(relPath)
  │
  ├─→ filepath.ToSlash(relPath)        [normaliza separators]
  │
  ├─→ 1. explicitExcludes.MatchesPath()  ──→ (true, IgnoreReasonExplicit)
  │
  ├─→ 2. explicitIncludes.MatchesPath()  ──→ (false, IgnoreReasonNone)       ← OVERRIDE
  │
  ├─→ 3. builtInMatcher.MatchesPath()    ──→ (true, IgnoreReasonBuiltIn)
  │
  ├─→ 4. gitignoreMatcher.MatchesPath()  ──→ (true, IgnoreReasonGitignore)
  │
  ├─→ 5. customMatcher.MatchesPath()     ──→ (true, IgnoreReasonCustom)
  │
  └─→ (false, IgnoreReasonNone)
```

**Regra de ouro:** `explicitIncludes` retorna `(false, None)` mesmo que built-in/gitignore/custom digam ignore — é o único **override positivo** do sistema.

---

## 6. Carregamento de Regras Externas

### `LoadGitignore(rootDir string)`
1. `filepath.Walk(rootDir)` — encontra **todos** os `.gitignore` na árvore.
2. Para cada arquivo encontrado:
   - `os.ReadFile` — lê conteúdo.
   - `filepath.Rel` — calcula caminho relativo para prefixar padrões em arquivos aninhados.
   - Split por `\n`, filter de `#` comments e linhas vazias.
   - **Padrões de negação** (`!...`) preservados com prefixo ajustado.
   - Arquivos aninhados: padrões são prefixados com `filepath.Join(relDir, pattern)`.
3. Todos os padrões agregados → `gitignore.CompileIgnoreLines(allPatterns...)`.
4. Sem arquivos encontrados → matcher vazio retornado (sem erro).

### `LoadShotgunignore(rootDir string)`
- **Mesma lógica** que `LoadGitignore`, mas para `.shotgunignore`.
- Padrões carregados são **adicionados como custom rules** via `AddCustomRules()`.
- Não retorna erro quando nenhum arquivo é encontrado (diferente de `LoadGitignore` que retorna matcher vazio).

### `AddCustomRule(pattern)` / `AddCustomRules(patterns)`
- **Acúmulo:** padrões são acumulados em `e.customPatterns`.
- **Recompilação:** após cada adição, `customMatcher` é recompilado com todos os padrões acumulados.
- Filtra padrões vazios/whitespace em `AddCustomRules`.

### `AddExplicitExclude(pattern)` / `AddExplicitInclude(pattern)`
- Mesma estratégia de acúmulo + recompilação.

---

## 7. Dependências

### Diretas
| Dependência | Uso |
|---|---|
| `github.com/sabhiram/go-gitignore` | 5 matchers de padrões estilo .gitignore |
| `path/filepath` | Normalização de path, `ToSlash`, `Rel`, `Join`, `Walk` |
| `strings` | `Split`, `TrimSpace`, `HasPrefix` |
| `os` | `ReadFile`, `FileInfo` (via Walk callback) |
| `fmt` | `Errorf` para wrapping de erros |

### Consumidores do módulo
| Consumidor | Uso |
|---|---|
| `internal/core/scanner/filesystem.go` | `FileSystemScanner.ignoreEngine` — engine injetada; usa `ShouldIgnore()`, `IsCustomIgnored()`, `IsGitignored()`, `LoadGitignore()`, `LoadShotgunignore()`, `AddCustomRules()` |
| `internal/core/scanner/scanner_test.go` | Tests de `classifyIgnoreReason` mapeando `IgnoreReason` para booleanos de exclusão |

---

## 8. Observações de Qualidade

### ✅ Pontos fortes
- **Interface limpa:** `IgnoreEngine` é um contrato mínimo e expressivo.
- **Testes robustos:** 12 testes cobrindo todos os caminhos, incluindo cenários edge (arquivos ausentes, vazios, nested, negação).
- **Benchmarks:** cobertura de performance para `ShouldIgnore` e `NewIgnoreEngine`.
- **Prioridade explícita:** ordem de avaliação documentada no código e nos tests.
- **Normalização cross-platform:** `filepath.ToSlash` garante consistência entre Windows/macOS/Linux.
- **Degeneração segura:** vazios e nulos resultam em matcher vazio (não panica).

### ⚠️ Riscos identificados
1. **Não thread-safe:** `LayeredIgnoreEngine` não possui sincronização. Se múltiplas goroutines chamarem `AddCustomRules` simultaneamente, haverá data race.
2. **Recompilação redundante:** `AddCustomRule` recompila o matcher a cada chamada individual. Em cenários com muitas adições individuais, seria mais eficiente batchar.
3. **Padrões built-in hardcoded:** 42+ padrões fixos no código — alteração requer recompilação do binário.
4. **Negation no gitignore:** O `go-gitignore` suporta `!negation`, mas a prioridade com `explicitIncludes` pode criar comportamentos surpreendentes se o usuário usar `!` em `.gitignore` para um path também coberto por `explicitInclude`.
5. **Walk recursivo de `.gitignore`:** `filepath.Walk` é síncrono e sequencial — para projetos enormes com milhares de arquivos, pode ser lentos.
6. **Alias deprecated:** `Reason = IgnoreReason` é um alias de tipo que gera confusão de naming (`ignore.Reason` vs `ignore.IgnoreReason`).

### 🔍 Nulos/Inferidos
- Nenhum gap crítico identificado. A implementação é coesa e bem testada.
