# Suggested Commands

## Development Commands

### Build and Run
```bash
# Development mode (run without building)
npm run dev
# OR directly: go run .

# Build for local development
npm run build:local
# OR directly: go build -o bin/shotgun-cli .

# Cross-platform builds
npm run build          # Current platform
npm run build:all      # All platforms (Windows, Linux, macOS, macOS ARM64)
npm run build:windows  # Windows only
npm run build:linux    # Linux only  
npm run build:macos    # macOS only

# Clean build artifacts
npm run clean
```

### Testing and Quality Assurance
```bash
# Run all tests
npm test
# OR directly: go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test -v ./internal/core
go test -v ./internal/ui

# Run benchmarks
go test -bench=. ./...

# Code quality checks
npm run lint           # Go vet linting
# OR directly: go vet ./...

npm run format         # Go formatting
# OR directly: go fmt ./...
```

### Debug and Development
```bash
# Run with debug logging enabled
DEBUG=1 go run .
# OR: DEBUG=1 npm run dev

# The debug log will be written to debug.log
```

## Windows-Specific Commands

Since this is a Windows development environment:

```cmd
# Windows equivalent commands
dir                    # List directory contents (ls equivalent)
cd                     # Change directory
findstr               # Search in files (grep equivalent) 
where                  # Find executable (which equivalent)
type                   # Display file contents (cat equivalent)

# Git operations (standard across platforms)
git status
git log --oneline
git diff
git add .
git commit -m "message"
```

## Application Usage Commands

```bash
# Run the application
shotgun-cli

# Show version
shotgun-cli --version

# Show help
shotgun-cli --help
# OR: shotgun-cli -h
```

## Package Management
```bash
# Install dependencies
go mod tidy

# Add new dependency
go get github.com/example/package

# Update dependencies
go get -u ./...

# Global npm installation (for end users)
npm install -g shotgun-cli

# Local development installation
npm install -g .
```