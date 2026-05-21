# Análise de Código — internal/platform/geminiapi

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/platform/geminiapi`
> **Nível de detalhe:** detalhado
> **Arquétipo:** Adaptador de fornecedor LLM (Strategy Pattern)
> **Arquivo-fonte:** Gerado pela fase archaeologist do Reversa

---

## 1. Visão Geral

O pacote `geminiapi` implementa o **provedor Google Gemini** do sistema shotgun-cli, seguindo a interface `llm.Provider` definida em `internal/core/llm`. O módulo atua como adaptador que converte requisições internas no formato esperado pela **Google Generative Language API (v1beta)**.

### Responsabilidades

1. **Fábrica** — `NewClient()` cria um `*Client` com valores padrão do Gemini.
2. **Envio de prompts** — `Send()` e `SendWithProgress()` encaminham conteúdo ao Gemini.
3. **Serialização/desserialização** — `BuildRequest()` / `ParseResponse()` mapeam entre o formato Gemini e `llm.Result`.
4. **Configuração HTTP** — `GetEndpoint()`, `GetHeaders()`, `NewResponse()` orientam o `BaseClient`.

### Dependências externas

| Pacote | Tipo | Uso |
|--------|------|-----|
| `github.com/quantmind-br/shotgun-cli/internal/core/llm` | Interna (core) | `llm.Config`, `llm.Result`, `llm.Provider` |
| `github.com/quantmind-br/shotgun-cli/internal/platform/llmbase` | Interna (infra) | `llmbase.BaseClient`, `llmbase.Sender`, `llmbase.DefaultConfig` |

### Dependência transitiva

```
geminiapi
  └─ llmbase
       └─ platformhttp (HTTP transport layer)
  └─ core/llm
```

---

## 2. Estrutura do Pacote

```
geminiapi/
├── client.go          ← Client struct + métodos Send, SendWithProgress, BuildRequest, ParseResponse, GetEndpoint, GetHeaders, NewResponse, GetProviderName, handleError
├── client_test.go     ← 7 testes (Send sucesso/erros, NoCandidates, NewClient validation, IsConfigured, SendWithProgress)
├── models.go          ← ValidModels(), IsKnownModel() (desprezado — sempre retorna true)
├── models_test.go     ← 7 testes (ValidModels, IsKnownModel com 7 cenários)
└── types.go           ← Todas as structs do request/response Gemini (GenerateRequest, Content, Part, GenerationConfig, SafetySetting, GenerateResponse, Candidate, UsageMetadata, APIError)
```

**Arquitetura interna:** 3 arquivos de código + 2 arquivos de teste. Os tipos de dados estão isolados em `types.go` para reuso. A lógica de negócio reside em `client.go`.

---

## 3. Análise Detalhada por Arquivo

### 3.1 `client.go` — Motor do Cliente Gemini (67 linhas)

#### Constantes

```go
const (
    defaultBaseURL   = "https://generativelanguage.googleapis.com/v1beta"
    defaultMaxTokens = 8192
)
```

`defaultBaseURL` é idêntico ao `DefaultConfigs()[ProviderGemini].BaseURL` em `core/llm`.
`defaultMaxTokens` (8192) **não** é refletido em `core/llm.DefaultConfigs()` — é um default duplicado em `client.go`.

#### Struct `Client`

```go
type Client struct {
    *llmbase.BaseClient
}
```

Composição simples — `Client` herda todos os campos de `BaseClient` (JSONClient, APIKey, Model, MaxTokens, ProviderName).

#### Função `NewClient(cfg llm.Config) (*Client, error)`

- Valida que `cfg.APIKey != ""` (via `llmbase.NewBaseClientWithDefaults`).
- Aplica defaults: `BaseURL="https://generativelanguage.googleapis.com/v1beta"`, `Model="gemini-2.5-flash"`, `MaxTokens=8192`, `Timeout=300s`.
- ProviderName fixo: `"Gemini"`.

#### Método `Send(ctx, content) (*llm.Result, error)`

- Delega para `c.BaseClient.Send(ctx, content, c)`.
- O `Client` passa a si mesmo como `Sender` (Strategy Pattern).
- Intercepta erros via `handleError()` para formatação consistente.

#### Método `SendWithProgress(ctx, content, progress) (*llm.Result, error)`

- Delega para `c.BaseClient.SendWithProgress(ctx, content, c, progress)`.
- `BaseClient.SendWithProgress` chama o callback com `"Connecting to Gemini..."` antes e `"Response received"` após sucesso.

#### Método `BuildRequest(content string) (interface{}, error)`

- Constrói um `GenerateRequest` Gemini com:
  - `Contents`: array com um único `Content` contendo um `Part` com `Text = content`.
  - `GenerationConfig`: apenas `MaxOutputTokens = c.MaxTokens`.
  - `SafetySettings`: omitido (nil).
- O `Role` do `Content` não é definido — a API Gemini usa `user` como default.

#### Método `ParseResponse(response, rawJSON) (*llm.Result, error)`

- Type-asserts para `*GenerateResponse`.
- Verifica `genResp.Error` e retorna erro formatado.
- Verifica `len(genResp.Candidates) > 0`.
- Concatena todos os `Part.Text` do primeiro candidato.
- Mapeia `UsageMetadata` → `llm.Usage`:
  - `PromptTokenCount` → `PromptTokens`
  - `CandidatesTokenCount` → `CompletionTokens`
  - `TotalTokenCount` → `TotalTokens`
- Retorna `llm.Result` completo.

#### Método `GetEndpoint() string`

- Retorna `/models/{model}:generateContent?key={apiKey}`.
- **Importante:** A autenticação Gemini é via query parameter `key=`, não header `Authorization`.

#### Método `GetHeaders() map[string]string`

- Retorna mapa vazio. Autenticação já está na URL.

#### Método `NewResponse() interface{}`

- Retorna `&GenerateResponse{}` para o `BaseClient` deserializar.

#### Método `GetProviderName() string`

- Retorna `c.ProviderName` (sempre `"Gemini"`).

#### Método `handleError(err error) error`

- Delega para `c.BaseClient.HandleHTTPError()` com `parseBody` retornando `""`.
- Resultado: mensagens de erro HTTP usam código de status, não body (o body é descartado).

---

### 3.2 `types.go` — Tipos do Request/Response Gemini (88 linhas)

Todas as structs são **POCO** (Plain Old Go Objects) com tags `json:` alinhadas à API Gemini v1beta.

| Tipo | Campo | JSON Tag | Tipo |
|------|-------|----------|------|
| `GenerateRequest` | Contents | `contents` | `[]Content` |
| `GenerateRequest` | GenerationConfig | `generationConfig` | `*GenerationConfig` |
| `GenerateRequest` | SafetySettings | `safetySettings` | `[]SafetySetting` |
| `Content` | Parts | `parts` | `[]Part` |
| `Content` | Role | `role` | `string` |
| `Part` | Text | `text` | `string` |
| `GenerationConfig` | Temperature | `temperature` | `float64` |
| `GenerationConfig` | TopK | `topK` | `int` |
| `GenerationConfig` | TopP | `topP` | `float64` |
| `GenerationConfig` | MaxOutputTokens | `maxOutputTokens` | `int` |
| `GenerationConfig` | StopSequences | `stopSequences` | `[]string` |
| `SafetySetting` | Category | `category` | `string` |
| `SafetySetting` | Threshold | `threshold` | `string` |
| `GenerateResponse` | Candidates | `candidates` | `[]Candidate` |
| `GenerateResponse` | UsageMetadata | `usageMetadata` | `*UsageMetadata` |
| `GenerateResponse` | Error | `error` | `*APIError` |
| `Candidate` | Content | `content` | `Content` |
| `Candidate` | FinishReason | `finishReason` | `string` |
| `Candidate` | Index | `index` | `int` |
| `UsageMetadata` | PromptTokenCount | `promptTokenCount` | `int` |
| `UsageMetadata` | CandidatesTokenCount | `candidatesTokenCount` | `int` |
| `UsageMetadata` | TotalTokenCount | `totalTokenCount` | `int` |
| `APIError` | Code | `code` | `int` |
| `APIError` | Message | `message` | `string` |
| `APIError` | Status | `status` | `string` |

---

### 3.3 `models.go` — Validação de Modelos (13 linhas)

- `ValidModels()`: retorna lista hardcoded de 5 modelos conhecidos.
- `IsKnownModel()`: **sempre retorna `true`**. O comentário indica que a validação foi removida para permitir modelos custom/preview. Marca como `Deprecated`.

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
| P1 | Duplicação de defaults | 🟡 Medium | `defaultMaxTokens=8192` em `client.go` e `DefaultConfigs()[ProviderGemini]` em `core/llm` são redundantes. `NewClient` não usa os defaults do `core/llm` — aplica os seus próprios via `llmbase.DefaultConfig`. |
| P2 | `HandleHTTPError` descarta body | 🟡 Medium | `handleError` passa `parseBody` retornando `""`, descartando o body de erro da API. O erro retornado é apenas `API error [<statusCode>]`. Perde informações diagnósticas. |
| P3 | `IsKnownModel` é um stub | 🟡 Medium | Sempre retorna `true`. A função `ValidModels()` ainda existe mas não é chamada por lugar algum no código. |
| P4 | `Content.Role` nunca definido | ℹ️ Low | `BuildRequest` não define `Role`. A Gemini API usa `"user"` como default, então funciona, mas é implícito. |
| P5 | Sem retry/rate-limit | 🔴 High | Identificado também no relatório do architect como debt `L3` (sem retry para LLM calls) e `L4` (sem rate limiting). |
| P6 | `SendWithProgress` não relata falhas | ℹ️ Low | O callback `progress` não é chamado no caminho de erro. |

---

## 6. Cobertura de Testes

### `client_test.go` — 7 testes

| Teste | O que verifica |
|-------|----------------|
| `TestClient_Send_Success` | HTTP 200 com candidatos + usage → `llm.Result` correto |
| `TestClient_Send_MultipleParts` | Concatenação de múltiplos Parts |
| `TestClient_Send_APIError` | `GenerateResponse.Error` → error message contendo código |
| `TestClient_Send_NoCandidates` | Array vazio → error "no candidates" |
| `TestClient_NewClient_Validation` | APIKey vazia → error "api key is required" |
| `TestClient_IsConfigured` | `IsConfigured()` retorna true/false conforme APIKey |
| `TestClient_SendWithProgress` | Callback recebe "Connecting to Gemini..." e "Response received" |

### `models_test.go` — 2 testes (7 sub-cases)

| Teste | O que verifica |
|-------|----------------|
| `TestValidModels` | Lista tem 5 modelos esperados |
| `TestIsKnownModel` | 7 cenários — todos retornam `true` |

### Cobertura estimada: **~85% das linhas de código**. Pontos não cobertos:
- Caminho de erro `handleError` (requer HTTP error real)
- `BuildRequest` com SafetySettings (não testado explicitamente)

---

## 7. Resumo Executável

O módulo `geminiapi` é um adaptador bem estruturado que segue o padrão Strategy em torno de `BaseClient`. Implementa todos os 6 métodos do `llmbase.Sender` interface, mapeia corretamente request/response Gemini, e integra-se com o sistema de progresso via callback. As principais lacunas são: ausência de retry, de rate limiting, e de tratamento rico de erros HTTP.
