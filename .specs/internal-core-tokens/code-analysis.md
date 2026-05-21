# Análise de Código — Módulo `internal/core/tokens`

> **Módulo:** `internal/core/tokens`  
> **Caminho do pacote:** `github.com/quantmind-br/shotgun-cli/internal/core/tokens`  
> **Nível de detalhe:** detalhado  
> **Data da análise:** 2026-05-20  
> **Gerado por:** reversa-archaeologist

---

## 1. Visão Geral

O pacote `internal/core/tokens` fornece **utilitários de estimativa de tokens** para otimização de contexto LLM. Ele implementa uma heurística baseada na proporção **1 token ≈ 4 bytes (caracteres ASCII)** para textos em inglês, aproximando-se do comportamento de tokenizers GPT-style (cl100k_base). O pacote é **zero-dependência** (apenas stdlib `fmt`) e é usado amplamente em toda a codebase para estimativas de tamanho de contexto, barras de progresso e relatórios de uso.

### Responsabilidades Principais

| Responsabilidade | Arquivo | Entidades |
|---|---|---|
| Estimativa de tokens a partir de texto | `estimator.go` | `Estimate()`, `EstimateFromBytes()` |
| Conversão inversa (tokens → bytes) | `estimator.go` | `BytesFromTokens()` |
| Agrupamento de estatísticas | `estimator.go` | `Stats`, `NewStats()`, `NewStatsFromText()`, `Stats.Add()` |
| Formatação humana de contagem de tokens | `estimator.go` | `FormatTokens()` |
| Verificação de adequação ao context window | `estimator.go` | `ContextFit`, `CheckContextFit()` |
| Constantes de heurística | `estimator.go` | `BytesPerToken`, `Window4K`–`Window128K` |

### Métricas do Módulo

| Métrica | Valor |
|---|---|
| Arquivos `.go` (excl. testes) | 1 |
| Arquivos `_test.go` | 1 |
| Funções globais (exportadas) | 6 |
| Funções globais (não exportadas) | 0 |
| Tipos definidos (exportados) | 2 (`Stats`, `ContextFit`) |
| Tipos construtores (exportados) | 3 (`NewStats`, `NewStatsFromText`, `CheckContextFit`) |
| Constantes definidas | 7 |
| Dependências externas (pacotes) | 0 (`fmt` é stdlib) |
| Dependências internas (pacotes) | 0 |
| Usos em outros pacotes | 6 arquivos, 9 chamadas |

---

## 2. Dependências Internas (pacotes)

Nenhum. O pacote é **autocontido** e depende apenas da biblioteca padrão Go.

---

## 3. Dependências Externas

| Pacote | Tipo | Uso |
|---|---|---|
| `fmt` | stdlib | `fmt.Sprintf()` em `FormatTokens()` |

---

## 4. Análise Detalhada por Arquivo

### 4.1 `estimator.go` — Estimador de Tokens

**Arquitetura:** Arquivo único, funcional. Sem structs complexas, sem interfaces. Apenas funções puras + structs simples + constantes.

#### 4.1.1 Constantes de Heurística

```go
BytesPerToken = 4
Window4K   = 4096
Window8K   = 8192
Window16K  = 16384
Window32K  = 32768
Window64K  = 65536
Window128K = 131072
```

- `BytesPerToken` é a constante central de toda a lógica do pacote. Valor `4` é uma aproximação padrão para tokenizers GPT-4/GPT-3.5 (cl100k_base), onde ~1 token corresponde a ~4 bytes de texto ASCII em inglês.
- As constantes `Window*` são **valores de referência** para context windows comuns de LLMs. São usadas como documentação de referência, não diretamente no código de estimativa.

**🟡 INFERIDO:** As constantes `Window*` não são usadas internamente pelo pacote — apenas documentadas. Elas são consumidas apenas externamente (se houver) para referência.

#### 4.1.2 `Estimate(text string) int`

Retorna a estimativa de tokens para o texto dado. Internamente delega para `EstimateFromBytes(int64(len(text)))`.

**Algoritmo:** `len(text)` retorna o número de bytes (não caracteres Unicode). Para strings ASCII, cada byte = 1 token/4. Para strings multibyte UTF-8, a estimativa é **conservadora** — textos com caracteres não-ASCII serão superestimados ligeiramente em tokens (pois cada byte conta separadamente, mas em tokenizers reais, caracteres multibyte podem ser tokens únicos).

#### 4.1.3 `EstimateFromBytes(size int64) int`

Função central da biblioteca. Implementa **ceil division** usando a fórmula `(size + divisor - 1) / divisor`:

- Se `size <= 0` → retorna `0`.
- Se `size > 0` → `(size + 3) / 4` (ceil de `size / 4`).

**Exemplos de comportamento (validado pelos testes):**
| Input | Saída | Racional |
|---|---|---|
| 0 | 0 | Sem conteúdo |
| 1 | 1 | 1/4 = 0.25 → ceil = 1 |
| 4 | 1 | 4/4 = 1.0 → ceil = 1 |
| 5 | 2 | 5/4 = 1.25 → ceil = 2 |
| 1024 | 256 | 1024/4 = 256 |
| -10 | 0 | Entrada negativa |

#### 4.1.4 `BytesFromTokens(tokens int) int64`

Conversão inversa. Multiplicação simples: `tokens * 4`.

**Assimetria:** Não é a inversa exata de `EstimateFromBytes()` devido ao arredondamento. Por exemplo:
- `EstimateFromBytes(5) = 2`
- `BytesFromTokens(2) = 8` ≠ 5

Esta é uma **limitação intrínseca** da heurística de 4 bytes/token. O pacote não tenta corrigir isso.

#### 4.1.5 `Stats` — Agregador de Estatísticas

Struct simples com dois campos:
| Campo | Tipo | Descrição |
|---|---|---|
| `Bytes` | `int64` | Contagem bruta de bytes |
| `Tokens` | `int` | Contagem estimada de tokens |

**Métodos/Construtores:**
- `NewStats(bytes int64) Stats` — Calcula tokens a partir de bytes usando `EstimateFromBytes()`.
- `NewStatsFromText(text string) Stats` — Usa `len(text)` como bytes.
- `Stats.Add(other Stats) Stats` — Soma ambos os campos. **Nota:** Não recalcula tokens a partir da soma de bytes (usa a soma direta de tokens individuais), o que pode levar a pequenas discrepâncias em relação a `EstimateFromBytes(s.Bytes + other.Bytes)`.

**Comportamento de `Add`:**
```
s1: {Bytes: 100, Tokens: 25}   // 100/4 = 25
s2: {Bytes: 200, Tokens: 50}   // 200/4 = 50
s1.Add(s2): {Bytes: 300, Tokens: 75}  // 25+50 = 75
EstimateFromBytes(300): 75       // 300/4 = 75 (mesmo resultado neste caso)
```
O resultado é idêntico quando ambas as quantidades são divisíveis por 4, mas pode divergir em casos de restolhos.

#### 4.1.6 `FormatTokens(tokens int) string`

Formatação humana com sufixos K/M:
| Faixa | Formato | Exemplo |
|---|---|---|
| 0–999 | `%d` | `"42"`, `"999"` |
| 1,000–999,999 | `"%.1fK"` | `"1.0K"`, `"10.0K"`, `"100.0K"` |
| ≥ 1,000,000 | `"%.1fM"` | `"1.0M"`, `"1.5M"` |

Nenhuma separação de milhares (ex: `"10,000"` não ocorre). A formatação é consistente e previsível.

#### 4.1.7 `ContextFit` e `CheckContextFit()`

Estrutura que descreve como um conjunto de tokens cabe dentro de uma janela de contexto:
| Campo | Tipo | Descrição |
|---|---|---|
| `WindowSize` | `int` | Tamanho alvo em tokens |
| `UsedTokens` | `int` | Tokens utilizados |
| `Percentage` | `float64` | Porcentagem da janela usada (0–100+) |
| `Fits` | `bool` | `true` se `UsedTokens <= WindowSize` |

**Lógica:**
```go
percentage = (tokens / windowSize) * 100  // zero se windowSize == 0
fits = (tokens <= windowSize)
```

**Caso especial — windowSize = 0:**
- `Percentage = 0.0` (divisão por zero é evitada pelo `if windowSize > 0`)
- `Fits = false` (tokens > 0 não cabe em janela de tamanho 0)

---

## 5. Dependentes do Pacote (Consumidores)

| Pacote Consumidor | Função Chamada | Contexto de Uso |
|---|---|---|
| `cmd/context.go` | `FormatTokens()` | Exibe estimativa de tokens no sumário de geração de contexto |
| `internal/app/service.go` | `EstimateFromBytes()` | Calcula `TokenEstimate` no resultado de `Generate()` |
| `internal/ui/components/progress.go` | `FormatTokens()` | Exibe tokens atuais na barra de progresso |
| `internal/ui/screens/file_selection.go` | `EstimateFromBytes()`, `FormatTokens()` | Exibe tokens estimados e formatados na tela de seleção de arquivos |
| `internal/ui/screens/review.go` | `EstimateFromBytes()`, `FormatTokens()` | Exibe tokens estimados na tela de revisão antes do envio |

**Padrão de uso:** O pacote é usado exclusivamente para **visualização/estimativa**, nunca para tomada de decisão lógica. Nenhum fluxo crítico depende exclusivamente dos resultados de `tokens`.

---

## 6. Fluxos Internos

### 6.1 Fluxo de Estimativa de Tokens

```
Texto → len(text) [bytes] → EstimateFromBytes() → ceil(bytes/4) → int
```

### 6.2 Fluxo de Estatísticas (Stats)

```
bytes → NewStats() → {Bytes: bytes, Tokens: EstimateFromBytes(bytes)}
text → NewStatsFromText() → NewStats(len(text))
stats1.Add(stats2) → {Bytes: s1.Bytes+s2.Bytes, Tokens: s1.Tokens+s2.Tokens}
```

### 6.3 Fluxo de Verificação de Contexto

```
tokens, windowSize → CheckContextFit()
  ├─ windowSize == 0 → percentage=0.0, fits=false (se tokens>0)
  ├─ percentage = (tokens/windowSize)*100
  └─ fits = (tokens <= windowSize)
  → ContextFit{WindowSize, UsedTokens, Percentage, Fits}
```

---

## 7. Padrões Arquiteturais Identificados

### 7.1 Heurística Sem Dependências Externas
O pacote implementa uma heurística simplificada (`1 token ≈ 4 bytes`) em vez de usar uma biblioteca de tokenizer real. Esta é uma escolha **intencional de design**: evitar dependências pesadas (`github.com/xtekky/gotokenizer`, etc.) em um módulo de utilidade central.

### 7.2 Funções Puras
Todas as funções são puras — sem estado global, sem efeitos colaterais, sem mutação. Isso as torna facilmente testáveis e previsíveis.

### 7.3 Operações com Int64/Int
- `EstimateFromBytes` aceita `int64` para suportar arquivos grandes.
- `Estimate` e `FormatTokens` usam `int`.
- `BytesFromTokens` retorna `int64`.
- `Stats.Bytes` é `int64`, `Stats.Tokens` é `int`.

**🟡 INFERIDO:** A mistura de `int` e `int64` é provavelmente para compatibilidade com APIs existentes. Em ambientes 64-bit, não há diferença prática, mas em teorias de overflow, `int` em sistemas 32-bit poderia ser problema para valores > 2B tokens.

### 7.4 Stats como Agregador Idempotente
`Stats.Add()` permite combinar estatísticas de partes distintas. O método é comutativo e associativo, facilitando agregação distribuída de contagens.

---

## 8. Qualidade de Código

### 8.1 Pontos Fortes

| Aspecto | Descrição |
|---|---|
| **Simplicidade** | Arquivo único, 80 linhas, sem dependências externas |
| **Pureza** | Todas as funções são puras — previsíveis e testáveis |
| **Testes** | 6 funções testadas, cobrindo todos os caminhos de execução, incluindo casos de borda (zero, negativo, divisões exatas e não-exatas) |
| **Performance** | O(1) — sem iterações, sem alocações, apenas aritmética |
| **Documentação** | Doc comments em todas as funções públicas, comentários inline explicando a heurística |

### 8.2 Pontos de Atenção

| Issue | Severidade | Descrição |
|---|---|---|
| Heurística de 4 bytes/token é imprecisa | Média | Textos em outros idiomas, código fonte, e textos com muitos espaços serão estimados de forma diferente da realidade. A heurística é reconhecida como aproximação. |
| `BytesFromTokens` não é inversa exata | Baixa | Devido ao ceil, `BytesFromTokens(EstimateFromBytes(n)) ≥ n` sempre. A diferença máxima é 3 bytes. |
| `Stats.Add()` não recalcula tokens | Baixa | Soma direta de tokens pode diferir de `EstimateFromBytes(s1.Bytes + s2.Bytes)` em casos de restolho. Diferença máxima: 2 tokens. |
| Sem suporte a encoding-aware estimation | Baixa | `len(text)` conta bytes UTF-8, não caracteres Unicode. Para textos em japonês/chinoês, a estimativa pode ser 2-3x maior que o real. |
| Constantes `Window*` não usadas internamente | Informativo | São apenas documentação de referência. Poderiam ser úteis como parâmetros de `CheckContextFit` (ex: `CheckContextFit4K()`). |

---

## 9. Gap Analysis

| Gap | Descrição | Impacto |
|---|---|---|
| Nenhuma biblioteca tokenizer real | Estimativa por heurística 4 bytes/token — pode ter erro de 20-40% em textos não-ASCII | Médio — para estimativas de custo LLM, um erro de ~25% é aceitável; para cortes precisos de contexto, não é suficiente |
| Sem `Stats.Subtraction` | Não há forma de subtrair estatísticas (útil para windowing/sliding window) | Baixo — uso raro |
| Sem `ContextFit.RemainingTokens` | Calcula percentage mas não retorna tokens restantes | Baixo — consumidores calculam `WindowSize - UsedTokens` externamente |
| Sem configuração de BytesPerToken | Heurística fixa em 4 bytes/token, sem ajuste por provider ou idioma | Médio — OpenAI, Anthropic e Gemini têm tokenizers ligeiramente diferentes |
| Sem `EstimateFromRune` ou `EstimateFromRuneCount` | Não há distinção entre bytes e caracteres Unicode | Baixo — uso específico para idiomas não-ASCII |
