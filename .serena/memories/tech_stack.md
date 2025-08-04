# Technology Stack

## Primary Language & Runtime
- **Go 1.24.5**: Main development language
- **Node.js >=14.0.0**: Build tooling and npm distribution

## Core UI Framework
- **BubbleTea (v1.3.6)**: Terminal UI framework for interactive applications
- **Lipgloss (v1.1.0)**: Styling and layout for terminal UI
- **Bubbles (v0.21.0)**: Pre-built UI components (progress bars, text inputs)

## Key Dependencies
- **go-gitignore**: Gitignore pattern matching for file filtering
- **go-openai**: OpenAI API integration for translation features
- **keyring (99designs)**: Secure credential storage
- **xdg**: XDG Base Directory specification support

## Build System
- **Hybrid npm/Go build system**: npm scripts orchestrate Go compilation
- **Cross-platform builds**: Supports Windows, Linux, macOS (x64 and ARM64)
- **Build optimization**: Uses `-ldflags="-s -w"` for smaller binaries

## Architecture Pattern
- **Clean Architecture**: Clear separation between UI, business logic, and external interfaces
- **Concurrent Processing**: Worker pools for file processing with progress tracking
- **Template-Based Generation**: Pluggable template system for different prompt types