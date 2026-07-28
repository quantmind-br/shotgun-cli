# PRD — Config and Output Single-Sourcing

**Source**: IDEATION_CODE_IMPROVEMENTS.md
**Generated**: 2026-07-28

## Implementation Order

1. CI-003 — Drive `ValidateValue` / `ConvertValue` from `ConfigMetadata.Type`, enforcing declared `MinValue` / `MaxValue`
2. CI-004 — Extract `app.DefaultOutputName()` and standardize the generated-context file mode to `0600`

---

## CI-003: Derive Validation From the Metadata Registry

### Scope

**In scope**:

- Replace `ValidateValue`'s per-key `switch` (`internal/config/validator.go:54-82`) with a dispatch on `GetMetadata(key).Type`.
- Replace `ConvertValue`'s per-key `switch` (`internal/config/validator.go:85-104`) with the same dispatch.
- Enforce `meta.MinValue` and `meta.MaxValue` for `TypeInt` and `TypeTimeout`.
- Replace `validateOutputFormat` and `validateLLMProvider` with one generic enum check reading `meta.EnumOptions`.
- Add a metadata-driven table test asserting every key in `AllConfigMetadata()` round-trips its own `DefaultValue` through `ValidateValue` without error.
- Add table cases for the newly enforced `scanner.max-files` upper bound.
- Update the two affected error-message rows in the README validation table.

**Out of scope**:

- `ValidKeys`, `IsValidKey`, `DeprecationMessage` and `deprecatedKeys` — all already correct.
- `validateSizeFormat`, `validatePath`, `validateURL`, `validateBooleanValue` — kept verbatim as the per-type implementations.
- Adding new config keys, changing any `DefaultValue`, or changing any `MinValue` / `MaxValue`.
- Importing anything outside stdlib + `internal/utils` into `internal/config`. The package is stdlib-only by layer rule; enum lists come from in-package `EnumOptions`, never from `core/llm`.
- Enforcing `Required`. That field is declared but unread, and wiring it is a separate decision.

### Technical Approach

- `internal/config/validator_test.go` asserts only `wantErr bool` — it never asserts error message text. Error wording is therefore free to change; the constraint is the error/no-error boundary, not the string.
- New `ValidateValue`:

  ```go
  func ValidateValue(key, value string) error {
      meta, ok := GetMetadata(key)
      if !ok {
          // Unknown key: IsValidKey is the gate, and the previous switch fell
          // through to nil here. Preserve that.
          return nil
      }

      switch meta.Type {
      case TypeBool:
          return validateBooleanValue(value)
      case TypeSize:
          return validateSizeFormat(value)
      case TypeEnum:
          return validateEnumValue(value, meta.EnumOptions)
      case TypeInt:
          return validateIntValue(value, meta)
      case TypeTimeout:
          return validateTimeoutValue(value, meta)
      case TypePath:
          return validatePath(value)
      case TypeURL:
          return validateURL(value)
      case TypeString:
          return nil
      }

      return nil
  }
  ```

- New `ConvertValue`:

  ```go
  func ConvertValue(key, value string) (interface{}, error) {
      meta, ok := GetMetadata(key)
      if !ok {
          return value, nil
      }

      switch meta.Type {
      case TypeInt, TypeTimeout:
          var intVal int
          if _, err := fmt.Sscanf(value, "%d", &intVal); err != nil {
              return nil, fmt.Errorf("failed to parse integer value: %w", err)
          }
          return intVal, nil
      case TypeBool:
          return strings.ToLower(value) == "true", nil
      default:
          return value, nil
      }
  }
  ```

  This is behaviour-identical to the current implementation: `KeyScannerMaxFiles` is the only `TypeInt`, `KeyLLMTimeout` the only `TypeTimeout`, and the eleven boolean keys are exactly the `TypeBool` set. Verify that correspondence against `buildAllMetadata()` before deleting the old arms.
- `validateIntValue` — keeps the size-format pre-check, which carries a distinct message and is the reason `TypeInt` does not simply reuse a generic parser:

  ```go
  func validateIntValue(value string, meta ConfigMetadata) error {
      if isSizeFormat(value) {
          return fmt.Errorf("expected a number, got size format")
      }

      var n int
      if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
          return fmt.Errorf("expected a positive integer")
      }
      if meta.MinValue > 0 && n < meta.MinValue {
          if meta.MinValue == 1 {
              return fmt.Errorf("must be positive, got %d", n)
          }
          return fmt.Errorf("must be at least %d, got %d", meta.MinValue, n)
      }
      if meta.MaxValue > 0 && n > meta.MaxValue {
          return fmt.Errorf("too large (max %d), got %d", meta.MaxValue, n)
      }

      return nil
  }
  ```

  Extract the size-format detection from the current `validateMaxFiles` body (`internal/config/validator.go:109-117`) into `isSizeFormat(value string) bool` verbatim. Delete `validateMaxFiles`.
- `validateTimeoutValue` — same shape, distinct wording, driven by metadata bounds:

  ```go
  func validateTimeoutValue(value string, meta ConfigMetadata) error {
      var timeout int
      if _, err := fmt.Sscanf(value, "%d", &timeout); err != nil {
          return fmt.Errorf("expected a positive integer (seconds)")
      }
      if meta.MinValue > 0 && timeout < meta.MinValue {
          return fmt.Errorf("timeout must be positive, got %d", timeout)
      }
      if meta.MaxValue > 0 && timeout > meta.MaxValue {
          return fmt.Errorf("timeout too large (max %d seconds), got %d", meta.MaxValue, timeout)
      }

      return nil
  }
  ```

  With `MaxValue: 3600` this renders byte-identically to the current message. Delete `validateTimeout` and the hardcoded `3600` at `internal/config/validator.go:189`.
- `validateEnumValue` — one wording for both enum keys:

  ```go
  func validateEnumValue(value string, options []string) error {
      for _, opt := range options {
          if value == opt {
              return nil
          }
      }

      return fmt.Errorf("expected one of: %s, got '%s'", strings.Join(options, ", "), value)
  }
  ```

  Delete `validateOutputFormat` and `validateLLMProvider`. The `llm.provider` message gains a `, got 'x'` suffix; the `output.format` message changes from `expected 'markdown' or 'text'` to `expected one of: markdown, text`. Both are message-only changes, and no test asserts on either string.
- Add the drift guard to `internal/config/validator_test.go`:

  ```go
  func TestValidateValue_AllDefaultsRoundTrip(t *testing.T) {
      t.Parallel()
      for _, m := range AllConfigMetadata() {
          m := m
          t.Run(m.Key, func(t *testing.T) {
              t.Parallel()
              if err := ValidateValue(m.Key, fmt.Sprint(m.DefaultValue)); err != nil {
                  t.Errorf("default for %s (%v) fails its own validator: %v", m.Key, m.DefaultValue, err)
              }
          })
      }
  }
  ```

- Extend `TestValidateValue_MaxFiles`'s table with `{"1000000", false}` and `{"1000001", true}`. The existing seven cases (`100`, `10000`, `0`, `-1`, `10MB`, `1KB`, `abc`) all keep their current `wantErr` values — the highest existing case is `10000`, well under the new ceiling.
- Update `README.md`'s "Validation Functions" table: the `output.format` row's error message becomes `"expected one of: markdown, text"` and the `llm.provider` row's becomes `"expected one of: openai, anthropic, gemini, got '<value>'"`. Change the `scanner.max-files` row's Rules cell to note the 1–1,000,000 range.

### Touchpoints

- `internal/config/validator.go` — `ValidateValue` and `ConvertValue` rewritten as type dispatch; `isSizeFormat`, `validateIntValue`, `validateTimeoutValue`, `validateEnumValue` added; `validateMaxFiles`, `validateTimeout`, `validateOutputFormat`, `validateLLMProvider` deleted.
- `internal/config/validator_test.go` — `TestValidateValue_AllDefaultsRoundTrip` added; `TestValidateValue_MaxFiles` table extended.
- `README.md` — three rows in the validation table corrected.

### Contracts

```go
// internal/config/validator.go — public surface unchanged
func ValidateValue(key, value string) error
func ConvertValue(key, value string) (interface{}, error)

// New unexported, metadata-driven validators
func isSizeFormat(value string) bool
func validateIntValue(value string, meta ConfigMetadata) error
func validateTimeoutValue(value string, meta ConfigMetadata) error
func validateEnumValue(value string, options []string) error

// Deleted
func validateMaxFiles(value string) error
func validateTimeout(value string) error
func validateOutputFormat(value string) error
func validateLLMProvider(value string) error

// internal/config/metadata.go — read, not modified
func GetMetadata(key string) (ConfigMetadata, bool)
func AllConfigMetadata() []ConfigMetadata
```

### Acceptance Criteria

- [ ] `ValidateValue` contains no `case Key...` arm; its only `switch` is over `meta.Type`.
- [ ] `ConvertValue` contains no `case Key...` arm; its only `switch` is over `meta.Type`.
- [ ] `grep -n "validateMaxFiles\|validateTimeout\b\|validateOutputFormat\|validateLLMProvider" internal/config/` returns no results.
- [ ] `grep -n "3600" internal/config/validator.go` returns no results.
- [ ] `grep -n "ProviderOpenAI\|FormatMarkdown" internal/config/validator.go` returns no results.
- [ ] `ValidateValue(KeyScannerMaxFiles, "1000000")` returns nil.
- [ ] `ValidateValue(KeyScannerMaxFiles, "1000001")` returns a non-nil error.
- [ ] `ValidateValue(KeyScannerMaxFiles, "0")`, `"-1"`, `"10MB"`, `"1KB"`, `"abc"` each still return a non-nil error.
- [ ] `ValidateValue(KeyLLMTimeout, "3600")` returns nil and `ValidateValue(KeyLLMTimeout, "3601")` returns an error whose message contains `"timeout too large (max 3600 seconds)"`.
- [ ] `ValidateValue("unknown.key", "anything")` returns nil.
- [ ] `ConvertValue(KeyScannerMaxFiles, "5000")` returns `int(5000)`; `ConvertValue(KeyLLMTimeout, "300")` returns `int(300)`; `ConvertValue(KeyOutputClipboard, "TRUE")` returns `true`; `ConvertValue(KeyLLMModel, "gpt-4o")` returns the string.
- [ ] `TestValidateValue_AllDefaultsRoundTrip` exists and passes for all 21 registered keys.
- [ ] Adding a temporary `TypeEnum` key to `buildAllMetadata()` with a valid `EnumOptions` list makes it validate correctly with no edit to `validator.go` (verify manually, then revert).
- [ ] `shotgun-cli config set scanner.max-files 999999999` exits non-zero.
- [ ] `shotgun-cli config set scanner.max-files 5000` still succeeds.
- [ ] Every pre-existing test in `internal/config/validator_test.go` passes unchanged apart from the extended max-files table.
- [ ] `go test -race ./internal/config/... ./cmd/...` passes.
- [ ] `golangci-lint run --config .golangci.yml ./...` reports no new findings.

### Dependencies

- None.

---

## CI-004: Single-Source the Output Filename and File Mode

### Scope

**In scope**:

- Add `app.DefaultOutputName()` as the single producer of the `shotgun-prompt-<timestamp>.md` name; have `GenerateConfig.GenerateOutputPath()` call it.
- Delete the duplicate filename literal in `cmd/context.go:213-217`, leaving `Output` empty so the service default applies, and read the final path from the service's `GenerateResult.OutputPath`.
- Delete the duplicate filename literal in `internal/ui/wizard.go:1264-1265` in favour of `app.DefaultOutputName()`.
- Change the TUI write at `internal/ui/wizard.go:1269` from `0644` to `0600` and delete the now-inaccurate `#nosec G306` comment above it.
- Update any `internal/ui/wizard_test.go` assertion that pins the generated path or file mode.

**Out of scope**:

- **The output directory of either front end.** `internal/ui/wizard.go:1266` does `filepath.Join(m.rootPath, filename)` while `cmd/context.go` and `internal/app/service.go:198` resolve a bare filename against the process CWD. Unifying the directory would silently relocate every TUI user's artifact. Both directory behaviours stay exactly as they are.
- The `llm-response-<timestamp>.md` name in `cmd/send.go:86-87` — a different artifact with a different convention.
- The `0600` writes at `internal/app/service.go:199` and `cmd/send.go:131` — already correct, unchanged.
- Removing the global `G306` exclusion from `.golangci.yml`.
- The timestamp format itself (`20060102-150405`), which stays as-is.

### Technical Approach

- In `internal/app/context.go`, promote the body of `GenerateOutputPath` to a package-level function:

  ```go
  // DefaultOutputName returns the default filename for a generated context
  // document. It is the single producer of this convention; both front ends
  // call it rather than formatting the name themselves.
  func DefaultOutputName() string {
      return fmt.Sprintf("shotgun-prompt-%s.md", time.Now().Format("20060102-150405"))
  }

  func (c *GenerateConfig) GenerateOutputPath() string {
      if c.OutputPath != "" {
          return c.OutputPath
      }

      return DefaultOutputName()
  }
  ```

- In `cmd/context.go`, delete the `if output == "" { ... }` block at lines 213-217. `Output` then flows through empty, `app.GenerateConfig.OutputPath` is empty, and `GenerateOutputPath()` supplies the default inside the service.
  - Audit every later use of the local `output` and of the command-level `GenerateConfig.Output` in `cmd/context.go`. Any site that echoes the destination to the user must switch to `result.OutputPath` from `GenerateResult` (`internal/app/service.go:214`), because the command no longer knows the name it did not generate. If no such echo exists, no further change is needed.
- In `internal/ui/wizard.go`:

  ```go
  func (m *WizardModel) saveGeneratedContent(content string) (string, error) {
      filePath := filepath.Join(m.rootPath, app.DefaultOutputName())
      if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
          return "", fmt.Errorf("failed to write file: %w", err)
      }

      return filePath, nil
  }
  ```

  `internal/ui` already imports `internal/app` (see `internal/ui/generate_coordinator.go`), so no new dependency direction is introduced. The `time` import in `wizard.go` stays — it is used elsewhere in the file; confirm before removing anything.
- Mode decision, recorded here so it is not relitigated: **`0600`**. Generated context documents embed the user's private source tree; the deleted `#nosec G306` justification ("Generated context files are meant to be world-readable") is not a defensible reason to widen permissions on that content. Both headless write sites already use `0600`, so this change affects only the TUI. State the change explicitly in the commit message, since it is user-visible for anyone whose workflow reads the artifact as a different user.
- `.golangci.yml:32` excludes `G306` globally, so gosec will not flag the mode either way; the change is not linter-driven.

### Touchpoints

- `internal/app/context.go` — `DefaultOutputName()` added; `GenerateOutputPath` delegates to it.
- `cmd/context.go` — the `if output == ""` default-filename block removed; any user-facing echo of the path repointed at `result.OutputPath`.
- `internal/ui/wizard.go` — `saveGeneratedContent` uses `app.DefaultOutputName()`; mode `0644` → `0600`; `#nosec G306` comment deleted.
- `internal/ui/wizard_test.go` — assertions on the generated filename and/or mode updated.
- `internal/app/context_test.go` — `TestGenerateConfig_GenerateOutputPath_*` must keep passing unchanged; add a direct `TestDefaultOutputName` if coverage of the new symbol is needed.

### Contracts

```go
// internal/app/context.go — new exported single producer
func DefaultOutputName() string  // "shotgun-prompt-20060102-150405.md"

// Unchanged signature, now delegating
func (c *GenerateConfig) GenerateOutputPath() string

// internal/ui/wizard.go — unchanged signature, new body
func (m *WizardModel) saveGeneratedContent(content string) (string, error)

// Write mode for generated context documents, at all three sites:
//   internal/app/service.go:199   0600 (unchanged)
//   cmd/send.go:131               0600 (unchanged)
//   internal/ui/wizard.go         0600 (was 0644)
```

### Acceptance Criteria

- [ ] `grep -rn "shotgun-prompt-%s.md\|shotgun-prompt-" --include='*.go' internal/ cmd/` returns exactly one non-test production hit: the body of `app.DefaultOutputName`.
- [ ] `cmd/context.go` contains no `time.Now().Format("20060102-150405")` call.
- [ ] `internal/ui/wizard.go` contains no `time.Now().Format("20060102-150405")` call inside `saveGeneratedContent`.
- [ ] `grep -n "0644" internal/ui/wizard.go` returns no results.
- [ ] `grep -n "nosec G306" internal/ui/wizard.go` returns no results.
- [ ] `shotgun-cli context generate -r .` writes a file named `shotgun-prompt-<YYYYMMDD>-<HHMMSS>.md` with mode `0600`, and any path it prints matches the file it actually wrote.
- [ ] `shotgun-cli context generate -r . -o custom.md` still writes exactly `custom.md`.
- [ ] A TUI run through to generation writes its artifact into the scanned root directory (unchanged behaviour) with mode `0600`.
- [ ] `TestGenerateConfig_GenerateOutputPath_CustomPath`, `_AutoGenerated` and `_TimestampFormat` in `internal/app/context_test.go` pass unchanged.
- [ ] `test/e2e/cli_test.go` `TestCLIContextGenerateProducesFile` passes.
- [ ] `go test -race ./internal/... ./cmd/...` passes.
- [ ] `golangci-lint run --config .golangci.yml ./...` reports no new findings.

### Dependencies

- None.
