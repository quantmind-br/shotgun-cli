# Repository Guidelines

## Project Overview

**shotgun-cli** is a Go CLI that turns a codebase into an LLM-optimized context document. It scans the filesystem with layered ignore rules, renders a file tree plus selected file contents into a prompt template, estimates tokens, and optionally sends the result to an LLM provider (OpenAI, Anthropic, Gemini).

Two front-ends share one application service: an interactive 5-step Bubble Tea wizard (launched with no arguments) and headless Cobra subcommands.

**Module**: `github.com/quantmind-br/shotgun-cli` | **Go 1.24.0** | Clean Architecture

## Architecture & Data Flow

### Layer Structure

```
cmd/               → Presentation: Cobra commands, viper init, composition root
internal/ui/       → Presentation: Bubble Tea TUI (wizard + config wizard), MVU
internal/app/      → Application: ContextService (generate/send orchestration), provider registry
internal/core/     → Domain: scanner, ignore, contextgen, template, tokens, selection, diff, llm interfaces
internal/platform/ → Infrastructure: openai, anthropic, geminiapi, llmbase, http, clipboard
internal/config/   → Config key constants, metadata/defaults, validation
internal/assets/   → go:embed of built-in prompt templates
```

### Dependency Rules (as actually enforced)

- `core/` → stdlib + a small set of pure libs (`go-gitignore`, `golang.org/x/text`, `adrg/xdg`) and other `core/` packages. **No viper, no app/ui/platform imports.**
- `platform/` → core interfaces + stdlib. **No viper.**
- `app/` → core + platform + config.
- `ui/` → app + core + config + Bubble Tea/Lipgloss. It *does* read viper directly in [config_wizard.go](file:///home/diogo/dev/shotgun-cli/internal/ui/config_wizard.go), [screens/config_category.go](file:///home/diogo/dev/shotgun-cli/internal/ui/screens/config_category.go), and [screens/template_selection.go](file:///home/diogo/dev/shotgun-cli/internal/ui/screens/template_selection.go) — a known deviation; do not extend it.
- `cmd/` → everything (composition root).

### Call Chains

**TUI wizard** (only when `len(os.Args) == 1`):
```
main.go → cmd.Execute() → rootCmd.Execute() → initConfig() → runRootCommand()
  → launchTUIWizard() → ui.NewWizard() → tea.NewProgram(...).Run()
  → Step 1 file selection (ScanCoordinator, async)
  → Step 2 template → Step 3 task → Step 4 rules
  → Step 5 review (GenerateCoordinator, async) → save / clipboard / SendToLLMWithProgress
```

**Headless generate**:
```
cmd/context.go: buildGenerateConfig() → generateContextHeadless()
  → app.NewContextService(WithSelectionStore(...)) → svc.Generate() / svc.GenerateWithProgress()
```

**Headless send** (separate from generation):
```
cmd/send.go: BuildLLMConfigWithOverrides() → CreateLLMProvider() → provider.Send()
```

**Config TUI**: `shotgun-cli config` (no subcommand) → `launchConfigTUI()` → `ui.NewConfigWizard()`.

### Generation Pipeline

1. `scanner.FileSystemScanner.ScanWithProgress()` — single-pass walk.
2. `ignore.LayeredIgnoreEngine.ShouldIgnore()` — precedence: explicit excludes → explicit includes → built-in patterns → `.gitignore` → custom/`.shotgunignore`. An ancestor exclusion blocks re-inclusion of nested paths.
3. `contextgen.DefaultContextGenerator.GenerateWithProgressEx()` — tree render + file content assembly (binary skip, language detection).
4. `contextgen/template.go` renders the final assembled context (Go `text/template`); the user-facing prompt template is loaded earlier via `internal/core/template`.
5. `tokens.EstimateFromBytes()` / `tokens.FormatTokens()`.
6. Output saved by `app` (headless) or `WizardModel.saveGeneratedContent()` (TUI); clipboard optional.

### LLM Provider Strategy

`llm.Provider` (core interface) ← `llmbase.BaseClient` (shared HTTP/auth/progress) + `llmbase.Sender` strategy implemented per provider (`BuildRequest`, `NewResponse`, `ParseResponse`, `GetEndpoint`, `GetHeaders`, `GetProviderName`). All HTTP goes through `platform/http.JSONClient`.

Two registries exist and both register openai/anthropic/gemini: [cmd/providers.go](file:///home/diogo/dev/shotgun-cli/cmd/providers.go) (headless path) and [internal/app/providers.go](file:///home/diogo/dev/shotgun-cli/internal/app/providers.go) (service path). Adding a provider means touching both.

## Key Directories

| Directory | Purpose |
|-----------|---------|
| `cmd/` | Cobra commands, viper/env init, provider factory, LLM config builder |
| `internal/app/` | `ContextService` (Generate, SendToLLM), config bridging, provider registry |
| `internal/config/` | `keys.go` constants, `metadata.go` defaults/categories, `validator.go` |
| `internal/core/scanner/` | `Scanner` interface, `FileNode`, `ScanConfig`, selection helpers |
| `internal/core/ignore/` | Layered ignore engine + built-in pattern list |
| `internal/core/contextgen/` | Context assembly, tree rendering, content collection |
| `internal/core/template/` | Template sources (embedded/user/custom), `{VAR}` renderer, metadata |
| `internal/core/tokens/` | Heuristic token estimation and formatting |
| `internal/core/selection/` | Per-project deselection persistence (one file per project) |
| `internal/core/diff/` | Intelligent diff splitting at file boundaries |
| `internal/core/llm/` | `Provider` interface, `Config`/defaults, `Registry` |
| `internal/platform/llmbase/` | `BaseClient` + `Sender` strategy interface |
| `internal/platform/{openai,anthropic,geminiapi}/` | Provider clients, models, DTOs |
| `internal/platform/http/` | Shared `JSONClient` and `HTTPError` |
| `internal/ui/` | Wizard orchestrator, config wizard, scan/generate coordinators |
| `internal/ui/{screens,components,styles}/` | Step screens, widgets, theme |
| `internal/assets/templates/` | Embedded built-in prompt templates |
| `test/e2e/`, `test/fixtures/sample-project/` | E2E tests and the fixture project they scan |

## Development Commands

```bash
# Build
make build                 # go build -ldflags "$(LDFLAGS)" -o build/shotgun-cli .
make build-all             # linux/darwin{amd64,arm64} + windows/amd64 → build/shotgun-cli-$os-$arch

# Test
make test                  # go test ./...
make test-race             # go test -race ./...          (preferred locally)
make test-e2e              # depends on build; go test ./test/e2e -v
make test-bench            # go test -bench=. -run=^$ ./...
go test -v -run TestFoo ./internal/core/scanner/...

# Quality
make lint                  # golangci-lint run --config .golangci.yml ./...
make fmt                   # go fmt ./...
make fmt-check             # gofmt -l . ; fails if non-empty (CI gate)
make vet                   # go vet ./...
make coverage              # go test -coverprofile=coverage.out ./... && go tool cover -func

# Run / install
go run . --help
./build/shotgun-cli
make install               # → ~/.local/bin/shotgun-cli
make uninstall

# Release
make version-patch|version-minor|version-major|version-set VERSION=x.y.z
make release-snapshot      # goreleaser release --snapshot --clean
make release VERSION=1.2.3 # release-tag + release-push
```

Build metadata is injected via ldflags into package `cmd`: `version`, `commit`, `date`, `builtBy`. There is no `make docker` target — container images are built by goreleaser.

## CLI Surface

| Command | Notable flags |
|---|---|
| `shotgun-cli` (no args) | launches TUI wizard |
| `context generate` | `--root/-r .`, `--include/-i *`, `--exclude/-e`, `--output/-o`, `--max-size 10MB`, `--enforce-limit true`, `--template/-t`, `--task`, `--rules`, `--var/-V KEY=VALUE`, `--include-hidden`, `--include-ignored`, `--progress none\|human\|json` |
| `context send [file]` | `--output/-o`, `--model/-m`, `--timeout`, `--raw` |
| `config` / `config show` / `config set <key> <value>` | bare `config` opens the config TUI |
| `llm status` / `llm doctor` / `llm list` | — |
| `template list\|render\|import\|export` | `render`: `--var`, `--output/-o` |
| `diff split` | `--input/-i` (required), `--output-dir/-o chunks`, `--approx-lines 500`, `--no-header` |
| `completion [bash\|zsh\|fish\|powershell]` | — |

Global flags: `--config <file>`, `--verbose/-v`, `--quiet/-q`, `--version`.

## Configuration

Keys are constants in [internal/config/keys.go](file:///home/diogo/dev/shotgun-cli/internal/config/keys.go); defaults/types in `metadata.go`; validation in `validator.go`.

| Key | Default | Notes |
|---|---|---|
| `scanner.max-files` | `10000` | positive int; rejects size strings |
| `scanner.max-file-size` | `1MB` | size format |
| `scanner.skip-binary` | `true` | |
| `scanner.include-hidden` | `false` | |
| `scanner.include-ignored` | `false` | |
| `scanner.respect-gitignore` | `true` | |
| `scanner.respect-shotgunignore` | `true` | |
| `context.include-tree` | `true` | |
| `context.include-summary` | `true` | |
| `context.max-size` | `10MB` | size format |
| `template.custom-path` | `""` | expands `~/` |
| `output.format` | `markdown` | `markdown\|text` |
| `output.clipboard` | `true` | |
| `llm.provider` | `openai` | `openai\|anthropic\|gemini` |
| `llm.api-key` | `""` | |
| `llm.base-url` | `""` | empty or `http(s)://` |
| `llm.model` | `""` | |
| `llm.timeout` | `300` | 1–3600 seconds |
| `llm.save-response` | `true` | |
| `verbose` / `quiet` | `false` | |

Resolution: `--config` file if given, else search platform config dir (`%APPDATA%\shotgun-cli`, `~/Library/Application Support/shotgun-cli`, or `$XDG_CONFIG_HOME/shotgun-cli`), then `$HOME`, then `.`; file is `config.yaml`. Env prefix `SHOTGUN_` with dots → underscores (`SHOTGUN_LLM_PROVIDER`). Precedence: flags > env > file > defaults.

## Templates

Built-in templates are `go:embed`-ed from `internal/assets/templates/*.md`: `prompt_analyzeBug.md`, `prompt_makePlan.md`, `prompt_makeDiffGitFormat.md`, `prompt_projectManager.md`. User templates load from `$XDG_CONFIG_HOME/shotgun-cli/templates`, plus an optional dir from `template.custom-path`; later sources override earlier ones by name.

Substitution syntax is **`{VARIABLE_NAME}`**, not Go template syntax. Built-in variables: `TASK`, `RULES`, `FILE_STRUCTURE`, `CURRENT_DATE` (auto-supplied). Missing required variables fail validation before render.

## Code Conventions & Common Patterns

**Imports** — three blank-line-separated groups: stdlib, third-party, `github.com/quantmind-br/shotgun-cli/...`.

**Config keys** — never raw strings:
```go
provider := viper.GetString(config.KeyLLMProvider)   // not "llm.provider"
```

**Logging** — zerolog only, structured, never `fmt.Println` for diagnostics:
```go
log.Info().Str("root", cfg.RootPath).Msg("Starting context generation...")
log.Error().Err(err).Str("config", name).Msg("Failed to load config")
```

**Error wrapping** — always add context:
```go
return fmt.Errorf("failed to save output: %w", err)
```

**Dependency injection** — functional options, no globals:
```go
svc := app.NewContextService(
    app.WithScanner(mockScanner),
    app.WithGenerator(mockGenerator),
    app.WithSelectionStore(store),
)
```

**Async in the TUI** — coordinator pattern; `Start()` returns a `tea.Cmd`, `Poll()` re-arms itself, `Result()` is mutex-guarded:
```go
return tea.Batch(m.fileSelection.Init(), m.scanCoordinator.Start(msg.rootPath, msg.config))
```

**Progress callbacks** — two styles, both mandatory on the TUI path (dropping them stalls the UI):
- channel-based: `ScanWithProgress(root, cfg, progress chan<- Progress)`
- callback-based: `GenerateWithProgress(ctx, cfg, func(stage, msg string, cur, total int64))`, `SendWithProgress(ctx, content, func(stage string))`

**Naming**

| Category | Pattern | Example |
|---|---|---|
| Interface | noun | `Scanner`, `Provider`, `ContextGenerator`, `TemplateSource` |
| Implementation | `<Impl><Name>` / `Default<Name>` | `FileSystemScanner`, `DefaultContextGenerator`, `LayeredIgnoreEngine` |
| Config struct | `<Name>Config` | `ScanConfig`, `GenerateConfig`, `LLMSendConfig` |
| Bubble Tea message | `<Action>Msg` | `ScanProgressMsg`, `GenerationCompleteMsg`, `LLMErrorMsg` |
| TUI model | `<Name>Model` | `WizardModel`, `FileSelectionModel` |
| Test double | `mock<Interface>` | `mockScanner`, `mockProvider`, `mockContextService` |

## Important Files

- [main.go](file:///home/diogo/dev/shotgun-cli/main.go) — entry point; zerolog console setup → `cmd.Execute()`
- [cmd/root.go](file:///home/diogo/dev/shotgun-cli/cmd/root.go) — root command, viper/env init, TUI launch
- [cmd/context.go](file:///home/diogo/dev/shotgun-cli/cmd/context.go) / [cmd/send.go](file:///home/diogo/dev/shotgun-cli/cmd/send.go) — headless generate / send
- [internal/app/service.go](file:///home/diogo/dev/shotgun-cli/internal/app/service.go) — generate + send orchestration, output, clipboard, token estimate
- [internal/app/context.go](file:///home/diogo/dev/shotgun-cli/internal/app/context.go) — `ContextService` interface and config/result types
- [internal/core/scanner/filesystem.go](file:///home/diogo/dev/shotgun-cli/internal/core/scanner/filesystem.go) — the walk
- [internal/core/ignore/engine.go](file:///home/diogo/dev/shotgun-cli/internal/core/ignore/engine.go) — ignore precedence + built-ins
- [internal/core/contextgen/generator.go](file:///home/diogo/dev/shotgun-cli/internal/core/contextgen/generator.go) — assembly pipeline
- [internal/core/template/manager.go](file:///home/diogo/dev/shotgun-cli/internal/core/template/manager.go) — source aggregation
- [internal/ui/wizard.go](file:///home/diogo/dev/shotgun-cli/internal/ui/wizard.go) — 5-step state machine
- [internal/ui/scan_coordinator.go](file:///home/diogo/dev/shotgun-cli/internal/ui/scan_coordinator.go) / [generate_coordinator.go](file:///home/diogo/dev/shotgun-cli/internal/ui/generate_coordinator.go) — async coordination

## Runtime/Tooling Preferences

- **Go 1.24.0** (`go.mod`; CI uses `1.24.x`). No `toolchain` directive.
- **Package manager**: Go modules.
- **Direct deps**: cobra `v1.10.2`, viper `v1.21.0`, bubbletea `v1.3.5`, bubbles `v0.21.0`, lipgloss `v1.1.0`, zerolog `v1.33.0`, go-gitignore, `adrg/xdg v0.5.3`, `atotto/clipboard v0.1.4`, `golang.org/x/text v0.32.0`, testify `v1.11.1`.
- **Tools**: `make`, `golangci-lint` (CI pins `v2.12.2` via action v8), `goreleaser` v2, `docker` (release images), `git`.
- **Lint config** (`.golangci.yml`, schema v2): explicitly enables `goconst`, `gocyclo`, `gosec`, `lll`, `misspell`, `prealloc`, `unconvert`, `unparam` *on top of* the v2 defaults (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`). Line length 120, gocyclo 25, goconst min-len 3 / min-occurrences 3 with `ignore-tests`, gosec excludes `G101`/`G306`. `_test.go` is exempt from `goconst`/`lll`/`gosec`; `internal/ui/` is exempt from `goconst`/`unparam`. Legacy/common-false-positive exclusion presets are deliberately **not** enabled.

## Testing & QA

**Layout**: 66 `*_test.go` files alongside source, 3 in `test/e2e/`, 8 inside the fixture project. All tests are **same-package** (no `_test` package suffix), so unexported helpers are tested directly. Heaviest suites: `internal/ui` (21 files), `internal/core` (15), `cmd` (8).

**Frameworks**: stdlib `testing` everywhere; `stretchr/testify` `assert`/`require` for most non-trivial checks; `net/http/httptest` for `platform/http` and every provider client; `bubbletea` for UI tests. No golden files, no build tags.

**Patterns**: table-driven with `t.Run()`, liberal `t.Parallel()`, `t.TempDir()` for filesystem isolation, and hand-written mocks:

| Mock | Interface | Location |
|---|---|---|
| `mockScanner` | `scanner.Scanner` | `internal/app/service_test.go` |
| `mockGenerator` | `contextgen.ContextGenerator` | `internal/app/service_test.go`, `internal/ui/generate_coordinator_test.go` |
| `mockProvider` / `mockLLMProvider` | `llm.Provider` | `internal/core/llm/registry_test.go`, `internal/app/*` |
| `mockContextService` | `app.ContextService` | `internal/ui/wizard_test.go` |
| `mockTemplateSource` | `template.TemplateSource` | `internal/core/template/manager_test.go` |
| `mockDirEntry` / `mockFileInfo` | `fs.DirEntry` / `os.FileInfo` | `internal/core/scanner/scanner_test.go` |

**Environment-guarded tests** (skip conditions are intentional — do not "fix" them):
- `TestCopySuccess`, `BenchmarkCopy` — skip when the clipboard is unavailable
- `requireColorPalette()` in `internal/ui/styles/theme_test.go` — skips under `NO_COLOR`
- `TestIntelligentSplit_ChunksApplyWithGit` — skips without `git`
- `TestScannerHandlesPermissionError`, `TestLoadIgnoreFiles_UnreadableDirectory`, `TestStore_Save_UnwritableDirReportsError` — skip when running as root
- `TestWriteDiffChunk_ReportsWriteError` — skips without `/dev/full`
- OS-gated: `TestGetConfigDir*`, `TestGetDefaultConfigPath_Windows`
- CI-only exclusion: `-skip "TestWizardClipboardCopyCmd"` (no clipboard utility on the runner)

**CI gates** (`.github/workflows/test.yml`, on push/PR to `main`/`master`): `make fmt-check` → `make vet` → `make build` → `go test -v -race -covermode=atomic -coverprofile=coverage.out -skip "TestWizardClipboardCopyCmd" ./...` → Codecov upload → **fail if total coverage < 80%**. Separate `Lint` and `Build` jobs run golangci-lint and `go build -v ./...`. Releases are tag-triggered (`v*`) and run goreleaser with `--skip=homebrew,docker,sign,sbom`.

Target 85%+ coverage overall, 90%+ for new code; 80% is the hard CI floor.

## Anti-Patterns

- Importing `viper` in `core/` or `platform/` (use DI); do not add new viper reads in `ui/` either.
- Global mutable state in `internal/`.
- Dropping progress callbacks on the TUI path — the UI stalls.
- Direct `net/http` in providers — use `platform/http.JSONClient` via `llmbase.BaseClient`.
- Registering a new LLM provider in only one of the two registries.

## Known Documentation Drift

`README.md` and `openwiki/` predate several changes. When they conflict with code, **the code wins**:

- The send command is `shotgun-cli context send`, not `shotgun-cli send` (README, quickstart, workflows).
- The config TUI is `shotgun-cli config`; there is no `--interactive` flag and no `config show --format json`.
- Template syntax is `{VAR}`, not `{{ .Var }}` — wrong in `openwiki/domains.md` and in `template render` help text. `template_selection.go` copy still mentions `{{FILES}}`; the real variable is `FILE_STRUCTURE`.
- Ignore precedence puts `.gitignore` before the custom/`.shotgunignore` layer, the reverse of `openwiki/domains.md`. There is no repo-local `.shotgun/templates/` source.
- Stale defaults: `llm.provider` is `openai` (not blank/`gemini`), `scanner.max-files` is `10000`, `scanner.max-file-size` is `1MB`, `output.clipboard` is `true`, `context.max-size` is `10MB`.
- `cmd/send.go` help text still says Gemini-only; the implementation uses the configured provider registry.
- The Gemini platform package is `geminiapi`, not `gemini`.

## OpenWiki

Longer-form docs live in `/openwiki`: [quickstart](openwiki/quickstart.md) (current), `architecture.md` (current), `workflows.md` (partly stale), `domains.md` and `operations.md` (materially stale — see drift list above). Read the quickstart first, then follow its links, and verify specifics against source.
