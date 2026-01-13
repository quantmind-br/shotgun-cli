# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details  
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

## Build/Test/Lint Commands

```bash
# Build
make build                    # Build for current platform
go build -o build/shotgun-cli .

# Test
go test ./...                 # Run all tests
go test -race ./...           # With race detector (recommended)
go test -v ./internal/core/scanner/...  # Single package
go test -v -run TestMyFunc ./path/to/pkg  # Single test by name
go test -coverprofile=coverage.out ./...  # With coverage

# Lint
golangci-lint run ./...       # Run linter (uses .golangci-local.yml)
make lint                     # Same via Makefile

# Coverage
go tool cover -func=coverage.out | grep total  # Check total %
go tool cover -html=coverage.out -o coverage.html  # HTML report
```

**Coverage Requirement**: 85% minimum (enforced by CI). Target 90%+ for new code.

## Code Style Guidelines

### Imports
Standard Go import grouping (stdlib, external, internal):
```go
import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    
    "github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)
```

### Error Handling
Always wrap errors with context using `%w`:
```go
if err != nil {
    return nil, fmt.Errorf("failed to scan directory %s: %w", rootPath, err)
}
```

### Naming Conventions
- **Packages**: lowercase, single word (`scanner`, `template`, `llm`)
- **Interfaces**: verb-like or -er suffix (`Scanner`, `Provider`, `ContextGenerator`)
- **Test functions**: `TestFunctionName_Scenario`
- **Config keys**: Use constants from `internal/config/keys.go`, not strings

### Testing Patterns
Use table-driven tests with `t.Parallel()` and `stretchr/testify`:
```go
func TestMyFunction_ValidInput(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {name: "basic case", input: "test", expected: "result"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            result, err := MyFunction(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Logging
Use `zerolog` for structured logging. Never use `fmt.Println` or stdlib `log`:
```go
log.Debug().Str("path", path).Msg("scanning directory")
```

## Architecture Rules

**Clean Architecture layers** - respect boundaries:
```
cmd/                    → Presentation (CLI commands, composition root)
internal/ui/            → Presentation (TUI wizard, Bubble Tea MVU)
internal/app/           → Application (orchestration, ContextService)
internal/core/          → Domain (scanner, template, llm, ignore, tokens)
internal/platform/      → Infrastructure (geminiweb, http, clipboard, openai, anthropic)
```

**Key rules**:
- `internal/core` must NOT import from `cmd`, `ui`, `app`, or `platform`
- Pass config via dependency injection, not global Viper access in core
- Use interfaces from core packages (`Scanner`, `Provider`, `ContextGenerator`)
- Create LLM providers via `app.DefaultProviderRegistry.Create()`
- Use `app.CLIConfig` from `internal/app/config.go` for CLI flag parsing
- Use `app.GenerateConfig` from `internal/app/context.go` for service layer configuration

### Adding a New LLM Provider
1. Implement `llm.Provider` interface in `internal/platform/<provider>/`
2. Register in `internal/app/providers.go`

### Shared HTTP Client
Most providers use the shared `JSONClient` in `internal/platform/http/` for standardized API calls.

**Usage Example:**
```go
var response TargetStruct
err := c.jsonClient.PostJSON(ctx, "/endpoint", headers, requestBody, &response)
if err != nil {
    return nil, c.handleError(err) // Map platformhttp.HTTPError to provider error
}
```

> **Note**: The browser-based Gemini integration is located in the `geminiweb` package.

## Linting Configuration
**Max line length**: 120 chars | **Max cyclomatic complexity**: 25  
**Enabled**: govet, errcheck, staticcheck, goconst, misspell, gocyclo, gosec, lll, prealloc, unconvert, unparam, unused

## Session Completion

**Work is NOT complete until `git push` succeeds.**

1. File issues for remaining work: `bd create "..." --description="..." -p 2`
2. Run quality gates: `go test -race ./... && golangci-lint run`
3. Update issue status: `bd close <id>` or `bd update <id> --status ...`
4. **PUSH TO REMOTE** (MANDATORY):
   ```bash
   git pull --rebase && bd sync && git push
   git status  # MUST show "up to date with origin"
   ```

## Issue Tracking with bd (beads)

**Use bd for ALL task tracking. Do NOT use markdown TODOs or task lists.**

```bash
bd ready --json                          # Check ready work
bd create "Title" -t task -p 2 --json    # Create issue
bd update bd-42 --status in_progress     # Claim work
bd close bd-42 --reason "Completed"      # Complete work
```

**Types**: `bug`, `feature`, `task`, `epic`, `chore`  
**Priorities**: `0`=Critical, `1`=High, `2`=Medium (default), `3`=Low, `4`=Backlog

## Key Application Services

### ContextService (`internal/app/context.go`)
- `Generate(ctx, cfg)` - Synchronous (CLI)
- `GenerateWithProgress(ctx, cfg, callback)` - Async with progress (TUI)
- `SendToLLM(ctx, content, provider)` - Send to LLM provider
- `SendToLLMWithProgress(ctx, content, config, callback)` - Send to LLM with progress reporting (TUI)

```go
// Using ContextService for LLM operations
svc := app.NewContextService()
result, err := svc.SendToLLMWithProgress(ctx, content, app.LLMSendConfig{
    Provider:     llm.ProviderOpenAI,
    APIKey:       "your-api-key",
    Model:        "gpt-4o",
    SaveResponse: true,
    OutputPath:   "./response.md",
}, func(stage string) {
    fmt.Printf("Progress: %s\n", stage)
})
```

### ProviderRegistry (`internal/app/providers.go`)
```go
provider, err := app.DefaultProviderRegistry.Create(llm.ProviderOpenAI, cfg)
```

### Core Utilities
- `diff.IntelligentSplit` - Split large diffs preserving file boundaries
- `scanner.CollectSelections` - Handle file tree selections
- `config.ValidateValue(key, value)` - Validate config before saving

## TUI Wizard

### Keyboard Navigation
- Standard shell shortcuts supported: Ctrl+N (next), Ctrl+P (previous)
- Function keys still work: F7 (back), F8 (next)
- Help screen: F1

### File Selection Screen
- Filter mode (`/`) shows match count in stats bar as "X/Y files"
- Animated spinner during directory scan
- Shows progress with file count during scanning

### Terminal Size Handling
- Minimum size: 40x10 (columns x rows)
- Warning overlay displayed for small terminals
- Check `minTerminalWidth` and `minTerminalHeight` constants in `internal/ui/wizard.go`

### Testing TUI Changes
```bash
# Run TUI tests
go test -v -race ./internal/ui/...

# Run specific screen tests
go test -v -run TestFileSelection ./internal/ui/screens/...
go test -v -run TestWizard ./internal/ui/...
```

### TUI Coordinators
The wizard uses dedicated coordinators for background operations:

- **ScanCoordinator**: Manages file system scanning state (`internal/ui/scan_coordinator.go`)
- **GenerateCoordinator**: Manages context generation state (`internal/ui/generate_coordinator.go`)

**Testing Coordinators:**
```bash
go test -v -run TestScanCoordinator ./internal/ui/
go test -v -run TestGenerateCoordinator ./internal/ui/
```

### Composed Screen Model Architecture

The `WizardModel` delegates screen-specific state to dedicated screen models in `internal/ui/screens/`:

| Screen Model | File | Owns State |
|--------------|------|------------|
| `FileSelectionModel` | `file_selection.go` | File tree, selections, filter |
| `TemplateSelectionModel` | `template_selection.go` | Template list, selected template |
| `TaskInputModel` | `task_input.go` | Task description text |
| `RulesInputModel` | `rules_input.go` | Rules text |
| `ReviewModel` | `review.go` | Summary display state |

**Adding a New Wizard Screen:**
1. Create `internal/ui/screens/my_screen.go` implementing `tea.Model`
2. Screen owns its own state (not stored in `WizardModel`)
3. Add screen model as field in `WizardModel`
4. Update `WizardModel.Update()` to delegate messages to the new screen
5. Update `WizardModel.View()` to render the new screen

**Accessing Screen State:**
```go
// WizardModel accessor methods delegate to screen models
func (m *WizardModel) getSelectedFiles() map[string]bool {
    if m.fileSelection != nil {
        return m.fileSelection.GetSelections()
    }
    return nil
}
```

### Testing External Binary Execution

Use the `CommandRunner` interface pattern for deterministic testing of external binaries:

**Production code** uses `DefaultRunner`:
```go
type CommandRunner interface {
    LookPath(file string) (string, error)
    Run(ctx context.Context, name string, args []string, stdin io.Reader) ([]byte, error)
}

type DefaultRunner struct{}
func (r *DefaultRunner) LookPath(file string) (string, error) {
    return exec.LookPath(file)
}
```

**Test code** uses `MockRunner`:
```go
type MockRunner struct {
    LookPathFunc func(file string) (string, error)
    RunFunc      func(ctx context.Context, name string, args []string, stdin io.Reader) ([]byte, error)
}

// Test with mock
runner := &MockRunner{
    LookPathFunc: func(file string) (string, error) {
        return "/usr/bin/geminiweb", nil
    },
    RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) ([]byte, error) {
        return []byte("mocked response"), nil
    },
}
provider := NewProvider(WithRunner(runner))
```

**Benefits:**
- Tests are deterministic (no dependency on installed binaries)
- Can test error conditions easily
- No network calls in tests
