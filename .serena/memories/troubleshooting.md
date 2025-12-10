# shotgun-cli Troubleshooting Guide

## Known Issues and Solutions

### TUI Freezes on Large Directories

**Symptom:** Application appears frozen when opening directories with many files.

**Cause:** The `iterativeScanCmd()` function in `internal/ui/wizard.go` has a recursive polling pattern. If the `default` case doesn't yield to the event loop, it creates a busy-loop.

**Solution (Fixed in commit 50bad98):**
```go
default:
    // MUST yield to event loop
    time.Sleep(10 * time.Millisecond)
    return m.iterativeScanCmd()()
```

**Related Files:**
- `internal/ui/wizard.go:872-940` - iterativeScanCmd()
- `internal/core/scanner/filesystem.go` - ScanWithProgress()

### No Progress During Initial Scan

**Symptom:** Progress indicator shows 0/0 or nothing during the initial phase.

**Cause:** The `countItems()` function (first pass) doesn't send progress updates - only `walkAndBuild()` (second pass) does.

**Solution (Fixed in commit 50bad98):**
Send a "counting" stage progress message before `countItems()`:
```go
if progress != nil {
    progress <- Progress{Stage: "counting", Message: "Counting files..."}
}
```

## Debugging Tips

### Enable Debug Logging
```bash
SHOTGUN_LOG_LEVEL=debug shotgun-cli
```

### Test with Specific Directory
```bash
cd /path/to/large/directory
shotgun-cli
```

### Run with Race Detector (Development)
```bash
make test-race
```

## Common Development Issues

### Tests Fail After Changing Progress Stages

If you change the order or names of progress stages, update the test in:
- `internal/core/scanner/scanner_test.go:402-405`

Example assertion:
```go
if first.Stage != "counting" {  // Was "scanning" before the fix
    t.Errorf("Expected first stage to be 'counting', got %q", first.Stage)
}
```

### TUI Doesn't Render During Async Operations

Check that async commands:
1. Return `tea.Msg` types properly
2. Don't block the main goroutine
3. Use channels with adequate buffer size (e.g., `make(chan Progress, 100)`)

### File Selection Not Working

Verify:
- `selectedFiles` map is initialized in `NewWizard()`
- FileTreeModel properly propagates selection changes
- `GetSelections()` returns correct paths
