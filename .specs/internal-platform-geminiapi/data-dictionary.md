# Dicionário de Dados — internal/platform/geminiapi

> **Módulo:** `github.com/quantmind-br/shotgun-cli/internal/platform/geminiapi`
> **Nível de detalhe:** detalhado
> **Gerado pela fase archaeologist do Reversa

---

## 1. Tipos Definidos no Pacote

### 1.1 `GenerateRequest`

Estrutura do corpo da requisição enviada à API Gemini `generateContent`.

```go
type GenerateRequest struct {
    Contents         []Content         `json:"contents"`
    GenerationConfig *GenerationConfig `json:"generationConfig,omitempty"`
    SafetySettings   []SafetySetting   `json:"safetySettings,omitempty"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Contents | `[]Content` | Não | Array de mensagens/conteúdo. Sempre contém pelo menos 1 elemento. |
| GenerationConfig | `*GenerationConfig` | Sim (`omitempty`) | Configurações de geração (max tokens, temperature, etc). |
| SafetySettings | `[]SafetySetting` | Sim (`omitempty`) | Configurações de segurança (nunca usado neste adaptador). |

**Fluxo de criação:** `Client.BuildRequest(content)` → cria `GenerateRequest` com 1 `Content` contendo 1 `Part` com o texto do prompt.

---

### 1.2 `Content`

Representa uma mensagem dentro de `Contents`.

```go
type Content struct {
    Parts []Part `json:"parts"`
    Role  string `json:"role,omitempty"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Parts | `[]Part` | Não | Lista de partes de texto. Sempre 1+ partes. |
| Role | `string` | Sim | Papel (`"user"`, `"model"`). Não definido no adaptador (default da API = `"user"`). |

**Nota:** 🟡 **INFERIDO** — `Role` é sempre omitido pelo adaptador. A API Gemini assume `"user"` implicitamente.

---

### 1.3 `Part`

Unidade mínima de conteúdo.

```go
type Part struct {
    Text string `json:"text"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Text | `string` | Não | Texto do prompt ou da resposta. |

---

### 1.4 `GenerationConfig`

Parâmetros de geração do LLM.

```go
type GenerationConfig struct {
    Temperature     float64  `json:"temperature,omitempty"`
    TopK            int      `json:"topK,omitempty"`
    TopP            float64  `json:"topP,omitempty"`
    MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
    StopSequences   []string `json:"stopSequences,omitempty"`
}
```

| Campo | Tipo | Opcional | Default | Descrição |
|-------|------|----------|---------|-----------|
| Temperature | `float64` | Sim | — | Creatividade (0.0–1.0). Não configurado pelo adaptador. |
| TopK | `int` | Sim | — | Amostrar entre os K tokens mais prováveis. Não configurado. |
| TopP | `float64` | Sim | — | Sampling de probabilidade acumulada. Não configurado. |
| MaxOutputTokens | `int` | Sim | `c.MaxTokens` (padrão 8192) | Máximo de tokens na saída. **Único campo configurado.** |
| StopSequences | `[]string` | Sim | — | Sequências que param a geração. Não configurado. |

---

### 1.5 `SafetySetting`

Configuração de filtros de segurança.

```go
type SafetySetting struct {
    Category  string `json:"category"`
    Threshold string `json:"threshold"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Category | `string` | Não | Categoria de segurança (e.g., `"HARM_CATEGORY_HARASSMENT"`). |
| Threshold | `string` | Não | Limiar (e.g., `"BLOCK_ONLY_HIGH"`). |

**Nota:** 🟡 **INFERIDO** — Esta struct é definida mas `SafetySettings` nunca é populado em `BuildRequest`. O campo é `nil` em todas as requisições.

---

### 1.6 `GenerateResponse`

Resposta da API Gemini.

```go
type GenerateResponse struct {
    Candidates    []Candidate    `json:"candidates"`
    UsageMetadata *UsageMetadata `json:"usageMetadata,omitempty"`
    Error         *APIError      `json:"error,omitempty"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Candidates | `[]Candidate` | Não | Lista de respostas candidatas. `len >= 1` indica sucesso. |
| UsageMetadata | `*UsageMetadata` | Sim | Métricas de uso de tokens. Pode ser nil. |
| Error | `*APIError` | Sim | Erro da API. Presente se falha. |

---

### 1.7 `Candidate`

Uma resposta candidata.

```go
type Candidate struct {
    Content      Content `json:"content"`
    FinishReason string  `json:"finishReason"`
    Index        int     `json:"index"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Content | `Content` | Não | Conteúdo da resposta. |
| FinishReason | `string` | Não | Motivo de parada: `"STOP"`, `"SAFETY"`, etc. |
| Index | `int` | Não | Índice no array de candidatos. |

---

### 1.8 `UsageMetadata`

Métricas de consumo de tokens.

```go
type UsageMetadata struct {
    PromptTokenCount     int `json:"promptTokenCount"`
    CandidatesTokenCount int `json:"candidatesTokenCount"`
    TotalTokenCount      int `json:"totalTokenCount"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| PromptTokenCount | `int` | Não | Tokens no prompt de entrada. |
| CandidatesTokenCount | `int` | Não | Tokens na resposta (candidatos). |
| TotalTokenCount | `int` | Não | Total (prompt + candidatos). |

**Mapeamento em `ParseResponse`:**
- `PromptTokenCount` → `llm.Usage.PromptTokens`
- `CandidatesTokenCount` → `llm.Usage.CompletionTokens`
- `TotalTokenCount` → `llm.Usage.TotalTokens`

---

### 1.9 `APIError`

Erro retornado pela API Gemini.

```go
type APIError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Status  string `json:"status"`
}
```

| Campo | Tipo | Opcional | Descrição |
|-------|------|----------|-----------|
| Code | `int` | Não | Código de status HTTP da API. |
| Message | `string` | Não | Mensagem legível do erro. |
| Status | `string` | Não | Status de erro (e.g., `"INVALID_ARGUMENT"`). |

**Nota:** 🟡 **INFERIDO** — `Code` e `Status` são usados para formatar a mensagem de erro, mas `Message` é descartado por `handleError` (body parse retorna `""`).

---

## 2. Tipos Externos Consumidos

### 2.1 `llm.Config` (de `internal/core/llm`)

```go
type Config struct {
    Provider  ProviderType
    APIKey    string
    BaseURL   string
    Model     string
    Timeout   int          // segundos
    MaxTokens int
    Temperature float64
}
```

Usado em `NewClient(cfg llm.Config)` para configurar o `BaseClient`.

### 2.2 `llm.Result` (de `internal/core/llm`)

```go
type Result struct {
    Response    string
    RawResponse string
    Model       string
    Provider    string
    Duration    time.Duration
    Usage       *Usage
}
```

Tipo retornado por `Send()` e `SendWithProgress()`. Preenchido em `ParseResponse()`.

### 2.3 `llm.Usage` (de `internal/core/llm`)

```go
type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
}
```

Preenchido a partir de `UsageMetadata` em `ParseResponse()`.

### 2.4 `llmbase.BaseClient` (de `internal/platform/llmbase`)

```go
type BaseClient struct {
    JSONClient   *platformhttp.JSONClient
    APIKey       string
    Model        string
    MaxTokens    int
    ProviderName string
}
```

Composto em `Client`. Fornece `Send()`, `SendWithProgress()`, `HandleHTTPError()`.

### 2.5 `llmbase.DefaultConfig` (de `internal/platform/llmbase`)

```go
type DefaultConfig struct {
    BaseURL   string
    Model     string
    MaxTokens int
    Timeout   time.Duration
}
```

Usado em `NewBaseClientWithDefaults()` para sobrepor defaults provider-specific.

---

## 3. Valores Constantes do Pacote

| Constante | Valor | Localização |
|-----------|-------|-------------|
| `defaultBaseURL` | `"https://generativelanguage.googleapis.com/v1beta"` | `client.go` |
| `defaultMaxTokens` | `8192` | `client.go` |

### Valores Hardcoded em `NewClient`

| Parâmetro | Valor |
|-----------|-------|
| `Model` | `"gemini-2.5-flash"` |
| `MaxTokens` | `8192` (via `defaultMaxTokens`) |
| `Timeout` | `300 * time.Second` |
| `ProviderName` | `"Gemini"` |

---

## 4. Tipos de Dados do Pacote

| Tipo | Tipo Go | Usado em |
|------|---------|----------|
| `ProviderType` (externo) | `string` (const `ProviderGemini = "gemini"`) | Registro em `core/llm` |

---

## 5. Fluxo de Dados End-to-End

```
[Usuario: content string]
        │
        ▼
  Client.Send(ctx, content)
        │
        ▼
  BaseClient.Send(ctx, content, sender=Client)
        │
        ├── sender.BuildRequest(content) → GenerateRequest
        │
        ├── BaseClient.PostJSON(endpoint, headers, reqBody, response)
        │       │
        │       ▼
        │  [HTTP POST → Gemini API]
        │       │
        │       ▼
        │  GenerateResponse (unmarshaled)
        │
        ├── sender.ParseResponse(response, rawJSON) → llm.Result
        │       │
        │       └── Mapeia: UsageMetadata → llm.Usage
        │              Candidates → responseText
        │
        └── [Result com Duration calculada]
```

---

## 6. Notas

- Todos os tipos em `types.go` seguem a convenção JSON da Gemini API v1beta.
- A conversão para `llm.Result` é feita exclusivamente em `ParseResponse()`.
- 🟡 **INFERIDO** — O campo `SafetySettings` da Gemini API não é utilizado pelo adaptador. Se a feature for adicionada no futuro, a struct `SafetySetting` está pronta para uso.
