# shotgun-cli

A cross-platform CLI tool written in Go that generates LLM-optimized codebase contexts. It features both an interactive TUI (Terminal User Interface) wizard and a headless CLI mode for automation.

## Project Overview

- **Language:** Go 1.24
- **Purpose:** Scans codebases, filters files (gitignore-style), and generates structured text context (file tree + content) optimized for LLM consumption.
- **Key Libraries:**
    - `spf13/cobra` (CLI commands)
    - `charmbracelet/bubbletea` (TUI framework)
    - `charmbracelet/lipgloss` (Styling)
    - `spf13/viper` (Configuration)
    - `rs/zerolog` (Logging)
    - `sabhiram/go-gitignore` (File filtering)

## Architecture

The project follows a standard Go CLI structure:

- **`cmd/`**: Entry points for the CLI commands (`root.go`, `context.go`, `config.go`, etc.).
- **`internal/core/`**: Core business logic.
    - `scanner/`: File system scanning and filtering.
    - `context/`: Logic for generating the context output (tree and content).
    - `template/`: Template engine for prompts.
    - `ignore/`: Gitignore pattern matching engine.
- **`internal/ui/`**: TUI implementation using Bubble Tea.
    - `wizard.go`: Main wizard model.
    - `screens/`: Individual wizard steps (File Selection, Template, Review, etc.).
- **`internal/platform/`**: Platform-specific code (Clipboard, etc.).

## Build and Run

The project uses a `Makefile` for common tasks.

### Build
```bash
make build          # Builds binary to ./build/shotgun-cli
make build-all      # Cross-compile for Linux, Mac, Windows
```

### Run
**Interactive TUI Mode:**
```bash
./build/shotgun-cli
```

**Headless CLI Mode:**
```bash
./build/shotgun-cli context generate --include "*.go" --output context.md
```

### Test
```bash
make test           # Run unit tests
make test-e2e       # Run end-to-end tests
make test-race      # Run tests with race detection
make lint           # Run golangci-lint
```

## Configuration

Configuration is managed via `~/.config/shotgun-cli/config.yaml` or environment variables (`SHOTGUN_*`).
Key settings include `scanner.max-files`, `context.max-size`, and `template.custom-path`.

## Current Focus & Roadmap

Refactoring and optimization are currently prioritized (see `plan.md`):
1.  **Code Cleanup:** Removing dead code and legacy implementations (e.g., old matcher logic in scanner).
2.  **Consolidation:** Centralizing utility functions (like size parsing) into `internal/utils/conversion`.
3.  **Simplification:** Streamlining the `clipboard` integration (removing complex platform-specific managers in favor of simpler libraries or abstractions).
