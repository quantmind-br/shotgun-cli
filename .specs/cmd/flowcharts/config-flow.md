# Fluxo: Config (show / set / TUI)

> **Arquivo fonte:** `config.go`  
> **Comandos:** `shotgun-cli config`, `shotgun-cli config show`, `shotgun-cli config set`

---

## Diagrama de Fluxo: config show

```mermaid
flowchart TD
    Start([config show]) --> showCurrentConfig["showCurrentConfig()"]
    
    showCurrentConfig --> configPath{"viper.ConfigFileUsed() != ''?"}
    
    configPath -->|Sim| printPath["Print: 'Config file: <path>'"]
    configPath -->|Não| printDefault["Print: 'Config file: Not found (using defaults)'"]
    
    printPath --> allKeys["viper.AllKeys() → []string"]
    printDefault --> allKeys
    
    allKeys --> sortKeys["sort.Strings(allKeys)"]
    sortKeys --> groupKeys["Agrupar por categoria (primeiro segmento da chave)"]
    
    groupKeys --> categoryOrder["scanner, context, template, output, llm"]
    categoryOrder --> displayLoop["Iterar sobre categoryOrder"]
    
    displayLoop --> hasCategory{"Categoria existe em keys?"}
    hasCategory -->|Sim| printCategory["Print: [CATEGORY] (styled)"]
    hasCategory -->|Não| nextCat["Próxima categoria"]
    
    printCategory --> printKey["Iterar keys da categoria:
      Print: key = value (source)"]
    printKey --> nextKey["Próxima key?"]
    nextKey -->|Sim| printKey
    nextKey -->|Não| nextCat2["Próxima categoria"]
    
    nextCat --> nextCat2
    
    nextCat2 --> remainingCategories["Categorias não na ordem pré-definida?"]
    remainingCategories -->|Sim| printRemaining["Print na ordem iterada (map iteration order)"]
    remainingCategories -->|Não| doneShow(["fim"])
    
    printRemaining --> doneShow
```

---

## Diagrama de Fluxo: config set

```mermaid
flowchart TD
    Start([config set]) --> preRunE["PreRunE: cobra.ExactArgs(2)"]
    
    preRunE --> getKey["key = args[0]"]
    getKey --> getValue["value = args[1]"]
    
    getValue --> isValidKey{"config.IsValidKey(key)?"}
    isValidKey -->|Não| errInvalidKey["Erro: invalid configuration key '<key>'"]
    isValidKey -->|Sim| validateValue{"config.ValidateValue(key, value)?"}
    
    errInvalidKey --> returnErr([return error])
    
    validateValue -->|Não| errInvalidValue["Erro: invalid value for '<key>'"]
    validateValue -->|Sim| runSet["setConfigValue(key, value)"]
    
    errInvalidValue --> returnErr
    
    runSet --> convertValue["config.ConvertValue(key, value) → converted"]
    convertValue --> convertErr["Erro de conversão?"]
    convertErr -->|Sim| errConvert["Erro de conversão"]
    convertErr -->|Não| setViper["viper.Set(key, converted)"]
    
    errConvert --> returnErr
    
    setViper --> configPath2{"viper.ConfigFileUsed() != ''?"}
    configPath2 -->|Sim| useUsedPath["configPath = ConfigFileUsed()"]
    configPath2 -->|Não| getDefault["configPath = getDefaultConfigPath()
      viper.SetConfigFile(configPath)"]
    
    useUsedPath --> ensureDir["os.MkdirAll(filepath.Dir(configPath), 0750)"]
    getDefault --> ensureDir
    
    ensureDir --> mkdirErr["Erro ao criar dir?"]
    mkdirErr -->|Sim| errMkdir["Erro: failed to create config directory"]
    
    mkdirErr -->|Não| writeConfig["viper.WriteConfig()"]
    
    writeConfig --> writeErr["Erro ao escrever?"]
    writeErr -->|Sim| checkNotExist{"os.IsNotExist?"}
    checkNotExist -->|Sim| safeWrite["viper.SafeWriteConfig()"]
    checkNotExist -->|Não| errWriteConfig["Erro: failed to write config file"]
    
    writeErr -->|Não| logUpdate["log.Debug: 'Configuration updated'"]
    
    safeWrite --> safeWriteErr["Erro no SafeWrite?"]
    safeWriteErr -->|Sim| errSafeWrite["Erro: failed to create config file"]
    safeWriteErr -->|Não| logUpdate
    
    errMkdir --> returnErr
    errWriteConfig --> returnErr
    errSafeWrite --> returnErr
    
    logUpdate --> printSuccess["Print:
  ✅ Configuration updated successfully!
  📝 Set <key> = <value>
  📁 Config file: <path>"]
    
    printSuccess --> templateMsg{"key == 'template.custom-path' E value != ''?"}
    templateMsg -->|Sim| templateHint["Print: 'The custom template directory will be created automatically'"]
    templateMsg -->|Não| endShow(["fim"])
    
    templateHint --> endShow
```

---

## Diagrama de Fluxo: config TUI

```mermaid
flowchart TD
    Start([config - sem subcomando]) --> launchConfigTUI["launchConfigTUI()"]
    
    launchConfigTUI --> newWizard["ui.NewConfigWizard()"]
    newWizard --> teaProgram["tea.NewProgram(wizard, altScreen, mouseCellMotion)"]
    teaProgram --> teaRun["program.Run()"]
    
    teaRun --> teaErr["Erro ao executar?"]
    teaErr -->|Sim| errTUI["Erro: failed to start config TUI"]
    teaErr -->|Não| tuiDone([fim])
    
    errTUI --> returnErr([return error])
```

---

## Resumo de Flags

| Comando | Flag | Tipo | Padrão | Descrição |
|---|---|---|---|---|
| `config` | — | — | — | Lança TUI interativo |
| `config show` | — | — | — | Exibe configuração |
| `config set` | `[key]` | positional | — | Chave de configuração |
| `config set` | `[value]` | positional | — | Valor |

---

## Categorias de Configuração

| Categoria | Chaves |
|---|---|
| `scanner` | `scanner.max-files`, `scanner.max-file-size`, `scanner.max-memory`, `scanner.respect-gitignore`, `scanner.skip-binary`, `scanner.workers`, `scanner.include-hidden`, `scanner.respect-shotgunignore` |
| `context` | `context.max-size`, `context.include-tree`, `context.include-summary` |
| `template` | `template.custom-path` |
| `output` | `output.format`, `output.clipboard` |
| `llm` | `llm.provider`, `llm.api-key`, `llm.base-url`, `llm.model`, `llm.timeout`, `llm.save-response` |
| `global` | `verbose`, `quiet` |

---

## Mapeamento de Tipos por Chave

| Chave | Tipo Convertido | Exemplos Válidos |
|---|---|---|
| `scanner.max-files` | `int64` | `10000` |
| `scanner.max-file-size` | `string` | `"1MB"` |
| `scanner.max-memory` | `string` | `"500MB"` |
| `scanner.respect-gitignore` | `bool` | `true`, `false` |
| `scanner.skip-binary` | `bool` | `true`, `false` |
| `scanner.workers` | `int` | `4` |
| `context.max-size` | `string` | `"10MB"` |
| `context.include-tree` | `bool` | `true`, `false` |
| `output.format` | `string` | `markdown`, `text` |
| `output.clipboard` | `bool` | `true`, `false` |
| `llm.provider` | `string` | `openai`, `anthropic`, `gemini` |
| `llm.api-key` | `string` | `sk-your-key` |
