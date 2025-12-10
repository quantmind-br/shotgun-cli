# AGENTS.md - Universal AI Agent Configuration

## Project Overview
This project is a Go CLI tool named `shotgun-cli` that generates LLM-optimized codebase contexts. It features an interactive TUI wizard for guided use and a headless mode for automation.

## Build & Test Commands
- Build the binary: `make build`
- Run all unit tests: `make test`
- Run tests with race detector: `make test-race`
- Lint the code: `make lint`
- Format the code: `make fmt`

## Architecture Overview
- **Clean Architecture**: The project follows a layered architecture separating presentation, core logic, and infrastructure.
- **Presentation Layer**: `cmd/` (Cobra CLI) and `internal/ui/` (Bubble Tea TUI).
- **Core Logic Layer**: `internal/core/` contains business logic for scanning, context generation, and template management.
- **Infrastructure Layer**: `internal/platform/` handles external integrations like the Gemini AI API and clipboard.
- **External Dependency**: Interacts with the Gemini API by executing an external `geminiweb` binary.

## Code Style Conventions
- Code is formatted with `go fmt`.
- Linting is enforced by `golangci-lint` using the `.golangci.yml` configuration. Run `make lint` to check.

## Testing Instructions
- Unit tests are located alongside the code. Run all tests with `make test`.
- To run tests for a specific package, use `go test ./internal/core/scanner/...`.
- End-to-end tests are available via `make test-e2e`.

## Key Conventions and Patterns
- **Dependency Injection**: Manual dependency injection is used, with `cmd/` acting as the composition root.
- **TUI Pattern**: The interactive wizard uses the Model-View-Update (MVU) pattern provided by the Bubble Tea framework.
- **Ignore Engine**: A layered ignore system is used, processing rules from `.gitignore`, `.shotgunignore`, and built-in patterns.
- **Template System**: Templates are loaded from multiple sources (embedded, user config, custom path) with a clear priority system.
