# Architecture

## Clean Architecture layers

Shotgun CLI follows **Clean Architecture** (also called hexagonal architecture) with strict one-directional import rules. Violating the layer boundaries is the most common way to break the build.

```
cmd/                → Presentation (CLI commands, composition root)
internal/ui/        → Presentation (TUI wizard, Bubble Tea MVU)
internal/app/       → Application (ContextService, ProviderRegistry)
internal/core/      → Domain (scanner, contextgen, template, llm,
                      ignore, tokens, diff, selection)
internal/platform/  → Infrastructure (openai, anthropic, geminiapi,
                      llmbase, http, clipboard)
internal/config/    → Configuration keys, validation, metadata
internal/utils/     → Utility functions
internal/assets/    → Embedded template files
```

## Dependency rules (strictly enforced)

| Package | May import |
|---------|-----------|
| `core/*` | Go stdlib only. No project-internal packages, no external deps. |
| `platform/*` | `core` interfaces only, plus stdlib and external HTTP client libs. |
| `app/` | `core` + `platform` + `config`. |
| `ui/` | `app` + `core` + `config` + Bubble Tea / Lipgloss. |
| `cmd/` | Everything (composition root). |

**Key restriction**: `viper` must never appear in `core/` or `platform/` — pass configuration via dependency injection.

Data flow direction: `cmd/` → everything, `ui/` → `app/` → `core/` + `platform/`.

## The central seam: `ContextService`

Located at `internal/app/context.go`, `ContextService` is the single API both front ends (TUI wizard and headless CLI) call. It coordinates scanner → generator → (optional) LLM send.

```go
svc := app.NewContextService()
result, err := svc.Generate(ctx, cfg)                           // CLI: synchronous
result, err := svc.GenerateWithProgress(ctx, cfg, callback)     // TUI: async w/ progress
result, err := svc.SendToLLMWithProgress(ctx, content, cfg, cb) // send to a provider
```

The service is built with functional options (`WithScanner`, `WithGenerator`, `WithRegistry`, `WithSelectionStore`) — use these for test doubles rather than touching globals.

Source: `internal/app/service.go`, `internal/app/context.go`

## Design Patterns

| Pattern | Usage | Location |
|---------|-------|----------|
| Command Pattern | CLI command structure | `cmd/*.go` |
| Builder/Generator | Context assembly pipeline | `internal/core/contextgen/` |
| Strategy Pattern | LLM provider abstraction | `internal/core/llm/` (Provider interface), `internal/platform/*/` (implementations) |
| MVU Pattern | TUI state management (Model-View-Update) | `internal/ui/*` (Bubble Tea) |
| Template Method | Standardized template rendering | `internal/core/template/` |
| Factory Pattern | Scanner and template manager creation | `internal/core/scanner/`, `internal/core/template/` |
| Functional Options | Service configuration | `internal/app/service.go` |

## LLM Provider Architecture

HTTP-based LLM providers use a shared `BaseClient` with the **Strategy pattern**:

```
                     BaseClient
                         │
           ┌─────────────┼─────────────┐
           ▼             ▼             ▼
        OpenAI        Anthropic     GeminiAPI
    (implements     (implements   (implements
     Sender)         Sender)       Sender)
```

- **BaseClient** (`internal/platform/llmbase/`): Provides common send/send-with-progress logic, embeds `JSONClient` from `internal/platform/http/`.
- **Sender interface**: Each provider implements `BuildRequest`, `ParseResponse`, `GetEndpoint`, `GetHeaders`, `GetProviderName`.
- **JSONClient** (`internal/platform/http/client.go`): Shared HTTP client with `PostJSON()` method and standardized `HTTPError` type.

Providers are created only through `app.DefaultProviderRegistry.Create(...)`; never construct a provider directly.

Source: `internal/platform/llmbase/base_client.go`, `internal/platform/llmbase/sender.go`, `internal/platform/http/client.go`, `internal/app/providers.go`, `internal/core/llm/provider.go`, `internal/core/llm/registry.go`

## Source map

| File | Role |
|------|------|
| `main.go` | Entry point, sets up zerolog, calls `cmd.Execute()` |
| `cmd/root.go` | Root Cobra command, TUI wizard launcher, config init |
| `cmd/context.go` | `context generate` headless command |
| `cmd/llm.go` | `llm status`, `llm doctor`, `llm list` commands |
| `cmd/send.go` | `send` command for LLM submission |
| `cmd/template.go` | `template list` command |
| `cmd/config.go` | `config set/get` commands |
| `cmd/diff.go` | `diff split` command |
| `cmd/completion.go` | Shell completion command |
| `cmd/providers.go` | Provider registration and metadata |
| `internal/app/service.go` | `DefaultContextService` implementation |
| `internal/app/context.go` | `GenerateConfig`, `GenerateResult` types, `ContextService` interface |
| `internal/app/providers.go` | `DefaultProviderRegistry` setup |
| `internal/app/config.go` | Config-to-LLM-config mapping |
| `internal/ui/wizard.go` | TUI `WizardModel`, 5-step state machine |
| `internal/ui/generate_coordinator.go` | Async context generation coordinator |
| `internal/ui/scan_coordinator.go` | Async filesystem scanning coordinator |
| `internal/ui/config_wizard.go` | Interactive config editor |
| `internal/ui/screens/` | Individual TUI screens (file_selection, template_selection, review, etc.) |
| `internal/ui/components/` | Reusable TUI widgets (tree, progress, config fields) |
| `internal/ui/styles/` | Lipgloss theme, colors, and styling |
| `internal/core/scanner/` | Filesystem scanner, `FileNode` tree type |
| `internal/core/ignore/` | Layered ignore rule engine |
| `internal/core/contextgen/` | Context content generation |
| `internal/core/template/` | Template loading, rendering, manager |
| `internal/core/tokens/` | Token estimation |
| `internal/core/llm/` | Provider interface, config, registry |
| `internal/core/selection/` | Per-project deselection persistence |
| `internal/core/diff/` | Diff splitting for LLM |
| `internal/platform/*/` | Provider implementations and HTTP client |
| `internal/config/` | Config key constants and validators |
| `internal/utils/` | Size parsing, conversion helpers |
| `internal/assets/` | Embedded prompt templates |

## Change guidance

| Task | Location |
|------|----------|
| New CLI subcommand | `cmd/<name>.go`, register in its `init()` |
| New TUI screen | `internal/ui/screens/<name>.go`, wire into `internal/ui/wizard.go` |
| New LLM provider | `internal/platform/<name>/` + register in `internal/app/providers.go` |
| New config key | `internal/config/keys.go` + `metadata.go` + `validator.go` |
| Scanning / ignore / context-gen / templates | `internal/core/{scanner,ignore,contextgen,template}/` |
| File selection persistence | `internal/core/selection/` + wire in `internal/app/service.go` and `cmd/root.go` |
