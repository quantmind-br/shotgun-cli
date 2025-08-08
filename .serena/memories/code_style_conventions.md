# Code Style and Conventions

## Go Code Style
- **Standard Go formatting**: Uses `go fmt` for consistent formatting
- **Standard Go linting**: Uses `go vet` for static analysis
- **Package naming**: Short, lowercase, single words (e.g., `core`, `ui`)
- **Struct tags**: JSON tags for serialization (`json:"name"`, `json:"children,omitempty"`)

## Naming Conventions

### Files
- **Snake case**: e.g., `custom_templates.go`, `enhanced_config.go`
- **Test files**: `*_test.go` suffix
- **Clear purpose**: File names describe their primary function

### Types and Structures
- **PascalCase**: `FileNode`, `DirectoryScanner`, `TemplateProcessor`
- **Descriptive names**: Names clearly indicate purpose
- **Interface suffix**: Interfaces often end with "Interface" (e.g., `ConfigManagerInterface`)

### Functions and Methods
- **camelCase**: `generateTreeRecursive`, `buildTreeRecursive`  
- **Receiver naming**: Single letter or short abbreviation (e.g., `(*ConfigManager).Load`)
- **Constructor pattern**: `New` prefix for constructors (e.g., `NewConfigManager`)

### Constants
- **PascalCase**: `TemplateDevKey`, `StatusExcluded`
- **Grouped by purpose**: Related constants grouped together

## Documentation Style
- **Comments on exported types**: All exported functions, types, and methods have comments
- **Package documentation**: Key packages have package-level documentation
- **Error handling**: Explicit error handling with descriptive error messages

## Error Handling Patterns
- **Custom error types**: `ShotgunError`, `TranslationError` with structured information
- **Error wrapping**: Uses Go's error wrapping patterns
- **Validation errors**: Structured validation with field-specific errors
- **Circuit breaker integration**: Resilient error handling for external services

## Testing Patterns
- **Table-driven tests**: Common pattern for testing multiple scenarios
- **Benchmark tests**: Performance testing with `Benchmark*` functions  
- **Mock interfaces**: Interface-based mocking for testing
- **Test helpers**: Reusable test setup functions (e.g., `createFastTestConfig`)

## Configuration Patterns
- **Struct tags**: Validation tags on configuration structs
- **Default values**: `Default*` functions provide sensible defaults
- **Environment overrides**: Configuration can be overridden by environment variables
- **Validation**: Configuration validation with detailed error reporting