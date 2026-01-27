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

# Uninstall from system and GOPATH
make uninstall

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

# Verbose output
go test -v ./...
```

## Code Quality

```bash
# Format code
make fmt

# Run golangci-lint (config: .golangci-local.yml)
make lint

# Run go vet
make vet

# Full quality check
make fmt lint vet test
```

## Coverage Requirements

- **85% minimum** (enforced by CI)
- Target **90%+ for new code**
- Check coverage: `go tool cover -func=coverage.out | grep total`
- HTML report: `go tool cover -html=coverage.out`

## Release

```bash
# Build release artifacts with Goreleaser
make release VERSION=1.2.3

# Create and push tag
make release-tag VERSION=1.2.3
make release-push VERSION=1.2.3

# Test release locally
make release-snapshot
```

## Dependencies

```bash
# Download and verify dependencies
make deps

# Tidy modules
go mod tidy
```

## CLI Commands During Development

### Testing LLM Integration
```bash
# Check LLM status
./build/shotgun-cli llm status

# Run LLM diagnostics
./build/shotgun-cli llm doctor

# List available providers
./build/shotgun-cli llm list

# Send context to LLM
./build/shotgun-cli send context.md --provider openai
```

### Testing Context Generation
```bash
# Generate context interactively (TUI)
./build/shotgun-cli

# Generate context from CLI
./build/shotgun-cli context generate --path . --output context.md

# List templates
./build/shotgun-cli template list
```

### Configuration Management
```bash
# Show all configuration
./build/shotgun-cli config show

# Interactive config TUI
./build/shotgun-cli config

# Set configuration values
./build/shotgun-cli config set llm.provider openai
./build/shotgun-cli config set llm.api-key sk-...
```

## Development Tips

1. **Test fixtures**: Use `test/fixtures/sample-project/` for integration testing
2. **E2E tests**: Located in `test/e2e/`, tests CLI commands
3. **Embedded templates**: Add new templates to `internal/assets/templates/`
4. **Config defaults**: Set in `cmd/root.go` → `setConfigDefaults()`
5. **Provider registry**: Register new providers in `internal/app/providers.go`

## Debugging

### Enable Debug Logging
```bash
SHOTGUN_LOG_LEVEL=debug ./build/shotgun-cli
```

### Run with Race Detector
```bash
make test-race
```

## Environment Variables

All config keys can be set via environment variables with `SHOTGUN_` prefix:

```bash
export SHOTGUN_LLM_PROVIDER=openai
export SHOTGUN_LLM_API_KEY=sk-...
export SHOTGUN_LLM_MODEL=gpt-4o
```

## Key Development Areas

### Adding a New LLM Provider
1. Create provider package in `internal/platform/<provider>/`
2. Embed `*llmbase.BaseClient` and implement `llmbase.Sender` interface
3. Register in `internal/app/providers.go`
4. Add provider constant to `internal/core/llm/provider.go` `AllProviders()`
5. Update documentation

### Adding a New Configuration Key
1. Add constant to `internal/config/keys.go`
2. Add metadata to `internal/config/metadata.go`
3. Add validation to `internal/config/validator.go`
4. Add to default config in `cmd/root.go`

### Adding a New CLI Command
1. Create `cmd/<command>.go`
2. Define cobra command
3. Add to `rootCmd` in `init()`
4. Add tests in `cmd/<command>_test.go`

## Binary Locations

- `build/shotgun-cli` - Local build output
- `$GOPATH/bin/shotgun-cli` - User installation
- `/usr/local/bin/shotgun-cli` - System installation

## Application Layer Service Usage

```go
// Create service with defaults
service := app.NewContextService()

// Create service with custom dependencies
service := app.NewContextService(
    app.WithScanner(customScanner),
    app.WithGenerator(customGenerator),
    app.WithRegistry(customRegistry),
)

// Generate context
result, err := service.Generate(ctx, cfg)

// Send to LLM with progress
result, err := service.SendToLLMWithProgress(ctx, content, llmConfig, progressCallback)
```

## Project Configuration Files

- `.golangci-local.yml` - Local linter configuration
- `.golangci.yml` - CI linter configuration
- `.goreleaser.yaml` - Release automation
- `.shotgunignore` - Patterns to ignore when scanning
- `Makefile` - Build automation
- `go.mod` - Go module definition (Go 1.24)
