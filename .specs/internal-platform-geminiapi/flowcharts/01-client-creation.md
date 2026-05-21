# Fluxo: Criação do Cliente Gemini (`NewClient`)

> **Arquivo:** `client.go` → `NewClient()`
> **Arquitetura:** Fábrica com defaults + validação

---

## Mermaid — Diagrama de Atividade

```mermaid
flowchart TD
    A[Entrada: llm.Config] --> B{APIKey vazio?}
    B -- sim --> C[Erro: 'api key is required']
    B -- não --> D{BaseURL vazio?}
    D -- sim --> E[BaseURL = defaults.BaseURL\n'https://generativelanguage.googleapis.com/v1beta']
    D -- não --> F[BaseURL = cfg.BaseURL]
    E --> G
    F --> G
    G{Timeout == 0?}
    G -- sim --> H[Timeout = defaults.Timeout = 300s]
    G -- não --> I[Timeout = cfg.Timeout]
    H --> J{Model vazio?}
    I --> J
    J -- sim --> K[Model = defaults.Model\n'gemini-2.5-flash']
    J -- não --> L[Model = cfg.Model]
    K --> M
    L --> M
    M{MaxTokens == 0?}
    M -- sim --> N[MaxTokens = defaults.MaxTokens = 8192]
    M -- não --> O[MaxTokens = cfg.MaxTokens]
    N --> P
    O --> P
    P[Create BaseClient:\n  - JSONClient(HTTP transport)\n  - APIKey, Model, MaxTokens,\n    ProviderName='Gemini'] --> Q{Erros?}
    Q -- sim --> R[Erro propagado]
    Q -- não --> S[Client{BaseClient} criado]
    S --> T[Saída: *Client]
```

## Descrição Textual

1. **Entrada:** `llm.Config` com campos opcionais.
2. **Validação:** Verifica `cfg.APIKey`. Se vazio, retorna erro imediato.
3. **BaseURL:** Usa `cfg.BaseURL` se fornecido, senão usa `defaults.BaseURL` (endpoint Gemini).
4. **Timeout:** Usa `cfg.Timeout` se > 0, senão `defaults.Timeout` (300s).
5. **Model:** Usa `cfg.Model` se fornecido, senão `defaults.Model` ("gemini-2.5-flash").
6. **MaxTokens:** Usa `cfg.MaxTokens` se > 0, senão `defaults.MaxTokens` (8192).
7. **Criação:** `llmbase.NewBaseClient()` cria `BaseClient` com `platformhttp.NewJSONClient()`.
8. **Retorno:** `&Client{BaseClient: base}`.

---

## Pontos de Decisão

| Nº | Condição | Caminho "sim" | Caminho "não" |
|----|----------|---------------|---------------|
| D1 | APIKey vazio? | Retorno de erro | Continua |
| D2 | BaseURL vazio? | Usa default | Usa cfg.BaseURL |
| D3 | Timeout == 0? | Usa default (300s) | Usa cfg.Timeout |
| D4 | Model vazio? | Usa default ("gemini-2.5-flash") | Usa cfg.Model |
| D5 | MaxTokens == 0? | Usa default (8192) | Usa cfg.MaxTokens |

---

## Valores Padrão (Hardcoded)

| Parâmetro | Valor Padrão | Fonte |
|-----------|-------------|-------|
| BaseURL | `"https://generativelanguage.googleapis.com/v1beta"` | `defaultBaseURL` |
| Model | `"gemini-2.5-flash"` | `DefaultConfig.Model` |
| Timeout | `300s` | `DefaultConfig.Timeout` |
| MaxTokens | `8192` | `defaultMaxTokens` |
| ProviderName | `"Gemini"` | Hardcoded em `NewBaseClientWithDefaults` |
