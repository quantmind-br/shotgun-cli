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

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **shotgun-cli** (5367 symbols, 15210 relationships, 259 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/shotgun-cli/context` | Codebase overview, check index freshness |
| `gitnexus://repo/shotgun-cli/clusters` | All functional areas |
| `gitnexus://repo/shotgun-cli/processes` | All execution flows |
| `gitnexus://repo/shotgun-cli/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
