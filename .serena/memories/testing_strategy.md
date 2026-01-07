# shotgun-cli Testing Strategy

## Test Structure

```
shotgun-cli/
├── cmd/*_test.go                 # CLI command tests
├── internal/
│   ├── app/                      # Application Layer tests (NEW)
│   │   ├── service_test.go
│   │   └── context_test.go
│   ├── config/                   # Configuration tests (NEW)
│   │   ├── keys_test.go
│   │   └── validator_test.go
│   ├── core/
│   │   ├── scanner/scanner_test.go
│   │   ├── context/*_test.go
│   │   ├── template/*_test.go
│   │   ├── ignore/engine_test.go
│   │   ├── tokens/estimator_test.go
│   │   ├── llm/                  # LLM provider tests (NEW)
│   │   │   ├── provider_test.go
│   │   │   ├── config_test.go
│   │   │   └── registry_test.go
│   │   └── diff/split_test.go
│   ├── ui/
│   │   ├── wizard_test.go
│   │   ├── screens/*_test.go
│   │   └── components/*_test.go
│   └── platform/
│       ├── clipboard/clipboard_test.go
│       ├── openai/client_test.go      # NEW
│       ├── anthropic/client_test.go   # NEW
│       ├── geminiapi/client_test.go   # NEW
│       └── gemini/gemini_test.go
├── test/
│   ├── e2e/                      # End-to-end CLI tests
│   │   ├── cli_test.go
│   │   └── filestructure_test.go
│   └── fixtures/
│       └── sample-project/       # Test fixture directory
```

## Test Commands

```bash
# Run all unit tests
make test

# Run with race detector
make test-race

# Run benchmarks
make test-bench

# Run E2E tests
make test-e2e

# Generate coverage
make coverage

# Run specific package tests
go test ./internal/core/scanner/...
go test ./internal/app/...
go test ./internal/core/llm/...
go test ./internal/platform/openai/...

# Run single test
go test -run TestFunctionName ./path/to/package

# Verbose output
go test -v ./...
```

## Test Fixtures

Location: `test/fixtures/sample-project/`

A comprehensive sample project structure including:
- Go source files with tests
- Configuration files (YAML, TOML, JSON)
- Documentation
- Scripts
- Various file types for scanner testing

## Testing Patterns

### Unit Tests
- Located alongside source files (`*_test.go`)
- Use `testify/assert` for assertions
- Table-driven tests for multiple cases

### Application Layer Tests
Located in `internal/app/`:
- Test service orchestration
- Test config validation
- Test error handling across layers
- Example: `service_test.go` tests `ContextService.Generate()`

### Configuration Tests
Located in `internal/config/`:
- Test key constants
- Test validation logic
- Test default values

### LLM Provider Tests
Located in `internal/core/llm/` and `internal/platform/*/`:
- Test provider interface implementation
- Test config validation
- Test provider availability checks
- Mock API responses for unit tests

### E2E Tests
Located in `test/e2e/`:
- Test actual CLI command execution
- Verify output files and clipboard behavior
- Test full workflows

### Bubble Tea UI Tests
- Test model updates with simulated messages
- Verify state changes and view rendering
- Example pattern:
```go
func TestFileTreeNavigation(t *testing.T) {
    model := NewFileTree(root)
    model.MoveDown()
    assert.Equal(t, 1, model.cursor)
}
```

## Coverage Requirements

Per project guidelines:
- **80% minimum coverage** for new features
- Generate coverage with `make coverage`
- Coverage output: `coverage.out`
- View HTML coverage: `go tool cover -html=coverage.out`

## Key Test Files

| Package | Test File | Description |
|---------|-----------|-------------|
| `cmd/` | `root_test.go` | Root command tests |
| `cmd/` | `context_test.go` | Context generation CLI |
| `cmd/` | `template_test.go` | Template commands |
| `cmd/` | `llm_test.go` | LLM commands (NEW) |
| `internal/app/` | `service_test.go` | ContextService tests (NEW) |
| `internal/config/` | `keys_test.go` | Config keys (NEW) |
| `internal/core/scanner/` | `scanner_test.go` | File scanning tests |
| `internal/core/context/` | `generator_test.go` | Context generation |
| `internal/core/llm/` | `provider_test.go` | LLM interface (NEW) |
| `internal/platform/openai/` | `client_test.go` | OpenAI tests (NEW) |
| `internal/ui/` | `wizard_test.go` | Wizard model tests |
| `internal/ui/components/` | `tree_test.go` | File tree component |

## Testing Tips

1. **Use fixtures**: `test/fixtures/sample-project/` for real file system tests
2. **Table-driven**: Use for multiple input/output combinations
3. **Mock interfaces**: Scanner, ContextGenerator, TemplateManager, Provider
4. **Parallel tests**: Add `t.Parallel()` for independent tests
5. **Coverage focus**: Core logic in `internal/core/` needs high coverage

## Testing LLM Providers

### Unit Testing Providers
```go
func TestOpenAIProvider(t *testing.T) {
    // Mock HTTP responses
    // Test Send() method
    // Test ValidateConfig()
    // Test IsAvailable()
}
```

### Integration Testing Providers
Use real API keys in environment variables for integration tests:
```bash
export TEST_OPENAI_API_KEY=sk-test-...
go test ./internal/platform/openai/ -run TestIntegration
```

## Mocking External Dependencies

For tests that don't require real API calls:

```go
type MockProvider struct {
    llm.Provider
    mockResponse string
    mockError    error
}

func (m *MockProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
    if m.mockError != nil {
        return nil, m.mockError
    }
    return &llm.Result{Response: m.mockResponse}, nil
}
```

## Coverage Reporting

```bash
# Generate coverage
make coverage

# View in browser
go tool cover -html=coverage.out

# Check specific package coverage
go test -coverprofile=cover.out ./internal/app/...
go tool cover -func=cover.out
```