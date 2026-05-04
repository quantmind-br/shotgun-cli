# UI Package - TUI Wizard

Parent: [../AGENTS.md](../AGENTS.md)

## OVERVIEW

Bubble Tea-based interactive wizard. **MVU pattern** with composed screen models and async coordinators.

## STRUCTURE

```
ui/
├── wizard.go              # Main orchestrator, 5-step state machine (1062 lines)
├── scan_coordinator.go    # Async scan state management
├── generate_coordinator.go # Async generation state management
├── config_wizard.go       # Config TUI (interactive settings editor)
├── screens/               # Individual wizard screens (each owns its state)
│   ├── file_selection.go  # File tree with filter, selection, ignore toggle
│   ├── template_selection.go  # Template list with preview modal
│   ├── task_input.go      # Task description textarea
│   ├── rules_input.go     # Rules textarea (optional)
│   └── review.go          # Summary, generate, copy, send to LLM
├── components/            # Reusable TUI components
│   ├── tree.go           # File tree renderer (683 lines)
│   ├── progress.go       # Progress indicators
│   ├── modal.go          # Modal dialogs
│   ├── config_field.go   # Config field editor
│   ├── config_toggle.go  # Boolean toggle
│   └── config_select.go  # Dropdown selector
└── styles/               # Theme and styling
    └── theme.go          # Colors, borders, layout constants
```

## SCREEN MODEL ARCHITECTURE

Each screen owns its state. `WizardModel` delegates via accessors:

```go
// WizardModel composes screen models
fileSelection      *FileSelectionModel
templateSelection  *TemplateSelectionModel
taskInput          *TaskInputModel
rulesInput         *RulesInputModel
review             *ReviewModel

// Accessors delegate to screen models
func (m *WizardModel) getSelectedFiles() map[string]bool
func (m *WizardModel) getSelectedTemplate() *template.Template
func (m *WizardModel) getTaskDesc() string
func (m *WizardModel) getRules() string
```

## COORDINATOR PATTERN

Background operations use coordinators (not direct service calls):

```go
// Scan: channel-based progress
scanCoordinator.Start(rootPath, config)     // Returns tea.Cmd
scanCoordinator.Poll()                      // Returns Batch(ProgressMsg, NextPoll)
scanCoordinator.Result()                    // Returns (*FileNode, error)

// Generate: same pattern
generateCoordinator.Start(generateConfig)
```

**Message flow**: Start → Poll loop → ProgressMsg updates UI → CompleteMsg triggers Result()

## WIZARD STEPS

| Step | Constant | Screen | Key Actions |
|------|----------|--------|-------------|
| 1 | `StepFileSelection` | File tree | Navigate, filter (`/`), select, toggle ignored (`i`) |
| 2 | `StepTemplateSelection` | Template list | Select, preview (`v`) |
| 3 | `StepTaskInput` | Textarea | Describe task |
| 4 | `StepRulesInput` | Textarea | Optional rules |
| 5 | `StepReview` | Summary | Generate (F8), copy (`c`), send to LLM (F9) |

## KEYBOARD SHORTCUTS

| Global | Action |
|--------|--------|
| F1 | Help |
| F7/Ctrl+P | Previous step |
| F8/Ctrl+N | Next step |
| Ctrl+Q | Quit |

| File Selection | Action |
|----------------|--------|
| ↑/↓/k/j | Navigate |
| ←/→/h/l | Collapse/Expand |
| Space | Toggle selection |
| / | Filter mode |
| a/A | Select/Deselect all |
| i | Toggle ignored files |
| F5 | Rescan |

**Terminal minimum**: 40x10 (columns x rows). Warning overlay if too small.

## ADDING A NEW SCREEN

1. Create `internal/ui/screens/my_screen.go` implementing `tea.Model`
2. Screen owns its state (NOT stored in `WizardModel`)
3. Add as field in `WizardModel`
4. Route messages in `WizardModel.Update()` based on current step
5. Render in `WizardModel.View()`

## TESTING

```bash
go test -v -race ./internal/ui/...                           # All UI tests
go test -v -run TestFileSelection ./internal/ui/screens/     # Specific screen
go test -v -run TestWizard ./internal/ui/                    # Wizard orchestration
go test -v -run TestScanCoordinator ./internal/ui/           # Coordinator
```

**CI skips**: `TestScanCoordinator`, `TestGenerateCoordinator`, `TestWizardClipboardCopyCmd` (env-dependent)

## ANTI-PATTERNS

- Storing screen state in WizardModel (use composed models)
- Blocking in Update() (use coordinators for async ops)
- Direct service calls from screens (go through wizard → coordinator)
- Skipping progress callbacks (UI freezes without them)