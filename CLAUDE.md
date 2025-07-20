# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

shotgun-cli is a terminal-based prompt generation tool built with Go and BubbleTea. It's designed to help developers generate structured LLM prompts from codebase context using an interactive TUI interface with inverse file selection.

## Development Commands

### Essential Commands
```bash
# Development workflow
npm run dev              # Run in development mode (go run .)
npm test                 # Run all Go tests (go test ./...)
npm run lint             # Run Go vet (go vet ./...)
npm run format           # Format Go code (go fmt ./...)

# Building
npm run build            # Build for current platform
npm run build:local      # Local development build (go build -o bin/shotgun-cli .)
npm run build:all        # Build for all platforms (Windows, Linux, macOS, ARM64)
npm run build:windows    # Build for Windows specifically
npm run build:linux      # Build for Linux specifically  
npm run build:macos      # Build for macOS specifically

# Maintenance
npm run clean            # Clean build artifacts
```

### Testing
- Run single test: `go test -v ./internal/core -run TestSpecificFunction`
- Run tests with coverage: `go test -v -cover ./...`
- Run benchmarks: `go test -bench=. ./...`

## Architecture Overview

### Core Structure
```
internal/
├── core/           # Business logic layer
│   ├── types.go    # Core data structures and interfaces
│   ├── scanner.go  # Directory scanning with gitignore support
│   ├── generator.go # Context generation and file processing
│   ├── template.go # Template processing (complex templates)
│   └── template_simple.go # Simple template processing
├── ui/             # BubbleTea UI layer
│   ├── app.go      # Main application model and state
│   ├── filetree.go # File tree UI with exclusion controls
│   └── views.go    # View rendering logic
```

### Key Components

**DirectoryScanner**: Handles recursive directory traversal with three-layer filtering:
1. Built-in ignore patterns (node_modules, .git, binaries, etc.)
2. Project .gitignore rules 
3. Custom user exclusions

**ContextGenerator**: Processes included files and generates prompt context with progress tracking and worker pools for performance.

**TemplateProcessor**: Manages four prompt templates:
- Dev (`prompt_makeDiffGitFormat.md`) - Git diff generation
- Architect (`prompt_makePlan.md`) - Design planning
- Debug (`prompt_analyzeBug.md`) - Bug analysis  
- Project Manager (`prompt_projectManager.md`) - Documentation sync

**FileTreeModel**: Interactive TUI component for inverse file selection (exclude rather than include files).

### State Management
- `ViewState` enum tracks current UI step (FileExclusion → PromptComposition → Generation → Complete)
- `SelectionState` manages file inclusion/exclusion with thread-safe operations
- Progress updates flow through channels for real-time UI feedback

### Template System
Templates use Go's `text/template` with these variables:
- `{TASK}` - User's task description
- `{RULES}` - Custom rules or "no additional rules"  
- `{FILE_STRUCTURE}` - Generated file structure and content

## Build System

The project uses a hybrid npm/Go build system:
- **npm scripts** orchestrate cross-platform builds
- **scripts/build.js** handles platform-specific Go compilation with optimized flags (`-ldflags="-s -w"`)
- **scripts/install.js** manages post-install binary setup
- **package.json** defines the CLI wrapper and npm distribution

Binary outputs:
- `bin/shotgun-cli.exe` (Windows x64)
- `bin/shotgun-cli-linux` (Linux x64)  
- `bin/shotgun-cli-macos` (macOS x64)
- `bin/shotgun-cli-macos-arm64` (macOS ARM64)

## Key Dependencies

- **BubbleTea**: TUI framework for interactive interface
- **Lipgloss**: Styling and layout for terminal UI
- **Bubbles**: Pre-built UI components (progress bars, text inputs)
- **go-gitignore**: Gitignore pattern matching

## Development Guidelines

### File Structure Conventions
- Use `internal/` for private packages not meant for external import
- Core business logic stays in `internal/core/`
- UI/presentation logic stays in `internal/ui/`
- Templates are external assets in `templates/` directory

### Error Handling
- Use `ShotgunError` struct for structured errors with operation and path context
- `ErrorCollector` aggregates multiple errors during batch operations
- Always validate directory access before operations

### Concurrency Patterns
- Use channels for progress updates between goroutines and UI
- Worker pools (`workerPool chan struct{}`) limit concurrent file operations
- Mutex protection for shared state (selection maps, progress counters)

### Template Development
Templates must output valid git diff format for Dev template, or structured markdown for others. Use `{TASK}`, `{RULES}`, and `{FILE_STRUCTURE}` placeholders.