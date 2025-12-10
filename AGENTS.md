# AGENTS.md - Universal AI Agent Configuration

## Build & Test Commands
- Build: `make build` | Run tests: `make test` | Single test: `go test ./internal/core/scanner -v` | Race detector: `make test-race` | E2E: `make test-e2e` | Lint: `make lint` | Format: `make fmt`

## Architecture & Patterns
- **Clean Architecture**: `cmd` (Presentation) → `internal/core` (Domain) → `internal/platform` (Infrastructure) | **DI**: Manual DI, `cmd` is composition root | **TUI**: MVU pattern with Bubble Tea | **Interfaces**: `Scanner`, `ContextGenerator`, `TemplateManager`, `IgnoreEngine`

## Code Style & Conventions
- **Error Handling**: Always wrap: `fmt.Errorf("context: %w", err)` | **Logging**: Use `zerolog`, never `fmt.Println` | **Config**: Viper in `cmd` only, pass structs down | **Imports**: Grouped (std, external, internal) with `goimports`

## Critical Rules & Gotchas
- **Gemini Integration**: **NEVER** call Gemini API directly. Use `internal/platform/gemini.Executor` with external `geminiweb` binary | **Core Isolation**: `internal/core` **MUST NOT** depend on `cmd`, `ui`, or `platform` | **Templates**: custom > user config > embedded | **State**: CLI is stateless, TUI manages state in `WizardModel`