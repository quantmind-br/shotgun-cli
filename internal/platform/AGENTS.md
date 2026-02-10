# Platform Package - Infrastructure

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

External system integrations. **Implements core interfaces.**

## PACKAGES

### llmbase/
Shared base client for HTTP-based LLM providers.

```go
type Sender interface {
    BuildRequest(content string) (interface{}, error)
    ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error)
    GetEndpoint() string
    GetHeaders() map[string]string
    NewResponse() interface{}
    GetProviderName() string
}

// Embed in provider
type MyProvider struct {
    *llmbase.BaseClient
}
```

### http/
Shared JSON HTTP client.

```go
client := platformhttp.NewJSONClient(baseURL, timeout)
var response MyResponse
err := client.PostJSON(ctx, "/endpoint", headers, body, &response)
```

### openai/, anthropic/, geminiapi/
LLM provider implementations.

```go
// Each implements llm.Provider via llmbase.BaseClient
client := openai.NewClient(cfg)
result, err := client.Send(ctx, content)
```

### clipboard/
Cross-platform clipboard access.

```go
err := clipboard.Write(content)
```

## ADDING A NEW PROVIDER

1. Create `internal/platform/<name>/`
2. Define struct embedding `*llmbase.BaseClient`
3. Implement `Sender` interface (6 methods)
4. Create `NewClient(cfg llm.Config)` constructor
5. Register in `internal/app/providers.go`

## PATTERNS

### Error Handling
Map platform errors to domain errors:
```go
func (c *Client) handleError(err error) error {
    if platformhttp.IsHTTPError(err) {
        return fmt.Errorf("API error: %w", err)
    }
    return err
}
```

### Configuration
Providers receive `llm.Config`, extract provider-specific values:
```go
func NewClient(cfg llm.Config) (*Client, error) {
    apiKey := cfg.APIKey
    model := cfg.Model
    // ...
}
```

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Add LLM provider | `platform/<name>/` + `app/providers.go` |
| HTTP client issues | `platform/http/client.go` |
| Provider base logic | `platform/llmbase/base_client.go` |
| Clipboard issues | `platform/clipboard/clipboard.go` |

## TESTING

```bash
go test -v -race ./internal/platform/...
```

## ANTI-PATTERNS

- Importing `app` or `ui` packages
- Direct Viper access (use config passed to constructor)
- Creating providers outside registry
