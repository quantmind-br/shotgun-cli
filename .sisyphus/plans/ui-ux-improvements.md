# UI/UX Improvements Plan

## Context

### Original Request
Implement UI/UX improvements defined in `uiux.md`, specifically targeting critical usability, visual polish, and interaction enhancements.

### Interview Summary
**Key Discussions**:
- **Scope**: Focus on Phase 1 (Critical Usability) and Phase 2 (Visual Polish). Phase 3 (Interaction) is lower priority but included.
- **Constraints**: Maintain terminal compatibility (min 40x10), strictly scoped refactoring (no business logic changes), and use existing `lipgloss` styles.

**Research Findings**:
- **Progress Bar**: `renderSizeLimitSection` in `review.go` contains conditional coloring logic (Green/Yellow/Red) missing from `theme.go`. Needs extraction to `internal/ui/components/progress.go`.
- **Scrollbars**: `tree.go` handles scrolling manually. Scrollbar logic needs to be added to `View()`, rendering side-by-side with the tree content.
- **CLI Styling**: `cmd/llm.go` uses standard `fmt`/`tabwriter`. Can safely import `internal/ui/styles` as `lipgloss` handles TTY detection.

### Metis Review
**Identified Gaps** (addressed):
- **Scope Drift**: Explicitly limited Phase 1 to extraction/reuse, no new logic.
- **Regressions**: Added regression testing steps for TUI layout and CLI output.
- **Edge Cases**: Included checks for small terminals and empty lists.
- **Scrollbar Visibility**: Clarified track visibility (only when exceeding viewport).
- **Threshold Config**: Confirmed thresholds must mirror `review.go` exactly.

---

## Work Objectives

### Core Objective
Enhance the usability and visual consistency of `shotgun-cli` by implementing a visual token budget bar, stylizing CLI output, and adding scrollbars to long lists.

### Concrete Deliverables
- **New Component**: `internal/ui/components/progress.go` with `UsageBar` struct.
- **Refactored Screens**: `file_selection.go` (uses UsageBar), `review.go` (uses UsageBar).
- **Refactored Components**: `tree.go` (adds scrollbar).
- **Refactored CLI**: `cmd/llm.go`, `cmd/config.go` (styled output).

### Definition of Done
- [x] File Selection screen shows a visual token budget bar that updates with selections.
- [x] CLI commands (`llm status`, `config show`) output styled text (colors/bold).
- [x] File Tree shows a visual scrollbar instead of "↑ more above" text (only when content exceeds viewport).
- [x] No regression in layout on 40x10 terminals.
- [x] Progress bar logic in `UsageBar` mirrors existing `review.go` thresholds exactly.

### Must Have
- Visual progress bar on File Selection.
- Styled CLI status output (labels and values).
- Vertical scrollbars on File Tree (conditionally visible).

### Must NOT Have (Guardrails)
- Business logic changes (scanning/generation remain untouched).
- New external dependencies (use existing `lipgloss` and `bubbletea`).
- Layout breakage on small screens.

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: Yes (Go tests).
- **User wants tests**: Manual verification is primary for UI; unit tests for new component logic.
- **QA Approach**: Manual TUI verification + Unit tests for `UsageBar` logic.

### Manual QA Only

**CRITICAL**: Verification must cover TUI interaction and visual correctness.

**For TUI changes:**
- [x] Run `go run .` (Wizard):
    - **Step 1 (File Selection)**: Select files until >80% context limit. Verify bar turns yellow/red.
    - **Scrollbar**: Navigate a long directory. Verify scrollbar moves correctly.
- [x] Run `go run . llm status` (CLI):
    - Verify output uses colors/bold styles instead of plain text.

---

## Task Flow

```
Task 1 → Task 2 → Task 3 → Task 4
```

## Parallelization

| Task | Depends On | Reason |
|------|------------|--------|
| 1 | - | Independent (New component) |
| 2 | 1 | Uses component from Task 1 |
| 3 | - | Independent (CLI styling) |
| 4 | - | Independent (Tree component) |

---

## TODOs

- [x] 1. Create UsageBar Component

  **What to do**:
  - Create `internal/ui/components/progress.go`.
  - Implement `UsageBar` struct (current, max, width).
  - Implement logic: Green (<80%), Yellow (<100%), Red (>=100%).
  - Add unit tests for color logic.

  **References**:
  - `internal/ui/screens/review.go:359` - Original logic source.

  **Acceptance Criteria**:
  - [x] `go test ./internal/ui/components/...` passes.
  - [x] Logic correctly identifies safe/warning/critical states.

  **Commit**: YES
  - Message: `feat(ui): add reusable UsageBar component`
  - Files: `internal/ui/components/progress.go`

---

- [x] 2. Integrate UsageBar into Screens

  **What to do**:
  - Modify `internal/ui/screens/review.go` to use `UsageBar`.
  - Modify `internal/ui/screens/file_selection.go` to use `UsageBar`.
  - Remove manual rendering logic from `review.go`.

  **Must NOT do**:
  - Change how context size is calculated.

  **References**:
  - `internal/ui/screens/review.go` - Replace `renderSizeLimitSection`.
  - `internal/ui/screens/file_selection.go` - Add to `View`.

  **Acceptance Criteria**:
  - [x] File Selection screen shows progress bar.
  - [x] Review screen still shows progress bar (no regression).

  **Commit**: YES
  - Message: `refactor(ui): integrate UsageBar into selection and review screens`
  - Files: `internal/ui/screens/review.go`, `internal/ui/screens/file_selection.go`

---

- [x] 3. Stylize CLI Status Commands

  **What to do**:
  - Import `internal/ui/styles` in `cmd/llm.go` and `cmd/config.go`.
  - Replace `fmt.Printf` with `styles.TitleStyle.Render`, `styles.SuccessStyle.Render`, etc.
  - Apply to `shotgun-cli llm status` and `shotgun-cli config show`.

  **Must NOT do**:
  - Change the content/data output, only the formatting.

  **Acceptance Criteria**:
  - [x] `shotgun-cli llm status` shows colored/styled output.
  - [x] `shotgun-cli config show` shows styled keys/values.

  **Commit**: YES
  - Message: `style(cli): apply rich styling to status commands`
  - Files: `cmd/llm.go`, `cmd/config.go`

---

- [x] 4. Implement Vertical Scrollbars

  **What to do**:
  - Modify `internal/ui/components/tree.go`.
  - Calculate `thumbHeight` and `thumbPos` based on `visibleItems`, `height`, and `topIndex`.
  - Render scrollbar (`│` track, `█` thumb) to the right of the tree view.
  - Remove "↑ more above" / "↓ more below" text lines.

  **References**:
  - `internal/ui/components/tree.go:View` - Target method.

  **Acceptance Criteria**:
  - [x] Long lists show a vertical scrollbar.
  - [x] Scrollbar thumb moves proportionally with navigation.
  - [x] No layout overflow on small screens.

  **Commit**: YES
  - Message: `feat(ui): add vertical scrollbars to file tree`
  - Files: `internal/ui/components/tree.go`

---

## Success Criteria

### Verification Commands
```bash
# Verify UI components
go run . config # Check scrollbars in config categories if long list
go run .        # Check file selection progress bar and tree scrollbar
go run . llm status # Check CLI styling
```

### Final Checklist
- [x] UsageBar component created and tested.
- [x] File Selection screen shows progress bar.
- [x] Review screen shows progress bar.
- [x] CLI commands are styled.
- [x] File tree has working scrollbars.
