# Plano de Implementacao: Reduce Logic in Command Layer

## Resumo Executivo

Este plano move a logica de construcao de configuracao de `cmd/context.go` para a camada de aplicacao (`internal/app`), seguindo os principios de Clean Architecture. A mudanca permite testar a logica de configuracao sem dependencia do Cobra e torna os componentes de progresso reutilizaveis para outros comandos CLI.

## Analise de Requisitos

### Requisitos Funcionais
- [ ] Mover tipo `GenerateConfig` de `cmd/` para `internal/app/`
- [ ] Criar `GenerateFlags` struct para encapsular flags do Cobra
- [ ] Criar `NewGenerateConfigFromFlags()` factory em `internal/app/`
- [ ] Mover tipos de progresso (`ProgressMode`, `ProgressOutput`) para `internal/app/`
- [ ] Mover funcoes de renderizacao de progresso para `internal/ui/cli/`
- [ ] Simplificar `cmd/context.go` para orquestracao apenas

### Requisitos Nao-Funcionais
- [ ] Manter compatibilidade total com CLI existente
- [ ] Cobertura de testes para nova logica em `internal/app/`
- [ ] Zero breaking changes para usuarios

## Analise Tecnica

### Arquitetura Atual (Problematica)

```
cmd/context.go (522 linhas)
├── GenerateConfig struct          # Deveria estar em app/
├── ProgressMode, ProgressOutput   # Deveria estar em app/ ou ui/cli/
├── buildGenerateConfig()          # ~100 linhas de logica de negocios
├── buildScannerConfig()           # ~25 linhas
├── buildTemplateVars()            # ~18 linhas  
├── loadTemplateContent()          # ~19 linhas
├── renderProgress*()              # ~30 linhas
└── printGenerationSummary()       # ~10 linhas
```

### Arquitetura Proposta

```
cmd/context.go (~150 linhas)
├── extractFlags() -> GenerateFlags
├── RunE: orquestracao simples
└── init(): definicao de flags

internal/app/cli_config.go (EXISTE PARCIALMENTE)
├── GenerateFlags struct           # Input do Cobra
├── NewGenerateConfigFromFlags()   # Factory com toda logica
├── buildScannerConfig()           # Movido
├── buildTemplateVars()            # Movido
└── loadTemplateContent()          # Movido

internal/app/config.go (JA EXISTE)
├── CLIConfig struct               # JA EXISTE - unificar com GenerateConfig
├── ProgressMode                   # JA EXISTE
├── ProgressOutput                 # JA EXISTE
└── GenerateConfig                 # MOVER para ca

internal/ui/cli/progress.go (CRIAR)
├── RenderProgress()
├── RenderProgressHuman()
├── RenderProgressJSON()
└── ClearProgressLine()
```

### Componentes Afetados

| Arquivo/Modulo | Tipo de Mudanca | Descricao |
|----------------|-----------------|-----------|
| `internal/app/config.go` | Modificar | Adicionar `GenerateFlags`, unificar configs |
| `internal/app/cli_config.go` | Criar | `NewGenerateConfigFromFlags()` e helpers |
| `internal/app/cli_config_test.go` | Criar | Testes para factory |
| `internal/ui/cli/progress.go` | Criar | Funcoes de renderizacao |
| `internal/ui/cli/progress_test.go` | Criar | Testes para renderizacao |
| `cmd/context.go` | Modificar | Simplificar para orquestracao |
| `cmd/context_test.go` | Modificar | Atualizar para nova estrutura |

### Dependencias
- `github.com/spf13/viper` - Acesso a config (ja usado em app/)
- `internal/config` - Keys de configuracao
- `internal/core/scanner` - `ScanConfig`
- `internal/core/template` - Template manager
- `internal/utils` - `ParseSize`, `ParseSizeWithDefault`

## Plano de Implementacao

### Fase 1: Consolidar Tipos em app/config.go

**Objetivo**: Unificar `CLIConfig` existente com `GenerateConfig` de cmd/

#### Tarefas:

1. **Analisar duplicacao entre `CLIConfig` e `GenerateConfig`**
   
   `internal/app/config.go` ja possui:
   ```go
   type CLIConfig struct {
       RootPath     string
       Include      []string
       Exclude      []string
       Output       string
       MaxSize      int64
       EnforceLimit bool
       // ... outros campos
       ProgressMode ProgressMode
   }
   ```
   
   `cmd/context.go` possui:
   ```go
   type GenerateConfig struct {
       RootPath      string
       Include       []string
       Exclude       []string
       Output        string
       MaxSize       int64
       EnforceLimit  bool
       // ... outros campos identicos
       ProgressMode ProgressMode  // duplicado!
   }
   ```

2. **Remover duplicacao em `internal/app/config.go`**
   
   Arquivos envolvidos: `internal/app/config.go`
   
   ```go
   package app

   import (
       "github.com/quantmind-br/shotgun-cli/internal/core/scanner"
   )

   // ProgressMode defines how progress is reported during CLI operations.
   type ProgressMode string

   const (
       ProgressNone  ProgressMode = "none"
       ProgressHuman ProgressMode = "human"
       ProgressJSON  ProgressMode = "json"
   )

   // ProgressOutput represents a progress event for output.
   type ProgressOutput struct {
       Timestamp string  `json:"timestamp"`
       Stage     string  `json:"stage"`
       Message   string  `json:"message"`
       Current   int64   `json:"current,omitempty"`
       Total     int64   `json:"total,omitempty"`
       Percent   float64 `json:"percent,omitempty"`
   }

   // GenerateFlags represents the raw CLI flags before processing.
   // This struct is filled directly from Cobra flag values.
   type GenerateFlags struct {
       RootPath       string
       Include        []string
       Exclude        []string
       Output         string
       MaxSize        string // String format, needs parsing
       EnforceLimit   bool
       Template       string
       Task           string
       Rules          string
       CustomVars     []string // KEY=VALUE format
       Workers        int
       IncludeHidden  bool
       IncludeIgnored bool
       ProgressMode   string // String format
       // Gemini-specific
       SendGemini    bool
       GeminiModel   string
       GeminiOutput  string
       GeminiTimeout int
   }

   // CLIGenerateConfig is the processed configuration for CLI context generation.
   // All values are validated and converted to their final types.
   type CLIGenerateConfig struct {
       RootPath      string
       Include       []string
       Exclude       []string
       Output        string
       MaxSize       int64 // Parsed from string
       EnforceLimit  bool
       Template      string
       Task          string
       Rules         string
       CustomVars    map[string]string // Parsed from []string
       Workers       int
       IncludeHidden  bool
       IncludeIgnored bool
       ProgressMode  ProgressMode // Parsed from string
       // Gemini-specific
       SendGemini    bool
       GeminiModel   string
       GeminiOutput  string
       GeminiTimeout int
   }
   ```

### Fase 2: Criar Factory em app/cli_config.go

**Objetivo**: Mover logica de `buildGenerateConfig()` para app layer

#### Tarefas:

1. **Criar `internal/app/cli_config.go`**
   
   ```go
   package app

   import (
       "fmt"
       "path/filepath"
       "strings"
       "time"

       "github.com/spf13/viper"

       cfgkeys "github.com/quantmind-br/shotgun-cli/internal/config"
       "github.com/quantmind-br/shotgun-cli/internal/core/scanner"
       "github.com/quantmind-br/shotgun-cli/internal/core/template"
       "github.com/quantmind-br/shotgun-cli/internal/utils"
   )

   // NewGenerateConfigFromFlags creates a CLIGenerateConfig from raw CLI flags.
   // This function handles all parsing, validation, and default application.
   func NewGenerateConfigFromFlags(flags GenerateFlags) (*CLIGenerateConfig, error) {
       // Parse progress mode
       progressMode, err := parseProgressMode(flags.ProgressMode)
       if err != nil {
           return nil, err
       }

       // Parse custom variables
       customVars, err := parseCustomVars(flags.CustomVars)
       if err != nil {
           return nil, err
       }

       // Convert root to absolute path
       absPath, err := filepath.Abs(flags.RootPath)
       if err != nil {
           return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
       }

       // Parse max size
       maxSize, err := utils.ParseSize(flags.MaxSize)
       if err != nil {
           return nil, fmt.Errorf("failed to parse max-size: %w", err)
       }

       // Generate default output filename if not specified
       output := flags.Output
       if output == "" {
           timestamp := time.Now().Format("20060102-150405")
           output = fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
       }

       // Apply Gemini defaults from config
       geminiModel := flags.GeminiModel
       if geminiModel == "" {
           geminiModel = viper.GetString(cfgkeys.KeyGeminiModel)
       }

       geminiTimeout := flags.GeminiTimeout
       if geminiTimeout == 0 {
           geminiTimeout = viper.GetInt(cfgkeys.KeyGeminiTimeout)
       }

       // Enable gemini via config if auto-send is enabled
       sendGemini := flags.SendGemini
       if !sendGemini && viper.GetBool(cfgkeys.KeyGeminiAutoSend) && viper.GetBool(cfgkeys.KeyGeminiEnabled) {
           sendGemini = true
       }

       return &CLIGenerateConfig{
           RootPath:       absPath,
           Include:        flags.Include,
           Exclude:        flags.Exclude,
           Output:         output,
           MaxSize:        maxSize,
           EnforceLimit:   flags.EnforceLimit,
           Template:       flags.Template,
           Task:           flags.Task,
           Rules:          flags.Rules,
           CustomVars:     customVars,
           Workers:        flags.Workers,
           IncludeHidden:  flags.IncludeHidden,
           IncludeIgnored: flags.IncludeIgnored,
           ProgressMode:   progressMode,
           SendGemini:     sendGemini,
           GeminiModel:    geminiModel,
           GeminiOutput:   flags.GeminiOutput,
           GeminiTimeout:  geminiTimeout,
       }, nil
   }

   func parseProgressMode(s string) (ProgressMode, error) {
       switch s {
       case "none", "":
           return ProgressNone, nil
       case "human":
           return ProgressHuman, nil
       case "json":
           return ProgressJSON, nil
       default:
           return "", fmt.Errorf("invalid progress mode: %q (expected: none, human, json)", s)
       }
   }

   func parseCustomVars(vars []string) (map[string]string, error) {
       result := make(map[string]string)
       for _, v := range vars {
           parts := strings.SplitN(v, "=", 2)
           if len(parts) != 2 {
               return nil, fmt.Errorf("invalid --var format: %q (expected KEY=VALUE)", v)
           }
           result[parts[0]] = parts[1]
       }
       return result, nil
   }

   // BuildScannerConfig creates a scanner.ScanConfig from CLIGenerateConfig.
   func BuildScannerConfig(cfg *CLIGenerateConfig) scanner.ScanConfig {
       scannerConfig := scanner.ScanConfig{
           MaxFiles:             viper.GetInt64(cfgkeys.KeyScannerMaxFiles),
           MaxFileSize:          utils.ParseSizeWithDefault(viper.GetString(cfgkeys.KeyScannerMaxFileSize), 1024*1024),
           MaxMemory:            utils.ParseSizeWithDefault(viper.GetString(cfgkeys.KeyScannerMaxMemory), 500*1024*1024),
           SkipBinary:           viper.GetBool(cfgkeys.KeyScannerSkipBinary),
           IncludeHidden:        viper.GetBool(cfgkeys.KeyScannerIncludeHidden),
           IncludeIgnored:       viper.GetBool(cfgkeys.KeyScannerIncludeIgnored),
           Workers:              viper.GetInt(cfgkeys.KeyScannerWorkers),
           RespectGitignore:     viper.GetBool(cfgkeys.KeyScannerRespectGitignore),
           RespectShotgunignore: viper.GetBool(cfgkeys.KeyScannerRespectShotgunignore),
           IgnorePatterns:       cfg.Exclude,
           IncludePatterns:      cfg.Include,
       }

       // Apply CLI overrides
       if cfg.Workers > 0 {
           scannerConfig.Workers = cfg.Workers
       }
       if cfg.IncludeHidden {
           scannerConfig.IncludeHidden = true
       }
       if cfg.IncludeIgnored {
           scannerConfig.IncludeIgnored = true
       }

       return scannerConfig
   }

   // BuildTemplateVars creates template variables from CLIGenerateConfig.
   func BuildTemplateVars(cfg *CLIGenerateConfig) map[string]string {
       taskValue := cfg.Task
       if taskValue == "" {
           taskValue = "Context generation"
       }

       vars := map[string]string{
           "TASK":           taskValue,
           "RULES":          cfg.Rules,
           "FILE_STRUCTURE": "",
           "CURRENT_DATE":   time.Now().Format("2006-01-02"),
       }

       for k, v := range cfg.CustomVars {
           vars[k] = v
       }

       return vars
   }

   // LoadTemplateContent loads template content by name.
   func LoadTemplateContent(templateName string) (string, error) {
       if templateName == "" {
           return "", nil
       }

       tmplMgr, err := template.NewManager(template.ManagerConfig{
           CustomPath: viper.GetString(cfgkeys.KeyTemplateCustomPath),
       })
       if err != nil {
           return "", fmt.Errorf("failed to initialize template manager: %w", err)
       }

       tmpl, err := tmplMgr.GetTemplate(templateName)
       if err != nil {
           return "", fmt.Errorf("failed to load template %q: %w", templateName, err)
       }

       return tmpl.Content, nil
   }
   ```

2. **Criar testes para a factory**
   
   Arquivos envolvidos: `internal/app/cli_config_test.go`
   
   ```go
   package app

   import (
       "testing"

       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/require"
   )

   func TestNewGenerateConfigFromFlags_Basic(t *testing.T) {
       t.Parallel()

       flags := GenerateFlags{
           RootPath:     "/tmp/test",
           Include:      []string{"*.go"},
           Exclude:      []string{"vendor/*"},
           MaxSize:      "10MB",
           EnforceLimit: true,
       }

       cfg, err := NewGenerateConfigFromFlags(flags)

       require.NoError(t, err)
       assert.Equal(t, "/tmp/test", cfg.RootPath)
       assert.Equal(t, []string{"*.go"}, cfg.Include)
       assert.Equal(t, int64(10*1024*1024), cfg.MaxSize)
       assert.NotEmpty(t, cfg.Output) // Auto-generated
   }

   func TestNewGenerateConfigFromFlags_ProgressModes(t *testing.T) {
       t.Parallel()

       tests := []struct {
           name     string
           input    string
           expected ProgressMode
           wantErr  bool
       }{
           {"empty", "", ProgressNone, false},
           {"none", "none", ProgressNone, false},
           {"human", "human", ProgressHuman, false},
           {"json", "json", ProgressJSON, false},
           {"invalid", "xml", "", true},
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               t.Parallel()

               flags := GenerateFlags{
                   RootPath:     "/tmp",
                   MaxSize:      "1MB",
                   ProgressMode: tt.input,
               }

               cfg, err := NewGenerateConfigFromFlags(flags)

               if tt.wantErr {
                   require.Error(t, err)
                   return
               }
               require.NoError(t, err)
               assert.Equal(t, tt.expected, cfg.ProgressMode)
           })
       }
   }

   func TestParseCustomVars(t *testing.T) {
       t.Parallel()

       tests := []struct {
           name     string
           input    []string
           expected map[string]string
           wantErr  bool
       }{
           {
               name:     "empty",
               input:    []string{},
               expected: map[string]string{},
           },
           {
               name:     "single var",
               input:    []string{"KEY=value"},
               expected: map[string]string{"KEY": "value"},
           },
           {
               name:     "value with equals",
               input:    []string{"KEY=value=with=equals"},
               expected: map[string]string{"KEY": "value=with=equals"},
           },
           {
               name:    "invalid format",
               input:   []string{"INVALID"},
               wantErr: true,
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               t.Parallel()

               result, err := parseCustomVars(tt.input)

               if tt.wantErr {
                   require.Error(t, err)
                   return
               }
               require.NoError(t, err)
               assert.Equal(t, tt.expected, result)
           })
       }
   }
   ```

### Fase 3: Criar Modulo de Progresso CLI

**Objetivo**: Mover funcoes de renderizacao para modulo reutilizavel

#### Tarefas:

1. **Criar diretorio e arquivo `internal/ui/cli/progress.go`**
   
   ```go
   // Package cli provides CLI-specific UI utilities for terminal output.
   package cli

   import (
       "encoding/json"
       "fmt"
       "os"

       "github.com/quantmind-br/shotgun-cli/internal/app"
   )

   // RenderProgress renders progress in the specified mode.
   func RenderProgress(mode app.ProgressMode, p app.ProgressOutput) {
       switch mode {
       case app.ProgressHuman:
           RenderProgressHuman(p)
       case app.ProgressJSON:
           RenderProgressJSON(p)
       case app.ProgressNone:
           // Silent
       }
   }

   // RenderProgressHuman renders progress for human consumption.
   func RenderProgressHuman(p app.ProgressOutput) {
       if p.Total > 0 {
           fmt.Fprintf(os.Stderr, "\r[%s] %s: %d/%d (%.1f%%)  ",
               p.Stage, p.Message, p.Current, p.Total, p.Percent)
       } else {
           fmt.Fprintf(os.Stderr, "\r[%s] %s  ", p.Stage, p.Message)
       }
   }

   // RenderProgressJSON renders progress as JSON (one line per event).
   func RenderProgressJSON(p app.ProgressOutput) {
       data, _ := json.Marshal(p)
       fmt.Fprintln(os.Stderr, string(data))
   }

   // ClearProgressLine clears the progress line (for human mode).
   func ClearProgressLine(mode app.ProgressMode) {
       if mode == app.ProgressHuman {
           fmt.Fprint(os.Stderr, "\r\033[K")
       }
   }
   ```

2. **Criar testes para progress**
   
   Arquivos envolvidos: `internal/ui/cli/progress_test.go`
   
   ```go
   package cli

   import (
       "bytes"
       "os"
       "testing"

       "github.com/stretchr/testify/assert"

       "github.com/quantmind-br/shotgun-cli/internal/app"
   )

   func TestRenderProgressHuman(t *testing.T) {
       t.Parallel()

       tests := []struct {
           name     string
           input    app.ProgressOutput
           contains string
       }{
           {
               name: "with total",
               input: app.ProgressOutput{
                   Stage:   "scanning",
                   Message: "Processing",
                   Current: 50,
                   Total:   100,
                   Percent: 50.0,
               },
               contains: "[scanning] Processing: 50/100 (50.0%)",
           },
           {
               name: "without total",
               input: app.ProgressOutput{
                   Stage:   "init",
                   Message: "Starting",
               },
               contains: "[init] Starting",
           },
       }

       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Capture stderr
               old := os.Stderr
               r, w, _ := os.Pipe()
               os.Stderr = w

               RenderProgressHuman(tt.input)

               w.Close()
               var buf bytes.Buffer
               buf.ReadFrom(r)
               os.Stderr = old

               assert.Contains(t, buf.String(), tt.contains)
           })
       }
   }

   func TestRenderProgressJSON(t *testing.T) {
       t.Parallel()

       // Capture stderr
       old := os.Stderr
       r, w, _ := os.Pipe()
       os.Stderr = w

       RenderProgressJSON(app.ProgressOutput{
           Timestamp: "2024-01-01T00:00:00Z",
           Stage:     "test",
           Message:   "Testing",
           Current:   1,
           Total:     10,
           Percent:   10.0,
       })

       w.Close()
       var buf bytes.Buffer
       buf.ReadFrom(r)
       os.Stderr = old

       output := buf.String()
       assert.Contains(t, output, `"stage":"test"`)
       assert.Contains(t, output, `"message":"Testing"`)
       assert.Contains(t, output, `"current":1`)
   }
   ```

### Fase 4: Simplificar cmd/context.go

**Objetivo**: Reduzir `cmd/context.go` para orquestracao apenas

#### Tarefas:

1. **Refatorar `cmd/context.go`**
   
   De (~522 linhas) para (~200 linhas):
   
   ```go
   package cmd

   import (
       "context"
       "fmt"
       "os"
       "path/filepath"
       "time"

       "github.com/rs/zerolog/log"
       "github.com/spf13/cobra"
       "github.com/spf13/viper"

       "github.com/quantmind-br/shotgun-cli/internal/app"
       cfgkeys "github.com/quantmind-br/shotgun-cli/internal/config"
       "github.com/quantmind-br/shotgun-cli/internal/core/tokens"
       "github.com/quantmind-br/shotgun-cli/internal/platform/geminiweb"
       "github.com/quantmind-br/shotgun-cli/internal/ui/cli"
       "github.com/quantmind-br/shotgun-cli/internal/utils"
   )

   var contextCmd = &cobra.Command{
       Use:   "context",
       Short: "Context generation tools",
       Long:  "Commands for generating and managing codebase context for LLMs",
   }

   var contextGenerateCmd = &cobra.Command{
       Use:   "generate",
       Short: "Generate context from codebase",
       Long: `Generate a structured text representation of your codebase...`,
       PreRunE: validateContextFlags,
       RunE:    runContextGenerate,
   }

   func validateContextFlags(cmd *cobra.Command, args []string) error {
       rootPath, _ := cmd.Flags().GetString("root")
       if rootPath == "" {
           return fmt.Errorf("root path cannot be empty")
       }

       absPath, err := filepath.Abs(rootPath)
       if err != nil {
           return fmt.Errorf("invalid root path '%s': %w", rootPath, err)
       }

       if _, err := os.Stat(absPath); os.IsNotExist(err) {
           return fmt.Errorf("root path does not exist: %s", absPath)
       }

       if info, err := os.Stat(absPath); err != nil {
           return fmt.Errorf("cannot access root path '%s': %w", absPath, err)
       } else if !info.IsDir() {
           return fmt.Errorf("root path must be a directory: %s", absPath)
       }

       maxSizeStr, _ := cmd.Flags().GetString("max-size")
       if _, err := utils.ParseSize(maxSizeStr); err != nil {
           return fmt.Errorf("invalid max-size format '%s': %w", maxSizeStr, err)
       }

       return nil
   }

   func runContextGenerate(cmd *cobra.Command, args []string) error {
       flags := extractFlags(cmd)
       
       config, err := app.NewGenerateConfigFromFlags(flags)
       if err != nil {
           return fmt.Errorf("failed to build configuration: %w", err)
       }

       log.Info().Str("root", config.RootPath).Msg("Starting context generation...")

       if err := generateContextHeadless(config); err != nil {
           return fmt.Errorf("context generation failed: %w", err)
       }

       log.Info().Msg("Context generated successfully")
       return nil
   }

   func extractFlags(cmd *cobra.Command) app.GenerateFlags {
       rootPath, _ := cmd.Flags().GetString("root")
       include, _ := cmd.Flags().GetStringSlice("include")
       exclude, _ := cmd.Flags().GetStringSlice("exclude")
       output, _ := cmd.Flags().GetString("output")
       maxSize, _ := cmd.Flags().GetString("max-size")
       enforceLimit, _ := cmd.Flags().GetBool("enforce-limit")
       sendGemini, _ := cmd.Flags().GetBool("send-gemini")
       geminiModel, _ := cmd.Flags().GetString("gemini-model")
       geminiOutput, _ := cmd.Flags().GetString("gemini-output")
       geminiTimeout, _ := cmd.Flags().GetInt("gemini-timeout")
       templateName, _ := cmd.Flags().GetString("template")
       task, _ := cmd.Flags().GetString("task")
       rules, _ := cmd.Flags().GetString("rules")
       customVars, _ := cmd.Flags().GetStringArray("var")
       workers, _ := cmd.Flags().GetInt("workers")
       includeHidden, _ := cmd.Flags().GetBool("include-hidden")
       includeIgnored, _ := cmd.Flags().GetBool("include-ignored")
       progressMode, _ := cmd.Flags().GetString("progress")

       return app.GenerateFlags{
           RootPath:       rootPath,
           Include:        include,
           Exclude:        exclude,
           Output:         output,
           MaxSize:        maxSize,
           EnforceLimit:   enforceLimit,
           SendGemini:     sendGemini,
           GeminiModel:    geminiModel,
           GeminiOutput:   geminiOutput,
           GeminiTimeout:  geminiTimeout,
           Template:       templateName,
           Task:           task,
           Rules:          rules,
           CustomVars:     customVars,
           Workers:        workers,
           IncludeHidden:  includeHidden,
           IncludeIgnored: includeIgnored,
           ProgressMode:   progressMode,
       }
   }

   func generateContextHeadless(cfg *app.CLIGenerateConfig) error {
       scannerConfig := app.BuildScannerConfig(cfg)
       templateVars := app.BuildTemplateVars(cfg)
       
       templateContent, err := app.LoadTemplateContent(cfg.Template)
       if err != nil {
           return err
       }

       svc := app.NewContextService()
       svcCfg := app.GenerateConfig{
           RootPath:        cfg.RootPath,
           ScanConfig:      &scannerConfig,
           Template:        templateContent,
           TemplateVars:    templateVars,
           MaxSize:         cfg.MaxSize,
           EnforceLimit:    cfg.EnforceLimit,
           OutputPath:      cfg.Output,
           CopyToClipboard: viper.GetBool(cfgkeys.KeyOutputClipboard),
           IncludeTree:     viper.GetBool(cfgkeys.KeyContextIncludeTree),
           IncludeSummary:  viper.GetBool(cfgkeys.KeyContextIncludeSummary),
           SkipBinary:      viper.GetBool(cfgkeys.KeyScannerSkipBinary),
       }

       var result *app.GenerateResult
       ctx := context.Background()

       if cfg.ProgressMode != app.ProgressNone {
           result, err = svc.GenerateWithProgress(ctx, svcCfg, func(stage, msg string, cur, total int64) {
               var percent float64
               if total > 0 {
                   percent = float64(cur) / float64(total) * 100
               }
               cli.RenderProgress(cfg.ProgressMode, app.ProgressOutput{
                   Timestamp: time.Now().Format(time.RFC3339),
                   Stage:     stage,
                   Message:   msg,
                   Current:   cur,
                   Total:     total,
                   Percent:   percent,
               })
           })
           cli.ClearProgressLine(cfg.ProgressMode)
       } else {
           result, err = svc.Generate(ctx, svcCfg)
       }

       if err != nil {
           return fmt.Errorf("context generation failed: %w", err)
       }

       log.Info().Int("files", result.FileCount).Msg("Files scanned")
       if result.CopiedToClipboard {
           log.Info().Msg("Context copied to clipboard")
       }

       printGenerationSummary(result, cfg)

       if cfg.SendGemini {
           if err := sendToGemini(cfg, result.Content); err != nil {
               log.Error().Err(err).Msg("Failed to send to Gemini")
               fmt.Printf("Error sending to Gemini: %v\n", err)
           }
       }

       return nil
   }

   func printGenerationSummary(result *app.GenerateResult, cfg *app.CLIGenerateConfig) {
       fmt.Printf("Context generated successfully!\n")
       fmt.Printf("Root path: %s\n", cfg.RootPath)
       fmt.Printf("Output file: %s\n", result.OutputPath)
       fmt.Printf("Files processed: %d\n", result.FileCount)
       fmt.Printf("Total size: %s (~%s tokens)\n",
           utils.FormatBytes(result.ContentSize),
           tokens.FormatTokens(int(result.TokenEstimate)))
       fmt.Printf("Size limit: %s\n", utils.FormatBytes(cfg.MaxSize))
   }

   // sendToGemini remains in cmd/ as it's CLI-specific orchestration
   func sendToGemini(cfg *app.CLIGenerateConfig, content string) error {
       // ... unchanged logic ...
   }

   func init() {
       // ... flag definitions unchanged ...
   }
   ```

### Fase 5: Atualizacao de Testes

**Objetivo**: Atualizar testes para refletir nova estrutura

#### Tarefas:

1. **Mover testes de `cmd/context_test.go` para `internal/app/cli_config_test.go`**
   
   Testes que testam logica de parsing devem migrar para o novo local.

2. **Manter testes de integracao em `cmd/context_test.go`**
   
   Testes que validam o comando completo devem permanecer.

## Estrategia de Testes

### Testes Unitarios
- [ ] `TestNewGenerateConfigFromFlags_*` - Factory com varios inputs
- [ ] `TestParseProgressMode_*` - Todos os modos validos e invalidos
- [ ] `TestParseCustomVars_*` - Formatos validos e invalidos
- [ ] `TestBuildScannerConfig_*` - Configuracao do scanner
- [ ] `TestBuildTemplateVars_*` - Variaveis de template
- [ ] `TestRenderProgress*` - Todas as funcoes de renderizacao

### Testes de Integracao
- [ ] Comando `context generate` funciona identicamente
- [ ] Flags sao corretamente parseados
- [ ] Progresso e exibido corretamente

### Casos de Teste Especificos

| ID | Cenario | Input | Output Esperado |
|----|---------|-------|-----------------|
| TC01 | MaxSize invalido | "invalid" | Error de parsing |
| TC02 | Progress mode invalido | "xml" | Error de parsing |
| TC03 | Custom var invalido | "NOVALUE" | Error "expected KEY=VALUE" |
| TC04 | Root path relativo | "." | Path absoluto |
| TC05 | Output vazio | "" | Nome auto-gerado com timestamp |

## Riscos e Mitigacoes

| Risco | Probabilidade | Impacto | Mitigacao |
|-------|---------------|---------|-----------|
| Regressao no CLI | Baixo | Alto | Testes de integracao existentes |
| Import cycle | Medio | Medio | Usar `app.ProgressMode` em vez de `cli.ProgressMode` |
| Duplicacao de tipos | Baixo | Baixo | Remover tipos duplicados de cmd/ |

## Checklist de Conclusao

- [ ] `internal/app/config.go` atualizado com `GenerateFlags` e `CLIGenerateConfig`
- [ ] `internal/app/cli_config.go` criado com factory e helpers
- [ ] `internal/app/cli_config_test.go` criado com testes completos
- [ ] `internal/ui/cli/progress.go` criado com funcoes de renderizacao
- [ ] `internal/ui/cli/progress_test.go` criado com testes
- [ ] `cmd/context.go` simplificado (~200 linhas)
- [ ] Tipos duplicados removidos de `cmd/context.go`
- [ ] Testes existentes atualizados e passando
- [ ] `go test -race ./...` passando
- [ ] `golangci-lint run` passando

## Notas Adicionais

### Ordem de Implementacao Recomendada

1. Consolidar tipos em `app/config.go` primeiro (evita conflitos)
2. Criar `app/cli_config.go` com factory
3. Criar `ui/cli/progress.go`
4. Atualizar `cmd/context.go` para usar novos modulos
5. Remover codigo duplicado
6. Atualizar testes

### Compatibilidade com Outros Comandos

O novo modulo `ui/cli` pode ser usado por outros comandos:
- `cmd/send.go` - ja tem logica de progresso similar
- Futuros comandos que precisem de output estruturado

### Metricas de Sucesso

- **Antes**: `cmd/context.go` com 522 linhas, misturando orquestracao e logica
- **Depois**: 
  - `cmd/context.go`: ~200 linhas (orquestracao)
  - `app/cli_config.go`: ~150 linhas (logica reutilizavel)
  - `ui/cli/progress.go`: ~50 linhas (renderizacao reutilizavel)
- **Beneficio**: Testabilidade sem Cobra, reutilizacao de componentes
