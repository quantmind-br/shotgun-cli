# Internal Package - Architecture Guide

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Clean Architecture implementation. **Core is pure domain logic; Platform is infrastructure.**

## LAYER BOUNDARIES

```
app/       → Application (orchestration, ContextService, ProviderRegistry)
core/      → Domain (scanner, context, template, llm, tokens, diff, ignore)
platform/  → Infrastructure (openai, anthropic, geminiapi, llmbase, http, clipboard)
ui/        → Presentation (TUI wizard, screens, components, coordinators)
config/    → Configuration (keys, validator, metadata)
assets/    → Embedded (templates)
```

## CRITICAL IMPORT RULES

| Package | CAN import from | CANNOT import from |
|---------|-----------------|-------------------|
| `core/*` | stdlib only | app, platform, ui, cmd |
| `app/` | core, platform | ui, cmd |
| `platform/*` | core (interfaces) | app, ui, cmd |
| `ui/` | app, core | cmd |
| `config/` | stdlib only | anything else |

**Violation = architecture break. LSP/linter will catch.**

## SHARED PATTERNS

### Progress Pattern
All async operations use callbacks/channels:
```go
// Channel-based (scanner)
func ScanWithProgress(path string, cfg *ScanConfig, progress chan<- Progress) (*FileNode, error)

// Callback-based (context, llm)
func GenerateWithProgress(cfg GenerateConfig, progress func(GenProgress)) (*ContextData, error)
func SendWithProgress(ctx context.Context, content string, progress func(stage string)) (*Result, error)
```

### Config Pattern
Every operation has a Config struct with DefaultConfig():
- `ScanConfig`, `GenerateConfig`, `SplitConfig`, `ManagerConfig`
- Validate at operation start, not construction

### Interface-First
Core defines interfaces; platform/app implement:
- `Scanner` → `FilesystemScanner`
- `Provider` → `openai.Client`, `anthropic.Client`
- `ContextGenerator` → `DefaultContextGenerator`

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Add LLM provider | platform/\<name\>/ + app/providers.go |
| Add config key | config/keys.go + config/metadata.go + config/validator.go |
| Add CLI command | ../cmd/\<name\>.go |
| Add TUI screen | ui/screens/\<name\>.go + ui/wizard.go |
| Modify scanning | core/scanner/ |
| Modify context output | core/context/ |
| Modify ignore logic | core/ignore/ |

## ANTI-PATTERNS

- Importing `viper` in core/ or platform/ (use DI)
- Global state anywhere in internal/
- Skipping progress callbacks (breaks TUI)
- Direct HTTP in providers (use platform/http/JSONClient)
