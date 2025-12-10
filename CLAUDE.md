# CLAUDE.md

This file provides guidance to Claude when working with code in this repository. It has been generated based on a detailed analysis of the codebase and incorporates conventions from the existing project documentation.

## Build & Development Commands

```bash
# Build
make build                    # Build binary for current platform (output: build/shotgun-cli)
make build-all                # Cross-compile for linux/darwin/windows (amd64/arm64)

# Test
make test                     # Run unit tests
make test-race                # Run tests with race detector
make test-e2e                 # Run end-to-end tests (./test/e2e)
go test ./internal/core/scanner/...  # Run tests for a specific package
go test -run TestFunctionName ./pkg  # Run a single test

# Code Quality
make lint                     # Run golangci-lint (config: .golangci.yml)
make fmt                      # Format code with go fmt
make vet                      # Run go vet static analysis
make coverage                 # Generate coverage report

# Install
make install                  # Install to GOPATH/bin
make install-system           # Install to /usr/local/bin (requires sudo)
```

## Architecture Overview

`shotgun-cli` is a Go CLI tool that generates LLM-optimized codebase contexts. It follows **Clean Architecture** principles, separating concerns into distinct layers. It has two primary modes of operation:
1.  **TUI Wizard Mode**: An interactive 5-step Bubble Tea interface (default when run without args).
2.  **Headless CLI Mode**: Commands for automation (`context generate`, `template`, `diff`, `config`).

### Architectural Layers

-   **Presentation/Adapter Layer (`cmd`, `internal/ui`)**: Handles user interaction through CLI commands (Cobra) and the interactive TUI wizard (Bubble Tea).
-   **Core/Domain Layer (`internal/core`)**: Contains the core business logic for context generation, file scanning, template management, and token estimation. This layer is independent of UI and infrastructure.
-   **Infrastructure/Platform Layer (`internal/platform`)**: Implements integrations with external systems, such as the Gemini AI API (via the `geminiweb` binary) and the system clipboard.

### Core Package Structure

```
cmd/                          # Cobra CLI commands (Presentation Layer)
├── root.go                   # Main command, launches TUI wizard
├── send.go                   # Sends context to AI
└── template.go               # Template management

internal/
├── core/                     # Core Business Logic (Domain Layer)
│   ├── context/              # Context generation (tree + content)
│   ├── ignore/               # Layered ignore rule engine
│   ├── scanner/              # File system scanning with gitignore support
│   ├── template/             # Template engine with multi-source loading
│   └── tokens/               # Token estimation for context size
├── ui/                       # TUI Wizard (Presentation Layer)
│   ├── wizard.go             # Main wizard orchestration (MVU pattern)
│   ├── screens/              # Individual wizard steps
│   └── components/           # Reusable UI components
├── platform/                 # External Integrations (Infrastructure Layer)
│   ├── clipboard/            # Cross-platform clipboard access
│   └── gemini/               # Gemini integration via external binary
└── assets/
    └── templates/            # Embedded templates via go:embed
```

## Key Components & Interfaces

### `internal/core/scanner`
Scans the filesystem and builds a file tree, applying ignore rules.
```go
type Scanner interface {
    Scan(rootPath string, config *ScanConfig) (*FileNode, error)
    ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}
```

### `internal/core/context`
Aggregates the file tree and file contents into the final structured context for the LLM.
```go
type ContextGenerator interface {
    Generate(root *scanner.FileNode, selections map[string]bool, config GenerateConfig) (string, error)
    GenerateWithProgressEx(root *scanner.FileNode, selections map[string]bool, config GenerateConfig, progress func(GenProgress)) (string, error)
}
```

### `internal/core/template`
Manages prompt templates, loading them from embedded assets, user configuration directories, or custom paths.
```go
type TemplateManager interface {
    ListTemplates() ([]Template, error)
    GetTemplate(name string) (*Template, error)
    RenderTemplate(name string, vars map[string]string) (string, error)
}
```

### `internal/core/ignore`
Implements a layered ignore engine that processes rules from built-in patterns, `.gitignore`, `.shotgunignore`, and custom rules.
```go
type IgnoreEngine interface {
    ShouldIgnore(relPath string) (bool, IgnoreReason)
    LoadGitignore(rootDir string) error
}
```

## Design Patterns
- **Command Pattern**: Used in the `cmd` package with Cobra. Each subcommand is a separate command object.
- **Model-View-Update (MVU)**: The entire `internal/ui` package is built on this pattern using the Bubble Tea framework. State is managed centrally in `wizard.go`.
- **Builder Pattern**: The `ContextGenerator` in `internal/core/context` acts as a builder, progressively assembling the final context from various pieces (file tree, file content, templates).
- **Strategy Pattern**: The integration with the AI provider in `internal/platform/gemini` is abstracted, allowing for potential future support of other providers.

## Development Gotchas & Warnings
- **External `geminiweb` Dependency**: The tool does not directly call the Gemini API. It executes a separate binary, `geminiweb`. Ensure this binary is installed and configured (`geminiweb auto-login`) for the `context send` command to work.
- **TUI Complexity**: The Bubble Tea TUI is stateful and event-driven. Debugging UI issues often involves tracing messages and state updates in `internal/ui/wizard.go`.
- **Template Loading Priority**: Remember the template override order: custom path > user directory (`~/.config/shotgun-cli/templates/`) > embedded templates. This can be a source of confusion if a template isn't loading as expected.
- **Configuration**: The app uses Viper, loading from `config.yaml`, environment variables (`SHOTGUN_...`), and flags. Check `shotgun-cli config show` to debug configuration issues.

## Key Libraries
- **CLI**: `github.com/spf13/cobra` (commands) + `github.com/spf13/viper` (config)
- **TUI**: `github.com/charmbracelet/bubbletea` (framework) + `lipgloss` (styling) + `bubbles` (components)
- **Logging**: `github.com/rs/zerolog` for structured, leveled logging.
- **Ignore Patterns**: `github.com/sabhiram/go-gitignore` for `.gitignore` style filtering.
