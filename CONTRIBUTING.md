# Contributing to Shotgun CLI

Thank you for your interest in contributing to Shotgun CLI! This document provides guidelines and best practices for contributors.

## Table of Contents

- [Development Setup](#development-setup)
- [Testing Guidelines](#testing-guidelines)
- [Code Coverage](#code-coverage)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [CI/CD Pipeline](#cicd-pipeline)

## Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/quantmind-br/shotgun-cli.git
   cd shotgun-cli
   ```

2. **Install Go** (version 1.23 or later):
   ```bash
   # Check your Go version
   go version
   ```

3. **Install dependencies**:
   ```bash
   go mod download
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

## Testing Guidelines

### Writing Tests

- **Use table-driven tests** for functions with multiple input/output scenarios
- **Use `t.Parallel()`** for independent tests to improve execution speed
- **Follow the naming convention**: `TestFunctionName_Scenario`
- **Test both success and error paths**

### Example Test Structure

```go
func TestMyFunction_ValidInput(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "basic case",
            input:    "test",
            expected: "result",
            wantErr:  false,
        },
        // Add more test cases...
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

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Code Coverage

### Coverage Requirements

- **Minimum threshold**: 85% overall coverage
- **Target**: 90% coverage for new code
- **Critical packages**: Core business logic should have 95%+ coverage

### Checking Coverage Locally

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Check specific package coverage
go test -cover ./internal/core/scanner/...
```

### Coverage Best Practices

1. **Test all public functions** in core packages
2. **Cover error paths** - don't just test the happy path
3. **Test boundary conditions** for numeric inputs
4. **Mock external dependencies** to isolate unit tests
5. **Use integration tests** for end-to-end scenarios

### Package Coverage Guidelines

| Package | Target Coverage | Priority |
|---------|----------------|----------|
| `internal/core/*` | 90%+ | Critical |
| `internal/app` | 85%+ | High |
| `internal/config` | 85%+ | High |
| `cmd` | 80%+ | Medium |
| `internal/ui` | 75%+ | Medium |
| `internal/platform` | 70%+ | Low |

## Pull Request Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes** following the code style guidelines

3. **Write tests** for new functionality

4. **Ensure all tests pass**:
   ```bash
   go test -race ./...
   ```

5. **Check coverage** meets requirements:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out | grep total
   ```

6. **Run linter**:
   ```bash
   golangci-lint run
   ```

7. **Create a pull request** with:
   - Clear description of changes
   - Reference to related issues
   - Test coverage information

### PR Checklist

- [ ] Tests added for new functionality
- [ ] All tests pass locally
- [ ] Coverage meets or exceeds 85%
- [ ] No linter warnings
- [ ] Documentation updated if needed
- [ ] Commit messages are clear and descriptive

## Code Style

### General Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Keep functions focused and small
- Document exported functions and types

### Project-Specific Patterns

- **Clean Architecture**: Respect layer boundaries (cmd -> app -> core)
- **Dependency Injection**: Pass dependencies explicitly, avoid globals
- **Error Handling**: Return errors, don't panic
- **Logging**: Use zerolog for structured logging

## CI/CD Pipeline

### Automated Checks

Every pull request runs:

1. **Tests** with race detection and coverage
2. **Coverage upload** to Codecov
3. **Coverage threshold check** (85% minimum)
4. **Linting** with golangci-lint
5. **Build verification**

### Coverage Enforcement

- PRs that drop coverage below 85% will fail CI
- Coverage reports are visible in Codecov
- Badge in README shows current coverage status

### Fixing CI Failures

1. **Test failures**: Run tests locally with `-v` flag
2. **Coverage below threshold**: Add tests for uncovered code
3. **Lint errors**: Run `golangci-lint run` locally
4. **Build failures**: Check for compilation errors

## Questions?

If you have questions about contributing, please open an issue or reach out to the maintainers.
