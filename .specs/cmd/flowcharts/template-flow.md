# Fluxo: Template Management (list / render / import / export)

> **Arquivo fonte:** `template.go`  
> **Comandos:** `shotgun-cli template list`, `shotgun-cli template render`, `shotgun-cli template import`, `shotgun-cli template export`

---

## Diagrama de Fluxo: template list

```mermaid
flowchart TD
    Start([template list]) --> runList["templateListCmd.RunE()"]
    
    runList --> newManager["template.NewManager(
      ManagerConfig{CustomPath: viper.GetString(template.custom-path)}
    )"]
    
    newManager --> mgrErr["Erro ao criar Manager?"]
    mgrErr -->|Sim| errMgr["Erro: failed to initialize template manager"]
    mgrErr -->|Não| listTmpl["manager.ListTemplates()"]
    
    errMgr --> returnErr([return error])
    
    listTmpl --> listErr["Erro ao listar?"]
    listErr -->|Sim| errList["Erro: failed to list templates"]
    listErr -->|Não| emptyCheck{"len(templates) == 0?"}
    
    errList --> returnErr
    
    emptyCheck -->|Sim| noTemplates["Print: 'No templates available.'"]
    emptyCheck -->|Não| printHeader["Print:
  Available Templates:
  =================="]
    
    noTemplates --> endList(["fim"])
    
    printHeader --> calcWidth["Calcular maxNameWidth:
  Para cada template:
    nameWithSource = name + ' (' + source + ')'
    maxNameWidth = max(maxNameWidth, len(nameWithSource))"]
    
    calcWidth --> printTable["Iterar templates:
  nameFormatted = '%-*s' % (maxNameWidth, nameWithSource)
  Print: '  ' + nameFormatted + '  ' + description"]
    
    printTable --> printTotal["Print: 'Total: <n> templates'"]
    printTotal --> printHint["Print: \"Use 'shotgun-cli template render <name>'\""]
    printHint --> endList
```

---

## Diagrama de Fluxo: template render

```mermaid
flowchart TD
    Start([template render]) --> preRunE["PreRunE: cobra.ExactArgs(1)"]
    
    preRunE --> templateName["args[0] = templateName"]
    
    templateName --> newManager1["template.NewManager(CustomPath)"]
    
    newManager1 --> mgrErr1["Erro Manager?"]
    mgrErr1 -->|Sim| errMgr1["Erro: failed to initialize template manager"]
    mgrErr1 -->|Não| listTmpl1["manager.ListTemplates()"]
    
    errMgr1 --> returnErr([return error])
    
    listTmpl1 --> listErr1["Erro listar?"]
    listErr1 -->|Sim| errList1["Erro: failed to verify template existence"]
    listErr1 -->|Não| findTmpl{"Template existe?"}
    
    errList1 --> returnErr
    
    findTmpl -->|Não| availableNames["Collect availableNames"]
    availableNames --> errNotFound["Erro: template '<name>' not found.
      Available templates: <names>"]
    
    findTmpl -->|Sim| runRender["templateRenderCmd.RunE()"]
    errNotFound --> returnErr
    
    runRender --> readVars["cmd.Flags().GetStringToString('var')"]
    readVars --> readOutput["cmd.Flags().GetString('output')"]
    
    readOutput --> logRender["log: 'Rendering template'"]
    logRender --> callRender["renderTemplate(templateName, variables, output)"]
    
    callRender --> renderErr["Erro?"]
    renderErr -->|Sim| errRender["Erro: template rendering failed"]
    renderErr -->|Não| renderSuccess[OK]
    
    errRender --> returnErr
    
    renderSuccess --> outputCheck{"output != ''?"}
    outputCheck -->|Sim| printToFile["Print: '✅ Template <name> rendered successfully to: <output>'"]
    outputCheck -->|Não| logStdout["log.Info: 'Template rendered to stdout'"]
    
    printToFile --> endRender(["fim"])
    logStdout --> endRender
```

---

## Sub-fluxo: renderTemplate (interno)

```mermaid
flowchart TD
    A["renderTemplate(name, variables, output)"] --> newMgr["template.NewManager(CustomPath)"]
    newMgr --> mgrErr["Erro?"]
    mgrErr -->|Sim| return("return error")
    mgrErr -->|Não| getRequired["manager.GetRequiredVariables(name)"]
    
    getRequired --> reqErr["Erro?"]
    reqErr -->|Sim| return("return error: template not found")
    reqErr -->|Não| checkMissing["Verificar variáveis faltando:
      missingVars = requiredVars - variables"]
    
    checkMissing --> hasMissing{"len(missingVars) > 0?"}
    hasMissing -->|Sim| returnMissing["return error: missing required variables"]
    hasMissing -->|Não| doRender["manager.RenderTemplate(name, variables)"]
    
    doRender --> renderErr["Erro?"]
    renderErr -->|Sim| returnRender["return error: failed to render"]
    renderErr -->|Não| content["content = resultado"]
    
    content --> outCheck{"output != ''?"}
    outCheck -->|Sim| writeFile["os.WriteFile(output, content, 0600)"]
    outCheck -->|Não| printStdout["fmt.Print(content)"]
    
    writeFile --> writeErr["Erro?"]
    writeErr -->|Sim| returnWrite["return error: failed to write output"]
    writeErr -->|Não| return("return nil")
    
    printStdout --> return
    returnMissing --> return
    returnRender --> return
```

---

## Diagrama de Fluxo: template import

```mermaid
flowchart TD
    Start([template import]) --> runImport["templateImportCmd.RunE()"]
    
    runImport --> argsCheck{"len(args) >= 1?"}
    argsCheck -->|Não| errArgs["Erro: cobra.RangeArgs(1,2) falha"]
    argsCheck -->|Sim| readFile["os.ReadFile(filePath)"]
    
    errArgs --> returnErr([return error])
    
    readFile --> readErr["Erro ao ler?"]
    readErr -->|Sim| errRead["Erro: failed to read template file"]
    readErr -->|Não| emptyCheck{"content vazio?"}
    
    errRead --> returnErr
    
    emptyCheck -->|Sim| errEmpty["Erro: template content is empty"]
    emptyCheck -->|Não| extractName["extrair nome do filename
      templateName = args[1] ou
      templateName = filename sem extensão
      (remover prefixo 'prompt_')]"]
    
    errEmpty --> returnErr
    
    extractName --> userDir["xdg.ConfigHome/shotgun-cli/templates/"]
    userDir --> mkdirAll["os.MkdirAll(userTemplatesDir, 0750)"]
    
    mkdirAll --> mkdirErr["Erro ao criar dir?"]
    mkdirErr -->|Sim| errMkdir["Erro: failed to create user templates directory"]
    mkdirErr -->|Não| destPath["filepath.Join(userDir, templateName + '.md')"]
    
    errMkdir --> returnErr
    
    destPath --> existsCheck{"os.Stat(destPath) existe?"}
    existsCheck -->|Não| writeNew["os.WriteFile(destPath, content, 0600)"]
    existsCheck -->|Sim| promptConfirm["Print: 'Template <name> already exists. Overwrite? (y/N):'"]
    
    promptConfirm --> reader["bufio.NewReader(os.Stdin)"]
    reader --> readResponse["response = reader.ReadString('\\n')"]
    
    readResponse --> readErr2["Erro ao ler?"]
    readErr2 -->|Sim| errReadResp["Erro: failed to read user input"]
    readErr2 -->|Não| normalize["TrimSpace, toLower(response)"]
    
    normalize --> confirm{"response == 'y' ou 'yes'?"}
    confirm -->|Não| canceled["Print: 'Import canceled.'"]
    confirm -->|Sim| writeNew
    
    errReadResp --> returnErr
    canceled --> endImport(["fim"])
    
    writeNew --> writeErr2["Erro ao escrever?"]
    writeErr2 -->|Sim| errWrite["Erro: failed to write template file"]
    writeErr2 -->|Não| printSuccess["Print:
  ✅ Template '<name>' imported successfully to: <destPath>
  Use 'shotgun-cli template list' to see all templates"]
    
    errWrite --> returnErr
    printSuccess --> endImport
```

---

## Diagrama de Fluxo: template export

```mermaid
flowchart TD
    Start([template export]) --> runExport["templateExportCmd.RunE()"]
    
    runExport --> argsCheck{"len(args) == 2?"}
    argsCheck -->|Não| errArgs["Erro: cobra.ExactArgs(2) falha"]
    argsCheck -->|Sim| getTemplate["templateName = args[0], outputPath = args[1]"]
    
    errArgs --> returnErr([return error])
    
    getTemplate --> newManager["template.NewManager(CustomPath)"]
    
    newManager --> mgrErr["Erro Manager?"]
    mgrErr -->|Sim| errMgr["Erro: failed to initialize template manager"]
    mgrErr -->|Não| getTmpl["manager.GetTemplate(templateName)"]
    
    errMgr --> returnErr
    
    getTmpl --> tmplErr["Template não encontrado?"]
    tmplErr -->|Sim| errTmpl["Erro: template '<name>' not found"]
    tmplErr -->|Não| tmpl["Template {Name, Content, Source, Description}"]
    
    errTmpl --> returnErr
    
    tmpl --> existsCheck{"os.Stat(outputPath) existe?"}
    existsCheck -->|Não| createDir["os.MkdirAll(filepath.Dir(outputPath), 0750)"]
    existsCheck -->|Sim| promptConfirm["Print: 'File <output> already exists. Overwrite? (y/N):'"]
    
    promptConfirm --> reader["bufio.NewReader(os.Stdin)"]
    reader --> readResponse["response = reader.ReadString('\\n')"]
    
    readResponse --> readErr2["Erro ao ler?"]
    readErr2 -->|Sim| errReadResp["Erro: failed to read user input"]
    readErr2 -->|Não| normalize["TrimSpace, toLower(response)"]
    
    normalize --> confirm{"response == 'y' ou 'yes'?"}
    confirm -->|Não| canceled["Print: 'Export canceled.'"]
    confirm -->|Sim| createDir
    
    errReadResp --> returnErr
    canceled --> endExport(["fim"])
    
    createDir --> mkdirErr["Erro ao criar dir?"]
    mkdirErr -->|Sim| errMkdir["Erro: failed to create output directory"]
    mkdirErr -->|Não| writeTmpl["os.WriteFile(outputPath, tmpl.Content, 0600)"]
    
    errMkdir --> returnErr
    
    writeTmpl --> writeErr["Erro ao escrever?"]
    writeErr -->|Sim| errWrite["Erro: failed to write template file"]
    writeErr -->|Não| printSuccess["Print:
  ✅ Template '<name>' exported successfully to: <outputPath>"]
    
    errWrite --> returnErr
    printSuccess --> endExport
```

---

## Resumo de Flags

| Comando | Flag | Tipo | Padrão | Obrigatório | Descrição |
|---|---|---|---|---|---|
| `template list` | — | — | — | — | Listar templates |
| `template render` | `[template-name]` | positional | — | Sim | Nome do template |
| `template render` | `--var` | stringToString | `{}` | Não | Variáveis KEY=VALUE |
| `template render` | `--output` / `-o` | string | `""` | Não | Arquivo de saída |
| `template import` | `<file>` | positional | — | Sim | Arquivo de template |
| `template import` | `[name]` | positional | auto | Não | Nome (auto do filename) |
| `template export` | `<name>` | positional | — | Sim | Nome do template |
| `template export` | `<file>` | positional | — | Sim | Arquivo de destino |

---

## Resumo de Variáveis de Configuração

| Chave Viper | Uso |
|---|---|
| `template.custom-path` | Caminho para templates customizados no Manager |

---

## Observações Importantes

🟡 **Inconsistência de caminho:** `template import` usa `xdg.ConfigHome` hardcoded para determinar o diretório de templates, enquanto `config show/set` usa `getDefaultConfigPath()` (que respeita a plataforma). Isso pode levar a caminhos diferentes em macOS e Windows.

🟡 **Validação de template:** O import não usa `template.NewManager()` para validação do conteúdo — usa validação básica (não vazio). A validação completa de template variables é feita apenas no render.
