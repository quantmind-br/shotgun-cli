# shotgun-cli Development Workflow

## Build Commands

```bash
# Build for current platform (output: build/shotgun-cli)
make build

# Cross-compile for all platforms (linux/darwin/windows, amd64/arm64)
make build-all

# Install to GOPATH/bin
make install

# Install system-wide (requires sudo)
make install-system

# Clean build artifacts
make clean
```

## Testing Commands

```bash
# Run unit tests
make test

# Run tests with race detector
make test-race

# Run benchmarks
make test-bench

# Run end-to-end tests
make test-e2e

# Generate coverage report
make coverage

# Run tests for specific package
go test ./internal/core/scanner/...

# Run single test
go test -run TestFunctionName ./pkg
```

## Code Quality

```bash
# Format code
make fmt

# Run golangci-lint (config: .golangci.yml)
make lint

# Run go vet
make vet

# Full quality check
make fmt lint vet test
```

## Release

```bash
# Build release artifacts with Goreleaser
make release
```

## Project Configuration Files

- `.golangci.yml` - Linter configuration
- `.goreleaser.yaml` - Release automation
- `.shotgunignore` - Patterns to ignore when using shotgun-cli on itself
- `Makefile` - Build automation

## Dependencies

```bash
# Download and verify dependencies
make deps
```

## Key Go Commands

```bash
# Generate (if using go:generate)
make generate

# Verify module checksums
go mod verify

# Tidy modules
go mod tidy
```

## Development Tips

1. **Test fixtures**: Use `test/fixtures/sample-project/` for integration testing
2. **E2E tests**: Located in `test/e2e/`, tests CLI commands
3. **Embedded templates**: Add new templates to `internal/assets/templates/`
4. **Config defaults**: Set in `cmd/root.go` → `setConfigDefaults()`

## Binary Locations

- `build/shotgun-cli` - Local build output
- `$GOPATH/bin/shotgun-cli` - User installation
- `/usr/local/bin/shotgun-cli` - System installation
