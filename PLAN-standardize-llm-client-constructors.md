# Plano de Implementacao: Standardize LLM Client Constructors

## Resumo Executivo

Este plano consolida a logica duplicada de inicializacao nos construtores de clientes LLM (`openai`, `anthropic`, `geminiapi`) em um helper centralizado no pacote `llmbase`. A mudanca elimina ~60 linhas de codigo duplicado, reduz risco de comportamento inconsistente entre providers e simplifica a adicao de novos providers.

## Analise de Requisitos

### Requisitos Funcionais
- [ ] Criar `ProviderDefaults` struct em `llmbase` para encapsular valores padrao por provider
- [ ] Criar `NewValidatedConfig()` helper que aplica defaults e valida configuracao
- [ ] Refatorar `NewClient()` em cada provider para usar o helper centralizado
- [ ] Manter compatibilidade total com a API existente (sem breaking changes)

### Requisitos Nao-Funcionais
- [ ] Manter ou melhorar cobertura de testes existente
- [ ] Zero breaking changes na interface publica
- [ ] Mensagens de erro consistentes entre providers

## Analise Tecnica

### Arquitetura Proposta

```
llm.Config (input)
      |
      v
+---------------------------+
| llmbase.NewValidatedConfig |  <-- NOVO: Validacao + defaults centralizados
| - Valida API key           |
| - Aplica BaseURL default   |
| - Aplica Model default     |
| - Converte Timeout         |
| - Aplica MaxTokens default |
+---------------------------+
      |
      v
llmbase.Config (validated)
      |
      v
llmbase.NewBaseClient()
      |
      v
Provider-specific Client (openai.Client, anthropic.Client, etc)
```

### Componentes Afetados

| Arquivo/Modulo | Tipo de Mudanca | Descricao |
|----------------|-----------------|-----------|
| `internal/platform/llmbase/config.go` | Criar | Novo arquivo com `ProviderDefaults` e `NewValidatedConfig()` |
| `internal/platform/llmbase/config_test.go` | Criar | Testes para novo helper |
| `internal/platform/openai/client.go` | Modificar | Simplificar `NewClient()` |
| `internal/platform/anthropic/client.go` | Modificar | Simplificar `NewClient()` |
| `internal/platform/geminiapi/client.go` | Modificar | Simplificar `NewClient()` |

### Dependencias
- Nenhuma nova dependencia de pacote
- Depende de `internal/core/llm` para `llm.Config`
- Depende de `internal/platform/http` para `platformhttp.JSONClient`

## Plano de Implementacao

### Fase 1: Criar Helper Centralizado

**Objetivo**: Implementar `ProviderDefaults` e `NewValidatedConfig()` em `llmbase`

#### Tarefas:

1. **Criar `internal/platform/llmbase/config.go`**
   
   Arquivos envolvidos: `internal/platform/llmbase/config.go`
   
   ```go
   package llmbase

   import (
       "fmt"
       "time"

       "github.com/quantmind-br/shotgun-cli/internal/core/llm"
   )

   // ProviderDefaults holds provider-specific default values.
   type ProviderDefaults struct {
       BaseURL      string
       Model        string
       Timeout      time.Duration
       MaxTokens    int
       ProviderName string
   }

   // NewValidatedConfig creates a validated llmbase.Config from llm.Config with defaults applied.
   // Returns error if required fields are missing.
   func NewValidatedConfig(cfg llm.Config, defaults ProviderDefaults) (*Config, error) {
       // Validate API key (required for all HTTP providers)
       if cfg.APIKey == "" {
           return nil, fmt.Errorf("api key is required")
       }

       // Apply BaseURL default
       baseURL := cfg.BaseURL
       if baseURL == "" {
           baseURL = defaults.BaseURL
       }

       // Apply Model default
       model := cfg.Model
       if model == "" {
           model = defaults.Model
       }

       // Convert and apply Timeout default
       timeout := time.Duration(cfg.Timeout) * time.Second
       if timeout == 0 {
           timeout = defaults.Timeout
       }
       if timeout == 0 {
           timeout = 300 * time.Second // fallback
       }

       // Apply MaxTokens default
       maxTokens := cfg.MaxTokens
       if maxTokens == 0 {
           maxTokens = defaults.MaxTokens
       }

       return &Config{
           APIKey:    cfg.APIKey,
           BaseURL:   baseURL,
           Model:     model,
           Timeout:   timeout,
           MaxTokens: maxTokens,
       }, nil
   }
   ```

2. **Criar testes para `NewValidatedConfig()`**
   
   Arquivos envolvidos: `internal/platform/llmbase/config_test.go`
   
   ```go
   package llmbase

   import (
       "testing"
       "time"

       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/require"

       "github.com/quantmind-br/shotgun-cli/internal/core/llm"
   )

   func TestNewValidatedConfig_Success(t *testing.T) {
       t.Parallel()
       
       defaults := ProviderDefaults{
           BaseURL:      "https://api.example.com",
           Model:        "default-model",
           Timeout:      300 * time.Second,
           MaxTokens:    8192,
           ProviderName: "Test",
       }
       
       cfg, err := NewValidatedConfig(llm.Config{
           APIKey: "test-key",
       }, defaults)
       
       require.NoError(t, err)
       assert.Equal(t, "test-key", cfg.APIKey)
       assert.Equal(t, "https://api.example.com", cfg.BaseURL)
       assert.Equal(t, "default-model", cfg.Model)
       assert.Equal(t, 300*time.Second, cfg.Timeout)
       assert.Equal(t, 8192, cfg.MaxTokens)
   }

   func TestNewValidatedConfig_CustomValues(t *testing.T) {
       t.Parallel()
       
       defaults := ProviderDefaults{
           BaseURL:   "https://api.example.com",
           Model:     "default-model",
           Timeout:   300 * time.Second,
           MaxTokens: 8192,
       }
       
       cfg, err := NewValidatedConfig(llm.Config{
           APIKey:    "test-key",
           BaseURL:   "https://custom.proxy.com",
           Model:     "custom-model",
           Timeout:   60,
           MaxTokens: 4096,
       }, defaults)
       
       require.NoError(t, err)
       assert.Equal(t, "https://custom.proxy.com", cfg.BaseURL)
       assert.Equal(t, "custom-model", cfg.Model)
       assert.Equal(t, 60*time.Second, cfg.Timeout)
       assert.Equal(t, 4096, cfg.MaxTokens)
   }

   func TestNewValidatedConfig_MissingAPIKey(t *testing.T) {
       t.Parallel()
       
       defaults := ProviderDefaults{
           BaseURL: "https://api.example.com",
           Model:   "default-model",
       }
       
       _, err := NewValidatedConfig(llm.Config{}, defaults)
       
       require.Error(t, err)
       assert.Contains(t, err.Error(), "api key is required")
   }

   func TestNewValidatedConfig_TimeoutFallback(t *testing.T) {
       t.Parallel()
       
       // Even with empty defaults.Timeout, should fallback to 300s
       defaults := ProviderDefaults{
           BaseURL: "https://api.example.com",
           Model:   "default-model",
       }
       
       cfg, err := NewValidatedConfig(llm.Config{
           APIKey: "test-key",
       }, defaults)
       
       require.NoError(t, err)
       assert.Equal(t, 300*time.Second, cfg.Timeout)
   }
   ```

### Fase 2: Refatorar OpenAI Client

**Objetivo**: Simplificar `openai.NewClient()` usando o helper centralizado

#### Tarefas:

1. **Modificar `internal/platform/openai/client.go`**
   
   De (~30 linhas):
   ```go
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

       model := cfg.Model
       if model == "" {
           model = "gpt-4o"
       }

       return &Client{
           BaseClient: llmbase.NewBaseClient(llmbase.Config{
               APIKey:    cfg.APIKey,
               BaseURL:   baseURL,
               Model:     model,
               MaxTokens: cfg.MaxTokens,
               Timeout:   timeout,
           }, "OpenAI"),
       }, nil
   }
   ```
   
   Para (~10 linhas):
   ```go
   func NewClient(cfg llm.Config) (*Client, error) {
       baseCfg, err := llmbase.NewValidatedConfig(cfg, llmbase.ProviderDefaults{
           BaseURL:      defaultBaseURL,
           Model:        "gpt-4o",
           Timeout:      300 * time.Second,
           ProviderName: "OpenAI",
       })
       if err != nil {
           return nil, err
       }

       return &Client{
           BaseClient: llmbase.NewBaseClient(*baseCfg, "OpenAI"),
       }, nil
   }
   ```

2. **Verificar testes existentes**
   
   Os testes em `internal/platform/openai/client_test.go` devem passar sem modificacao, pois a interface publica nao muda.

### Fase 3: Refatorar Anthropic Client

**Objetivo**: Simplificar `anthropic.NewClient()` usando o helper centralizado

#### Tarefas:

1. **Modificar `internal/platform/anthropic/client.go`**
   
   De (~35 linhas):
   ```go
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

       model := cfg.Model
       if model == "" {
           model = "claude-sonnet-4-20250514"
       }

       maxTokens := cfg.MaxTokens
       if maxTokens == 0 {
           maxTokens = defaultMaxTokens
       }

       return &Client{
           BaseClient: llmbase.NewBaseClient(llmbase.Config{...}, "Anthropic"),
       }, nil
   }
   ```
   
   Para (~10 linhas):
   ```go
   func NewClient(cfg llm.Config) (*Client, error) {
       baseCfg, err := llmbase.NewValidatedConfig(cfg, llmbase.ProviderDefaults{
           BaseURL:      defaultBaseURL,
           Model:        "claude-sonnet-4-20250514",
           Timeout:      300 * time.Second,
           MaxTokens:    defaultMaxTokens,
           ProviderName: "Anthropic",
       })
       if err != nil {
           return nil, err
       }

       return &Client{
           BaseClient: llmbase.NewBaseClient(*baseCfg, "Anthropic"),
       }, nil
   }
   ```

### Fase 4: Refatorar Gemini API Client

**Objetivo**: Simplificar `geminiapi.NewClient()` usando o helper centralizado

#### Tarefas:

1. **Modificar `internal/platform/geminiapi/client.go`**
   
   Para (~10 linhas):
   ```go
   func NewClient(cfg llm.Config) (*Client, error) {
       baseCfg, err := llmbase.NewValidatedConfig(cfg, llmbase.ProviderDefaults{
           BaseURL:      defaultBaseURL,
           Model:        "gemini-2.5-flash",
           Timeout:      300 * time.Second,
           MaxTokens:    defaultMaxTokens,
           ProviderName: "Gemini",
       })
       if err != nil {
           return nil, err
       }

       return &Client{
           BaseClient: llmbase.NewBaseClient(*baseCfg, "Gemini"),
       }, nil
   }
   ```

### Fase 5: Validacao Final

**Objetivo**: Garantir que todas as mudancas funcionam corretamente

#### Tarefas:

1. **Executar suite de testes completa**
   ```bash
   go test -race ./internal/platform/...
   ```

2. **Verificar lint**
   ```bash
   golangci-lint run ./internal/platform/...
   ```

3. **Teste de integracao manual**
   - Testar `shotgun-cli llm status` com cada provider configurado
   - Testar envio para cada provider (se keys disponiveis)

## Estrategia de Testes

### Testes Unitarios
- [x] `TestNewValidatedConfig_Success` - Config valida com defaults
- [x] `TestNewValidatedConfig_CustomValues` - Override de todos os campos
- [x] `TestNewValidatedConfig_MissingAPIKey` - Erro para API key vazia
- [x] `TestNewValidatedConfig_TimeoutFallback` - Fallback de timeout

### Testes de Integracao
- [ ] Executar testes existentes de cada provider (`openai`, `anthropic`, `geminiapi`)
- [ ] Verificar que testes httptest existentes continuam passando

### Casos de Teste Especificos

| ID | Cenario | Input | Output Esperado |
|----|---------|-------|-----------------|
| TC01 | API key vazia | `llm.Config{}` | Error "api key is required" |
| TC02 | Apenas API key | `llm.Config{APIKey: "x"}` | Config com todos os defaults |
| TC03 | Override de BaseURL | `llm.Config{APIKey: "x", BaseURL: "custom"}` | Config com BaseURL customizada |
| TC04 | Timeout zero | `llm.Config{APIKey: "x", Timeout: 0}` | Config com timeout default (300s) |
| TC05 | Timeout customizado | `llm.Config{APIKey: "x", Timeout: 60}` | Config com timeout 60s |

## Riscos e Mitigacoes

| Risco | Probabilidade | Impacto | Mitigacao |
|-------|---------------|---------|-----------|
| Testes existentes falham | Baixo | Alto | Interface publica identica; rodar testes cedo |
| Mensagens de erro diferentes | Baixo | Baixo | Usar mesma mensagem "api key is required" |
| Comportamento sutil diferente | Baixo | Medio | Comparar logica linha a linha antes de deletar |

## Checklist de Conclusao

- [ ] `internal/platform/llmbase/config.go` criado com `ProviderDefaults` e `NewValidatedConfig()`
- [ ] `internal/platform/llmbase/config_test.go` criado com testes abrangentes
- [ ] `internal/platform/openai/client.go` refatorado
- [ ] `internal/platform/anthropic/client.go` refatorado
- [ ] `internal/platform/geminiapi/client.go` refatorado
- [ ] Todos os testes existentes passando (`go test -race ./internal/platform/...`)
- [ ] Lint passando (`golangci-lint run ./internal/platform/...`)
- [ ] Codigo revisado

## Notas Adicionais

### Consideracoes sobre `llm.Config.WithDefaults()`

O pacote `internal/core/llm/config.go` ja possui `WithDefaults()` e `DefaultConfigs()`. Porem:
- Essas funcoes operam em `llm.Config` (segundos para timeout)
- Os providers precisam converter para `time.Duration`
- A validacao de API key nao esta centralizada

A nova abordagem em `llmbase` e complementar:
- `llm.Config.WithDefaults()` pode ser usada antes de passar para o provider
- `llmbase.NewValidatedConfig()` faz a conversao e validacao no ponto de uso

### Ordem de Implementacao Recomendada

1. Implementar e testar `llmbase/config.go` primeiro
2. Refatorar OpenAI (mais simples, sem MaxTokens obrigatorio)
3. Refatorar Anthropic (adiciona MaxTokens)
4. Refatorar Gemini (endpoint dinamico, mas construtor similar)
5. Validacao final

### Metricas de Sucesso

- **Antes**: ~90 linhas de codigo duplicado em construtores
- **Depois**: ~30 linhas centralizadas + ~30 linhas nos providers
- **Reducao**: ~30 linhas liquidas + melhor manutencao
