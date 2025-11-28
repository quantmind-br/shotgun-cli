# AGENTS.md

Go 1.24+ CLI tool for generating LLM-optimized codebase contexts. Uses Cobra/Viper for CLI, Bubble Tea for TUI.

## Commands
- `make build` - Build binary to `build/shotgun-cli`
- `make test` - Run all unit tests
- `go test -v -run TestName ./path/to/pkg` - Run single test
- `make test-e2e` - Run E2E tests in `test/e2e/`
- `make lint` - Run golangci-lint (max line length: 120, cyclomatic complexity: 15)
- `make fmt` - Format code
- Pre-commit: `make fmt lint vet test`

## Architecture
- `cmd/` - Cobra commands (root.go launches TUI, context.go for headless mode)
- `internal/core/` - Business logic: scanner/, context/, template/, ignore/
- `internal/ui/` - Bubble Tea TUI: wizard.go, screens/, components/, styles/
- `internal/platform/` - Cross-platform utilities (clipboard)

## Code Style
- Import path: `github.com/quantmind-br/shotgun-cli`
- Pattern matching uses gitignore-style (`**`, `!`, trailing `/`) via go-gitignore, NOT filepath.Match
- Table-driven tests in `*_test.go` alongside source
- Zerolog for structured logging
- Config: `~/.config/shotgun-cli/config.yaml` (XDG compliant)
