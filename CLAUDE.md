# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

shotgun-cli is a cross-platform CLI tool that generates LLM-optimized codebase contexts. It has two modes:
- **TUI Wizard**: Interactive 5-step interface using Bubble Tea framework
- **Headless CLI**: Command-line interface for automation using Cobra commands

## Build and Development Commands

### Building
- `make build` - Build for current platform (output: `build/shotgun-cli`)
- `make build-all` - Cross-compile for linux/darwin/windows (amd64/arm64)
- `make install` or `make install-local` - Install to `$GOPATH/bin`
- `make install-system` - Install to `/usr/local/bin` (requires sudo)

### Testing
- `make test` - Run unit tests
- `make test-race` - Run tests with race detector
- `make test-bench` - Run benchmarks
- `make test-e2e` - Run end-to-end CLI tests (located in `test/e2e`)
- `make coverage` - Generate coverage report to `coverage.out`

### Code Quality
- `make fmt` - Format code with `go fmt`
- `make lint` - Run golangci-lint (see `.golangci.yml` for enabled linters)
- `make vet` - Run `go vet` static analysis
- Pre-commit: Run `make fmt lint vet test` before committing

### Dependencies
- `make deps` - Download and verify module dependencies
- Go 1.24+ required

## Architecture

### Command Structure (cmd/)
All commands use Cobra framework with Viper configuration:
- **root.go**: Main entry point - launches TUI wizard when no args provided
- **context.go**: `context generate` command for headless context generation
- **template.go**: `template list/render` commands for template management
- **diff.go**: `diff split` command for splitting large diff files
- **config.go**: `config show/set` commands for configuration management
- **completion.go**: Shell completion generation

### Core Business Logic (internal/core/)
- **scanner/**: File system scanning with gitignore-style pattern matching
- **context/**: Context generation logic (creates LLM-optimized text output)
- **template/**: Template rendering with variable substitution
- **ignore/**: Gitignore pattern matching using `github.com/sabhiram/go-gitignore`

### TUI Components (internal/ui/)
The wizard uses Bubble Tea's Elm Architecture (Model-Update-View):
- **wizard.go**: Main orchestrator with WizardModel, handles navigation between steps
- **screens/**: Individual wizard steps (file_selection, template_selection, task_input, rules_input, review)
- **components/**: Reusable UI widgets
- **styles/**: Lip Gloss styling definitions

Key wizard flow:
1. File selection → scans directory with patterns
2. Template selection → picks prompt template
3. Task input → defines specific task
4. Rules input → adds constraints
5. Review → generates and optionally copies to clipboard

### Platform Integration (internal/platform/)
- **clipboard/**: Cross-platform clipboard operations using `github.com/atotto/clipboard`

### Key Dependencies
- CLI: Cobra (commands) + Viper (config)
- TUI: Bubble Tea (framework) + Lip Gloss (styling) + Bubbles (components)
- Logging: Zerolog for structured logging
- Pattern Matching: go-gitignore for gitignore-style patterns (NOT filepath.Match)

## Pattern Matching Behavior

**IMPORTANT**: This project uses gitignore-style patterns via `github.com/sabhiram/go-gitignore`, NOT Go's `filepath.Match`.

Key differences:
- `**` for recursive directory matching (e.g., `**/vendor/` matches vendor at any depth)
- `!` prefix for negation (include previously excluded files)
- Trailing `/` matches only directories
- Leading `/` anchors to repository root

Examples:
- `*.log` - All .log files
- `**/node_modules/` - node_modules at any depth
- `!important.tmp` - Include this file even if *.tmp is excluded
- `/build/` - Only root-level build directory

## Testing Patterns

- Unit tests: `*_test.go` files alongside source
- E2E tests: `test/e2e/` directory
- Use table-driven tests where applicable
- Mock file system operations for scanner tests

## Configuration

Default config location: `~/.config/shotgun-cli/config.yaml`
Override with `--config` flag or `SHOTGUN_CONFIG` env var

Common config keys:
- `scanner.max-files`: File count limit
- `scanner.max-file-size`: Size limit per file
- `output.clipboard`: Auto-copy to clipboard

## Module Path

**Import path**: `github.com/quantmind-br/shotgun-cli`

When adding new packages, use this as the base import path.
