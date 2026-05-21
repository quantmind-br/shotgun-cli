# Fluxo: Construção da Requisição (`BuildRequest`)

> **Arquivo:** `client.go` → `Client.BuildRequest()`
> **Tipo:** Método do Strategy `llmbase.Sender`

---

## Mermaid — Diagrama de Atividade

```mermaid
flowchart TD
    A[Entrada: content string] --> B[Cria GenerateRequest]
    
    B --> C[Cria Contents array]
    C --> D[Cria Content{Parts: []Part}]
    D --> E[Cria Part{{Text: content}}]
    E --> F[Content.Parts = []Part{E}]
    F --> G[Contents = []Content{F}]
    
    G --> H[Cria GenerationConfig]
    H --> I[MaxOutputTokens = c.MaxTokens]
    I --> J[Temperature, TopK, TopP, StopSequences omitidos]
    
    J --> K[SafetySettings = nil]
    
    K --> L[Retorna GenerateRequest{Contents, GenerationConfig}]
```

---

## Estrutura Gerada

```
GenerateRequest {
    Contents: [
        Content {
            Parts: [
                Part {
                    Text: "<content string do caller>"
                }
            ],
            Role: <não definido>
        }
    ],
    GenerationConfig: {
        MaxOutputTokens: <c.MaxTokens, default 8192>
    },
    SafetySettings: nil
}
```

---

## Campo a Campo

| Campo | Valor | Fonte | Opcional? |
|-------|-------|-------|-----------|
| `Contents[0].Parts[0].Text` | `content` (argumento) | Parâmetro | Não |
| `Contents[0].Role` | **não definido** | — | Sim (`omitempty`) |
| `GenerationConfig.MaxOutputTokens` | `c.MaxTokens` | `Client.MaxTokens` | Sim (`omitempty`) |
| `GenerationConfig.Temperature` | **não definido** | — | Sim (`omitempty`) |
| `GenerationConfig.TopK` | **não definido** | — | Sim (`omitempty`) |
| `GenerationConfig.TopP` | **não definido** | — | Sim (`omitempty`) |
| `GenerationConfig.StopSequences` | **não definido** | — | Sim (`omitempty`) |
| `SafetySettings` | `nil` | — | Sim (`omitempty`) |

---

## Observações

- 🟡 **INFERIDO** — `Role` não é definido. A Gemini API assume `"user"` como default quando omitido.
- A estrutura é mínima — apenas texto puro do prompt. Sem system prompt, sem conversação multi-turn.
- `GenerationConfig` tem apenas `MaxOutputTokens` definido; os demais campos são omitidos e a API usa seus defaults.
- `SafetySettings` é sempre `nil` neste adaptador.
