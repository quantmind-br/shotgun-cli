# shotgun-cli UI Patterns (Bubble Tea)

## Wizard Architecture

### Composed Screen Model Pattern

The `WizardModel` uses composition to delegate screen-specific state to dedicated screen models. This keeps `WizardModel` focused on coordination while each screen owns its own state.

```
WizardModel (coordination only)
├── fileSelection     → FileSelectionModel (tree, selections, filter)
├── templateSelection → TemplateSelectionModel (templates, selected)
├── taskInput         → TaskInputModel (task description)
├── rulesInput        → RulesInputModel (rules text)
└── review            → ReviewModel (summary display)
```

**Accessor Pattern:**
```go
// WizardModel delegates state access to screen models
func (m *WizardModel) getSelectedFiles() map[string]bool {
    if m.fileSelection != nil {
        return m.fileSelection.GetSelections()
    }
    return nil
}
```

### Main Wizard Model (`internal/ui/wizard.go`)
- `WizardModel` orchestrates all 5 steps
- Steps defined as constants: `StepFileSelection`, `StepTemplateSelection`, `StepTaskInput`, `StepRulesInput`, `StepReview`
- Handles navigation between steps and async operations
- Screen-specific state delegated to screen models

### Message Types
```go
// Scan messages
ScanProgressMsg    // Progress update during scan
ScanCompleteMsg    // Scan completed successfully
ScanErrorMsg       // Scan failed

// Generation messages
GenerationProgressMsg   // Progress during context generation
GenerationCompleteMsg   // Generation completed
GenerationErrorMsg      // Generation failed

// LLM messages (provider-agnostic)
LLMProgressMsg     // Progress during LLM send
LLMCompleteMsg     // LLM send completed
LLMErrorMsg        // LLM send failed

// Clipboard
ClipboardCompleteMsg    // Clipboard copy done
```

### Step Handling Pattern
1. `handleNextStep()` / `handlePrevStep()` - Navigation
2. `canAdvanceStep()` - Validation before advancing
3. `initStep()` - Initialize step when entering
4. `handleStepInput()` - Process input for current step
5. `getNextStep()` / `getPrevStep()` - Step skipping logic

## Screen Models

### File Selection (`internal/ui/screens/file_selection.go`)
```go
type FileSelectionModel struct {
    tree         *components.FileTreeModel
    filterMode   bool
    filterInput  textinput.Model
    // ...
}
```
- Filter mode: Toggle with `/`, clear with `Ctrl+C`
- Vim navigation: `hjkl` keys
- Directory toggle: `d` key

### File Tree Component (`internal/ui/components/tree.go`)
Key methods:
- `MoveUp()`, `MoveDown()` - Navigation
- `ExpandNode()`, `CollapseNode()` - Tree expansion
- `ToggleSelection()` - File selection
- `SetFilter()`, `ClearFilter()` - Filtering
- `GetSelections()` - Get selected files

### Template Selection (`internal/ui/screens/template_selection.go`)
- Simple list selection
- Shows template source (embedded/user/custom)

### Task/Rules Input (`internal/ui/screens/task_input.go`, `rules_input.go`)
- Text area with Bubble Tea textarea component
- Focus handling with `Esc` toggle

### Review Screen (`internal/ui/screens/review.go`)
- Preview of generated content
- Generation trigger on F8
- LLM send trigger on F9

## Styling (`internal/ui/styles/theme.go`)
- Uses Lip Gloss for styling
- Consistent color scheme across screens
- Selection states: normal, selected, partial, ignored

## Key Bindings
```
Global:
- F8 / Ctrl+PgDn: Next step / Generate
- F10 / Ctrl+PgUp: Previous step
- F1: Toggle help
- Ctrl+Q / Ctrl+C: Quit

File Selection:
- ↑/↓ or k/j: Navigate
- ←/→ or h/l: Collapse/expand
- Space: Toggle selection
- d: Toggle directory selection
- i: Toggle ignored files
- /: Filter mode
- F5: Rescan
```

## TUI Coordinator Pattern

The wizard uses dedicated coordinators for asynchronous operations in `internal/ui/`:

### ScanCoordinator (`scan_coordinator.go`)
Manages file system scanning state machine:
- `Start(rootPath, config)` - Initiates async scan
- `Poll()` - Checks for progress updates
- `Result()` - Returns `(*scanner.FileNode, error)`

### GenerateCoordinator (`generate_coordinator.go`)
Manages context generation state machine:
- `Start(config)` - Initiates async generation
- `Poll()` - Checks for progress updates
- `Result()` - Returns `(string, error)`

### Coordinator Message Flow

1. **Start**: `wizard.Update` calls `coordinator.Start()` → returns `tea.Cmd`
2. **Poll**: `coordinator.Poll()` checks channels and returns `Batch(Msg, NextPoll)`
3. **Progress**: UI receives progress messages
4. **Completion**: Coordinator signals completion, UI retrieves result via `Result()`

## Async Operations Pattern

Use Bubble Tea commands for async work:
```go
func scanDirectoryCmd(root string, config scanner.ScanConfig) tea.Cmd {
    return func() tea.Msg {
        result, err := scanner.Scan(root, config)
        if err != nil {
            return ScanErrorMsg{Error: err}
        }
        return ScanCompleteMsg{Root: result}
    }
}
```

### Iterative Polling Pattern

For long-running operations with progress reporting:

```go
func (m *WizardModel) iterativeScanCmd() tea.Cmd {
    return func() tea.Msg {
        select {
        case progress, ok := <-m.scanState.progressCh:
            if !ok || isComplete(progress) {
                return ScanCompleteMsg{Tree: result}
            }
            return ScanProgressMsg{...}
        case <-m.scanState.done:
            return ScanCompleteMsg{Tree: result}
        default:
            // CRITICAL: Must yield to event loop!
            time.Sleep(10 * time.Millisecond)
            return m.iterativeScanCmd()()
        }
    }
}
```

**Important:** The `default` case MUST include `time.Sleep` to yield to the Bubble Tea event loop.

## Testing Patterns

### CommandRunner Interface

For deterministic testing of external operations:

```go
type MockRunner struct {
    LookPathFunc func(file string) (string, error)
    RunFunc      func(ctx context.Context, name string, args []string, stdin io.Reader) ([]byte, error)
}
```

### Screen Model Test Helpers

```go
// Set selections via screen model
func setWizardSelectedFiles(wizard *WizardModel, selections map[string]bool) {
    if wizard.fileSelection == nil {
        wizard.fileSelection = screens.NewFileSelection(nil, selections)
    } else {
        wizard.fileSelection.SetSelectionsForTest(selections)
    }
}
```

## Configuration TUI (`internal/ui/config_wizard.go`)

Separate TUI for interactive configuration:
- Navigate between configuration categories
- Edit values with real-time validation
- Toggle boolean settings with Space
- Save all changes with Ctrl+S
