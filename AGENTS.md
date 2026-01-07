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
internal/platform/      → Infrastructure (gemini, clipboard, openai, anthropic)
```

**Key rules**:
- `internal/core` must NOT import from `cmd`, `ui`, `app`, or `platform`
- Pass config via dependency injection, not global Viper access in core
- Use interfaces from core packages (`Scanner`, `Provider`, `ContextGenerator`)
- Create LLM providers via `app.DefaultProviderRegistry.Create()`

### Adding a New LLM Provider
1. Implement `llm.Provider` interface in `internal/platform/<provider>/`
2. Register in `internal/app/providers.go`

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

### ProviderRegistry (`internal/app/providers.go`)
```go
provider, err := app.DefaultProviderRegistry.Create(llm.ProviderOpenAI, cfg)
```

### Core Utilities
- `diff.IntelligentSplit` - Split large diffs preserving file boundaries
- `scanner.CollectSelections` - Handle file tree selections
- `config.ValidateValue(key, value)` - Validate config before saving
