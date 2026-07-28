# Domain Concepts

## Scanner (`internal/core/scanner/`)

The scanner performs recursive filesystem traversal and builds a `FileNode` tree.

### FileNode tree

`FileNode` (`internal/core/scanner/scanner.go`) represents a file or directory with:
- `Path`, `RelPath`, `Name`, `IsDir`
- `IsGitignored`, `IsCustomIgnored`, `IsIgnored()` helper
- `Children` (dir entries)
- `Size` and `IsBinary`

### ScanConfig

| Field | Default | Description |
|-------|---------|-------------|
| `MaxFiles` | 10000 | Maximum files to scan |
| `MaxFileSize` | 1MB | Skip files larger than this |
| `MaxMemory` | 500MB | Memory limit for scanning |
| `Workers` | runtime.NumCPU() | Parallel scan workers |
| `IncludeHidden` | false | Include dot-files and dot-dirs |
| `IncludeIgnored` | false | Include gitignored/custom-ignored files |
| `RespectGitignore` | true | Apply `.gitignore` rules |
| `RespectShotgunignore` | true | Apply `.shotgunignore` rules |
| `SkipBinary` | true | Skip binary file content |

Source: `internal/core/scanner/filesystem.go`, `internal/core/scanner/scanner.go`

### Selection helpers (`internal/core/scanner/helpers.go`)

| Function | Purpose |
|----------|---------|
| `CollectSelections(node, selections)` | Recursively collects all non-ignored paths into a selection map |
| `NewSelectAll(root)` | Creates selection map with all non-ignored files |
| `SelectAllExcept(root, deselected)` | Like `NewSelectAll` but omits file paths in the `deselected` list. If `deselected` is nil, equivalent to `NewSelectAll`. |
| `CollectDeselected(root, selections)` | Computes the sorted list of deselected (non-ignored, unselected) file relative paths |

---

## Ignore Engine (`internal/core/ignore/`)

Layered ignore rule engine that evaluates multiple ignore sources in priority order:

1. **Explicit ignore rules** — passed programmatically
2. **Built-in rules** — hardcoded common ignore patterns (e.g., `.git/`, `node_modules/`)
3. **`.shotgunignore`** — project-level custom ignore file (uses `go-gitignore` library)
4. **`.gitignore`** — standard Git ignore rules

Each source can be independently enabled. The engine provides `IsIgnored(path, isDir)` for the scanner and `AddCustomRules(rules)` for dynamic rule injection.

Source: `internal/core/ignore/engine.go`

---

## Context Generation (`internal/core/contextgen/`)

Assembles the final context string from the scanned file tree and selections.

### GenerateConfig

| Field | Purpose |
|-------|---------|
| `MaxTotalSize` | Maximum output size in bytes |
| `TemplateVars` | Variables for template substitution |
| `Template` | Optional template to wrap output |
| `SkipBinary` | Exclude binary file content |
| `IncludeTree` | Include directory tree in output |
| `IncludeSummary` | Include file summary section |

### Generation steps

1. **Tree rendering** — formats the directory tree with selection/ignore state markers
2. **File assembly** — collects selected file contents with path headers
3. **Template substitution** — wraps assembled content in the selected template, interpolating variables
4. **Size enforcement** — truncates or rejects if max size exceeded

Source: `internal/core/contextgen/generator.go`, `internal/core/contextgen/content.go`, `internal/core/contextgen/tree.go`, `internal/core/contextgen/template.go`

---

## Template Management (`internal/core/template/`)

Multi-source template loading and rendering system.

### Sources (in priority order)

1. **Embedded templates** — compiled into binary via `internal/assets/embed.go`
2. **Custom templates directory** — `~/.config/shotgun-cli/templates/`
3. **Repository-local `.shotgun/templates/`** — project-specific templates

### Manager lifecycle

```
Manager.NewManager(embeddedFS, customDir, projectDir)
  → Load() loads all templates from all sources
  → Get(name) retrieves a parsed template
  → Render(tmpl, vars) substitutes variables into template
```

Templates use Go standard library `text/template` syntax with `{{ .VarName }}` substitution. Variables can be provided via CLI flags, config file, or environment variables.

Source: `internal/core/template/manager.go`, `internal/core/template/loader.go`, `internal/core/template/renderer.go`, `internal/core/template/template.go`

---

## Token Estimation (`internal/core/tokens/`)

Rough token estimation for LLM context fit checking. Uses a simple ~4 bytes per token heuristic.

| Function | Purpose |
|----------|---------|
| `EstimateFromBytes(size)` | Estimates token count from content size |
| `ContextFit(size, maxTokens)` | Checks if content fits within token limit |
| `FormatTokens(count)` | Human-readable token count formatting |

Source: `internal/core/tokens/estimator.go`

---

## LLM Provider Interfaces (`internal/core/llm/`)

### Provider interface

```go
type Provider interface {
    Name() string
    IsAvailable() bool
    IsConfigured() bool
    ValidateConfig() error
    Send(ctx context.Context, content string) (*Result, error)
    SendWithProgress(ctx context.Context, content string, progress func(Progress)) (*Result, error)
}
```

### Config

`LLMConfig` struct holds provider name, API key, base URL, model, and timeout. Validation checks for required fields per provider.

### Registry

`Registry` manages a map of provider names to providers. `Create(name)` returns the named provider. Used by `app.DefaultProviderRegistry` which registers OpenAI, Anthropic, and Gemini.

Source: `internal/core/llm/provider.go`, `internal/core/llm/config.go`, `internal/core/llm/registry.go`

---

## Selection Store (`internal/core/selection/`)

Persists per-project file deselection preferences to a single JSON file. This is the most recently added domain package.

### Store

```go
type Store struct{ path string }

func NewStore(path string) *Store
func (s *Store) Load(projectPath string) ([]string, error)
func (s *Store) Save(projectPath string, deselected []string) error
```

### On-disk format

```json
{
  "deselected": {
    "/home/user/proj": ["vendor/pkg/a.go", "dist/bundle.js"],
    "/home/user/other": ["generated/..."],
    ...
  }
}
```

### Write safety

`Save()` writes to a `.tmp` file first, then renames atomically. This prevents data loss on crash.

### Integration points

- `app.WithSelectionStore()` functional option on `DefaultContextService`
- `WizardModel.SetSelectionStore()` attaches to the TUI wizard
- `WizardModel.persistSelections()` called on advancing past file selection step
- Headless CLI: `generateContextHeadless()` passes store to service

Source: `internal/core/selection/store.go`

---

## Diff Splitting (`internal/core/diff/`)

Intelligently splits a git diff into chunks suitable for LLM consumption. Preserves file boundaries and context.

Source: `internal/core/diff/split.go`
