# Code Quality Improvements: cq-002 + cq-003

## Context

### Original Request
Implement two code quality improvements from the validated analysis:
1. **cq-002**: Rename `internal/core/context` → `internal/core/contextgen` (package naming collision fix)
2. **cq-003**: Consolidate LLM provider initialization logic (DRY refactoring)

### Interview Summary
**Key Discussions**:
- **PR Strategy**: Separate PRs for each improvement (easier to review/revert)
- **Test Strategy**: Use existing tests (85%+ coverage), run `go test ./...` after changes
- **No TDD**: Existing coverage is sufficient for these refactors

**Research Findings**:
- cq-002 affects 8 files (confirmed via grep)
- cq-003 involves 3 providers with 80% duplicated `NewClient` code
- `BaseClient` already handles HTTP/progress, but initialization is duplicated

### Metis Review
**Identified Gaps** (addressed):
- Package name should be `contextgen` (both path and package declaration)
- Preserve all existing behavior, error messages, defaults, and logging
- No changes to config keys, env var names, or provider registration behavior
- Keep PRs strictly separated
- Validate no docs/README references to old package path

---

## Work Objectives

### Core Objective
Improve code quality through package rename and DRY refactoring without changing behavior.

### Concrete Deliverables
1. **PR #1 (cq-002)**: Renamed package `internal/core/contextgen` with clean imports
2. **PR #2 (cq-003)**: New `llmbase.NewBaseClient` factory with simplified providers

### Definition of Done
- [ ] `go test -race ./...` passes (0 failures)
- [ ] `golangci-lint run ./...` passes (0 new errors)
- [ ] No `ctxgen` aliases remain in codebase
- [ ] All 3 providers use the new factory
- [ ] PRs created and passing CI

### Must Have
- Behavioral equivalence (same defaults, same validation, same errors)
- All existing tests pass unmodified
- Clean separation between the two PRs

### Must NOT Have (Guardrails)
- Changes to request/response parsing logic
- Changes to config keys or env var names
- New abstractions beyond `llmbase.NewBaseClient`
- Formatting changes to unrelated files
- Test modifications (tests should pass as-is)
- Any cq-003 changes in cq-002 PR (and vice versa)

---

## Verification Strategy (MANDATORY)

### Test Decision
- **Infrastructure exists**: YES
- **User wants tests**: Existing tests (no TDD)
- **Framework**: `go test`

### Verification Commands
```bash
# After each task
go test -race ./...

# Before committing
golangci-lint run ./...

# Verify no regressions
go build ./...
```

---

## Task Flow

```
PR #1 (cq-002):
  Task 1 → Task 2 → Task 3 → Commit/PR

PR #2 (cq-003):
  Task 4 → Task 5 → Task 6 → Task 7 → Commit/PR
```

## Parallelization

| Task | Depends On | Reason |
|------|------------|--------|
| 1 | None | First task |
| 2 | 1 | Imports depend on renamed package |
| 3 | 2 | Verification after imports fixed |
| 4 | 3 (PR #1 merged) | cq-003 starts after cq-002 complete |
| 5 | 4 | Providers depend on factory |
| 6 | 5 | All providers must be updated |
| 7 | 6 | Final verification |

---

## TODOs

---

### PR #1: Package Rename (cq-002)

---

- [x] 1. Rename package directory and update package declaration

  **What to do**:
  - Rename directory: `internal/core/context` → `internal/core/contextgen`
  - Update package declaration in all files within the directory to `package contextgen`
  - Files to update: `generator.go`, `content.go`, and any other `.go` files in the package

  **Must NOT do**:
  - Change any function signatures or behavior
  - Rename types or functions
  - Modify any logic

  **Parallelizable**: NO (first task)

  **References**:
  - `internal/core/context/generator.go` - Main generator interface and implementation
  - `internal/core/context/content.go` - Content collection logic

  **Acceptance Criteria**:
  - [ ] Directory renamed to `internal/core/contextgen/`
  - [ ] All files declare `package contextgen`
  - [ ] `go build ./internal/core/contextgen/...` succeeds

  **Commit**: NO (groups with task 2)

---

- [x] 2. Update all import paths and remove aliases

  **What to do**:
  - Update imports from `"github.com/quantmind-br/shotgun-cli/internal/core/context"` to `"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"`
  - Remove `ctxgen` aliases where used
  - Replace `ctxgen.` prefixes with `contextgen.`

  **Affected files**:
  - `internal/app/service.go` - Uses `ctxgen` alias
  - `internal/app/service_test.go` - Uses `ctxgen` alias
  - `internal/ui/wizard.go` - Direct import
  - `internal/ui/wizard_test.go` - Direct import
  - `internal/ui/generate_coordinator.go` - Direct import
  - `internal/ui/generate_coordinator_test.go` - Direct import

  **Must NOT do**:
  - Change any code logic
  - Modify test assertions
  - Touch unrelated imports

  **Parallelizable**: NO (depends on task 1)

  **References**:
  - `internal/app/service.go:3-15` - Import block with `ctxgen` alias
  - `internal/ui/wizard.go:3-20` - Import block
  - `internal/ui/generate_coordinator.go:3-15` - Import block

  **Acceptance Criteria**:
  - [ ] `grep -r "internal/core/context\"" .` returns no matches
  - [ ] `grep -r "ctxgen" .` returns no matches
  - [ ] `go build ./...` succeeds

  **Commit**: NO (groups with task 3)

---

- [x] 3. Verify and commit PR #1

  **What to do**:
  - Run full test suite
  - Run linter
  - Create commit with descriptive message
  - Push and create PR

  **Must NOT do**:
  - Include any cq-003 related changes

  **Parallelizable**: NO (depends on task 2)

  **References**:
  - `AGENTS.md:Build/Test/Lint Commands` - Test and lint commands

  **Acceptance Criteria**:
  - [ ] `go test -race ./...` passes (all tests green)
  - [ ] `golangci-lint run ./...` passes
  - [ ] `git status` shows only expected files changed

  **Commit**: YES
  - Message: `refactor(core): rename context package to contextgen`
  - Files: `internal/core/contextgen/*`, `internal/app/service*.go`, `internal/ui/*.go`
  - Pre-commit: `go test -race ./... && golangci-lint run`

---

### PR #2: LLM Provider DRY (cq-003)

---

- [x] 4. Create `DefaultConfig` struct and `NewBaseClient` factory in llmbase

  **What to do**:
  - Add `DefaultConfig` struct to hold provider defaults (BaseURL, Model, MaxTokens, Timeout)
  - Create `NewBaseClient(cfg llm.Config, defaults DefaultConfig, sender Sender) (*BaseClient, error)` factory
  - Move common validation and defaulting logic into the factory:
    - API key required check
    - BaseURL default fallback
    - Timeout default (convert int seconds to time.Duration)
    - Model default fallback
    - MaxTokens default
  - Store `Sender` in `BaseClient` struct

  **Must NOT do**:
  - Change existing `BaseClient` behavior
  - Modify HTTP handling or progress reporting
  - Change error message text

  **Parallelizable**: NO (first task of PR #2)

  **References**:
  - `internal/platform/llmbase/base_client.go` - Existing BaseClient implementation
  - `internal/platform/llmbase/sender.go` - Sender interface definition
  - `internal/platform/openai/client.go:NewClient` - Pattern to consolidate

  **Acceptance Criteria**:
  - [ ] `DefaultConfig` struct exists with BaseURL, Model, MaxTokens, Timeout fields
  - [ ] `NewBaseClient` function exists and compiles
  - [ ] Factory performs API key validation (returns error if empty)
  - [ ] Factory applies default values for BaseURL, Model, Timeout
  - [ ] `go build ./internal/platform/llmbase/...` succeeds

  **Commit**: NO (groups with tasks 5, 6)

---

- [x] 5. Refactor OpenAI provider to use new factory

  **What to do**:
  - Update `openai.NewClient` to use `llmbase.NewBaseClient`
  - Pass OpenAI-specific defaults:
    - BaseURL: `https://api.openai.com/v1`
    - Model: `gpt-4o`
    - Timeout: `300 * time.Second`
  - Remove duplicated validation and defaulting logic
  - Keep `Sender` interface implementation (BuildRequest, ParseResponse, etc.)

  **Must NOT do**:
  - Change default values
  - Modify request/response handling
  - Change error messages

  **Parallelizable**: YES (with tasks 5, 6 after task 4)

  **References**:
  - `internal/platform/openai/client.go` - Current implementation
  - `internal/platform/openai/client.go:NewClient` - Function to refactor

  **Acceptance Criteria**:
  - [ ] `NewClient` uses `llmbase.NewBaseClient`
  - [ ] No duplicated validation logic remains
  - [ ] `go test -race ./internal/platform/openai/...` passes

  **Commit**: NO (groups with tasks 4, 6, 7)

---

- [x] 6. Refactor Anthropic and GeminiAPI providers

  **What to do**:
  - Update `anthropic.NewClient` to use `llmbase.NewBaseClient`
    - BaseURL: `https://api.anthropic.com`
    - Model: `claude-sonnet-4-20250514`
    - MaxTokens: `8192`
    - Timeout: `300 * time.Second`
  - Update `geminiapi.NewClient` to use `llmbase.NewBaseClient`
    - BaseURL: `https://generativelanguage.googleapis.com`
    - Model: `gemini-2.5-flash`
    - MaxTokens: `8192`
    - Timeout: `300 * time.Second`
  - Remove duplicated validation and defaulting logic from both

  **Must NOT do**:
  - Change default values
  - Modify request/response handling
  - Change error messages

  **Parallelizable**: YES (with task 5 after task 4)

  **References**:
  - `internal/platform/anthropic/client.go` - Current implementation
  - `internal/platform/geminiapi/client.go` - Current implementation

  **Acceptance Criteria**:
  - [ ] Both `NewClient` functions use `llmbase.NewBaseClient`
  - [ ] No duplicated validation logic remains
  - [ ] `go test -race ./internal/platform/anthropic/...` passes
  - [ ] `go test -race ./internal/platform/geminiapi/...` passes

  **Commit**: NO (groups with tasks 4, 5, 7)

---

- [x] 7. Verify and commit PR #2

  **What to do**:
  - Run full test suite
  - Run linter
  - Verify all 3 providers work correctly
  - Create commit with descriptive message
  - Push and create PR

  **Must NOT do**:
  - Include any cq-002 related changes

  **Parallelizable**: NO (depends on tasks 4, 5, 6)

  **References**:
  - `AGENTS.md:Build/Test/Lint Commands` - Test and lint commands

  **Acceptance Criteria**:
  - [ ] `go test -race ./...` passes (all tests green)
  - [ ] `golangci-lint run ./...` passes
  - [ ] Provider tests verify behavioral equivalence

  **Commit**: YES
  - Message: `refactor(llm): consolidate provider initialization with factory`
  - Files: `internal/platform/llmbase/*`, `internal/platform/openai/*`, `internal/platform/anthropic/*`, `internal/platform/geminiapi/*`
  - Pre-commit: `go test -race ./... && golangci-lint run`

---

## Commit Strategy

| After Task | Message | Files | Verification |
|------------|---------|-------|--------------|
| 3 | `refactor(core): rename context package to contextgen` | contextgen/*, app/service*.go, ui/*.go | go test -race ./... |
| 7 | `refactor(llm): consolidate provider initialization with factory` | llmbase/*, openai/*, anthropic/*, geminiapi/* | go test -race ./... |

---

## Success Criteria

### Verification Commands
```bash
# After PR #1
grep -r "internal/core/context\"" .  # Expected: no matches
grep -r "ctxgen" .                   # Expected: no matches

# After PR #2
# All providers should have simplified NewClient (~30 lines vs ~80)

# Both PRs
go test -race ./...                  # Expected: PASS
golangci-lint run ./...              # Expected: 0 errors
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All tests pass
- [ ] No `ctxgen` aliases in codebase
- [ ] All 3 providers use factory
- [ ] 2 separate PRs created
