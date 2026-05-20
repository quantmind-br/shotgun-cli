# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`shotgun-cli` (module `github.com/quantmind-br/shotgun-cli`, Go 1.24) is a CLI that scans a
codebase, applies layered ignore rules, and generates LLM-optimized context files — then
optionally sends them to OpenAI/Anthropic/Gemini. It has two front ends sharing one core:

- **TUI wizard** (Bubble Tea, 5 steps): launched when `shotgun-cli` runs with no args.
- **Headless CLI**: any subcommand runs without the TUI. Entry: `main.go` → `cmd.Execute()`.

## Commands

```bash
make build                 # → build/shotgun-cli (current platform)
make build-all             # cross-compile linux/darwin/windows × amd64/arm64
make install               # build + copy to ~/.local/bin

go test -race ./...        # full test suite (preferred — race detector on)
go test -v -run TestName_Scenario ./internal/core/scanner/...   # single test / package
make test-e2e              # end-to-end CLI tests in ./test/e2e

golangci-lint run ./...    # or `make lint` (auto-picks .golangci-local.yml if present)
make coverage              # coverage.out + per-func report (target: 85%+ overall, 90%+ new code)
```

Linting is strict: 120-char lines, gocyclo max 25, plus errcheck/staticcheck/gosec/goconst/
lll/prealloc/unparam/unused. Run lint before considering work done.

## Architecture

Clean Architecture with strict one-directional imports — violating these is the most common
way to break the build:

```
cmd/                → CLI commands + composition root (may import everything)
internal/ui/        → TUI wizard, Bubble Tea MVU (imports app + core)
internal/app/       → ContextService, ProviderRegistry (imports core + platform)
internal/core/      → domain: scanner, contextgen, template, llm, ignore, tokens, diff (stdlib ONLY)
internal/platform/  → infra: openai, anthropic, geminiapi, llmbase, http, clipboard (imports core interfaces only)
internal/config/    → config keys, validation, metadata
```

Rule of thumb: **`core` never imports anything project-internal except stdlib**, and **`viper`
must never appear in `core/` or `platform/`** — pass config in via dependency injection.

### The central seam: `ContextService`

`internal/app/context.go` is the single API both front ends call. It coordinates scanner →
generator → (optional) LLM send, and is built with functional options (`WithScanner`,
`WithGenerator`, `WithRegistry`) — use those for test doubles rather than touching globals.

```go
svc := app.NewContextService()
result, err := svc.Generate(ctx, cfg)                            // CLI: synchronous
result, err := svc.GenerateWithProgress(ctx, cfg, callback)      // TUI: async w/ progress
result, err := svc.SendToLLMWithProgress(ctx, content, cfg, cb)  // send to a provider
```

Progress callbacks are not optional in the TUI path — dropping them stalls the wizard UI.

LLM providers are created only through `app.DefaultProviderRegistry.Create(...)`; never
construct a provider directly. Providers issue HTTP via `platform/http`'s JSON client, not raw
`net/http`.

## Where to make common changes

| Task | Location |
|------|----------|
| New CLI subcommand | `cmd/<name>.go`, register in its `init()` |
| New TUI screen | `internal/ui/screens/<name>.go`, wire into `internal/ui/wizard.go` |
| New LLM provider | `internal/platform/<name>/` + register in `internal/app/providers.go` |
| New config key | `internal/config/keys.go` + `metadata.go` + `validator.go` (never raw key strings) |
| Scanning / ignore / context-gen / templates | `internal/core/{scanner,ignore,contextgen,template}/` |

## Conventions

- Logging is `zerolog` only — never `fmt.Println` or stdlib `log`.
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`.
- Tests are table-driven with `t.Parallel()` and `stretchr/testify` (`require` for fatal,
  `assert` for soft checks); name them `TestFunc_Scenario`.
- Imports in three blank-line-separated groups: stdlib → external → internal.

## Per-directory docs

Each major package has a detailed `AGENTS.md` (root `AGENTS.md`, `cmd/`, `internal/` and its
subpackages). Consult the relevant one before larger changes — they go deeper than this file.

## Issue tracking

Work items live in **bd** (beads), not markdown TODOs: `bd ready`, `bd show <id>`,
`bd update <id> --status in_progress`, `bd close <id>`, `bd sync`.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **shotgun-cli** (5115 symbols, 14961 relationships, 259 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

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