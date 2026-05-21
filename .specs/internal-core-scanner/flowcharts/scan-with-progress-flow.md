# Fluxo: ScanWithProgress — Varredura Completa do Sistema de Arquivos

> **Módulo:** `internal/core/scanner`
> **Função analisada:** `FileSystemScanner.ScanWithProgress(rootPath, config, progress chan<- Progress)`
> **Arquivo fonte:** `internal/core/scanner/filesystem.go`, linhas 36–113

---

## 1. Visão Geral

Este é o **fluxo principal** do módulo. Realiza uma varredura recursiva do sistema de arquivos a partir de um root path, constrói uma árvore de `FileNode`, filtra por ignorância (gitignore, shotgunignore, custom, hidden), aplica limites (MaxFileSize, MaxFiles), reporta progresso, e ordena os resultados.

---

## 2. Diagrama Mermaid

```mermaid
flowchart TD
    A([INÍCIO: ScanWithProgress]) --> B[config == nil ?\nDefaultScanConfig()]
    
    B --> C[os.Stat rootPath]
    C --> D{path é dir válido?}
    D -- NÃO --> E[erro: invalid/not a directory]
    E --> Z([FIM])
    
    D -- SIM --> F{RespectGitignore?}
    F -- SIM --> G[ignoreEngine.LoadGitignore rootPath]
    F -- NÃO --> H{RespectShotgunignore?}
    G --> H
    
    H -- SIM --> I[ignoreEngine.LoadShotgunignore rootPath]
    H -- NÃO --> J{IgnorePatterns len > 0?}
    I --> J
    
    J -- SIM --> K[ignoreEngine.AddCustomRules IgnorePatterns]
    J -- NÃO --> L{progress != nil?}
    K --> L
    
    L -- SIM --> L1[progress <- {Current:0, Total:-1,\nStage:scanning, Msg:Scanning...}]
    L1 --> M[walkAndBuild rootPath config progress total=-1]
    L -- NÃO --> M
    
    M --> N{walkAndBuild erro?}
    N -- SIM --> O[erro: failed to scan directory]
    O --> Z
    
    N -- NÃO --> P[sortChildren root]
    
    P --> Q{progress != nil?}
    Q -- SIM --> R[progress <- {Current:actualCount,\nTotal:actualCount,\nStage:complete}]
    Q -- NÃO --> S
    R --> S
    
    S --> T([RETORNO root, nil])
    
    Z([FIM com erro])
```

---

## 3. Descrição Detalhada do Fluxo

### Fase 1: Validação de Entrada

```go
if config == nil {
    config = DefaultScanConfig()
}
```
- Se `config` é nil, usa defaults: `Workers=1`, `RespectGitignore=true`, `RespectShotgunignore=true`, `IncludeHidden=false`.

```go
info, err := os.Stat(rootPath)
if err != nil { ... }
if !info.IsDir() { ... }
```
- Valida que o caminho existe e é um diretório.
- Erro: `"invalid root path: %w"` ou `"root path is not a directory: %s"`.

### Fase 2: Carregamento de Regras de Ignorância

O scanner carrega regras **antes** de iniciar a varredura:

| Condição | Ação | Método do Engine |
|---|---|---|
| `config.RespectGitignore == true` | Carrega `.gitignore` recursivo | `LoadGitignore(rootPath)` |
| `config.RespectShotgunignore == true` | Carrega `.shotgunignore` recursivo | `LoadShotgunignore(rootPath)` |
| `len(config.IgnorePatterns) > 0` | Adiciona padrões custom | `AddCustomRules(…)` |

- `LoadGitignore` e `LoadShotgunignore` varrem **todos** os arquivos `.gitignore` / `.shotgunignore` na árvore (incluindo aninhados).
- Se nenhum arquivo for encontrado, o matcher fica vazio (sem erro).
- Regras shotgunignore são tratadas como **custom rules** (AddCustomRules).

### Fase 3: Notificação de Progresso Inicial

```go
if progress != nil {
    progress <- Progress{
        Current:   0,
        Total:     -1,  // Streaming mode
        Stage:     "scanning",
        Message:   "Scanning files...",
        Timestamp: time.Now(),
    }
}
```
- `Total = -1` indica modo streaming — a UI sabe que o total é desconhecido.
- Mostra spinner/progresso indeterminado.

### Fase 4: `walkAndBuild(rootPath, config, progress, total=-1)`

Esta função realiza toda a varredura recursiva. Detalhada no arquivo `walk-and-build-flow.md`.

Resulta em:
- `root` — árvore de FileNode
- `actualCount` — contagem total de itens processados
- `err` — erro, se qualquer

### Fase 5: Ordenação da Árvore

```go
fs.sortChildren(root)
```
- Recorre toda a árvore.
- Em cada nó de directory, ordena `Children`:
  - Diretóros antes de arquivos
  - Dentro de cada grupo: ordem alfabética (case-insensitive)
- **Complexidade:** O(n × log n) por nível da árvore.

### Fase 6: Notificação de Progresso Final

```go
progress <- Progress{
    Current:   actualCount,
    Total:     actualCount,  // Agora sabemos o total
    Stage:     "complete",
    Message:   "Scan completed successfully",
    Timestamp: time.Now(),
}
```
- Agora `Total` é o valor real, permitindo progresso determinístico na UI.

### Fase 7: Retorno

```go
return root, nil
```

---

## 4. Fluxo de Dados — Estado do FileSystemScanner

```
NewFileSystemScanner()
  │
  └─ ignoreEngine = NewIgnoreEngine()
      ├─ builtInMatcher = CompileIgnoreLines(42+ patterns)
      └─ gitignoreMatcher = CompileIgnoreLines()   ← vazio
      └─ customMatcher = CompileIgnoreLines()       ← vazio
      └─ explicitExcludes = CompileIgnoreLines()    ← vazio
      └─ explicitIncludes = CompileIgnoreLines()    ← vazio
  │
  ▼
[Scanner pronto — scan com built-in ativo]
  │
  ▼
ScanWithProgress(rootPath, config, progress)
  │
  ├─ RespectGitignore
  │   └─ LoadGitignore(rootPath)
  │       └─ Walk(rootPath) → .gitignore files
  │       └─ Read + parse
  │       └─ gitignoreMatcher = CompileIgnoreLines(all)
  │
  ├─ RespectShotgunignore
  │   └─ LoadShotgunignore(rootPath)
  │       └─ Walk(rootPath) → .shotgunignore files
  │       └─ AddCustomRules(all)
  │       └─ customMatcher = CompileIgnoreLines(all)
  │
  ├─ IgnorePatterns
  │   └─ AddCustomRules(config.IgnorePatterns)
  │   └─ customMatcher = CompileIgnoreLines(accumulated)
  │
  ▼
walkAndBuild(rootPath, config, progress, total=-1)
  │
  ├─ root = FileNode{rootPath basename, IsDir=true, Expanded=true}
  ├─ dirNodes["."] = root
  ├─ filepath.WalkDir(rootPath, callback)
  │   └─ Para cada path → shouldIgnore → createFileNode → addNodeToTree
  │
  ├─ sortChildren(root)
  │   └─ dirs first, then files, alphabetical case-insensitive
  │
  └─ progress <- {Current: actualCount, Total: actualCount, Stage: "complete"}
  │
  ▼
return root, nil
```

---

## 5. Configuração de Progresso — Throttling

```go
func reportProgress(progress chan<- Progress, current, total int64, relPath string) {
    if progress != nil && current%100 == 0 {
        progress <- Progress{
            Current: current,
            Total: total,
            Stage: "scanning",
            Message: fmt.Sprintf("Processing: %s", relPath),
            Timestamp: time.Now(),
        }
    }
}
```

- **Throttle factor:** `100` — envia update a cada 100 itens processados.
- **Motivo:** evita flooding de canal para diretórios com milhares de arquivos.
- **Exceções:** progresso inicial (0) e final (complete) não passam pelo throttle.
- **Caminho:** chamado no final de cada iteração de `WalkDir`.

---

## 6. Casos de Borda do Fluxo

| Caso | Comportamento | Razão |
|---|---|---|
| `config == nil` | Usa `DefaultScanConfig()` | Config padrão seguro |
| `rootPath` não existe | Erro `"invalid root path"` | Validação prévia |
| `rootPath` é arquivo | Erro `"root path is not a directory"` | Validação prévia |
| `progress == nil` | Nenhum canal enviado | Scan silencioso (via `Scan()`) |
| `.gitignore` ausente | `gitignoreMatcher` vazio | Sem erro |
| `.shotgunignore` ausente | Nenhuma adição | Sem erro |
| `MaxFiles` limitado | `filepath.SkipDir` ao atingir limite | Stop early |
| `MaxFileSize` limitado | Arquivo grande é pulado (skipFile=true) | Não entra na árvore |
| Permission denied (dir) | `filepath.SkipDir` (handleWalkError) | Tolerância a erros |
| Permission denied (file) | `nil` retornado (suppressed) | Tolerância a erros |
| `IncludeIgnored = false` | Arquivos ignorados não entram | Skip na árvore |
| `IncludeIgnored = true` | Arquivos ignorados entram com flags | Visíveis |
| `IncludeHidden = false` | Arquivos `.*` não entram | Skip na árvore |
| `IncludePatterns` especificado | Só arquivos matching entram | Filtro prévio |
| Árvore grande (100k files) | Memória linear O(n) | Sem streaming de saída |
