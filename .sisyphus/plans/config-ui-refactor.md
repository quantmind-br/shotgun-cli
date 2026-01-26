# Refactor: Config UI Toggle Visibility

## Context

### Original Request
Modify the "on-off" buttons in `shotgun-cli config` TUI to be more intuitive, as the current visualization is confusing.

### Interview Summary
**Key Discussions**:
- **Visual Preference**: User selected "Explicit Text" (e.g., `[ENABLED]` / `[DISABLED]`) over switches or icons.
- **Pain Point**: Current "dot" switch is ambiguous.

**Research Findings**:
- **Target File**: `internal/ui/components/config_toggle.go`
- **Method**: `renderToggle` (lines 120-145) responsible for rendering.
- **Current Style**: Custom "OFF ━━◉ ON" switch using lipgloss.

### Metis Review
**Identified Gaps** (addressed):
- **Accessibility**: Ensured text labels (`[ENABLED]`/`[DISABLED]`) provide non-color cues.
- **Focus State**: Explicitly required preservation of focus styling to differentiate the active row.
- **Layout**: Noted potential width increase; explicit text is wider than the switch.

---

## Work Objectives

### Core Objective
Replace the ambiguous switch visualization in the configuration wizard with explicit, text-based status indicators (`[ENABLED]` / `[DISABLED]`) to improve clarity and reduce user error.

### Concrete Deliverables
- Modified `internal/ui/components/config_toggle.go` implementing the new visual style.

### Definition of Done
- [x] Toggle displays `[ENABLED]` in Green when true.
- [x] Toggle displays `[DISABLED]` in Gray/Dim when false.
- [x] Focus state (when user navigates to the row) remains clearly visible.
- [x] Manual verification in TUI confirms readability and behavior.

### Must Have
- Explicit text labels.
- Color coding (Green=On, Gray=Off).
- Brackets `[]` for structure.

### Must NOT Have (Guardrails)
- Logic changes (no changing how config is saved).
- New dependencies (use existing `lipgloss`).
- Layout breakage (must fit within standard terminal width).

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: Yes (Go tests).
- **User wants tests**: Manual verification is sufficient for this pure UI change, but we will run existing tests to ensure no regressions.
- **QA Approach**: Manual TUI verification.

### Manual QA Only

**CRITICAL**: Verification must cover both states and focus interaction.

**For TUI changes:**
- [x] Using `interactive_bash` (tmux session) or manual run:
    - Command: `go run . config`
    - Action: Navigate to a boolean setting (e.g., `context.save_response` or similar).
    - Verify:
        1. Default state is readable (`[DISABLED]` or `[ENABLED]`).
        2. Press `Space` to toggle.
        3. State changes to the opposite text and color immediately.
        4. Focus indicator (border/highlight) is visible on the active row.

---

## TODOs

- [x] 1. Refactor `renderToggle` to use explicit text

  **What to do**:
  - Modify `internal/ui/components/config_toggle.go`.
  - Replace the "OFF ━━◉ ON" string builder logic.
  - Implement:
    - **Enabled**: `[ENABLED]` (Green style).
    - **Disabled**: `[DISABLED]` (Gray/Faint style).
  - Ensure the `m.Focused()` check continues to apply the focus style (likely a border or background) around the component or label.

  **References**:
  - `internal/ui/components/config_toggle.go:120` - `renderToggle` method.
  - `internal/ui/styles.go` - Check for existing common styles (e.g., `ActiveColor`, `InactiveColor`).

  **Acceptance Criteria**:
  - [x] Run `go run . config`
  - [x] Observe `[ENABLED]` in Green for active settings.
  - [x] Observe `[DISABLED]` in Gray for inactive settings.
  - [x] Toggle works with Spacebar.

  **Commit**: YES
  - Message: `ui(config): replace toggle switch with explicit enabled/disabled text`
  - Files: `internal/ui/components/config_toggle.go`

---

## Success Criteria

### Verification Commands
```bash
go run . config
# Navigate and toggle options to verify visual changes
```

### Final Checklist
- [x] "On" state says `[ENABLED]` (Green)
- [x] "Off" state says `[DISABLED]` (Gray)
- [x] Toggling works instantly
- [x] No layout alignment issues
