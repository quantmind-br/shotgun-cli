# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`shotgun-cli` is a Go CLI tool that generates LLM-optimized codebase contexts. It follows **Clean Architecture** principles, separating concerns into distinct layers. It has two primary modes of operation: an interactive TUI wizard for guided use and a headless CLI mode for automation.

## Build & Development Commands

**Go Version**: 1.24+ required (see `go.mod`)

Here are the common commands you'll use during development, all managed via the `Makefile`.

```bash
# Build the binary for your current platform
make build

# Cross-compile for all target platforms (linux, darwin, windows)
make build-all

# Run all unit tests
make test

# Run tests with the race condition detector enabled
make test-race

# Run the end-to-end tests
make test-e2e

# Run tests for a specific package
go test ./internal/core/scanner/...

# Run a single test by name
go test -run TestMyFunction ./internal/core/scanner

# Run the linter
make lint

# Run static analysis
make vet

# Format all Go code
make fmt

# Generate coverage report
make coverage
```

## Architecture Overview

The application uses a **Clean/Hexagonal Architecture** to isolate business logic from external concerns like the UI and APIs.

### Architectural Layers

1.  **Presentation/Adapter Layer (`cmd`, `internal/ui`)**: Handles all user interaction. This includes the Cobra-based CLI commands and the interactive Bubble Tea TUI wizard.
2.  **Core/Domain Layer (`internal/core`)**: Contains the pure business logic for context generation, file scanning, template management, and token estimation. This layer has no knowledge of the UI or any specific external tools.
3.  **Infrastructure/Platform Layer (`internal/platform`)**: Implements integrations with external systems. This is where we interact with the Gemini AI API (via an external binary) and the system clipboard.

### Core Package Structure

```
shotgun-cli/
├── cmd/                      # Cobra CLI commands (Presentation)
│   ├── root.go               # Main entry point, launches TUI or commands
│   ├── context.go            # `context generate` and `context send` commands
│   └── template.go           # `template list` and `template render` commands
│
├── internal/
│   ├── core/                 # Core Business Logic (Domain)
│   │   ├── context/          # Context generation engine
│   │   ├── ignore/           # Layered ignore rule engine (.gitignore, etc.)
│   │   ├── scanner/          # Filesystem scanning and tree building
│   │   ├── template/         # Template management (load, render, validate)
│   │   └── tokens/           # Token estimation utilities
│   │
│   ├── ui/                   # TUI Wizard (Presentation)
│   │   ├── wizard.go         # Main TUI orchestrator (Bubble Tea MVU model)
│   │   ├── screens/          # Each step of the wizard is a "screen"
│   │   └── components/       # Reusable TUI components (e.g., progress bar)
│   │
│   ├── platform/             # External Integrations (Infrastructure)
│   │   ├── clipboard/        # Cross-platform clipboard access
│   │   └── gemini/           # Gemini integration via external `geminiweb` binary
│   │
│   └── assets/
│       └── templates/        # Embedded templates via go:embed
│
└── Makefile                  # Build, test, and lint commands
```

## Key Components & Interfaces

The core logic is built around a set of key interfaces that promote loose coupling.

### `internal/core/scanner.Scanner`
Scans the filesystem, applies ignore rules, and builds a tree of `FileNode` objects.
```go
type Scanner interface {
    Scan(rootPath string, config *ScanConfig) (*FileNode, error)
    ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}
```

### `internal/core/context.ContextGenerator`
Takes the file tree and user selections to generate the final, structured context string for the LLM.
```go
type ContextGenerator interface {
    Generate(root *scanner.FileNode, selections map[string]bool, config GenerateConfig) (string, error)
    GenerateWithProgressEx(root *scanner.FileNode, selections map[string]bool, config GenerateConfig, progress func(GenProgress)) (string, error)
}
```

### `internal/core/template.TemplateManager`
Manages prompt templates, loading them from embedded assets, user config directories, or custom paths.
```go
type TemplateManager interface {
    ListTemplates() ([]Template, error)
    GetTemplate(name string) (*Template, error)
    RenderTemplate(name string, vars map[string]string) (string, error)
}
```

### `internal/core/ignore.IgnoreEngine`
Implements a layered ignore system that processes rules from built-in patterns, `.gitignore`, `.shotgunignore`, and custom rules provided at runtime.
```go
type IgnoreEngine interface {
    ShouldIgnore(relPath string) (bool, IgnoreReason)
    LoadGitignore(rootDir string) error
}
```

## Design Patterns
- **Command Pattern**: Used in the `cmd` package with Cobra. Each subcommand is a separate, self-contained command object.
- **Model-View-Update (MVU)**: The entire `internal/ui` package is built on this pattern using the Bubble Tea framework. State is managed centrally in `wizard.go` and updated via messages.
- **Builder Pattern**: The `ContextGenerator` in `internal/core/context` acts as a builder, progressively assembling the final context from the file tree, file content, and templates.
- **Strategy Pattern**: The integration with the AI provider in `internal/platform/gemini` is abstracted, allowing for potential future support of other providers. The `Executor` handles the "how" of sending a request.

## Development Gotchas & Warnings
- **External `geminiweb` Dependency**: The tool **does not** directly call the Gemini API. It executes a separate binary, `geminiweb`. For the `context send` command to work, this binary must be installed and configured (`geminiweb auto-login`).
- **TUI Complexity**: The Bubble Tea TUI is stateful and event-driven. Debugging UI issues often involves tracing messages (`tea.Msg`) and state updates in `internal/ui/wizard.go`.
- **Template Loading Priority**: Remember the template override order: custom path > user config (`~/.config/shotgun-cli/templates`) > embedded assets. This can be a source of confusion if a template isn't loading as expected.
- **Configuration**: The app uses Viper, loading from `config.yaml`, environment variables (`SHOTGUN_...`), and flags. Use `shotgun-cli config show` to debug configuration issues and see where values are coming from.
- **Core Layer Isolation**: The `internal/core` package **MUST NOT** depend on `cmd`, `internal/ui`, or `internal/platform`. This is critical for maintaining clean architecture boundaries.

## Gemini Execution Flow

When modifying the Gemini integration (`internal/platform/gemini`), follow this sequence:

1. **Prerequisite Checks**: Call `gemini.IsAvailable()` and `gemini.IsConfigured()` first
2. **Build Command**: Use `buildArgs` to construct CLI arguments for `geminiweb`
3. **Execute via stdin**: Pass the context/prompt to `geminiweb` via standard input
4. **Capture stdout**: Read the AI response from the process's standard output
5. **Process Response**: Clean ANSI codes with `StripANSI` and parse with `ParseResponse`

**Anti-Pattern**: Never make direct HTTP requests to Google's APIs. All interactions must go through `gemini.Executor`.

## Code Conventions

### Configuration Access
- **DO NOT** access Viper globally (e.g., `viper.GetString()`) from within `internal/core` or `internal/platform`.
- **DO** pass configuration values or dedicated config structs from the `cmd` layer down to the components.
- The `cmd` package acts as the **composition root** where services are instantiated and wired together.

### Error Handling
Always wrap errors with context using `fmt.Errorf` with the `%w` verb:
```go
if err != nil {
    return nil, fmt.Errorf("failed to scan directory %s: %w", rootPath, err)
}
```

### Logging
Use `zerolog` for structured logging. Inject `zerolog.Logger` into components. Do not use `fmt.Println` or standard `log` package for application logging.

### Testing
Use `github.com/stretchr/testify` for assertions. Tests follow Go conventions with `_test.go` suffixes. E2E tests are in `test/e2e/`.

### Imports
Group imports: standard library, external packages, internal packages. Use `goimports` for formatting.

## TUI Navigation & Step Skipping

The wizard has 5 steps defined in `internal/ui/wizard.go`:
1. **StepFileSelection** - Select files to include in context
2. **StepTemplateSelection** - Choose a prompt template
3. **StepTaskInput** - Enter task description (skipped if template lacks `{{TASK}}`)
4. **StepRulesInput** - Enter custom rules (skipped if template lacks `{{RULES}}`)
5. **StepReview** - Review and generate

The `getNextStep()` and `getPrevStep()` methods handle automatic step skipping based on template variables. When modifying navigation, update both methods to maintain consistency.

## Key Libraries
- **CLI**: `github.com/spf13/cobra` (commands) & `github.com/spf13/viper` (config)
- **TUI**: `github.com/charmbracelet/bubbletea` (framework), `lipgloss` (styling), `bubbles` (components)
- **Logging**: `github.com/rs/zerolog` for structured, leveled logging.
- **Ignore Patterns**: `github.com/sabhiram/go-gitignore` for `.gitignore` style filtering.