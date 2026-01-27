# Code Quality Improvements - Validated

**Analysis Date**: 2026-01-26
**Validated By**: Deep codebase investigation with explore agents

---

## Priority 1: Package Naming (Quick Win)

### cq-002: Rename `internal/core/context` to `internal/core/contextgen`

**Status**: Approved
**Effort**: Small (30 min)
**Breaking**: Internal only

**Problem**:
- Package name collides with stdlib `context` package
- 8 files import it, 50% require `ctxgen` alias
- Inconsistent import styles across codebase

**Solution**:
1. Rename directory: `internal/core/context` -> `internal/core/contextgen`
2. Update package declaration in all files
3. Remove all `ctxgen` aliases, use default package name
4. Run `go mod tidy`

**Affected Files**:
- `internal/core/context/` (rename)
- `internal/app/service.go`
- `internal/app/service_test.go`
- `internal/ui/wizard.go`
- `internal/ui/wizard_test.go`
- `internal/ui/generate_coordinator.go`
- `internal/ui/generate_coordinator_test.go`

---

## Priority 2: LLM Provider DRY (Convenience)

### cq-003: Consolidate LLM Provider Initialization

**Status**: Approved with adjustments
**Effort**: Small
**Breaking**: No

**Problem**:
- `NewClient` functions are 80% identical across OpenAI, Anthropic, GeminiAPI
- Same API key validation, timeout defaults, URL defaults
- `Send`/`SendWithProgress` are 100% boilerplate wrappers

**Solution**:
1. Create `llmbase.NewBaseClient(cfg Config, defaults DefaultConfig, sender Sender)` factory
2. Move validation and defaulting logic into factory
3. Store `Sender` in `BaseClient` to eliminate wrapper methods
4. Providers become ~30 lines instead of ~80

**Implementation Pattern**:
```go
// In llmbase/base_client.go
type DefaultConfig struct {
    BaseURL    string
    Model      string
    MaxTokens  int
    Timeout    time.Duration
}

func NewBaseClient(cfg llm.Config, defaults DefaultConfig, sender Sender) (*BaseClient, error) {
    if cfg.APIKey == "" {
        return nil, errors.New("API key is required")
    }
    baseURL := cfg.BaseURL
    if baseURL == "" {
        baseURL = defaults.BaseURL
    }
    // ... apply all defaults
    return &BaseClient{...}, nil
}

// In openai/client.go - becomes trivial
func NewClient(cfg llm.Config) (llm.Provider, error) {
    client := &Client{}
    base, err := llmbase.NewBaseClient(cfg, DefaultConfig{
        BaseURL:   "https://api.openai.com/v1",
        Model:     "gpt-4o",
        Timeout:   300 * time.Second,
    }, client)
    if err != nil {
        return nil, err
    }
    client.BaseClient = base
    return client, nil
}
```

**When to Do**: Before adding a 4th LLM provider.

---

## Discarded Items

### cq-001: Wizard "God Object" Refactoring

**Status**: Discarded
**Reason**: Misdiagnosis - architecture already follows recommended patterns

**Evidence**:
- Update method is 130 lines (acceptable for MVU orchestrator)
- Already delegates to 5 screen models (`FileSelectionModel`, `TemplateSelectionModel`, etc.)
- Already uses coordinators (`ScanCoordinator`, `GenerateCoordinator`)
- Complexity (33) is documented as "required by Bubble Tea framework"
- Comprehensive test coverage exists

**Conclusion**: The wizard is a Mediator/Orchestrator, not a God Object. Refactoring would add abstraction without functional benefit.

---

### cq-004: IntelligentSplit Complexity

**Status**: Discarded
**Reason**: Function is already well-designed and stable

**Evidence**:
- Cyclomatic complexity is **11** (project limit is 25)
- Function is only **62 lines**
- Test coverage is **98.4%**
- Logic is linear and deterministic

**Conclusion**: The "fragility" claim is false. Refactoring would increase complexity without benefit.

---

### cq-005: XML Encoding for File Content

**Status**: Discarded
**Reason**: Intentional design choice, not a code smell

**Evidence**:
- The `<file path="...">` format is pseudo-XML designed for LLM readability
- Not intended to be valid XML
- Using `xml.Encoder` would add overhead and potentially break LLM parsing
- File paths don't contain XML-sensitive characters in practice

**Conclusion**: The current implementation is explicit, debuggable, and fit for purpose.

---

## Summary

| ID | Status | Priority | Effort |
|----|--------|----------|--------|
| cq-002 | **Approved** | 1 | Small |
| cq-003 | **Approved** | 2 | Small |
| cq-001 | Discarded | - | - |
| cq-004 | Discarded | - | - |
| cq-005 | Discarded | - | - |

**Original issues found**: 5
**Valid after investigation**: 2 (40%)
**False positives**: 3 (60%)
