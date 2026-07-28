---
name: shotgun-ignore
description: >
  Use when creating, updating, or regenerating a project `.shotgunignore` that EXCLUDES
  key/core source files from shotgun-cli scans so they are not mapped or sent to an LLM,
  protecting proprietary code, IP-sensitive modules, or secrets-adjacent logic. Also use for
  "shotgunignore", "shotgun ignore", "criar .shotgunignore", "não enviar código chave pro LLM",
  "bloquear arquivos chave no shotgun", "hide core code from shotgun", or when the user wants
  shotgun-cli to omit primary application source from LLM context.
---

# shotgun-ignore

Scan a repo, identify **key/core code files**, write root `.shotgunignore` so shotgun-cli
**does not map or send those files** to the LLM.

## Goal (literal)

| Class | Action in `.shotgunignore` | Effect on LLM context |
|-------|----------------------------|------------------------|
| **Key / core source** | **IGNORE (list them)** | Excluded — not sent |
| **Peripheral / safe** | leave unlisted | Still scannable / sendable |

This is an **exclusion** skill for proprietary or sensitive core code. It is **not** a
"trim tests to save tokens" skill. Default behavior matches the user ask: key files out.

## When to use

- User wants core/business logic kept out of shotgun → LLM pipelines
- IP/privacy: only structure, docs, tests, or non-core packages may go to the model
- Regenerating or tightening an existing key-exclusion `.shotgunignore`

## When NOT to use

- User wants maximum useful code **in** the LLM context (noise-trim only) — different goal; do not use this skill's default
- User only wants git ignore → edit `.gitignore`
- One-off file pick for a single prompt → TUI selection / `--exclude`, not a durable ignore file

## Built-in (do NOT re-list)

shotgun-cli already ignores noise like `.git/`, `node_modules/`, `vendor/`, build dirs, caches,
media, binaries, logs, `shotgun-prompt*.md`. Do not pad `.shotgunignore` with those.

This skill adds **project key-source paths** on top of built-ins.

## Workflow

### 1. Scope root

- Default = workspace / git root.
- User path overrides.
- One `.shotgunignore` at project root (unless user scopes a monorepo package).

### 2. Inventory (read-only)

1. List root; detect stack (`go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, etc.).
2. Find source roots: `cmd/`, `internal/`, `src/`, `app/`, `lib/`, `pkg/`, `packages/`, …
3. Read manifests, README, entrypoints — learn what the product *is*.
4. If `.shotgunignore` exists, read it and merge (never blind-overwrite).
5. Prefer `glob` / targeted reads over dumping the whole tree.

### 3. Identify KEY files (must exclude)

Key = code that defines the product's proprietary behavior or crown jewels.

**Usually KEY (exclude):**

- Domain / business logic modules
- Core services, engines, algorithms, pricing, authz policy internals
- Primary CLI/API command implementations (not just `main` stubs if they hold logic)
- Internal packages that encode product rules (`internal/`, private packages)
- Proprietary ML/model code, closed workflows, licensed adapters
- Config that embeds secrets or non-public endpoints (paths only — never paste secret values into the ignore file)

**Usually NOT key (leave sendable unless user says otherwise):**

- Tests, fixtures, mocks, testdata
- Public docs, README, examples meant for sharing
- Generated glue with no unique logic
- Thin `main` that only wires public packages
- Open scaffolding, CI configs, linters

**Heuristics by signal:**

| Signal | Lean |
|--------|------|
| Under `internal/`, `core/`, `domain/`, `engine/`, `service/` with real logic | KEY |
| Package is the published product surface but implementation is closed | KEY |
| File is pure test / fixture / snapshot | not key |
| File is vendored third-party already built-in-ignored | skip (already out) |
| Ambiguous utility used only by core | KEY if removing it hides the secret sauce |

When unsure: **exclude** (safer for IP). Note ambiguity in the handoff.

### 4. Express patterns

Prefer the **smallest set of globs/paths** that covers key code without nuking the whole repo by accident.

Order of preference:

1. **Directory** of a core package: `internal/billing/`, `src/engine/`
2. **Glob** for a family: `internal/core/**/*.go` (only if entire tree is key)
3. **Explicit file** for outliers: `cmd/secret-tool/main.go`

Avoid:

- Ignoring the entire repo (`*` / `**`) unless user explicitly demands full lock-down
- Thousands of per-file lines when one directory pattern works
- Patterns that only match noise (tests) while leaving core in — that inverts the goal

gitignore syntax:

- `#` comments
- `dir/` for directories
- `**` / `*` globs
- `!path` negation to re-include a safe file under an ignored parent (use sparingly)

### 5. Write `.shotgunignore`

Path: `<project-root>/.shotgunignore`

Body structure:

```gitignore
# shotgun-cli: EXCLUDE key/core source from LLM context
# skill: shotgun-ignore | generated: YYYY-MM-DD
# intent: key files listed below are NOT mapped/sent

# --- Core / domain ---
internal/core/
internal/billing/

# --- Proprietary services ---
src/engine/

# --- user-custom (preserved) ---
```

Merge rules when file exists:

1. Parse old non-comment patterns.
2. Compute new key-exclusion set from scan.
3. Union with old patterns that still look intentional.
4. Drop patterns that only duplicate built-ins.
5. Rewrite with clean sections; put unrecognized old lines under `# user-custom`.

Do not ignore `.shotgunignore` itself.

### 6. Verify

1. File exists at root.
2. Spot-check: at least one known **key** path matches a pattern; at least one **peripheral** path (e.g. test or README) does **not**.
3. Optional if `shotgun-cli` present — generate context and confirm key paths absent, periphery present:

```bash
shotgun-cli context generate . -o /tmp/shotgun-ignore-check.md
# or with a template that only needs structure
```

4. Handoff to user (required fields below).

## Output contract (always report)

- Absolute path of `.shotgunignore`
- Stack detected
- Created vs merged
- **Key paths excluded** (grouped)
- **Left sendable** (what periphery remains for the LLM)
- Ambiguous items and default taken (exclude)
- Reminder: this mode **hides core code** from the LLM by design

## Anti-patterns

| Don't | Why |
|-------|-----|
| Keep key source sendable and only ignore tests | Opposite of this skill's goal |
| Re-list built-in `node_modules/`, `*.png`, `.git/` | Redundant |
| Ignore everything with `*` without user OK | May be intended lockdown — ask first if ambiguous |
| Blind overwrite existing `.shotgunignore` | Loses custom rules |
| Paste secret values into the ignore file | Patterns only; secrets stay out of git via real secret hygiene |
| "Clarify away" a user who asked to exclude key files | Literal request is the default |

## Minimal example (Go CLI with proprietary core)

```gitignore
# shotgun-cli: EXCLUDE key/core source from LLM context
# skill: shotgun-ignore

# Domain + engines (proprietary)
internal/core/
internal/engine/
internal/billing/

# Closed command implementations
cmd/shotgun-cli/app.go
cmd/shotgun-cli/service/

# Leave sendable: tests, docs, thin main, public types (unlisted)
```

## Red flags — re-read goal

- About to write a noise-only ignore (tests/fixtures) while core stays exposed → wrong skill mode
- About to exclude only one file when whole `internal/core/` is key → widen pattern
- Monorepo: only scanned one package → prefix patterns or confirm package list
- User later says "I meant drop noise, keep core in context" → invert policy only after explicit flip; until then, exclude key source
