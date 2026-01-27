# shotgun-cli LLM Provider Architecture

## Overview

shotgun-cli supports three LLM providers through a unified interface. All providers use HTTP APIs and share common base infrastructure.

## Provider Interface

All providers implement `llm.Provider` from `internal/core/llm/provider.go`:

```go
type Provider interface {
    Send(ctx context.Context, content string) (*Result, error)
    SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error)
    Name() string
    IsAvailable() bool
    IsConfigured() bool
    ValidateConfig() error
}
```

## BaseClient Architecture

All HTTP-based providers use a shared `BaseClient` in `internal/platform/llmbase/`:

- **BaseClient**: Implements common Provider methods and HTTP execution
- **Sender Interface**: Provider-specific strategy for request/response handling

Providers embed `*llmbase.BaseClient` and implement `llmbase.Sender`:

```go
type Sender interface {
    BuildRequest(content string) (interface{}, error)
    ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error)
    GetEndpoint() string
    GetHeaders() map[string]string
    NewResponse() interface{}
    GetProviderName() string
}
```

## Supported Providers

### 1. OpenAI (`ProviderOpenAI`)
- **Location**: `internal/platform/openai/`
- **Models**: GPT-4o, GPT-4, o1, o3-mini
- **Default Base URL**: `https://api.openai.com/v1`
- **Configuration**:
  - `llm.provider`: openai
  - `llm.api-key`: Your OpenAI API key
  - `llm.model`: Model name (default: gpt-4o)
  - `llm.base-url`: Optional custom endpoint
  - `llm.timeout`: Request timeout in seconds
- **Get API Key**: https://platform.openai.com/api-keys

### 2. Anthropic (`ProviderAnthropic`)
- **Location**: `internal/platform/anthropic/`
- **Models**: Claude 4 Sonnet (20250514), Claude 3.5 Sonnet
- **Default Base URL**: `https://api.anthropic.com`
- **Configuration**:
  - `llm.provider`: anthropic
  - `llm.api-key`: Your Anthropic API key
  - `llm.model`: Model name (default: claude-sonnet-4-20250514)
  - `llm.base-url`: Optional custom endpoint
  - `llm.timeout`: Request timeout in seconds
- **Get API Key**: https://console.anthropic.com/settings/keys

### 3. Google Gemini API (`ProviderGemini`)
- **Location**: `internal/platform/geminiapi/`
- **Models**: gemini-2.5-flash, gemini-2.0-flash-exp
- **Default Base URL**: `https://generativelanguage.googleapis.com/v1beta`
- **Configuration**:
  - `llm.provider`: gemini
  - `llm.api-key`: Your Gemini API key
  - `llm.model`: Model name (default: gemini-2.5-flash)
  - `llm.timeout`: Request timeout in seconds
- **Get API Key**: https://aistudio.google.com/app/apikey

## Provider Registry

Providers are registered in `internal/app/providers.go`:

```go
var DefaultProviderRegistry = llm.NewRegistry()

func init() {
    DefaultProviderRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
        return openai.NewClient(cfg)
    })
    DefaultProviderRegistry.Register(llm.ProviderAnthropic, func(cfg llm.Config) (llm.Provider, error) {
        return anthropic.NewClient(cfg)
    })
    DefaultProviderRegistry.Register(llm.ProviderGemini, func(cfg llm.Config) (llm.Provider, error) {
        return geminiapi.NewClient(cfg)
    })
}
```

## Configuration Keys

From `internal/config/keys.go`:

```go
const (
    KeyLLMProvider = "llm.provider"
    KeyLLMAPIKey   = "llm.api-key"
    KeyLLMBaseURL  = "llm.base-url"
    KeyLLMModel    = "llm.model"
    KeyLLMTimeout  = "llm.timeout"
)
```

## CLI Commands

### Set Provider
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.provider anthropic
shotgun-cli config set llm.provider gemini
```

### LLM Management Commands
```bash
shotgun-cli llm status   # Show current config and status
shotgun-cli llm doctor   # Run diagnostics
shotgun-cli llm list     # List all supported providers
```

## Using Providers in Code

### Via ContextService (Recommended)
```go
svc := app.NewContextService()
result, err := svc.SendToLLMWithProgress(ctx, content, app.LLMSendConfig{
    Provider:     llm.ProviderOpenAI,
    APIKey:       "your-api-key",
    Model:        "gpt-4o",
    SaveResponse: true,
    OutputPath:   "./response.md",
}, func(stage string) {
    fmt.Printf("Progress: %s\n", stage)
})
```

### Via Registry Directly
```go
cfg := llm.Config{
    Provider: llm.ProviderOpenAI,
    APIKey:   "your-key",
    Model:    "gpt-4o",
}
provider, err := app.DefaultProviderRegistry.Create(cfg.Provider, cfg)
result, err := provider.Send(ctx, content)
```

## Provider Result

All providers return `*llm.Result`:

```go
type Result struct {
    Response    string        // Processed/cleaned response
    RawResponse string        // Raw response from API
    Model       string        // Model used
    Provider    string        // Provider name
    Duration    time.Duration // Execution time
    Usage       *Usage        // Token usage metrics
}
```

## Custom Endpoints

OpenAI-compatible endpoints work with the OpenAI provider:

### OpenRouter
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.base-url https://openrouter.ai/api/v1
shotgun-cli config set llm.api-key YOUR_OPENROUTER_KEY
shotgun-cli config set llm.model anthropic/claude-sonnet-4
```

### Azure OpenAI
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.base-url https://YOUR_RESOURCE.openai.azure.com/openai/deployments/YOUR_DEPLOYMENT
shotgun-cli config set llm.api-key YOUR_AZURE_KEY
```

## Adding a New Provider

1. Create package in `internal/platform/<provider>/`
2. Create struct embedding `*llmbase.BaseClient`
3. Implement `llmbase.Sender` interface (6 methods)
4. Register in `internal/app/providers.go`
5. Add provider constant to `internal/core/llm/provider.go`
