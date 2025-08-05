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
├── core/                      # Business logic layer
│   ├── types.go              # Core data structures and interfaces
│   ├── scanner.go            # Directory scanning with gitignore support
│   ├── generator.go          # Context generation and file processing
│   ├── template.go           # Template processing (complex templates)
│   ├── template_simple.go    # Simple template processing
│   ├── config.go             # Basic configuration management
│   ├── config_manager.go     # Enhanced configuration with Koanf
│   ├── enhanced_config.go    # Enhanced config structures with validation
│   ├── keyring.go            # Secure API key management
│   ├── enhanced_translator.go # Production translation service
│   ├── translation_errors.go # Structured translation error types
│   ├── retry_handler.go      # Retry logic with exponential backoff
│   └── circuit_breaker.go    # Circuit breaker for fault tolerance
├── ui/                       # BubbleTea UI layer
│   ├── app.go               # Main application model and state
│   ├── filetree.go          # File tree UI with exclusion controls
│   ├── components.go        # TextArea component for user input
│   ├── views.go             # View rendering logic
│   ├── enhanced_config_form.go # Advanced configuration UI
│   ├── translation_views.go  # Translation UI with progress indicators
│   ├── config_sections.go    # Configuration form sections
│   └── windows_input.go      # Windows-specific input handling
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

**NumberedTextArea**: Simple multiline input component for Task Description and Custom Rules fields.

### Enhanced Configuration System

**EnhancedConfigManager**: Advanced configuration management using Koanf with:
- Multi-source configuration (files, environment variables, defaults)
- Real-time configuration validation with structured error reporting
- File watching for live configuration updates
- XDG Base Directory specification compliance for cross-platform config storage

**SecureKeyManager**: Secure API key management with:
- Windows Credential Manager integration for encrypted storage
- Cross-platform keyring support with filesystem fallback
- Configurable API key aliases for multiple service providers

**Configuration Hierarchy** (highest to lowest priority):
1. Environment variables (`SHOTGUN_*` prefix)
2. Configuration file (`~/.config/shotgun-cli/config.json`)
3. Built-in defaults

### Translation System Architecture

**EnhancedTranslationService**: Production-ready translation service with:
- Circuit breaker pattern for fault tolerance
- Intelligent retry logic with exponential backoff and jitter
- LRU caching with TTL for performance optimization
- Rate limiting to respect API quotas
- Comprehensive metrics and monitoring

**Translation Flow**:
1. Input validation and preprocessing
2. Cache lookup for existing translations
3. Rate limiting and circuit breaker checks
4. API request with retry logic
5. Response processing and cache storage
6. Metrics collection and error tracking

**Error Handling**: Structured error types (Network, Auth, RateLimit, Validation, etc.) with automatic classification and retry eligibility determination.

### State Management
- `ViewState` enum tracks current UI step (FileExclusion → PromptComposition → Generation → Complete)
- `SelectionState` manages file inclusion/exclusion with thread-safe operations
- Progress updates flow through channels for real-time UI feedback

### Template System
Templates use simple string replacement:
- `{TASK}` - User's task description
- `{RULES}` - Custom rules or "no additional rules"
- `{FILE_STRUCTURE}` - Generated file structure and content
- `{CURRENT_DATE}` - Current date in YYYY-MM-DD format

## Input System

The application supports multiline input for both Task Description and Custom Rules fields.

### Keyboard Shortcuts (Updated for Separate Screens)
**General Navigation:**
- **Esc**: Go back to previous step
- **?**: Toggle help menu
- **Ctrl+Q, Ctrl+C**: Quit application

**File Exclusion Screen:**
- **hjkl or ↑↓←→**: Navigate file tree
- **Space**: Toggle file/directory exclusion
- **c**: Continue to next step
- **o**: Access configuration/settings
- **r**: Reset all exclusions
- **a/A**: Exclude/include all files

**Template Selection Screen:**
- **↑/↓ (or k/j)**: Navigate template options
- **1-4**: Quick select template by number
- **Enter**: Confirm selection and continue

**Task Description Screen:**
- **Tab**: Focus/unfocus input field
- **F5**: Continue to Custom Rules
- **Esc**: Go back to template selection

**Custom Rules Screen:**
- **Tab**: Focus/unfocus input field
- **F5**: Generate prompt
- **Esc**: Go back to task description

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
- **Koanf**: Configuration management with multi-source support
- **go-playground/validator**: Configuration validation with struct tags
- **go-keyring**: Cross-platform secure credential storage
- **sony/gobreaker**: Circuit breaker implementation for fault tolerance
- **golang.org/x/time/rate**: Token bucket rate limiting
- **sashabaranov/go-openai**: OpenAI API client for translation services

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

## Configuration Management

### Configuration File Location
- **Primary**: `~/.config/shotgun-cli/config.json` (XDG standard)
- **Windows**: `%LOCALAPPDATA%\shotgun-cli\config.json`
- **macOS**: `~/Library/Application Support/shotgun-cli/config.json`
- **Linux**: `~/.config/shotgun-cli/config.json`

### API Key Storage
- **Windows**: Windows Credential Manager (encrypted)
- **macOS**: Keychain (encrypted)
- **Linux**: Secret Service API or filesystem fallback
- **Fallback**: `~/.config/shotgun-cli/keyring/` directory

### Environment Variables
All configuration can be overridden via environment variables with `SHOTGUN_` prefix:
- `SHOTGUN_OPENAI_MODEL=gpt-4o`
- `SHOTGUN_TRANSLATION_ENABLED=true`
- `SHOTGUN_APP_THEME=dark`

### Configuration Validation
The system validates all configuration on load with detailed error reporting. Common validation rules:
- Model names must be from supported OpenAI models list
- URLs must be valid HTTP/HTTPS endpoints
- Numeric ranges enforced (timeouts, token limits, etc.)
- Language codes must be supported for translation

## Testing Architecture

### Test Structure
- **Unit Tests**: `*_test.go` files alongside source code
- **Integration Tests**: Complex workflows testing multiple components
- **Benchmark Tests**: Performance validation for critical paths

### Mock Strategy
- **Translation Service**: Mock OpenAI client with configurable responses and error simulation
- **Configuration**: In-memory configuration for isolated testing
- **Keyring**: Mock key manager to avoid credential dependencies

### Windows Compatibility Testing
Special attention to Windows-specific behaviors:
- Path handling with backslashes vs forward slashes
- CRLF line endings in git operations
- Windows Credential Manager integration
- File permissions and temporary file creation