# shotgun-cli UI Patterns (Bubble Tea)

## Wizard Architecture

### Main Wizard Model (`internal/ui/wizard.go`)
- `WizardModel` orchestrates all 5 steps
- Steps defined as constants: `StepFileSelection`, `StepTemplateSelection`, `StepTaskInput`, `StepRulesInput`, `StepReview`
- Handles navigation between steps and async operations

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

// Clipboard
ClipboardCompleteMsg    // Clipboard copy done

// Gemini integration
GeminiSendMsg, GeminiProgressMsg, GeminiCompleteMsg, GeminiErrorMsg
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
```go
type FileTreeModel struct {
    root        *scanner.FileNode
    selections  map[string]bool
    expanded    map[string]bool
    showIgnored bool
    filter      string
    // ...
}
```
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

## Async Operations Pattern

Use Bubble Tea commands for async work:
```go
func scanDirectoryCmd(root string, config scanner.ScanConfig) tea.Cmd {
    return func() tea.Msg {
        // Do async work
        result, err := scanner.Scan(root, config)
        if err != nil {
            return ScanErrorMsg{Error: err}
        }
        return ScanCompleteMsg{Root: result}
    }
}
```

### Iterative Polling Pattern (Scan/Generate)

For long-running operations with progress reporting, use iterative polling:

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

**Important:** The `default` case MUST include a `time.Sleep` to yield to the Bubble Tea event loop. Without this, the recursive call creates a busy-loop that blocks the UI, causing the application to appear frozen on large directories.

### Progress Stages

Scanner progress uses stages to indicate current phase:
- `"counting"` - First pass counting files (no progress numbers yet)
- `"scanning"` - Second pass building file tree (with progress)
- `"complete"` - Scan finished
