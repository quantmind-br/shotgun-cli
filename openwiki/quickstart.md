# Shotgun CLI — Quickstart

**Shotgun CLI** is a cross-platform Go CLI that transforms codebases into LLM-optimized context. It provides both an interactive TUI wizard (Bubble Tea) and headless CLI commands (Cobra) for scanning, assembling context with templates, and sending to AI providers (OpenAI, Anthropic, Gemini).

**Module**: `github.com/quantmind-br/shotgun-cli` · **Go 1.24** · **Clean Architecture**

---

## Quick setup

```bash
# Build for current platform
make build                    # → build/shotgun-cli

# Or install to ~/.local/bin
make install                  # build + copy

# Run without args → TUI wizard
./build/shotgun-cli

# Run headless subcommands
./build/shotgun-cli context generate --help
./build/shotgun-cli llm status
./build/shotgun-cli config set llm.provider gemini
```

See [Operations](operations.md) for full build, test, and config reference.

---

## Architecture at a glance

The project follows **Clean Architecture / Hexagonal Architecture** with four layers and strict one-directional imports:

```
cmd/                → CLI commands + composition root (may import everything)
internal/ui/        → TUI wizard, Bubble Tea MVU (imports app + core)
internal/app/       → ContextService, ProviderRegistry (imports core + platform)
internal/core/      → domain: scanner, contextgen, template, llm, ignore,
  │                    tokens, diff, selection (stdlib ONLY)
internal/platform/  → infra: openai, anthropic, geminiapi, llmbase, http,
                       clipboard (imports core interfaces only)
internal/config/    → config key constants, validation, metadata
```

**Core rule**: `internal/core/` must never import project-internal packages or external dependencies beyond the Go standard library.

See [Architecture](architecture.md) for details.

---

## Repository map

| Directory | Purpose |
|-----------|---------|
| `cmd/` | CLI commands (Cobra), composition root |
| `internal/app/` | `ContextService` main API, `ProviderRegistry` |
| `internal/ui/` | TUI wizard, screens, components, coordinators |
| `internal/core/scanner/` | Filesystem traversal, layered ignore rules |
| `internal/core/contextgen/` | Context assembly, tree rendering |
| `internal/core/template/` | Template loading, variable substitution |
| `internal/core/ignore/` | Layered ignore engine |
| `internal/core/llm/` | Provider interfaces, config, registry |
| `internal/core/tokens/` | Token estimation |
| `internal/core/selection/` | Per-project file-deselection persistence |
| `internal/core/diff/` | Intelligent diff splitting |
| `internal/platform/openai/` | OpenAI LLM implementation |
| `internal/platform/anthropic/` | Anthropic LLM implementation |
| `internal/platform/geminiapi/` | Gemini LLM implementation |
| `internal/platform/llmbase/` | Base client with shared HTTP logic |
| `internal/platform/http/` | Shared JSON HTTP client |
| `internal/platform/clipboard/` | Clipboard integration |
| `internal/config/` | Config key constants, validation, metadata |
| `internal/utils/` | Utility functions (size parsing, conversion) |
| `internal/assets/` | Embedded templates |
| `test/e2e/` | End-to-end CLI tests |
| `test/fixtures/` | Sample project fixtures |

---

## Key workflows

- **TUI Wizard**: 5-step interactive flow (file selection → template → task → rules → review/send). Launched when `shotgun-cli` runs with no arguments.
- **Headless CLI**: `shotgun-cli context generate` runs the generation pipeline synchronously. Supports `--progress json` for machine-readable output.
- **LLM send**: Generated context can be sent to OpenAI, Anthropic, or Gemini via `shotgun-cli send` or from the review step in the TUI.

See [Workflows](workflows.md) for detailed flow descriptions.

---

## What's new

The most recent changes introduced:

- **File selection persistence** — per-project deselected files are saved to `~/.config/shotgun-cli/selections.json` and restored on subsequent scans. The TUI wizard and headless CLI both use `selection.Store` to load and save deselections.
- **Toggle ignored scan** — pressing `i` in the file-selection screen now rescans the directory with the `IncludeIgnored` flag toggled, instead of just toggling display.
- **SetShowIgnored** — the file tree component no longer toggles show-ignored state; it receives it from the wizard, which keeps scan-config and display in sync.
- **Generated artifact cleanup** — project caches, analysis outputs, and temp files removed from the repository.

See [Domains](domains.md) (selection store) and [Workflows](workflows.md) (selection persistence flow) for details.

---

## Domain concepts

Read about each core domain in [Domains](domains.md):

- Scanner — filesystem traversal, FileNode tree, layered ignore rules
- Ignore engine — layered rules from .gitignore, .shotgunignore, custom rules
- Context generation — tree rendering, file assembly, template substitution
- Template management — multi-source loading, variable interpolation
- Token estimation — rough bytes-to-tokens conversion
- LLM provider interfaces — Provider interface, Registry, config
- Selection store — JSON-backed per-project deselection persistence
- Diff splitting — intelligent chunking for LLM consumption

---

## Build, test, config

See [Operations](operations.md) for:

- Build commands (`make build`, `make build-all`, `make install`)
- Test commands (`make test`, `make test-race`, `make test-e2e`, `make coverage`)
- Linting (`make lint`)
- Configuration system (Viper, config file, env vars)
- CI (GitHub Actions)
- Release (GoReleaser)
