# Fluxo: Análise da Resposta (`ParseResponse`)

> **Arquivo:** `client.go` → `Client.ParseResponse()`
> **Tipo:** Método do Strategy `llmbase.Sender`

---

## Mermaid — Diagrama de Atividade

```mermaid
flowchart TD
    A[Entrada: response interface{}, rawJSON []byte] --> B[Type-assert: *GenerateResponse]
    B --> C{Type assert sucesso?}
    C -- não --> D[Erro: 'unexpected response type']
    C -- sim --> E{genResp.Error != nil?}
    
    E -- sim --> F[Erro: 'API error [genResp.Error.Code]: genResp.Error.Message']
    
    E -- não --> G{len(genResp.Candidates) > 0?}
    G -- não --> H[Erro: 'no candidates in response']
    
    G -- sim --> I[responseText = '']
    I --> J[Loop: genResp.Candidates[0].Content.Parts]
    J --> K[responseText += part.Text]
    K --> L{Mais partes?}
    L -- sim --> J
    L -- não --> M
    
    M{genResp.UsageMetadata != nil?}
    M -- sim --> N[Cria llm.Usage]
    N --> O[PromptTokens = UsageMetadata.PromptTokenCount]
    O --> P[CompletionTokens = UsageMetadata.CandidatesTokenCount]
    P --> Q[TotalTokens = UsageMetadata.TotalTokenCount]
    M -- não --> R[Usage = nil]
    
    N --> S
    R --> S[Monta llm.Result]
    S --> T[Response = responseText]
    T --> U[RawResponse = string(rawJSON)]
    U --> V[Model = c.Model]
    V --> W[Provider = c.ProviderName = 'Gemini']
    W --> X[Usage = usage ou nil]
    X --> Y[Saída: *llm.Result]
    D --> Z[Saída: nil, error]
    F --> Z
    H --> Z
```

---

## Descrição Textual das Etapas

### Etapa 1: Type Assert

```go
genResp, ok := response.(*GenerateResponse)
if !ok {
    return nil, fmt.Errorf("unexpected response type")
}
```

- O `response` é o objeto `&GenerateResponse{}` retornado por `NewResponse()` e unmarshaled pelo `BaseClient`.
- Em condições normais, o assert sempre succeeds.

### Etapa 2: Verificação de Erro da API

```go
if genResp.Error != nil {
    return nil, fmt.Errorf("API error [%d]: %s", genResp.Error.Code, genResp.Error.Message)
}
```

- Se a API retornou um erro JSON, extrai `Code` e `Message`.
- `Status` da `APIError` **não** é incluído na mensagem.

### Etapa 3: Verificação de Candidatos

```go
if len(genResp.Candidates) == 0 {
    return nil, fmt.Errorf("no candidates in response")
}
```

- Respostas válidas sempre têm pelo menos 1 candidato.

### Etapa 4: Extração do Texto de Resposta

```go
var responseText string
for _, part := range genResp.Candidates[0].Content.Parts {
    responseText += part.Text
}
```

- Itera sobre todos os `Part` do primeiro candidato.
- Concatena sequencialmente (sem separador).
- Usa apenas o primeiro candidato (`Candidates[0]`).

### Etapa 5: Extração de Usage

```go
var usage *llm.Usage
if genResp.UsageMetadata != nil {
    usage = &llm.Usage{
        PromptTokens:     genResp.UsageMetadata.PromptTokenCount,
        CompletionTokens: genResp.UsageMetadata.CandidatesTokenCount,
        TotalTokens:      genResp.UsageMetadata.TotalTokenCount,
    }
}
```

- Mapeia 1:1 entre `UsageMetadata` e `llm.Usage`.
- Se `UsageMetadata` é nil, `usage` permanece nil.

### Etapa 6: Construção do `llm.Result`

```go
return &llm.Result{
    Response:    responseText,
    RawResponse: string(rawJSON),
    Model:       c.Model,
    Provider:    c.ProviderName,   // "Gemini"
    Usage:       usage,
}, nil
```

- `Duration` **não** é preenchido aqui — é calculado pelo `BaseClient.Send()` após `ParseResponse` retornar.

---

## Mapeamento de Campos

| Origem (`GenerateResponse`) | Destino (`llm.Result`) | Tipo |
|-----------------------------|------------------------|------|
| `Candidates[0].Content.Parts[].Text` | `Result.Response` | `string` (concatenado) |
| `rawJSON` (input) | `Result.RawResponse` | `string` |
| `c.Model` | `Result.Model` | `string` |
| `c.ProviderName` | `Result.Provider` | `string` ("Gemini") |
| `UsageMetadata.PromptTokenCount` | `Result.Usage.PromptTokens` | `int` |
| `UsageMetadata.CandidatesTokenCount` | `Result.Usage.CompletionTokens` | `int` |
| `UsageMetadata.TotalTokenCount` | `Result.Usage.TotalTokens` | `int` |

---

## Caminhos de Retorno

| Cenário | Retorno |
|---------|---------|
| Type assert falha | `nil, "unexpected response type"` |
| `genResp.Error != nil` | `nil, "API error [X]: Y"` |
| `len(Candidates) == 0` | `nil, "no candidates in response"` |
| `UsageMetadata == nil` | `&llm.Result{...Usage: nil}, nil` |
| Sucesso | `&llm.Result{Response, RawResponse, Model, Provider, Usage}, nil` |
