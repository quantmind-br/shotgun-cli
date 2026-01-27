# CMD Package - CLI Commands

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Cobra command definitions. **Composition root** - wires dependencies, delegates to `app.ContextService`.

## COMMAND HIERARCHY

```
shotgun-cli (root)
├── [no args]        → TUI Wizard (internal/ui/wizard.go)
├── context
│   └── generate     → Headless context generation
├── config
│   ├── [no args]    → Config TUI
│   ├── show         → Display config
│   └── set          → Update config
├── llm
│   ├── status       → Provider status
│   ├── doctor       → Diagnostics
│   └── list         → List providers
├── template
│   ├── list/render/import/export
├── diff
│   └── split        → Split large diffs
├── send             → Send to LLM
└── completion       → Shell completion
```

## ADDING A NEW COMMAND

1. Create `cmd/<name>.go`
2. Define command: `var <name>Cmd = &cobra.Command{...}`
3. Add in `init()`: `rootCmd.AddCommand(<name>Cmd)` or parent
4. Create `cmd/<name>_test.go` with table-driven tests
5. Document in parent AGENTS.md

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
CLI commands delegate to `app.ContextService`:
```go
svc := app.NewContextService()
result, err := svc.Generate(ctx, cfg)
```

## KEY FILES

| File | Purpose |
|------|---------|
| `root.go` | Main entry, TUI launch, config defaults |
| `providers.go` | LLM provider registry init |
| `config_llm.go` | LLM config helpers |
| `context.go` | Context generation command |
| `send.go` | Send to LLM command |

## TESTING

```bash
go test -v ./cmd/...
go test -v -run TestSendCommand ./cmd/
```

Tests use stdout/stderr capture. See `cmd/*_test.go` for patterns.
