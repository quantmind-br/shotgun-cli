# Código-Fonte — Módulo `internal/platform/anthropic`

> **Nível de detalhamento:** detalhado  
> **Idioma do documento:** pt-br  
> **Data:** 2026-05-20

---

## 1. Visão Geral

| Campo | Valor |
|-------|-------|
| **Nome do módulo** | `internal/platform/anthropic` |
| **Package import path** | `github.com/quantmind-br/shotgun-cli/internal/platform/anthropic` |
| **Tipo** | Adaptador de plataforma — implementação concreta do provedor Anthropic |
| **Propósito** | Implementar o contrato `llm.Provider` via a Anthropic Messages API (`/v1/messages`) |
| **Arquivos analisados** | 4 (3 fontes + 1 teste) |

### Relações de dependência

```
internal/platform/anthropic
├── depends on → internal/core/llm         (Provider interface, Result, Usage, Config)
├── depends on → internal/platform/llmbase (BaseClient, Sender interface, Strategy)
└── depends on → internal/platform/http    (JSONClient, HTTPError) [transitiva via llmbase]
```

**Dependência externa:** `github.com/stretchr/testify` (apenas no teste).

---

## 2. Arquivo `types.go`

### 2.1 Visão Geral

Define todas as estruturas de dados utilizadas nas comunicações com a API Anthropic. É o **único** arquivo que não depende de `llm` ou `llmbase` (zero dependências internas), sendo puramente dados.

### 2.2 Tipos Detalhados

#### `MessagesRequest` (struct, linhas ~6-11)

```go
type MessagesRequest struct {
    Model     string    `json:"model"`
    MaxTokens int       `json:"max_tokens"`
    Messages  []Message `json:"messages"`
    System    string    `json:"system,omitempty"`
    Stream    bool      `json:"stream,omitempty"`
}
```

| Campo | Tipo | Tag JSON | Opcional? | Descrição |
|-------|------|----------|-----------|-----------|
| `Model` | `string` | `model` | Não | Identificador do modelo Claude (ex: `claude-sonnet-4-20250514`) |
| `MaxTokens` | `int` | `max_tokens` | Não | Limite máximo de tokens na resposta |
| `Messages` | `[]Message` | `messages` | Não | Lista de mensagens da conversa |
| `System` | `string` | `system,omitempty` | Sim | Instruções do sistema (omissas se vazias) |
| `Stream` | `bool` | `stream,omitempty` | Sim | Habilita streaming (sempre `false` na implementação atual) |

**Observação:** O campo `System` permite envio de system prompt, embora o `BuildRequest()` do client atual **nunca o preencha**. Isso representa uma funcionalidade não utilizada (dead code path).

#### `Message` (struct, linhas ~13-16)

```go
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}
```

| Campo | Tipo | Valores possíveis |
|-------|------|-------------------|
| `Role` | `string` | `"user"`, `"assistant"` |
| `Content` | `string` | Texto livre da mensagem |

#### `MessagesResponse` (struct, linhas ~18-29)

```go
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
```

Todos os campos são lidos via `json.Unmarshal` pelo `JSONClient.PostJSON()`. O `ParseResponse()` consome apenas `Content` e `Usage`.

#### `ContentBlock` (struct, linhas ~31-35)

```go
type ContentBlock struct {
    Type string `json:"type"`
    Text string `json:"text"`
}
```

**Observação:** O `ParseResponse()` itera sobre `Content` e concatena blocos do tipo `"text"`. Blocos de outros tipos (ex: `"tool_use"`, `"thinking"`) são **silentemente ignorados**. Isto pode causar perda de dados se a API retornar blocos não-texto.

#### `UsageAPI` (struct, linhas ~37-40)

```go
type UsageAPI struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
}
```

#### `ErrorResponse` (struct, linhas ~42-49)

```go
type ErrorResponse struct {
    Type  string `json:"type"`
    Error struct {
        Type    string `json:"type"`
        Message string `json:"message"`
    } `json:"error"`
}
```

Usado no `handleError()` para extrair a mensagem de erro da resposta JSON em casos de falha.

---

## 3. Arquivo `models.go`

### 3.1 Visão Geral

Arquivo mínimo (3 funções) que fornece metadados sobre os modelos Anthropic suportados.

### 3.2 Funções Detalhadas

#### `ValidModels() []string` (linhas ~3-14)

Retorna a lista fixa de 8 modelos conhecidos:

| # | Modelo | Família |
|---|--------|---------|
| 1 | `claude-sonnet-4-20250514` | Sonnet (padrão) |
| 2 | `claude-3-5-sonnet-latest` | Sonnet 3.5 |
| 3 | `claude-3-5-sonnet-20241022` | Sonnet 3.5 (data) |
| 4 | `claude-3-5-haiku-latest` | Haiku 3.5 |
| 5 | `claude-3-opus-latest` | Opus 3 |
| 6 | `claude-3-opus-20240229` | Opus 3 (data) |
| 7 | `claude-3-sonnet-20240229` | Sonnet 3 |
| 8 | `claude-3-haiku-20240307` | Haiku 3 |

**Padrão de uso:** chamada pela UI para popular menus de seleção de modelo.

#### `IsKnownModel(model, baseURL string) bool` (linhas ~16-19)

```go
func IsKnownModel(model, baseURL string) bool {
    return true
}
```

**🟡 INFERIDO:** A docstring indica que a validação foi removida propositalmente para permitir modelos customizados/preview. A função foi marcada como `@deprecated`. Isto significa que **nenhum modelo será rejeitado** pelo validador.

---

## 4. Arquivo `client.go`

### 4.1 Visão Geral

O arquivo principal que implementa a estratégia `llmbase.Sender` e define o adaptador `Client`. Contém a lógica de construção de requisições, parsing de respostas e tratamento de erros.

### 4.2 Constantes (linhas ~10-13)

| Constante | Valor | Propósito |
|-----------|-------|-----------|
| `defaultBaseURL` | `"https://api.anthropic.com"` | Endpoint base da API |
| `anthropicVersion` | `"2023-06-01"` | Versão da API (header `anthropic-version`) |
| `defaultMaxTokens` | `8192` | Limite padrão de tokens |

### 4.3 Tipo `Client`

```go
type Client struct {
    *llmbase.BaseClient
}
```

Embeds `BaseClient` (do pacote `llmbase`), herdando:
- `JSONClient` (HTTP client JSON)
- `APIKey`, `Model`, `MaxTokens`, `ProviderName`

**Padrão de projeto:** Herança via embedding para reuso. O `Client` é um *concrete strategy* que implementa os métodos da interface `llmbase.Sender`.

### 4.4 Função `NewClient(cfg llm.Config)`

```go
func NewClient(cfg llm.Config) (*Client, error)
```

**Responsabilidades:**
1. Chama `llmbase.NewBaseClientWithDefaults()` com valores padrão do Anthropic
2. Aplica defaults: base URL, modelo (`claude-sonnet-4-20250514`), max tokens (8192), timeout (300s)
3. Valida que a API key não está vazia (via `NewBaseClientWithDefaults`)
4. Retorna `&Client{BaseClient: base}`

**Fluxo de erro:** Se `cfg.APIKey == ""`, retorna `"api key is required"` (do `llmbase`).

### 4.5 Método `Send(ctx, content) → *llm.Result, error`

```go
func (c *Client) Send(ctx context.Context, content string) (*llm.Result, error)
```

**Fluxo:**
1. Delega para `c.BaseClient.Send(ctx, content, c)` — o `c` é passado como `Sender`
2. `BaseClient.Send` executa o ciclo: `BuildRequest → GetHeaders → GetEndpoint → PostJSON → ParseResponse`
3. Em caso de erro, delega para `c.handleError(err)` antes de retornar
4. Retorna o `llm.Result` processado

### 4.6 Método `SendWithProgress(ctx, content, progress) → *llm.Result, error`

```go
func (c *Client) SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*llm.Result, error)
```

**Fluxo:**
1. Delega para `c.BaseClient.SendWithProgress(ctx, content, c, progress)`
2. O `BaseClient` invoca `progress("Connecting to Anthropic...")` antes da requisição
3. Se sem erro, invoca `progress("Response received")` após
4. Em caso de erro, delega para `c.handleError(err)`

**Observação:** A progressão é mínima — apenas dois callbacks. Não há progresso incremental durante a leitura da resposta (a API não é SSE/streaming na implementação atual).

### 4.7 Método `BuildRequest(content string) → (interface{}, error)`

Implementa a estratégia `Sender.BuildRequest`:

```go
func (c *Client) BuildRequest(content string) (interface{}, error) {
    return MessagesRequest{
        Model:     c.Model,
        MaxTokens: c.MaxTokens,
        Messages: []Message{
            {Role: "user", Content: content},
        },
    }, nil
}
```

- Cria um array com **uma única mensagem** do tipo `"user"`
- **NÃO** preenche o campo `System` (ver observação em `types.go`)
- **NÃO** ativa `Stream`
- O modelo e max tokens vêm do `Client` (já com defaults aplicados)

### 4.8 Método `ParseResponse(response, rawJSON) → *llm.Result, error`

Implementa a estratégia `Sender.ParseResponse`:

```go
func (c *Client) ParseResponse(response interface{}, rawJSON []byte) (*llm.Result, error)
```

**Fluxo:**
1. Tipo assert `*MessagesResponse` — retorna erro se tipo inesperado
2. Concatena todos os `ContentBlock` do tipo `"text"` (ignorando outros tipos)
3. Mapeia `UsageAPI.InputTokens → Usage.PromptTokens`
4. Mapeia `UsageAPI.OutputTokens → Usage.CompletionTokens`
5. `Usage.TotalTokens = Input + Output`
6. Retorna `llm.Result` com `Response`, `RawResponse`, `Model`, `Provider`, `Usage`

**🟡 INFERIDO:** O campo `msgResp.Usage.InputTokens > 0 \|\| msgResp.Usage.OutputTokens > 0` é o guard para decidir se `usage` é não-nulo. Se ambos forem 0, o `Usage` será `nil` no `Result`. Isto é aceitável porque a Anthropic sempre reporta usage em respostas bem-sucedidas.

### 4.9 Método `GetEndpoint() → string`

Retorna `"/v1/messages"` — o endpoint da Anthropic Messages API.

### 4.10 Método `GetHeaders() → map[string]string`

Retorna:
```json
{
  "x-api-key": "<api-key-do-client>",
  "anthropic-version": "2023-06-01"
}
```

A `Content-Type: application/json` é adicionada pelo `JSONClient`.

### 4.11 Método `NewResponse() → interface{}`

Retorna `&MessagesResponse{}` — um *factory method* para a struct de resposta, usado pelo `BaseClient` antes de `PostJSON`.

### 4.12 Método `GetProviderName() → string`

Retorna `c.ProviderName` (herdado de `BaseClient`, que é `"Anthropic"` por padrão).

### 4.13 Método `handleError(err error) → error`

```go
func (c *Client) handleError(err error) error {
    return c.BaseClient.HandleHTTPError(err, func(body []byte) string {
        var errResp ErrorResponse
        if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
            return errResp.Error.Message
        }
        return ""
    })
}
```

**Fluxo:**
1. Delega para `BaseClient.HandleHTTPError()`
2. Passa uma closure que tenta deserializar o corpo como `ErrorResponse`
3. Se a deserialização succeeds e há `Error.Message`, retorna a mensagem
4. Se falhar, retorna string vazia (o `HandleHTTPError` cai no fallback de corpo bruto)

**Resultado de erro típico:** `"API error [401]: Invalid API key"` ou `"API error [429]: Rate limit exceeded"` etc.

---

## 5. Arquivo `client_test.go`

### 5.1 Visão Geral

Contém 6 casos de teste (5 funções `TestClient_*` + `TestClient_SendWithProgress`) que cobrem o happy path, múltiplos blocos de conteúdo, erros de API, validação de configuração, verificação de configuração e progressão.

### 5.2 Casos de Teste Detalhados

| Função | Cobre |
|--------|-------|
| `TestClient_Send_Success` | Requisição HTTP válida → resposta correta, headers, tokens |
| `TestClient_Send_MultipleContentBlocks` | Concatenação de múltiplos blocos de texto |
| `TestClient_Send_APIError` | Código 401 → mensagem de erro extraída |
| `TestClient_NewClient_Validation` | Ausência de API key gera erro; defaults aplicados corretamente |
| `TestClient_IsConfigured` | `IsConfigured()` retorna true quando API key e modelo presentes |
| `TestClient_SendWithProgress` | Callbacks de progresso são invocados corretamente |

### 5.3 Estratégia de Teste

- Usa `httptest.NewServer` para mocking do endpoint Anthropic
- Usa `testify/assert` e `testify/require` para asserções
- Cria clientes com `BaseURL` sobreposto ao server mock (ignora a URL real)

### 5.4 Lacunas de Teste

| O que falta | Razão |
|------------|-------|
| Teste para resposta com `ContentBlock` de tipo não-texto (ex: `tool_use`) | Não coberto — blocos não-texto são silentemente ignorados |
| Teste para `BuildRequest` com system prompt | `BuildRequest` não suporta system prompt na implementação atual |
| Teste para timeout | O timeout é do `JSONClient` (testado em outro módulo) |
| Teste para `ValidModels()` | Nenhum teste explícito para a lista de modelos |

---

## 6. Padrões de Projeto Identificados

| Padrão | Onde | Descrição |
|--------|------|-----------|
| **Strategy** | `llmbase.Sender` interface, `Client` como impl | Cada provedor (Anthropic, OpenAI, Gemini) implementa `Sender` para personalizar request/response |
| **Template Method** | `BaseClient.Send()` | Algoritmo fixo: build → headers → endpoint → POST → parse, com callbacks variáveis por estratégia |
| **Factory** | `NewClient()`, `NewBaseClientWithDefaults()` | Criação de clientes com defaults aplicados |
| **Delegation** | `Client` embeds `BaseClient` | Delega HTTP ao `BaseClient`, apenas a transformação específica do Anthropic é feita no `Client` |
| **Adapter** | `Client` implementa `llm.Provider` | Adapta a API específica do Anthropic ao contrato unificado `llm.Provider` |

---

## 7. Riscos e Observações de Código

### 7.1 Bloqueio de Conteúdo Não-Texto (`🔴 Risco Médio`)

O `ParseResponse()` itera sobre `ContentBlock` e concatena apenas blocos de tipo `"text"`. A Anthropic Messages API pode retornar blocos de outros tipos (`tool_use`, `thinking`, `reflection`) quando o modelo usa ferramentas ou modo thinking. Estes são **silentemente ignorados**, resultando em respostas incompletas.

### 7.2 System Prompt Não Utilizado (`🟡 Risco Baixo`)

`MessagesRequest` possui campo `System`, mas `BuildRequest()` nunca o preenche. O sistema poderia enviar system prompts via configuração adicional.

### 7.3 Sem Retrying (`🔴 Risco Médio`)

Nenhuma política de retry é implementada. Falhas transitórias (503, timeouts) resultam em erro imediato. Isto é mencionado como dívida técnica L3 no relatório do arquiteto.

### 7.4 Sem Rate Limiting (`🔴 Risco Crítico`)

Não há controle de taxa antes de enviar requisições. Em uso intensivo, o cliente pode exceder os limites da API.

### 7.5 Versão da API Hardcoded

`anthropicVersion = "2023-06-01"` é uma constante. Atualizações de versão requerem mudança de código.

---

## 8. Métricas de Código

| Métrica | Valor |
|---------|-------|
| Arquivos fonte | 3 |
| Linhas totais (fontes) | ~175 |
| Linhas de teste | ~155 |
| Funções/métodos (fontes) | 10 |
| Tipos (structs) | 5 |
| Constantes | 3 |
| Importações externas (fonte) | 1 (`testify` em testes) |
| Importações internas | 2 (`llm`, `llmbase`) |
| Funções de teste | 6 |
| Cobertura (estimada) | ~65-70% (funções principais cobertas, edge cases não) |
