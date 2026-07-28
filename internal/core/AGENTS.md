# Core Package - Domain Logic

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

**Pure business logic. Zero external dependencies.** Only stdlib allowed.

## PACKAGES

### scanner/
Filesystem traversal with layered ignore support. Single-pass, sequential walk.

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

**ScanConfig**: MaxFileSize, MaxFiles, SkipBinary, IncludeHidden, IncludeIgnored, IgnorePatterns, IncludePatterns, RespectGitignore, RespectShotgunignore.

### contextgen/
Generates LLM context from templates, file tree, and file contents.

```go
generator := contextgen.NewGenerator()
result, err := generator.Generate(cfg)
```

**GenerateConfig**: MaxFileSize, MaxTotalSize, MaxFiles, SkipBinary, TemplateVars, Template, IncludeTree, IncludeSummary, IncludeIgnored.

### Zero means "no limit"

Both `ScanConfig` and `GenerateConfig` treat a zero limit as **no limit**. Never
substitute a default for a caller's zero: doing so turns a missing field into a
policy nobody chose, and `MaxFiles`/`MaxTotalSize` abort generation rather than
degrade. Callers wanting the documented ceilings ask for them:

```go
cfg := contextgen.DefaultGenerateConfig() // 10MB / 10MB / 1000 files
```

**Front ends do not build a `GenerateConfig` directly** — `app.BuildGeneratorConfig`
is the only producer, so the TUI and the headless path cannot drift apart.

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