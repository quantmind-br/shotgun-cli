# PROJECT KNOWLEDGE BASE

**Generated:** 2026-02-12T02:36:00Z  
**Commit:** 0917a6c  
**Branch:** main

---

## Overview

**shotgun-cli** — Go CLI that generates LLM-optimized codebase contexts. Interactive TUI wizard (Bubble Tea) + headless CLI. Sends context to OpenAI/Anthropic/Gemini.

**Module**: `github.com/quantmind-br/shotgun-cli` | **Go 1.24** | **Clean Architecture**

## Issue Tracking

Uses **bd** (beads). Do NOT use markdown TODOs.

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress
bd close <id>         # Complete work
bd sync               # Sync with git
```

**Types**: `bug`, `feature`, `task`, `epic`, `chore` | **Priorities**: `0`=Critical → `4`=Backlog

## Build/Test/Lint

```bash
# Build
make build                    # → build/shotgun-cli

# Test
go test -race ./...           # All tests with race detector (preferred)
go test -v -run TestFoo ./internal/core/scanner/...  # Single test

# Lint
golangci-lint run ./...       # Uses .golangci-local.yml
make lint

# Coverage (85% minimum, 90%+ target for new code)
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total

# E2E
make test-e2e
```

**Linting**: 120 char lines | 25 cyclomatic complexity max | govet, errcheck, staticcheck, goconst, misspell, gocyclo, gosec, lll, prealloc, unconvert, unparam, unused

## Architecture

```
cmd/                    → Presentation (CLI commands, composition root)
internal/ui/            → Presentation (TUI wizard, Bubble Tea MVU)
internal/app/           → Application (ContextService, ProviderRegistry)
internal/core/          → Domain (scanner, contextgen, template, llm, ignore, tokens, diff)
internal/platform/      → Infrastructure (openai, anthropic, geminiapi, llmbase, http, clipboard)
internal/config/        → Config keys, validation, metadata
internal/assets/        → Embedded templates
```

**Import rules**: `core` → stdlib only. `platform` → core interfaces only. `app` → core+platform. `ui` → app+core. `cmd` → everything (composition root).

## Code Style

- **Imports**: stdlib → external → internal (3 groups, blank-line separated)
- **Errors**: Always `fmt.Errorf("context: %w", err)`
- **Logging**: `zerolog` only. Never `fmt.Println` or stdlib `log`
- **Config keys**: Constants from `internal/config/keys.go`, never raw strings
- **Tests**: Table-driven, `t.Parallel()`, `stretchr/testify` (require for fatal, assert for checks)
- **Naming**: `TestFunctionName_Scenario`, lowercase single-word packages, `-er` suffix interfaces

## Key Services

```go
// ContextService — main API (internal/app/context.go)
svc := app.NewContextService()
result, err := svc.Generate(ctx, cfg)                              // CLI (sync)
result, err := svc.GenerateWithProgress(ctx, cfg, callback)        // TUI (async)
result, err := svc.SendToLLMWithProgress(ctx, content, cfg, cb)   // Send to LLM

// ProviderRegistry — LLM provider factory (internal/app/providers.go)
provider, err := app.DefaultProviderRegistry.Create(llm.ProviderOpenAI, cfg)
```

## WHERE TO LOOK

| Task | Location |
|------|----------|
| Add CLI command | `cmd/<name>.go` + register in `init()` |
| Add TUI screen | `internal/ui/screens/<name>.go` + wire in `wizard.go` |
| Add LLM provider | `internal/platform/<name>/` + `internal/app/providers.go` |
| Add config key | `internal/config/keys.go` + `metadata.go` + `validator.go` |
| Modify scanning | `internal/core/scanner/` |
| Modify context gen | `internal/core/contextgen/` |
| Modify templates | `internal/core/template/` |
| Modify ignore rules | `internal/core/ignore/` |

## Anti-Patterns

- Importing `viper` in core/ or platform/ (use DI)
- Global state anywhere in internal/
- Skipping progress callbacks (breaks TUI)
- Direct HTTP in providers (use `platform/http/JSONClient`)
- `fmt.Println` or stdlib `log` (use `zerolog`)
- Creating providers outside registry

## Session Completion

**Work is NOT complete until `git push` succeeds.**

1. Run quality gates: `go test -race ./... && golangci-lint run`
2. File issues for remaining work: `bd create "..." -p 2`
3. Update issue status: `bd close <id>`
4. Push: `git pull --rebase && bd sync && git push`

## Hierarchy

```
AGENTS.md (this file)
├── cmd/AGENTS.md              → CLI commands
└── internal/AGENTS.md         → Internal architecture
    ├── ui/AGENTS.md           → TUI wizard
    ├── app/AGENTS.md          → Application services
    ├── core/AGENTS.md         → Domain logic
    └── platform/AGENTS.md     → Infrastructure
```
