# Essential Development Commands

## Daily Development Workflow
```bash
# Run in development mode
npm run dev              # Equivalent to: go run .

# Testing
npm test                 # Run all Go tests: go test ./...
go test -v ./internal/core -run TestSpecificFunction  # Run specific test
go test -v -cover ./...  # Run tests with coverage
go test -bench=. ./...   # Run benchmarks

# Code Quality
npm run lint             # Run Go vet: go vet ./...
npm run format           # Format Go code: go fmt ./...
```

## Building and Distribution
```bash
# Local development build
npm run build:local      # Build for current platform: go build -o bin/shotgun-cli .

# Cross-platform builds
npm run build            # Build for current platform
npm run build:all        # Build for all platforms (Windows, Linux, macOS, ARM64)
npm run build:windows    # Build for Windows specifically
npm run build:linux      # Build for Linux specifically
npm run build:macos      # Build for macOS specifically

# Cleanup
npm run clean            # Clean build artifacts
```

## Windows-Specific Commands
Since this is a Windows development environment:
```cmd
# File operations
dir                      # List directory contents (instead of ls)
type filename.txt        # View file contents (instead of cat)
find "text" filename.txt # Search in files (instead of grep)
where filename.exe       # Find executable location (instead of which)

# Navigation
cd path\to\directory     # Change directory (backslashes on Windows)
pushd path && popd       # Save/restore directory location
```

## Application Usage
```bash
# Run the built application
./bin/shotgun-cli        # Run the interactive TUI
./bin/shotgun-cli --version  # Show version
./bin/shotgun-cli --help     # Show help
```

## Task Completion Checklist
When completing a development task:
1. Run `npm run format` to format code
2. Run `npm run lint` to check for issues
3. Run `npm test` to ensure all tests pass
4. Test the application with `npm run dev`
5. If making cross-platform changes, test with `npm run build:all`