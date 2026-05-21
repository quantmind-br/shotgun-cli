# Fluxograma: Função `Copy`

| Item | Detalhe |
|------|---------|
| **Módulo** | `internal/platform/clipboard` |
| **Arquivo** | `clipboard.go:27-33` |
| **Fluxo** | Copiar conteúdo para o clipboard do sistema operacional |

---

## 1. Fluxo Principal `Copy(content string) error`

```mermaid
flowchart TD
    Start(["Início: Copy(content)"]) --> Arg["Receber content (string)"]

    Arg --> Call["clipboard.WriteAll(content)\n[da biblioteca atotto/clipboard]"]

    Call --> Check{"WriteAll retornou\nerror?"}

    Check -->|nil (sucesso)| Success["Retornar nil"]
    Check -->|err (falha)| Wrap["Criar &ClipboardError{Err: err}"]

    Wrap --> RetErr["Retornar *ClipboardError"]

    Success --> EndOK(["Fim: nil"])
    RetErr --> EndErr(["Fim: *ClipboardError"])

    classDef success fill:#90EE90,stroke:#333,stroke-width:2px
    classDef error fill:#FFB6C1,stroke:#333,stroke-width:2px
    classDef process fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef terminal fill:#e8f5e9,stroke:#388e3c

    class Success,EndOK success
    class RetErr,EndErr error
    class Arg,Call,Wrap process
    class Start terminal
```

---

## 2. Detalhamento Passo a Passo

### Passo 1: Receber Parâmetro

- **Entrada:** `content string` — qualquer string válida
- **Restrições:** Nenhuma — `WriteAll` aceita strings arbitrárias
- **Tipos suportados:** ASCII, UTF-8 (unicode), multiline, vazia

### Passo 2: Chamar `clipboard.WriteAll(content)`

- **Origem:** `github.com/atotto/clipboard` (third-party)
- **Implementação:** Pluralizada por plataforma:
  - **Linux:** `xclip` ou `wl-copy` (Wayland)
  - **macOS:** `pbcopy` (execução de subprocess)
  - **Windows:** `OpenClipboard` + `GlobalAlloc` (Win32 API)
- **Thread-safety:** `atotto/clipboard` usa locking interno por plataforma
- **Bloqueio:** Sincrono — bloqueia até o lock do clipboard ser liberado

### Passo 3: Verificar Retorno

- **Sucesso (`nil`):** Operação completa — conteúdo visível no clipboard
- **Erro (`err`):** Falha durante a operação

### Passo 4: Retornar Resultado

- **Sucesso:** `return nil` — caller verifica `err == nil`
- **Erro:** `return &ClipboardError{Err: err}` — wrapper typed

---

## 3. Caminhos de Erro

| Erro | Causa Provável | Tipo de `err` | Exemplo |
|------|----------------|---------------|---------|
| Clipboard já em uso | Outro processo escreveu ao clipboard | `error` do sistema operacional | `"clipboard locked"` |
| Sem display server (Linux) | Headless / SSH | `*errors.errorString` | `"X11 not available"` |
| Subprocesso falhou (macOS) | `pbcopy` não encontrado | `exec.Error` | `"pbcopy: command not found"` |
| Permissão negada | Sandbox / AppArmor | `*os.PathError` | `"permission denied"` |

**Observação:** O `ClipboardError` **não altera** o erro original — apenas o empacota. O caller pode usar `errors.As(err, &originalErr)` para identificar o tipo exato.

---

## 4. Casos de Uso no Projeto

### 4.1 Via `app.GenerateWithProgress`

```go
// service.go:149 — dentro de GenerateWithProgress
if cfg.CopyToClipboard {
    err := clipboard.Copy(content)
    if err != nil {
        log.Printf("failed to copy to clipboard: %v", err)
    } else {
        result.CopiedToClipboard = true
    }
}
```

**Características:**
- Chamada única — não há loop
- Erro é logged mas não bloqueia o fluxo (non-fatal)
- `result.CopiedToClipboard` reflete sucesso/falha

### 4.2 Gate Check com `IsAvailable()`

```go
if clipboard.IsAvailable() {
    clipboard.Copy(content)
}
```

**Padrão:** Verificar disponibilidade antes de tentar copiar, especialmente em ambientes CI/CD.

---

## 5. Complexidade

| Métrica | Valor |
|---------|-------|
| Cyclomatic Complexity | 2 (1 condição: `if err != nil`) |
| Linhas de código | 7 |
| Chamadas externas | 1 (`WriteAll`) |
| Alocações de memória | 1 (`&ClipboardError{}`) em caso de erro |
| Goroutines | 0 |
| Sincronização | 0 (delegado a `atotto/clipboard`) |
