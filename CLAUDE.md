# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

shotgun-cli is a terminal-based prompt generation tool built with Go and BubbleTea. It's designed to generate structured LLM prompts from codebase context using an interactive terminal user interface. The application uses inverse file selection (exclude files rather than include) and supports both built-in and custom prompt templates.

## Development Commands

### Build Commands
```bash
# Local development build
npm run build:local
# or directly: go build -o bin/shotgun-cli .

# Cross-platform builds
npm run build          # Current platform
npm run build:all      # All platforms
npm run build:windows  # Windows only
npm run build:linux    # Linux only
npm run build:macos    # macOS only

# Development mode
npm run dev
# or directly: go run .
```

### Testing and Quality
```bash
npm test               # Run Go tests
# or directly: go test ./...

npm run lint           # Go vet linting
# or directly: go vet ./...

npm run format         # Go formatting  
# or directly: go fmt ./...

# Clean build artifacts
npm run clean
```

### Debug Mode
```bash
DEBUG=1 go run .       # Enable debug logging to debug.log
```

## Architecture Overview

### Core Structure
- **main.go**: Application entry point with platform-specific initialization (Windows UTF-8 support)
- **internal/ui/**: BubbleTea UI components and application flow
- **internal/core/**: Business logic, file scanning, template processing, and configuration

### Key Components

#### UI Layer (`internal/ui/`)
- **app.go**: Main BubbleTea model with ViewState management (FileExclusion → TemplateSelection → TaskDescription → CustomRules → Generation)
- **filetree.go**: Interactive file tree with exclusion toggling
- **views.go**: View rendering for each application step
- **components.go**: Reusable UI components and styling
- **config_*.go**: Configuration management UI components

#### Core Logic (`internal/core/`)
- **scanner.go**: Directory scanning with gitignore integration
- **generator.go**: Context generation and prompt creation
- **template*.go**: Template system (built-in and custom template loading)
- **types.go**: Core data structures (FileNode, DirectoryScanner, ContextGenerator)
- **config*.go**: Configuration management and validation
- **enhanced_*.go**: Advanced features for translation and configuration

### Template System
- **Built-in templates**: Located in `templates/` directory with specific purposes:
  - `prompt_makeDiffGitFormat.md` (dev template)
  - `prompt_makePlan.md` (architect template)  
  - `prompt_analyzeBug.md` (debug template)
  - `prompt_projectManager.md` (project template)
- **Custom templates**: User-defined templates with YAML frontmatter in platform-specific directories
- **Template variables**: `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`, `{CURRENT_DATE}`

### File Processing
- **Three-layer filtering**: Built-in ignore patterns → .gitignore → custom rules
- **Inverse selection**: Users exclude files rather than include them
- **Progress tracking**: Real-time progress bars for large directory scans
- **UTF-8 support**: Platform-specific handling, especially Windows console configuration

### Configuration Management
- **Platform-specific paths**: Uses XDG standards on Unix, appropriate paths on Windows
- **JSON configuration**: Stored in user config directory
- **Custom templates directory**: Automatically created on first run
- **Keyring integration**: Secure credential storage via OS keychain

## Development Guidelines

### BubbleTea Pattern
The application follows the standard BubbleTea pattern with:
- **Model**: Application state and data
- **Update**: Message handling and state transitions  
- **View**: Rendering based on ViewState
- **Commands**: Async operations (scanning, generation)

### Error Handling
- Graceful degradation for permission issues
- User-friendly error messages
- Debug logging when DEBUG environment variable is set
- Platform-specific error handling (Windows console issues)

### Cross-Platform Considerations
- Windows UTF-8 console configuration in main.go
- Platform-specific directory paths using XDG library
- File path handling with filepath.Join for cross-platform compatibility

### Testing
- Unit tests for core logic components
- UI component tests where applicable  
- Platform-specific test cases for Windows input handling

## Key Dependencies

- **BubbleTea**: Terminal UI framework
- **Lipgloss**: Styling and layout
- **go-gitignore**: .gitignore parsing
- **xdg**: Cross-platform directory paths
- **koanf**: Configuration management
- **keyring**: Secure credential storage

## Build System

The project uses a hybrid npm/Go build system:
- **npm scripts**: Cross-platform build orchestration
- **Go toolchain**: Binary compilation
- **GoReleaser**: Release automation and packaging
- **Platform targets**: Windows, Linux, macOS (x64 and ARM64)