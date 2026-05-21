# Fluxo: Pipeline de Geração de Contexto

| Campo | Valor |
|-------|-------|
| **Módulo** | `internal/core/contextgen` |
| **Arquivo fonte** | `generator.go` (principal) |
| **Funções envolvidas** | `GenerateWithProgressEx`, `validateConfig`, `RenderTree`, `collectFileContents`, `buildCompleteFileStructure`, `RenderTemplate`, `convertTemplateVariables` |
| **Nível de detalhe** | detalhado |

---

## Mermaid

```mermaid
flowchart TD
    Start([Início: GenerateWithProgressEx]) --> Validate

    subgraph ConfigValidation [Validação e Defaults]
        Validate[validateConfig(&config)]
        Validate --> ApplyDefaults["Aplica defaults:\n- MaxFileSize → 10MB se 0\n- MaxTotalSize → 10MB se 0\n- MaxFiles → 1000 se ≤0\n- TemplateVars → map vazio se nil"]
    end

    ApplyDefaults --> CheckTree{config.IncludeTree?}

    CheckTree -->|Sim| GenTree["treeRenderer.RenderTree(root)\n→ Árvore ASCII"]
    CheckTree -->|Não| SkipTree["tree = \"\""]

    GenTree --> CollectFiles["collectFileContents(root, selections, config)\nDFS sobre FileNode"]
    SkipTree --> CollectFiles

    subgraph FileCollection [Coleta de Conteúdo]
        CollectFiles --> WalkRoot["walkSelectedNodes(root, fn)\nDFS recursivo"]
        WalkRoot --> CheckIsDir["node.IsDir?"]
        CheckIsDir -->|Sim| SkipNodeDir["skip (retorna nil)"]
        CheckIsDir -->|Não| CheckSelection

        CheckSelection["selections != nil && !selections[node.Path]?"] -->|Sim| SkipSelection["skip (não selecionado)"]
        CheckSelection -->|Não| CheckMaxFiles["fileCount >= config.MaxFiles?"]
        
        SkipSelection --> NextNode
        CheckMaxFiles -->|Sim| ErrMaxFiles["ERROR: 'maximum file count exceeded: %d'"]
        CheckMaxFiles -->|Não| CheckSkip

        CheckSkip["shouldSkipFile(node, config)?\n(node.Size > MaxFileSize)"] -->|Sim| SkipSize["skip (arquivo muito grande)"]
        CheckSkip -->|Não| CheckBinary

        CheckBinary["config.SkipBinary?\npeekFileHeader + isTextFile"] -->|Sim, binário| SkipBinary["skip (arquivo binário)"]
        CheckBinary -->|Não, ou SkipBinary=false| ReadFile

        SkipNodeDir --> NextNode
        SkipSelection --> NextNode
        SkipSize --> NextNode
        SkipBinary --> NextNode

        ReadFile["readFileContent(node.Path)\n→ string content"] --> CheckTotalSize
        CheckTotalSize["totalSize + len(content) > MaxTotalSize?"] -->|Sim| ErrTotalSize["ERROR: 'cumulative content size exceeds total size limit'"]
        CheckTotalSize -->|Não| BuildFile

        BuildFile["FileContent{\n  Path, RelPath, Language,\n  Content, Size\n}\n\nLanguage = detectLanguage(node.Name)"] --> AppendFiles["files = append(files, FileContent)\ntotalSize += Size\nfileCount++"]
        AppendFiles --> NextNode["Próximo nó (DFS)"]
        NextNode --> CheckIsDir
    end

    ErrMaxFiles --> EndErr
    ErrTotalSize --> EndErr
    EndErr([Erro retornado])

    CollectFiles --> BuildStructure

    subgraph StructureBuild [Construção da Estrutura]
        BuildStructure["buildCompleteFileStructure(tree, files)"]
        BuildStructure --> CheckTreeMode{config.IncludeTree?}
        CheckTreeMode -->|Sim| TreePlusContent["tree + '\\n' + renderFileContentBlocks(files)\n→ Árvore ASCII + blocos XML"]
        CheckTreeMode -->|Não| OnlyContent["renderFileContentBlocks(files)\n→ Apenas blocos XML"]
    end

    TreePlusContent --> BuildContext
    OnlyContent --> BuildContext

    BuildContext["ContextData{\n  Task = TemplateVars['TASK'],\n  Rules = TemplateVars['RULES'],\n  FileStructure,\n  Files,\n  CurrentDate = time.Now(),\n  Config\n}"] --> GetTemplate

    GetTemplate["template = config.Template\nse vazio → getDefaultTemplate()"] --> ConvertVars["convertTemplateVariables(template)\n{TASK} → {{.Task}}\n{RULES} → {{.Rules}}\n{FILE_STRUCTURE} → {{.FileStructure}}\n{CURRENT_DATE} → {{.CurrentDate}}"]

    ConvertVars --> ValidateTemplateVars
    ValidateTemplateVars["TemplateRenderer.validateRequiredVars(data)\n(validação só para default template)\nrequisita 'TASK'"] --> ParseTemplate["text/template.New('context')\n.Funcs(templateFuncs)\n.Parse(template)"]
    
    ParseTemplate --> ExecuteTemplate["tmpl.Execute(&buf, contextData)"]

    ExecuteTemplate --> CheckOutputSize["len(result) > MaxTotalSize?"]
    CheckOutputSize -->|Sim| ErrOutputSize["ERROR: 'generated context exceeds total size limit'"]
    CheckOutputSize -->|Não| Complete("GenProgress{Stage:'complete', Message:'Context generation completed'}")

    ErrOutputSize --> EndErr
    Complete --> EndOk([Retorna result (string)])

    style Start fill:#e1f5fe
    style EndOk fill:#c8e6c9
    style EndErr fill:#ffcdd2
    style CheckOutputSize fill:#fff9c4
```

---

## Descrição Detalhada do Fluxo

### Fase 1: Validação de Configuração
A função `validateConfig` é chamada no início. Se `MaxFileSize`, `MaxTotalSize` ou `MaxFiles` estiverem em zero/valor inválido, valores padrão são aplicados:
- `MaxFileSize` → `10 * 1024 * 1024` (10 MB)
- `MaxTotalSize` → `10 * 1024 * 1024` (10 MB)
- `MaxFiles` → `1000`
- `TemplateVars` → `map[string]string{}` se nil

### Fase 2: Geração da Árvore (condicional)
Se `config.IncludeTree == true`:
1. `treeRenderer.RenderTree(root)` é chamado.
2. O `TreeRenderer` faz DFS recursivo sobre `*scanner.FileNode`.
3. Para cada nó:
   - Se `maxDepth >= 0 && depth > maxDepth` → nó é pulado.
   - Se `!showIgnored && node.IsIgnored()` → nó é pulado.
   - Se ignorado e visível: indicador `" (g)"` (git) ou `" (c)"` (custom).
   - Se arquivo com tamanho > 0: tamanho formatado é adicionado (ex: `[1.0KB]`).
   - Filhos são ordenados: diretórios primeiro, depois alfabético.
   - Prefixo é construído com conectores ASCII: `├──`, `└──`, `│`.
4. Se `IncludeTree == false`, `fileStructure` fica vazio.

### Fase 3: Coleta de Conteúdo de Arquivos
A função `collectFileContents` faz DFS sobre o `FileNode`:

Para cada nó não-diretório:
1. **Seleção**: se `selections` não é nil e o caminho do nó não está no mapa → skip.
2. **Limite de arquivos**: se `fileCount >= config.MaxFiles` → erro.
3. **Tamanho do arquivo**: se `node.Size > config.MaxFileSize` → skip.
4. **Detecção de binário**: se `config.SkipBinary`, lê primeiros 1024 bytes, verifica se contém byte `0x00` ou se não é UTF-8 válido → skip.
5. **Leitura do arquivo**: `os.Open` + `io.ReadAll` → string.
6. **Limite de tamanho total**: se `totalSize + len(content) > config.MaxTotalSize` → erro.
7. **Construção do FileContent**: calcula `RelPath` via `filepath.Rel(root.Path, node.Path)`, detecta linguagem, armazena.

A linguagem é detectada por:
1. **Basename** (prioridade): `dockerfile`, `makefile`, `rakefile`, `Gemfile`, `package.json`, `Cargo.toml`, `go.mod`, `requirements.txt`, etc.
2. **Extensão**: mapeamento de 50+ extensões (.go → go, .py → python, etc.), fallback `"text"`.

### Fase 4: Construção da Estrutura Final
Se `IncludeTree == true`: combina árvore ASCII + `\n` + blocos de conteúdo.
Se `IncludeTree == false`: apenas blocos de conteúdo.

Cada bloco de conteúdo usa formato XML-like:
```xml
<file path="rel/path">
content here
</file>
```

### Fase 5: Preparação do Template
1. `ContextData` é construído com `Task`, `Rules`, `FileStructure`, `Files`, `CurrentDate` (formato `2006-01-02 15:04:05`), e `Config`.
2. Template é definido: `config.Template` se fornecido, senão `getDefaultTemplate()`.
3. Conversão de variáveis: `{TASK}` → `{{.Task}}`, `{RULES}` → `{{.Rules}}`, `{FILE_STRUCTURE}` → `{{.FileStructure}}`, `{CURRENT_DATE}` → `{{.CurrentDate}}`.

### Fase 6: Renderização do Template
1. **Validação de variáveis**: para o template padrão, valida que `TASK` existe em `TemplateVars`. Templates customizados não passam por esta validação.
2. **Parse**: `text/template.New("context").Funcs(tr.funcs).Parse(template)`.
3. **Execução**: `tmpl.Execute(&buf, contextData)`.

Funções customizadas disponíveis no template:
- `truncate` — trunca string com `...`
- `formatSize` — formata bytes (KB/MB/GB)
- `detectLang` — detecta linguagem por nome
- `now` — data/hora atual
- `join` — join de strings
- `title` — title case (inglês)
- `upper` / `lower` — case conversion

### Fase 7: Validação de Tamanho de Saída
Se `len(result) > config.MaxTotalSize` → erro.

### Fase 8: Conclusão
Progresso `GenProgress{Stage: "complete", Message: "Context generation completed"}` é disparado. Resultado é retornado como `string`.

---

## Estados de Progresso

| Stage | Mensagem | Quando |
|-------|----------|--------|
| `tree_generation` | "Generating file structure..." | Antes de renderizar árvore |
| `content_collection` | "Collecting file contents..." | Antes de coletar arquivos |
| `template_rendering` | "Rendering template..." | Antes de executar template |
| `complete` | "Context generation completed" | Após sucesso |

---

## Pontos de Falha

| Ponto | Erro | Causa |
|-------|------|-------|
| Config | `invalid config: ...` | Nunca disparado (validateConfig só aplica defaults) |
| Tree | `failed to render tree: ...` | Erro interno do TreeRenderer (raro, só nil root) |
| File count | `maximum file count exceeded: %d` | `fileCount >= config.MaxFiles` |
| Binary peek | `failed to peek file header %s: %w` | Falha ao abrir arquivo para peek |
| Read file | `failed to read file %s: %w` | Falha ao ler arquivo |
| Total size | `cumulative content size exceeds total size limit: %d + %d > %d` | Acumulador ultrapassa `MaxTotalSize` |
| Template validation | `template variable validation failed: ...` | Variável obrigatória (`TASK`) ausente |
| Template parse | `failed to parse template: %w` | Template com sintaxe inválida Go template |
| Template execute | `failed to execute template: %w` | Erro durante execução do template |
| Output size | `generated context exceeds total size limit: %d bytes > %d bytes` | Resultado final ultrapassa `MaxTotalSize` |
