# PRD — LLM Send Consolidation

**Source**: IDEATION_CODE_IMPROVEMENTS.md
**Generated**: 2026-07-28

## Implementation Order

1. CI-001 — Route `context send` through `app.ContextService`, delete `cmd/providers.go`, repoint `cmd/llm.go`
2. CI-005 — Build each platform client's defaults from `llm.DefaultConfigs()`; replace the manual default block in `cmd/config_llm.go` with `cfg.WithDefaults()`
3. CI-006 — Rewrite `TestRunContextSend_FromFile` against an injected fake registry

---

## CI-001: Route `context send` Through `ContextService`

### Scope

**In scope**:

- Delete `cmd/providers.go` entirely, removing the package-level `providerRegistry`, `CreateLLMProvider` and `GetProviderRegistry`.
- Repoint `cmd/llm.go:85` (`runLLMStatus`) and `cmd/llm.go:151` (`runLLMDoctor`) at `app.DefaultProviderRegistry.Create(cfg)`.
- Replace the provider-creation, availability-check, validation and send block in `runContextSend` (`cmd/send.go:94-137`) with a single `app.ContextService.SendToLLMWithProgress` call.
- Introduce a package-level `newSendService` function variable in `cmd/send.go` as the injection seam CI-006 needs.
- Delete or repoint the eight now-orphaned test functions in `cmd/llm_test.go`: `TestCreateLLMProvider_OpenAI`, `TestCreateLLMProvider_Anthropic`, `TestCreateLLMProvider_Gemini`, `TestCreateLLMProvider_InvalidProvider`, `TestGetProviderRegistry`, `TestGetProviderRegistry_AllProvidersPresent`, `TestGetProviderRegistry_CreatesProvider`, `TestGetProviderRegistry_Singleton`.
- Replace the Gemini-only help text in `cmd/send.go`: the command `Short`, the `Long` body, and the `--output` / `--model` flag descriptions.

**Out of scope**:

- `internal/app/providers.go` — `DefaultProviderRegistry` and its three registrations are already correct and stay unchanged.
- Changing `ContextService.SendToLLMWithProgress`'s signature or behaviour.
- Rewriting `TestRunContextSend_FromFile` — that is CI-006, which consumes the seam this item creates.
- Progress rendering for headless send: pass `nil` as the progress callback, matching current behaviour (`cmd/send.go` prints one `Sending to %s (%s)...` line).

### Technical Approach

- Delete `cmd/providers.go`. Both `cmd/llm.go` call sites become:

  ```go
  provider, err := app.DefaultProviderRegistry.Create(cfg)
  ```

  Note the real signature is `Create(cfg llm.Config)` — single argument. `cmd/llm.go` must gain the `internal/app` import; the `internal/core/llm` import stays (it is still used by `displayURL` and `llm.ProviderType`).

  `CreateLLMProvider` wrapped the registry error as `"failed to create provider: %w"`. `runLLMStatus` renders that string into `Status: Not ready - <err>`, so preserve the wrap at both call sites to keep the rendered output identical:

  ```go
  provider, err := app.DefaultProviderRegistry.Create(cfg)
  if err != nil {
      err = fmt.Errorf("failed to create provider: %w", err)
      // ... existing handling
  }
  ```

- Add the seam at package scope in `cmd/send.go`:

  ```go
  // newSendService builds the service used by `context send`. It is a variable
  // so tests can substitute a registry-injected service.
  var newSendService = func() app.ContextService {
      return app.NewContextService()
  }
  ```

- Replace `cmd/send.go:90-137` with the service call. `SendToLLMWithProgress` saves `result.Response` when `SaveResponse` is true and `OutputPath` is non-empty (`internal/app/service.go:283-287`), so `--raw` — which must write `result.RawResponse` — keeps its write in the command:

  ```go
  cfg := BuildLLMConfigWithOverrides(model, timeout)

  sendCfg := app.LLMSendConfig{
      Provider:     cfg.Provider,
      APIKey:       cfg.APIKey,
      BaseURL:      cfg.BaseURL,
      Model:        cfg.Model,
      Timeout:      cfg.Timeout,
      SaveResponse: outputFile != "" && !raw,
      OutputPath:   outputFile,
  }

  fmt.Printf("Sending to %s (%s)...\n", cfg.Provider, cfg.Model)

  result, err := newSendService().SendToLLMWithProgress(context.Background(), content, sendCfg, nil)
  if err != nil {
      return err
  }

  response := result.Response
  if raw {
      response = result.RawResponse
  }

  if outputFile != "" {
      if raw {
          if writeErr := os.WriteFile(outputFile, []byte(response), 0600); writeErr != nil {
              return fmt.Errorf("failed to save response to '%s': %w", outputFile, writeErr)
          }
      }
      fmt.Printf("Response saved to: %s\n", outputFile)
  } else {
      fmt.Println(response)
  }
  ```

- The pre-send `Sending to %s (%s)...` line previously printed `llmProvider.Name()` (the client's display name, e.g. `"OpenAI"`). After the change the provider instance is owned by the service, so print `cfg.Provider` (the configured key, e.g. `"openai"`). This is a deliberate, cosmetic-only change; note it in the commit message.
- Error wrapping shifts to the service's wording. `SendToLLMWithProgress` returns `"failed to create LLM provider: %w"`, `"%s not available"`, `"invalid provider config: %w"` and `"LLM request failed: %w"`. Return those unwrapped — re-wrapping would double the prefix. The old `"Run 'shotgun-cli llm doctor' for help"` hints are lost from the send path; do not reintroduce them here, since `llm doctor` remains the documented diagnostic entry point.
- Keep the trailing token-usage and duration output block (`cmd/send.go:139-146`) verbatim.
- Keep `formatDuration` in `cmd/send.go`; `TestFormatDuration` in `cmd/send_test.go` stays green untouched.
- For `cmd/llm_test.go`: `TestCreateLLMProvider_*` becomes `TestDefaultProviderRegistry_Create*` asserting the same four cases against `app.DefaultProviderRegistry`, or is deleted if `internal/app` already covers registry creation. `TestGetProviderRegistry_Singleton` has no meaning after the deletion — remove it rather than porting it.

### Touchpoints

- `cmd/providers.go` — deleted.
- `cmd/send.go` — `newSendService` added; `runContextSend` body from config build through response write replaced; command `Short`/`Long` and two flag descriptions de-Geminified.
- `cmd/llm.go` — two `CreateLLMProvider` call sites repointed at `app.DefaultProviderRegistry.Create`; `internal/app` import added.
- `cmd/llm_test.go` — eight test functions deleted or repointed.
- `cmd/send_test.go` — compiles unchanged; `TestRunContextSend_FromFile` is rewritten in CI-006.

### Contracts

```go
// cmd/send.go — new package-level injection seam
var newSendService = func() app.ContextService { return app.NewContextService() }

// internal/app — consumed as-is, unchanged
type LLMSendConfig struct {
    Provider     llm.ProviderType
    APIKey       string
    BaseURL      string
    Model        string
    Timeout      int
    SaveResponse bool
    OutputPath   string
}

func (s *DefaultContextService) SendToLLMWithProgress(
    ctx context.Context,
    content string,
    cfg LLMSendConfig,
    progress LLMProgressCallback,
) (*llm.Result, error)

// internal/core/llm — real registry signature (note: single argument)
func (r *Registry) Create(cfg Config) (Provider, error)

// Removed from package cmd. No replacement.
func CreateLLMProvider(cfg llm.Config) (llm.Provider, error)
func GetProviderRegistry() *llm.Registry
```

### Acceptance Criteria

- [ ] `cmd/providers.go` does not exist.
- [ ] `grep -rn "CreateLLMProvider\|GetProviderRegistry\|providerRegistry" --include='*.go' .` returns no results.
- [ ] `grep -rn "llm.NewRegistry()" --include='*.go' internal/ cmd/` returns exactly one production hit: `internal/app/providers.go:10`.
- [ ] `cmd/send.go` declares `var newSendService` at package scope returning `app.ContextService`.
- [ ] `runContextSend` contains no direct call to `provider.Send`, `provider.IsAvailable` or `provider.ValidateConfig`.
- [ ] `shotgun-cli context send prompt.md -o out.md` with a configured provider writes `out.md` and prints `Response saved to: out.md`.
- [ ] `shotgun-cli context send prompt.md -o out.md --raw` writes `result.RawResponse` to `out.md`, not `result.Response`.
- [ ] `shotgun-cli context send prompt.md` with no `--output` and `llm.save-response=false` prints the response to stdout and writes no file.
- [ ] `shotgun-cli context send prompt.md` with no `--output` and `llm.save-response=true` writes `llm-response-<timestamp>.md`.
- [ ] Both files written by the send path have mode `0600`.
- [ ] `shotgun-cli llm status` and `shotgun-cli llm doctor` produce byte-identical output to the pre-change binary for a configured provider and for an unset provider.
- [ ] `shotgun-cli context send --help` contains no occurrence of "Gemini" or "gemini" outside the usage examples.
- [ ] `go build ./...` exits 0 and `golangci-lint run --config .golangci.yml ./...` reports no new findings.
- [ ] `go test -race ./cmd/... ./internal/app/...` passes.

### Dependencies

- None.

---

## CI-005: Single-Source Provider Defaults

### Scope

**In scope**:

- Build each platform client's `llmbase.DefaultConfig` from `llm.DefaultConfigs()[llm.Provider<Name>]` in `openai.NewClient`, `anthropic.NewClient` and `geminiapi.NewClient`.
- Delete the now-unused `defaultBaseURL` constant from each of the three client packages.
- Replace the manual three-field default block in `BuildLLMConfig` (`cmd/config_llm.go:14-34`) with `cfg.WithDefaults()`.

**Out of scope**:

- `defaultMaxTokens` in `anthropic` and `geminiapi` — `MaxTokens` is absent from `llm.DefaultConfigs()` and stays a package constant.
- `anthropicVersion` in `anthropic`.
- Adding `MaxTokens` to `llm.Config`'s default table.
- Changing any default *value*. This item changes only where each value is read from.
- `llmbase.NewBaseClientWithDefaults` — its signature and fallback logic (including the hardcoded `300 * time.Second` last resort at `internal/platform/llmbase/base_client.go:73-75`) stay as-is.

### Technical Approach

- `llm.Config.Timeout` is an `int` in seconds; `llmbase.DefaultConfig.Timeout` is a `time.Duration`. Every client must convert.
- `internal/platform/openai/client.go`:

  ```go
  func NewClient(cfg llm.Config) (*Client, error) {
      d := llm.DefaultConfigs()[llm.ProviderOpenAI]
      base, err := llmbase.NewBaseClientWithDefaults(cfg, llmbase.DefaultConfig{
          BaseURL: d.BaseURL,
          Model:   d.Model,
          Timeout: time.Duration(d.Timeout) * time.Second,
      }, "OpenAI")
      if err != nil {
          return nil, err
      }
      return &Client{BaseClient: base}, nil
  }
  ```

  Delete `const defaultBaseURL = "https://api.openai.com/v1"` (line 13).
- `internal/platform/anthropic/client.go` — same shape with `llm.ProviderAnthropic` and `"Anthropic"`, retaining `MaxTokens: defaultMaxTokens`. Delete `defaultBaseURL` from the `const` block; keep `anthropicVersion` and `defaultMaxTokens`.
- `internal/platform/geminiapi/client.go` — same shape with `llm.ProviderGemini` and `"Gemini"`, retaining `MaxTokens: defaultMaxTokens`. Delete `defaultBaseURL`; keep `defaultMaxTokens`.
- Before deleting each `defaultBaseURL`, grep its package — if a test or another file in the same package references it, keep the constant and have `NewClient` still read from `llm.DefaultConfigs()`, so the table remains the single producer.
- `cmd/config_llm.go`:

  ```go
  func BuildLLMConfig() llm.Config {
      cfg := llm.Config{
          Provider: llm.ProviderType(viper.GetString(config.KeyLLMProvider)),
          APIKey:   viper.GetString(config.KeyLLMAPIKey),
          BaseURL:  viper.GetString(config.KeyLLMBaseURL),
          Model:    viper.GetString(config.KeyLLMModel),
          Timeout:  viper.GetInt(config.KeyLLMTimeout),
      }
      cfg.WithDefaults()
      return cfg
  }
  ```

  `WithDefaults` mutates the receiver in place and also returns `*Config` (`internal/core/llm/config.go:90-104`); calling it for its side effect and returning `cfg` is correct. The `defaults := llm.DefaultConfigs()[provider]` local becomes unused and must be removed or `golangci-lint` fails the build.
- `BuildLLMConfigWithOverrides` is unchanged — it already delegates to `BuildLLMConfig`.

### Touchpoints

- `internal/platform/openai/client.go` — `NewClient` defaults sourced from `llm.DefaultConfigs()`; `defaultBaseURL` removed.
- `internal/platform/anthropic/client.go` — same; `defaultBaseURL` removed from the `const` block.
- `internal/platform/geminiapi/client.go` — same; `defaultBaseURL` removed from the `const` block.
- `cmd/config_llm.go` — `BuildLLMConfig` manual default block replaced by `cfg.WithDefaults()`.

### Contracts

```go
// internal/core/llm/config.go — the single producer, unchanged
func DefaultConfigs() map[ProviderType]Config
func (c *Config) WithDefaults() *Config

// internal/platform/llmbase — unchanged consumer signature
type DefaultConfig struct {
    BaseURL   string
    Model     string
    MaxTokens int
    Timeout   time.Duration
}
func NewBaseClientWithDefaults(cfg llm.Config, defaults DefaultConfig, providerName string) (*BaseClient, error)

// Constants removed from the three platform packages:
//   openai:    defaultBaseURL
//   anthropic: defaultBaseURL   (anthropicVersion, defaultMaxTokens retained)
//   geminiapi: defaultBaseURL   (defaultMaxTokens retained)
```

### Acceptance Criteria

- [ ] `grep -rn "defaultBaseURL" internal/platform/` returns no results.
- [ ] `grep -rn "api.openai.com\|api.anthropic.com\|generativelanguage.googleapis.com" --include='*.go' internal/ cmd/` returns hits only in `internal/core/llm/config.go` and test files.
- [ ] `grep -rn '"gpt-4o"\|"claude-sonnet-4-20250514"\|"gemini-2.5-flash"' --include='*.go' internal/platform/ cmd/` returns no non-test results.
- [ ] Each of `openai.NewClient`, `anthropic.NewClient`, `geminiapi.NewClient` reads `llm.DefaultConfigs()` exactly once and converts `Timeout` with `time.Duration(d.Timeout) * time.Second`.
- [ ] `anthropic.NewClient` and `geminiapi.NewClient` still pass `MaxTokens: defaultMaxTokens`.
- [ ] `cmd/config_llm.go` contains no `if cfg.BaseURL == ""`, `if cfg.Model == ""` or `if cfg.Timeout == 0` block, and calls `cfg.WithDefaults()`.
- [ ] Existing `client_test.go` assertions on constructed base URL, model and timeout pass unchanged in all three platform packages.
- [ ] `TestDisplayURL` in `cmd/llm_test.go` passes unchanged.
- [ ] `shotgun-cli llm status` with an empty `llm.base-url` reports the same base URL for each provider as the pre-change binary.
- [ ] `go build ./...` exits 0 and `golangci-lint run --config .golangci.yml ./...` reports no unused-variable findings.

### Dependencies

- CI-001 — leaves `cmd/config_llm.go` with a single consumer path to change.

---

## CI-006: Registry-Injected `context send` Test

### Scope

**In scope**:

- Rewrite `TestRunContextSend_FromFile` in `cmd/send_test.go` to run against a fake provider injected through `newSendService`, asserting the real output contract.
- Add a `mockLLMProvider` and a `newMockRegistry` helper to `cmd/send_test.go`, mirroring `internal/app/service_llm_test.go:16-47`.
- Add a failure case asserting the exact error surfaced when the provider reports unavailable.
- Delete `isExpectedProviderError` (`cmd/send_test.go:15-30`).
- Remove `viper.Reset()` from the rewritten test; use `t.Setenv`-free explicit `viper.Set` plus a `t.Cleanup` that restores the prior values.

**Out of scope**:

- `TestContextSendCmd_PreRunE` and `TestFormatDuration` — both are real assertions and stay untouched.
- Adding an e2e test for `context send`.
- Testing the LLM providers themselves; `internal/platform/*/client_test.go` already covers them with `httptest`.

### Technical Approach

- Define the fake in `cmd/send_test.go`:

  ```go
  type mockLLMProvider struct {
      available bool
      result    *llm.Result
      sendErr   error
  }

  func (m *mockLLMProvider) Name() string          { return "mock" }
  func (m *mockLLMProvider) IsAvailable() bool     { return m.available }
  func (m *mockLLMProvider) IsConfigured() bool    { return true }
  func (m *mockLLMProvider) ValidateConfig() error { return nil }
  func (m *mockLLMProvider) Send(ctx context.Context, content string) (*llm.Result, error) {
      return m.result, m.sendErr
  }
  func (m *mockLLMProvider) SendWithProgress(
      ctx context.Context, content string, progress func(stage string),
  ) (*llm.Result, error) {
      return m.result, m.sendErr
  }

  func newMockRegistry(p llm.Provider) *llm.Registry {
      r := llm.NewRegistry()
      r.Register(llm.ProviderOpenAI, func(cfg llm.Config) (llm.Provider, error) { return p, nil })
      return r
  }
  ```

  Match the exact `llm.Provider` method set against `internal/core/llm/` before writing; the compiler is the authority.
- Install and restore the seam per subtest:

  ```go
  prev := newSendService
  t.Cleanup(func() { newSendService = prev })
  newSendService = func() app.ContextService {
      return app.NewContextService(app.WithRegistry(newMockRegistry(mock)))
  }
  ```

- Build the `*cobra.Command` with the same four flags `contextSendCmd` registers (`output`, `model`, `timeout`, `raw`), as the existing test already does.
- Required subtests:
  1. **writes to `--output`** — `result.Response = "processed"`, `--output` set to a `t.TempDir()` path; assert the file exists, its content is `"processed"`, and its mode is `0600`.
  2. **`--raw` selects `RawResponse`** — `result.Response = "processed"`, `result.RawResponse = "{\"raw\":true}"`, `--raw` set; assert the written file contains the raw payload and not `"processed"`.
  3. **auto-generated filename** — no `--output`, `llm.save-response=true`; assert a file matching `llm-response-\d{8}-\d{6}\.md` was created in the working directory, and clean it up.
  4. **stdout when nothing to save** — no `--output`, `llm.save-response=false`; capture stdout with the existing `cmd/*_test.go` capture pattern and assert the response body appears; assert no file was created.
  5. **provider unavailable** — `available: false`; assert `runContextSend` returns an error whose message contains `"not available"`, and that it is not nil. This subtest must fail if the error is nil.
  6. **send failure** — `sendErr` set; assert the returned error wraps it and contains `"LLM request failed"`.
- Subtest 3 writes into the process working directory. Use `t.Chdir(t.TempDir())` (Go 1.24) so the artifact is isolated and swept automatically.
- No subtest may reach the network. Assert this structurally: the injected registry returns the fake for `llm.ProviderOpenAI`, and `viper` is set to `llm.provider=openai`, so `Registry.Create` never constructs a real client.

### Touchpoints

- `cmd/send_test.go` — `isExpectedProviderError` deleted; `TestRunContextSend_FromFile` rewritten; `mockLLMProvider` and `newMockRegistry` added.
- `cmd/send.go` — no further change beyond the `newSendService` seam added in CI-001.

### Contracts

```go
// cmd/send_test.go — new test doubles
type mockLLMProvider struct {
    available bool
    result    *llm.Result
    sendErr   error
}
func newMockRegistry(p llm.Provider) *llm.Registry

// Seam override pattern used by every subtest
prev := newSendService
t.Cleanup(func() { newSendService = prev })
newSendService = func() app.ContextService {
    return app.NewContextService(app.WithRegistry(newMockRegistry(mock)))
}

// Deleted
func isExpectedProviderError(err error) bool
```

### Acceptance Criteria

- [ ] `grep -n "isExpectedProviderError" cmd/send_test.go` returns no results.
- [ ] `grep -n "viper.Reset()" cmd/send_test.go` returns no results.
- [ ] `TestRunContextSend_FromFile` contains no `assert.True` over a `strings.Contains` disjunction.
- [ ] No subtest accepts `err == nil` and `err != nil` as equally passing outcomes.
- [ ] `go test -run TestRunContextSend ./cmd/...` passes with the network unavailable (verify by running it with outbound DNS blocked, or in an offline container).
- [ ] Mutating `cmd/send.go` to write `result.Response` instead of `result.RawResponse` under `--raw` makes the `--raw` subtest fail.
- [ ] Mutating `cmd/send.go` to swallow the unavailable-provider error makes the unavailable subtest fail.
- [ ] The auto-generated-filename subtest asserts a name matching `llm-response-\d{8}-\d{6}\.md`.
- [ ] The written-file subtests assert mode `0600`.
- [ ] `go test -race ./cmd/...` passes and leaves no stray `llm-response-*.md` in the repository working directory.
- [ ] Total coverage reported by `make coverage` remains at or above 80%.

### Dependencies

- CI-001 — supplies the `newSendService` injection seam.
