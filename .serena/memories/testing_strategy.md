# shotgun-cli Testing Strategy

## Test Structure

```
shotgun-cli/
├── cmd/*_test.go                 # CLI command tests
├── internal/
│   ├── core/
│   │   ├── scanner/scanner_test.go
│   │   ├── context/*_test.go
│   │   ├── template/*_test.go
│   │   ├── ignore/engine_test.go
│   │   └── tokens/estimator_test.go
│   ├── ui/
│   │   ├── wizard_test.go
│   │   ├── screens/*_test.go
│   │   └── components/*_test.go
│   └── platform/
│       ├── clipboard/clipboard_test.go
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

### E2E Tests
- Located in `test/e2e/`
- Test actual CLI command execution
- Verify output files and clipboard behavior

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

## Key Test Files

| Package | Test File | Description |
|---------|-----------|-------------|
| `cmd/` | `root_test.go` | Root command tests |
| `cmd/` | `context_test.go` | Context generation CLI |
| `cmd/` | `template_test.go` | Template commands |
| `internal/core/scanner/` | `scanner_test.go` | File scanning tests |
| `internal/core/context/` | `generator_test.go` | Context generation |
| `internal/ui/` | `wizard_test.go` | Wizard model tests |
| `internal/ui/components/` | `tree_test.go` | File tree component |

## Testing Tips

1. **Use fixtures**: `test/fixtures/sample-project/` for real file system tests
2. **Table-driven**: Use for multiple input/output combinations
3. **Mock interfaces**: Scanner, ContextGenerator, TemplateManager
4. **Parallel tests**: Add `t.Parallel()` for independent tests
5. **Coverage focus**: Core logic in `internal/core/` needs high coverage
