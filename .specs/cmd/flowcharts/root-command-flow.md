# Fluxo: Root Command & TUI Wizard

> **Arquivo fonte:** `root.go`  
> **Comando:** `shotgun-cli` (sem argumentos) / `shotgun-cli --help` / `shotgun-cli --version`

---

## Diagrama de Fluxo

```mermaid
flowchart TD
    Start([main.go: cmd.Execute()]) --> Execute["Execute()"]
    Execute --> rootCmdExec["rootCmd.Execute()"]
    rootCmdExec --> cobraInit["cobra.OnInitialize(initConfig)"]
    
    cobraInit --> initConfig["initConfig()"]
    
    initConfig --> cfgFlag["cfgFile definido via flag --config?"]
    cfgFlag -->|Sim| useFlag["viper.SetConfigFile(cfgFile)"]
    cfgFlag -->|Não| searchPaths["Search paths: getConfigDir(), ~/.config, ."]
    searchPaths --> envPrefix["viper.SetEnvPrefix('SHOTGUN') + AutomaticEnv()"]
    
    useFlag --> envPrefix
    
    envPrefix --> setDefaults["setConfigDefaults()"]
    
    setDefaults --> bindFlags["Bind PFlags: verbose, quiet"]
    bindFlags --> readConfig["viper.ReadInConfig()"]
    
    readConfig --> configFileFound["Config file encontrado?"]
    configFileFound -->|Não| useDefaults["Usar defaults apenas"]
    configFileFound -->|Sim| logConfig["log: 'Using config file'"]
    
    useDefaults --> updateLogging["updateLoggingLevel()"]
    logConfig --> updateLogging
    
    updateLogging --> loggingReady[Logging ready]
    
    rootCmdExec --> runRoot["runRootCommand(cmd, args)"]
    
    runRoot --> versionFlag["Flags --version?"]
    versionFlag -->|Sim| printVersion["Print: shotgun-cli version X.Y.Z"]
    versionFlag -->|Não| checkArgs["len(args) == 0 && len(os.Args) == 1?"]
    
    printVersion --> end([fim])
    
    checkArgs -->|Sim| launchTUI["launchTUIWizard()"]
    checkArgs -->|Não| showHelp["_ = cmd.Help()"]
    
    showHelp --> end
    
    launchTUI --> getWd["os.Getwd() → rootPath"]
    getWd --> wdErr["Falha ao obter cwd?"]
    wdErr -->|Sim| showError1["Error: Could not determine directory"]
    wdErr -->|Não| buildScanCfg["Build ScanConfig from Viper"]
    
    showError1 --> exit1["os.Exit(1)"]
    exit1 --> end
    
    buildScanCfg --> buildWizardCfg["Build WizardConfig from Viper"]
    buildWizardCfg --> newWizard["ui.NewWizard(rootPath, scanConfig, wizardConfig, nil)"]
    newWizard --> teaProgram["tea.NewProgram(wizard, altScreen, mouseCellMotion)"]
    teaProgram --> teaRun["program.Run()"]
    teaRun --> teaErr["Falha ao executar wizard?"]
    teaErr -->|Sim| showError2["Error starting wizard"]
    teaErr -->|Não| wizardDone["Wizard concluído"]
    
    showError2 --> exit2["os.Exit(1)"]
    exit2 --> end
    wizardDone --> end
```

---

## Hierarquia de Comandos Cobra

```
shotgun-cli (rootCmd)
├── [sem subcomando]     → runRootCommand → TUI Wizard
├── --help               → Show help
├── --version            → Show version
├── --config             → Config file path
├── -v                   → Verbose (Debug)
├── -q                   → Quiet (Error only)
├── context              → contextCmd
│   ├── generate         → contextGenerateCmd
│   └── send             → contextSendCmd
├── config               → configCmd
│   ├── [sem subcmd]     → launchConfigTUI
│   ├── show             → showCurrentConfig
│   └── set              → configSetCmd
├── llm                  → llmCmd
│   ├── status           → llmStatusCmd
│   ├── doctor           → llmDoctorCmd
│   └── list             → llmListCmd
├── template             → templateCmd
│   ├── list             → templateListCmd
│   ├── render           → templateRenderCmd
│   ├── import           → templateImportCmd
│   └── export           → templateExportCmd
├── diff                 → diffCmd
│   └── split            → diffSplitCmd
└── completion           → completionCmd
```

---

## Detalhes da Inicialização de Configuração

### Ordem de Prioridade Viper (da menor para a maior)

| Prioridade | Fonte | Exemplo |
|---|---|---|
| 1 | Defaults (setConfigDefaults) | `scanner.max-files = 10000` |
| 2 | Config file YAML | `~/.config/shotgun-cli/config.yaml` |
| 3 | Environment variables | `SHOTGUN_SCANNER_MAX_FILES=5000` |
| 4 | Flags CLI | `--verbose` |

### Caminhos de Configuração por Plataforma

| SO | Caminho | Variável Ambiente |
|---|---|---|
| Linux | `$XDG_CONFIG_HOME/shotgun-cli/config.yaml` | `XDG_CONFIG_HOME` |
| Linux (fallback) | `~/.config/shotgun-cli/config.yaml` | — |
| macOS | `~/Library/Application Support/shotgun-cli/config.yaml` | — |
| Windows | `%APPDATA%/shotgun-cli/config.yaml` | `APPDATA` |
