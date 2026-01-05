# Plano de Implementacao: Integracao Multi-Provider LLM

## Sumario Executivo

Este plano detalha a implementacao de uma arquitetura multi-provider para o shotgun-cli, permitindo que usuarios escolham entre diferentes provedores de LLM:

1. **OpenAI** - API compativel com OpenAI (GPT-4, GPT-4o, etc.)
2. **Anthropic** - API Claude (Claude 3.5 Sonnet, Claude 3 Opus, etc.)
3. **Google Gemini** - API Google Generative AI (Gemini 2.5, etc.)
4. **GeminiWeb** (legado) - Integracao via binario externo `geminiweb`

A arquitetura permitira:
- Configuracao flexivel de `base-url`, `api-key` e `model` por provider
- Facil adicao de novos providers no futuro
- Suporte a endpoints customizados (OpenRouter, Azure, proxies locais)
- Migracao transparente do sistema atual (GeminiWeb)

---

## Arquitetura de Alto Nivel

```
                    ┌─────────────────────────────────────────────────────┐
                    │                    cmd/send.go                       │
                    │                  cmd/providers.go                    │
                    └─────────────────────┬───────────────────────────────┘
                                          │
                                          ▼
                    ┌─────────────────────────────────────────────────────┐
                    │              internal/core/llm                       │
                    │  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
                    │  │ Provider │  │  Config  │  │ ProviderRegistry │  │
                    │  │Interface │  │  Struct  │  │     (Factory)    │  │
                    │  └──────────┘  └──────────┘  └──────────────────┘  │
                    └─────────────────────┬───────────────────────────────┘
                                          │
              ┌───────────────────────────┼───────────────────────────┐
              │                           │                           │
              ▼                           ▼                           ▼
┌─────────────────────────┐  ┌─────────────────────────┐  ┌─────────────────────────┐
│ internal/platform/openai │  │internal/platform/anthropic│  │internal/platform/gemini │
│                         │  │                         │  │                         │
│  ┌───────────────────┐  │  │  ┌───────────────────┐  │  │  ┌───────────────────┐  │
│  │  OpenAIClient     │  │  │  │ AnthropicClient   │  │  │  │ GeminiAPIClient   │  │
│  │  (implements      │  │  │  │ (implements       │  │  │  │ (implements       │  │
│  │   Provider)       │  │  │  │  Provider)        │  │  │  │  Provider)        │  │
│  └───────────────────┘  │  │  └───────────────────┘  │  │  └───────────────────┘  │
│                         │  │                         │  │  ┌───────────────────┐  │
│                         │  │                         │  │  │ GeminiWebProvider │  │
│                         │  │                         │  │  │ (implements       │  │
│                         │  │                         │  │  │  Provider)        │  │
│                         │  │                         │  │  └───────────────────┘  │
└─────────────────────────┘  └─────────────────────────┘  └─────────────────────────┘
```

---

## Fase 1: Interface Provider e Tipos Base

### 1.1 Interface Provider

**Arquivo**: `internal/core/llm/provider.go`

```go
package llm

import (
    "context"
    "time"
)

// Result representa o resultado de uma chamada ao LLM
type Result struct {
    Response    string        // Resposta processada/limpa
    RawResponse string        // Resposta bruta da API
    Model       string        // Modelo usado
    Provider    string        // Nome do provider
    Duration    time.Duration // Tempo de execucao
    Usage       *Usage        // Metricas de uso (tokens, etc)
}

// Usage contem metricas de uso da API
type Usage struct {
    PromptTokens     int // Tokens do prompt
    CompletionTokens int // Tokens da resposta
    TotalTokens      int // Total de tokens
}

// Provider define a interface comum para provedores de LLM
type Provider interface {
    // Send envia um prompt e retorna a resposta
    Send(ctx context.Context, content string) (*Result, error)
    
    // SendWithProgress envia com callback de progresso (para TUI)
    SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error)
    
    // Name retorna o nome do provider (ex: "OpenAI", "Anthropic", "Gemini")
    Name() string
    
    // IsAvailable verifica se o provider esta disponivel (ex: binario existe, etc)
    IsAvailable() bool
    
    // IsConfigured verifica se o provider esta configurado (ex: API key presente)
    IsConfigured() bool
    
    // ValidateConfig valida a configuracao antes de usar
    ValidateConfig() error
}

// ProviderType identifica o tipo de provider
type ProviderType string

const (
    ProviderOpenAI    ProviderType = "openai"
    ProviderAnthropic ProviderType = "anthropic"
    ProviderGemini    ProviderType = "gemini"
    ProviderGeminiWeb ProviderType = "geminiweb" // Legado
)

// AllProviders retorna todos os providers suportados
func AllProviders() []ProviderType {
    return []ProviderType{
        ProviderOpenAI,
        ProviderAnthropic,
        ProviderGemini,
        ProviderGeminiWeb,
    }
}

// IsValidProvider verifica se o provider e valido
func IsValidProvider(p string) bool {
    for _, valid := range AllProviders() {
        if string(valid) == p {
            return true
        }
    }
    return false
}
```

### 1.2 Configuracao Unificada

**Arquivo**: `internal/core/llm/config.go`

```go
package llm

import (
    "fmt"
    "net/url"
    "strings"
)

// Config contem configuracao unificada para qualquer provider LLM
type Config struct {
    // Provider especifica qual provider usar
    Provider ProviderType
    
    // Configuracoes comuns a todos os providers
    APIKey  string // Chave da API
    BaseURL string // URL base da API (permite custom endpoints)
    Model   string // Modelo a usar
    Timeout int    // Timeout em segundos
    
    // Configuracoes especificas do GeminiWeb (legado)
    BinaryPath     string
    BrowserRefresh string
    
    // Configuracoes opcionais
    MaxTokens   int     // Limite de tokens na resposta
    Temperature float64 // Temperatura (0.0 - 2.0)
}

// DefaultConfigs retorna configuracoes padrao por provider
func DefaultConfigs() map[ProviderType]Config {
    return map[ProviderType]Config{
        ProviderOpenAI: {
            Provider: ProviderOpenAI,
            BaseURL:  "https://api.openai.com/v1",
            Model:    "gpt-4o",
            Timeout:  300,
        },
        ProviderAnthropic: {
            Provider: ProviderAnthropic,
            BaseURL:  "https://api.anthropic.com",
            Model:    "claude-3-5-sonnet-latest",
            Timeout:  300,
        },
        ProviderGemini: {
            Provider: ProviderGemini,
            BaseURL:  "https://generativelanguage.googleapis.com/v1beta",
            Model:    "gemini-2.5-flash",
            Timeout:  300,
        },
        ProviderGeminiWeb: {
            Provider: ProviderGeminiWeb,
            Model:    "gemini-2.5-flash",
            Timeout:  300,
        },
    }
}

// Validate valida a configuracao
func (c *Config) Validate() error {
    if c.Provider == "" {
        return fmt.Errorf("provider is required")
    }
    
    if !IsValidProvider(string(c.Provider)) {
        return fmt.Errorf("invalid provider: %s", c.Provider)
    }
    
    // GeminiWeb nao requer API key
    if c.Provider != ProviderGeminiWeb {
        if c.APIKey == "" {
            return fmt.Errorf("api-key is required for provider %s", c.Provider)
        }
    }
    
    if c.Model == "" {
        return fmt.Errorf("model is required")
    }
    
    if c.BaseURL != "" && c.Provider != ProviderGeminiWeb {
        if _, err := url.Parse(c.BaseURL); err != nil {
            return fmt.Errorf("invalid base-url: %w", err)
        }
    }
    
    if c.Timeout <= 0 {
        return fmt.Errorf("timeout must be positive")
    }
    
    return nil
}

// MaskAPIKey retorna a API key mascarada para exibicao
func (c *Config) MaskAPIKey() string {
    if c.APIKey == "" {
        return "(not configured)"
    }
    if len(c.APIKey) <= 8 {
        return "***"
    }
    return c.APIKey[:4] + "..." + c.APIKey[len(c.APIKey)-4:]
}
```

### 1.3 Registry de Providers (Factory)

**Arquivo**: `internal/core/llm/registry.go`

```go
package llm

import (
    "fmt"
    "sync"
)

// ProviderCreator e uma funcao que cria um Provider a partir de Config
type ProviderCreator func(cfg Config) (Provider, error)

// Registry gerencia o registro de providers
type Registry struct {
    mu       sync.RWMutex
    creators map[ProviderType]ProviderCreator
}

// NewRegistry cria um novo registry vazio
func NewRegistry() *Registry {
    return &Registry{
        creators: make(map[ProviderType]ProviderCreator),
    }
}

// Register registra um creator para um tipo de provider
func (r *Registry) Register(providerType ProviderType, creator ProviderCreator) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.creators[providerType] = creator
}

// Create cria um provider a partir da configuracao
func (r *Registry) Create(cfg Config) (Provider, error) {
    r.mu.RLock()
    creator, ok := r.creators[cfg.Provider]
    r.mu.RUnlock()
    
    if !ok {
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
    }
    
    return creator(cfg)
}

// SupportedProviders retorna os providers registrados
func (r *Registry) SupportedProviders() []ProviderType {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    result := make([]ProviderType, 0, len(r.creators))
    for pt := range r.creators {
        result = append(result, pt)
    }
    return result
}
```

---

## Fase 2: Implementar Provider OpenAI

### 2.1 Cliente OpenAI

**Arquivo**: `internal/platform/openai/client.go`

```go
package openai

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// Client implementa llm.Provider para APIs compativeis com OpenAI
type Client struct {
    apiKey     string
    baseURL    string
    model      string
    timeout    time.Duration
    httpClient *http.Client
}

// NewClient cria um novo cliente OpenAI
func NewClient(cfg llm.Config) (*Client, error) {
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("api key is required")
    }
    
    baseURL := cfg.BaseURL
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }
    
    timeout := time.Duration(cfg.Timeout) * time.Second
    if timeout == 0 {
        timeout = 300 * time.Second
    }
    
    return &Client{
        apiKey:  cfg.APIKey,
        baseURL: baseURL,
        model:   cfg.Model,
        timeout: timeout,
        httpClient: &http.Client{
            Timeout: timeout,
        },
    }, nil
}

func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
    startTime := time.Now()
    
    req := ChatCompletionRequest{
        Model: c.model,
        Messages: []Message{
            {Role: "user", Content: content},
        },
    }
    
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    url := c.baseURL + "/chat/completions"
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    if resp.StatusCode != http.StatusOK {
        var errResp ErrorResponse
        if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
            return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, errResp.Error.Message)
        }
        return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(respBody))
    }
    
    var chatResp ChatCompletionResponse
    if err := json.Unmarshal(respBody, &chatResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    if len(chatResp.Choices) == 0 {
        return nil, fmt.Errorf("no choices in response")
    }
    
    var usage *llm.Usage
    if chatResp.Usage.TotalTokens > 0 {
        usage = &llm.Usage{
            PromptTokens:     chatResp.Usage.PromptTokens,
            CompletionTokens: chatResp.Usage.CompletionTokens,
            TotalTokens:      chatResp.Usage.TotalTokens,
        }
    }
    
    return &llm.Result{
        Response:    chatResp.Choices[0].Message.Content,
        RawResponse: string(respBody),
        Model:       c.model,
        Provider:    "OpenAI",
        Duration:    time.Since(startTime),
        Usage:       usage,
    }, nil
}

func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
    progress("Connecting to OpenAI...")
    result, err := c.Send(ctx, content)
    if err == nil {
        progress("Response received")
    }
    return result, err
}

func (c *Client) Name() string {
    return "OpenAI"
}

func (c *Client) IsAvailable() bool {
    return true // API sempre disponivel se tiver internet
}

func (c *Client) IsConfigured() bool {
    return c.apiKey != "" && c.model != ""
}

func (c *Client) ValidateConfig() error {
    if c.apiKey == "" {
        return fmt.Errorf("API key is required")
    }
    if c.model == "" {
        return fmt.Errorf("model is required")
    }
    return nil
}
```

### 2.2 Tipos OpenAI

**Arquivo**: `internal/platform/openai/types.go`

```go
package openai

// ChatCompletionRequest representa o corpo da requisicao
type ChatCompletionRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// ChatCompletionResponse representa a resposta da API
type ChatCompletionResponse struct {
    ID      string   `json:"id"`
    Object  string   `json:"object"`
    Created int64    `json:"created"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   UsageAPI `json:"usage"`
}

type Choice struct {
    Index        int     `json:"index"`
    Message      Message `json:"message"`
    FinishReason string  `json:"finish_reason"`
}

type UsageAPI struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

// ErrorResponse representa um erro da API
type ErrorResponse struct {
    Error struct {
        Message string `json:"message"`
        Type    string `json:"type"`
        Code    string `json:"code"`
    } `json:"error"`
}
```

### 2.3 Modelos Validos OpenAI

**Arquivo**: `internal/platform/openai/models.go`

```go
package openai

// ValidModels retorna modelos conhecidos para OpenAI
// Nota: Nao restringimos a esses pois custom endpoints podem ter outros
func ValidModels() []string {
    return []string{
        "gpt-4o",
        "gpt-4o-mini",
        "gpt-4-turbo",
        "gpt-4",
        "gpt-3.5-turbo",
        "o1-preview",
        "o1-mini",
    }
}

// IsKnownModel verifica se e um modelo conhecido
// Retorna true para qualquer modelo se base-url for customizado
func IsKnownModel(model, baseURL string) bool {
    // Se usando endpoint customizado, aceitar qualquer modelo
    if baseURL != "" && baseURL != "https://api.openai.com/v1" {
        return true
    }
    
    for _, known := range ValidModels() {
        if model == known {
            return true
        }
    }
    return false
}
```

---

## Fase 3: Implementar Provider Anthropic

### 3.1 Cliente Anthropic

**Arquivo**: `internal/platform/anthropic/client.go`

```go
package anthropic

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

const (
    defaultBaseURL   = "https://api.anthropic.com"
    anthropicVersion = "2023-06-01"
)

// Client implementa llm.Provider para API Anthropic
type Client struct {
    apiKey     string
    baseURL    string
    model      string
    timeout    time.Duration
    httpClient *http.Client
}

// NewClient cria um novo cliente Anthropic
func NewClient(cfg llm.Config) (*Client, error) {
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("api key is required")
    }
    
    baseURL := cfg.BaseURL
    if baseURL == "" {
        baseURL = defaultBaseURL
    }
    
    timeout := time.Duration(cfg.Timeout) * time.Second
    if timeout == 0 {
        timeout = 300 * time.Second
    }
    
    return &Client{
        apiKey:  cfg.APIKey,
        baseURL: baseURL,
        model:   cfg.Model,
        timeout: timeout,
        httpClient: &http.Client{
            Timeout: timeout,
        },
    }, nil
}

func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
    startTime := time.Now()
    
    req := MessagesRequest{
        Model:     c.model,
        MaxTokens: 8192,
        Messages: []Message{
            {Role: "user", Content: content},
        },
    }
    
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    url := c.baseURL + "/v1/messages"
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("x-api-key", c.apiKey)
    httpReq.Header.Set("anthropic-version", anthropicVersion)
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    if resp.StatusCode != http.StatusOK {
        var errResp ErrorResponse
        if json.Unmarshal(respBody, &errResp) == nil && errResp.Error.Message != "" {
            return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, errResp.Error.Message)
        }
        return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(respBody))
    }
    
    var msgResp MessagesResponse
    if err := json.Unmarshal(respBody, &msgResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    // Extrair texto dos content blocks
    var responseText string
    for _, block := range msgResp.Content {
        if block.Type == "text" {
            responseText += block.Text
        }
    }
    
    var usage *llm.Usage
    if msgResp.Usage.InputTokens > 0 || msgResp.Usage.OutputTokens > 0 {
        usage = &llm.Usage{
            PromptTokens:     msgResp.Usage.InputTokens,
            CompletionTokens: msgResp.Usage.OutputTokens,
            TotalTokens:      msgResp.Usage.InputTokens + msgResp.Usage.OutputTokens,
        }
    }
    
    return &llm.Result{
        Response:    responseText,
        RawResponse: string(respBody),
        Model:       c.model,
        Provider:    "Anthropic",
        Duration:    time.Since(startTime),
        Usage:       usage,
    }, nil
}

func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
    progress("Connecting to Anthropic...")
    result, err := c.Send(ctx, content)
    if err == nil {
        progress("Response received")
    }
    return result, err
}

func (c *Client) Name() string {
    return "Anthropic"
}

func (c *Client) IsAvailable() bool {
    return true
}

func (c *Client) IsConfigured() bool {
    return c.apiKey != "" && c.model != ""
}

func (c *Client) ValidateConfig() error {
    if c.apiKey == "" {
        return fmt.Errorf("API key is required")
    }
    if c.model == "" {
        return fmt.Errorf("model is required")
    }
    return nil
}
```

### 3.2 Tipos Anthropic

**Arquivo**: `internal/platform/anthropic/types.go`

```go
package anthropic

// MessagesRequest representa o corpo da requisicao para /v1/messages
type MessagesRequest struct {
    Model     string    `json:"model"`
    MaxTokens int       `json:"max_tokens"`
    Messages  []Message `json:"messages"`
    System    string    `json:"system,omitempty"`
    Stream    bool      `json:"stream,omitempty"`
}

type Message struct {
    Role    string `json:"role"` // "user" ou "assistant"
    Content string `json:"content"`
}

// MessagesResponse representa a resposta da API
type MessagesResponse struct {
    ID           string         `json:"id"`
    Type         string         `json:"type"`
    Role         string         `json:"role"`
    Content      []ContentBlock `json:"content"`
    Model        string         `json:"model"`
    StopReason   string         `json:"stop_reason"`
    StopSequence string         `json:"stop_sequence,omitempty"`
    Usage        UsageAPI       `json:"usage"`
}

type ContentBlock struct {
    Type string `json:"type"` // "text"
    Text string `json:"text"`
}

type UsageAPI struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
}

// ErrorResponse representa um erro da API
type ErrorResponse struct {
    Type  string `json:"type"`
    Error struct {
        Type    string `json:"type"`
        Message string `json:"message"`
    } `json:"error"`
}
```

### 3.3 Modelos Validos Anthropic

**Arquivo**: `internal/platform/anthropic/models.go`

```go
package anthropic

// ValidModels retorna modelos conhecidos para Anthropic
func ValidModels() []string {
    return []string{
        "claude-3-5-sonnet-latest",
        "claude-3-5-sonnet-20241022",
        "claude-3-5-haiku-latest",
        "claude-3-opus-latest",
        "claude-3-opus-20240229",
        "claude-3-sonnet-20240229",
        "claude-3-haiku-20240307",
    }
}

// IsKnownModel verifica se e um modelo conhecido
func IsKnownModel(model, baseURL string) bool {
    // Se usando endpoint customizado, aceitar qualquer modelo
    if baseURL != "" && baseURL != "https://api.anthropic.com" {
        return true
    }
    
    for _, known := range ValidModels() {
        if model == known {
            return true
        }
    }
    return false
}
```

---

## Fase 4: Implementar Provider Gemini API

### 4.1 Cliente Gemini

**Arquivo**: `internal/platform/geminiapi/client.go`

```go
package geminiapi

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Client implementa llm.Provider para Google Gemini API
type Client struct {
    apiKey     string
    baseURL    string
    model      string
    timeout    time.Duration
    httpClient *http.Client
}

// NewClient cria um novo cliente Gemini API
func NewClient(cfg llm.Config) (*Client, error) {
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("api key is required")
    }
    
    baseURL := cfg.BaseURL
    if baseURL == "" {
        baseURL = defaultBaseURL
    }
    
    timeout := time.Duration(cfg.Timeout) * time.Second
    if timeout == 0 {
        timeout = 300 * time.Second
    }
    
    return &Client{
        apiKey:  cfg.APIKey,
        baseURL: baseURL,
        model:   cfg.Model,
        timeout: timeout,
        httpClient: &http.Client{
            Timeout: timeout,
        },
    }, nil
}

func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error) {
    startTime := time.Now()
    
    req := GenerateRequest{
        Contents: []Content{
            {
                Parts: []Part{{Text: content}},
            },
        },
        GenerationConfig: &GenerationConfig{
            MaxOutputTokens: 8192,
        },
    }
    
    body, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    url := fmt.Sprintf("%s/models/%s:generateContent?key=%s",
        c.baseURL, c.model, c.apiKey)
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()
    
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    var genResp GenerateResponse
    if err := json.Unmarshal(respBody, &genResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    if genResp.Error != nil {
        return nil, fmt.Errorf("API error [%d]: %s", genResp.Error.Code, genResp.Error.Message)
    }
    
    if len(genResp.Candidates) == 0 {
        return nil, fmt.Errorf("no candidates in response")
    }
    
    var responseText string
    for _, part := range genResp.Candidates[0].Content.Parts {
        responseText += part.Text
    }
    
    var usage *llm.Usage
    if genResp.UsageMetadata != nil {
        usage = &llm.Usage{
            PromptTokens:     genResp.UsageMetadata.PromptTokenCount,
            CompletionTokens: genResp.UsageMetadata.CandidatesTokenCount,
            TotalTokens:      genResp.UsageMetadata.TotalTokenCount,
        }
    }
    
    return &llm.Result{
        Response:    responseText,
        RawResponse: string(respBody),
        Model:       c.model,
        Provider:    "Gemini",
        Duration:    time.Since(startTime),
        Usage:       usage,
    }, nil
}

func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
    progress("Connecting to Gemini API...")
    result, err := c.Send(ctx, content)
    if err == nil {
        progress("Response received")
    }
    return result, err
}

func (c *Client) Name() string {
    return "Gemini"
}

func (c *Client) IsAvailable() bool {
    return true
}

func (c *Client) IsConfigured() bool {
    return c.apiKey != "" && c.model != ""
}

func (c *Client) ValidateConfig() error {
    if c.apiKey == "" {
        return fmt.Errorf("API key is required")
    }
    if c.model == "" {
        return fmt.Errorf("model is required")
    }
    return nil
}
```

### 4.2 Tipos Gemini

**Arquivo**: `internal/platform/geminiapi/types.go`

```go
package geminiapi

// GenerateRequest representa o corpo da requisicao
type GenerateRequest struct {
    Contents         []Content         `json:"contents"`
    GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
    SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}

type Content struct {
    Parts []Part `json:"parts"`
    Role  string `json:"role,omitempty"`
}

type Part struct {
    Text string `json:"text"`
}

type GenerationConfig struct {
    Temperature     float64  `json:"temperature,omitempty"`
    TopK            int      `json:"topK,omitempty"`
    TopP            float64  `json:"topP,omitempty"`
    MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
    StopSequences   []string `json:"stopSequences,omitempty"`
}

type SafetySetting struct {
    Category  string `json:"category"`
    Threshold string `json:"threshold"`
}

// GenerateResponse representa a resposta da API
type GenerateResponse struct {
    Candidates    []Candidate    `json:"candidates"`
    UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
    Error         *APIError      `json:"error,omitempty"`
}

type Candidate struct {
    Content      Content `json:"content"`
    FinishReason string  `json:"finishReason"`
    Index        int     `json:"index"`
}

type UsageMetadata struct {
    PromptTokenCount     int `json:"promptTokenCount"`
    CandidatesTokenCount int `json:"candidatesTokenCount"`
    TotalTokenCount      int `json:"totalTokenCount"`
}

type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Status  string `json:"status"`
}
```

### 4.3 Modelos Validos Gemini

**Arquivo**: `internal/platform/geminiapi/models.go`

```go
package geminiapi

// ValidModels retorna modelos conhecidos para Gemini API
func ValidModels() []string {
    return []string{
        "gemini-2.5-flash",
        "gemini-2.5-pro",
        "gemini-2.0-flash",
        "gemini-1.5-flash",
        "gemini-1.5-pro",
    }
}

// IsKnownModel verifica se e um modelo conhecido
func IsKnownModel(model, baseURL string) bool {
    // Se usando endpoint customizado, aceitar qualquer modelo
    if baseURL != "" && baseURL != "https://generativelanguage.googleapis.com/v1beta" {
        return true
    }
    
    for _, known := range ValidModels() {
        if model == known {
            return true
        }
    }
    return false
}
```

---

## Fase 5: Adaptar GeminiWeb como Provider (Legado)

### 5.1 Provider GeminiWeb

**Arquivo**: `internal/platform/gemini/provider.go` (novo arquivo)

```go
package gemini

import (
    "context"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// WebProvider adapta o Executor existente para interface llm.Provider
type WebProvider struct {
    executor *Executor
    config   Config
}

// NewWebProvider cria um provider GeminiWeb
func NewWebProvider(cfg llm.Config) (*WebProvider, error) {
    execCfg := Config{
        BinaryPath:     cfg.BinaryPath,
        Model:          cfg.Model,
        Timeout:        cfg.Timeout,
        BrowserRefresh: cfg.BrowserRefresh,
    }
    
    return &WebProvider{
        executor: NewExecutor(execCfg),
        config:   execCfg,
    }, nil
}

func (p *WebProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
    result, err := p.executor.Send(ctx, content)
    if err != nil {
        return nil, err
    }
    
    return &llm.Result{
        Response:    result.Response,
        RawResponse: result.RawResponse,
        Model:       result.Model,
        Provider:    "GeminiWeb",
        Duration:    result.Duration,
    }, nil
}

func (p *WebProvider) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error) {
    result, err := p.executor.SendWithProgress(ctx, content, progress)
    if err != nil {
        return nil, err
    }
    
    return &llm.Result{
        Response:    result.Response,
        RawResponse: result.RawResponse,
        Model:       result.Model,
        Provider:    "GeminiWeb",
        Duration:    result.Duration,
    }, nil
}

func (p *WebProvider) Name() string {
    return "GeminiWeb"
}

func (p *WebProvider) IsAvailable() bool {
    return IsAvailable()
}

func (p *WebProvider) IsConfigured() bool {
    return IsConfigured()
}

func (p *WebProvider) ValidateConfig() error {
    if !p.IsAvailable() {
        return fmt.Errorf("geminiweb binary not found")
    }
    if !p.IsConfigured() {
        return fmt.Errorf("geminiweb not configured (run: geminiweb auto-login)")
    }
    return nil
}
```

---

## Fase 6: Composition Root - Registro de Providers

### 6.1 Registro no cmd

**Arquivo**: `cmd/providers.go` (novo arquivo)

```go
package cmd

import (
    "fmt"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
    "github.com/quantmind-br/shotgun-cli/internal/platform/anthropic"
    "github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
    "github.com/quantmind-br/shotgun-cli/internal/platform/geminiapi"
    "github.com/quantmind-br/shotgun-cli/internal/platform/openai"
)

// providerRegistry e o registro global de providers
var providerRegistry *llm.Registry

func init() {
    providerRegistry = llm.NewRegistry()
    
    // Registrar OpenAI
    providerRegistry.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) {
        return openai.NewClient(cfg)
    })
    
    // Registrar Anthropic
    providerRegistry.Register(llm.ProviderAnthropic, func(cfg llm.Config) (llm.Provider, error) {
        return anthropic.NewClient(cfg)
    })
    
    // Registrar Gemini API
    providerRegistry.Register(llm.ProviderGemini, func(cfg llm.Config) (llm.Provider, error) {
        return geminiapi.NewClient(cfg)
    })
    
    // Registrar GeminiWeb (legado)
    providerRegistry.Register(llm.ProviderGeminiWeb, func(cfg llm.Config) (llm.Provider, error) {
        return gemini.NewWebProvider(cfg)
    })
}

// CreateLLMProvider cria um provider baseado na configuracao
func CreateLLMProvider(cfg llm.Config) (llm.Provider, error) {
    provider, err := providerRegistry.Create(cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create provider: %w", err)
    }
    
    if err := provider.ValidateConfig(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return provider, nil
}

// GetProviderRegistry retorna o registry para uso externo
func GetProviderRegistry() *llm.Registry {
    return providerRegistry
}
```

---

## Fase 7: Atualizar Sistema de Configuracao

### 7.1 Novos Defaults

**Arquivo**: `cmd/root.go` - Atualizar `setConfigDefaults()`

```go
func setConfigDefaults() {
    // ... existing defaults ...
    
    // LLM Provider defaults (NOVO)
    viper.SetDefault("llm.provider", "geminiweb")     // Provider padrao (retrocompativel)
    viper.SetDefault("llm.api-key", "")               // API key do provider atual
    viper.SetDefault("llm.base-url", "")              // URL customizada (opcional)
    viper.SetDefault("llm.model", "")                 // Modelo (usa default do provider se vazio)
    viper.SetDefault("llm.timeout", 300)              // Timeout em segundos
    
    // Gemini integration defaults (manter para retrocompatibilidade)
    viper.SetDefault("gemini.enabled", false)         // Deprecado - usar llm.provider
    viper.SetDefault("gemini.binary-path", "")
    viper.SetDefault("gemini.model", "gemini-2.5-flash")
    viper.SetDefault("gemini.timeout", 300)
    viper.SetDefault("gemini.browser-refresh", "auto")
    viper.SetDefault("gemini.auto-send", false)
    viper.SetDefault("gemini.save-response", true)
}
```

### 7.2 Novas Validacoes

**Arquivo**: `cmd/config.go` - Adicionar novas chaves e validacoes

```go
func isValidConfigKey(key string) bool {
    validKeys := []string{
        // ... existing keys ...
        
        // LLM Provider keys (NOVO)
        "llm.provider",
        "llm.api-key",
        "llm.base-url",
        "llm.model",
        "llm.timeout",
        
        // Gemini integration keys (manter para compatibilidade)
        "gemini.enabled",
        "gemini.binary-path",
        "gemini.model",
        "gemini.timeout",
        "gemini.browser-refresh",
        "gemini.auto-send",
        "gemini.save-response",
    }
    // ...
}

func validateConfigValue(key, value string) error {
    switch key {
    // ... existing cases ...
    
    case "llm.provider":
        return validateLLMProvider(value)
    case "llm.api-key":
        // API key pode ser vazia (usa geminiweb) ou deve ter conteudo
        return nil
    case "llm.base-url":
        if value != "" {
            return validateURL(value)
        }
        return nil
    case "llm.model":
        // Modelo pode ser qualquer string - validacao especifica por provider
        return nil
    case "llm.timeout":
        return validateTimeout(value)
    }
    // ...
}

func validateLLMProvider(value string) error {
    if !llm.IsValidProvider(value) {
        return fmt.Errorf("expected one of: openai, anthropic, gemini, geminiweb")
    }
    return nil
}

func validateURL(value string) error {
    _, err := url.Parse(value)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }
    return nil
}

func validateTimeout(value string) error {
    var timeout int
    if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil {
        return fmt.Errorf("expected a positive integer (seconds)")
    }
    if timeout <= 0 || timeout > 3600 {
        return fmt.Errorf("timeout must be between 1 and 3600 seconds")
    }
    return nil
}
```

### 7.3 Helper para Construir Config

**Arquivo**: `cmd/config_llm.go` (novo arquivo)

```go
package cmd

import (
    "github.com/spf13/viper"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

// BuildLLMConfig constroi a configuracao LLM a partir do Viper
func BuildLLMConfig() llm.Config {
    provider := llm.ProviderType(viper.GetString("llm.provider"))
    
    // Obter defaults do provider
    defaults := llm.DefaultConfigs()[provider]
    
    cfg := llm.Config{
        Provider: provider,
        APIKey:   viper.GetString("llm.api-key"),
        BaseURL:  viper.GetString("llm.base-url"),
        Model:    viper.GetString("llm.model"),
        Timeout:  viper.GetInt("llm.timeout"),
    }
    
    // Aplicar defaults se nao configurado
    if cfg.BaseURL == "" {
        cfg.BaseURL = defaults.BaseURL
    }
    if cfg.Model == "" {
        cfg.Model = defaults.Model
    }
    if cfg.Timeout == 0 {
        cfg.Timeout = defaults.Timeout
    }
    
    // Configuracoes especificas do GeminiWeb
    if provider == llm.ProviderGeminiWeb {
        cfg.BinaryPath = viper.GetString("gemini.binary-path")
        cfg.BrowserRefresh = viper.GetString("gemini.browser-refresh")
        // Usar model do gemini se llm.model nao estiver definido
        if viper.GetString("llm.model") == "" {
            cfg.Model = viper.GetString("gemini.model")
        }
    }
    
    return cfg
}

// BuildLLMConfigWithOverrides constroi config com overrides de flags
func BuildLLMConfigWithOverrides(model string, timeout int) llm.Config {
    cfg := BuildLLMConfig()
    
    if model != "" {
        cfg.Model = model
    }
    if timeout > 0 {
        cfg.Timeout = timeout
    }
    
    return cfg
}
```

---

## Fase 8: Atualizar Comandos CLI

### 8.1 Atualizar `cmd/send.go`

```go
func runContextSend(cmd *cobra.Command, args []string) error {
    // ... existing content reading logic ...
    
    // Check if any LLM provider is enabled/configured
    provider := viper.GetString("llm.provider")
    
    // Retrocompatibilidade: se llm.provider nao definido, verificar gemini.enabled
    if provider == "" || provider == "geminiweb" {
        if !viper.GetBool("gemini.enabled") {
            return fmt.Errorf("LLM integration is disabled. Enable with: shotgun-cli config set gemini.enabled true")
        }
    }
    
    // Get flag overrides
    model, _ := cmd.Flags().GetString("model")
    timeout, _ := cmd.Flags().GetInt("timeout")
    
    // Build config
    cfg := BuildLLMConfigWithOverrides(model, timeout)
    
    // Create provider
    llmProvider, err := CreateLLMProvider(cfg)
    if err != nil {
        return fmt.Errorf("failed to create provider: %w", err)
    }
    
    if !llmProvider.IsAvailable() {
        return fmt.Errorf("%s not available. Run 'shotgun-cli llm doctor' for help", llmProvider.Name())
    }
    
    if !llmProvider.IsConfigured() {
        return fmt.Errorf("%s not configured. Run 'shotgun-cli llm doctor' for help", llmProvider.Name())
    }
    
    // Send
    log.Info().
        Str("provider", llmProvider.Name()).
        Str("model", cfg.Model).
        Msg("Sending to LLM")
    
    fmt.Printf("Sending to %s (%s)...\n", llmProvider.Name(), cfg.Model)
    
    ctx := context.Background()
    result, err := llmProvider.Send(ctx, content)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    
    // Output handling
    response := result.Response
    if raw, _ := cmd.Flags().GetBool("raw"); raw {
        response = result.RawResponse
    }
    
    outputFile, _ := cmd.Flags().GetString("output")
    if outputFile != "" {
        if err := os.WriteFile(outputFile, []byte(response), 0600); err != nil {
            return fmt.Errorf("failed to save response: %w", err)
        }
        fmt.Printf("Response saved to: %s\n", outputFile)
    } else {
        fmt.Println(response)
    }
    
    // Show usage if available
    if result.Usage != nil {
        fmt.Printf("Tokens: %d (prompt: %d, completion: %d)\n",
            result.Usage.TotalTokens,
            result.Usage.PromptTokens,
            result.Usage.CompletionTokens)
    }
    fmt.Printf("Duration: %s\n", result.Duration.Round(time.Millisecond))
    
    return nil
}
```

### 8.2 Novo Comando `cmd/llm.go`

```go
package cmd

import (
    "fmt"
    "os"
    "text/tabwriter"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

var llmCmd = &cobra.Command{
    Use:   "llm",
    Short: "LLM provider management",
    Long:  "Commands for managing and diagnosing LLM provider configuration",
}

var llmStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show LLM provider status",
    RunE:  runLLMStatus,
}

var llmDoctorCmd = &cobra.Command{
    Use:   "doctor",
    Short: "Diagnose LLM configuration",
    RunE:  runLLMDoctor,
}

var llmListCmd = &cobra.Command{
    Use:   "list",
    Short: "List supported providers",
    RunE:  runLLMList,
}

func runLLMStatus(cmd *cobra.Command, args []string) error {
    cfg := BuildLLMConfig()
    
    w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
    
    fmt.Fprintln(w, "=== LLM Configuration ===")
    fmt.Fprintln(w)
    fmt.Fprintf(w, "Provider:\t%s\n", cfg.Provider)
    fmt.Fprintf(w, "Model:\t%s\n", cfg.Model)
    fmt.Fprintf(w, "Base URL:\t%s\n", displayURL(cfg.BaseURL, cfg.Provider))
    fmt.Fprintf(w, "API Key:\t%s\n", cfg.MaskAPIKey())
    fmt.Fprintf(w, "Timeout:\t%ds\n", cfg.Timeout)
    
    w.Flush()
    
    // Test provider
    fmt.Println()
    provider, err := CreateLLMProvider(cfg)
    if err != nil {
        fmt.Printf("Status: Not ready - %s\n", err)
        return nil
    }
    
    if !provider.IsConfigured() {
        fmt.Printf("Status: Not configured\n")
        return nil
    }
    
    fmt.Printf("Status: Ready\n")
    return nil
}

func runLLMDoctor(cmd *cobra.Command, args []string) error {
    cfg := BuildLLMConfig()
    
    fmt.Printf("Running diagnostics for %s...\n\n", cfg.Provider)
    
    issues := []string{}
    
    // Check 1: Provider type
    fmt.Print("Checking provider... ")
    if llm.IsValidProvider(string(cfg.Provider)) {
        fmt.Printf("%s\n", cfg.Provider)
    } else {
        fmt.Printf("invalid: %s\n", cfg.Provider)
        issues = append(issues, fmt.Sprintf("Invalid provider: %s", cfg.Provider))
    }
    
    // Check 2: API Key (exceto GeminiWeb)
    if cfg.Provider != llm.ProviderGeminiWeb {
        fmt.Print("Checking API key... ")
        if cfg.APIKey != "" {
            fmt.Println("configured")
        } else {
            fmt.Println("not configured")
            issues = append(issues, "API key not configured")
        }
    }
    
    // Check 3: Model
    fmt.Print("Checking model... ")
    if cfg.Model != "" {
        fmt.Printf("%s\n", cfg.Model)
    } else {
        fmt.Println("not configured")
        issues = append(issues, "Model not configured")
    }
    
    // Check 4: Provider-specific
    provider, err := CreateLLMProvider(cfg)
    if err == nil {
        fmt.Print("Checking provider availability... ")
        if provider.IsAvailable() {
            fmt.Println("OK")
        } else {
            fmt.Println("not available")
            issues = append(issues, fmt.Sprintf("%s is not available", provider.Name()))
        }
    }
    
    // Summary
    fmt.Println()
    if len(issues) == 0 {
        fmt.Printf("No issues found! %s is ready.\n", cfg.Provider)
        return nil
    }
    
    fmt.Printf("Found %d issue(s):\n", len(issues))
    for i, issue := range issues {
        fmt.Printf("  %d. %s\n", i+1, issue)
    }
    
    // Provider-specific help
    fmt.Println("\nNext steps:")
    switch cfg.Provider {
    case llm.ProviderOpenAI:
        fmt.Println("  1. Get API key from: https://platform.openai.com/api-keys")
        fmt.Println("  2. Configure: shotgun-cli config set llm.api-key YOUR_KEY")
    case llm.ProviderAnthropic:
        fmt.Println("  1. Get API key from: https://console.anthropic.com/settings/keys")
        fmt.Println("  2. Configure: shotgun-cli config set llm.api-key YOUR_KEY")
    case llm.ProviderGemini:
        fmt.Println("  1. Get API key from: https://aistudio.google.com/app/apikey")
        fmt.Println("  2. Configure: shotgun-cli config set llm.api-key YOUR_KEY")
    case llm.ProviderGeminiWeb:
        fmt.Println("  1. Install: go install github.com/diogo/geminiweb/cmd/geminiweb@latest")
        fmt.Println("  2. Configure: geminiweb auto-login")
    }
    
    return nil
}

func runLLMList(cmd *cobra.Command, args []string) error {
    fmt.Println("Supported LLM Providers:")
    fmt.Println()
    
    providers := []struct {
        id      llm.ProviderType
        name    string
        desc    string
        apiKeyURL string
    }{
        {llm.ProviderOpenAI, "OpenAI", "GPT-4, GPT-4o, o1", "https://platform.openai.com/api-keys"},
        {llm.ProviderAnthropic, "Anthropic", "Claude 3.5, Claude 3", "https://console.anthropic.com/settings/keys"},
        {llm.ProviderGemini, "Google Gemini", "Gemini 2.5, Gemini 1.5", "https://aistudio.google.com/app/apikey"},
        {llm.ProviderGeminiWeb, "GeminiWeb", "Browser-based (no API key)", "N/A"},
    }
    
    current := viper.GetString("llm.provider")
    
    for _, p := range providers {
        marker := "  "
        if string(p.id) == current {
            marker = "* "
        }
        fmt.Printf("%s%-12s - %s (%s)\n", marker, p.id, p.name, p.desc)
    }
    
    fmt.Println()
    fmt.Println("Configure with:")
    fmt.Println("  shotgun-cli config set llm.provider <provider>")
    fmt.Println("  shotgun-cli config set llm.api-key <your-api-key>")
    
    return nil
}

func displayURL(url string, provider llm.ProviderType) string {
    if url == "" {
        defaults := llm.DefaultConfigs()
        if d, ok := defaults[provider]; ok && d.BaseURL != "" {
            return fmt.Sprintf("(default: %s)", d.BaseURL)
        }
        return "(default)"
    }
    return url
}

func init() {
    llmCmd.AddCommand(llmStatusCmd)
    llmCmd.AddCommand(llmDoctorCmd)
    llmCmd.AddCommand(llmListCmd)
    rootCmd.AddCommand(llmCmd)
}
```

---

## Fase 9: Atualizar TUI Wizard

### 9.1 Atualizar `internal/ui/wizard.go`

```go
// Atualizar GeminiConfig para LLMConfig
type LLMConfig struct {
    Provider       string // "openai", "anthropic", "gemini", "geminiweb"
    APIKey         string
    BaseURL        string
    Model          string
    Timeout        int
    SaveResponse   bool
    // Legado GeminiWeb
    BinaryPath     string
    BrowserRefresh string
}

// WizardConfig atualizado
type WizardConfig struct {
    LLM     LLMConfig  // NOVO - substitui Gemini
    Gemini  GeminiConfig // Manter para retrocompatibilidade
    Context ContextConfig
}

// Atualizar sendToLLMCmd
func (m *WizardModel) sendToLLMCmd() tea.Cmd {
    return func() tea.Msg {
        cfg := llm.Config{
            Provider:       llm.ProviderType(m.wizardConfig.LLM.Provider),
            APIKey:         m.wizardConfig.LLM.APIKey,
            BaseURL:        m.wizardConfig.LLM.BaseURL,
            Model:          m.wizardConfig.LLM.Model,
            Timeout:        m.wizardConfig.LLM.Timeout,
            BinaryPath:     m.wizardConfig.LLM.BinaryPath,
            BrowserRefresh: m.wizardConfig.LLM.BrowserRefresh,
        }
        
        provider, err := cmd.CreateLLMProvider(cfg)
        if err != nil {
            return LLMErrorMsg{Err: err}
        }
        
        ctx := gocontext.Background()
        result, err := provider.SendWithProgress(ctx, m.generatedContent, func(stage string) {
            // Progress callback para TUI
        })
        if err != nil {
            return LLMErrorMsg{Err: err}
        }
        
        // Save response
        outputFile := strings.TrimSuffix(m.generatedFilePath, ".md") + "_response.md"
        if m.wizardConfig.LLM.SaveResponse {
            if err := os.WriteFile(outputFile, []byte(result.Response), 0600); err != nil {
                return LLMErrorMsg{Err: fmt.Errorf("failed to save response: %w", err)}
            }
        }
        
        return LLMCompleteMsg{
            Response:   result.Response,
            OutputFile: outputFile,
            Duration:   result.Duration,
            Provider:   result.Provider,
            Usage:      result.Usage,
        }
    }
}
```

### 9.2 Atualizar `cmd/root.go` - Passar Config para Wizard

```go
func launchTUIWizard() {
    // ...
    
    wizardConfig := &ui.WizardConfig{
        LLM: ui.LLMConfig{
            Provider:       viper.GetString("llm.provider"),
            APIKey:         viper.GetString("llm.api-key"),
            BaseURL:        viper.GetString("llm.base-url"),
            Model:          viper.GetString("llm.model"),
            Timeout:        viper.GetInt("llm.timeout"),
            SaveResponse:   viper.GetBool("gemini.save-response"),
            // Legado GeminiWeb
            BinaryPath:     viper.GetString("gemini.binary-path"),
            BrowserRefresh: viper.GetString("gemini.browser-refresh"),
        },
        // Manter Gemini para retrocompatibilidade
        Gemini: ui.GeminiConfig{
            BinaryPath:     viper.GetString("gemini.binary-path"),
            Model:          viper.GetString("gemini.model"),
            Timeout:        viper.GetInt("gemini.timeout"),
            BrowserRefresh: viper.GetString("gemini.browser-refresh"),
            SaveResponse:   viper.GetBool("gemini.save-response"),
        },
        Context: ui.ContextConfig{
            IncludeTree:    viper.GetBool("context.include-tree"),
            IncludeSummary: viper.GetBool("context.include-summary"),
            MaxSize:        viper.GetString("context.max-size"),
        },
    }
    
    // ...
}
```

---

## Fase 10: Testes

### 10.1 Testes da Interface Provider

**Arquivo**: `internal/core/llm/provider_test.go`

```go
package llm

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestIsValidProvider(t *testing.T) {
    tests := []struct {
        provider string
        want     bool
    }{
        {"openai", true},
        {"anthropic", true},
        {"gemini", true},
        {"geminiweb", true},
        {"invalid", false},
        {"", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.provider, func(t *testing.T) {
            got := IsValidProvider(tt.provider)
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestConfigValidate(t *testing.T) {
    tests := []struct {
        name    string
        cfg     Config
        wantErr bool
    }{
        {
            name: "valid openai config",
            cfg: Config{
                Provider: ProviderOpenAI,
                APIKey:   "sk-test",
                Model:    "gpt-4o",
                Timeout:  300,
            },
            wantErr: false,
        },
        {
            name: "missing api key for openai",
            cfg: Config{
                Provider: ProviderOpenAI,
                Model:    "gpt-4o",
                Timeout:  300,
            },
            wantErr: true,
        },
        {
            name: "geminiweb without api key is ok",
            cfg: Config{
                Provider: ProviderGeminiWeb,
                Model:    "gemini-2.5-flash",
                Timeout:  300,
            },
            wantErr: false,
        },
        {
            name: "invalid provider",
            cfg: Config{
                Provider: "invalid",
                APIKey:   "key",
                Model:    "model",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.cfg.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 10.2 Testes do Cliente OpenAI (Mock)

**Arquivo**: `internal/platform/openai/client_test.go`

```go
package openai

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

func TestClient_Send_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/chat/completions", r.URL.Path)
        assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
        
        resp := ChatCompletionResponse{
            ID:    "test-id",
            Model: "gpt-4o",
            Choices: []Choice{
                {
                    Message: Message{Role: "assistant", Content: "Hello!"},
                    FinishReason: "stop",
                },
            },
            Usage: UsageAPI{
                PromptTokens:     10,
                CompletionTokens: 5,
                TotalTokens:      15,
            },
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
    }))
    defer server.Close()
    
    client, err := NewClient(llm.Config{
        APIKey:  "test-key",
        BaseURL: server.URL,
        Model:   "gpt-4o",
    })
    require.NoError(t, err)
    
    result, err := client.Send(context.Background(), "test prompt")
    require.NoError(t, err)
    
    assert.Equal(t, "Hello!", result.Response)
    assert.Equal(t, "OpenAI", result.Provider)
    assert.NotNil(t, result.Usage)
    assert.Equal(t, 15, result.Usage.TotalTokens)
}

func TestClient_Send_APIError(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        resp := ErrorResponse{}
        resp.Error.Message = "Invalid API key"
        resp.Error.Type = "invalid_request_error"
        
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(resp)
    }))
    defer server.Close()
    
    client, err := NewClient(llm.Config{
        APIKey:  "bad-key",
        BaseURL: server.URL,
        Model:   "gpt-4o",
    })
    require.NoError(t, err)
    
    _, err = client.Send(context.Background(), "test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "Invalid API key")
}
```

### 10.3 Testes de Integracao (Opcional)

**Arquivo**: `internal/platform/openai/integration_test.go`

```go
//go:build integration

package openai

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/quantmind-br/shotgun-cli/internal/core/llm"
)

func TestClient_Integration_RealAPI(t *testing.T) {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        t.Skip("OPENAI_API_KEY not set")
    }
    
    client, err := NewClient(llm.Config{
        APIKey: apiKey,
        Model:  "gpt-4o-mini",
    })
    require.NoError(t, err)
    
    result, err := client.Send(context.Background(), "Say hello in one word")
    require.NoError(t, err)
    require.NotEmpty(t, result.Response)
    
    t.Logf("Response: %s", result.Response)
    t.Logf("Tokens: %d", result.Usage.TotalTokens)
    t.Logf("Duration: %s", result.Duration)
}
```

---

## Fase 11: Documentacao

### 11.1 Atualizar README

```markdown
## LLM Integration

shotgun-cli supports multiple LLM providers:

### Supported Providers

| Provider | API Key Required | Models |
|----------|------------------|--------|
| openai | Yes | gpt-4o, gpt-4-turbo, o1-preview |
| anthropic | Yes | claude-3-5-sonnet, claude-3-opus |
| gemini | Yes | gemini-2.5-flash, gemini-2.5-pro |
| geminiweb | No (browser auth) | gemini-2.5-flash |

### Quick Setup

#### OpenAI
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.api-key sk-your-api-key
shotgun-cli config set llm.model gpt-4o
```

#### Anthropic (Claude)
```bash
shotgun-cli config set llm.provider anthropic
shotgun-cli config set llm.api-key sk-ant-your-api-key
shotgun-cli config set llm.model claude-3-5-sonnet-latest
```

#### Google Gemini
```bash
shotgun-cli config set llm.provider gemini
shotgun-cli config set llm.api-key AIza-your-api-key
shotgun-cli config set llm.model gemini-2.5-flash
```

#### GeminiWeb (Browser-based, no API key)
```bash
# Install geminiweb
go install github.com/diogo/geminiweb/cmd/geminiweb@latest
geminiweb auto-login

# Configure
shotgun-cli config set llm.provider geminiweb
shotgun-cli config set gemini.enabled true
```

### Custom Endpoints (OpenRouter, Azure, etc.)

For OpenRouter or other OpenAI-compatible endpoints:
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.base-url https://openrouter.ai/api/v1
shotgun-cli config set llm.api-key your-openrouter-key
shotgun-cli config set llm.model anthropic/claude-3.5-sonnet
```

### Diagnostics

```bash
shotgun-cli llm status   # Show current configuration
shotgun-cli llm doctor   # Diagnose issues
shotgun-cli llm list     # List supported providers
```
```

---

## Fase 12: Checklist de Implementacao

### Arquivos a Criar
- [ ] `internal/core/llm/provider.go` - Interface Provider
- [ ] `internal/core/llm/config.go` - Config unificada
- [ ] `internal/core/llm/registry.go` - Factory/Registry
- [ ] `internal/core/llm/provider_test.go` - Testes
- [ ] `internal/platform/openai/client.go` - Cliente OpenAI
- [ ] `internal/platform/openai/types.go` - Tipos request/response
- [ ] `internal/platform/openai/models.go` - Modelos validos
- [ ] `internal/platform/openai/client_test.go` - Testes
- [ ] `internal/platform/anthropic/client.go` - Cliente Anthropic
- [ ] `internal/platform/anthropic/types.go` - Tipos request/response
- [ ] `internal/platform/anthropic/models.go` - Modelos validos
- [ ] `internal/platform/anthropic/client_test.go` - Testes
- [ ] `internal/platform/geminiapi/client.go` - Cliente Gemini API
- [ ] `internal/platform/geminiapi/types.go` - Tipos request/response
- [ ] `internal/platform/geminiapi/models.go` - Modelos validos
- [ ] `internal/platform/geminiapi/client_test.go` - Testes
- [ ] `internal/platform/gemini/provider.go` - Adapter GeminiWeb
- [ ] `cmd/providers.go` - Registro de providers
- [ ] `cmd/config_llm.go` - Helper para config LLM
- [ ] `cmd/llm.go` - Comando `llm status/doctor/list`

### Arquivos a Modificar
- [ ] `cmd/root.go` - Novos defaults
- [ ] `cmd/config.go` - Novas validacoes
- [ ] `cmd/send.go` - Usar Provider interface
- [ ] `internal/ui/wizard.go` - LLMConfig atualizado
- [ ] `internal/ui/screens/review.go` - Provider availability check
- [ ] `go.mod` - Nenhuma nova dependencia necessaria (HTTP puro)

### Configuracoes Novas
- [ ] `llm.provider` - "openai", "anthropic", "gemini", "geminiweb"
- [ ] `llm.api-key` - Chave da API
- [ ] `llm.base-url` - URL customizada (opcional)
- [ ] `llm.model` - Modelo a usar
- [ ] `llm.timeout` - Timeout em segundos

### Testes
- [ ] Testes unitarios para interface Provider
- [ ] Testes unitarios para cada cliente (mock HTTP)
- [ ] Testes de integracao (opcional, requer API keys reais)
- [ ] Teste E2E do fluxo completo

---

## Ordem de Execucao Recomendada

1. **Fase 1**: Criar interface Provider, Config e Registry no core
2. **Fase 2**: Implementar cliente OpenAI
3. **Fase 3**: Implementar cliente Anthropic
4. **Fase 4**: Implementar cliente Gemini API
5. **Fase 5**: Adaptar GeminiWeb existente para interface
6. **Fase 6**: Criar registro de providers no cmd
7. **Fase 7**: Atualizar sistema de configuracao
8. **Fase 8**: Atualizar comandos CLI
9. **Fase 9**: Atualizar TUI Wizard
10. **Fase 10**: Escrever testes
11. **Fase 11**: Documentar
12. **Fase 12**: Revisar e testar E2E

---

## Estimativa de Esforco

| Fase | Complexidade | Tempo Estimado |
|------|--------------|----------------|
| 1    | Media        | 1 hora         |
| 2    | Media        | 1.5 horas      |
| 3    | Media        | 1.5 horas      |
| 4    | Media        | 1 hora         |
| 5    | Baixa        | 30 min         |
| 6    | Baixa        | 30 min         |
| 7    | Media        | 1 hora         |
| 8    | Media        | 1.5 horas      |
| 9    | Media        | 1 hora         |
| 10   | Media        | 2 horas        |
| 11   | Baixa        | 30 min         |
| 12   | Baixa        | 30 min         |
| **Total** |        | **~12 horas**  |

---

## Notas Finais

### Vantagens da Arquitetura Multi-Provider
- Flexibilidade para escolher o melhor LLM para cada caso
- Suporte a endpoints customizados (OpenRouter, Azure, proxies)
- Facil adicao de novos providers
- Custo controlado - usuarios podem usar modelos mais baratos
- Resiliencia - se um provider falhar, pode usar outro

### Retrocompatibilidade
- GeminiWeb continua funcionando sem mudancas
- Configuracoes existentes (`gemini.*`) sao respeitadas
- Usuarios existentes nao precisam mudar nada

### Extensibilidade Futura
A interface Provider permite adicionar facilmente:
- Ollama (modelos locais)
- Azure OpenAI
- AWS Bedrock
- Groq
- Together AI
- Qualquer provider OpenAI-compativel

### Seguranca
- API keys nunca sao logadas
- API keys sao mascaradas na exibicao
- Configuracao armazenada em arquivo local com permissoes restritas
