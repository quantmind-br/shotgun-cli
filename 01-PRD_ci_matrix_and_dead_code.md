# PRD — CI Matrix and Dead Code

**Source**: IDEATION_CODE_IMPROVEMENTS.md
**Generated**: 2026-07-28

## Implementation Order

1. CI-007 — Add a macOS/Windows test matrix to `.github/workflows/test.yml`, keeping race detection and coverage on Linux only
2. CI-008 — Delete the unreachable `internal/app/config.go` and its test

---

## CI-007: CI Test Matrix for macOS and Windows

### Scope

**In scope**:

- Add `strategy.matrix.os: [ubuntu-latest, macos-latest, windows-latest]` to the `test` job in `.github/workflows/test.yml` and set `runs-on: ${{ matrix.os }}`.
- Set `fail-fast: false` on the matrix.
- Split the test step into two mutually exclusive steps: a Linux step retaining `-race -covermode=atomic -coverprofile=coverage.out`, and a non-Linux step running the same suite without race/coverage flags.
- Gate the `Upload coverage to Codecov` and `Check coverage threshold` steps on `matrix.os == 'ubuntu-latest'`.
- Preserve the existing `-skip "TestWizardClipboardCopyCmd"` argument on every runner.
- Preserve `make fmt-check`, `make vet` and `make build` on every runner.

**Out of scope**:

- The `lint` and `build` jobs — both stay `runs-on: ubuntu-latest`, unchanged.
- Changing the 80% coverage threshold value.
- Fixing whatever platform-specific test failures the matrix surfaces. Those are follow-up commits; this change only creates the signal.
- Adding a Go version matrix.

### Technical Approach

- Under `jobs.test`, replace the scalar `runs-on: ubuntu-latest` with a matrix:

  ```yaml
  jobs:
    test:
      name: Test
      runs-on: ${{ matrix.os }}
      strategy:
        fail-fast: false
        matrix:
          os: [ubuntu-latest, macos-latest, windows-latest]
  ```

- `make fmt-check`, `make vet` and `make build` require GNU `make`. It is preinstalled on all three GitHub-hosted runner images (`windows-latest` ships it via the Git-for-Windows / MSYS2 toolchain on `PATH`), so the existing steps need no `shell:` override. If `make` is not resolvable on `windows-latest` at implementation time, replace those three steps with their direct Go equivalents (`gofmt -l .`, `go vet ./...`, `go build -o build/shotgun-cli .`) rather than dropping the runner from the matrix.
- Replace the single `Run tests with coverage` step with two conditioned steps. Only the Linux one produces `coverage.out`, so the two downstream coverage steps must carry the same condition or they will fail on macOS/Windows with a missing file:

  ```yaml
  - name: Run tests with race and coverage (Linux)
    if: matrix.os == 'ubuntu-latest'
    run: go test -v -race -covermode=atomic -coverprofile=coverage.out -skip "TestWizardClipboardCopyCmd" ./...

  - name: Run tests (macOS, Windows)
    if: matrix.os != 'ubuntu-latest'
    run: go test -v -skip "TestWizardClipboardCopyCmd" ./...

  - name: Upload coverage to Codecov
    if: matrix.os == 'ubuntu-latest'
    uses: codecov/codecov-action@v4
    # ... existing `with:` block unchanged

  - name: Check coverage threshold
    if: matrix.os == 'ubuntu-latest'
    run: |
      # ... existing script unchanged
  ```

- The `Check coverage threshold` script uses `bc`, which is not available on `windows-latest`. Gating the step on `ubuntu-latest` is what keeps it working; do not relax that condition.
- Keep the existing `actions/checkout@v6` `fetch-depth: 0` on every runner — `make build` injects the version via `-ldflags` from git tags and the e2e suite asserts on it.

### Touchpoints

- `.github/workflows/test.yml` — the `test` job only: matrix block, `runs-on`, test step split, `if:` conditions on the two coverage steps.

### Contracts

```yaml
# .github/workflows/test.yml — jobs.test, resulting shape
test:
  name: Test
  runs-on: ${{ matrix.os }}
  strategy:
    fail-fast: false
    matrix:
      os: [ubuntu-latest, macos-latest, windows-latest]
  steps:
    - Checkout code            # all runners, fetch-depth: 0
    - Set up Go                # all runners, go-version 1.24.x
    - Download dependencies    # all runners
    - Check formatting         # all runners
    - Run go vet               # all runners
    - Build binary for e2e     # all runners
    - Run tests (Linux)        # if: matrix.os == 'ubuntu-latest'   -race -covermode=atomic
    - Run tests (non-Linux)    # if: matrix.os != 'ubuntu-latest'   plain
    - Upload coverage          # if: matrix.os == 'ubuntu-latest'
    - Check coverage threshold # if: matrix.os == 'ubuntu-latest'
```

### Acceptance Criteria

- [ ] `.github/workflows/test.yml` declares `strategy.matrix.os` with exactly `ubuntu-latest`, `macos-latest`, `windows-latest` on the `test` job.
- [ ] `strategy.fail-fast` is `false`.
- [ ] `jobs.test.runs-on` is `${{ matrix.os }}`.
- [ ] Exactly one test step carries `-race`, and it is guarded by `if: matrix.os == 'ubuntu-latest'`.
- [ ] Exactly one test step is guarded by `if: matrix.os != 'ubuntu-latest'` and carries neither `-race` nor `-coverprofile`.
- [ ] Both test steps pass `-skip "TestWizardClipboardCopyCmd"`.
- [ ] The `Upload coverage to Codecov` step carries `if: matrix.os == 'ubuntu-latest'`.
- [ ] The `Check coverage threshold` step carries `if: matrix.os == 'ubuntu-latest'` and its `THRESHOLD` remains `80`.
- [ ] The `lint` and `build` jobs still declare `runs-on: ubuntu-latest` with no matrix.
- [ ] A workflow run reports three `Test` job instances, one per OS.
- [ ] On the `windows-latest` run, `TestGetConfigDir_Windows` and `TestGetDefaultConfigPath_Windows` execute rather than skip.
- [ ] On the `macos-latest` run, the `runtime.GOOS == "darwin"` branch of `cmd/root_test.go` executes rather than skips.
- [ ] The `ubuntu-latest` run still fails the job when total coverage drops below 80%.

### Dependencies

- None.

---

## CI-008: Remove Dead `app.CLIConfig` / `app.ProgressMode` / `app.ProgressOutput`

### Scope

**In scope**:

- Delete `internal/app/config.go` in full — `CLIConfig`, `ProgressMode`, the three `ProgressNone`/`ProgressHuman`/`ProgressJSON` constants, and `ProgressOutput`.
- Delete `internal/app/config_test.go` in full.
- Remove the `CLIConfig` row from the `## CONFIG TYPES` table in `internal/app/AGENTS.md`.

**Out of scope**:

- `cmd/context.go:25-42` — the live `ProgressMode` / `ProgressOutput` declarations stay exactly where they are. Progress *rendering* is a presentation concern and `cmd` is its only consumer.
- Introducing any replacement type in `internal/app`.
- Touching `internal/app/context.go`, which declares the live `GenerateConfig` and is unrelated.

### Technical Approach

- Confirm no production reference survives before deleting. The pre-change state has exactly two files referencing these symbols:

  ```bash
  grep -rn "CLIConfig\|app\.ProgressMode\|app\.ProgressOutput" --include='*.go' .
  # internal/app/config.go        (declaration)
  # internal/app/config_test.go:9-12  (its only test)
  ```

- `rm internal/app/config.go internal/app/config_test.go`.
- In `internal/app/AGENTS.md`, the `## CONFIG TYPES` table currently reads:

  | Type | File | Purpose |
  |------|------|---------|
  | `CLIConfig` | `config.go` | CLI flag parsing, Viper-bound |
  | `GenerateConfig` | `service.go` | Context generation parameters |
  | `LLMSendConfig` | `service_llm.go` | LLM send parameters (provider, key, model, output) |

  Delete the `CLIConfig` row. Leave the other two rows untouched.
- Coverage is unaffected: `internal/app/config.go` contains only type and constant declarations and contributes no statements to the coverage denominator, so removing it and its test cannot move the 80% floor.

### Touchpoints

- `internal/app/config.go` — deleted.
- `internal/app/config_test.go` — deleted.
- `internal/app/AGENTS.md` — one table row removed.

### Contracts

```go
// Symbols removed from package app. No replacement is introduced.
type CLIConfig struct{ /* ... */ }
type ProgressMode string
const ( ProgressNone ProgressMode = "none"; ProgressHuman ProgressMode = "human"; ProgressJSON ProgressMode = "json" )
type ProgressOutput struct{ /* ... */ }

// Unchanged and still live in package cmd (cmd/context.go:25-42):
type ProgressMode string
const ( ProgressNone ProgressMode = "none"; ProgressHuman ProgressMode = "human"; ProgressJSON ProgressMode = "json" )
type ProgressOutput struct {
    Timestamp string  `json:"timestamp"`
    Stage     string  `json:"stage"`
    Message   string  `json:"message"`
    Current   int64   `json:"current,omitempty"`
    Total     int64   `json:"total,omitempty"`
    Percent   float64 `json:"percent,omitempty"`
}
```

### Acceptance Criteria

- [ ] `internal/app/config.go` does not exist.
- [ ] `internal/app/config_test.go` does not exist.
- [ ] `grep -rn "CLIConfig" --include='*.go' .` returns no results.
- [ ] `grep -rn "app\.ProgressMode\|app\.ProgressOutput" --include='*.go' .` returns no results.
- [ ] `cmd/context.go` still declares `ProgressMode`, `ProgressNone`, `ProgressHuman`, `ProgressJSON` and `ProgressOutput`.
- [ ] `internal/app/AGENTS.md` no longer lists a `CLIConfig` row and still lists `GenerateConfig` and `LLMSendConfig`.
- [ ] `go build ./...` exits 0.
- [ ] `go vet ./...` exits 0.
- [ ] `go test ./...` passes with no new failures.
- [ ] Total coverage reported by `make coverage` remains at or above 80%.

### Dependencies

- None.
