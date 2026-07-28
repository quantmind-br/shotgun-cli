# Repository Guidelines

## Project Overview

**shotgun-cli** is a Go CLI that generates LLM-optimized codebase contexts. It provides both an interactive TUI wizard (Bubble Tea) and headless CLI commands (Cobra) for scanning codebases, assembling context with templates, and sending to LLM providers (OpenAI, Anthropic, Gemini).

**Module**: `github.com/quantmind-br/shotgun-cli` | **Go 1.24** | **Clean Architecture**

## Architecture & Data Flow

### Layer Structure

```
cmd/              → Presentation (CLI commands, composition root)
internal/ui/      → Presentation (TUI wizard, Bubble Tea MVU)
internal/app/     → Application (ContextService, ProviderRegistry)
internal/core/    → Domain (scanner, contextgen, template, ignore, tokens, llm interfaces)
internal/platform/ → Infrastructure (openai, anthropic, geminiapi, http, clipboard)
internal/config/  → Configuration keys, validation, metadata
```

### Dependency Rules (Strictly Enforced)

- `core/` → stdlib ONLY (no external deps, no internal packages)
- `platform/` → core interfaces only
- `app/` → core + platform + config
- `ui/` → app + core + config + Bubble Tea/Lipgloss
- `cmd/` → everything (composition root)

### Data Flow

**TUI Wizard** (no args):
```
main.go → cmd.Execute() → launchTUIWizard()
  → ui.NewWizard() → 5-step state machine
  → Step 1: File selection (ScanCoordinator async scan)
  → Step 2: Template selection
  → Step 3-4: Task/Rules input
  → Step 5: Review & send (GenerateCoordinator async generation)
  → Optional: SendToLLMWithProgress() via ProviderRegistry
```

**Headless CLI** (subcommand):
```
cmd/context.go → Build GenerateConfig → app.NewContextService()
  → svc.Generate(ctx, cfg) [synchronous]
  → Return result → print/save/copy
```

**Core Generation Pipeline**:
```
1. Validate config
2. Scan filesystem (single-pass sequential walk, respect .gitignore/.shotgunignore)
3. Apply selections
4. Generate context (tree rendering + file assembly + template substitution)
5. Estimate tokens (~4 bytes per token)
6. Save output or send to LLM
```

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/` | CLI commands (Cobra), config initialization, composition root |
| `internal/ui/` | TUI wizard, screens, components, coordinators |
| `internal/app/` | ContextService (main API), ProviderRegistry |
| `internal/core/scanner/` | Filesystem traversal, layered ignore rules |
| `internal/core/contextgen/` | Context assembly, tree rendering |
| `internal/core/template/` | Template loading, variable substitution |
| `internal/core/ignore/` | Layered ignore engine (explicit → built-in → gitignore) |
| `internal/core/llm/` | Provider interfaces, Config, Registry |
| `internal/platform/openai/` | OpenAI implementation |
| `internal/platform/anthropic/` | Anthropic implementation |
| `internal/platform/geminiapi/` | Gemini implementation |
| `internal/platform/llmbase/` | Base client with common HTTP logic |
| `internal/platform/http/` | Shared JSON HTTP client |
| `internal/config/` | Config key constants, validation, metadata |
| `test/e2e/` | End-to-end CLI tests |
| `test/fixtures/` | Sample project fixtures |

## Development Commands

```bash
# Build
make build                    # → build/shotgun-cli (current platform)
make build-all                # Cross-compile: linux/darwin/windows × amd64/arm64

# Test
make test                     # Unit tests (go test ./...)
make test-race                # Tests with race detector (preferred)
make test-e2e                 # End-to-end tests (requires build first)
go test -v -run TestFoo ./internal/core/scanner/...  # Single test

# Lint
make lint                     # golangci-lint with .golangci.yml
golangci-lint run ./...       # Direct invocation

# Coverage
make coverage                 # Generate coverage.out + report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# Quality
make fmt                      # go fmt
make vet                      # go vet

# Run
go run . --help               # Run directly
./build/shotgun-cli           # Run built binary

# Install
make install                  # → ~/.local/bin/shotgun-cli
make uninstall                # Remove from ~/.local/bin

# Release
make release-snapshot         # Local test build (goreleaser)
make release VERSION=1.2.3    # Full release (tag + push)
```

## Code Conventions & Common Patterns

### Imports (Three Groups, Blank-Line Separated)

```go
import (
    "context"
    "fmt"
    "path/filepath"

    "github.com/spf13/cobra"
    "github.com/rs/zerolog"

    "github.com/quantmind-br/shotgun-cli/internal/app"
    "github.com/quantmind-br/shotgun-cli/internal/config"
)
```

### Configuration Keys (Never Raw Strings)

```go
// ✅ CORRECT
maxFiles := viper.GetInt(config.KeyScannerMaxFiles)

// ❌ WRONG
maxFiles := viper.GetInt("scanner.max-files")
```

All config keys defined in `internal/config/keys.go` as constants.

### Logging (Zerolog Only)

```go
// ✅ CORRECT
log.Debug().Str("file", path).Msg("Processing file")
log.Error().Err(err).Str("config", name).Msg("Failed to load config")

// ❌ WRONG
fmt.Println("Processing file")
log.Printf("Error: %v", err)
```

### Error Wrapping

```go
// Always wrap with context
absPath, err := filepath.Abs(rootPath)
if err != nil {
    return fmt.Errorf("invalid root path: %w", err)
}
```

### Dependency Injection (Functional Options)

```go
// Service with injectable dependencies
svc := NewContextService(
    WithScanner(mockScanner),
    WithGenerator(mockGenerator),
    WithRegistry(mockRegistry),
)

// Usage in tests
mockScanner := &mockScanner{tree: testTree, err: nil}
svc := NewContextService(WithScanner(mockScanner))
```

### Async Operations (Coordinator Pattern)

```go
// TUI: Non-blocking async with progress
coordinator := NewScanCoordinator(scanner)
cmd := coordinator.Start(rootPath, config)  // Returns tea.Cmd

// Poll for updates in Update() loop
func (m *WizardModel) Update(msg tea.Msg) {
    if coordinator.IsComplete() {
        tree, err := coordinator.Result()  // Mutex-guarded
    }
}
```

### Progress Callbacks

Two styles:
- **Channel-based** (scanner): `progress chan<- Progress`
- **Callback-based** (generators, LLM): `progress func(GenProgress)`

**Critical**: Progress callbacks in TUI path are NOT optional — dropping them stalls the UI.

### Naming Conventions

| Category | Pattern | Example |
|----------|---------|---------|
| Interface | `<Name>` (noun) | `Scanner`, `Provider` |
| Implementation | `<Impl><Name>` or `Default<Name>` | `FilesystemScanner`, `DefaultContextGenerator` |
| Config struct | `<Name>Config` | `ScanConfig`, `GenerateConfig` |
| Message struct | `<Action>Msg` | `ScanProgressMsg`, `ScanCompleteMsg` |
| Test helper | `mock<Interface>` | `mockScanner`, `mockProvider` |
| Model struct | `<Name>Model` | `FileSelectionModel`, `WizardModel` |

### Testing (Table-Driven, Parallel)

```go
func TestScanConfig_Validate(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name    string
        config  *ScanConfig
        wantErr bool
        errMsg  string
    }{
        {"valid config", &ScanConfig{MaxFiles: 100}, false, ""},
        {"negative max files", &ScanConfig{MaxFiles: -1}, true, "must be non-negative"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### LLM Provider Strategy Pattern

```go
// Core interface (internal/core/llm/)
type Provider interface {
    Send(ctx context.Context, content string) (*Result, error)
    SendWithProgress(ctx context.Context, content string, progress func(string)) (*Result, error)
    Name() string
    IsAvailable() bool
}

// Platform implementation (internal/platform/openai/)
type OpenAI struct {
    *llmbase.BaseClient
}

func (c *OpenAI) BuildRequest(content string) (interface{}, error) { ... }
func (c *OpenAI) ParseResponse(resp interface{}, raw []byte) (*llm.Result, error) { ... }
func (c *OpenAI) GetEndpoint() string { return "/chat/completions" }
func (c *OpenAI) GetHeaders() map[string]string { return map[string]string{"Authorization": "Bearer " + c.APIKey} }

// Registry factory (internal/app/providers.go)
provider, err := app.DefaultProviderRegistry.Create(llm.ProviderOpenAI, cfg)
```

### Layered Ignore Engine

Priority (high → low):
1. Explicit excludes (forced ignore)
2. Explicit includes (forced include)
3. Built-in patterns (node_modules, .git, etc.)
4. .gitignore rules
5. .shotgunignore rules
6. Custom patterns

## Important Files

### Entry Points
- `main.go` — Application entry point
- `cmd/root.go` — Root command, TUI launch, config initialization
- `cmd/context.go` — `context generate` command
- `cmd/send.go` — `send` command (send to LLM)

### Core Services
- `internal/app/context.go` — ContextService (main API: Generate, SendToLLM)
- `internal/app/providers.go` — ProviderRegistry factory
- `internal/core/scanner/scanner.go` — Scanner interface + FilesystemScanner
- `internal/core/contextgen/generator.go` — Context generation pipeline
- `internal/core/template/manager.go` — Template loading and rendering
- `internal/core/ignore/engine.go` — Layered ignore engine

### Configuration
- `internal/config/keys.go` — All config key constants
- `internal/config/validator.go` — Validation rules
- `internal/config/metadata.go` — Config descriptions and types

### TUI
- `internal/ui/wizard.go` — Main wizard coordination (5-step state machine)
- `internal/ui/screens/file_selection.go` — File selection screen
- `internal/ui/screens/review.go` — Review & send screen
- `internal/ui/scan_coordinator.go` — Async scan coordination
- `internal/ui/generate_coordinator.go` — Async generation coordination

### Testing
- `test/e2e/cli_test.go` — E2E CLI tests
- `test/fixtures/sample-project/` — Test fixture project
- `internal/app/service_test.go` — ContextService unit tests with mocks

## Runtime/Tooling Preferences

**Required Runtime**: Go 1.24.0+

**Package Manager**: Go modules (`go.mod`, `go.sum`)

**Key Dependencies**:
- CLI: `github.com/spf13/cobra` v1.10.2
- Config: `github.com/spf13/viper` v1.21.0
- TUI: `github.com/charmbracelet/bubbletea` v1.3.5
- Styling: `github.com/charmbracelet/lipgloss` v1.1.0
- Logging: `github.com/rs/zerolog` v1.33.0
- Testing: `github.com/stretchr/testify` v1.11.1

**Build Tools**:
- `make` — Build orchestration
- `golangci-lint` — Linting (v6)
- `goreleaser` — Release automation
- `docker` — Container builds (optional)

**Linting Rules** (`.golangci.yml`):
- Line length: 120 characters
- Cyclomatic complexity: 25 max
- Enabled linters: govet, errcheck, staticcheck, goconst, misspell, gocyclo, gosec, lll, prealloc, unconvert, unparam, unused

## Testing & QA

**Test Framework**: Go `testing` + `stretchr/testify`

**Test Organization**:
- Unit tests: Co-located with source (`*_test.go`)
- Integration tests: `internal/app/integration_test.go`
- E2E tests: `test/e2e/`
- Fixtures: `test/fixtures/sample-project/`

**Running Tests**:
```bash
go test ./...                          # All tests
go test -race ./...                    # With race detector (preferred)
go test -v -run TestName ./pkg         # Specific test
make test-e2e                          # E2E only (requires build first)
```

**Coverage**:
- Target: 85% minimum, 90%+ for new code
- CI threshold: 80% (enforced)
- Generate: `make coverage` or `go test -coverprofile=coverage.out ./...`

**Test Patterns**:
- Table-driven tests with `t.Run()`
- `t.Parallel()` for concurrent execution
- Manual mocks (no code generation)
- `t.TempDir()` for isolation
- `httptest.NewServer()` for HTTP mocking

**Mocking Example**:
```go
type mockScanner struct {
    tree *scanner.FileNode
    err  error
}

func (m *mockScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
    return m.tree, m.err
}

// Usage
mock := &mockScanner{tree: testTree, err: nil}
svc := NewContextService(WithScanner(mock))
```

**CI Skips** (environment limitations):
- `TestScanCoordinator`
- `TestGenerateCoordinator`
- `TestWizardClipboardCopyCmd`

## Anti-Patterns

- ❌ Importing `viper` in `core/` or `platform/` (use DI)
- ❌ Global state in `internal/`
- ❌ Skipping progress callbacks in TUI (breaks UI)
- ❌ Direct HTTP in providers (use `platform/http/JSONClient`)

## OpenWiki

This repository has documentation located in the /openwiki directory.

Start here:
- [OpenWiki quickstart](openwiki/quickstart.md)

OpenWiki includes repository overview, architecture notes, workflows, domain concepts, operations, integrations, testing guidance, and source maps.

When working in this repository, read the OpenWiki quickstart first, then follow its links to the relevant architecture, workflow, domain, operation, and testing notes.
