# Fluxo de Ciclo de Vida do Cliente — `NewClient()` e Configuração

> **Nível de detalhamento:** detalhado  
> **Idioma do documento:** pt-br  
> **Data:** 2026-05-20

---

## 1. Fluxo `NewClient(cfg llm.Config)`

```mermaid
flowchart TD
    A[NewClient cfg] --> B[llmbase.NewBaseClientWithDefaults cfg defaults Anthropic]
    
    B --> B1{cfg.APIKey == ""?}
    B1 -->|Yes| B2[return nil error api key is required]
    B1 -->|No| B3[baseURL = cfg.BaseURL if non-empty else defaults.BaseURL]
    
    B3 --> B4[timeout = cfg.Timeout * 1s]
    B4 --> B5{timeout == 0?}
    B5 -->|Yes| B6[timeout = defaults.Timeout]
    B5 -->|No| B7[timeout stays]
    B6 --> B8{timeout == 0?}
    B8 -->|Yes| B9[timeout = 300s]
    B8 -->|No| B7
    B9 --> B7
    
    B7 --> B10[model = cfg.Model if non-empty else defaults.Model]
    B10 --> B11[maxTokens = cfg.MaxTokens if non-zero else defaults.MaxTokens]
    
    B11 --> B12[JSONClient = platformhttp.NewJSONClient baseURL timeout]
    
    B12 --> B13[BaseClient = struct JSONClient APIKey cfg.APIKey Model model MaxTokens maxTokens ProviderName Anthropic]
    
    B13 --> B14[Client = Client BaseClient]
    B14 --> B15[return Client nil]
    
    classDef success fill:#d4edda,color:#155724
    classDef error fill:#f8d7da,color:#721c24
    classDef default fill:#fff3cd,color:#856404
    class B2 error
    class B15 success
    class B3,B6,B8,B9,B11 default
```

### Passos de Default

| Config Field | Valor fornecido | Valor default (Anthropic) | Resultado |
|-------------|-----------------|--------------------------|-----------|
| `BaseURL` | `""` (vazio) | `"https://api.anthropic.com"` | Usa default |
| `BaseURL` | `"https://custom.api"` | `"https://api.anthropic.com"` | Usa fornecido |
| `Model` | `""` | `"claude-sonnet-4-20250514"` | Usa default |
| `Model` | `"claude-3-opus"` | `"claude-sonnet-4-20250514"` | Usa fornecido |
| `Timeout` | `0` | `300` segundos | 300s |
| `Timeout` | `60` | `300` segundos | 60s |
| `MaxTokens` | `0` | `8192` | Usa default |
| `MaxTokens` | `4096` | `8192` | Usa 4096 |
| `APIKey` | `""` | — | **ERRO** |

---

## 2. Fluxo de Configuração e Validação

```mermaid
flowchart TD
    A[Client created] --> B[IsConfigured]
    
    B --> B1[c.APIKey != "" AND c.Model != ""]
    B1 -->|Yes| B2[return true]
    B1 -->|No| B3[return false]
    
    B --> C[ValidateConfig]
    C --> C1[c.APIKey == ""? ]
    C1 -->|Yes| C2[return error API key is required]
    C1 -->|No| C3[c.Model == ""? ]
    C3 -->|Yes| C4[return error model is required]
    C3 -->|No| C5[return nil]
    
    B --> D[IsAvailable]
    D --> D1[return true]
    
    B --> E[Name]
    E --> E1[return c.ProviderName]
    
    classDef success fill:#d4edda,color:#155724
    classDef fail fill:#f8d7da,color:#721c24
    class B2,E1 success
    class C2,C4 fail
```

---

## 3. Fluxo de Interação com `llmbase.BaseClient.Send()`

```mermaid
flowchart TD
    A[BaseClient.Send ctx content sender] --> A1[startTime = time.Now]
    
    A1 --> A2[reqBody sender.BuildRequest content]
    A2 --> A3[headers sender.GetHeaders]
    A3 --> A4[endpoint sender.GetEndpoint]
    A4 --> A5[response sender.NewResponse]
    
    A5 --> A6[c.JSONClient.PostJSON endpoint headers reqBody response]
    
    A6 --> A7{success?}
    A7 -->|No| A8[return nil err]
    A7 -->|Yes| A9[rawJSON json.Marshal response]
    
    A9 --> A10[sender.ParseResponse response rawJSON]
    A10 --> A11{success?}
    A11 -->|No| A8
    A11 -->|Yes| A12[result.Duration time.Since startTime]
    
    A12 --> A13[return result nil]
    
    classDef success fill:#d4edda,color:#155724
    classDef error fill:#f8d7da,color:#721c24
    class A13 success
    class A8 error
```

---

## 4. Fluxo de Integração com `llm.Registry` (como provider)

```mermaid
flowchart TD
    A[main: registry.Register anthropic func cfg] --> A1[func cfg -> Client, error]
    A1 --> A2[NewClient cfg]
    A2 --> A3[NewBaseClientWithDefaults cfg defaults Anthropic]
    A3 --> A4[returns Client]
    
    A4 --> B[main: registry.Create cfg with Provider=anthropic]
    B --> B1[lookup anthropic in creators map]
    B1 --> B2{found?}
    B2 -->|No| B3[return error unsupported provider]
    B2 -->|Yes| B4[call creator cfg]
    B4 --> B5[NewClient cfg]
    B5 --> B6[returns Provider]
    B6 --> B7[main uses Provider.Send]
    
    classDef success fill:#d4edda,color:#155724
    classDef error fill:#f8d7da,color:#721c24
    class B7 success
    class B3 error
```

---

## 5. Fluxo de Uso Completo (Tentativa de Envio)

```mermaid
flowchart TD
    A[App creates llm.Registry] --> A1[Register ProviderAnthropic with func cfg]
    A1 --> A2[func calls NewClient cfg]
    A2 --> A3[Client created with API key, model, baseURL]
    
    A3 --> B[App calls registry.Create with Config Provider=anthropic]
    B --> B1[registry looks up anthropic]
    B1 --> B2[call creator func Config]
    B2 --> B3[returns *Client as Provider]
    
    B3 --> C[App calls provider.Send ctx prompt]
    C --> C1[Client.Send calls BaseClient.Send]
    C1 --> C2[BuildRequest: MessagesRequest user prompt]
    C2 --> C3[PostJSON: POST /v1/messages with headers]
    C3 --> C4{API OK?}
    
    C4 -->|No| C5[handleError parse error body]
    C5 --> C6[return error to caller]
    
    C4 -->|Yes| C7[ParseResponse extract text + usage]
    C7 --> C8[return Result to caller]
    
    classDef success fill:#d4edda,color:#155724
    classDef error fill:#f8d7da,color:#721c24
    class C8 success
    class C6 error
```

---

## 6. Fluxo de Validação de Modelos (`models.go`)

```mermaid
flowchart LR
    A[ValidModels call] --> B[return fixed array of 8 strings]
    
    A2[IsKnownModel model baseURL] --> C[return true]
    
    classDef return fill:#d4edda,color:#155724
    class B,C return
```

**Observação:** `IsKnownModel` foi modificado propositalmente para sempre retornar `true`. A validação de modelo foi removida para aceitar modelos customizados e preview. A função está marcada como `@deprecated`.

---

## 7. Tabela de Referência de Fluxos

| Fluxo | Arquivo Principal | Função Entry | Complexidade |
|-------|-------------------|-------------|-------------|
| **Criação do cliente** | `client.go` | `NewClient()` | Média — aplica defaults, valida API key |
| **Envio simples** | `client.go` + `base_client.go` | `Send()` | Alta — build → HTTP → parse |
| **Envio com progresso** | `client.go` + `base_client.go` | `SendWithProgress()` | Alta — idêntico a Send + 2 callbacks |
| **Tratamento de erro** | `client.go` + `base_client.go` | `handleError()` | Média — tenta parse JSON de erro, fallback |
| **Parsing de resposta** | `client.go` | `ParseResponse()` | Média — iterar content blocks, mapear usage |
| **Construção de requisição** | `client.go` | `BuildRequest()` | Baixa — cria struct simples |
| **Headers e endpoint** | `client.go` | `GetHeaders()` / `GetEndpoint()` | Baixa — retorne maps/strings |
| **Validação de modelo** | `models.go` | `ValidModels()` / `IsKnownModel()` | Baixa — retorna array / true |
