# Dicionário de Dados — Módulo `internal/platform/anthropic`

> **Nível de detalhamento:** detalhado  
> **Idioma do documento:** pt-br  
> **Data:** 2026-05-20

---

## 1. Visão Geral

Este dicionário documenta **todos** os tipos, campos, constantes e variáveis de estado expostos pelo módulo `internal/platform/anthropic`. Para cada item são fornecidos: nome, tipo, origem (arquivo/linha), descrição, valores possíveis e restrições.

---

## 2. Constantes

### `defaultBaseURL`

| Campo | Valor |
|-------|-------|
| **Arquivo** | `client.go` (linha 11) |
| **Tipo** | `string` |
| **Valor** | `"https://api.anthropic.com"` |
| **Escopo** | Package (lowercase = privada) |
| **Uso** | Valor padrão do `BaseURL` quando `cfg.BaseURL` está vazio |

### `anthropicVersion`

| Campo | Valor |
|-------|-------|
| **Arquivo** | `client.go` (linha 12) |
| **Tipo** | `string` |
| **Valor** | `"2023-06-01"` |
| **Escopo** | Package (lowercase = privada) |
| **Uso** | Enviado como header `anthropic-version` em todas as requisições |

### `defaultMaxTokens`

| Campo | Valor |
|-------|-------|
| **Arquivo** | `client.go` (linha 13) |
| **Tipo** | `int` |
| **Valor** | `8192` |
| **Escopo** | Package (lowercase = privada) |
| **Uso** | Valor padrão de `MaxTokens` quando `cfg.MaxTokens` está vazio |

---

## 3. Tipos do Módulo (definidos localmente)

### 3.1 `MessagesRequest`

**Arquivo:** `types.go` (linha 6)  
**Descrição:** Corpo da requisição enviado à Anthropic Messages API (`POST /v1/messages`).

| Campo | Tipo | Tag JSON | Nulo/Opcional | Valores Válidos | Obrigatório? |
|-------|------|----------|---------------|-----------------|--------------|
| `Model` | `string` | `"model"` | Não | Qualquer string válida de modelo Anthropic | Sim |
| `MaxTokens` | `int` | `"max_tokens"` | Não | Inteiro ≥ 1 | Sim |
| `Messages` | `[]Message` | `"messages"` | Não | Array de 1..N mensagens | Sim |
| `System` | `string` | `"system,omitempty"` | Sim (omissão) | Instruções do sistema | Não |
| `Stream` | `bool` | `"stream,omitempty"` | Sim (omissão) | `false` (sempre) | Não |

**Regras de validação:** Nenhum — a validação é feita pelo `BuildRequest()` (formato mínimo) e pela API.

**Exemplo de JSON gerado:**
```json
{
  "model": "claude-sonnet-4-20250514",
  "max_tokens": 8192,
  "messages": [
    {"role": "user", "content": "Hello!"}
  ]
}
```

### 3.2 `Message`

**Arquivo:** `types.go` (linha 13)  
**Descrição:** Uma mensagem individual dentro de uma conversa.

| Campo | Tipo | Tag JSON | Valores Válidos | Obrigatório? |
|-------|------|----------|-----------------|--------------|
| `Role` | `string` | `"role"` | `"user"`, `"assistant"` | Sim |
| `Content` | `string` | `"content"` | Texto livre (UTF-8) | Sim |

**Regras:** `Role` deve ser `"user"` ou `"assistant"`. Na implementação atual, todas as mensagens são criadas com `Role: "user"`.

### 3.3 `MessagesResponse`

**Arquivo:** `types.go` (linha 18)  
**Descrição:** Resposta deserializada da Anthropic Messages API.

| Campo | Tipo | Tag JSON | Usado pelo módulo? | Observação |
|-------|------|----------|---------------------|------------|
| `ID` | `string` | `"id"` | Não | ID da mensagem |
| `Type` | `string` | `"type"` | Não | Sempre `"message"` |
| `Role` | `string` | `"role"` | Não | Sempre `"assistant"` |
| `Content` | `[]ContentBlock` | `"content"` | **Sim** | Blocos de texto concatenados |
| `Model` | `string` | `"model"` | Não | Modelo que gerou a resposta |
| `StopReason` | `string` | `"stop_reason"` | Não | `"end_turn"`, `"max_tokens"`, `"stop_sequence"` |
| `StopSequence` | `string` | `"stop_sequence,omitempty"` | Não | Sequência que causou parada |
| `Usage` | `UsageAPI` | `"usage"` | **Sim** | Métricas de tokens |

### 3.4 `ContentBlock`

**Arquivo:** `types.go` (linha 31)  
**Descrição:** Um bloco de conteúdo dentro da resposta.

| Campo | Tipo | Tag JSON | Valores Válidos | Usado? |
|-------|------|----------|-----------------|--------|
| `Type` | `string` | `"type"` | `"text"`, `"tool_use"`, `"thinking"`, ... | **Sim** — filtrado por tipo `"text"` |
| `Text` | `string` | `"text"` | Texto livre | **Sim** — lido quando `Type == "text"` |

**Comportamento no módulo:** `ParseResponse()` itera sobre o array e concatena `Text` de todos os blocos onde `Type == "text"`. Blocos de outros tipos são ignorados.

### 3.5 `UsageAPI`

**Arquivo:** `types.go` (linha 37)  
**Descrição:** Métricas de uso de tokens da Anthropic.

| Campo | Tipo | Tag JSON | Faça Válida | Usado? |
|-------|------|----------|-------------|--------|
| `InputTokens` | `int` | `"input_tokens"` | ≥ 0 | **Sim** → `PromptTokens` |
| `OutputTokens` | `int` | `"output_tokens"` | ≥ 0 | **Sim** → `CompletionTokens` |

**Mapeamento para `llm.Usage`:**
```
Usage.PromptTokens     = UsageAPI.InputTokens
Usage.CompletionTokens = UsageAPI.OutputTokens
Usage.TotalTokens      = InputTokens + OutputTokens
```

### 3.6 `ErrorResponse`

**Arquivo:** `types.go` (linha 42)  
**Descrição:** Estrutura para parsing de respostas de erro da API.

| Campo | Tipo | Tag JSON | Usado? |
|-------|------|----------|--------|
| `Type` | `string` | `"type"` | Não |
| `Error.Type` | `string` | `error.type` | Não (mas lido) |
| `Error.Message` | `string` | `error.message` | **Sim** — extraído no `handleError()` |

**Exemplo de JSON:**
```json
{
  "type": "error",
  "error": {
    "type": "authentication_error",
    "message": "Invalid API key"
  }
}
```

---

## 4. Tipos do Módulo (definidos em `client.go`)

### 4.1 `Client`

**Arquivo:** `client.go` (linha 16)

| Campo | Tipo | Origem | Descrição |
|-------|------|--------|-----------|
| `BaseClient` | `*llmbase.BaseClient` (embedded) | `llmbase` | Cliente base com HTTP, configuração e métodos compartilhados |

**Campos herdados de `BaseClient`:**

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `JSONClient` | `*platformhttp.JSONClient` | Cliente HTTP com marshal/unmarshal JSON |
| `APIKey` | `string` | Chave de API do Anthropic |
| `Model` | `string` | Modelo Claude ativo |
| `MaxTokens` | `int` | Limite máximo de tokens |
| `ProviderName` | `string` | Nome do provedor (`"Anthropic"`) |

---

## 5. Tipos Externos Consumidos

### 5.1 `llm.Config` (pacote `internal/core/llm`)

| Campo | Tipo | Obrigatório? | Usado por |
|-------|------|--------------|-----------|
| `Provider` | `llm.ProviderType` | Sim | Validação (via `llmbase`) |
| `APIKey` | `string` | Sim | `NewBaseClientWithDefaults` |
| `BaseURL` | `string` | Não (default aplicado) | `NewBaseClientWithDefaults` |
| `Model` | `string` | Sim (default aplicado) | `NewBaseClientWithDefaults` |
| `Timeout` | `int` | Sim (segundos) | `NewBaseClientWithDefaults` |
| `MaxTokens` | `int` | Não (default aplicado) | `NewBaseClientWithDefaults` |
| `Temperature` | `float64` | Não | Não utilizado pelo Anthropic |

### 5.2 `llm.Result` (pacote `internal/core/llm`)

| Campo | Tipo | Preenchido por `ParseResponse`? |
|-------|------|---------------------------------|
| `Response` | `string` | **Sim** — texto concatenado |
| `RawResponse` | `string` | **Sim** — JSON bruto da API |
| `Model` | `string` | **Sim** — `c.Model` |
| `Provider` | `string` | **Sim** — `c.ProviderName` |
| `Duration` | `time.Duration` | **Sim** — calculado pelo `BaseClient.Send()` |
| `Usage` | `*llm.Usage` | **Condicional** — não-nulo se tokens > 0 |

### 5.3 `llm.Usage` (pacote `internal/core/llm`)

| Campo | Tipo | Valor |
|-------|------|-------|
| `PromptTokens` | `int` | `msgResp.Usage.InputTokens` |
| `CompletionTokens` | `int` | `msgResp.Usage.OutputTokens` |
| `TotalTokens` | `int` | `InputTokens + OutputTokens` |

### 5.4 `llmbase.Sender` (interface, pacote `internal/platform/llmbase`)

| Método | Assinatura | Implementado por |
|--------|-----------|------------------|
| `BuildRequest` | `func(content string) (interface{}, error)` | `Client` |
| `ParseResponse` | `func(response, rawJSON) (*llm.Result, error)` | `Client` |
| `GetEndpoint` | `func() string` | `Client` |
| `GetHeaders` | `func() map[string]string` | `Client` |
| `NewResponse` | `func() interface{}` | `Client` |
| `GetProviderName` | `func() string` | `Client` |

### 5.5 `llm.Provider` (interface, pacote `internal/core/llm`)

O `Client` satisfaz esta interface **indiretamente** através dos métodos herdados de `BaseClient` (`Name()`, `IsAvailable()`, `IsConfigured()`, `ValidateConfig()`) e dos métodos propios (`Send()`, `SendWithProgress()`).

**Métodos da interface:**

| Método | Assinatura | Origem |
|--------|-----------|--------|
| `Send` | `func(ctx, content) (*llm.Result, error)` | `Client` |
| `SendWithProgress` | `func(ctx, content, progress) (*llm.Result, error)` | `Client` |
| `Name` | `func() string` | `BaseClient.Name()` |
| `IsAvailable` | `func() bool` | `BaseClient.IsAvailable()` |
| `IsConfigured` | `func() bool` | `BaseClient.IsConfigured()` |
| `ValidateConfig` | `func() error` | `BaseClient.ValidateConfig()` |

---

## 6. Dicionário de Valores de Constantes

### Modelos da Anthropic (lista fixa)

**Origem:** `models.go` → `ValidModels()`

| # | String | Família |
|---|--------|---------|
| 1 | `claude-sonnet-4-20250514` | Sonnet 4 |
| 2 | `claude-3-5-sonnet-latest` | Sonnet 3.5 |
| 3 | `claude-3-5-sonnet-20241022` | Sonnet 3.5 |
| 4 | `claude-3-5-haiku-latest` | Haiku 3.5 |
| 5 | `claude-3-opus-latest` | Opus 3 |
| 6 | `claude-3-opus-20240229` | Opus 3 |
| 7 | `claude-3-sonnet-20240229` | Sonnet 3 |
| 8 | `claude-3-haiku-20240307` | Haiku 3 |

**Nota:** Todos os valores são strings não-vazias. O formato segue o padrão `{familia}-{versão}-{data}`.

---

## 7. Glossário de Domínio

| Termo | Definição |
|-------|-----------|
| **Provider** | Implementação concreta de um provedor de IA (Anthropic, OpenAI, Gemini). O `Client` é o provider Anthropic. |
| **Sender** | Estratégia que define como construir requisições e interpretar respostas para um provider específico. |
| **BaseClient** | Implementação base que lida com HTTP, timeout, e fluxo genérico de Send. |
| **ContentBlock** | Unidade de conteúdo na resposta da Anthropic (texto, ferramentas, etc.). |
| **Usage** | Métricas de consumo de tokens (input, output, total). |
| **anthropic-version** | Header HTTP que indica a versão da API. |
| **end_turn** | Razão de parada indicando que o modelo completou naturalmente sua resposta. |
