# Fluxo: walkAndBuild — Percursos e Construção da Árvore de Arquivos

> **Módulo:** `internal/core/scanner`
> **Função analisada:** `FileSystemScanner.walkAndBuild(rootPath, config, progress, total int64) (*FileNode, int64, error)`
> **Arquivo fonte:** `internal/core/scanner/filesystem.go`, linhas 115–225

---

## 1. Visão Geral

`walkAndBuild` é o **motor central** do scanner. Usa `filepath.WalkDir` para percorrer recursivamente o diretório raiz e, para cada caminho encontrado, decide se cria um `FileNode`, onde posicioná-lo na árvore, e como reportar progresso.

A construção da árvore usa um **mapa de nós de diretório** (`dirNodes map[string]*FileNode`) para resolver a relação pai-filho em O(1) por arquivo.

---

## 2. Diagrama Mermaid

```mermaid
flowchart TD
    A([INÍCIO: walkAndBuild]) --> B[Create root FileNode\nName=basename, IsDir=true\nExpanded=true, Children=[]]
    B --> C[dirNodes["."] = root]
    C --> D[filepath.WalkDir rootPath callback]
    
    D --> E{err != nil?}
    E -- SIM --> F[handleWalkError d]
    F --> F1{d.IsDir?}
    F1 -- SIM --> F2[return filepath.SkipDir]
    F1 -- NÃO --> F3[return nil\nsuppress error]
    F2 --> G
    F3 --> G
    
    E -- NÃO --> G{shouldStopWalking\nconfig, d, fileCount}
    
    G -- SIM (MaxFiles reached) --> H[return filepath.SkipDir]
    H --> I
    
    G -- NÃO --> J[relPath = filepath.Rel\nrootPath, path]
    
    J --> K{relPath erro ou "."?}
    K -- SIM --> L[return nil\ncontinue walking]
    K -- NÃO --> M
    
    M --> N{shouldIgnore\nrelPath, d.IsDir config}
    
    N -- SIM --> O{d.IsDir?}
    O -- SIM --> P[return filepath.SkipDir\nskip entire subtree]
    O -- NÃO --> Q[return nil\nskip single file]
    
    N -- NÃO --> R[size, skipFile = getFilesize\n d, config]
    
    R --> S{skipFile?}
    S -- SIM (MaxFileSize exceeded) --> L
    S -- NÃO --> T
    
    T --> U[createFileNode\npath, relPath, d, size, config]
    
    U --> V[addNodeToTree\nnode, relPath, dirNodes]
    
    V --> W[findParentNode\nrelPath, dirNodes]
    
    W --> X{parentNode exists?}
    X -- SIM --> Y[parentNode.Children.append node\nnode.Parent = parentNode]
    X -- NÃO --> Z[parentNode = dirNodes["."]\nfallback to root]
    
    Y --> AA
    Z --> AA
    
    AA --> BB{node.IsDir?}
    BB -- SIM --> CC[dirNodes[normRel relPath] = node]
    BB -- NÃO --> DD
    
    CC --> DD[current++\nif !d.IsDir: fileCount++]
    DD --> EE[reportProgress\nprogress, current, total, relPath\nif current % 100 == 0]
    DD --> F
    EE --> F
    
    F --> GG{WalkDir done?}
    GG -- SIM --> HH[return root, current, nil]
    
    H --> I
    I --> GG
    L --> GG
    P --> GG
    Q --> GG
    GG --> HH
    
    GG -- erro durante walk --> II[return root, current, error]
    II --> ZZ([FIM com erro])
    HH --> ZZ2([RETORNO root, current, nil])
```

---

## 3. Descrição Detalhada do Fluxo

### Passo 1: Criação do Nó Raiz

```go
root := &FileNode{
    Name:     filepath.Base(rootPath),
    Path:     rootPath,
    RelPath:  ".",
    IsDir:    true,
    Children: make([]*FileNode, 0),
    Expanded: true,
}
```
- O root tem `RelPath = "."`, `Expanded = true`, `Children` inicializado vazio.
- É registrado em `dirNodes["."]` como ponto de referência pai.

### Passo 2: `filepath.WalkDir(rootPath, callback)`

O `filepath.WalkDir` do Go garante ordem **DFS (pre-order traversal)**. Cada chamada do callback recebe:
- `path` — caminho absoluto
- `d` — `os.DirEntry`
- `err` — erro de acesso (nil = sucesso)

### Passo 3: Tratamento de Erros — `handleWalkError`

```go
func (fs *FileSystemScanner) handleWalkError(d os.DirEntry) error {
    if d != nil && d.IsDir() {
        return filepath.SkipDir  // Skip the directory
    }
    return nil  // Suppress file errors
}
```
- **Diretórios com erro** → `filepath.SkipDir` — evita travar o walker.
- **Arquivos com erro** → `nil` — error silencioso, arquivo simplesmente não aparece.

### Passo 4: Limite de Arquivos — `shouldStopWalking`

```go
func (fs *FileSystemScanner) shouldStopWalking(config *ScanConfig, d os.DirEntry, fileCount int64) bool {
    return config.MaxFiles > 0 && !d.IsDir() && fileCount >= config.MaxFiles
}
```
- Só aplica para **arquivos** (`!d.IsDir()`).
- Quando `fileCount >= MaxFiles`, retorna `true` → `filepath.SkipDir`.
- **Nota:** este é um "skip dir" mesmo para arquivos — o walker continua mas não processa o arquivo.

### Passo 5: Cálculo de RelPath

```go
relPath, err := filepath.Rel(rootPath, path)
```
- Converte caminho absoluto para relativo ao root do scan.
- Exemplo: `/project/src/main.go` → `src/main.go`
- Se `relPath == "."` → ignora (é o próprio root, já criado).

### Passo 6: Filtragem de Ignorância — `shouldIgnore`

```go
if fs.shouldIgnore(relPath, d.IsDir(), config) {
    return fs.skipIfDirectory(d)
}
```
- Se ignorado:
  - Se é diretório → `filepath.SkipDir` (pula toda subárvore)
  - Se é arquivo → `nil` (pula apenas este arquivo)
- **Importante:** diretórios ignorados são **pulsados completamente**, evitando traversal de `node_modules/`, `.git/`, etc.

### Passo 7: Tamanho do Arquivo — `getFileSize`

```go
func (fs *FileSystemScanner) getFileSize(d os.DirEntry, config *ScanConfig) (int64, bool) {
```
- Diretórios → `size = 0`, `skipFile = false`
- Arquivos com `MaxFileSize > 0` e `size > MaxFileSize` → `skipFile = true`
- Arquivos normais → `size`, `skipFile = false`

### Passo 8: Criação do Nó — `createFileNode`

```go
func (fs *FileSystemScanner) createFileNode(
    path, relPath string, d os.DirEntry, size int64, config *ScanConfig,
) *FileNode
```
- Chama `getIgnoreStatus(relPath, d.IsDir(), config)` → `(isGitignored, isCustomIgnored)`
- Cria `FileNode` com todos os campos preenchidos.
- `Expanded = false` para todos (exceto root que é `true`).

### Passo 9: Inserção na Árvore — `addNodeToTree`

```go
func (fs *FileSystemScanner) addNodeToTree(node *FileNode, relPath string, dirNodes map[string]*FileNode)
```

**Sub-passo: `findParentNode`**
```go
func (fs *FileSystemScanner) findParentNode(relPath string, dirNodes map[string]*FileNode) *FileNode
```
1. `parentPath = filepath.Dir(relPath)` — ex: `"src/main.go"` → `"src"`.
2. Se `parentPath == "."` → retorna `dirNodes["."]` (root).
3. Normaliza: `parentPath = normRel(parentPath)` → `filepath.ToSlash(parentPath)`.
4. Lookup direto: `dirNodes[parentPath]` — O(1).
5. **Fallback linear:** se não encontrado, split por `/` e tenta cada prefixo crescente.
6. Se ainda não encontrou → retorna root (`dirNodes["."]`).

**Inserção:**
- `node.Parent = parentNode`
- `parentNode.Children = append(parentNode.Children, node)`
- Se `node.IsDir`: registra em `dirNodes[normRel(relPath)] = node` (para futuros filhos).

### Passo 10: Reporte de Progresso — `reportProgress`

```go
func (fs *FileSystemScanner) reportProgress(progress chan<- Progress, current, total int64, relPath string)
```
- Só executa se `progress != nil && current % 100 == 0`.
- Envia: `{Current: current, Total: total, Stage: "scanning", Message: "Processing: <relPath>"}`
- **Throttle:** 100 items — evita flooding de canal.

### Passo 11: Retorno

```go
return root, current, nil
```
- `current` = contagem total de itens processados (arquivos + diretórios).

---

## 4. Mapa de Diretórios — `dirNodes`

Estrutura auxiliar crucial para construção eficiente da árvore:

```
dirNodes map[string]*FileNode

Exemplo para /project/src/main.go:
  dirNodes["."]       → FileNode{Name: "project", IsDir=true, Children: [...]}
  dirNodes["src"]      → FileNode{Name: "src", IsDir=true, Children: [...]}
  dirNodes["src/sub"]  → FileNode{Name: "sub", IsDir=true, Children: [...]}
```

**Por que funciona:** `filepath.WalkDir` visita em ordem DFS pre-order. Ao visitar `src/main.go`, o nó `src` já foi criado e registrado em `dirNodes["src"]`. Lookup O(1) do pai.

**Normalização:** `normRel(relPath)` usa `filepath.ToSlash()` para garantir consistência entre plataformas.

---

## 5. Casos de Borda do Fluxo

| Caso | Comportamento | Razão |
|---|---|---|
| `filepath.Rel` falha | `return nil` (continue) | Erro não fatal |
| `relPath == "."` | Ignora (root já criado) | Evita duplicação |
| `findParentNode` não encontra | Fallback linear → depois root | Segurança: nunca panica |
| `current % 100 == 0` | Reporte de progresso | Throttle de 100 |
| `current % 100 != 0` | Skip report | Otimização |
| `total = -1` | Streaming mode | UI mostra spinner |
| WalkDir encontra erro de permissão | `handleWalkError` decide | Skip dir, suppress file |
| MaxFiles = 0 (default) | `shouldStopWalking` sempre false | Sem limite |
| MaxFileSize = 0 (default) | `getFilesize` nunca skip | Sem limite |
