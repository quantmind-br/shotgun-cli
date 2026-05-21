# Análise de Código — `internal/core/contextgen`

| Campo           | Valor                                                                 |
|-----------------|-----------------------------------------------------------------------|
| **Módulo**      | `internal/core/contextgen`                                            |
| **Package**     | `github.com/quantmind-br/shotgun-cli/internal/core/contextgen`        |
| **Nível de detalhe** | detalhado                                                      |
| **Dependência externa** | `github.com/quantmind-br/shotgun-cli/internal/core/scanner`  |
| **Arquivos**    | 8 (5 `.go` + 3 `_test.go`)                                            |
| **Total de linhas (estimado)** | ~950                                                      |

---

## 1. Visão Geral

O pacote `contextgen` é o **gerador de contexto** do Shotgun CLI. Sua responsabilidade é escanear um projeto, ler arquivos, construir uma estrutura de árvore ASCII, coletar conteúdos, e renderizar tudo usando um template Go (`text/template`) para produzir um bloco de contexto textual — tipicamente usado como prompt para LLMs (chat completions).

É um módulo **core (zero dependências externas)**, dependendo apenas do `scanner` do mesmo projeto.

---

## 2. Estrutura de Arquivos

| Arquivo | Linhas (est.) | Tipo | Descrição |
|---------|---------------|------|-----------|
| `content.go` | ~190 | package | Coleta de conteúdo de arquivos, detecção de linguagem, filtro binário, renderização XML-like |
| `generator.go` | ~180 | package | Interface `ContextGenerator`, `DefaultContextGenerator`, pipeline principal `Generate*`, configuração |
| `template.go` | ~70 | package | `TemplateRenderer`, template padrão, funções personalizadas, validação de variáveis |
| `tree.go` | ~150 | package | `TreeRenderer`, renderização de árvore ASCII, ignorados, depth limit, ordenação |
| `content_test.go` | ~150 | test | Tests de `isTextFile`, `detectLanguage`, `readFileContent`, `peekFileHeader` |
| `filestructure_test.go` | ~100 | test | Testes de renderização de árvore + conteúdo combinados |
| `generator_test.go` | ~300 | test | Cenários completos de geração, progress reporting, benchmarks |
| `tree_test.go` | ~200 | test | Tests unitários extensivos de `TreeRenderer` |

---

## 3. Interfaces e Tipos Públicos

### 3.1 Interface `ContextGenerator`

```go
type ContextGenerator interface {
    Generate(root *scanner.FileNode, selections map[string]bool, config GenerateConfig) (string, error)
    GenerateWithProgress(root *scanner.FileNode, selections map[string]bool, config GenerateConfig, progress func(string)) (string, error)
    GenerateWithProgressEx(root *scanner.FileNode, selections map[string]bool, config GenerateConfig, progress func(GenProgress)) (string, error)
}
```

Três métodos para diferentes níveis de observabilidade:
- **`Generate`** — interface mínima, sem progress callbacks.
- **`GenerateWithProgress`** — callback `func(string)`, adaptado internamente para `GenProgress`.
- **`GenerateWithProgressEx`** — callback `func(GenProgress)` com estágio e mensagem estruturados.

### 3.2 `DefaultContextGenerator`

```go
type DefaultContextGenerator struct {
    treeRenderer     *TreeRenderer
    templateRenderer *TemplateRenderer
}
```

- **`NewDefaultContextGenerator()`** — construtor que inicializa `TreeRenderer` e `TemplateRenderer`.
- **`Generate(...)`** — delega para `GenerateWithProgress` com progress = `nil`.
- **`GenerateWithProgress(...)`** — adapta `func(string)` → `func(GenProgress)`, delega para `GenerateWithProgressEx`.
- **`GenerateWithProgressEx(...)`** — **pipeline principal** (detalhado na §5).
- **`validateConfig(...)`** — aplica valores padrão se campos não definidos.
- **`collectFileContents(...)`** — delega para a função livre `collectFileContents`.
- **`buildCompleteFileStructure(...)`** — combina árvore ASCII + blocos de conteúdo XML-like.
- **`convertTemplateVariables(...)`** — converte `{VAR}` → `{{.Var}}`.

### 3.3 `GenerateConfig`

```go
type GenerateConfig struct {
    MaxFileSize    int64             // Máximo por arquivo individual
    MaxTotalSize   int64             // Máximo total acumulado
    MaxFiles       int               // Limite de arquivos
    SkipBinary     bool              // Pular arquivos binários
    TemplateVars   map[string]string // Variáveis do template
    Template       string            // Template personalizado (opcional)
    IncludeTree    bool              // Incluir árvore ASCII
    IncludeSummary bool              // Incluir sumários (não usado ativamente)
}
```

**Valores padrão** (aplicados por `validateConfig`):
- `MaxFileSize` = `10 * 1024 * 1024` (10 MB)
- `MaxTotalSize` = `10 * 1024 * 1024` (10 MB)
- `MaxFiles` = `1000`

### 3.4 `ContextData`

```go
type ContextData struct {
    Task          string         // Extraído de TemplateVars["TASK"]
    Rules         string         // Extraído de TemplateVars["RULES"]
    FileStructure string         // Árvore + conteúdo
    Files         []FileContent
    CurrentDate   string         // Formatação: "2006-01-02 15:04:05"
    Config        GenerateConfig
}
```

### 3.5 `GenProgress`

```go
type GenProgress struct {
    Stage   string // "tree_generation" | "content_collection" | "template_rendering" | "complete"
    Message string // Mensagem legível
}
```

### 3.6 `FileContent` (definido em `content.go`)

```go
type FileContent struct {
    Path     string `json:"path"`
    RelPath  string `json:"relPath"`
    Language string `json:"language"`
    Content  string `json:"content"`
    Size     int64  `json:"size"`
}
```

### 3.7 `TreeRenderer` (definido em `tree.go`)

```go
type TreeRenderer struct {
    showIgnored bool   // Default: false
    maxDepth    int    // Default: -1 (sem limite)
}
```

Métodos de builder:
- **`WithShowIgnored(bool)`** — retorna o mesmo `*TreeRenderer` modificado (mutação direta).
- **`WithMaxDepth(int)`** — retorna o mesmo `*TreeRenderer` modificado.
- **`RenderTree(*scanner.FileNode)`** — retorna árvore ASCII ou erro.

### 3.8 `TemplateRenderer` (definido em `template.go`)

```go
type TemplateRenderer struct {
    funcs        template.FuncMap
    requiredVars []string // Default: ["TASK"]
}
```

- **`NewTemplateRenderer()`** — construtor.
- **`RenderTemplate(string, ContextData)`** — valida variáveis obrigatórias (somente para template padrão), parse e execute.
- **`validateRequiredVars(ContextData)`** — verifica `TemplateVars` contêm chaves exigidas.
- **`getDefaultTemplate()`** — retorna template Markdown padrão.

---

## 4. Funções Privadas

### 4.1 Em `content.go`

| Função | Assinatura | Descrição |
|--------|-----------|-----------|
| `collectFileContents` | `(root, selections, config) ([]FileContent, error)` | Percorre árvore, lê arquivos, aplica filtros (tamanho, binário, count), retorna `[]FileContent`. |
| `walkSelectedNodes` | `(node, fn func(*FileNode) error) error` | DFS recursivo sobre `FileNode` chamando `fn`. |
| `peekFileHeader` | `(path string) ([]byte, error)` | Abre arquivo e lê primeiros 1024 bytes para detecção binária. |
| `readFileContent` | `(path string) (string, error)` | Lê arquivo inteiro como `string`. |
| `isTextFile` | `(content string) bool` | Verifica bytes nulos (0x00) nos primeiros 1024 bytes + validação UTF-8. |
| `detectLanguage` | `(filename string) string` | Tenta basename primeiro, depois extensão. |
| `detectLanguageByBasename` | `(base string) string` | Mapeamento por nome base: Dockerfile, Makefile, Gemfile, etc. |
| `detectLanguageByExtension` | `(ext string) string` | Mapeamento por extensão (map de 50+ entradas), fallback `"text"`. |
| `shouldSkipFile` | `(node *FileNode, config GenerateConfig) bool` | Pula se `node.Size > config.MaxFileSize` ou se for dir. |
| `renderFileContentBlocks` | `(files []FileContent) string` | Renderiza `<file path="...">\ncontent\n</file>` para cada arquivo. |

**Mapas de extensão (50+ entradas):**

| Extensão | Linguagem | Extensão | Linguagem |
|----------|-----------|----------|-----------|
| `.go` | go | `.py/.pyw` | python |
| `.js/.jsx/.mjs` | javascript | `.java` | java |
| `.ts/.tsx` | typescript | `.c/.cpp/.cc/.cxx/.c++` | cpp |
| `.rb` | ruby | `.rs` | rust |
| `.sh/.bash/.zsh` | bash | `.ps1` | powershell |
| `.sql` | sql | `.html/.htm` | html |
| `.css` | css | `.scss/.sass` | scss |
| `.xml` | xml | `.json` | json |
| `.yaml/.yml` | yaml | `.toml` | toml |
| `.ini` | ini | `.md/.markdown` | markdown |
| `.tex` | latex | `.r` | r |
| `.m` | matlab | `.swift` | swift |
| `.kt` | kotlin | `.scala` | scala |
| `.clj/.cljs` | clojure | `.hs` | haskell |
| `.elm` | elm | `.dart` | dart |
| `.lua` | lua | `.vim` | vim |
| `.dockerfile` | dockerfile | `.cs` | csharp |
| `.php` | php | | |

### 4.2 Em `template.go`

| Função | Assinatura | Descrição |
|--------|-----------|-----------|
| `getTemplateFunctions` | `() template.FuncMap` | Retorna map com: `truncate`, `formatSize`, `detectLang`, `now`, `join`, `title`, `upper`, `lower`. |

### 4.3 Em `tree.go`

| Função | Assinatura | Descrição |
|--------|-----------|-----------|
| `formatNodeLine` | `(node, prefix, isLast) string` | Formata linha: `prefix├─── name/ (g) [1.0KB]`. |
| `getIgnoreIndicator` | `(node) string` | Retorna `" (g)"` se gitignored, `" (c)"` se custom ignored. |
| `getSizeInfo` | `(node) string` | Retorna `"[1.0KB]"` se arquivo com size > 0. |
| `renderChildren` | `(node, prefix, isLast, depth, result)` | Itera filhos visíveis, atualiza prefixo (│ ou espaço), chama `renderNode` recursivamente. |
| `getVisibleChildren` | `(node) []*FileNode` | Filtra `IsIgnored()` baseado em `showIgnored`. |
| `sortChildren` | `(children)` | Ordena: dirs primeiro, depois alfabetico. |
| `shouldSkipNode` | `(node, depth) bool` | Verifica `maxDepth` e `showIgnored`. |
| `formatFileSize` | `(bytes int64) string` | Formata: `100B`, `1.0KB`, `1.5MB`, `1.0GB`. |

### 4.4 Em `generator.go`

| Função | Assinatura | Descrição |
|--------|-----------|-----------|
| `convertTemplateVariables` | `(template string) string` | Converte `{TASK}` → `{{.Task}}`, `{RULES}` → `{{.Rules}}`, `{FILE_STRUCTURE}` → `{{.FileStructure}}`, `{CURRENT_DATE}` → `{{.CurrentDate}}`. |

---

## 5. Pipeline Principal de Geração (`GenerateWithProgressEx`)

```
┌─────────────────────────────────────────────────────────────────────┐
│ 1. validateConfig(&config)                                           │
│    • Aplica defaults (MaxFileSize, MaxTotalSize, MaxFiles)           │
│    • Garante TemplateVars não é nil                                  │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 2. if config.IncludeTree:                                            │
│    tree = treeRenderer.RenderTree(root)                              │
│    progress("tree_generation")                                       │
│    else: tree = ""                                                   │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 3. progress("content_collection")                                    │
│    files = collectFileContents(root, selections, config)             │
│    • DFS sobre FileNode                                              │
│    • Filtra: ignored, maxFiles, maxSize, binary                      │
│    • detectLanguage, relPath, size                                   │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 4. fileStructureComplete =                                           │
│      buildCompleteFileStructure(tree, files)                         │
│    • Inclui tree + separator + content blocks                        │
│    • Se IncludeTree=false: apenas content blocks                     │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 5. contextData = ContextData{                                       │
│      Task, Rules, FileStructure, Files, currentDate, Config         │
│    }                                                                 │
│    template = config.Template ?? defaultTemplate()                   │
│    template = convertTemplateVariables(template)                     │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 6. result, err = templateRenderer.RenderTemplate(template, ctxData) │
│    • Valida requiredVars (para default template)                     │
│    • text/template Parse + Execute                                   │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 7. if len(result) > config.MaxTotalSize:                            │
│      return error "exceeds total size limit"                         │
└──────────────────┬──────────────────────────────────────────────────┘
                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│ 8. progress("complete")                                              │
│    return result                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 6. Template Padrão

O `TemplateRenderer.getDefaultTemplate()` retorna:

```markdown
# Project Context

**Generated:** {{now}}
{{if .Task}}**Task:** {{.Task}}{{end}}
{{if .Rules}}**Rules:** {{.Rules}}{{end}}
{{if .FileStructure}}
## File Structure

{{.FileStructure}}
{{end}}
## File Contents
{{range .Files}}
### {{.RelPath}}{{if .Language}} ({{.Language}}){{end}}

```{{if .Language}}{{.Language}}{{end}}
{{.Content}}
```
{{end}}

---
*Context generated with {{formatSize .Config.MaxTotalSize}} size limit*
```

---

## 7. Análise de Dependências

### 7.1 Dependência Interna

| Depende de | Uso |
|------------|-----|
| `internal/core/scanner.FileNode` | Nó da árvore de arquivos escaneada. O `contextgen` **não escanea** — assume que o scanner já populou a árvore. Usa `node.Path`, `node.Name`, `node.IsDir`, `node.Size`, `node.Children`, `node.IsIgnored()`, `node.IsGitignored`, `node.IsCustomIgnored`. |

### 7.2 Dependências Externas

| Pacote | Uso |
|--------|-----|
| `github.com/stretchr/testify` | **Testes apenas** — `assert` e `require` |
| `golang.org/x/text/cases` | `cases.Title(language.English)` para função template `title` |
| `text/template` | Go standard library — parse + execute templates |
| `io`, `os`, `path/filepath`, `strings`, `fmt`, `errors`, `unicode/utf8`, `time`, `sort` | Standard library Go |

---

## 8. Pontos de Atenção

### 8.1 🔴 `IncludeSummary` não implementado
O campo `IncludeSummary bool` existe em `GenerateConfig` mas **nunca é utilizado** no pipeline. É um campo "zombie" — legado de um futuro recurso de sumário de arquivos.

### 8.2 🟡 Mutação em `TreeRenderer`
Os métodos `WithShowIgnored` e `WithMaxDepth` mutam `*TreeRenderer` diretamente (return `tr`). Isso significa que **o builder não é imutável** — se o mesmo renderer for reusado, o estado persiste. Em testes isso é inofensivo, mas pode causar bugs em produção se um renderer for compartilhado.

### 8.3 🟡 `convertTemplateVariables` — substituição insegura
A conversão `{TASK}` → `{{.Task}}` usa `strings.ReplaceAll` simples. Se um template contiver a string literal `{TASK}` que **não** é uma variável (ex: `{TASK_ID}`), seria convertido erroneamente para `{{.Task}}_ID`. O mapeamento é limitado a 4 chaves fixas, então o risco atual é baixo.

### 8.4 🟡 `readFileContent` não faz validação de tamanho
Ao contrário de `collectFileContents` (que verifica `MaxFileSize` por nó), a função `readFileContent` lê o arquivo inteiro sem limite. Se um arquivo fosse lido diretamente por outro caller, poderia causar OOM. Dentro de `collectFileContents` isso é mitigado pelo check `node.Size > config.MaxFileSize`.

### 8.5 🟢 Cobertura de testes robusta
- **`content_test.go`**: `isTextFile` (12 cenários), `detectLanguage` (50+ cenários por extensão e basename), `readFileContent`, `peekFileHeader`.
- **`generator_test.go`**: 5 cenários completos de geração (`single go file`, `binary skipped`, `total size limit`, `max files enforcement`, `custom template`), progress reporting, error propagation, template validation, benchmark.
- **`tree_test.go`**: 14 testes unitários de `TreeRenderer` (nil root, single file/dir, nested, depth limit, ignored files, sorting, ignore indicators).
- **`filestructure_test.go`**: Testes integrados de tree + content blocks + generator completo.

### 8.6 🟢 Validação de variáveis obrigatórias
O template padrão requer `TASK`. Se o usuário não passar `TASK` e não definir `Template`, a geração falha com erro claro: `"required template variable 'TASK' is missing or empty"`. Se um template customizado for fornecido, a validação é pulada.

---

## 9. Métricas

| Métrica | Valor |
|---------|-------|
| Funções públicas | 11 (interface 3 + struct 5 + constructor 2 + template 3) |
| Funções privadas | 17 |
| Tipos públicos | 6 (`ContextGenerator`, `GenProgress`, `GenerateConfig`, `ContextData`, `DefaultContextGenerator`, `FileContent`, `TreeRenderer`, `TemplateRenderer`) |
| Constantes | 8 (3 idiomas + 2 default sizes + 1 default files) |
| Mapas de extensão | 1 mapa com 50+ entradas |
| Testes unitários | ~30 testes |
| Benchmark | 1 (`BenchmarkDefaultContextGenerator` com 50 arquivos) |
| Linhas de código (prod) | ~590 |
| Linhas de teste | ~750 |
