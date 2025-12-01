# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build
make build                    # Build binary for current platform (output: build/shotgun-cli)
make build-all                # Cross-compile for linux/darwin/windows (amd64/arm64)

# Test
make test                     # Run unit tests
make test-race                # Run tests with race detector
make test-e2e                 # Run end-to-end tests (./test/e2e)
go test ./internal/core/scanner/...  # Run tests for specific package
go test -run TestFunctionName ./pkg  # Run single test

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

shotgun-cli is a CLI tool that generates LLM-optimized codebase contexts. It has two modes:
1. **TUI Wizard Mode** - Interactive 5-step Bubble Tea interface (default when run without args)
2. **Headless CLI Mode** - Commands for automation (`context generate`, `template`, `diff`, `config`)

### Core Package Structure

```
cmd/                          # Cobra CLI commands
├── root.go                   # Main command, launches TUI wizard
├── context.go                # Context generation commands
├── template.go               # Template management
├── diff.go                   # Diff splitting tools
└── config.go                 # Configuration management

internal/
├── core/
│   ├── scanner/              # File system scanning with gitignore support
│   │   ├── scanner.go        # Scanner interface and FileNode struct
│   │   └── filesystem.go     # FilesystemScanner implementation
│   ├── context/              # Context generation (tree + content)
│   │   ├── generator.go      # ContextGenerator interface
│   │   ├── tree.go           # TreeRenderer for file structure
│   │   └── content.go        # File content extraction
│   ├── template/             # Template engine with multi-source loading
│   │   ├── manager.go        # Template discovery and management
│   │   ├── loader.go         # Multi-source template loading
│   │   └── renderer.go       # Template rendering with variables
│   └── ignore/               # Gitignore pattern matching (uses go-gitignore)
├── ui/
│   ├── wizard.go             # Main wizard orchestration (WizardModel)
│   ├── screens/              # Individual wizard steps
│   │   ├── file_selection.go # Step 1: File tree selection
│   │   ├── template_selection.go  # Step 2: Template picker
│   │   ├── task_input.go     # Step 3: Task description
│   │   ├── rules_input.go    # Step 4: Custom rules
│   │   └── review.go         # Step 5: Preview and generate
│   ├── components/           # Reusable UI components
│   └── styles/               # Theme and styling (Lip Gloss)
├── platform/
│   ├── clipboard/            # Cross-platform clipboard (darwin/linux/windows/wsl)
│   └── gemini/               # Gemini integration
└── assets/
    └── embed.go              # Embedded templates via go:embed
```

### Key Interfaces

- `scanner.Scanner` - File system scanning with progress reporting
- `context.ContextGenerator` - Generates context from selected files
- Template loading priority: custom path > user directory (~/.config/shotgun-cli/templates/) > embedded

### Configuration

- Config file: `~/.config/shotgun-cli/config.yaml`
- Environment prefix: `SHOTGUN_` (e.g., `SHOTGUN_LOG_LEVEL`)
- Uses Viper for configuration management

## Key Libraries

- **CLI**: Cobra (commands) + Viper (config)
- **TUI**: Bubble Tea (framework) + Lip Gloss (styling) + Bubbles (components)
- **Logging**: Zerolog
- **Patterns**: go-gitignore for gitignore-style filtering
