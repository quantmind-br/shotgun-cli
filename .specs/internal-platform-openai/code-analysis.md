# Análise de Código — internal/platform/openai

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/platform/openai`
> **Nível de detalhe:** detalhado
> **Arquétipo:** Adaptador de fornecedor LLM (Strategy Pattern)
> **Arquivo-fonte:** Gerado pela fase archaeologist do Reversa

---

## 1. Visão Geral

O pacote `openai` implementa o **provedor OpenAI** do sistema shotgun-cli, seguindo a interface `llm.Provider` definida em `internal/core/llm`. O módulo atua como adaptador que converte requisições internas no formato esperado pela **OpenAI Chat Completions API (v1)**.

### Responsabilidades

1. **Fábrica** — `NewClient()` cria um `*Client` com valores padrão do OpenAI.
2. **Envio de prompts** — `Send()` e `SendWithProgress()` encaminham conteúdo ao endpoint `/chat/completions`.
3. **Serialização/desserialização** — `BuildRequest()` / `ParseResponse()` mapeiam entre o formato Chat Completions e `llm.Result`.
4. **Configuração HTTP** — `GetEndpoint()`, `GetHeaders()`, `NewResponse()` orientam o `BaseClient`.

### Dependências externas

| Pacote | Tipo | Uso |
|--------|------|-----|
| `github.com/quantmind-br/shotgun-cli/internal/core/llm` | Interna (core) | `llm.Config`, `llm.Result`, `llm.Provider` |
| `github.com/quantmind-br/shotgun-cli/internal/platform/llmbase` | Interna (infra) | `llmbase.BaseClient`, `llmbase.Sender`, `llmbase.DefaultConfig` |

### Dependência transitiva

```
openai
  └─ llmbase
       ├─ platformhttp (HTTP transport layer)
       └─ core/llm
  └─ core/llm
```

---

## 2. Estrutura do Pacote

```
openai/
├── client.go          ← Client struct + Send, SendWithProgress, BuildRequest, ParseResponse, GetEndpoint, GetHeaders, NewResponse, GetProviderName, handleError
├── client_test.go     ← 7 testes (Send sucesso/erros/empty, NewClient validation, IsConfigured, ValidateConfig, SendWithProgress)
├── models.go          ← ValidModels(), IsKnownModel() (desprezado — sempre retorna true)
├── models_test.go     ← 2 testes (ValidModels com 9 checks, IsKnownModel com 7 cenários)
└── types.go           ← Todas as structs do request/response Chat Completions (ChatCompletionRequest, Message, Choice, UsageAPI, ChatCompletionResponse, ErrorResponse)
```

**Arquitetura interna:** 3 arquivos de código + 2 arquivos de teste. Os tipos de dados estão isolados em `types.go` para reuso. A lógica de negócio reside em `client.go`.

---

## 3. Análise Detalhada por Arquivo

### 3.1 `client.go` — Motor do Cliente OpenAI (82 linhas)

#### Constantes

```go
const defaultBaseURL = "https://api.openai.com/v1"
```

Identico ao `DefaultConfigs()[ProviderOpenAI].BaseURL` em `core/llm`. **Não há `defaultMaxTokens`** — o OpenAI usa o default do `BaseClient`.

#### Struct `Client`

```go
type Client struct {
    *llmbase.BaseClient
}
```

Composição simples — `Client` herda todos os campos de `BaseClient` (JSONClient, APIKey, Model, MaxTokens, ProviderName).

#### Função `NewClient(cfg llm.Config) (*Client, error)`

- Valida que `cfg.APIKey != ""` (via `llmbase.NewBaseClientWithDefaults`).
- Aplica defaults: `BaseURL="https://api.openai.com/v1"`, `Model="gpt-4o"`, `Timeout=300s`.
- **Nota:** Não aplica `MaxTokens` via `DefaultConfig` — se `cfg.MaxTokens == 0`, o campo fica 0 e `BuildRequest` ignora (não adiciona `max_tokens` ao request).
- ProviderName fixo: `"OpenAI"`.

#### Método `Send(ctx, content) (*llm.Result, error)`

- Delega para `c.BaseClient.Send(ctx, content, c)`.
- O `Client` passa a si mesmo como `Sender` (Strategy Pattern).
- Intercepta erros via `handleError()` para formatação com message do body JSON.
- **Diferença em relação ao Gemini:** O OpenAI retorna o message de erro no body JSON, que é parseado e incluído na mensagem de erro.

#### Método `SendWithProgress(ctx, content, progress) (*llm.Result, error)`

- Delega para `c.BaseClient.SendWithProgress(ctx, content, c, progress)`.
- `BaseClient.SendWithProgress` chama o callback com `"Connecting to OpenAI..."` antes e `"Response received"` após sucesso.
- **Mesmo padrão do Gemini.**

#### Método `BuildRequest(content string) (interface{}, error)`

- Constrói um `ChatCompletionRequest` com:
  - `Model = c.Model`.
  - `Messages = []{{Role: "user", Content: content}}`.
  - `MaxTokens = c.MaxTokens` **apenas se `c.MaxTokens > 0`** (diferente do Gemini, que sempre define).
- **Nota:** `Temperature` e `Stream` **nunca** são configurados pelo adaptador — ficam com zero-values.
- O `Role` é explicitamente `"user"`.

#### Método `ParseResponse(response, rawJSON) (*llm.Result, error)`

- Type-asserts para `*ChatCompletionResponse`.
- Verifica `len(chatResp.Choices) > 0`.
- Extrai `chatResp.Choices[0].Message.Content` como resposta.
- Mapeia `Usage` → `llm.Usage`:
  - `PromptTokens` → `PromptTokens`
  - `CompletionTokens` → `CompletionTokens`
  - `TotalTokens` → `TotalTokens`
  - **Guarda nil se `TotalTokens == 0`** (diferente do Gemini, que mapeia mesmo quando zero).
- Retorna `llm.Result` completo com `RawResponse = string(rawJSON)`.

#### Método `GetEndpoint() string`

- Retorna `"/chat/completions"` — apenas o path, sem query params.

#### Método `GetHeaders() map[string]string`

- Retorna `{"Authorization": "Bearer " + c.APIKey}`.
- A autenticação é via header (diferente do Gemini que usa query param).

#### Método `NewResponse() interface{}`

- Retorna `&ChatCompletionResponse{}` para o `BaseClient` deserializar.

#### Método `GetProviderName() string`

- Retorna `c.ProviderName` (sempre `"OpenAI"`).

#### Método `handleError(err error) error`

- **Diferença crítica do Gemini:** Pára o body JSON e tenta deserializar em `ErrorResponse`.
- Se `errResp.Error.Message != ""`, retorna `"API error [<statusCode>]: <message>"`.
- Isso significa que erros como `"Invalid API key"`, `"Rate limit exceeded"`, `"Model not found"` são reportados com a mensagem original.
- O Gemini descartava o body, o OpenAI o utiliza.

---

### 3.2 `types.go` — Tipos do Request/Response Chat Completions (45 linhas)

Todas as structs são **POCO** com tags `json:` alinhadas à API Chat Completions v1.

| Tipo | Campo | JSON Tag | Tipo |
|------|-------|----------|------|
| `ChatCompletionRequest` | Model | `model` | `string` |
| `ChatCompletionRequest` | Messages | `messages` | `[]Message` |
| `ChatCompletionRequest` | MaxTokens | `max_tokens,omitempty` | `int` |
| `ChatCompletionRequest` | Temperature | `temperature,omitempty` | `float64` |
| `ChatCompletionRequest` | Stream | `stream,omitempty` | `bool` |
| `Message` | Role | `role` | `string` |
| `Message` | Content | `content` | `string` |
| `ChatCompletionResponse` | ID | `id` | `string` |
| `ChatCompletionResponse` | Object | `object` | `string` |
| `ChatCompletionResponse` | Created | `created` | `int64` |
| `ChatCompletionResponse` | Model | `model` | `string` |
| `ChatCompletionResponse` | Choices | `choices` | `[]Choice` |
| `ChatCompletionResponse` | Usage | `usage` | `UsageAPI` |
| `Choice` | Index | `index` | `int` |
| `Choice` | Message | `message` | `Message` |
| `Choice` | FinishReason | `finish_reason` | `string` |
| `UsageAPI` | PromptTokens | `prompt_tokens` | `int` |
| `UsageAPI` | CompletionTokens | `completion_tokens` | `int` |
| `UsageAPI` | TotalTokens | `total_tokens` | `int` |
| `ErrorResponse` | Error.Message | `error.message` | `string` |
| `ErrorResponse` | Error.Type | `error.type` | `string` |
| `ErrorResponse` | Error.Code | `error.code` | `string` |

**Observação:** `ErrorResponse.Error.Code` é `string` (diferente do Gemini que usa `int`). A API OpenAI retorna códigos como `"invalid_api_key"`, não códigos HTTP numéricos.

---

### 3.3 `models.go` — Validação de Modelos (20 linhas)

- `ValidModels()`: retorna lista hardcoded de **9** modelos conhecidos.
- `IsKnownModel()`: **sempre retorna `true`**. Comentário indica que validação foi removida para permitir endpoints custom e modelos preview. Marca como `Deprecated`.

**Lista de modelos (9):**
1. `gpt-4o`
2. `gpt-4o-mini`
3. `gpt-4-turbo`
4. `gpt-4`
5. `gpt-3.5-turbo`
6. `o1-preview`
7. `o1-mini`
8. `o1`
9. `o3-mini`

**Nota:** `o3-mini` é o modelo mais novo listado, indicando manutenção recente. Ausência de `gpt-4o-realtime` ou `gpt-4.5`.

---

## 4. Padrões de Design

| Padrão | Aplicação |
|--------|-----------|
| **Strategy** | `Client` implementa `llmbase.Sender`; `BaseClient` invoca os métodos do sender para operações provider-specific. |
| **Template Method** | `BaseClient.Send()` define o esqueleto (build → post → parse); providers preenchem os métodos do Sender. |
| **Composition** | `Client` contém `*llmbase.BaseClient` (não herança, composição explícita). |
| **Factory** | `NewClient()` / `NewBaseClientWithDefaults()` encapsulam criação + defaults + validação. |

---

## 5. Pontos de Atenção / Riscos

| # | Issue | Severidade | Descrição |
|---|-------|------------|-----------|
| P1 | `MaxTokens` não é default | 🟡 Medium | `NewClient` não define um `MaxTokens` default em `DefaultConfig`. Se `cfg.MaxTokens == 0`, `BuildRequest` pula a linha `req.MaxTokens = c.MaxTokens`, deixando a API usar seu default. O Gemini (8192) tem default explícito. |
| P2 | `Temperature` nunca configurado | 🟡 Medium | O campo `Temperature` existe em `ChatCompletionRequest` e em `llm.Config`, mas **nunca** é lido de `llm.Config` em `BuildRequest`. O valor zero da API OpenAI é 1.0 (default da API), não 0.0. |
| P3 | `IsKnownModel` é um stub | 🟡 Medium | Sempre retorna `true`. `ValidModels()` existe mas não é chamada por lugar algum. |
| P4 | `Usage.TotalTokens == 0` gera nil | ℹ️ Low | `ParseResponse` retorna `usage = nil` quando `TotalTokens == 0`. A maioria das respostas tem tokens, mas em edge cases (modelos com streaming disabled e prompts muito curtos) isso pode acontecer. Consistente com a intenção. |
| P5 | Sem retry/rate-limit | 🔴 High | Identificado também no relatório do architect como debt `L3` (sem retry para LLM calls) e `L4` (sem rate limiting). |
| P6 | `SendWithProgress` não relata falhas | ℹ️ Low | O callback `progress` não é chamado no caminho de erro (mesmo padrão do Gemini). |
| P7 | Duplicação de defaults | 🟡 Medium | `defaultBaseURL` em `client.go` duplica `DefaultConfigs()[ProviderOpenAI].BaseURL` em `core/llm`. `NewClient` não usa os defaults do `core/llm` — aplica seus próprios via `llmbase.DefaultConfig`. |

---

## 6. Cobertura de Testes

### `client_test.go` — 7 testes

| Teste | O que verifica |
|-------|----------------|
| `TestClient_Send_Success` | HTTP 200 com choices + usage → `llm.Result` com tokens corretos, provider "OpenAI", model "gpt-4o" |
| `TestClient_Send_APIError` | HTTP 401 com `ErrorResponse` → error contendo "Invalid API key" e "401" |
| `TestClient_Send_EmptyChoices` | Choices vazio → error "no choices" |
| `TestClient_NewClient_Validation` | APIKey vazia → error "api key is required"; Config válida → client com model "gpt-4o" |
| `TestClient_IsConfigured` | APIKey + Model → true; APIKey + Model vazio → true (🟡 INFERIDO — `IsConfigured` só verifica APIKey); APIKey vazia → false |
| `TestClient_ValidateConfig` | APIKey + Model → nil; APIKey vazia → error; Model vazio → error |
| `TestClient_SendWithProgress` | Callback recebe "Connecting to OpenAI..." e "Response received" |

### `models_test.go` — 2 testes (9 checks + 7 sub-cases)

| Teste | O que verifica |
|-------|----------------|
| `TestValidModels` | Lista tem 9 modelos, verifica 9 itens específicos |
| `TestIsKnownModel` | 7 cenários — todos retornam `true` (conforme stub) |

### Cobertura estimada: **~88% das linhas de código**. Pontos não cobertos:
- Caminho `ParseResponse` com `TotalTokens == 0` → usage nil (não testado explicitamente)
- `BuildRequest` com `MaxTokens == 0` (não testado, mas implícito em Test_NewClient_Validation)
- `GetHeaders()` retorna o header Authorization (não testado, mas implicitamente testado via HTTP mock)

---

## 7. Resumo Executável

O módulo `openai` é um adaptador bem estruturado que segue o padrão Strategy em torno de `BaseClient`. Implementa todos os 6 métodos do `llmbase.Sender` interface, mapeia corretamente request/response Chat Completions, e integra-se com o sistema de progresso via callback.

**Diferenças-chave em relação ao Gemini:**
- Autenticação via header `Authorization: Bearer` (não query param).
- Endpoint `/chat/completions` (não `/models/{model}:generateContent`).
- Parse do body de erro na API (mensagens legíveis como "Invalid API key").
- `MaxTokens` condicional (apenas se > 0), vs. Gemini que sempre define.
- `Usage` pode ser nil se `TotalTokens == 0`, vs. Gemini que sempre mapeia.
- Default model `gpt-4o`, vs. Gemini `gemini-2.5-flash`.
- Sem `defaultMaxTokens` — depende do default da API.

As principais lacunas são: ausência de retry, de rate limiting, de propagação de `Temperature` do config, e de um `MaxTokens` default explícito.
