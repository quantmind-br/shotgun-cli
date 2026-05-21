# Análise de Código — Módulo `internal/platform/clipboard`

**Nível de detalhe:** detalhado
**Caminho do pacote:** `github.com/quantmind-br/shotgun-cli/internal/platform/clipboard`
**Número de arquivos de código:** 2 (1 fonte + 1 teste)
**Dependências externas:** 1 (`github.com/atotto/clipboard`)
**Dependências internas:** 0

---

## 1. Visão Geral Arquitetural

O módulo `internal/platform/clipboard` é uma **abstração de plataforma** que fornece operações de clipboard cross-platform. É a menor unidade de código dentro do package `internal/platform/` e funciona como uma **wrapper thin** sobre a biblioteca `github.com/atotto/clipboard`.

### Responsabilidades

1. **Copy:** copiar texto para o clipboard do sistema operacional
2. **IsAvailable:** verificar se o sistema suporta operações de clipboard (ex.: desktop com X11/Wayland/Windows, não headless/CI)
3. **Erros typed:** retornar `*ClipboardError` para diferenciação de erros de clipboard vs erros de sistema

### Regras de Fronteira

- ❌ **Não importa** nenhum package interno do projeto — é completamente isolado
- ❌ **Não faz caching** de estado — cada chamada é independente
- ❌ **Não é thread-safe por design** — delega thread-safety ao `atotto/clipboard`
- ✅ **Não exporta variáveis globais** — tudo é função ou método

---

## 2. Inventário de Arquivos

### 2.1 Arquivos de Produção

| Arquivo | Linhas | Pacote | Função |
|---------|--------|--------|--------|
| `clipboard.go` | ~40 | `clipboard` | Definição de `ClipboardError`, funções `Copy()` e `IsAvailable()` |

### 2.2 Arquivos de Teste

| Arquivo | Linhas | Função |
|---------|--------|--------|
| `clipboard_test.go` | ~75 | Testes unitários de `ClipboardError`, `IsAvailable()`, `Copy()` (com dados reais), `BenchmarkCopy` |

---

## 3. Análise Detalhada por Arquivo

### 3.1 `clipboard.go` — Wrap do atotto/clipboard

**Funções/Tipos exportados:** 4

| Tipo | Exportado | Descrição |
|------|-----------|-----------|
| `ClipboardError` (struct) | ✅ | Error typed para erros de clipboard |
| `ClipboardError.Error()` | ✅ | Formatação: `"clipboard error: <erro>"` |
| `ClipboardError.Unwrap()` | ✅ | Retorna erro subjacente (implements `error` interface de wrapping) |
| `Copy(content string)` | ✅ | Copia conteúdo ao clipboard |
| `IsAvailable()` | ✅ | Verifica suporte de clipboard |

**Arquitetura interna:**

```
┌─────────────────────┐
│  shotgun-cli        │
│  (user code)        │
├─────────────────────┤
│  clipboard.Copy()   │  ───►  atotto/clipboard.WriteAll()
│  clipboard.IsAvail. │  ───►  atotto/clipboard.Unsupported
├─────────────────────┤
│  atotto/clipboard   │
│  (third-party)      │
│  ┌───┬───┬───┐      │
│  │Win│Linux│Mac│     │
│  └───┴───┴───┘      │
└─────────────────────┘
```

**Dependências de import:**

| Import | Origem | Uso |
|--------|--------|-----|
| `fmt` | stdlib | `fmt.Sprintf` para formatação de erro |
| `github.com/atotto/clipboard` | externo | `WriteAll()`, `Unsupported` |

**Análise da função `Copy`:**

```go
func Copy(content string) error {
    if err := clipboard.WriteAll(content); err != nil {
        return &ClipboardError{Err: err}
    }
    return nil
}
```

- **Padrão:** Wrapper idiomático com error wrapping (`&ClipboardError{Err: err}`)
- **Go 1.20+:** O `Unwrap()` permite `errors.As()` e `errors.Is()` no erro original
- **Imutabilidade:** `content` é passado por valor (string é imutável em Go), sem cópia explícita

**Análise da função `IsAvailable`:**

```go
func IsAvailable() bool {
    return !clipboard.Unsupported
}
```

- **Padrão:** Inversão de boolean — `Unsupported` é `bool` definido pelo `atotto/clipboard` durante init
- **Detecção:** A biblioteca verifica X11/Wayland no Linux, Windows API no Windows, e NSClipBoard no macOS
- **Valores falsos:** Em CI/CD headless, servidores sem display, ou ambientes sandboxed
- **Sem caching:** `clipboard.Unsupported` é setado em `init()` da biblioteca, mas a função lê o valor diretamente — não há memoização local

### 3.2 `clipboard_test.go` — Cobertura de Testes

**Total de testes:** 4 (3 unitários + 1 benchmark)

| Teste | Tipo | Cobertura |
|-------|------|-----------|
| `TestClipboardErrorFormat` | Unitário | `ClipboardError.Error()` retorna formato `"clipboard error: <nil>"` |
| `TestClipboardErrorUnwrap` | Unitário | `Unwrap()` retorna o erro original (`same identity` via `==`) |
| `TestIsAvailable` | Unitário | Verifica que `IsAvailable()` não panics |
| `BenchmarkCopy` | Benchmark | Mede throughput de `Copy()` com conteúdo pequeno |
| `TestCopySuccess` | Unitário (dados) | 6 subtests de `Copy()` com diferentes tipos de conteúdo |

**Dados de teste:**

| Caso | Conteúdo | Tamanho |
|------|----------|---------|
| simple text | `"hello world"` | 11 bytes |
| empty string | `""` | 0 bytes |
| unicode | `"こんにちは世界"` | 15 bytes (UTF-8) |
| multiline | `"line1\nline2\nline3"` | 17 bytes |
| special chars | `"tab\there\nnewline"` | 17 bytes |
| long text | `str.Repeat("x", 10000)` | 10000 bytes |

**Observações sobre o teste:**

- `TestCopySuccess` **executa operações reais de clipboard** — não é mockado
- `t.Skip()` se `IsAvailable()` retornar `false` (headless/CI)
- O benchmark usa `b.ResetTimer()` correto, pula em headless
- **Nenhum teste simula erro de `WriteAll()`** — não há forma de mockar sem mudar o package under test
- 🟡 **INFERIDO:** A biblioteca `atotto/clipboard` não expõe um interface — é uma implementação concreta. Para testar falhas de clipboard seria necessário um fork ou uma interface abstrata.

---

## 4. Dependências de Pacote

### 4.1 Dependências Diretas

| Dependência | Origem | Uso |
|-------------|--------|-----|
| `fmt` | stdlib | Formatação de erro |
| `github.com/atotto/clipboard` | externo (`go.mod`) | `WriteAll()`, `Unsupported` |

### 4.2 Dependências Transitivas (via atotto/clipboard)

A biblioteca `atotto/clipboard` depende de packages internos do projeto:

| Dependência | Origem | Uso |
|-------------|--------|-----|
| `internal/app` | interno | `app.CopyToClipboard()` chama `clipboard.Copy()` |

### 4.3 Matriz de Dependências

```
internal/platform/clipboard
├── (externo) github.com/atotto/clipboard
│   ├── (externo) golang.org/x/sys (platform detection)
│   └── (externo) syscall (platform detection)
└── (interno) internal/app
    └── (uso) Copy(content string) error
```

**Consumidores identificados:**

| Consumidor | Pacote | Uso | Linha/Arquivo |
|------------|--------|-----|---------------|
| `DefaultContextService.GenerateWithProgress` | `app` | `clipboard.Copy(content)` após salvar arquivo | `service.go:149` |

**🟡 INFERIDO:** O consumo do package `clipboard` é exclusivamente via `app.GenerateWithProgress`. Não há outras referências diretas encontradas no código-fonte. Poderia haver imports indiretos via `cmd` que delega para `app`.

---

## 5. Padrões de Design Identificados

### 5.1 Thin Wrapper Pattern

`clipboard.go` é um **wrapper mínimo** — apenas uma função e uma verificação. Não adiciona lógica de negócio, apenas adiciona tipagem de erro.

**Motivação:**
- Uniformiza erros de clipboard no código do projeto
- Permite substituir a implementação subjacente sem alterar callers
- Facilita testes: `ClipboardError` pode ser esperado por tipo em testes do `app`

### 5.2 Error Wrapping

```go
type ClipboardError struct {
    Err error
}
```

- Implementa `Error() string` — satisfaz `error` interface
- Implementa `Unwrap() error` — permite `errors.Is()`/`errors.As()` no erro original
- **Go 1.13+ error wrapping** — padrão moderno

### 5.3 Platform Abstraction

A função `IsAvailable()` fornece um gate check antes de operações de clipboard. Padrão comum em libs de plataforma:

```go
if !clipboard.IsAvailable() {
    // fallback or error
}
```

---

## 6. Cobertura de Testes

### 6.1 Resumo

| Arquivo de Teste | Casos de Teste | Funções Cobertas |
|------------------|----------------|------------------|
| `clipboard_test.go` | 4 + 6 subtests | `Copy()`, `IsAvailable()`, `ClipboardError.Error()`, `ClipboardError.Unwrap()` |
| `clipboard_test.go` (bench) | 1 benchmark | `BenchmarkCopy` |

**Total:** 10 casos de teste/benchmark

### 6.2 Cobertura por Caminho

| Caminho de Execução | Cobertura |
|---------------------|-----------|
| `Copy()` sucesso (simple text) | ✅ |
| `Copy()` sucesso (empty string) | ✅ |
| `Copy()` sucesso (unicode) | ✅ |
| `Copy()` sucesso (multiline) | ✅ |
| `Copy()` sucesso (special chars) | ✅ |
| `Copy()` sucesso (long text) | ✅ |
| `Copy()` sucesso (skipped) | ✅ (headless) |
| `IsAvailable()` sem panic | ✅ |
| `ClipboardError.Error()` formato | ✅ |
| `ClipboardError.Unwrap()` identidade | ✅ |
| `ClipboardError.Error()` com erro não-nil | ❌ |
| `Copy()` com erro de atotto/clipboard | ❌ |

**Caminhos não testados:**
- 🟡 **Erro de clipboard não pode ser testado** — `atotto/clipboard` não expõe interface mockable
- 🟡 **`ClipboardError.Error()` com erro não-nil** — teste existe apenas com `Err: nil`
- 🟡 **Concorrência** — não há teste de chamada concorrente de `Copy()` (embora `atotto/clipboard` seja thread-safe)

### 6.3 Limitações de Teste

O principal desafio é que `atotto/clipboard` é uma **implementação concreta** e não uma interface:

```go
// atotto/clipboard (sem interface)
func WriteAll(text string) error  // função concreta
var Unsupported bool              // variável global
```

**Alternativas para melhor testabilidade (não implementadas):**
1. Definir interface local: `type Clipboard interface { WriteAll(string) error }`
2. Usar `testify/mock` com interface — requer refatoração do `clipboard.go`
3. Usar variável de função: `var WriteAllFunc = clipboard.WriteAll`

---

## 7. Questões Técnicas e Debt

### 7.1 Sem Interface para Testabilidade

**Problema:** `clipboard.go` chama `clipboard.WriteAll()` diretamente — não há como mockar em testes.

**Impacto:** Erros de clipboard não podem ser testados em unit tests.

**Sugestão:** Definir uma interface local:
```go
type clipboardOps interface {
    WriteAll(string) error
}
```

### 7.2 `ClipboardError` com `Err: nil`

O teste `TestClipboardErrorFormat` usa `&ClipboardError{Err: nil}`, resultando em `"clipboard error: <nil>"`. Isso é **estético mas não funcional** — um `ClipboardError` com `Err: nil` é um anti-pattern (erros nil não devem ser retornados).

**Observação:** A função `Copy()` nunca retorna `&ClipboardError{Err: nil}` — o erro só é criado quando `WriteAll()` retorna um erro não-nil. O teste é puramente de formato de string.

### 7.3 `IsAvailable()` Sem Caching

`IsAvailable()` lê `clipboard.Unsupported` diretamente a cada chamada. Embora a biblioteca interna cache esse valor em `init()`, uma chamada direta à variável é menos explícita.

**Impacto:** Mínimo — apenas uma leitura de bool, mas sem memoização local o código não é auto-documentado sobre a intenção de cache.

### 7.4 Sem Buffer de Clipboard

A operação `Copy()` é **síncrona e sem buffer**. Se o caller chamar `Copy()` múltiplas vezes rapidamente, cada chamada bloqueia até que o clipboard libere o lock do sistema.

**🟡 INFERIDO:** `atotto/clipboard` tem locking interno por plataforma, mas não há semáforo ou debounce no wrapper. Em workflows de geração de contexto, isso é aceitável (uma única chamada), mas não em loops.

---

## 8. Métricas de Código

| Métrica | Valor |
|---------|-------|
| Linhas de código (excluindo testes) | ~40 |
| Linhas de código (testes) | ~75 |
| Ratio teste:código | ~1.88x |
| Tipos definidos | 1 (struct `ClipboardError`) |
| Constantes | 0 |
| Variáveis globais | 0 |
| Funções init | 0 |
| Funções não-método | 2 (`Copy`, `IsAvailable`) |
| Métodos total | 2 (`Error()`, `Unwrap()`) |
| Importações stdlib | 1 (`fmt`) |
| Importações internas | 0 |
| Importações externas de terceiros | 1 (`atotto/clipboard`) |
| Cyclomatic Complexity (Copy) | 1 |
| Cyclomatic Complexity (IsAvailable) | 1 |

---

## 9. Resumo Executivo

O módulo `internal/platform/clipboard` é a **menor abstração de plataforma** do projeto. É um wrapper de ~40 linhas sobre `atotto/clipboard`, fornecendo:

1. **`Copy(content)`** — Copiar texto ao clipboard do sistema
2. **`IsAvailable()`** — Verificar se clipboard é suportado

**Pontos fortes:**
- Código minimalista e fácil de entender
- Error wrapping idiomático (`ClipboardError` com `Unwrap()`)
- Testes com dados reais (não mockados) — valida funcionalidade real
- Zero dependências internas — completamente isolado

**Pontos de atenção:**
- 🟡 **Sem interface** para o `atotto/clipboard` — não pode ser mockado
- 🟡 **Erro de clipboard não testável** em unit tests
- 🟡 **Sem buffer/debounce** para chamadas rápidas de `Copy()`
- 🟡 **`IsAvailable()` sem memoização local** — depende do cache da biblioteca third-party

**Posição na arquitetura:**
```
cmd → app.GenerateWithProgress() → clipboard.Copy(content)
                                    ↘→ clipboard.IsAvailable() (gate check)
```
