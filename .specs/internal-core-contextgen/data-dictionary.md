# Dicionário de Dados — `internal/core/contextgen`

| Campo           | Valor                                                                 |
|-----------------|-----------------------------------------------------------------------|
| **Módulo**      | `internal/core/contextgen`                                            |
| **Package**     | `github.com/quantmind-br/shotgun-cli/internal/core/contextgen`        |
| **Nível de detalhe** | detalhado                                                      |

---

## 1. Tipos Públicos

### 1.1 `ContextGenerator` (interface)

Interface que define o contrato do gerador de contexto.

| Método | Retorno | Descrição |
|--------|---------|-----------|
| `Generate(root, selections, config)` | `(string, error)` | Gera contexto sem callback de progresso. |
| `GenerateWithProgress(root, selections, config, progress)` | `(string, error)` | Gera com callback `func(string)` de progresso. |
| `GenerateWithProgressEx(root, selections, config, progress)` | `(string, error)` | Gera com callback `func(GenProgress)` estruturado. |

---

### 1.2 `GenProgress`

Estrutura de progresso estruturado.

| Campo | Tipo | JSON | Obrigatório | Descrição |
|-------|------|------|-------------|-----------|
| `Stage` | `string` | `"stage"` | Sim | Estágio da geração: `"tree_generation"`, `"content_collection"`, `"template_rendering"`, `"complete"`. |
| `Message` | `string` | `"message"` | Sim | Mensagem legível do progresso. |

**Valores possíveis de `Stage`:**
| Valor | Quando disparado |
|-------|-----------------|
| `tree_generation` | Antes de renderizar a árvore ASCII. |
| `content_collection` | Antes de coletar conteúdos de arquivos. |
| `template_rendering` | Antes de executar o template. |
| `complete` | Após geração bem-sucedida. |

---

### 1.3 `GenerateConfig`

Configuração do gerador de contexto.

| Campo | Tipo | JSON | Obrigatório | Padrão | Descrição |
|-------|------|------|-------------|--------|-----------|
| `MaxFileSize` | `int64` | `"maxFileSize"` | Não | `10 * 1024 * 1024` (10 MB) | Tamanho máximo por arquivo individual. |
| `MaxTotalSize` | `int64` | `"maxTotalSize"` | Não | `10 * 1024 * 1024` (10 MB) | Tamanho total acumulado de todo o conteúdo. |
| `MaxFiles` | `int` | `"maxFiles"` | Não | `1000` | Número máximo de arquivos a processar. |
| `SkipBinary` | `bool` | `"skipBinary"` | Não | `false` | Pular arquivos binários (detectados por byte null + UTF-8). |
| `TemplateVars` | `map[string]string` | `"templateVars"` | Não | `nil` → vazio | Variáveis injetáveis no template. |
| `Template` | `string` | `"template,omitempty"` | Não | `""` (usa default) | Template Go template personalizado. |
| `IncludeTree` | `bool` | `"includeTree"` | Não | `false` | Incluir árvore ASCII na saída. |
| `IncludeSummary` | `bool` | `"includeSummary"` | Não | `false` | Incluir sumários de arquivos. **🟡 Não implementado.** |

**Validações (aplicadas em `validateConfig`):**
- Se `MaxFileSize == 0` → definido para `DefaultMaxSize` (10 MB).
- Se `MaxTotalSize == 0` → definido para `DefaultMaxSize` (10 MB).
- Se `MaxFiles <= 0` → definido para `DefaultMaxFiles` (1000).
- Se `TemplateVars == nil` → inicializado como `map[string]string{}`.

---

### 1.4 `ContextData`

Dados passados ao template Go.

| Campo | Tipo | JSON | Obrigatório | Descrição |
|-------|------|------|-------------|-----------|
| `Task` | `string` | `"task"` | Sim | Valor de `TemplateVars["TASK"]`. |
| `Rules` | `string` | `"rules"` | Não | Valor de `TemplateVars["RULES"]`. |
| `FileStructure` | `string` | `"fileStructure"` | Sim | Árvore ASCII + blocos de conteúdo. |
| `Files` | `[]FileContent` | `"files"` | Sim | Lista de arquivos processados. |
| `CurrentDate` | `string` | `"currentDate"` | Sim | Data/hora no formato `"2006-01-02 15:04:05"`. |
| `Config` | `GenerateConfig` | `"config"` | Sim | Configuração utilizada. |

---

### 1.5 `FileContent`

Conteúdo de um único arquivo.

| Campo | Tipo | JSON | Obrigatório | Descrição |
|-------|------|------|-------------|-----------|
| `Path` | `string` | `"path"` | Sim | Caminho absoluto no sistema de arquivos. |
| `RelPath` | `string` | `"relPath"` | Sim | Caminho relativo ao root do projeto. |
| `Language` | `string` | `"language"` | Sim | Linguagem detectada (ex: `"go"`, `"python"`, `"text"`). |
| `Content` | `string` | `"content"` | Sim | Conteúdo textual do arquivo. |
| `Size` | `int64` | `"size"` | Sim | Tamanho em bytes do conteúdo. |

---

### 1.6 `TreeRenderer`

Renderizador de árvore de arquivos ASCII.

| Campo | Tipo | Padrão | Descrição |
|-------|------|--------|-----------|
| `showIgnored` | `bool` | `false` | Se `true`, arquivos ignorados são incluídos com indicador. |
| `maxDepth` | `int` | `-1` | Limite de profundidade (−1 = sem limite). |

**Métodos Públicos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `NewTreeRenderer()` | `*TreeRenderer` | Construtor. |
| `WithShowIgnored(bool)` | `*TreeRenderer` | Builder: define `showIgnored`. **Mutável.** |
| `WithMaxDepth(int)` | `*TreeRenderer` | Builder: define `maxDepth`. **Mutável.** |
| `RenderTree(*FileNode)` | `(string, error)` | Renderiza árvore ASCII ou erro. |

---

### 1.7 `TemplateRenderer`

Renderizador de templates Go.

| Campo | Tipo | Padrão | Descrição |
|-------|------|--------|-----------|
| `funcs` | `template.FuncMap` | — | Funções customizadas para templates. |
| `requiredVars` | `[]string` | `["TASK"]` | Variáveis obrigatórias para validação. |

**Métodos Públicos:**
| Método | Retorno | Descrição |
|--------|---------|-----------|
| `NewTemplateRenderer()` | `*TemplateRenderer` | Construtor. |
| `RenderTemplate(string, ContextData)` | `(string, error)` | Valida vars obrigatórias, parse + exec template. |

---

## 2. Funções Privadas — Dicionário

### 2.1 `content.go`

| Função | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|-----------|
| `collectFileContents` | `root *FileNode`, `selections map[string]bool`, `config GenerateConfig` | `[]FileContent, error` | DFS, lê arquivos, aplica filtros. |
| `walkSelectedNodes` | `node *FileNode`, `fn func(*FileNode) error` | `error` | DFS recursivo chamando `fn`. |
| `peekFileHeader` | `path string` | `[]byte, error` | Lê primeiros 1024 bytes. |
| `readFileContent` | `path string` | `string, error` | Lê arquivo inteiro. |
| `isTextFile` | `content string` | `bool` | True se não contiver 0x00 + UTF-8 válido nos primeiros 1024 bytes. |
| `detectLanguage` | `filename string` | `string` | Tenta basename, depois extensão. |
| `detectLanguageByBasename` | `base string` | `string` | Mapeamento por nome base. |
| `detectLanguageByExtension` | `ext string` | `string` | Mapeamento por extensão (fallback `"text"`). |
| `shouldSkipFile` | `node *FileNode`, `config GenerateConfig` | `bool` | True se dir ou tamanho > `MaxFileSize`. |
| `renderFileContentBlocks` | `files []FileContent` | `string` | `<file path="...">\ncontent\n</file>` por arquivo. |

### 2.2 `tree.go`

| Função | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|-----------|
| `formatNodeLine` | `node *FileNode`, `prefix string`, `isLast bool` | `string` | `"prefix├── name/ (g) [1.0KB]\n"` |
| `getIgnoreIndicator` | `node *FileNode` | `string` | `" (g)"` ou `" (c)"` ou `""`. |
| `getSizeInfo` | `node *FileNode` | `string` | `"[1.0KB]"` ou `""`. |
| `renderChildren` | `node *FileNode`, `prefix string`, `isLast bool`, `depth int`, `result *strings.Builder` | — | Itera filhos visíveis, recursivo. |
| `getVisibleChildren` | `node *FileNode` | `[]*FileNode` | Filtra ignorados. |
| `sortChildren` | `children []*FileNode` | — | Dirs primeiro, depois alfabético. |
| `shouldSkipNode` | `node *FileNode`, `depth int` | `bool` | Verifica depth limit e ignore. |
| `formatFileSize` | `bytes int64` | `string` | Formata bytes para KB/MB/GB. |

### 2.3 `template.go`

| Função | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|-----------|
| `getTemplateFunctions` | — | `template.FuncMap` | `truncate`, `formatSize`, `detectLang`, `now`, `join`, `title`, `upper`, `lower`. |

### 2.4 `generator.go`

| Função | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|-----------|
| `convertTemplateVariables` | `template string` | `string` | `{TASK}` → `{{.Task}}` etc. |

---

## 3. Constantes

| Constante | Valor | Tipo | Arquivo | Descrição |
|-----------|-------|------|---------|-----------|
| `langDockerfile` | `"dockerfile"` | `string` | `content.go` | Identificador de linguagem Dockerfile. |
| `langJSON` | `"json"` | `string` | `content.go` | Identificador de linguagem JSON. |
| `langRuby` | `"ruby"` | `string` | `content.go` | Identificador de linguagem Ruby. |
| `DefaultMaxSize` | `10 * 1024 * 1024` (10 MB) | `int` | `generator.go` | Tamanho default máximo. |
| `DefaultMaxFiles` | `1000` | `int` | `generator.go` | Arquivos default máximo. |

---

## 4. Mapa de Extensões para Linguagem

50+ entradas. Abaixo as principais:

| Extensão | Linguagem | Extensão | Linguagem |
|----------|-----------|----------|-----------|
| `.go` | go | `.py`, `.pyw` | python |
| `.js`, `.jsx`, `.mjs` | javascript | `.java` | java |
| `.ts`, `.tsx` | typescript | `.c`, `.cpp`, `.cc`, `.cxx`, `.c++` | cpp |
| `.h`, `.hpp`, `.hh`, `.hxx`, `.h++` | cpp | `.cs` | csharp |
| `.php` | php | `.rb` | ruby |
| `.rs` | rust | `.sh`, `.bash`, `.zsh` | bash |
| `.ps1` | powershell | `.sql` | sql |
| `.html`, `.htm` | html | `.css` | css |
| `.scss`, `.sass` | scss | `.xml` | xml |
| `.json` | json | `.yaml`, `.yml` | yaml |
| `.toml` | toml | `.ini` | ini |
| `.md`, `.markdown` | markdown | `.tex` | latex |
| `.r` | r | `.m` | matlab |
| `.swift` | swift | `.kt` | kotlin |
| `.scala` | scala | `.clj`, `.cljs` | clojure |
| `.hs` | haskell | `.elm` | elm |
| `.dart` | dart | `.lua` | lua |
| `.vim` | vim | `.dockerfile` | dockerfile |

**Fallback:** qualquer extensão não reconhecida retorna `"text"`.

**Prioridade de detecção:** basename primeiro, extensão segundo.

---

## 5. Conversão de Variáveis de Template

| Síntaxe Original | Síntaxe Go Template | Membro ContextData |
|-----------------|---------------------|-------------------|
| `{TASK}` | `{{.Task}}` | `.Task` |
| `{RULES}` | `{{.Rules}}` | `.Rules` |
| `{FILE_STRUCTURE}` | `{{.FileStructure}}` | `.FileStructure` |
| `{CURRENT_DATE}` | `{{.CurrentDate}}` | `.CurrentDate` |

---

## 6. Funções Customizadas do Template

| Função | Parâmetros | Retorno | Descrição |
|--------|-----------|---------|-----------|
| `truncate` | `s string`, `length int` | `string` | Trunca string, adiciona `"..."` se cortado. |
| `formatSize` | `bytes int64` | `string` | Formata tamanho (usado na função de `tree.go`). |
| `detectLang` | `filename string` | `string` | Detecta linguagem por nome de arquivo. |
| `now` | — | `string` | Retorna data/hora atual no formato `"2006-01-02 15:04:05"`. |
| `join` | `elems []string`, `sep string` | `string` | Junta strings com separador. |
| `title` | `s string` | `string` | Capitaliza (title case, inglês). |
| `upper` | `s string` | `string` | Maiúsculas. |
| `lower` | `s string` | `string` | Minúsculas. |
