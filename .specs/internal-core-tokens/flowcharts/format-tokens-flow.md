# Fluxo: Formatação de Tokens (FormatTokens)

> **Arquivo fonte:** `estimator.go`  
> **Função:** `FormatTokens()`

---

## Diagrama de Fluxo

```mermaid
flowchart TD
    Start([FormatTokens(tokens int)]) --> checkM["tokens >= 1,000,000?"]
    
    checkM -->|Sim| fmtM["return fmt.Sprintf("%.1fM", float64(tokens)/1000000)"]
    checkM -->|Não| checkK["tokens >= 1,000?"]
    
    checkK -->|Sim| fmtK["return fmt.Sprintf("%.1fK", float64(tokens)/1000)"]
    checkK -->|Não| fmtD["return fmt.Sprintf("%d", tokens)"]
    
    fmtM --> End([string])
    fmtK --> End
    fmtD --> End
```

---

## Fluxo Detalhado

```
Input: tokens (int)
  │
  ├─ Se tokens >= 1_000_000:
  │   ├─ value = float64(tokens) / 1_000_000
  │   ├─ result = "%.1fM" format(value)
  │   │   // Ex: 1500000 → "1.5M"
  │   │
  ├─ Senão se tokens >= 1_000:
  │   ├─ value = float64(tokens) / 1_000
  │   ├─ result = "%.1fK" format(value)
  │   │   // Ex: 1500 → "1.5K"
  │   │
  └─ Senão (tokens < 1_000):
      ├─ result = "%d" format(tokens)
      │   // Ex: 42 → "42", 999 → "999"
      │
      └─ return result
```

---

## Tabela de Formatação Completa

| Tokens | Faixa | Formato | Resultado |
|--------|-------|---------|-----------|
| `-100` | default | `"%d"` | `"-100"` |
| `0` | default | `"%d"` | `"0"` |
| `1` | default | `"%d"` | `"1"` |
| `999` | default | `"%d"` | `"999"` |
| `1000` | K | `"%.1fK"` | `"1.0K"` |
| `1001` | K | `"%.1fK"` | `"1.0K"` |
| `1500` | K | `"%.1fK"` | `"1.5K"` |
| `9999` | K | `"%.1fK"` | `"10.0K"` |
| `10000` | K | `"%.1fK"` | `"10.0K"` |
| `99999` | K | `"%.1fK"` | `"100.0K"` |
| `100000` | K | `"%.1fK"` | `"100.0K"` |
| `999999` | K | `"%.1fK"` | `"1000.0K"` |
| `1000000` | M | `"%.1fM"` | `"1.0M"` |
| `1500000` | M | `"%.1fM"` | `"1.5M"` |
| `2000000` | M | `"%.1fM"` | `"2.0M"` |
| `1234567` | M | `"%.1fM"` | `"1.2M"` |

---

## Comportamentos de Borda

### Valores Negativos
```
FormatTokens(-100) → "-100"
```
Valores negativos não são tratados separadamente — caem no `default` case. **Informativo:** tokens negativos não têm significado semântico neste contexto, mas a função não valida nem rejeita entradas negativas.

### Limite K (1.000)
```
FormatTokens(999)  → "999"    (default case)
FormatTokens(1000) → "1.0K"   (K case)
```
Transição limpa: não há valores K com `.0` fracionário no limite exato.

### Limite M (1.000.000)
```
FormatTokens(999999) → "1000.0K"  (K case)
FormatTokens(1000000) → "1.0M"    (M case)
```

**Nota:** `999999` é formatado como `"1000.0K"`, não como `"999.9K"`. Isso ocorre porque `999999 / 1000 = 999.999`, que é arredondado para `"1000.0K"` pelo formato `%.1f`.

---

## Propriedades

| Propriedade | Valor |
|-------------|-------|
| Monotonicidade | ✅ `tokens1 < tokens2` → `FormatTokens(tokens1) ≤ lexicographic FormatTokens(tokens2)` |
| Determinismo | ✅ Mesma entrada → mesma saída |
| Sem state | ✅ Função pura |
| Allocated string | ✅ Cada chamada aloca uma nova string (via `fmt.Sprintf`) |

---

## Consumidores na Codebase

| Arquivo | Chamada | Resultado de Uso |
|---------|---------|-----------------|
| `cmd/context.go` | `tokens.FormatTokens(int(result.TokenEstimate))` | Exibe "~1.2K tokens" no sumário de geração |
| `internal/ui/components/progress.go` | `tokens.FormatTokens(b.TotalTokens)` | Exibe tokens atuais na barra de progresso TUI |
| `internal/ui/screens/file_selection.go` | `tokens.FormatTokens(estimatedTokens)` | Exibe tokens estimados na tela de seleção |
| `internal/ui/screens/review.go` | `tokens.FormatTokens(m.totalTokens)` | Exibe tokens totais na tela de revisão |

**Padrão:** Todos os consumidores usam `FormatTokens()` para **exibição** — nunca para lógica de decisão. O formato é projetado para legibilidade em interfaces de terminal (TUI).
