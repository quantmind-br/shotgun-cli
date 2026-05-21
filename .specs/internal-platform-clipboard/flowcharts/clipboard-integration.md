# Fluxograma: Integração do Clipboard no Fluxo `GenerateWithProgress`

| Item | Detalhe |
|------|---------|
| **Módulo** | `internal/platform/clipboard` + `internal/app` |
| **Arquivos** | `clipboard.go` + `service.go:140-154` |
| **Fluxo** | Como o clipboard é invocado durante a geração de contexto |

---

## 1. Contexto

A função `clipboard.Copy()` é chamada **dentro de** `app.DefaultContextService.GenerateWithProgress()`, exclusivamente quando a opção `CopyToClipboard` está habilitada.

**Localização exata:** `service.go:140-154` (dentro do workflow de geração)

---

## 2. Fluxograma Mermaid

```mermaid
flowchart TD
    Start([Início: GenerateWithProgress]) --> Validate["cfg.Validate()\n- rootPath vazio?\n- rootPath existe?\n- rootPath é diretório?"]

    Validate -->|Erro| ErrCfg["Erro: invalid config"]
    Validate -->|OK| Scan["scanner.Scan()\n→ FileNode tree"]

    Scan -->|Erro| ErrScan["Erro: scan failed"]
    Scan -->|OK| Select["scanner.NewSelectAll(tree)\n→ selections map"]

    Select --> Gen["generator.Generate()\n→ string content"]
    Gen -->|Erro| ErrGen["Erro: generation failed"]
    Gen -->|OK| CheckLimit["cfg.EnforceLimit?\ncontentSize > cfg.MaxSize?"]

    CheckLimit -->|Sim| ErrLimit["Erro: content exceeds limit"]
    CheckLimit -->|Não| WriteFile["os.WriteFile(\n    outputPath,\n    []byte(content),\n    0600)"] --> WriteErr{"Save erro?"}

    WriteErr -->|Sim| ErrSave["Erro: failed to save"]
    WriteErr -->|Não| CheckClipboard{"cfg.CopyToClipboard?"}

    CheckClipboard -->|Não| BuildResult

    CheckClipboard -->|Sim| IsAvail["clipboard.IsAvailable()"]
    IsAvail -->|false| LogWarn["log.Printf: clipboard not available"]
    LogWarn --> SetCopiedFalse["result.CopiedToClipboard = false"]

    IsAvail -->|true| DoCopy["clipboard.Copy(content)"]
    DoCopy --> CopyErr{"Copy erro?"}

    CopyErr -->|Sim| LogCopyErr["log.Printf: copy failed"]
    LogCopyErr --> SetCopiedFalse

    CopyErr -->|Não| SetCopiedTrue["result.CopiedToClipboard = true"]

    SetCopiedFalse --> BuildResult
    SetCopiedTrue --> BuildResult

    BuildResult["Resultado:\nGenerateResult{\n    Content,\n    OutputPath,\n    FileCount,\n    ContentSize,\n    TokenEstimate,\n    CopiedToClipboard}"] --> EndSuccess([Fim: Success])

    ErrCfg --> End1([Fim: Error])
    ErrScan --> End2([Fim: Error])
    ErrGen --> End3([Fim: Error])
    ErrLimit --> End4([Fim: Error])
    ErrSave --> End5([Fim: Error])

    classDef success fill:#90EE90,stroke:#333,stroke-width:2px
    classDef error fill:#FFB6C1,stroke:#333,stroke-width:2px
    classDef process fill:#87CEEB,stroke:#333,stroke-width:2px
    classDef decision fill:#FFD700,stroke:#333,stroke-width:2px
    classDef clipboard fill:#DDA0DD,stroke:#6A0DAD,stroke-width:2px
    classDef log fill:#F5F5DC,stroke:#8B8B00,stroke-width:1px

    class BuildResult,EndSuccess success
    class ErrCfg,ErrScan,ErrGen,ErrLimit,ErrSave error
    class Validate,Scan,Select,Gen,CheckLimit,WriteFile,DoCopy process
    class CheckClipboard,CheckLimit,WriteErr,CopyErr,IsAvail decision
    class IsAvail,DoCopy clipboard
    class LogWarn,LogCopyErr log
```

---

## 3. Detalhamento Passo a Passo

### Etapa A: Gate Check (`IsAvailable`)

```go
// service.go — antes de Copy(), o caller NÃO verifica IsAvailable()
// A verificação é feita DENTRO de Copy() ou não é feita — Copy() tenta direto
```

**Observação importante:** `GenerateWithProgress` **não chama explicitamente** `IsAvailable()` antes de `Copy()`. A função `Copy()` é chamada diretamente. Se `IsAvailable()` retornasse `false`, `WriteAll()` falharia, e o erro seria capturado pelo `if err != nil` em `Copy()`.

### Etapa B: Execução (`Copy`)

```go
err := clipboard.Copy(content)
```

- Chamada síncrona
- Bloqueia até o lock do clipboard ser liberado
- Retorna `nil` (sucesso) ou `*ClipboardError` (falha)

### Etapa C: Tratamento (Non-Fatal)

```go
if err != nil {
    log.Printf("failed to copy to clipboard: %v", err)
} else {
    result.CopiedToClipboard = true
}
```

**Características:**
- Erro é **logged mas não retornado** — non-fatal
- O fluxo de geração continua mesmo se clipboard falhar
- `result.CopiedToClipboard` reflete o resultado exato
- **Não há retry** — uma única tentativa

### Etapa D: Resultado

```go
result := GenerateResult{
    Content:             content,
    OutputPath:          outputPath,
    FileCount:           tree.CountFiles(),
    ContentSize:         int64(len(content)),
    TokenEstimate:       tokens.EstimateFromBytes(contentSize),
    CopiedToClipboard:   !copyErr,
}
```

---

## 4. Fluxo Alternativo: `CopyToClipboard = false`

```
GenerateWithProgress
    ↓
... (geração e salvamento do arquivo)
    ↓
Check cfg.CopyToClipboard
    ↓
false → Pular completamente o bloco clipboard
    ↓
BuildResult (CopiedToClipboard = false por padrão)
    ↓
End Success
```

Neste caso, o package `clipboard` **não é invocado** — a biblioteca `atotto/clipboard` nem é carregada.

---

## 5. Propagação do Erro

| Nível | Comportamento |
|-------|---------------|
| `atotto/clipboard` | Retorna `error` nativo do SO |
| `clipboard.Copy()` | Empacota em `*ClipboardError` |
| `app.GenerateWithProgress()` | Logs e ignora — não propaga ao caller |
| Caller final | Nunca recebe erro de clipboard |

**🟡 INFERIDO:** Esta é uma escolha de design intencional. O caller (CLI/TUI) recebe apenas `GenerateResult.CopiedToClipboard` como indicador. Não há como o caller distinguir entre:
1. Clipboard não disponível (sistema)
2. Clipboard falhou (erro temporário)
3. `CopyToClipboard` não foi solicitado

---

## 6. Métricas de Integração

| Métrica | Valor |
|---------|-------|
| Linhas de código do bloco clipboard | ~10 (service.go:140-154) |
| Chamadas de `clipboard.Copy()` | 1 por geração |
| Chamadas de `IsAvailable()` | 0 (não explícito) |
| Tratamento de erro | Non-fatal, log-only |
| Retry | 0 |
| Timeout | Não aplicável (operação local) |
| Concorrência | Não aplicável (chamada única síncrona) |
