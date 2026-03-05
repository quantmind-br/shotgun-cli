# Platform Package - Infrastructure

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

External system integrations. **Implements core interfaces via Strategy pattern.**

## PACKAGES

### llmbase/
Shared base client for HTTP-based LLM providers. Implements common `llm.Provider` methods (Name, IsAvailable, IsConfigured, ValidateConfig, Send, SendWithProgress).

```go
// Sender interface — 6 methods each provider must implement
type Sender interface {
    BuildRequest(content string) (interface{}, error)
    ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error)
    GetEndpoint() string
    GetHeaders() map[string]string
    NewResponse() interface{}
    GetProviderName() string
}
```

### http/
Shared JSON HTTP client. All providers use this.

```go
client := platformhttp.NewJSONClient(baseURL, timeout)
err := client.PostJSON(ctx, "/endpoint", headers, body, &response)
```

Returns `HTTPError` with `StatusCode` and `Body` for provider-specific error parsing.

### openai/
- BaseURL: `https://api.openai.com/v1` | Model: `gpt-4o`
- Endpoint: `/chat/completions` | Auth: `Authorization: Bearer {key}`

### anthropic/
- BaseURL: `https://api.anthropic.com` | Model: `claude-sonnet-4-20250514`
- Endpoint: `/v1/messages` | Auth: `x-api-key`, `anthropic-version: 2023-06-01`
- Default MaxTokens: 8192

### geminiapi/
- BaseURL: `https://generativelanguage.googleapis.com/v1beta` | Model: `gemini-2.5-flash`
- Endpoint: `/models/{model}:generateContent?key={key}` (auth via query param)
- Default MaxTokens: 8192

### clipboard/
Cross-platform clipboard via `atotto/clipboard`. `Copy(content)`, `IsAvailable()`.

## ADDING A NEW PROVIDER

1. Create `internal/platform/<name>/`
2. Embed `*llmbase.BaseClient` in struct
3. Implement `Sender` interface (6 methods)
4. Create `NewClient(cfg llm.Config)` constructor
5. Register in `internal/app/providers.go`

## ERROR HANDLING PATTERN

```go
func (c *Client) handleError(err error) error {
    var httpErr *platformhttp.HTTPError
    if errors.As(err, &httpErr) {
        // Parse provider-specific error from httpErr.Body
        return fmt.Errorf("API error [%d]: %s", httpErr.StatusCode, message)
    }
    return err
}
```

## TESTING

```bash
go test -v -race ./internal/platform/...
```

Provider tests use `httptest.NewServer` for HTTP mocking. No external API calls in tests.

## ANTI-PATTERNS

- Importing `app` or `ui` packages
- Direct Viper access (use config passed to constructor)
- Creating providers outside registry
- Skipping `handleError()` when processing HTTP responses