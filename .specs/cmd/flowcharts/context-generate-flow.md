# Fluxo: Context Generate

> **Arquivo fonte:** `context.go`  
> **Comando:** `shotgun-cli context generate`

---

## Diagrama de Fluxo

```mermaid
flowchart TD
    Start([context generate]) --> preRunE["PreRunE"]
    
    preRunE --> validateRoot["Validar --root flag"]
    validateRoot --> rootEmpty["--root vazio?"]
    rootEmpty -->|Sim| errRootEmpty["Erro: root path cannot be empty"]
    rootEmpty -->|Não| toAbs["filepath.Abs(rootPath)"]
    
    toAbs --> absErr["Erro ao converter?"]
    absErr -->|Sim| errAbs["Erro: invalid root path"]
    
    absErr -->|Não| pathExists["os.Stat(absPath) existe?"]
    pathExists -->|Não| errNotExist["Erro: root path does not exist"]
    
    pathExists -->|Não (is dir?)| errNotDir["Erro: root path must be a directory"]
    pathExists -->|Sim| validateMaxSize["Validar --max-size format"]
    
    errNotExist --> returnErr([return error])
    errNotDir --> returnErr
    errAbs --> returnErr
    errRootEmpty --> returnErr
    validateMaxSize --> maxParseErr["ParseSize falhou?"]
    maxParseErr -->|Sim| errMaxSize["Erro: invalid max-size format"]
    errMaxSize --> returnErr
    
    validateMaxSize -->|OK| buildConfig["buildGenerateConfig(cmd)"]
    
    buildConfig --> readFlags["Ler flags: root, include, exclude, output, max-size, enforce-limit, template, task, rules, var, workers, include-hidden, include-ignored, progress"]
    
    readFlags --> parseVars["Parse --var KEY=VALUE → map[string]string"]
    parseVars --> parseErr["Formato inválido?"]
    parseErr -->|Sim| errVar["Erro: invalid --var format"]
    parseErr -->|Não| toAbsPath["filepath.Abs(rootPath)"]
    
    toAbsPath --> absPathErr["Erro?"]
    absPathErr -->|Sim| errAbsPath["Erro: failed to resolve absolute path"]
    
    absPathErr -->|Não| parseMax["utils.ParseSize(maxSizeStr)"]
    parseMax --> parseMaxErr["Erro?"]
    parseMaxErr -->|Sim| errMax["Erro: failed to parse max-size"]
    
    parseMaxErr -->|Não| genOutput["Output vazio? → shotgun-prompt-TIMESTAMP.md"]
    
    genOutput --> genConfig["GenerateConfig construído"]
    errAbsPath --> returnErr
    errMax --> returnErr
    errVar --> returnErr
    
    buildConfig --> genHeadless["generateContextHeadless(config)"]
    genConfig --> genHeadless
    
    genHeadless --> buildScannerCfg["buildScannerConfig(cfg)"]
    buildScannerCfg --> scanCfg["scanner.ScanConfig montado a partir de Viper + overrides"]
    scanCfg --> buildTemplateVars["buildTemplateVars(cfg)"]
    
    buildTemplateVars -> tVars["TASK, RULES, FILE_STRUCTURE, CURRENT_DATE, custom vars"]
    tVars --> loadTemplate["loadTemplateContent(cfg.Template)"]
    
    loadTemplate --> templateEmpty["Template nome vazio?"]
    templateEmpty -->|Sim| emptyTpl["templateContent = ''"]
    templateEmpty -->|Não| tplMgr["template.NewManager(CustomPath)"]
    
    tplMgr --> tplGet["tmplMgr.GetTemplate(templateName)"]
    tplGet --> tplErr["Template não encontrado?"]
    tplErr -->|Sim| errTpl["Erro: failed to load template"]
    tplErr -->|Não| tplContent["templateContent = tmpl.Content"]
    
    emptyTpl --> newService["app.NewContextService()"]
    tplContent --> newService
    
    errTpl --> returnErr
    
    newService --> svcCfg["app.GenerateConfig montado: RootPath, ScanConfig, Template, TemplateVars, MaxSize, EnforceLimit, OutputPath, CopyToClipboard, IncludeTree, IncludeSummary, SkipBinary"]
    svcCfg --> progressMode["cfg.ProgressMode diferente de 'none'?"]
    
    progressMode -->|Sim| withProgress["svc.GenerateWithProgress(ctx, svcCfg, callback)"]
    progressMode -->|Não| plainGen["svc.Generate(ctx, svcCfg)"]
    
    withProgress --> genCallback["Callback: renderProgress(mode, ProgressOutput{stage, msg, cur, total, percent})"]
    plainGen --> resultGen
    
    genCallback --> genDone["Generate concluído"]
    clearProgress["clearProgressLine(mode)"]
    clearProgress --> genDone
    
    genDone --> genErr["Erro na geração?"]
    genErr -->|Sim| errGen["Erro: context generation failed"]
    genErr -->|Não| printSummary["printGenerationSummary(result, cfg)"]
    
    errGen --> returnErr
    
    printSummary --> summary["  ✅ Context generated successfully!
  📁 Root path: ...
  📄 Output file: ...
  📊 Files processed: N
  📏 Total size: X (~Y tokens)
  🎯 Size limit: Z"]
    
    summary --> logSuccess["log.Info: 'Context generated successfully'"]
    logSuccess --> end([fim])
```

---

## Sub-fluxo: buildGenerateConfig

```mermaid
flowchart LR
    A[buildGenerateConfig(cmd)] --> B["Ler todas as flags"]
    B --> C["Parse --var KEY=VALUE"]
    C --> D{Formato válido?}
    D -->|Não| E["Erro: invalid --var format"]
    D -->|Sim| F["filepath.Abs(rootPath)"]
    F --> G["utils.ParseSize(maxSizeStr)"]
    G --> H["Output padrão se vazio: shotgun-prompt-TIMESTAMP.md"]
    H --> I[GenerateConfig{}]
    E --> Z[return error]
    I --> Z
```

---

## Sub-fluxo: buildScannerConfig

```mermaid
flowchart TD
    A[buildScannerConfig(cfg)] --> B["scanner.ScanConfig{}"]
    B --> C["Set defaults de Viper:
      - MaxFiles
      - MaxFileSize
      - MaxMemory
      - SkipBinary
      - IncludeHidden
      - IncludeIgnored
      - Workers
      - RespectGitignore
      - RespectShotgunignore"]
    C --> D["cfg.Exclude → IgnorePatterns"]
    D --> E["cfg.Include → IncludePatterns"]
    E --> F["cfg.Workers > 0?"]
    F -->|Sim| G["Overrides: Workers = cfg.Workers"]
    F -->|Não| H["cfg.IncludeHidden"]
    G --> H
    H --> I["cfg.IncludeIgnored"]
    I --> J[scanner.ScanConfig]
```

---

## Sub-fluxo: buildTemplateVars

```mermaid
flowchart TD
    A[buildTemplateVars(cfg)] --> B["TASK = cfg.Task ou 'Context generation'"]
    B --> C["RULES = cfg.Rules"]
    C --> D["FILE_STRUCTURE = ''"]
    D --> E["CURRENT_DATE = now().Format(YYYY-MM-DD)"]
    E --> F["Merge cfg.CustomVars"]
    F --> G[map[string]string]
```

---

## Resumo de Flags

| Flag | Tipo | Obrigatório | Padrão | Descrição |
|---|---|---|---|---|
| `--root` / `-r` | string | Sim (validado) | `"."` | Diretório raiz |
| `--include` / `-i` | []string | Não | `["*"]` | Padrões de inclusão |
| `--exclude` / `-e` | []string | Não | `[]` | Padrões de exclusão |
| `--output` / `-o` | string | Não | auto | Arquivo de saída |
| `--max-size` | string | Não | `"10MB"` | Tamanho máximo |
| `--enforce-limit` | bool | Não | `true` | Aplicar limite |
| `--template` / `-t` | string | Não | `""` | Template |
| `--task` | string | Não | `""` | Task description |
| `--rules` | string | Não | `""` | Rules |
| `--var` / `-V` | []string | Não | `[]` | Custom vars |
| `--workers` | int | Não | `0` | Workers |
| `--include-hidden` | bool | Não | `false` | Incluir ocultos |
| `--include-ignored` | bool | Não | `false` | Incluir ignorados |
| `--progress` | string | Não | `"none"` | none/human/json |
