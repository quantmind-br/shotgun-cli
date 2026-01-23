# Code Quality Improvements

This document tracks code quality improvements for shotgun-cli, prioritized by implementation order.

## Implementation Priority

| Priority | ID | Title | Effort | Status |
|----------|-----|-------|--------|--------|
| 1 | cq-004 | Standardize LLM Client Constructors | Small | Pending |
| 2 | cq-003 | Reduce Logic in Command Layer | Small | Pending |
| 3 | cq-002 | Refactor Configuration Validation | Medium | Deferred |

---

## Priority 1: Standardize LLM Client Constructors (cq-004)

**Category**: duplication  
**Severity**: suggestion  
**Effort**: small  
**Breaking Change**: false

### Problem

The `NewClient` functions in `internal/platform/openai`, `anthropic`, and `geminiapi` contain nearly identical boilerplate code:
- API key validation
- Base URL fallback logic
- Timeout conversion (seconds → `time.Duration`)
- Model default assignment
- Identical `Send/SendWithProgress` wrappers

### Solution

Move validation and default-setting logic into `llmbase.NewValidatedConfig` or a factory method:

```go
// internal/platform/llmbase/config.go
type ProviderDefaults struct {
    BaseURL string
    Model   string
    Timeout time.Duration
}

func NewValidatedConfig(cfg llm.Config, defaults ProviderDefaults) (*Config, error) {
    if cfg.APIKey == "" {
        return nil, fmt.Errorf("API key is required")
    }
    
    baseURL := cfg.BaseURL
    if baseURL == "" {
        baseURL = defaults.BaseURL
    }
    
    model := cfg.Model
    if model == "" {
        model = defaults.Model
    }
    
    timeout := time.Duration(cfg.Timeout) * time.Second
    if timeout == 0 {
        timeout = defaults.Timeout
    }
    
    return &Config{
        APIKey:   cfg.APIKey,
        BaseURL:  baseURL,
        Model:    model,
        Timeout:  timeout,
    }, nil
}
```

Provider constructors become ~10 lines:

```go
// internal/platform/openai/client.go
func NewClient(cfg llm.Config) (*Client, error) {
    baseCfg, err := llmbase.NewValidatedConfig(cfg, llmbase.ProviderDefaults{
        BaseURL: "https://api.openai.com/v1",
        Model:   "gpt-4o",
        Timeout: 300 * time.Second,
    })
    if err != nil {
        return nil, err
    }
    return &Client{BaseClient: llmbase.NewBaseClient(baseCfg)}, nil
}
```

### Affected Files

- `internal/platform/llmbase/base_client.go`
- `internal/platform/openai/client.go`
- `internal/platform/anthropic/client.go`
- `internal/platform/geminiapi/client.go`

### Benefits

- Eliminates ~60 lines of duplicated code
- Reduces risk of inconsistent provider behavior
- Makes adding new providers simpler

---

## Priority 2: Reduce Logic in Command Layer (cq-003)

**Category**: structure  
**Severity**: minor  
**Effort**: small  
**Breaking Change**: false

### Problem

The `cmd/context.go` file contains ~100 lines of business logic in `buildGenerateConfig`:
- Flag parsing and conversion
- Path resolution and validation
- Configuration merging (flags + viper)
- Progress rendering

This violates Clean Architecture by placing orchestration logic in the presentation layer.

### Solution

1. **Move config builder to application layer**:

```go
// internal/app/cli_config.go
type GenerateFlags struct {
    RootPath       string
    Include        []string
    Exclude        []string
    Output         string
    MaxSize        string
    EnforceLimit   bool
    Template       string
    Task           string
    Rules          string
    CustomVars     []string
    Workers        int
    IncludeHidden  bool
    IncludeIgnored bool
    ProgressMode   string
}

func NewGenerateConfigFromFlags(flags GenerateFlags) (GenerateConfig, error) {
    // All the logic currently in buildGenerateConfig
}
```

2. **Move progress rendering to shared location**:

```go
// internal/ui/cli/progress.go
type ProgressMode string

const (
    ProgressNone  ProgressMode = "none"
    ProgressHuman ProgressMode = "human"
    ProgressJSON  ProgressMode = "json"
)

func RenderProgress(mode ProgressMode, output ProgressOutput) { ... }
```

3. **Simplify cmd/context.go**:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    flags := extractFlags(cmd)
    config, err := app.NewGenerateConfigFromFlags(flags)
    if err != nil {
        return err
    }
    return generateContextHeadless(config)
}
```

### Affected Files

- `cmd/context.go` (simplify)
- `internal/app/cli_config.go` (create)
- `internal/ui/cli/progress.go` (create)

### Benefits

- Enables testing config building without Cobra
- Reusable progress rendering for other CLI commands
- Cleaner separation of concerns

---

## Deferred: Refactor Configuration Validation (cq-002)

**Category**: design_pattern  
**Severity**: minor  
**Effort**: medium  
**Breaking Change**: false  
**Defer Until**: Config keys exceed 40-50

### Problem

The `ValidateValue` function uses a hardcoded switch statement (~25 cases). Adding a new key requires modifying both `metadata.go` and `validator.go`.

### Current Assessment

This is **not a significant problem** currently because:
- All config-related code is in `internal/config/`
- The switch is well-organized and explicit
- Config keys change infrequently
- The current pattern is debuggable

### If Implemented

Use a validator registry instead of embedding in metadata (avoids circular dependencies):

```go
// internal/config/validators.go
var validatorRegistry = map[string]func(string) error{
    KeyScannerMaxFiles:    validateMaxFiles,
    KeyScannerMaxFileSize: validateSizeFormat,
    // ...
}

func ValidateValue(key, value string) error {
    if validator, ok := validatorRegistry[key]; ok {
        return validator(value)
    }
    return nil
}
```

---

## Removed Items

### ~~cq-001: Refactor ConfigWizardModel (God Object)~~

**Status**: Already implemented

The codebase already uses the Composed Screen Model Architecture:
- `WizardModel` delegates to 5 screen models in `internal/ui/screens/`
- `ConfigWizardModel` delegates to `ConfigCategoryModel` instances
- Heavy async operations use `ScanCoordinator` and `GenerateCoordinator`

See README section "Composed Screen Model Architecture" for documentation.

### ~~cq-005: Consolidate UI Styling~~

**Status**: Deferred indefinitely

**Reason**: The proposed DI-based theming adds significant complexity without clear benefit:
- Global styles work well for the current use case
- `lipgloss.Style` values are immutable (copied on modification)
- No current requirement for runtime theme switching
- Breaking change that affects all UI components

**Revisit when**: There's a concrete requirement for light/dark mode toggle or user-customizable themes.

---

## Summary Statistics

| Category | Count |
|----------|-------|
| Active (to implement) | 2 |
| Deferred | 1 |
| Removed (already done) | 1 |
| Removed (low value) | 1 |

**Estimated Total Effort**: Small (2-3 focused sessions)
