# Code Improvements — Decisions of Record

**Analyzed:** 2026-07-28 | **Repository:** `shotgun-cli` @ `c2376b3` | **Reviewed:** 2026-07-28

> This file records what was decided, in the order it should be built. Every claim below was re-verified against source during critical review. Four items were rescoped; the corrections are noted inline under **Correction on review**. Build specs live in `01-PRD_ci_matrix_and_dead_code.md`, `02-PRD_llm_send_consolidation.md`, `03-PRD_config_and_output_single_sourcing.md`, and `04-PRD_tui_send_progress_ticker.md`.

## Executive Summary

Scope: the whole repository (`cmd/`, `internal/{app,core,platform,ui,config,utils}`, `test/e2e/`, CI and Makefile tooling).

The codebase has three strong, already-implemented consolidation mechanisms: the **config metadata registry** (`internal/config/metadata.go`), the **LLM provider registry + `ContextService`** (`internal/app/`), and the **`llmbase.BaseClient` + `Sender` strategy** (`internal/platform/llmbase/`). Recent history shows a deliberate campaign of routing duplicated logic through single producers — `BuildGeneratorConfig` (`9eca9e2`), `setConfigDefaults` from metadata (`87521c8`), `ValidKeys` from metadata (`8a72c25`), `utils.ParseSize` (`4cfcf52`), `utils.FormatBytes` (`c4cb855`). Every item below is an unfinished edge of one of those campaigns, not a new idea.

Highest-impact findings: the headless `context send` path bypasses `ContextService` entirely and keeps a **second, redundant provider registry** alive (CI-001); the metadata registry migration stopped short of validation, leaving `MinValue`/`MaxValue` declared but **never enforced** (CI-003); and the Review screen's elapsed-time counter is **rendered but never re-rendered**, so it sits frozen at `(0s)` for the whole request (CI-002).

Limitations: no runtime profiling and no live LLM calls. All evidence is static source, git history, the test suite, `Makefile`, and `.github/workflows/test.yml`.

## Existing Patterns Being Extended

### P-001: Single-producer consolidation for cross-front-end config
- **Evidence:** `internal/app/generator_config.go:33-50` (`BuildGeneratorConfig`), `internal/app/generator_config.go:56-67` (`ScannerLimits`)
- **Extension Boundary:** Any other value both front ends compute independently — LLM config defaults, output filenames, send orchestration.

### P-002: Metadata registry as the single source of config truth
- **Evidence:** `internal/config/metadata.go:152-320` (`buildAllMetadata`), `internal/config/validator.go:16-24` (`ValidKeys` derived), `cmd/root.go:245-249` (`setConfigDefaults` derived), `cmd/completion.go:108`, `internal/ui/components/config_select.go:23`
- **Extension Boundary:** `ValidateValue`/`ConvertValue`, which still hand-maintain per-key `switch` statements.

### P-003: Provider registry + `ContextService` as the LLM composition point
- **Evidence:** `internal/core/llm/registry.go:25-42`, `internal/app/providers.go:12-22`, `internal/app/service.go:37`, `internal/app/service.go:244-290`
- **Extension Boundary:** The headless `context send` command, which reimplements all of it against its own registry.

### P-004: `llmbase.BaseClient` + `Sender` strategy for providers
- **Evidence:** `internal/platform/llmbase/base_client.go:127-153` (`Send`), `internal/platform/llmbase/sender.go:12-34`, `internal/platform/llmbase/base_client.go:166-174` (`HandleHTTPError`)
- **Extension Boundary:** Provider default values (base URL, model, timeout), which each client still declares privately.

### P-005: Registry-injection for testable orchestration
- **Evidence:** `internal/app/service.go:46-50` (`WithRegistry`), `internal/app/service_llm_test.go:16-47` (`mockLLMProvider`, `newMockRegistry`), `internal/app/service_llm_test.go:75`
- **Extension Boundary:** `cmd/send.go`, which reaches a package-level global and therefore cannot be tested against a fake.

### P-006: Coordinator-mediated async messaging in the TUI
- **Evidence:** `internal/ui/scan_coordinator.go` (single-flight, per-run channels, `tea.Tick` poll chain), `internal/ui/generate_coordinator.go`
- **Extension Boundary:** The LLM send, which runs as a single blocking `tea.Cmd` and therefore emits no intermediate messages at all.

---

## Phase 1 — Zero-behavior-change groundwork

Land first: neither item changes runtime behavior, and the CI matrix gives every later phase cross-platform signal. Spec: `01-PRD_ci_matrix_and_dead_code.md`.

### CI-007: Extend the CI test job to a macOS/Windows matrix

**Category:** tooling | **Effort:** small | **Confidence:** high

Three platform-specific code paths are asserted by tests that never execute in CI, while `make build-all` ships binaries for exactly those platforms. The matrix makes existing assertions real instead of silently skipped.

**Correction on review — scope narrowed.** The original proposal was a bare `strategy.matrix` addition with `-race` and coverage running on all three runners. That triples wall-clock for signal that is not OS-sensitive: the race detector tests Go-runtime concurrency, which is identical across platforms, and on `windows-latest` it additionally requires a C toolchain. The matrix's actual value is **path, permission and filesystem semantics**. Adjusted scope:

- Add `strategy.matrix.os: [ubuntu-latest, macos-latest, windows-latest]` with `fail-fast: false`, so one platform failing does not mask another's result.
- Keep `-race`, `-covermode=atomic`, the Codecov upload, and the 80% threshold gate **on `ubuntu-latest` only**. Non-Linux runners run the plain suite.
- The Linux job keeps the single existing coverage gate; nothing about the 80% floor changes.

**Evidence:**
- `cmd/root_test.go:250-273` — three sibling assertions guarded by `runtime.GOOS` checks, covering the `getConfigDir()` branches at `cmd/root.go:211-220`. `cmd/config_test.go:234` and `:243` add two more Windows guards.
- `.github/workflows/test.yml:14-16` — the `test` job declares `runs-on: ubuntu-latest` with no matrix.
- `Makefile:37` — `build-all` cross-compiles for `darwin` and `windows`.

**Affected Areas:** `.github/workflows/test.yml`

**Risks:** Will surface latent Windows path/permission failures on first run — which is the point, and why it lands before any behavior change. `fail-fast: false` ensures the first run reports every platform's failures at once rather than one per iteration. GitHub Actions minutes are free for this public repository, so runner cost is not a constraint.

---

### CI-008: Remove the dead `app.CLIConfig` / `app.ProgressMode` / `app.ProgressOutput` triplicate

**Category:** consolidation | **Effort:** trivial | **Confidence:** high

`internal/app/config.go` looks like the shared front-end config contract, and `internal/app/AGENTS.md` documents it as one, but it is unreachable — the live definitions sit in `cmd`. It also carries pre-rename Gemini-specific field names (`SendGemini`, `GeminiModel`, `GeminiOutput`, `GeminiTimeout`) that contradict the provider-agnostic model the rest of the code now uses.

**Correction on review — dependency dropped, confidence raised.** The original entry sequenced this after CI-001 at `medium` confidence, on the reasoning that CI-001 should first settle "what belongs in `app` as a front-end config type". That dependency is philosophical, not technical: CI-001 touches `cmd/send.go`, `cmd/providers.go` and `cmd/llm.go`, none of which reference `app.CLIConfig`. Verified by grep — the only references anywhere are the declaration and its own test. The item is standalone and zero-risk.

Coverage impact is nil: `internal/app/config.go` declares only types and constants, contributing no statements to the coverage denominator.

**Evidence:**
- `internal/app/config.go:5-30` (`CLIConfig`), `:33-42` (`ProgressMode`), `:46-53` (`ProgressOutput`).
- Repository-wide grep for `CLIConfig`, `app.ProgressMode`, `app.ProgressOutput`: hits only `internal/app/config.go` and `internal/app/config_test.go:9-12`.
- The live, used definitions are byte-identical duplicates in the presentation layer: `cmd/context.go:25-32` (`ProgressMode`) and `cmd/context.go:34-42` (`ProgressOutput`), consumed at `cmd/context.go:264-279` and `:373-406`.

**Affected Areas:** `internal/app/config.go`, `internal/app/config_test.go`, `internal/app/AGENTS.md`

**Risks:** None. The progress types stay in `cmd/context.go` — progress *rendering* is a presentation concern and `cmd` is its only consumer.

---

## Phase 2 — LLM send path consolidation

The structural keystone. CI-001 must land before CI-005 and CI-006. Spec: `02-PRD_llm_send_consolidation.md`.

### CI-001: Route `context send` through `ContextService` and delete the duplicate provider registry

**Category:** consolidation | **Effort:** medium | **Confidence:** high

Registering a new LLM provider becomes one edit instead of two, eliminating a silent class of bug where a provider works in the TUI and reports `unsupported provider` headlessly. Headless send gains the response-saving and validation semantics the service already guarantees, and becomes injectable for tests (CI-006).

**Evidence:**
- `internal/app/providers.go:12-22` — registers openai/anthropic/gemini into `DefaultProviderRegistry`; `internal/app/service.go:37` wires it as the service default.
- `cmd/providers.go:15-32` — a second `init()` registering the identical three factories into a private `providerRegistry`, exposed via `CreateLLMProvider` (`cmd/providers.go:35-42`) and `GetProviderRegistry` (`:44-47`).
- `cmd/send.go:94-137` — reimplements the exact sequence `ContextService.SendToLLMWithProgress` already performs at `internal/app/service.go:259-287`: create provider, `IsAvailable()`, `ValidateConfig()`, `Send()`, write the response with `0600`.
- Documented intent contradicting current code: `.cursor/rules/go-patterns.mdc:130` states "Always use `app.DefaultProviderRegistry` to create LLM providers. This ensures consistency across CLI and TUI"; `internal/AGENTS.md:61` lists "Add LLM provider → `platform/<name>/` + `app/providers.go`" as the single registration point.

**Affected Areas:** `cmd/send.go`, `cmd/providers.go` (deleted), `cmd/llm.go` (2 call sites), `cmd/send_test.go`, `cmd/llm_test.go` (8 test functions to repoint or delete)

**Implementation notes:**
1. Delete `cmd/providers.go`; repoint `cmd/llm.go:85` and `cmd/llm.go:151` at `app.DefaultProviderRegistry.Create(cfg)`.
2. Replace `cmd/send.go:94-137` with a `ContextService.SendToLLMWithProgress` call. `--raw` handling stays in the command: `LLMSendConfig` saves `result.Response` only (`internal/app/service.go:284`), so a `--raw` run must write `result.RawResponse` itself with `SaveResponse` left off.
3. While rewriting `cmd/send.go`, fix its Gemini-only help text (`Short`, `Long`, and the `--output`/`--model` flag descriptions all say "Gemini"; the command has used the configured provider since the registry landed).

**Risks:** `CreateLLMProvider` and `GetProviderRegistry` are exported from `cmd`; they are only reachable in-repo, so removal is safe, but a fork importing them breaks. `--raw` semantics must be preserved explicitly.

---

### CI-005: Single-source provider defaults between `llm.DefaultConfigs()` and the platform clients

**Category:** consolidation | **Effort:** small | **Confidence:** high

`llm.DefaultConfigs()` is what `llm status` shows the user; the platform clients declare their own copies of the same three values. Collapsing them removes a latent class of defect where the diagnostic output and the dialled endpoint disagree.

**Correction on review — impact restated as latent, not live.** The original entry claimed `llm status` "currently reports a default base URL read from one table while the client actually dials a constant declared in another file". Verified: the two tables hold **identical values today** (openai `https://api.openai.com/v1` / `gpt-4o` / 300s; anthropic `https://api.anthropic.com` / `claude-sonnet-4-20250514` / 300s; gemini `https://generativelanguage.googleapis.com/v1beta` / `gemini-2.5-flash` / 300s). Further, `cmd/config_llm.go:14-34` and `internal/app/service.go:257` both fill `BaseURL`/`Model`/`Timeout` from `DefaultConfigs()` *before* the client is constructed, so the client-local constants are unreachable fallbacks on both live paths. Nothing is observably wrong today. The item is worth doing to remove the drift hazard and delete a third copy of the defaults block — not to fix a live bug.

**Evidence:**
- `internal/core/llm/config.go:25-46` — `DefaultConfigs()` declares base URL, model, and timeout for all three providers.
- `internal/platform/openai/client.go:13,22-25`; `internal/platform/anthropic/client.go:14-31`; `internal/platform/geminiapi/client.go:14-30` — each declares the same values independently.
- `internal/platform/llmbase/base_client.go:60-98` — `NewBaseClientWithDefaults` already takes a `DefaultConfig` parameter; the injection point exists.
- `cmd/config_llm.go:26-34` hand-applies the same three defaults a third time, duplicating `llm.Config.WithDefaults()` (`internal/core/llm/config.go:90-104`).

**Affected Areas:** `internal/platform/{openai,anthropic,geminiapi}/client.go`, `cmd/config_llm.go`

**Risks:** Adds a `core/llm` import to each platform client. Permitted by the layer rules, and `llmbase` already imports `core/llm` (`internal/platform/llmbase/sender.go:6`), so no new dependency direction is introduced. `MaxTokens` is absent from the core table and stays a package constant in anthropic/geminiapi.

**Depends on:** CI-001 (so `cmd/config_llm.go` has one consumer left to change).

---

### CI-006: Replace the tautological `context send` test with a registry-injected fake

**Category:** tests | **Effort:** small | **Confidence:** high

`TestRunContextSend_FromFile` cannot fail, and it issues a real outbound HTTPS request from the unit suite. Replacing it gives the headless send path the same real assertions the service path has, and removes network dependence from `go test ./...`.

**Correction on review — billing claim withdrawn.** The original entry stated the test "when a valid API key is present in the environment, issues a live billed API call". Verified false: `cmd/send_test.go:97-101` calls `viper.Reset()` and then `viper.Set("llm.api-key", "test-key")`. An explicit `viper.Set` has top precedence, so the request always carries `test-key` and always returns 401 — a real billed call is not reachable. The genuine defects are unchanged and still disqualifying:

- The assertion accepts `err == nil` unconditionally (`cmd/send_test.go:129`: "If err is nil, it means gemini is working and the test succeeded"), and the eight-alternative `strings.Contains` disjunction includes `"request failed"`, which is the prefix `cmd/send.go:119` puts on *every* send error — so the failure branch is also unfailable. Three of the eight alternatives ("gemini integration is disabled", "LLM integration is disabled", "gemini request failed") match no string in the current codebase.
- The test makes a real DNS + TLS round trip to `https://api.openai.com/v1` on every `go test ./cmd/...`, making the unit suite network-dependent and slow.
- `viper.Reset()` inside the subtest clears global config state for every later test in the `cmd` package.

**Evidence:**
- The working pattern to copy: `internal/app/service_llm_test.go:16-47` defines `mockLLMProvider` and `newMockRegistry`; `:75` injects it via `NewContextService(WithRegistry(registry))`.
- The blocker is structural: `cmd/send.go:94` calls `CreateLLMProvider`, which reads the package-level global at `cmd/providers.go:13` with no injection seam.

**Affected Areas:** `cmd/send_test.go`, `cmd/send.go`

**Implementation notes:** After CI-001, give `runContextSend` a service seam (a package-level `newSendService` function variable) so a test can supply `app.NewContextService(app.WithRegistry(...))`. Assert the actual contract: response written to `--output`; `--raw` selecting `RawResponse` over `Response`; the auto-generated `llm-response-<timestamp>.md` filename when `llm.save-response` is true (`cmd/send.go:83-88`); and the exact error when the provider reports unavailable. Delete `isExpectedProviderError` (`cmd/send_test.go:15-30`), which exists only to serve the disjunction.

**Depends on:** CI-001 (which supplies the injection seam).

---

## Phase 3 — Config and output single-sourcing

Independent of Phase 2; can land in parallel. Spec: `03-PRD_config_and_output_single_sourcing.md`.

### CI-003: Derive `ValidateValue`/`ConvertValue` from the metadata registry

**Category:** configuration | **Effort:** medium | **Confidence:** high

Completes the registry migration started in `8a72c25`/`87521c8`: adding a config key becomes one edit instead of three (metadata + `ValidateValue` arm + `ConvertValue` arm). Fixes a live drift — `scanner.max-files` declares `MaxValue: 1000000` that nothing enforces, so `config set scanner.max-files 999999999` is accepted today.

**Evidence:**
- `internal/config/validator.go:16-24` — `ValidKeys()` already derives from `AllConfigMetadata()`, with the comment "Registering a key in `buildAllMetadata()` is the single edit needed to make it usable."
- `internal/config/validator.go:54-82` and `:85-104` — `ValidateValue` and `ConvertValue` still enumerate every key by hand in parallel `switch` statements; the boolean arm alone lists eleven keys twice.
- Declared-but-unenforced metadata: `internal/config/metadata.go:156-163` sets `MinValue: 1, MaxValue: 1000000` for `scanner.max-files`, while `validateMaxFiles` (`internal/config/validator.go:107-127`) checks positivity only. A repository-wide grep for `MinValue`/`MaxValue` outside `metadata.go` returns no production reads — only `EnumOptions` is consumed (`cmd/completion.go:108`, `internal/ui/components/config_select.go:23`).
- Enum lists duplicated against metadata: `internal/config/validator.go:197` rebuilds the provider list already declared at `internal/config/metadata.go:263`; `:148` duplicates `metadata.go:246`. `internal/config/validator.go:189` hardcodes `3600`, already declared as `MaxValue` at `metadata.go:293`.

**Affected Areas:** `internal/config/validator.go`, `internal/config/validator_test.go`

**Implementation notes:** Dispatch on `GetMetadata(key).Type`. Two behaviours must be preserved deliberately: `ValidateValue`'s current `default:` arm returns `nil` for an unrecognised key (`IsValidKey` is the separate gate), so a metadata miss must also return `nil` rather than an error; and `validateMaxFiles`'s size-format rejection carries a distinct error message, so it stays as a `TypeInt` pre-check. Add a metadata-driven table test asserting every key in `AllConfigMetadata()` round-trips through `ValidateValue(key, fmt.Sprint(m.DefaultValue))` without error — that test is what prevents the next drift.

**Risks:** Enforcing `MaxValue` on `scanner.max-files` is a behavior change: a `config set` above 1,000,000 will now fail (viper reads on load are unaffected). `internal/config` must stay stdlib-only per `internal/AGENTS.md`; deriving enums from in-package `EnumOptions` respects that, whereas importing `core/llm` would not.

---

### CI-004: Single-source the default output filename and standardize the file mode

**Category:** consolidation | **Effort:** small | **Confidence:** high

The `shotgun-prompt-<timestamp>.md` convention stops being three independent literals, and the same product artifact stops being written `0600` headlessly and `0644` from the TUI.

**Correction on review — directory semantics explicitly excluded.** The original entry treated the three sites as pure duplicate literals. They are not: `internal/ui/wizard.go:1265-1266` does `filepath.Join(m.rootPath, filename)`, writing into the scanned project root, while `cmd/context.go:213-217` and `internal/app/service.go:198` produce a bare filename resolved against the process CWD. Unifying the *name* is safe; unifying the *directory* would silently relocate every TUI user's output. Adjusted scope: extract the filename helper only, and leave each front end's directory choice exactly as it is.

The mode decision is settled rather than left open: standardize on **`0600`**. Generated context files routinely embed the user's entire private source tree, and the `#nosec G306` justification at `internal/ui/wizard.go:1268` ("Generated context files are meant to be world-readable") is not a defensible reason to widen permissions on that content. `0600` also matches what both headless write sites already do, so the change is TUI-only.

**Evidence:**
- `internal/app/context.go:104-112` — `GenerateOutputPath()` produces `shotgun-prompt-%s.md` from a `20060102-150405` timestamp.
- `cmd/context.go:213-217` — the headless command builds the identical filename itself before handing `Output` to the service, so `GenerateOutputPath`'s default branch is dead on that path.
- `internal/ui/wizard.go:1263-1273` — `saveGeneratedContent` builds the same filename a third time and writes with `0644`.
- Divergent modes for the same artifact: `internal/app/service.go:199` uses `0600`, `cmd/send.go:131` uses `0600`, `internal/ui/wizard.go:1269` uses `0644` behind a `#nosec G306` comment; `.golangci.yml:32` globally excludes `G306`, so the linter cannot flag the divergence.

**Affected Areas:** `internal/app/context.go`, `cmd/context.go`, `internal/ui/wizard.go`, `internal/ui/wizard_test.go`

**Risks:** Changing the TUI's mode from `0644` to `0600` is user-visible for anyone whose workflow reads the artifact as another user. Record the choice in the commit message.

---

## Phase 4 — TUI send feedback

Independent of every other phase. Spec: `04-PRD_tui_send_progress_ticker.md`.

### CI-002: Drive the Review screen's elapsed-time counter during an LLM send

**Category:** pattern extension | **Effort:** small | **Confidence:** high

During a send the Review screen shows `⏳ Sending to LLM... (0s)` and the counter never advances — for up to the configured 300s timeout. The screen already computes and renders elapsed time; nothing re-renders it.

**Correction on review — mechanism replaced.** The original entry proposed passing a non-`nil` progress callback at `internal/ui/wizard.go:927` in place of `nil`, described as "a one-argument change". Verification shows that fix does not work and would not deliver the stated outcome:

1. **A callback cannot emit messages.** `sendToLLMCmd` (`internal/ui/wizard.go:922-938`) is a `tea.Cmd` — a function returning exactly **one** `tea.Msg`. A progress callback firing inside that closure has nowhere to deliver a message; Bubble Tea consumes only the single returned value. Delivering N messages requires a channel + `tea.Tick` poll chain (pattern P-006) or a retained `*tea.Program` for `Send()`. The wizard has neither for the LLM path, so the change is not one argument.
2. **The provider emits two stages, both useless here.** `internal/platform/llmbase/base_client.go:156-163` emits `"Connecting to <provider>..."` immediately and `"Response received"` *after* `Send` returns — microseconds before `LLMCompleteMsg`. There is no mid-flight signal to forward.
3. **The receiving side would discard it anyway.** `handleLLMProgress` (`internal/ui/wizard.go:857-864`) writes `msg.Stage` into `m.progress.Stage` and then hardcodes `UpdateMessage("", "Sending to LLM...")`, so a delivered stage string changes no rendered output.
4. **Delivering progress would actively regress the display.** `internal/ui/screens/review.go:147-150` resets `m.llmStartTime = time.Now()` on every `LLMProgressMsg`, so any mid-flight progress message would reset the elapsed counter to zero.

Adjusted scope — deliver the actual user-visible outcome with the simple mechanism:

- Start a repeating 1-second `tea.Tick` when a send begins and stop it on `LLMCompleteMsg`/`LLMErrorMsg`. Each tick re-renders, so `review.go:429`'s `time.Since(m.llmStartTime)` advances. This mirrors the existing tick at `internal/ui/wizard.go:826`.
- Remove the `m.llmStartTime = time.Now()` reset from the `LLMProgressMsg` arm at `internal/ui/screens/review.go:147-150`; `SetLLMSending(true)` already sets it once at send start.
- Explicitly **out of scope**: forwarding the provider callback and building an LLM coordinator. Two stages, one of which arrives at completion, do not justify a coordinator.

**Evidence:**
- `internal/ui/screens/review.go:427-430` — renders `⏳ Sending to LLM... (%s)` from `time.Since(m.llmStartTime).Round(time.Second)`.
- `internal/ui/wizard.go:927` — `SendToLLMWithProgress(ctx, m.generatedContent, cfg, nil)`; the command blocks for the whole request and returns one message, so no re-render occurs in between.
- `internal/ui/wizard.go:846-849` / `internal/ui/screens/review.go:467-473` — `SetLLMSending(true)` already stamps `llmStartTime` at send start.
- `internal/ui/wizard.go:822-829` — the existing `tea.Tick` precedent (`copyToastClearMsg`).

**Affected Areas:** `internal/ui/wizard.go`, `internal/ui/screens/review.go`, `internal/ui/wizard_test.go`

**Risks:** The tick must stop on both terminal messages or it leaks a repeating command for the life of the program. `go test -race ./internal/ui/...` covers the concurrency.

---

## Build Order

| # | Phase | Items | PRD |
|---|---|---|---|
| 1 | Zero-behavior-change groundwork | CI-007, CI-008 | `01-PRD_ci_matrix_and_dead_code.md` |
| 2 | LLM send path consolidation | CI-001 → CI-005, CI-006 | `02-PRD_llm_send_consolidation.md` |
| 3 | Config and output single-sourcing | CI-003, CI-004 | `03-PRD_config_and_output_single_sourcing.md` |
| 4 | TUI send feedback | CI-002 | `04-PRD_tui_send_progress_ticker.md` |

Phase 1 lands first so that any latent Windows/macOS failure is attributable to the matrix rather than to a concurrent refactor, and so every later phase gets cross-platform signal. Phases 2, 3 and 4 have no edges between them and may land in any order or in parallel.

## Summary

| ID | Opportunity | Category | Effort | Confidence |
|---|---|---|---|---|
| CI-007 | CI test matrix for macOS/Windows (Linux-only race + coverage) | tooling | small | high |
| CI-008 | Remove dead `app.CLIConfig`/`ProgressMode`/`ProgressOutput` | consolidation | trivial | high |
| CI-001 | Route `context send` through `ContextService`; delete duplicate registry | consolidation | medium | high |
| CI-005 | Single-source provider defaults | consolidation | small | high |
| CI-006 | Replace the tautological send test with an injected fake | tests | small | high |
| CI-003 | Derive validation from the metadata registry | configuration | medium | high |
| CI-004 | Single-source output filename; standardize mode to `0600` | consolidation | small | high |
| CI-002 | Drive the Review screen's elapsed-time counter during a send | pattern extension | small | high |

**Patterns extended:** 6
**Opportunities retained:** 8 (4 as originally scoped, 4 rescoped on review — CI-002, CI-004, CI-007, CI-008)
**Ideas dropped for insufficient evidence:** 4

Dropped, with reasons:
- *Split `internal/ui/wizard.go` (1327 lines, top churn hotspot)* — size and churn alone are a smell, not an opportunity; the screen-model composition pattern is already applied and no specific state was found that still belongs in a screen model.
- *Add e2e coverage for `diff split` and `config set`* — both have real in-package unit tests; no analogous regression or observable contract gap was established beyond "no e2e file exists".
- *Extract a shared result printer for `cmd/send.go` and the Review screen* — the two renderings differ materially (tabwriter/stdout vs. lipgloss TUI state); consolidating would abstract over two genuinely different outputs.
- *Retire the `contextgen.TemplateRenderer` Go-`text/template` engine in favor of the `{VAR}` renderer in `core/template`* — the two serve different stages (final context assembly vs. user-facing prompt templates) and both are live; no evidence establishes one as the intended replacement.
