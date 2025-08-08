# Task Completion Workflow

## Standard Development Workflow

When completing any development task on this project, follow these steps:

### 1. Code Quality Checks
```bash
# Format code (must be run)
npm run format
# OR: go fmt ./...

# Lint code (must be run)  
npm run lint
# OR: go vet ./...
```

### 2. Testing
```bash
# Run all tests (must pass)
npm test
# OR: go test ./...

# For specific changes, run focused tests
go test -v ./internal/core      # If core logic changed
go test -v ./internal/ui        # If UI components changed
```

### 3. Build Verification
```bash
# Verify local build works
npm run build:local
# OR: go build -o bin/shotgun-cli .

# Test the built binary
./bin/shotgun-cli --version
./bin/shotgun-cli --help
```

### 4. Cross-Platform Considerations
Since this is a cross-platform tool, be aware of:
- **Path handling**: Use `filepath.Join()` for cross-platform paths
- **Platform-specific features**: Check for Windows-specific code (UTF-8 handling, console setup)
- **File permissions**: Consider different permission models across platforms

## Before Committing Code

### Required Checks
1. **All tests pass**: `npm test` or `go test ./...`
2. **Code is formatted**: `npm run format` or `go fmt ./...`  
3. **Code passes linting**: `npm run lint` or `go vet ./...`
4. **Application builds**: `npm run build:local` or `go build`
5. **Application runs**: Basic functionality test

### Optional but Recommended
```bash
# Run benchmarks if performance-related changes
go test -bench=. ./...

# Test with debug mode
DEBUG=1 go run .

# Cross-platform build test (if significant changes)
npm run build:all
```

## Error Handling Standards

### When Adding New Features
- **Use structured errors**: Follow patterns in `translation_errors.go`
- **Add proper validation**: Use validation tags and error reporting
- **Include tests**: Both success and error cases
- **Add logging**: Use appropriate log levels for debugging

### Configuration Changes  
- **Update defaults**: Modify `Default*` functions if needed
- **Add validation**: Include validation rules for new config fields
- **Test migration**: Ensure existing configs still work
- **Update documentation**: Keep CLAUDE.md and README.md current

## Release Readiness

### Before Creating Releases
1. **All quality checks pass**
2. **Cross-platform builds successful**: `npm run build:all`
3. **Manual testing on target platforms**
4. **Documentation updated**
5. **Version numbers updated** (package.json, main.go)

### Platform Testing
- **Windows**: UTF-8 handling, console behavior, path separators
- **macOS**: File permissions, keyring integration
- **Linux**: File permissions, XDG compliance