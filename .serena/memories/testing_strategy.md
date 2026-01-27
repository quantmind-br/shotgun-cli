# shotgun-cli Testing Strategy

## Test Structure

```
shotgun-cli/
├── cmd/*_test.go                 # CLI command tests
├── internal/
│   ├── app/                      # Application Layer tests
│   │   ├── service_test.go
│   │   ├── context_test.go
│   │   ├── config_test.go
│   │   ├── service_llm_test.go
│   │   └── integration_test.go
│   │
│   ├── config/                   # Configuration tests
│   │   ├── keys_test.go
│   │   ├── validator_test.go
│   │   └── metadata_test.go
│   │
│   ├── core/
│   │   ├── scanner/scanner_test.go
│   │   ├── context/*_test.go
│   │   ├── template/*_test.go
│   │   ├── ignore/engine_test.go
│   │   ├── tokens/estimator_test.go
│   │   ├── llm/
│   │   │   ├── provider_test.go
│   │   │   ├── config_test.go
│   │   │   └── registry_test.go
│   │   └── diff/split_test.go
│   │
│   ├── ui/
│   │   ├── wizard_test.go
│   │   ├── config_wizard_test.go
│   │   ├── scan_coordinator_test.go
│   │   ├── generate_coordinator_test.go
│   │   ├── screens/*_test.go
│   │   └── components/*_test.go
│   │
│   └── platform/
│       ├── openai/client_test.go
│       ├── openai/models_test.go
│       ├── anthropic/client_test.go
│       ├── geminiapi/client_test.go
│       └── clipboard/clipboard_test.go
│
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

## Coverage Requirements

Per project guidelines (AGENTS.md):
- **85% minimum coverage** (enforced by CI)
- Target **90%+ for new code**
- Generate coverage with `make coverage`
- Coverage output: `coverage.out`
- View HTML coverage: `go tool cover -html=coverage.out`
- Check total: `go tool cover -func=coverage.out | grep total`

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
- Use `testify/assert` and `testify/require` for assertions
- Table-driven tests for multiple cases
- Use `t.Parallel()` for independent tests

### Example Test Pattern
```go
func TestMyFunction_ValidInput(t *testing.T) {
    t.Parallel()
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {name: "basic case", input: "test", expected: "result"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            result, err := MyFunction(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Application Layer Tests
Located in `internal/app/`:
- Test service orchestration
- Test LLM send operations (`service_llm_test.go`)
- Test integration flows (`integration_test.go`)

### Configuration Tests
Located in `internal/config/`:
- Test key constants
- Test validation logic for all types
- Test metadata registration

### LLM Provider Tests
Located in `internal/platform/*/`:
- Test provider interface implementation
- Mock HTTP responses for unit tests
- Test error handling paths

### TUI Coordinator Tests
Located in `internal/ui/`:
- `scan_coordinator_test.go` - Tests scanning state machine
- `generate_coordinator_test.go` - Tests generation state machine

### E2E Tests
Located in `test/e2e/`:
- Test actual CLI command execution
- Verify output files and clipboard behavior
- Test full workflows

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

## Key Test Files

| Package | Test File | Description |
|---------|-----------|-------------|
| `cmd/` | `root_test.go` | Root command tests |
| `cmd/` | `context_test.go` | Context generation CLI |
| `cmd/` | `llm_test.go` | LLM commands |
| `cmd/` | `send_test.go` | Send command |
| `internal/app/` | `service_test.go` | ContextService tests |
| `internal/app/` | `service_llm_test.go` | LLM service tests |
| `internal/config/` | `validator_test.go` | Config validation |
| `internal/config/` | `metadata_test.go` | Config metadata |
| `internal/core/llm/` | `registry_test.go` | Provider registry |
| `internal/platform/openai/` | `client_test.go` | OpenAI client |
| `internal/ui/` | `wizard_test.go` | Wizard model |
| `internal/ui/` | `scan_coordinator_test.go` | Scan coordinator |

## Testing Tips

1. **Use fixtures**: `test/fixtures/sample-project/` for file system tests
2. **Table-driven**: Use for multiple input/output combinations
3. **Mock interfaces**: Scanner, ContextGenerator, TemplateManager, Provider
4. **Parallel tests**: Add `t.Parallel()` for independent tests
5. **Coverage focus**: Core logic in `internal/core/` needs high coverage
