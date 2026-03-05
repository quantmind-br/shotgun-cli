# App Package - Application Services

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Orchestration layer between CLI/TUI and Core. **No framework dependencies.**

## KEY SERVICES

### ContextService (`service.go`, `service_llm.go`)
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
Unified factory for LLM providers. Stores factory functions, not instances.

```go
// Create provider
provider, err := app.DefaultProviderRegistry.Create(llm.ProviderOpenAI, cfg)

// Register new provider
DefaultProviderRegistry.Register(llm.ProviderMyProvider, func(cfg llm.Config) (llm.Provider, error) {
    return myprovider.NewClient(cfg)
})
```

**Registered**: OpenAI, Anthropic, Gemini (via `init()` in providers.go)

## CONFIG TYPES

| Type | File | Purpose |
|------|------|---------|
| `CLIConfig` | `config.go` | CLI flag parsing, Viper-bound |
| `GenerateConfig` | `service.go` | Context generation parameters |
| `LLMSendConfig` | `service_llm.go` | LLM send parameters (provider, key, model, output) |

## KEY FILES

| File | Purpose |
|------|---------|
| `service.go` | ContextService: Generate, GenerateWithProgress |
| `service_llm.go` | ContextService: SendToLLM*, provider creation |
| `providers.go` | DefaultProviderRegistry, provider init |
| `config.go` | CLIConfig, config-to-service bridging |

## TESTING

```bash
go test -v -race ./internal/app/...
```

**Mock patterns** in `service_test.go`: `mockScanner`, `mockGenerator`, `mockProvider` — all implement core interfaces.

## ANTI-PATTERNS

- Importing `ui` or `cmd` packages
- Using Viper directly (accept config via parameters)
- Creating global service instances
- Skipping progress callbacks in async methods