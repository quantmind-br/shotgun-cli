# Specification: Visual Selection Feedback for File Tree

> Ingest the information from this file, implement the Low-Level Tasks, and generate the code that will satisfy the High and Mid-Level Objectives.

## High-Level Objective

Transform the TUI file tree component to provide **immediate visual feedback** about file/directory selection states using a color-coded system that distinguishes between unselected (muted blue-gray), partially selected (warning yellow), and fully selected (success green) items.

## Mid-Level Objectives

1. **Define Selection State Type System** - Create a strongly-typed `SelectionState` enum with three states (Unselected, Partial, Selected) and corresponding styled colors integrated into the existing theme system

2. **Implement Selection State Computation** - Add caching mechanism to efficiently compute and store selection states for all visible nodes in the tree, recalculating only when selections or visibility changes

3. **Integrate Visual Feedback into Rendering** - Modify the tree item renderer to apply state-based colors to both checkboxes and file/directory names while maintaining cursor highlight functionality

4. **Ensure Performance at Scale** - Validate that the caching approach maintains O(1) lookup performance during rendering and acceptable O(n) recomputation when selections change

## Implementation Notes

### Color Scheme (Based on Existing Theme)
```go
// From internal/ui/styles/theme.go
PrimaryColor   = "#00ADD8" // Cyan blue - titles and highlight
AccentColor    = "#A3BE8C" // Green - success and progress
WarningColor   = "#EBCB8B" // Yellow - warnings
MutedColor     = "#5C7E8C" // Muted blue-gray - secondary text

// New selection state colors
FileUnselectedColor = "#5C7E8C" // Muted - inactive state
FileSelectedColor   = "#A3BE8C" // Green - complete selection
FilePartialColor    = "#EBCB8B" // Yellow - intermediate state
```

### Technical Constraints
- Must preserve existing keyboard navigation (arrow keys, hjkl vim keys)
- Must maintain cursor highlight behavior (background color override)
- Must respect `showIgnored` and `filter` visibility rules
- Must handle nested directory trees correctly (propagate partial state upward)
- Cannot modify `scanner.FileNode.Selected` field (selections managed by map)

### Performance Targets
- Initial rendering: < 500ms for 1000+ files
- Navigation response: < 50ms per cursor movement
- Selection toggle: < 200ms for directories with 500+ files
- Memory overhead: < 10% increase from selection state cache

### Dependencies
- `github.com/charmbracelet/lipgloss` - Already used for styling
- Existing color theme system in `internal/ui/styles/theme.go`
- Current tree rendering in `internal/ui/components/tree.go`

## Context

### Current State Files

**internal/ui/styles/theme.go** (246 lines)
- Defines color palette and base styles
- Exports `TreeStyle` with foreground color `#ECEFF4`
- Contains helper functions like `RenderFileTree()` (unused in tree.go)

**internal/ui/components/tree.go** (388 lines)
- `FileTreeModel` manages tree state with `selections map[string]bool`
- `renderTreeItem()` (lines 170-240) builds tree prefix, checkbox, name, and applies cursor highlight
- `areAllFilesInDirSelected()` (lines 338-352) already computes "fully selected" logic
- `rebuildVisibleItems()` (lines 242-255) called when filter/expand changes
- `ToggleSelection()` and `ToggleDirectorySelection()` modify selections map

**internal/core/scanner/scanner.go**
- `FileNode.Selected` field exists but **is not used** (selections tracked separately)
- Used only in wizard.go:726 for context generation marking

### Desired State Files

**internal/ui/styles/theme.go** (enhanced)
- Add `SelectionState` type and constants
- Add three color constants for selection states
- Add three styled variants: `UnselectedNameStyle`, `SelectedNameStyle`, `PartialNameStyle`
- Add `RenderFileName(name string, selectionState SelectionState) string` helper

**internal/ui/components/tree.go** (enhanced)
- Add `selectionStates map[string]styles.SelectionState` to `FileTreeModel`
- Add `recomputeSelectionStates()` method - traverses tree post-order to compute states
- Add `selectionStateFor(path string) styles.SelectionState` accessor
- Modify `renderTreeItem()` to use `RenderFileName()` for checkbox and name
- Call `recomputeSelectionStates()` after selection changes and visibility updates

## State Transformation Analysis

### Current Behavior (Problem)
```
📁 src/                    # No color indication of selection state
  ├── [ ] file1.go         # Individual file checkbox only
  ├── [✓] file2.go         # User must expand directories to see
  └── 📁 nested/           # what is selected inside
      └── [✓] file3.go
```

**Issues:**
1. Directories show no visual state (cannot tell if children are selected)
2. Checkbox/name use default tree color regardless of selection
3. Requires expanding all directories to verify selections

### Desired Behavior (Solution)
```
📁 src/                    # YELLOW (partial: 2 of 3 files selected)
  ├── [ ] file1.go         # MUTED (not selected)
  ├── [✓] file2.go         # GREEN (selected)
  └── 📁 nested/           # GREEN (all children selected)
      └── [✓] file3.go     # GREEN (selected)
```

**Benefits:**
1. Immediate visual feedback at directory level
2. Color reinforces checkbox state for files
3. Can identify selection patterns without expanding

### State Computation Logic
```
For each node (post-order traversal):
  If FILE:
    state = selections[path] ? SELECTED : UNSELECTED

  If DIRECTORY:
    hasSelected = any child is SELECTED or PARTIAL
    hasUnselected = any child is UNSELECTED or PARTIAL

    if hasSelected && !hasUnselected:
      state = SELECTED
    else if hasSelected && hasUnselected:
      state = PARTIAL
    else:
      state = UNSELECTED
```

## Low-Level Tasks

> Tasks are ordered by dependency and should be executed sequentially

### Task 1: Add SelectionState Type System to Theme

**Action:** ADD
**File:** `internal/ui/styles/theme.go`
**Location:** After line 52 (after `TreeStyle` definition)

```go
// Add SelectionState type and constants
type SelectionState int

const (
    SelectionUnselected SelectionState = iota
    SelectionPartial
    SelectionSelected
)

// Selection state colors for file tree
var (
    FileUnselectedColor = lipgloss.Color("#5C7E8C") // Muted blue-gray
    FileSelectedColor   = lipgloss.Color("#A3BE8C") // Success green
    FilePartialColor    = lipgloss.Color("#EBCB8B") // Warning yellow

    // Styles for file/directory names based on selection state
    UnselectedNameStyle = lipgloss.NewStyle().
        Foreground(FileUnselectedColor)

    SelectedNameStyle = lipgloss.NewStyle().
        Foreground(FileSelectedColor).
        Bold(true)

    PartialNameStyle = lipgloss.NewStyle().
        Foreground(FilePartialColor).
        Bold(true)
)
```

**Validation:**
```bash
make build
# Expected: Build succeeds with no errors
# Verify: New type and constants are available for import
```

---

### Task 2: Add RenderFileName Helper Function

**Action:** ADD
**File:** `internal/ui/styles/theme.go`
**Location:** After line 246 (end of file, after `RenderFileTree`)

```go
// RenderFileName applies color styling to file/directory names based on selection state
func RenderFileName(name string, selectionState SelectionState) string {
    switch selectionState {
    case SelectionSelected:
        return SelectedNameStyle.Render(name)
    case SelectionPartial:
        return PartialNameStyle.Render(name)
    case SelectionUnselected:
        return UnselectedNameStyle.Render(name)
    default:
        return TreeStyle.Render(name)
    }
}
```

**Details:**
- Takes raw name string and selection state
- Returns lip gloss styled string with appropriate color
- Default case uses existing `TreeStyle` as fallback

**Validation:**
```bash
make build
# Expected: Build succeeds
go test ./internal/ui/styles/...
# Expected: All existing tests pass
```

---

### Task 3: Add Selection State Cache to FileTreeModel

**Action:** MODIFY
**File:** `internal/ui/components/tree.go`
**Function:** `FileTreeModel` struct and `NewFileTree` constructor

**Changes to FileTreeModel struct (line 13):**
```go
type FileTreeModel struct {
    tree         *scanner.FileNode
    cursor       int
    selections   map[string]bool
    selectionStates map[string]styles.SelectionState  // ADD THIS LINE
    showIgnored  bool
    filter       string
    expanded     map[string]bool
    width        int
    height       int
    visibleItems []treeItem
    topIndex     int
}
```

**Changes to NewFileTree function (after line 50):**
```go
func NewFileTree(tree *scanner.FileNode, selections map[string]bool) *FileTreeModel {
    expanded := make(map[string]bool)
    if tree != nil {
        expanded[tree.Path] = true
    }

    model := &FileTreeModel{
        tree:            tree,
        selections:      make(map[string]bool),
        selectionStates: make(map[string]styles.SelectionState), // ADD THIS LINE
        expanded:        expanded,
        showIgnored:     false,
        filter:          "",
    }

    // Copy selections
    for k, v := range selections {
        model.selections[k] = v
    }

    model.rebuildVisibleItems()
    model.recomputeSelectionStates() // ADD THIS LINE
    return model
}
```

**Validation:**
```bash
make build
# Expected: Build succeeds (recomputeSelectionStates not yet defined - next task)
```

---

### Task 4: Implement Selection State Computation

**Action:** ADD
**File:** `internal/ui/components/tree.go`
**Location:** After line 375 (after `walkNode`, before `formatFileSize`)

```go
// recomputeSelectionStates traverses the tree and computes selection state for each node
func (m *FileTreeModel) recomputeSelectionStates() {
    m.selectionStates = make(map[string]styles.SelectionState)
    if m.tree == nil {
        return
    }

    // Post-order traversal to compute states bottom-up
    var visit func(node *scanner.FileNode) styles.SelectionState
    visit = func(node *scanner.FileNode) styles.SelectionState {
        // Skip nodes that shouldn't be shown
        if !m.shouldShowNode(node) {
            return styles.SelectionUnselected
        }

        // Base case: file nodes
        if !node.IsDir {
            state := styles.SelectionUnselected
            if m.selections[node.Path] {
                state = styles.SelectionSelected
            }
            m.selectionStates[node.Path] = state
            return state
        }

        // Recursive case: directory nodes
        hasSelected := false
        hasUnselected := false

        for _, child := range node.Children {
            childState := visit(child)
            switch childState {
            case styles.SelectionSelected:
                hasSelected = true
            case styles.SelectionUnselected:
                hasUnselected = true
            case styles.SelectionPartial:
                // Partial counts as both
                hasSelected = true
                hasUnselected = true
            }
        }

        // Determine directory state
        state := styles.SelectionUnselected
        switch {
        case hasSelected && !hasUnselected:
            state = styles.SelectionSelected
        case hasSelected && hasUnselected:
            state = styles.SelectionPartial
        }

        m.selectionStates[node.Path] = state
        return state
    }

    visit(m.tree)
}

// selectionStateFor returns the cached selection state for a path
func (m *FileTreeModel) selectionStateFor(path string) styles.SelectionState {
    if state, ok := m.selectionStates[path]; ok {
        return state
    }
    return styles.SelectionUnselected
}
```

**Details:**
- Post-order traversal ensures children are computed before parents
- Respects `shouldShowNode()` to honor filter and ignore rules
- Partial state propagates upward (if any child is partial, parent is partial)
- O(n) complexity where n is number of visible nodes

**Validation:**
```bash
make build
# Expected: Build succeeds
go test ./internal/ui/components/...
# Expected: All tests pass
```

---

### Task 5: Trigger Recomputation After Selection Changes

**Action:** MODIFY
**File:** `internal/ui/components/tree.go`
**Functions:** `ToggleSelection`, `ToggleDirectorySelection`, `setDirectorySelection`, `rebuildVisibleItems`

**Modify ToggleSelection (after line 102):**
```go
func (m *FileTreeModel) ToggleSelection() {
    if m.cursor < len(m.visibleItems) {
        item := m.visibleItems[m.cursor]
        if !item.node.IsDir {
            m.selections[item.path] = !m.selections[item.path]
            m.recomputeSelectionStates() // ADD THIS LINE
        }
    }
}
```

**Modify ToggleDirectorySelection (after line 111):**
```go
func (m *FileTreeModel) ToggleDirectorySelection() {
    if m.cursor < len(m.visibleItems) {
        item := m.visibleItems[m.cursor]
        if item.node.IsDir {
            allSelected := m.areAllFilesInDirSelected(item.node)
            m.setDirectorySelection(item.node, !allSelected)
            m.recomputeSelectionStates() // ADD THIS LINE
        }
    }
}
```

**Modify setDirectorySelection (after line 363):**
```go
func (m *FileTreeModel) setDirectorySelection(dir *scanner.FileNode, selected bool) {
    m.walkNode(dir, func(node *scanner.FileNode) {
        if !node.IsDir {
            if selected {
                m.selections[node.Path] = true
            } else {
                delete(m.selections, node.Path)
            }
        }
    })
    m.recomputeSelectionStates() // ADD THIS LINE
}
```

**Modify rebuildVisibleItems (after line 254):**
```go
func (m *FileTreeModel) rebuildVisibleItems() {
    m.visibleItems = nil
    if m.tree != nil {
        m.buildVisibleItems(m.tree, "", 0, true, nil)
    }

    // Adjust cursor if it's out of bounds
    if m.cursor >= len(m.visibleItems) {
        m.cursor = len(m.visibleItems) - 1
    }
    if m.cursor < 0 {
        m.cursor = 0
    }

    m.recomputeSelectionStates() // ADD THIS LINE
}
```

**Validation:**
```bash
make build
# Expected: Build succeeds
# Manual test: Toggle selections and verify state cache updates
```

---

### Task 6: Apply Color to File/Directory Names in Renderer

**Action:** MODIFY
**File:** `internal/ui/components/tree.go`
**Function:** `renderTreeItem` (lines 170-240)

**Add state lookup after prefix construction (after line 189):**
```go
func (m *FileTreeModel) renderTreeItem(item treeItem, isCursor bool) string {
    var prefix strings.Builder

    // ... existing prefix construction code (lines 173-189) ...

    // ADD: Determine selection state for styling
    selectionState := m.selectionStateFor(item.path)

    // ... rest of function continues ...
}
```

**Modify checkbox section (replace lines 192-199):**
```go
// ORIGINAL CODE:
// var checkbox string
// if !item.node.IsDir {
//     if m.selections[item.path] {
//         checkbox = "[✓] "
//     } else {
//         checkbox = "[ ] "
//     }
// }

// REPLACE WITH:
var checkbox string
if !item.node.IsDir {
    checkboxText := "[ ] "
    if m.selections[item.path] {
        checkboxText = "[✓] "
    }
    // Apply color based on selection state
    checkbox = styles.RenderFileName(checkboxText, selectionState)
}
```

**Modify name section (replace lines 211-215):**
```go
// ORIGINAL CODE:
// name := filepath.Base(item.path)
// if item.node.IsDir {
//     name += "/"
// }

// REPLACE WITH:
baseName := filepath.Base(item.path)
if item.node.IsDir {
    baseName += "/"
}
// Apply color based on selection state
name := styles.RenderFileName(baseName, selectionState)
```

**Keep cursor highlighting unchanged (lines 234-237):**
```go
// PRESERVE EXISTING CODE:
if isCursor {
    line = styles.SelectedStyle.Render(line)
}
```

**Validation:**
```bash
make build
./build/shotgun-cli
# Manual test:
# 1. Navigate to a directory with files
# 2. Select individual files - verify checkbox and name turn green
# 3. Select directory - verify all files green, directory green
# 4. Partially select directory - verify yellow color
# 5. Verify cursor highlight (background) still works over colors
```

---

### Task 7: Manual Testing - Basic Selection States

**Action:** VALIDATE
**Verification Steps:**

```bash
# Test 1: Individual file selection
./build/shotgun-cli
# Navigate to any file, press Space
# Expected: Checkbox and name change from muted (#5C7E8C) to green (#A3BE8C)

# Test 2: Directory full selection
# Navigate to a directory, press 'd'
# Expected: Directory name becomes green and bold
# Expected: All child files have green checkboxes and names

# Test 3: Directory partial selection
# Expand directory with 3+ files
# Select only first file with Space
# Collapse directory
# Expected: Directory name becomes yellow (#EBCB8B) and bold

# Test 4: Nested directories
# Create structure: parent/child1/file1.txt, parent/child2/file2.txt
# Select only file1.txt
# Expected: child1 is green, child2 is muted, parent is yellow

# Test 5: State transitions
# Start with empty selection (all muted)
# Select 1 file in directory (directory turns yellow)
# Select all files in directory (directory turns green)
# Deselect 1 file (directory turns yellow)
# Deselect all files (directory turns muted)
```

**Success Criteria:**
- All color transitions occur immediately
- Colors match specification (muted/yellow/green)
- Cursor highlight overrides colors correctly
- No visual glitches or flicker

---

### Task 8: Manual Testing - Integration with Existing Features

**Action:** VALIDATE
**Verification Steps:**

```bash
# Test 6: Filter interaction
./build/shotgun-cli
# Select some files
# Press '/' and filter (e.g., ".go")
# Expected: Colors reflect only visible filtered files
# Clear filter (Ctrl+C)
# Expected: Colors restore to reflect all files

# Test 7: Ignored files toggle
# Press 'i' to show ignored files
# Expected: Ignored files show in muted color (default unselected)
# Select an ignored file
# Expected: Ignored file turns green
# Press 'i' to hide ignored again
# Expected: Parent directory color accounts for hidden ignored selections

# Test 8: Cursor navigation
# Use arrow keys and hjkl to navigate
# Expected: Navigation works smoothly
# Expected: Response time < 50ms per keystroke

# Test 9: All keyboard shortcuts work
# F5 (rescan), F8 (next step), Space (toggle), d (dir toggle), i (ignored)
# Expected: All shortcuts function identically to before changes
```

**Success Criteria:**
- Filter updates colors correctly
- Ignored files interact properly with color system
- Navigation feels responsive
- All existing shortcuts work

---

### Task 9: Performance Testing at Scale

**Action:** VALIDATE
**Performance Benchmarks:**

```bash
# Setup: Clone large repository for testing
cd /tmp
git clone https://github.com/golang/go go-test
cd /path/to/shotgun-cli

# Test P1: Initial load (1000+ files)
time ./build/shotgun-cli
# Navigate to /tmp/go-test in TUI
# Expected: Initial tree render < 500ms
# Expected: UI remains responsive

# Test P2: Navigation responsiveness
# Navigate with arrow keys rapidly (10+ moves)
# Expected: Each move response < 50ms
# Expected: No perceptible lag

# Test P3: Directory selection (500+ files)
# Navigate to large directory (e.g., src/ in Go repo)
# Press 'd' to toggle directory selection
# Expected: Response time < 200ms for 500 files
# Expected: Response time < 1s for 5000 files

# Test P4: Memory usage
# Compare before/after implementation
pmap $(pgrep shotgun-cli) | grep total
# Expected: Memory increase < 10% from baseline
```

**Success Criteria:**
- All timing targets met
- No memory leaks detected
- Smooth UX with large repositories

---

### Task 10: Regression Testing

**Action:** VALIDATE
**Regression Checklist:**

```bash
# R1: File selection functionality
# [ ] Space toggles file selection
# [ ] Checkboxes appear correctly
# [ ] Selected files tracked in map

# R2: Directory selection functionality
# [ ] 'd' key toggles directory selection
# [ ] All files in directory toggle together
# [ ] Nested directories handled correctly

# R3: Navigation
# [ ] Arrow keys (up/down/left/right) work
# [ ] Vim keys (h/j/k/l) work
# [ ] Expand/collapse directories works
# [ ] Scroll behavior correct

# R4: Filter and visibility
# [ ] '/' enters filter mode
# [ ] Filter applies correctly
# [ ] Ctrl+C clears filter
# [ ] 'i' toggles ignored files
# [ ] F5 rescans directory

# R5: Other wizard steps
# [ ] Template selection works
# [ ] Task input works
# [ ] Rules input works
# [ ] Review screen works
# [ ] Context generation completes

# R6: Visual elements
# [ ] Tree structure characters (│├└) display correctly
# [ ] Directory emojis (📁📂) display correctly
# [ ] File size info displays
# [ ] Ignore status markers (g/c) display
```

**Success Criteria:**
- All checkboxes pass
- No regressions detected
- User experience equivalent or better

---

### Task 11: Documentation and Finalization

**Action:** MODIFY
**File:** `plans/VISUAL_SELECTION_FEEDBACK.md`
**Changes:** Add "Implementation Status" section

```markdown
## Implementation Status

✅ **Completed** - 2025-10-02

### Changes Made
1. Added `SelectionState` type system to `internal/ui/styles/theme.go`
2. Implemented `RenderFileName()` helper with color styling
3. Added `selectionStates` cache to `FileTreeModel`
4. Implemented `recomputeSelectionStates()` with post-order traversal
5. Modified `renderTreeItem()` to apply colors to checkboxes and names
6. Triggered recomputation on selection and visibility changes

### Testing Results
- ✅ All manual tests passed
- ✅ Performance targets met
- ✅ No regressions detected

### Performance Metrics
- Initial render (1000 files): XXX ms
- Navigation response: XX ms
- Directory toggle (500 files): XX ms
- Memory overhead: X%
```

**Action:** MODIFY
**File:** `CLAUDE.md` (if needed)
**Changes:** Document new color system for future reference

**Validation:**
```bash
git add .
git commit -m "feat: implement visual selection feedback with color states

- add SelectionState type (Unselected/Partial/Selected) to theme
- implement cached selection state computation via post-order traversal
- colorize tree checkboxes and names based on selection state
- colors: unselected (#5C7E8C), partial (#EBCB8B), selected (#A3BE8C)
- maintain O(1) render lookup, O(n) recomputation on changes
- all tests passed, no performance regressions"

# Expected: Clean commit with descriptive message
```

---

## Risk Assessment and Mitigations

### Risk 1: Performance Degradation on Large Trees
**Probability:** Low
**Impact:** Medium
**Mitigation:**
- Cache-based design ensures O(1) lookup during render
- Recomputation is O(n) but only triggered on state changes
- Task 9 validates performance targets explicitly

### Risk 2: Color Visibility in Different Terminals
**Probability:** Medium
**Impact:** Medium
**Mitigation:**
- Colors chosen from existing theme (already tested)
- Use bold for selected/partial states to increase contrast
- Test in common terminals: iTerm2, Alacritty, GNOME Terminal, Windows Terminal

### Risk 3: Cursor Highlight Visual Conflicts
**Probability:** Low
**Impact:** Low
**Mitigation:**
- Cursor uses background color, which overrides foreground
- Task 7 explicitly tests cursor interaction
- Existing behavior preserved

### Risk 4: Filter/Ignored State Bugs
**Probability:** Low
**Impact:** Medium
**Mitigation:**
- `recomputeSelectionStates()` respects `shouldShowNode()` filter
- Task 8 tests filter and ignored interactions explicitly
- Existing visibility logic unchanged

## Rollback Plan

If critical issues are discovered after implementation:

```bash
# Immediate rollback
git revert HEAD
make build

# Alternative: Feature flag (future enhancement)
# Add to config.yaml:
# ui:
#   enable_selection_colors: false
```

## Future Enhancements

1. **Short-term:** Add legend in footer showing color meanings
2. **Mid-term:** Make colors configurable via config.yaml
3. **Long-term:** Support high-contrast accessibility mode

## Success Criteria

- [ ] All 11 tasks completed successfully
- [ ] All manual tests pass (Tasks 7-8)
- [ ] Performance targets met (Task 9)
- [ ] No regressions detected (Task 10)
- [ ] Code committed with descriptive message (Task 11)
- [ ] Visual feedback provides clear, immediate selection state indication

---

**Document Version:** 1.0
**Created:** 2025-10-02
**Status:** Ready for Implementation
**Estimated Implementation Time:** 3-4 hours
