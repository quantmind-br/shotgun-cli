name: "Hierarchical File Selection System Refactor"
description: |

## Purpose

Context-rich PRP for refactoring shotgun-cli's file selection system from individual file tracking to an efficient hierarchical model that dramatically improves performance and reduces memory usage for large directory trees.

## Core Principles

1. **Performance First**: Target O(1) directory operations vs current O(n)
2. **Memory Efficiency**: Store only explicit user decisions, not computed states  
3. **UI Responsiveness**: Sub-millisecond View() rendering in BubbleTea
4. **Cross-Platform**: Robust Windows/Unix path handling

---

## Goal

Replace the current flat `map[string]bool` selection system with a hierarchical model that tracks only explicit user exclusions and computes child states through parent inheritance. Achieve 50-90% memory reduction and 60-80% performance improvement for large directory operations.

## Why

- **Performance Crisis**: Current system requires O(n) operations to exclude directories with thousands of files
- **Memory Waste**: Excluding `node_modules/` (10K files) creates 10K individual map entries  
- **UI Lag**: Directory statistics recalculated on every render causing 100-200ms delays
- **Scalability Limits**: Large repositories (10K+ files) become unusable due to recursive UI calculations

## What

Transform file selection from individual file tracking to hierarchical parent-child inheritance:

**Current State**: `SelectionState.excluded["src/components/Button.tsx"] = true` (1000+ entries for directory)
**Target State**: `SelectionState.selection["src/components"] = StatusExcluded` (1 entry for directory)

### Success Criteria

- [ ] Directory exclusion operations complete in <1ms (currently 50-100ms)
- [ ] Memory usage reduced 50-90% for hierarchical exclusions
- [ ] UI View() rendering <10ms for large trees (currently 100-200ms)
- [ ] All existing functionality preserved (keyboard shortcuts, visual indicators)
- [ ] Cross-platform path handling without regressions
- [ ] Thread-safe operations maintained

## All Needed Context

### Documentation & References

```yaml
# MUST READ - Critical for implementation success
- url: https://pkg.go.dev/path/filepath
  why: Cross-platform path operations, Clean, Join, Dir functions for hierarchical path checking
  
- url: https://github.com/charmbracelet/bubbletea
  why: BubbleTea performance patterns, avoiding expensive operations in View() method

- url: https://github.com/julienschmidt/httprouter  
  why: Radix tree implementation patterns for efficient path hierarchies
  critical: Zero-allocation lookup patterns for performance

- url: https://github.com/tchap/go-patricia
  why: Patricia trie algorithms for hierarchical path storage and prefix matching

- file: internal/core/types.go
  why: Current SelectionState implementation and thread-safety patterns
  critical: Mutex usage patterns and method signatures to preserve

- file: internal/ui/filetree.go  
  why: BubbleTea integration patterns and getStatusIndicator performance bottlenecks
  critical: Avoid recursive operations in View() method

- file: internal/core/scanner.go
  why: Integration points with file tree traversal and exclusion checking
  critical: walkFileTree function modification patterns
```

### Current Codebase Structure

```bash
shotgun-cli/
├── internal/
│   ├── core/
│   │   ├── types.go              # SelectionState with flat map[string]bool
│   │   ├── scanner.go            # File tree traversal with exclusion checking  
│   │   ├── template.go           # Template processing (unrelated)
│   │   └── types_test.go         # Basic unit tests
│   └── ui/
│       ├── filetree.go          # BubbleTea model with recursive UI operations
│       ├── app.go               # Main application flow
│       └── views.go             # View rendering logic
├── templates/                   # Prompt templates (unrelated)
└── main.go                     # Entry point
```

### Target Codebase Structure

```bash
shotgun-cli/
├── internal/
│   ├── core/
│   │   ├── types.go              # MODIFIED: Hierarchical SelectionState
│   │   ├── scanner.go            # MODIFIED: Optimized exclusion checking
│   │   ├── types_test.go         # ENHANCED: Hierarchical tests
│   │   └── selection_test.go     # NEW: Comprehensive hierarchical tests
│   └── ui/
│       ├── filetree.go          # MODIFIED: Cached statistics, optimized View()
│       └── filetree_test.go     # NEW: UI component tests
```

### Known Gotchas & Library Quirks

```go
// CRITICAL: BubbleTea View() called frequently - avoid expensive operations
func (m FileTreeModel) View() string {
    // ❌ BAD: O(n) operations on every render
    stats := m.getStatsRecursive(m.root)  // Current implementation
    
    // ✅ GOOD: Cached statistics updated only on state change
    stats := m.cachedStats  // Target implementation
}

// CRITICAL: Windows path handling requires filepath.Clean and filepath.Join
// ❌ BAD: strings.HasPrefix("C:\\path", "C:/path") - fails on Windows
// ✅ GOOD: filepath.Clean + proper separator handling

// CRITICAL: Thread safety with RWMutex - match existing patterns
// Current: ss.mu.RLock() for reads, ss.mu.Lock() for writes

// CRITICAL: BubbleTea model immutability - return new instances on updates
// Don't mutate existing model state directly
```

## Implementation Blueprint

### Data Models and Structure

Transform core selection state to hierarchical model with parent-child inheritance:

```go
// Target enum for selection state
type SelectionStatus int

const (
    StatusInherit SelectionStatus = iota  // Inherit from parent (default)
    StatusExcluded                        // Explicitly excluded
)

// Target hierarchical selection state  
type SelectionState struct {
    selection map[string]SelectionStatus  // Only explicit decisions
    mu        sync.RWMutex               // Thread safety (preserve existing)
}
```

### Task Implementation Order

```yaml
Task 1 - Core Data Structure Refactor:
MODIFY internal/core/types.go:
  - FIND pattern: "type SelectionState struct"
  - REPLACE excluded map[string]bool with selection map[string]SelectionStatus  
  - ADD SelectionStatus enum (StatusInherit, StatusExcluded)
  - PRESERVE mutex and thread-safety patterns
  - REMOVE unused fields: included, patterns

CREATE internal/core/types.go - IsPathExcluded method:
  - IMPLEMENT hierarchical parent checking using filepath.Dir
  - PATTERN: Walk parent paths using filepath.Dir until root
  - CRITICAL: Use filepath.Clean for cross-platform normalization
  - THREAD-SAFE: Use RLock for read operations

Task 2 - Update Existing Methods:
MODIFY internal/core/types.go - ToggleFile method:
  - FIND pattern: "func (ss *SelectionState) ToggleFile"
  - SIMPLIFY: Only toggle specific path, remove recursive children logic
  - PRESERVE: Thread safety with mutex locking
  - LOGIC: StatusInherit -> StatusExcluded -> StatusInherit

MODIFY internal/core/types.go - IsFileIncluded method:
  - REPLACE: Direct map lookup with hierarchical IsPathExcluded call
  - MAINTAIN: Same method signature for backward compatibility
  - THREAD-SAFE: Use RLock through IsPathExcluded

Task 3 - Scanner Integration:
MODIFY internal/core/scanner.go - walkFileTree function:
  - FIND pattern: "selection.IsFileIncluded(node.RelPath)"
  - OPTIMIZE: Skip directory traversal if parent excluded
  - PATTERN: Check IsPathExcluded before recursing into directories
  - PRESERVE: Three-layer filtering logic (gitignore, custom, user)

Task 4 - UI Performance Optimization:
MODIFY internal/ui/filetree.go - Remove recursive functions:
  - DELETE: toggleDirectoryChildren and toggleDirectoryChildrenRecursive
  - DELETE: getDirectoryStats function
  - SIMPLIFY: Space key toggle calls only m.selection.ToggleFile(node.RelPath)

ADD internal/ui/filetree.go - Directory statistics caching:
  - CREATE: cachedStats field in FileTreeModel
  - CREATE: invalidateStatsCache method called on selection changes
  - MODIFY: getStatusIndicator to use cached statistics
  - PATTERN: Only recalculate when selection state actually changes

Task 5 - Enhanced Testing:
CREATE internal/core/selection_test.go:
  - TEST: Hierarchical exclusion inheritance
  - TEST: Parent-child relationship logic  
  - TEST: Cross-platform path handling
  - TEST: Thread safety with concurrent operations
  - BENCHMARK: Performance comparison vs old implementation

CREATE internal/ui/filetree_test.go:
  - TEST: UI rendering with hierarchical state
  - TEST: Keyboard interaction behavior
  - TEST: Visual status indicators accuracy
  - BENCHMARK: View() rendering performance
```

### Per Task Pseudocode

```go
// Task 1 - Core hierarchical checking algorithm
func (ss *SelectionState) IsPathExcluded(path string) bool {
    ss.mu.RLock()
    defer ss.mu.RUnlock()
    
    // Clean path for cross-platform compatibility
    cleanPath := filepath.Clean(path)
    
    // Check each parent level for exclusions
    currentPath := cleanPath
    for {
        // Check explicit exclusion at this level
        if status, exists := ss.selection[currentPath]; exists {
            return status == StatusExcluded
        }
        
        // Move to parent directory
        parentPath := filepath.Dir(currentPath)
        if parentPath == currentPath {
            // Reached root without finding exclusion
            break
        }
        currentPath = parentPath
    }
    
    // Default: included (no exclusions found)
    return false
}

// Task 2 - Simplified toggle logic
func (ss *SelectionState) ToggleFile(path string) {
    ss.mu.Lock()
    defer ss.mu.Unlock()
    
    cleanPath := filepath.Clean(path)
    
    // Simple state cycling: inherit -> excluded -> inherit
    current := ss.selection[cleanPath]  // Zero value is StatusInherit
    if current == StatusInherit {
        ss.selection[cleanPath] = StatusExcluded
    } else {
        delete(ss.selection, cleanPath)  // Return to inherit state
    }
}

// Task 4 - Cached directory statistics
type FileTreeModel struct {
    // ... existing fields ...
    cachedStats      map[string]DirectoryStats  // NEW: Cache stats by path
    statsCacheValid  bool                       // NEW: Cache validity flag
}

func (m *FileTreeModel) invalidateStatsCache() {
    m.statsCacheValid = false
    m.cachedStats = make(map[string]DirectoryStats)
}

func (m *FileTreeModel) getDirectoryStatsEfficient(dir *core.FileNode) (included, excluded int) {
    if m.statsCacheValid {
        if stats, exists := m.cachedStats[dir.RelPath]; exists {
            return stats.Included, stats.Excluded
        }
    }
    
    // Calculate once, cache result
    included, excluded = m.calculateDirectoryStats(dir)
    if m.cachedStats == nil {
        m.cachedStats = make(map[string]DirectoryStats)
    }
    m.cachedStats[dir.RelPath] = DirectoryStats{Included: included, Excluded: excluded}
    
    return included, excluded
}
```

### Integration Points

```yaml
THREAD_SAFETY:
  - pattern: "Preserve existing RWMutex usage in SelectionState"
  - maintain: "ss.mu.RLock() for reads, ss.mu.Lock() for writes"

BUBBLETEA_INTEGRATION:
  - modify: "Update() method to call invalidateStatsCache on state changes"
  - optimize: "View() method to use cached statistics instead of recursive calculation"

CROSS_PLATFORM:
  - add: "filepath.Clean() calls for path normalization"
  - use: "filepath.Dir() and filepath.Join() for path operations"
  - avoid: "Hard-coded path separators or string manipulation"

BACKWARD_COMPATIBILITY:
  - preserve: "All existing method signatures (IsFileIncluded, ToggleFile, etc.)"
  - maintain: "Same keyboard shortcuts and UI behavior"
  - keep: "Export formats and external interfaces unchanged"
```

## Validation Loop

### Level 1: Syntax & Style

```bash
# Run these FIRST - fix any errors before proceeding
go fmt ./internal/core/types.go      # Format code
go vet ./internal/core/types.go      # Static analysis
go build ./...                       # Compilation check

# Expected: No errors. If errors, READ the error and fix.
```

### Level 2: Unit Tests

```go
// CREATE internal/core/selection_test.go with these critical test cases:
func TestHierarchicalExclusion(t *testing.T) {
    ss := NewSelectionState()
    
    // Exclude parent directory
    ss.selection["src/components"] = StatusExcluded
    
    // Child should be excluded via inheritance
    if !ss.IsPathExcluded("src/components/Button.tsx") {
        t.Error("Child file should inherit exclusion from parent directory")
    }
}

func TestToggleDirectoryBehavior(t *testing.T) {
    ss := NewSelectionState()
    
    // Toggle directory should affect only that path
    ss.ToggleFile("src/components")
    if ss.selection["src/components"] != StatusExcluded {
        t.Error("Directory should be explicitly excluded after toggle")
    }
    
    // Toggle again should return to inherit
    ss.ToggleFile("src/components")
    if _, exists := ss.selection["src/components"]; exists {
        t.Error("Directory should return to inherit state after second toggle")
    }
}

func TestCrossPlatformPaths(t *testing.T) {
    ss := NewSelectionState() 
    
    // Test Windows-style path handling
    ss.selection["src\\components"] = StatusExcluded
    if !ss.IsPathExcluded("src/components/Button.tsx") {
        t.Error("Should handle mixed path separators correctly")
    }
}

func TestConcurrentAccess(t *testing.T) {
    ss := NewSelectionState()
    
    // Test thread safety with goroutines
    done := make(chan bool, 2)
    
    go func() {
        for i := 0; i < 1000; i++ {
            ss.ToggleFile(fmt.Sprintf("path%d", i))
        }
        done <- true
    }()
    
    go func() {
        for i := 0; i < 1000; i++ {
            _ = ss.IsPathExcluded(fmt.Sprintf("path%d", i))
        }
        done <- true
    }()
    
    <-done
    <-done
    // If we reach here without panic, thread safety works
}
```

```bash
# Run and iterate until passing:
go test ./internal/core -v
# If failing: Read error, understand root cause, fix code, re-run
```

### Level 3: Performance Benchmarks

```bash
# Create and run performance benchmarks
go test -bench=. ./internal/core -benchmem

# Expected performance improvements:
# - Directory toggle: >90% faster (50ms -> <1ms)
# - Memory usage: 50-90% reduction for hierarchical exclusions
# - IsPathExcluded: O(log d) where d = directory depth
```

```go
// CREATE performance benchmarks in selection_test.go
func BenchmarkHierarchicalVsFlat(b *testing.B) {
    // Compare old flat approach vs new hierarchical
    // Measure directory exclusion performance
    // Measure memory allocations
}

func BenchmarkUIRendering(b *testing.B) {
    // Benchmark FileTreeModel.View() performance
    // Test with large file trees (1000+ files)
    // Measure memory allocations per render
}
```

### Level 4: Integration & UI Testing

```bash
# Test the full application functionality
go run . # Start shotgun-cli

# Manual test cases:
# 1. Navigate to a large directory (node_modules, vendor)
# 2. Press 'space' to exclude - should be instant (<1ms)
# 3. Check visual indicators - directory should show [x]
# 4. Navigate to child files - should show [x] (inherited)
# 5. Toggle child file - should work independently
# 6. Check memory usage with large exclusions

# Expected: Responsive UI, correct visual feedback, memory efficiency
```

### Level 5: Cross-Platform Validation

```bash
# Test on different platforms for path handling
# Windows: Check backslash vs forward slash handling
# Unix: Check path normalization
# Test with paths containing spaces, special characters

# Validate with npm commands:
npm run dev     # Development build
npm test        # Full test suite  
npm run lint    # Go vet validation
npm run format  # Go fmt validation
```

## Final Validation Checklist

- [ ] All tests pass: `go test ./... -v`
- [ ] No linting errors: `go vet ./...`
- [ ] Code formatted: `go fmt ./...`
- [ ] Performance benchmarks show >50% improvement
- [ ] Memory usage reduced for hierarchical exclusions
- [ ] UI remains responsive (<10ms View() rendering)
- [ ] Cross-platform path handling works correctly
- [ ] Thread safety maintained under concurrent access
- [ ] All existing keyboard shortcuts work unchanged
- [ ] Visual indicators ([ ], [x], [~]) display correctly

---

## Anti-Patterns to Avoid

- ❌ Don't add expensive operations to BubbleTea View() method
- ❌ Don't use string manipulation for path operations (use filepath package)
- ❌ Don't break thread safety - preserve existing mutex patterns
- ❌ Don't populate hierarchical state recursively - store only explicit decisions
- ❌ Don't ignore cross-platform path normalization
- ❌ Don't sacrifice UI responsiveness for feature completeness
- ❌ Don't change method signatures - maintain backward compatibility

## Expected Performance Gains

**Directory Operations**: O(n) → O(1) (90%+ improvement)  
**Memory Usage**: 50-90% reduction for hierarchical exclusions  
**UI Rendering**: 60-80% faster for large trees  
**User Experience**: Sub-millisecond operations vs 50-100ms delays

---

**PRP Confidence Score: 9/10**

This PRP provides comprehensive context from deep codebase analysis, proven algorithmic approaches from research, detailed implementation steps, and robust validation gates. The one-pass implementation success probability is very high due to the rich context and specific technical guidance provided.