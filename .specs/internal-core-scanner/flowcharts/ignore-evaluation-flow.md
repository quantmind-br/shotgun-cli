# Fluxo: shouldIgnore — Avaliação de Ignorância por Caminho

> **Módulo:** `internal/core/scanner`
> **Função analisada:** `FileSystemScanner.shouldIgnore(relPath string, isDir bool, config *ScanConfig) bool`
> **Arquivo fonte:** `internal/core/scanner/filesystem.go`, linhas 276–292

---

## 1. Visão Geral

O método `shouldIgnore` é o **guardião** de cada arquivo durante a varredura. Antes de criar um `FileNode`, o scanner consulta este método para decidir se o arquivo deve ser incluído ou excluído da árvore. A lógica combina **3 filtros independentes**:

1. **IncludePatterns** (inclusão seletiva) — filtro de whitelist
2. **IgnoreEngine** (.gitignore, .shotgunignore, custom rules) — filtro de blacklist
3. **Hidden files** (arquivos ocultos) — filtro de prefixo `.`

Todos os três filtros são **OR de exclusão** — se qualquer filtro diz "ignore", o arquivo é excluído.

---

## 2. Diagrama Mermaid

```mermaid
flowchart TD
    A([INÍCIO: shouldIgnore\nrelPath, isDir, config]) --> B{IncludePatterns\nconfigado?}
    
    B -- SIM --> C{matchesIncludePatterns\nrelPath, isDir, config}
    C -- NÃO --> D[RETORNO: true\n(ignorar)]
    C -- SIM --> E[continua]
    
    B -- NÃO --> E
    
    E --> F{ignoreEngine.ShouldIgnore\nrelPath}
    F -- SIM, ignored=true --> G{config.IncludeIgnored?}
    F -- NÃO, ignored=false --> H{config.IncludeHidden?}
    
    G -- SIM --> I[RETORNO: false\n(incluir com flag)]
    G -- NÃO --> D
    
    H -- SIM --> J[RETORNO: false\n(incluir)]
    H -- NÃO --> K{relPath starts with\n'.' ?}
    
    K -- SIM --> D
    K -- NÃO --> L[RETORNO: false\n(incluir)]
    
    D --> Z([FIM: ignora])
    I --> M([FIM: não ignora])
    J --> M
    L --> M
```

---

## 3. Descrição Detalhada do Fluxo

### Camada 1: IncludePatterns (Whitelist)

```go
func (fs *FileSystemScanner) matchesIncludePatterns(relPath string, isDir bool, config *ScanConfig) bool {
```

**Regras:**
1. Se `len(config.IncludePatterns) == 0` → `true` (inclui tudo) — filtro desligado.
2. Se `isDir == true` → `true` (sempre inclui diretórios para permitir traversal).
3. Para arquivos: verifica `filepath.Match(pattern, relPath)` ou `filepath.Match(pattern, fileName)`.
   - Compara contra **caminho relativo** E **basename do arquivo**.
   - `filepath.Match` — padrão glob (não regex): `*`, `?`, `[abc]`.

**Exemplos:**
| IncludePatterns | relPath | Resultado |
|---|---|---|
| `["*.go"]` | `"src/main.go"` | `true` (basename match) |
| `["*.go"]` | `"README.md"` | `false` → ignorado |
| `["src/*"]` | `"src/app.go"` | `true` (relPath match) |
| `[]` | qualquer | `true` (filtro desligado) |

### Camada 2: IgnoreEngine (Blacklist)

```go
ignored, _ := fs.ignoreEngine.ShouldIgnore(relPath)
if ignored {
    return !fs.shouldIncludeIgnored(config)
}
```

O `ignoreEngine.ShouldIgnore` retorna `(bool, IgnoreReason)` seguindo a ordem de prioridade de 5 camadas (documentada em `internal/core/ignore/flowcharts/should-ignore-flow.md`).

**Decisão:**
| IgnoreEngine | IncludeIgnored | Resultado |
|---|---|---|
| Ignored (qualquer reason) | `true` | `false` — inclui na árvore |
| Ignored (qualquer reason) | `false` | `true` — exclui da árvore |
| Not ignored | qualquer | `false` — passa para próxima camada |

### Camada 3: Hidden Files (Prefixo `.`)

```go
if !config.IncludeHidden {
    baseName := filepath.Base(relPath)
    if strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".." {
        return true
    }
}
```

**Regras:**
- Só aplicável se `IncludeHidden == false` (default).
- Verifica se o basename do caminho começa com `.`.
- Exceções: `"."` e `".."` nunca são considerados hidden.
- Aplica-se a **qualquer arquivo/diretório** começando com `.`: `.hidden.txt`, `.git/`, `.vscode/`, etc.

**Exemplos:**
| relPath | IncludeHidden | Resultado |
|---|---|---|
| `.hidden.txt` | `false` | `true` — ignorado |
| `.hidden.txt` | `true` | `false` — incluído |
| `src/.env` | `false` | `true` — ignorado |
| `src/main.go` | `false` | `false` — incluído |
| `.` | `false` | `false` — nunca ignorado |

---

## 4. Classificação de Ignorância — `classifyIgnoreReason()`

```go
func (fs *FileSystemScanner) classifyIgnoreReason(reason ignore.IgnoreReason) (bool, bool)
```

Esta função traduz `IgnoreReason` (do engine) em dois booleanos para `FileNode`:

| IgnoreReason | isGitignored | isCustomIgnored | Semântica |
|---|---|---|---|
| `IgnoreReasonGitignore` | `true` | `false` | Regra `.gitignore` |
| `IgnoreReasonBuiltIn` | `false` | `true` | Padrão embutido (node_modules, *.png, etc.) |
| `IgnoreReasonCustom` | `false` | `true` | `.shotgunignore` ou custom patterns |
| `IgnoreReasonExplicit` | `false` | `true` | Regra explícita (não usável pelo scanner) |
| Qualquer outro | `false` | `false` | Unknown — fallback seguro |

---

## 5. Integração com createFileNode

O fluxo completo de criação de um nó:

```
shouldIgnore(relPath, isDir, config)
  │
  ├─ matchesIncludePatterns(relPath, isDir, config) → false
  │   └─ SKIP: não cria FileNode, filepath.SkipDir (se dir)
  │
  └─ (se passou)
      │
      ├─ isGitignored, isCustomIgnored = getIgnoreStatusWithEngine(relPath, config)
      │   │
      │   ├─ ignored, reason = ignoreEngine.ShouldIgnore(relPath)
      │   │   ├─ ignored == true → classifyIgnoreReason(reason)
      │   │   └─ ignored == false
      │   │       ├─ isHiddenFile(relPath, config) → isCustomIgnored=true
      │   │       └─ else: isGitignored, isCustomIgnored from engine
      │   └─ retorna (isGitignored, isCustomIgnored)
      │
      └─ createFileNode(relPath, d, size, config)
          └─ FileNode{IsGitignored, IsCustomIgnored, Size, ...}
```

---

## 6. Casos de Borda

| Caso | Comportamento | Razão |
|---|---|---|
| `IncludePatterns` + `IgnorePatterns` | Include primeiro, depois ignore | Filtro de whitelist + blacklist |
| Dir com `IncludePatterns` | Sempre incluído (mesmo sem match) | Permite traversal |
| `IgnoreEngine` + `IncludeIgnored` | Ignorado mas incluído com flags | Visível na árvore |
| Arquivo `.env` em `src/.env` | Ignorado se `IncludeHidden=false` | Prefixo `.` |
| `.` ou `..` como basename | Nunca ignorado | Exceção explícita |
| `IgnoreReason(999)` | `false, false` | Fallback seguro |
| Pattern `*` em IncludePatterns | Matcha tudo | Glob wildcard |
