# Dicionário de Dados — Módulo `internal/platform/clipboard`

**Nível de detalhe:** detalhado
**Pacote:** `github.com/quantmind-br/shotgun-cli/internal/platform/clipboard`
**Total de tipos documentados:** 2

---

## 1. Structs (Tipos de Erro)

### 1.1 `ClipboardError`

**Localização:** `clipboard.go:14-16`
**Exportada:** ✅
**Finalidade:** Tipo de erro específico para operações de clipboard. Permite diferenciação de erros de clipboard vs erros de sistema usando `errors.As()`.

| Campo | Tipo | Exportado | Valor Padrão | Obrigatório | Descrição |
|-------|------|-----------|-------------|-------------|-----------|
| `Err` | `error` | ❌ | `nil` | ✅ | Erro subjacente retornado por `atotto/clipboard.WriteAll()` |

**Métodos:**

| Método | Recebedor | Parâmetros | Retorno | Descrição |
|--------|-----------|------------|---------|-----------|
| `Error()` | `*ClipboardError` | — | `string` | Retorna `"clipboard error: <err>"` — satisfaz interface `error` |
| `Unwrap()` | `*ClipboardError` | — | `error` | Retorna `e.Err` — permite `errors.Is()`/`errors.As()` na cadeia de erros (Go 1.20+) |

**Uso típico:**
```go
err := clipboard.Copy("text")
if err != nil {
    var clipboardErr *clipboarderrorerror
    if errors.As(err, &clipboardErr) {
        // trata erro específico de clipboard
    }
}
```

---

## 2. Funções (Sem estado)

> **Nota:** O package `clipboard` não define structs de dados de domínio. Todas as operações são **stateless functions** — não há structs de configuração, DTO, ou resultado.

### 2.1 `Copy(content string) error`

**Localização:** `clipboard.go:27-33`
**Exportada:** ✅
**Tipo:** Função global (não-método)
**Finalidade:** Copiar texto ao clipboard do sistema operacional

| Parâmetro | Tipo | Exportado | Obrigatório | Descrição |
|-----------|------|-----------|-------------|-----------|
| `content` | `string` | ✅ | ✅ | Conteúdo a copiar. Aceita qualquer string válida (incluindo unicode, multiline, vazia) |

**Retorno:**

| Valor | Tipo | Descrição |
|-------|------|-----------|
| `nil` | `error` | Operação bem-sucedida |
| `*ClipboardError` | `error` | Erro durante a escrita ao clipboard |

**Caminhos:**

| Caminho | Condição | Retorno |
|---------|----------|---------|
| Sucesso | `clipboard.WriteAll(content)` retorna `nil` | `nil` |
| Erro | `clipboard.WriteAll(content)` retorna `err` | `&ClipboardError{Err: err}` |

**Limitações:**
- Sincrono — bloqueia até que `atotto/clipboard` libere o lock
- Sem buffer interno — chamadas consecutivas são independentes
- Não verifica `IsAvailable()` — o caller deve fazer gate check se necessário

---

### 2.2 `IsAvailable() bool`

**Localização:** `clipboard.go:36-38`
**Exportada:** ✅
**Tipo:** Função global (não-método)
**Finalidade:** Verificar se o sistema suporta operações de clipboard

**Retorno:**

| Valor | Tipo | Descrição |
|-------|------|-----------|
| `true` | `bool` | Clipboard suportado (display server disponível) |
| `false` | `bool` | Clipboard não suportado (headless, CI, sandbox) |

**Implementação:**
```go
return !clipboard.Unsupported
```

**Comportamento de detecção (via `atotto/clipboard`):**

| Plataforma | Critério | Resultado se não atendido |
|------------|----------|--------------------------|
| Linux | `X11` ou `Wayland` disponível | `Unsupported = true` |
| Windows | Windows API acessível | `Unsupported = true` (improvável) |
| macOS | `NSPasteboard` disponível | `Unsupported = true` (improvável) |
| Headless/CI | Sem display server | `Unsupported = true` |

**Uso típico:**
```go
if !clipboard.IsAvailable() {
    fmt.Println("Clipboard não disponível neste ambiente")
    os.Exit(1)
}
```

---

## 3. Enums / Constantes

| Enum | Valor | Descrição |
|------|-------|-----------|
| N/A | — | Nenhum enum ou constante definido no package |

---

## 4. Tipos Não-Exportados (Internos)

| Tipo | Exportada | Descrição |
|------|-----------|-----------|
| N/A | ❌ | Nenhum tipo interno — package é minimalista (3 exports, 0 internals) |

---

## 5. Fluxo de Dados

### 5.1 Fluxo de `Copy()`

```
string (content)
    ↓ [input]
clipboard.Copy(content)
    ↓
atotto/clipboard.WriteAll(content)
    ├── ✅ → nil ──→ return nil
    └── ❌ → err ──→ return &ClipboardError{Err: err}
    ↓
error (nil ou *ClipboardError)
```

### 5.2 Fluxo de `IsAvailable()`

```
[system state]
    ↓ [X11/Wayland/Windows/macOS]
atotto/clipboard.Unsupported (bool, set in init)
    ↓ [inversão]
!clipboard.Unsupported
    ↓
bool (true ou false)
```

---

## 6. Relação entre Tipos do Package `clipboard` e Tipos Externos

### 6.1 `ClipboardError` → `error`

| Tipo em `clipboard` | Tipo em `stdlib` | Relação |
|---------------------|------------------|---------|
| `*ClipboardError` | `error` | Satisfaz interface via `Error()` |
| `*ClipboardError.Unwrap()` | `error` (Go 1.20+) | Cadeia de erros via `errors.Is()`/`errors.As()` |

### 6.2 `Copy()` → `atotto/clipboard.WriteAll()`

| Tipo em `clipboard` | Tipo em `atotto/clipboard` | Relação |
|---------------------|---------------------------|---------|
| `Copy(content string)` | `WriteAll(text string) error` | Wrapper direto — 1:1 |
| `&ClipboardError{Err: err}` | `err` do `WriteAll()` | Error wrapping |

---

## 7. Notas sobre Ausência de Tipos de Dados

O package `clipboard` **não define**:
- ❌ Structs de configuração
- ❌ Structs de resultado (DTOs)
- ❌ Enums ou constantes
- ❌ Interfaces
- ❌ Tipos não-exportados

Esta é uma escolha de design intencional: o package é um **utility module** (pure functions) em vez de um **service module** (stateful types). Isso é consistente com a filosofia de plataformas no shotgun-cli: `platform/` fornece funções simples, enquanto `app/` orquestra os serviços com structs complexos.
