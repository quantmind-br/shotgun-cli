# Fluxo: Shell Completion

> **Arquivo fonte:** `completion.go`  
> **Comando:** `shotgun-cli completion [bash\|zsh\|fish\|powershell]`

---

## Diagrama de Fluxo

```mermaid
flowchart TD
    Start([completion]) --> completionCmd["completionCmd"]
    
    completionCmd --> validArgs["cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)"]
    validArgs --> validShells{"args[0] em ['bash', 'zsh', 'fish', 'powershell']?"}
    validShells -->|Não| errShell["Erro: cobra.ExactArgs + OnlyValidArgs falha"]
    validShells -->|Sim| shellSwitch["switch args[0]"]
    
    errShell --> returnErr([return error])
    
    shellSwitch --> caseBash["'bash': rootCmd.GenBashCompletion(os.Stdout)"]
    shellSwitch --> caseZsh["'zsh': rootCmd.GenZshCompletion(os.Stdout)"]
    shellSwitch --> caseFish["'fish': rootCmd.GenFishCompletion(os.Stdout, true)"]
    shellSwitch --> casePwsh["'powershell': rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)"]
    shellSwitch --> caseDefault["default: fmt.Errorf('unsupported shell: <shell>')"]
    
    caseBash --> endComp(["fim"])
    caseZsh --> endComp
    caseFish --> endComp
    casePwsh --> endComp
    caseDefault --> errDefault["Erro: unsupported shell"]
    errDefault --> returnErr
```

---

## Diagrama de Fluxo: Config Key Completion

```mermaid
flowchart TD
    A["configKeyCompletion(cmd, args, toComplete)"] --> checkArgs{"len(args) > 0?"}
    checkArgs -->|Sim| noResult["return nil, NoFileComp"]
    checkArgs -->|Não| keys["Retornar lista de config keys:
  - scanner.max-files (Maximum number of files to scan)
  - scanner.max-file-size (Maximum size per file)
  - scanner.respect-gitignore (Respect .gitignore files)
  - scanner.skip-binary (Skip binary files)
  - scanner.workers (Number of parallel workers)
  - scanner.include-hidden (Include hidden files)
  - scanner.include-ignored (Include ignored files)
  - scanner.respect-shotgunignore (Respect .shotgunignore files)
  - scanner.max-memory (Max memory usage)
  - context.max-size (Maximum context size)
  - context.include-tree (Include directory tree)
  - context.include-summary (Include file summaries)
  - template.custom-path (Path to custom templates)
  - output.format (Output format)
  - output.clipboard (Copy to clipboard)"]
    
    noResult --> R([return])
    keys --> R
```

---

## Diagrama de Fluxo: Bool Value Completion

```mermaid
flowchart TD
    A["boolValueCompletion(cmd, args, toComplete)"] --> checkArgs1{"len(args) == 1?"}
    checkArgs1 -->|Não| noResult1["return nil, NoFileComp"]
    checkArgs1 -->|Sim| key["key = args[0]"]
    
    key --> checkBool{"key em chaves booleanas?"}
    checkBool -->|Sim| bools["return ['true', 'false'], NoFileComp"]
    checkBool -->|Não| checkFormat{"key == 'output.format'?"}
    
    checkFormat -->|Sim| formats["return ['markdown', 'text'], NoFileComp"]
    checkFormat -->|Não| checkPath{"key == 'template.custom-path'?"}
    
    checkPath -->|Sim| fileComp["return nil, Default (file completion)"]
    checkPath -->|Não| noResult2["return nil, NoFileComp"]
    
    bools --> R([return])
    formats --> R
    fileComp --> R
    noResult1 --> R
    noResult2 --> R
```

---

## Registro de Completers

| Comando | Completador | Função | Quando é chamado |
|---|---|---|---|
| `template render` | Template names | `ValidArgsFunction` | No 1º argumento |
| `config set` | Config keys | `configKeyCompletion` | No 1º argumento |
| `config set` | Values | `boolValueCompletion` | No 2º argumento |

---

## Observações

🟡 **Order dependency:** O completer de `config set` é registrado em `init()` após a verificação `if configSetCmd != nil`, o que depende que `configCmd` e `configSetCmd` já tenham sido criados (ordem de `init()` nos arquivos Go).
