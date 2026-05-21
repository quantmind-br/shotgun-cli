# Dicionário de Dados — `internal/core/tokens`

| Campo           | Valor                                                                 |
|-----------------|-----------------------------------------------------------------------|
| **Módulo**      | `internal/core/tokens`                                                |
| **Package**     | `github.com/quantmind-br/shotgun-cli/internal/core/tokens`            |
| **Nível de detalhe** | detalhado                                                      |
| **Arquivos analisados** | `estimator.go` (1), `estimator_test.go` (1)                   |

---

## 1. Constantes Globais

### 1.1 `BytesPerToken`

| Campo | Valor |
|-------|-------|
| **Tipo** | `int` (não tipado, constante numérica) |
| **Valor** | `4` |
| **Escopo** | Pacote (não exportada como identificador, mas referenciada em fórmulas) |
| **Descrição** | Heurística: número médio de bytes por token para textos em inglês (GPT-style tokenizers, cl100k_base). |
| **Uso em** | `EstimateFromBytes()`, `BytesFromTokens()` |

**Nota:** A constante é usada diretamente em expressões aritméticas (`+ BytesPerToken - 1`), não como variável global nomeada exportada.

### 1.2 Constantes de Context Window

| Constante | Valor | Tipo | Descrição |
|-----------|-------|------|-----------|
| `Window4K` | `4096` | `int` | Context window de 4K tokens |
| `Window8K` | `8192` | `int` | Context window de 8K tokens |
| `Window16K` | `16384` | `int` | Context window de 16K tokens |
| `Window32K` | `32768` | `int` | Context window de 32K tokens |
| `Window64K` | `65536` | `int` | Context window de 64K tokens |
| `Window128K` | `131072` | `int` | Context window de 128K tokens |

**Nota:** Estas constantes são definidas no pacote mas **não são consumidas internamente** por nenhuma função. Servem como referência de documentação. Podem ser usadas por consumidores para criar `ContextFit` com valores padrão.

---

## 2. Tipos Públicos

### 2.1 `Stats`

Estrutura que agrega estatísticas de tamanho (bytes) e estimativa de tokens para conteúdo.

| Campo | Tipo | Zero Value | Obrigatório | Mutável | Descrição |
|-------|------|------------|-------------|---------|-----------|
| `Bytes` | `int64` | `0` | Não | Sim | Contagem bruta de bytes do conteúdo. |
| `Tokens` | `int` | `0` | Não | Sim | Estimativa de tokens calculada a partir de `Bytes`. |

**Valores típicos:**
| Cenário | `Bytes` | `Tokens` (aproximado) |
|---------|---------|----------------------|
| String vazia | `0` | `0` |
| "Hello" | `5` | `2` |
| README.md (50KB) | `51200` | `12800` |
| 1MB de texto | `1048576` | `262144` |

**Métodos/Construtores associados:**

| Entidade | Tipo | Parâmetros | Retorno | Descrição |
|----------|------|------------|---------|-----------|
| `NewStats` | Função | `bytes int64` | `Stats` | Cria `Stats` calculando `Tokens` a partir de `Bytes`. |
| `NewStatsFromText` | Função | `text string` | `Stats` | Converte `len(text)` para bytes e constrói `Stats`. |
| `Stats.Add` | Método receiver | `other Stats` | `Stats` | Retorna nova `Stats` com soma dos campos. |

**Comportamento de `Add`:**
```go
// Exemplo:
s1 := Stats{Bytes: 100, Tokens: 25}
s2 := Stats{Bytes: 200, Tokens: 50}
s1.Add(s2) → Stats{Bytes: 300, Tokens: 75}
```

**🟡 INFERIDO — Limitação de precisão:** `Add()` soma tokens diretamente em vez de recalculá-los a partir da soma de bytes. Para valores pequenos com restolhos, pode haver diferença máxima de 2 tokens em relação a `EstimateFromBytes(s1.Bytes + s2.Bytes)`.

---

### 2.2 `ContextFit`

Estrutura que descreve como um conjunto de tokens cabe dentro de uma janela de contexto específica.

| Campo | Tipo | Zero Value | Obrigatório | Mutável | Descrição |
|-------|------|------------|-------------|---------|-----------|
| `WindowSize` | `int` | `0` | Sim | Não | Tamanho alvo da janela de contexto em tokens (fornecido pelo caller). |
| `UsedTokens` | `int` | `0` | Sim | Não | Número de tokens estimados que o conteúdo ocupa. |
| `Percentage` | `float64` | `0.0` | Sim | Não | Porcentagem da janela utilizada: `(UsedTokens / WindowSize) * 100`. Zero se `WindowSize == 0`. |
| `Fits` | `bool` | `false` | Sim | Não | `true` se `UsedTokens <= WindowSize`, indicando que o conteúdo cabe na janela. |

**Valores típicos de `Percentage`:**

| Cenário | UsedTokens | WindowSize | Percentage | Fits |
|---------|------------|------------|------------|------|
| Conteúdo vazio | 0 | 8192 | `0.0` | `true` |
| Metade preenchido | 4096 | 8192 | `50.0` | `true` |
| Exatamente cheio | 8192 | 8192 | `100.0` | `true` |
| Excedendo | 10000 | 8192 | `122.0703125` | `false` |
| Janela zero | 100 | 0 | `0.0` | `false` |
| Muito excedendo | 200000 | 4096 | `4882.8125` | `false` |

**Método/Construtor associado:**

| Entidade | Tipo | Parâmetros | Retorno | Descrição |
|----------|------|------------|---------|-----------|
| `CheckContextFit` | Função | `tokens int`, `windowSize int` | `ContextFit` | Avalia se tokens cabem na janela e calcula porcentagem. |

**Caso especial — `windowSize == 0`:**
- `Percentage = 0.0` (divisão por zero evitada)
- `Fits = (tokens <= 0)` — só `true` se tokens for zero ou negativo

---

## 3. Funções Globais (Exportadas)

### 3.1 `Estimate(text string) int`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(text string) int` |
| **Retorno** | `int` — Estimativa de tokens |
| **Complexidade** | O(1) — `len(text)` + divisão |
| **Efeitos colaterais** | Nenhum |
| **Purity** | Pura |
| **Descrição** | Retorna a estimativa de tokens para a string. Internamente chama `EstimateFromBytes(int64(len(text)))`. |

**Algoritmo:** `ceil(len(text) / 4)`

**Casos de borda:**
| Entrada | Resultado |
|---------|-----------|
| `""` (vazio) | `0` |
| `"a"` (1 byte) | `1` |
| `"abcd"` (4 bytes) | `1` |
| `"abcde"` (5 bytes) | `2` |
| String com emojis (bytes > caracteres) | Sobreestimado — emojis UTF-8 são 4 bytes cada |

---

### 3.2 `EstimateFromBytes(size int64) int`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(size int64) int` |
| **Retorno** | `int` — Estimativa de tokens |
| **Complexidade** | O(1) |
| **Efeitos colaterais** | Nenhum |
| **Purity** | Pura |
| **Descrição** | Função central. Converte bytes para tokens usando a heurística 4 bytes/token com arredondamento para cima. |

**Algoritmo:**
```go
if size <= 0: return 0
return (size + 3) / 4  // ceil division
```

**Casos de borda:**
| Entrada | Resultado | Racional |
|---------|-----------|----------|
| `0` | `0` | Sem conteúdo |
| `-1` | `0` | Negativo tratado como zero |
| `1` | `1` | 1 byte ≥ 1 token mínimo |
| `4` | `1` | Exatamente 1 token |
| `5` | `2` | Restolho = mais 1 token |
| `1024` | `256` | 1 KB = 256 tokens |
| `1048576` | `262144` | 1 MB = 262.144 tokens |

**Propriedades matemáticas:**
- Monotonicamente crescente: `size1 < size2` → `EstimateFromBytes(size1) ≤ EstimateFromBytes(size2)`
- Homogeneidade aproximada: `EstimateFromBytes(k * n) ≈ k * EstimateFromBytes(n)` para k > 0
- Idempotência parcial: `EstimateFromBytes(EstimateFromBytes(n) * 4) == n` apenas quando `n` é divisível por 4

---

### 3.3 `BytesFromTokens(tokens int) int64`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(tokens int) int64` |
| **Retorno** | `int64` — Tamanho aproximado em bytes |
| **Complexidade** | O(1) |
| **Efeitos colaterais** | Nenhum |
| **Purity** | Pura |
| **Descrição** | Converte tokens de volta para bytes aproximados. Multiplicação simples. |

**Algoritmo:** `int64(tokens) * 4`

**Relação com `EstimateFromBytes`:**
- `BytesFromTokens(EstimateFromBytes(n)) ≥ n` — sempre igual ou maior devido ao ceil
- `EstimateFromBytes(BytesFromTokens(t)) ≥ t` — sempre igual ou maior (pelo menos `t * 4` bytes, que arredonda para `t` tokens)
- A inversa exata só vale quando `n` é múltiplo de 4: `BytesFromTokens(EstimateFromBytes(4*k)) == 4*k`

---

### 3.4 `FormatTokens(tokens int) string`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(tokens int) string` |
| **Retorno** | `string` — Representação humana legível |
| **Complexidade** | O(1) — switch com comparações constantes |
| **Efeitos colaterais** | Nenhum (alocação de string por `fmt.Sprintf`) |
| **Purity** | Pura |
| **Descrição** | Formata contagem de tokens com sufixos K/M para legibilidade. |

**Regras de formatação:**
| Faixa de input | Formato | Exemplos |
|----------------|---------|----------|
| `0–999` | `"%d"` | `"0"`, `"42"`, `"999"` |
| `1,000–999,999` | `"%.1fK"` | `"1.0K"`, `"1.5K"`, `"999.9K"` |
| `≥ 1,000,000` | `"%.1fM"` | `"1.0M"`, `"1.5M"`, `"100.0M"` |

**Comportamentos de borda:**
| Entrada | Resultado | Observação |
|---------|-----------|------------|
| `0` | `"0"` | Zero é tratado no `default` |
| `999` | `"999"` | Limite inferior do default |
| `1000` | `"1.0K"` | Limite superior do default |
| `-1` | `"-1"` | Valores negativos não são tratados separadamente |

---

### 3.5 `NewStats(bytes int64) Stats`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(bytes int64) Stats` |
| **Retorno** | `Stats` |
| **Efeito colateral** | Nenhum |
| **Descrição** | Construtor que calcula tokens a partir de bytes usando `EstimateFromBytes()`. |

---

### 3.6 `NewStatsFromText(text string) Stats`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(text string) Stats` |
| **Retorno** | `Stats` |
| **Efeito colateral** | Nenhum |
| **Descrição** | Construtor que calcula estatísticas a partir do conteúdo de texto. Internamente chama `NewStats(int64(len(text)))`. |

---

### 3.7 `CheckContextFit(tokens int, windowSize int) ContextFit`

| Campo | Valor |
|-------|-------|
| **Tipo** | `func(tokens int, windowSize int) ContextFit` |
| **Retorno** | `ContextFit` |
| **Efeito colateral** | Nenhum |
| **Descrição** | Avalia se tokens cabem em uma janela de contexto específica. |

**Lógica detalhada:**
```
se windowSize > 0:
    percentage = (float64(tokens) / float64(windowSize)) * 100
senão:
    percentage = 0.0

fits = (tokens <= windowSize)

return ContextFit{
    WindowSize: windowSize,
    UsedTokens: tokens,
    Percentage: percentage,
    Fits: fits,
}
```

**Casos de borda:**
| tokens | windowSize | Fits | Percentage | Observação |
|--------|------------|------|------------|------------|
| 0 | 8192 | `true` | `0.0` | Conteúdo vazio |
| 0 | 0 | `true` | `0.0` | Zero cabe em qualquer janela |
| 100 | 0 | `false` | `0.0` | Janela zero, tokens não cabe |
| 8192 | 8192 | `true` | `100.0` | Exatamente cheio |
| 8193 | 8192 | `false` | `100.012207...` | Ultrajante de 1 token |

---

## 4. Testes — Casos Cobertos

### 4.1 `TestEstimate`

| Test Case | Entrada | Saída Esperada | Cobertura |
|-----------|---------|----------------|-----------|
| `empty` | `""` | `0` | String vazia |
| `single char` | `"a"` | `1` | 1 byte |
| `four chars` | `"abcd"` | `1` | Exato divisor (4) |
| `five chars` | `"abcde"` | `2` | Restolho (5) |
| `eight chars` | `"abcdefgh"` | `2` | Exato (8) |
| `typical sentence` | `"Hello, world!"` | `4` | 13 bytes → ceil(13/4) = 4 |

### 4.2 `TestEstimateFromBytes`

| Test Case | Entrada | Saída Esperada | Cobertura |
|-----------|---------|----------------|-----------|
| `zero` | `0` | `0` | Limite inferior |
| `negative` | `-10` | `0` | Entrada negativa |
| `one byte` | `1` | `1` | 1 byte |
| `four bytes` | `4` | `1` | Divisor exato |
| `five bytes` | `5` | `2` | Restolho |
| `1KB` | `1024` | `256` | Potência de 2 |
| `1MB` | `1048576` | `262144` | Valor grande |

### 4.3 `TestBytesFromTokens`

| Test Case | Tokens | Saída Esperada | Cobertura |
|-----------|--------|----------------|-----------|
| — | `0` | `0` | Zero |
| — | `1` | `4` | Mínimo |
| — | `100` | `400` | Escala |
| — | `1000` | `4000` | Grande |

### 4.4 `TestStats`

| Sub-test | Cobertura |
|----------|-----------|
| `NewStats` | Verifica que `Stats{Bytes:1024, Tokens:256}` é criado corretamente |
| `NewStatsFromText` | Verifica `"Hello"` → `Stats{Bytes:5, Tokens:2}` |
| `Add` | Verifica `s1.Add(s2)` = `{300, 75}` (25+50 tokens) |

### 4.5 `TestFormatTokens`

| Test Case | Tokens | Saída Esperada | Cobertura |
|-----------|--------|----------------|-----------|
| — | `0` | `"0"` | Zero |
| — | `100` | `"100"` | Abaixo de K |
| — | `999` | `"999"` | Limite inferior K |
| — | `1000` | `"1.0K"` | Limite superior default, entrada K |
| — | `1500` | `"1.5K"` | Meio K |
| — | `10000` | `"10.0K"` | Múltiplos de K |
| — | `100000` | `"100.0K"` | Quase M |
| — | `1000000` | `"1.0M"` | Limite M |
| — | `1500000` | `"1.5M"` | Acima de M |

### 4.6 `TestCheckContextFit`

| Test Case | Tokens | WindowSize | Fits | Percentage | Cobertura |
|-----------|--------|------------|------|------------|-----------|
| `empty in 8K` | `0` | `8192` | `true` | `0` | Zero tokens |
| `half filled` | `4096` | `8192` | `true` | `50` | 50% |
| `exactly full` | `8192` | `8192` | `true` | `100` | 100% exato |
| `overflow` | `10000` | `8192` | `false` | `~122.07` | Excesso |
| `zero window` | `100` | `0` | `false` | `0` | Janela zero |

---

## 5. Relações entre Tipos

```
┌─────────────────┐         ┌──────────────┐
│  string/text    │─len()──▶│  int64 (bytes)│
└─────────────────┘         └──────┬───────┘
                                   │
                    ┌──────────────┼──────────────┐
                    │              │              │
                    ▼              ▼              ▼
              EstimateFromBytes  NewStats    NewStatsFromText
                    │              │              │
                    │         ┌────┴────┐         │
                    │         │   Stats  │◀────────┘
                    │         └────┬────┘         │
                    │              │              │
                    │     ┌────────┴────────┐     │
                    │     │                 │     │
                    ▼     ▼                 ▼     ▼
              BytesFromTokens  Stats.Add()  (combined)
                    │                           │
                    │              ┌────────────┴────────────┐
                    │              │                           │
                    ▼              ▼                           ▼
              int64 (bytes)   Stats                       ContextFit
                                             CheckContextFit
                                                │
                                                ▼
                                         ContextFit
                                         {WindowSize, UsedTokens,
                                          Percentage, Fits}
```
