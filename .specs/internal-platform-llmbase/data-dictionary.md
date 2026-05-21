# Dicionário de Dados — internal/platform/llmbase

| Item              | Detalhe                                                        |
|-------------------|----------------------------------------------------------------|
| **Módulo**        | `internal/platform/llmbase`                                    |
| **Package path**  | `github.com/quantmind-br/shotgun-cli/internal/platform/llmbase` |
| **Document Language** | Português (pt-br)                                           |

---

## 1. Interfaces

### 1.1 `Sender` (interface)

Interface Strategy que define as operações provider-specific. É implementada por cada provedor concreto (`openai.Client`, `anthropic.Client`, `geminiapi.Client`).

| Método | Retorno | Obrigatório | Descrição |
|--------|---------|-------------|-----------|
| `BuildRequest(content string)` | `(interface{}, error)` | Sim | Cria o payload JSON específico do provider. O resultado deve ser serializável por `encoding/json`. |
| `ParseResponse(response interface{}, rawJSON []byte)` | `(*llm.Result, error)` | Sim | Extrai `llm.Result` (da package `internal/core/llm`) a partir da resposta deserializada e do JSON bruto. |
| `GetEndpoint()` | `string` | Sim | Path da API endpoint. Para Gemini inclui a API key na query string. Ex: `"/v1/chat/completions"`. |
| `GetHeaders()` | `map[string]string` | Sim | Headers HTTP provider-specific. Content-Type é setado por `BaseClient`. Ex: `{"Authorization": "Bearer ..."}`. |
| `NewResponse()` | `interface{}` | Sim | Retorna uma instância vazia (zero value) do struct de response que será passado para `json.Unmarshal`. |
| `GetProviderName()` | `string` | Sim | Nome de exibição do provider (ex: "OpenAI", "Anthropic", "Gemini"). |

**Responsabilidade**: Permitir que `BaseClient` execute o fluxo genérico de envio/recebimento enquanto cada provider controla os detalhes da API.

---

## 2. Structs

### 2.1 `BaseClient`

Estrutura principal que encapsula toda a lógica comum de comunicação HTTP com LLMs.

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `JSONClient` | `*platformhttp.JSONClient` | Sim | Cliente HTTP reutilizável. Gerencia timeouts, marshaling, headers comuns, e tratamento de erros HTTP. |
| `APIKey` | `string` | Sim | Chave de API do provedor. Usada em headers (OpenAI, Anthropic) ou URL (Gemini). |
| `Model` | `string` | Sim | Nome do modelo LLM (ex: "gpt-4o", "claude-sonnet-4-20250514"). |
| `MaxTokens` | `int` | Não | Limite máximo de tokens na resposta. 0 significa sem limite. |
| `ProviderName` | `string` | Sim | Nome legível do provedor para exibição e logging. |

**Métodos embutidos via embedding** (quando concretos embed `*BaseClient`):
- `Name()` → retorna `ProviderName`
- `IsAvailable()` → retorna `true`
- `IsConfigured()` → `APIKey != "" && Model != ""`
- `ValidateConfig()` → valida `APIKey` e `Model` não vazios
- `Send(ctx, content, sender)` → fluxo principal de envio
- `SendWithProgress(ctx, content, sender, progress)` → envio com callbacks
- `HandleHTTPError(err, parseBody)` → tratamento de erros HTTP

---

### 2.2 `Config`

Configuração minimalista para criação direta de `BaseClient` via `NewBaseClient()`.

| Campo | Tipo | Obrigatório | Padrão | Descrição |
|-------|------|-------------|--------|-----------|
| `APIKey` | `string` | Sim | — | Chave de API |
| `BaseURL` | `string` | Sim | — | URL base da API (ex: `"https://api.openai.com/v1"`) |
| `Model` | `string` | Sim | — | Nome do modelo |
| `MaxTokens` | `int` | Não | `0` | Máximo de tokens na resposta |
| `Timeout` | `time.Duration` | Não | `300s` | Timeout para requisições HTTP |

**Nota**: Esta struct é usada exclusivamente por `NewBaseClient()`. Os provedores concretos usam `llm.Config` + `DefaultConfig` via `NewBaseClientWithDefaults()`.

---

### 2.3 `DefaultConfig`

Valores padrão provider-specific aplicados por `NewBaseClientWithDefaults()`.

| Campo | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `BaseURL` | `string` | Sim | URL base do provider (ex: `"https://api.openai.com/v1"`) |
| `Model` | `string` | Sim | Modelo padrão do provider (ex: `"gpt-4o"`) |
| `MaxTokens` | `int` | Não | Tokens máximos padrão. 0 = sem limite. |
| `Timeout` | `time.Duration` | Não | Timeout padrão. 0 = 300s. |

**Uso**: Cada provedor concrete passa sua própria instância:

| Provider | BaseURL | Model | MaxTokens | Timeout |
|----------|---------|-------|-----------|---------|
| OpenAI | `"https://api.openai.com/v1"` | `"gpt-4o"` | 0 (sem limite) | `300s` |
| Anthropic | `"https://api.anthropic.com"` | `"claude-sonnet-4-20250514"` | `8192` | `300s` |
| Gemini | `"https://generativelanguage.googleapis.com/v1beta"` | `"gemini-2.5-flash"` | `8192` | `300s` |

---

## 3. Relações entre Tipos

```
llm.Config  ───[via]──→  DefaultConfig  ───[constrói]──→  BaseClient
                                                                    │
llm.Provider  ◄──[implementado por]──  Client{*BaseClient}  ◄───[embed]──  BaseClient
                                                                    │
                                                    BaseClient.Send()  ───[delega para]──→  Sender (interface)
```

**Fluxo de criação típico** (ex: OpenAI):

1. `llm.Config` é preenchido pelo usuário (via CLI ou config file)
2. `NewBaseClientWithDefaults(cfg, DefaultConfig{...}, "OpenAI")` é chamado
3. Internamente: `llm.Config` é mapeado para `llmbase.Config` → `NewBaseClient` cria `BaseClient`
4. `BaseClient` é embedido no `Client` struct do provider
5. `Client` implementa `Sender` com lógica OpenAI-specific

---

## 4. Resumo de Todos os Tipos

| Tipo | Categoria | Arquivo | Campos/Elementos |
|------|-----------|---------|-----------------|
| `Sender` | Interface | `sender.go` | 6 métodos |
| `BaseClient` | Struct | `base_client.go` | 5 campos |
| `Config` | Struct | `base_client.go` | 5 campos |
| `DefaultConfig` | Struct | `base_client.go` | 4 campos |

---

## 5. Dependências Externas aos Tipos

| Tipo | Importa de | Finalidade |
|------|-----------|------------|
| `BaseClient.Send()` | `internal/core/llm` | `llm.Result` como tipo de retorno |
| `BaseClient` | `internal/platform/http` | `*platformhttp.JSONClient` para HTTP |
| `Sender.ParseResponse()` | `internal/core/llm` | `*llm.Result` como tipo de retorno |
| `BaseClient` | `time` | `time.Duration` para Timeout |
| `BaseClient` | `context` | `context.Context` para cancelamento |
