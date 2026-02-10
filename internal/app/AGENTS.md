# App Package - Application Services

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Orchestration layer between CLI/TUI and Core. **No framework dependencies.**

## KEY SERVICES

### ContextService (`context.go`)
Main API for context generation and LLM operations.

```go
svc := app.NewContextService()

// Synchronous (CLI)
result, err := svc.Generate(ctx, cfg)

// Async with progress (TUI)
result, err := svc.GenerateWithProgress(ctx, cfg, func(stage string) {
    // Update UI
})

// Send to LLM
result, err := svc.SendToLLM(ctx, content, provider)
result, err := svc.SendToLLMWithProgress(ctx, content, config, callback)
```

### ProviderRegistry (`providers.go`)
Unified factory for LLM providers.

```go
// Create provider
provider, err := app.DefaultProviderRegistry.Create(llm.ProviderOpenAI, cfg)

// Register new provider
DefaultProviderRegistry.Register(llm.ProviderMyProvider, func(cfg llm.Config) (llm.Provider, error) {
    return myprovider.NewClient(cfg)
})
```

## CONFIG TYPES

| Type | Purpose |
|------|---------|
| `CLIConfig` | CLI flag parsing |
| `GenerateConfig` | Service layer configuration |
| `LLMSendConfig` | LLM operation parameters |

## PATTERNS

### Service Initialization
```go
func NewContextService() *ContextService {
    return &ContextService{
        scanner:       scanner.NewFilesystemScanner(),
        generator:     contextgen.NewGenerator(),
        templateMgr:   template.NewManager(),
        providerReg:   DefaultProviderRegistry,
    }
}
```

### Progress Callbacks
All async operations support progress reporting:
```go
func (s *ContextService) GenerateWithProgress(
    ctx context.Context,
    cfg GenerateConfig,
    progress func(stage string),
) (*GenerateResult, error)
```

## WHERE TO LOOK

| Task | File |
|------|------|
| Context generation | `context.go` |
| LLM operations | `context.go` (SendToLLM*) |
| Provider registry | `providers.go` |
| Config types | `config.go` |

## TESTING

```bash
go test -v -race ./internal/app/...
```

## ANTI-PATTERNS

- Importing `ui` or `cmd` packages
- Using Viper directly (accept config via parameters)
- Creating global service instances
