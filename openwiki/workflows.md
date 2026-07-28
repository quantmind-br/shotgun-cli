# Workflows

## TUI Wizard (interactive mode)

When `shotgun-cli` runs with no arguments, it launches a 5-step interactive TUI wizard built with Bubble Tea.

```
main.go → cmd.Execute() → launchTUIWizard()
  → ui.NewWizard() → 5-step state machine
  → Init: ScanCoordinator starts async scan
```

### 5 steps

| Step | Screen | Behavior |
|------|--------|----------|
| 1 | **File Selection** | Shows scanned file tree. Space to toggle, `a`/`A` to select/deselect all visible, `i` to rescan with ignored files toggled, `/` to filter, `F5`/`r` to rescan. Excludes are persisted per-project (see Selection Persistence below). |
| 2 | **Template Selection** | Lists available templates. User picks one or skips (plain context). |
| 3 | **Task Input** | Free-text task description for the LLM. |
| 4 | **Rules Input** | Free-text rules/constraints for the LLM. |
| 5 | **Review** | Shows summary of selections, template, task, rules. User can edit, generate, or send to LLM. |

On advancing past Step 1, deselections are saved to `selectionStore.Save()`.

Source: `internal/ui/wizard.go`, `internal/ui/screens/`

### Async coordinators

- **ScanCoordinator** (`internal/ui/scan_coordinator.go`): Initiates an async filesystem scan with progress reporting. Sends `ScanProgressMsg` and `ScanCompleteMsg` back to the wizard model.
- **GenerateCoordinator** (`internal/ui/generate_coordinator.go`): Manages async context generation with progress callbacks.

### State machine

The wizard uses a linear step progression (`StepFileSelection = 1` → `StepReview = 5`). Each step can be navigated forward/backward. Validation is enforced at each step boundary (e.g., at least one file must be selected before advancing).

See wizard fields and `getNextStep()`/`canAdvanceStep()` in `internal/ui/wizard.go`.

---

## Headless CLI mode

### `context generate`

```
cmd/context.go → Build GenerateConfig → app.NewContextService()
  → svc.Generate(ctx, cfg) [synchronous]
  → Return result → print/save/copy
```

Flags: `--output`, `--max-size`, `--template`, `--task`, `--rules`, `--progress` (none/human/json), `--no-enforce-limit`, `--include-hidden`, `--include-ignored`.

Source: `cmd/context.go`

### `llm` commands

- `shotgun-cli llm status` — shows current provider, model, API key masked, timeout, status.
- `shotgun-cli llm doctor` — runs diagnostics on provider config, checks API key and availability.
- `shotgun-cli llm list` — lists supported providers (OpenAI, Anthropic, Gemini).

### `send`

`shotgun-cli send` reads a generated context file and sends it to the configured LLM provider.

### `config set/get`

Manages persistent configuration via Viper. Also supports an interactive TUI config editor when run with `--interactive`.

### `diff split`

Intelligently splits a diff into LLM-sized chunks.

---

## Context generation pipeline

Both TUI and headless CLI share this pipeline through `DefaultContextService.GenerateWithProgress()` (`internal/app/service.go`):

```
1. Validate config
2. Scan filesystem (parallel workers, respect .gitignore/.shotgunignore)
3. Apply selections (default to SelectAllExcept with saved deselections if none provided)
4. Generate context (tree rendering + file assembly + template substitution)
5. Enforce size limits (optional)
6. Save output to file
7. Copy to clipboard (optional)
```

### Selections resolution

If `cfg.Selections` is nil and a `selectionStore` is configured, the service loads the project's saved deselections and applies them via `scanner.SelectAllExcept()`. If no store or no saved deselections, all non-ignored files are selected.

Source: `internal/app/service.go` (lines 134-141)

---

## Selection persistence flow

The selection store (`internal/core/selection/store.go`) persists per-project file deselections so users don't need to re-exclude the same files on every scan.

### Architecture

```
JSON file (~/.config/shotgun-cli/selections.json)
  └── deselected (map[string][]string)
        └── "/path/to/project" → ["vendor/...", "dist/...", ...]
```

### Lifecycle

1. **Init**: `cmd/root.go` creates `selection.NewStore(selectionStorePath())` and attaches to the wizard or headless service.
2. **Load**: On scan complete, `handleScanComplete()` seeds the file selection model with `scanner.SelectAllExcept(tree, deselected)`.
3. **Save**: On advancing past the file-selection step, `persistSelections()` computes the diff of the deselected files using `scanner.CollectDeselected()` and writes it via `selectionStore.Save()`.
4. **Display sync**: After a rescan, any previously saved deselections are reapplied automatically.

### Key source files

| File | Role |
|------|------|
| `internal/core/selection/store.go` | `Store` struct with `Load()`/`Save()`, atomic file writes (write to `.tmp` then rename) |
| `internal/core/scanner/helpers.go` | `SelectAllExcept()` applies deselected list, `CollectDeselected()` computes what's deselected from current selections |
| `internal/app/service.go` | `WithSelectionStore()` option, deselection loading in `GenerateWithProgress()` |
| `internal/ui/wizard.go` | `selectionStore`/`deselected`/`selectionsSeeded` fields, `SetSelectionStore()`, `persistSelections()` |
| `internal/ui/screens/file_selection.go` | `SetSelections()` replaces selection map, `SetShowIgnored()` syncs display |
| `internal/ui/components/tree.go` | `SetShowIgnored(show bool)` replaces toggle behavior |
| `cmd/root.go` | Wizard selection store wiring, `selectionStorePath()` |
| `cmd/context.go` | Headless CLI selection store wiring |

---

## LLM send flow

```
Generate content
  → SendToLLMWithProgress(ctx, content, cfg, callback)
  → ProviderRegistry.Create(providerName)
  → Provider.Send(ctx, content)
  → Parse response → return result
```

The `ProviderRegistry` (`internal/core/llm/registry.go`) manages available providers. `DefaultProviderRegistry` (`internal/app/providers.go`) registers OpenAI, Anthropic, and Gemini.

Each provider implements the `Sender` interface through `BaseClient` (`internal/platform/llmbase/base_client.go`), which handles common HTTP send/retry/progress logic.
