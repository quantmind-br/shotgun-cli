# CMD Package - CLI Commands

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Cobra command definitions. **Composition root** — wires dependencies, delegates to `app.ContextService`.

## COMMAND TREE

```
shotgun-cli (root)
├── [no args]        → TUI Wizard (internal/ui/wizard.go)
├── context
│   └── generate     → Headless context generation
├── config
│   ├── [no args]    → Config TUI (interactive editor)
│   ├── show         → Display config with sources
│   └── set          → Update config (validates before saving)
├── llm
│   ├── status       → Provider status
│   ├── doctor       → Diagnostics with fix guidance
│   └── list         → List providers (* marks current)
├── template
│   ├── list/render/import/export
├── diff
│   └── split        → Split large diffs at file boundaries
├── send             → Send context to LLM provider
└── completion       → Shell completion (bash/zsh/fish/powershell)
```

## KEY FILES

| File | Purpose |
|------|---------|
| `root.go` | Main entry, TUI launch, Viper config init, logging setup |
| `providers.go` | LLM provider registry init (OpenAI, Anthropic, Gemini) |
| `config_llm.go` | `BuildLLMConfig()`, `BuildLLMConfigWithOverrides()`, `CreateLLMProvider()` |
| `context.go` | Context generate command, progress rendering (human/JSON/none) |
| `send.go` | Send to LLM command, `formatDuration()` helper |
| `config.go` | Config show/set + interactive Config TUI launcher |
| `llm.go` | LLM status/doctor/list, `displayURL()` helper |
| `template.go` | Template list/render/import/export |
| `diff.go` | Diff split command |
| `completion.go` | Shell completion generation |

## ADDING A NEW COMMAND

1. Create `cmd/<name>.go`
2. Define: `var <name>Cmd = &cobra.Command{...}`
3. Register in `init()`: `rootCmd.AddCommand(<name>Cmd)` or parent
4. Create `cmd/<name>_test.go` with table-driven tests

## PATTERNS

### Flag Binding
```go
func init() {
    myCmd.Flags().StringP("output", "o", "", "Output file")
    viper.BindPFlag("my.output", myCmd.Flags().Lookup("output"))
}
```

### Config Access
Always use constants from `internal/config/keys.go`:
```go
import "github.com/quantmind-br/shotgun-cli/internal/config"
viper.GetString(config.KeyLLMProvider)
```

### Service Delegation
```go
svc := app.NewContextService()
result, err := svc.Generate(ctx, cfg)
```

### Config Initialization Flow
`main.go` → `cmd.Execute()` → `rootCmd` → `cobra.OnInitialize(initConfig)`:
1. Sets up zerolog logging
2. Loads config from `~/.config/shotgun-cli/config.yaml`
3. Binds env vars (prefix: `SHOTGUN_`)
4. Sets defaults via `setConfigDefaults()`

## TESTING

```bash
go test -v ./cmd/...
go test -v -run TestSendCommand ./cmd/
```

Tests use stdout/stderr capture. See `cmd/*_test.go` for patterns.