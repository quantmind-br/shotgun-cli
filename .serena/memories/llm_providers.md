# shotgun-cli LLM Provider Architecture

## Overview

shotgun-cli supports multiple LLM providers through a unified interface. The provider abstraction allows easy switching between different AI services while maintaining consistent behavior across the application.

## Provider Interface

All providers implement the `llm.Provider` interface defined in `internal/core/llm/provider.go`:

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

## Supported Providers

### 1. OpenAI (`ProviderOpenAI`)
- **Location**: `internal/platform/openai/`
- **Models**: GPT-4o, GPT-4, o1, o3-mini
- **API Base**: `https://api.openai.com/v1` (default)
- **Configuration**:
  - `llm.provider`: openai
  - `llm.api-key`: Your OpenAI API key
  - `llm.model`: Model name (default: gpt-4o)
  - `llm.base-url`: Optional custom endpoint (for OpenRouter, Azure, etc.)
  - `llm.timeout`: Request timeout in seconds
- **Get API Key**: https://platform.openai.com/api-keys

### 2. Anthropic (`ProviderAnthropic`)
- **Location**: `internal/platform/anthropic/`
- **Models**: Claude 4 Sonnet (20250514), Claude 3.5 Sonnet
- **API Base**: `https://api.anthropic.com` (default)
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
- **API Base**: `https://generativelanguage.googleapis.com/v1beta` (default)
- **Configuration**:
  - `llm.provider`: gemini
  - `llm.api-key`: Your Gemini API key
  - `llm.model`: Model name (default: gemini-2.5-flash)
  - `llm.timeout`: Request timeout in seconds
- **Get API Key**: https://aistudio.google.com/app/apikey

### 4. GeminiWeb (`ProviderGeminiWeb`)
- **Location**: `internal/platform/gemini/`
- **Description**: Browser-based integration using the `geminiweb` CLI tool
- **Configuration**:
  - `llm.provider`: geminiweb
  - `gemini.enabled`: true
  - `gemini.binary-path`: Path to geminiweb binary
- **Setup**:
  ```bash
  go install github.com/diogo/geminiweb/cmd/geminiweb@latest
  geminiweb auto-login
  ```
- **Note**: This provider doesn't require an API key but requires the geminiweb tool

## Provider Registry

Providers are registered in `cmd/providers.go`:

```go
providerRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
    return openai.NewClient(cfg)
})
```

## Configuration Keys

All provider configuration uses the `internal/config/keys.go` constants:

```go
const (
    KeyLLMProvider   = "llm.provider"
    KeyLLMAPIKey     = "llm.api-key"
    KeyLLMBaseURL    = "llm.base-url"
    KeyLLMModel      = "llm.model"
    KeyLLMTimeout    = "llm.timeout"
    // ...
)
```

## Configuration Commands

### Set Provider
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.provider anthropic
shotgun-cli config set llm.provider gemini
shotgun-cli config set llm.provider geminiweb
```

### Set API Key
```bash
shotgun-cli config set llm.api-key YOUR_API_KEY
```

### Set Model
```bash
shotgun-cli config set llm.model gpt-4o
```

### Set Custom Base URL (for OpenRouter, Azure, etc.)
```bash
shotgun-cli config set llm.base-url https://openrouter.ai/api/v1
```

## LLM Management Commands

### Check Status
```bash
shotgun-cli llm status
```
Shows current provider, model, API key status, and readiness.

### Run Diagnostics
```bash
shotgun-cli llm doctor
```
Comprehensive diagnostics with specific guidance for fixing issues.

### List Providers
```bash
shotgun-cli llm list
```
Shows all supported providers with descriptions.

## Usage in Code

### Creating a Provider
```go
import "github.com/quantmind-br/shotgun-cli/cmd"

cfg := BuildLLMConfig()
provider, err := CreateLLMProvider(cfg)
```

### Using ContextService
```go
import "github.com/quantmind-br/shotgun-cli/internal/app"

service := app.NewContextService()
result, err := service.SendToLLM(ctx, content, provider)
```

## Provider Result

All providers return a `llm.Result`:

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

The OpenAI-compatible interface allows using custom endpoints:

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

## Testing Provider Configuration

```bash
# Check if configured correctly
shotgun-cli llm status

# Run diagnostics
shotgun-cli llm doctor

# Test with actual send
shotgun-cli send --provider openai myfile.md
```