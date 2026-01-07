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
go test ./internal/app/...
go test ./internal/core/llm/...

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

# Tidy modules
go mod tidy
```

## Key Go Commands

```bash
# Generate (if using go:generate)
make generate

# Verify module checksums
go mod verify
```

## CLI Commands During Development

### Testing LLM Integration
```bash
# Check LLM status
./shotgun-cli llm status

# Run LLM diagnostics
./shotgun-cli llm doctor

# List available providers
./shotgun-cli llm list

# Send context to LLM
./shotgun-cli send context.md --provider openai
```

### Testing Context Generation
```bash
# Generate context interactively
./shotgun-cli

# Generate context from CLI
./shotgun-cli context generate --path . --output context.md

# List templates
./shotgun-cli template list
```

### Configuration Management
```bash
# Show all configuration
./shotgun-cli config show

# Set configuration values
./shotgun-cli config set llm.provider openai
./shotgun-cli config set llm.api-key sk-...
```

## Development Tips

1. **Test fixtures**: Use `test/fixtures/sample-project/` for integration testing
2. **E2E tests**: Located in `test/e2e/`, tests CLI commands
3. **Embedded templates**: Add new templates to `internal/assets/templates/`
4. **Config defaults**: Set in `cmd/root.go` → `setConfigDefaults()`
5. **Provider registry**: Add new providers in `cmd/providers.go`

## Debugging

### Enable Debug Logging
```bash
SHOTGUN_LOG_LEVEL=debug ./shotgun-cli
```

### Test with Specific Directory
```bash
cd /path/to/test/directory
./shotgun-cli
```

### Run with Race Detector
```bash
make test-race
```

### Generate Coverage Report
```bash
make coverage
go tool cover -html=coverage.out
```

## Key Development Areas

### Adding a New LLM Provider
1. Create provider package in `internal/platform/<provider>/`
2. Implement `llm.Provider` interface
3. Register in `cmd/providers.go`
4. Add to `internal/core/llm/provider.go` `AllProviders()`
5. Update documentation

### Adding a New Configuration Key
1. Add constant to `internal/config/keys.go`
2. Add to default config in `cmd/root.go`
3. Update documentation

### Adding a New CLI Command
1. Create `cmd/<command>.go`
2. Define cobra command
3. Add to `rootCmd` in `init()`
4. Add tests in `cmd/<command>_test.go`

## Binary Locations

- `build/shotgun-cli` - Local build output
- `$GOPATH/bin/shotgun-cli` - User installation
- `/usr/local/bin/shotgun-cli` - System installation

## Application Layer Service

When working with the application layer (`internal/app/`):

```go
// Create service with defaults
service := app.NewContextService()

// Create service with custom dependencies
service := app.NewContextService(
    app.WithScanner(customScanner),
    app.WithGenerator(customGenerator),
)

// Generate context
result, err := service.Generate(ctx, cfg)

// Send to LLM
llmResult, err := service.SendToLLM(ctx, content, provider)
```

## Environment Variables

All config keys can be set via environment variables with `SHOTGUN_` prefix:

```bash
export SHOTGUN_LLM_PROVIDER=openai
export SHOTGUN_LLM_API_KEY=sk-...
export SHOTGUN_LLM_MODEL=gpt-4o
```