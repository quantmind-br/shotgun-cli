# Fluxo: Carregamento de Templates

> **Funções:** `loadTemplatesFromFS()`, `TemplateSource.LoadTemplates()`
> **Arquivos:** `loader.go`

## Fluxograma: loadTemplatesFromFS

```mermaid
flowchart TD
    Start(["🟢 loadTemplatesFromFS"]) --> ReadDir["📂 fs.ReadDir(fsys, basePath)\n→ entries"]
    ReadDir --> ReadDirErr{"ReadDir\nsucesso?"}
    ReadDirErr -->|❌ erro| ReturnDirErr["❌ return nil, error\n'failed to read directory'"]
    ReadDirErr -->|✅ ok| ForEntry["🔁 Para cada entry em entries"]

    ForEntry --> IsDir{"entry.IsDir()\n||\n!HasSuffix('.md')?"}
    IsDir -->|✅ sim| NextEntry["⏭️ continue (ignora)"]
    IsDir -->|❌ não| ReadFile["📄 fs.ReadFile(fsys, filePath)\n→ content"]
    ReadFile --> ReadFileErr{"ReadFile\nsucesso?"}
    ReadFileErr -->|❌ erro| NextEntry
    ReadFileErr -->|✅ ok| Parse["🔧 parseTemplate(content, name, path)"]
    Parse --> ParseErr{"parseTemplate\nsucesso?"}
    ParseErr -->|❌ erro| NextEntry
    ParseErr -->|✅ ok| SetMeta["🏷️ template.IsEmbedded = isEmbedded\n   template.Source = sourceName"]
    SetMeta --> ExtractName["📛 templateName = extractTemplateName(entry.Name)"]
    ExtractName --> Store["🗂️ templates[templateName] = template"]
    Store --> NextEntry

    NextEntry --> HasMore{"mais entries?"}
    HasMore -->|✅ sim| ForEntry
    HasMore -->|❌ não| ReturnOk["✅ return templates"]
    ReturnDirErr --> End(["🔵 FIM"])
    ReturnOk --> End
```

## Fluxograma: EmbeddedSource

```mermaid
flowchart TD
    Start(["🟢 EmbeddedSource.LoadTemplates"]) --> LoadFS["📂 loadTemplatesFromFS(\n  s.fsys, '.', true, 'embedded')"]
    LoadFS --> Return["✅ retorna map[string]*Template"]
    Return --> End(["🔵 FIM"])
```

## Fluxograma: FilesystemSource

```mermaid
flowchart TD
    Start(["🟢 FilesystemSource.LoadTemplates"]) --> DirFS["📁 os.DirFS(s.path)\n→ fsys virtual"]
    DirFS --> LoadFS["📂 loadTemplatesFromFS(\n  fsys, '.', false, s.sourceName)"]
    LoadFS --> Return["✅ retorna map[string]*Template"]
    Return --> End(["🔵 FIM"])
```

## Regras de Filtragem

| Critério | Ação |
|----------|------|
| `entry.IsDir() == true` | Ignora (continue) |
| `!strings.HasSuffix(name, ".md")` | Ignora (continue) |
| `fs.ReadFile` falha | Ignora com continue (silencioso) |
| `parseTemplate` retorna erro | Ignora com continue (silencioso) |
| `fs.ReadDir` falha | Retorna erro (não silencia) |
