# Fluxo: Verificação de Adequação ao Context Window

> **Arquivo fonte:** `estimator.go`  
> **Função:** `CheckContextFit()`

---

## Diagrama de Fluxo

```mermaid
flowchart TD
    Start([CheckContextFit(tokens, windowSize)]) --> checkWS["windowSize > 0?"]
    
    checkWS -->|Sim| calcPct["percentage = float64(tokens) / float64(windowSize) * 100"]
    checkWS -->|Não| pctZero["percentage = 0.0"]
    
    pctZero --> checkFits["fits = (tokens <= windowSize)"]
    calcPct --> checkFits
    
    checkFits --> buildResult["ContextFit{WindowSize: windowSize, UsedTokens: tokens, Percentage: percentage, Fits: fits}"]
    
    buildResult --> End([ContextFit])
```

---

## Fluxo Detalhado

```
Input: tokens (int), windowSize (int)
  │
  ├─ Se windowSize > 0:
  │   ├─ percentage = (float64(tokens) / float64(windowSize)) * 100
  │   │   // Conversão para float64 evita overflow de divisão inteira
  │   │   // Exemplo: 10000 / 8192 * 100 = 122.0703125
  │   │
  │─ Senão (windowSize == 0):
  │   ├─ percentage = 0.0
  │   │   // Evita divisão por zero
  │   │
  ├─ fits = (tokens <= windowSize)
  │   │   // true se tokens cabe, false se excede
  │   │   // Nota: tokens <= 0 também resulta em fits = true (mesmo com windowSize = 0)
  │   │
  └─ return ContextFit{
       WindowSize: windowSize,
       UsedTokens: tokens,
       Percentage: percentage,
       Fits: fits,
     }
```

---

## Tabelas de Casos

### Cenário 1: Janela 8K (8192 tokens)

| Tokens | Fits | Percentage | Observação |
|--------|------|------------|------------|
| 0 | `true` | `0.0` | Vazio |
| 100 | `true` | `1.220703125` | Mínimo uso |
| 4096 | `true` | `50.0` | Metade da janela |
| 8192 | `true` | `100.0` | Exatamente cheio |
| 8193 | `false` | `100.01220703125` | 1 token acima |
| 10000 | `false` | `122.0703125` | ~22% excedente |
| 65536 | `false` | `800.0` | 8x a janela |

### Cenário 2: Janela 4K (4096 tokens)

| Tokens | Fits | Percentage | Observação |
|--------|------|------------|------------|
| 0 | `true` | `0.0` | |
| 2000 | `true` | `48.828125` | ~metade |
| 4096 | `true` | `100.0` | Exato |
| 5000 | `false` | `122.0703125` | ~22% excedente |
| 131072 | `false` | `3200.0` | 32x a janela |

### Cenário 3: Janela Zero (0 tokens)

| Tokens | Fits | Percentage | Observação |
|--------|------|------------|------------|
| 0 | `true` | `0.0` | Zero cabe em zero |
| 1 | `false` | `0.0` | 1 token não cabe em janela vazia |
| 100 | `false` | `0.0` | windowSize=0 → percentage sempre 0 |

**Nota:** Quando `windowSize == 0`, o `percentage` é sempre `0.0` (não é NaN ou inf) devido ao `if windowSize > 0` guard.

---

## Casos de Uso na Codebase

| Consumidor | Chamada | Contexto |
|------------|---------|----------|
| `internal/ui/components/progress.go` | Indireto — `FormatTokens` é usado para exibir tokens na barra de progresso | Nenhum `CheckContextFit` direto, mas a lógica de progresso usa tokens |
| `cmd/context.go` | Indireto — `FormatTokens` é usado no sumário | Nenhum `CheckContextFit` direto |

**🟡 INFERIDO:** A função `CheckContextFit` não é chamada diretamente por nenhum código existente. Ela está disponível como utilidade pública, mas o consumo atual usa `EstimateFromBytes()` e `FormatTokens()` diretamente sem a verificação de contexto window. Pode ser utilidade para futuras funcionalidades de corte/trim de contexto.

---

## Limitações

| Limitação | Impacto |
|-----------|---------|
| `Percentage` usa `float64` — possível perda de precisão para valores muito grandes | Baixo — 53 bits de mantissa são suficientes para > 99% dos casos |
| Sem limite máximo de `Percentage` — pode exceder 10000% sem aviso | Baixo — `Fits` já indica overflow |
| `windowSize = 0` retorna `percentage = 0.0` mas `fits = false` (se tokens > 0) — comportamento coerente mas potencialmente confuso | Baixo |
