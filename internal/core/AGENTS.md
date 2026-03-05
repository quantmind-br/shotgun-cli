# Core Package - Domain Logic

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

**Pure business logic. Zero external dependencies.** Only stdlib allowed.

## PACKAGES

### scanner/
Filesystem traversal with layered ignore support. Parallel workers.

```go
type FileNode struct {
    Name, Path, RelPath string
    IsDir               bool
    Children            []*FileNode
    IsGitignored, IsCustomIgnored bool
    Size                int64
    Parent              *FileNode
}

scanner := scanner.NewFilesystemScanner()
tree, err := scanner.Scan(rootPath, config)
```

**ScanConfig**: MaxFileSize, MaxFiles, MaxMemory, Workers (1-32), SkipBinary, IncludeHidden, IncludeIgnored, IgnorePatterns, IncludePatterns, RespectGitignore, RespectShotgunignore.

### contextgen/
Generates LLM context from templates, file tree, and file contents.

```go
generator := contextgen.NewGenerator()
result, err := generator.Generate(cfg)
```

### template/
Template loading from embedded FS + custom paths. Variable substitution with `{VARIABLE_NAME}` pattern.

```go
mgr := template.NewManager(cfg)
tmpl, err := mgr.Load("code-review")
rendered, err := mgr.Render(tmpl, variables)
```

**Variables**: `{TASK}`, `{RULES}`, `{FILE_STRUCTURE}`, `{CURRENT_DATE}` — uppercase, alphanumeric+underscore.

### ignore/
Layered ignore engine. Priority (high→low): explicit excludes → explicit includes → built-in → .gitignore → .shotgunignore → custom.

```go
engine := ignore.NewEngine()
engine.LoadGitignore(rootDir)
shouldIgnore, reason := engine.ShouldIgnore(relPath)
```

### llm/
Provider interface, types, config, registry.

```go
type Provider interface {
    Send(ctx context.Context, content string) (*Result, error)
    SendWithProgress(ctx context.Context, content string, progress func(string)) (*Result, error)
    Name() string
    IsAvailable() bool
    IsConfigured() bool
    ValidateConfig() error
}
```

**ProviderTypes**: `ProviderOpenAI`, `ProviderAnthropic`, `ProviderGemini`

### tokens/
Token estimation (heuristic: 1 token ≈ 4 bytes). No heavy tokenizer dependency.

```go
tokens.Estimate(text)           // From text
tokens.EstimateFromBytes(size)  // From byte count
tokens.FormatTokens(count)      // "32K", "1.2M"
```

### diff/
Intelligent diff splitting at file boundaries.

```go
chunks := diff.IntelligentSplit(diffContent, maxLines)
```

## CRITICAL RULES

1. **No external imports** (except test helpers like testify)
2. **Define interfaces here**, implement in platform/
3. **Config structs** with `DefaultConfig()` method
4. **Error wrapping**: `fmt.Errorf("...: %w", err)`

## TESTING

```bash
go test -v -race ./internal/core/...           # All core
go test -v ./internal/core/scanner/...         # Single package
```

## ANTI-PATTERNS

- Importing from `app`, `platform`, `ui`, or `cmd`
- Using Viper or any config library
- Making HTTP calls directly
- Global state