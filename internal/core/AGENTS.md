# Core Package - Domain Logic

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

**Pure business logic. Zero external dependencies.** Only stdlib allowed.

## PACKAGES

### scanner/
Filesystem traversal with ignore support.

```go
type FileNode struct {
    Name     string
    Path     string
    IsDir    bool
    Children []*FileNode
}

scanner := scanner.NewFilesystemScanner()
tree, err := scanner.Scan(rootPath, config)
```

### contextgen/
Generates LLM context from templates and files.

```go
generator := contextgen.NewGenerator()
context, err := generator.Generate(cfg)
```

### template/
Template loading and rendering with variable substitution.

```go
mgr := template.NewManager(cfg)
tmpl, err := mgr.Load("code-review")
rendered, err := mgr.Render(tmpl, variables)
```

### ignore/
Layered ignore rule processing (.gitignore + .shotgunignore).

```go
engine := ignore.NewEngine()
engine.AddRules(patterns)
if engine.ShouldIgnore(path) { ... }
```

### llm/
Provider interfaces and types.

```go
type Provider interface {
    Name() string
    Send(ctx context.Context, content string) (*Result, error)
    IsAvailable() bool
    IsConfigured() bool
}
```

### tokens/
Token estimation for context size validation.

```go
estimate := tokens.Estimate(text)
count := estimate.Tokens
```

### diff/
Intelligent diff splitting for large changes.

```go
chunks := diff.IntelligentSplit(diffContent, maxSize)
```

## CRITICAL RULES

1. **No external imports** (except test helpers)
2. **Define interfaces here**, implement in platform/
3. **Config structs** with `DefaultConfig()` method
4. **Error wrapping** with context: `fmt.Errorf("...: %w", err)`

## INTERFACE DEFINITIONS

| Interface | Implementations |
|-----------|----------------|
| `Scanner` | `FilesystemScanner` (platform) |
| `Provider` | `openai.Client`, `anthropic.Client`, `geminiapi.Client` |
| `ContextGenerator` | `DefaultContextGenerator` |

## TESTING

```bash
# All core tests
go test -v -race ./internal/core/...

# Specific package
go test -v ./internal/core/scanner/...
```

## ANTI-PATTERNS

- Importing from `app`, `platform`, `ui`, or `cmd`
- Using Viper or any config library
- Making HTTP calls directly
- Global state
