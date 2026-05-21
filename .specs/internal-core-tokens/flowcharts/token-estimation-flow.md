# Fluxo: Estimativa de Tokens

> **Arquivo fonte:** `estimator.go`  
> **Funções:** `Estimate()`, `EstimateFromBytes()`

---

## Diagrama de Fluxo

```mermaid
flowchart TD
    Start([Estimate(text)]) --> lenText["text = len(text) bytes"]
    lenText --> toFromBytes["EstimateFromBytes(int64(lenText))"]
    
    toFromBytes --> fromBytesStart([EstimateFromBytes])
    fromBytesStart --> checkZero["size <= 0?"]
    checkZero -->|Sim| returnZero["return 0"]
    checkZero -->|Não| calcCeil["calcCeil: (size + BytesPerToken - 1) / BytesPerToken"]
    
    calcCeil --> detailCalc["  (size + 4 - 1) / 4"]
    detailCalc --> detailCalc2["  = (size + 3) / 4"]
    detailCalc2 --> returnCeil["return int(calcCeil)"]
    
    returnZero --> End([int: estimated tokens])
    returnCeil --> End
```

### Explicação Detalhada do Algoritmo `EstimateFromBytes`

O cálculo de estimativa usa **ceil division** (arredondamento para cima) para garantir que cada byte adicional gere pelo menos 1 token:

```
tokens = ceil(bytes / 4)

Implementação Go: (bytes + 4 - 1) / 4
                = (bytes + 3) / 4   ← divisão inteira com arredondamento para cima
```

**Tabela de conversão (bytes → tokens):**

| Bytes | Cálculo | Tokens | Observação |
|-------|---------|--------|------------|
| 0 | ceil(0/4) | 0 | Sem conteúdo |
| 1 | ceil(1/4) | 1 | Mínimo 1 token |
| 2 | ceil(2/4) | 1 | |
| 3 | ceil(3/4) | 1 | |
| 4 | ceil(4/4) | 1 | Exato divisor |
| 5 | ceil(5/4) | 2 | Primeiro restolho |
| 8 | ceil(8/4) | 2 | Exato divisor |
| 100 | ceil(100/4) | 25 | |
| 1024 | ceil(1024/4) | 256 | 1 KB |
| 13 | ceil(13/4) | 4 | "Hello, world!" |

---

## Caminhos de Execução

### Caminho 1: Entrada Vazia ou Negativa
```
Input: bytes = 0 (ou qualquer valor <= 0)
  → checkZero = true
  → returnZero
  → Output: 0 tokens
```

### Caminho 2: Entrada Positiva — Divisão Exata
```
Input: bytes = 8
  → checkZero = false
  → calcCeil = (8 + 3) / 4 = 11 / 4 = 2 (divisão inteira)
  → Output: 2 tokens
```

### Caminho 3: Entrada Positiva — Restolho
```
Input: bytes = 5
  → checkZero = false
  → calcCeil = (5 + 3) / 4 = 8 / 4 = 2
  → Output: 2 tokens (1 token para os 4 primeiros bytes + 1 para o byte restante)
```

---

## Dependências Internas

| Entidade | Chamada | Localização |
|----------|---------|-------------|
| `BytesPerToken` (constante = 4) | Usada em fórmula de ceil division | `estimator.go:36` |

---

## Pontos de Atenção

| Item | Severidade | Descrição |
|------|------------|-----------|
| `len(text)` conta bytes UTF-8 | Informativo | Textos multibyte (emojis, japonês) serão sobreestimados em ~2-4x vs tokenizers reais |
| Funções puras, sem side effects | Positivo | Totalmente determinísticas e testáveis |
| `int64` para `EstimateFromBytes` | Positivo | Suporta tamanhos > 2GB (64-bit) |
