# shotgun-cli Troubleshooting Guide

## Known Issues and Solutions

### TUI Freezes on Large Directories

**Symptom:** Application appears frozen when opening directories with many files.

**Cause:** The `iterativeScanCmd()` function in `internal/ui/wizard.go` has a recursive polling pattern. If the `default` case doesn't yield to the event loop, it creates a busy-loop.

**Solution:**
```go
default:
    // MUST yield to event loop
    time.Sleep(10 * time.Millisecond)
    return m.iterativeScanCmd()()
```

**Related Files:**
- `internal/ui/wizard.go` - iterativeScanCmd()
- `internal/ui/scan_coordinator.go` - ScanCoordinator
- `internal/core/scanner/filesystem.go` - ScanWithProgress()

### No Progress During Initial Scan

**Symptom:** Progress indicator shows 0/0 or nothing during the initial phase.

**Cause:** The `countItems()` function (first pass) doesn't send progress updates - only `walkAndBuild()` (second pass) does.

**Solution:**
Send a "counting" stage progress message before `countItems()`:
```go
if progress != nil {
    progress <- Progress{Stage: "counting", Message: "Counting files..."}
}
```

## LLM Provider Issues

### "Provider not configured" Error

**Diagnosis:**
```bash
shotgun-cli llm status
shotgun-cli llm doctor
```

**Common Causes:**
1. API key not set
2. Wrong provider name
3. Invalid model name

**Solution:**
```bash
shotgun-cli config set llm.provider openai
shotgun-cli config set llm.api-key YOUR_API_KEY
shotgun-cli config set llm.model gpt-4o
```

### API Request Timeout

**Symptom:** LLM send fails with timeout error.

**Solution:** Increase timeout (default is 300 seconds):
```bash
shotgun-cli config set llm.timeout 600
```

### Custom Endpoint Not Working

**Symptom:** Using OpenRouter or Azure but getting connection errors.

**Checklist:**
1. Verify base URL includes full path (e.g., `https://openrouter.ai/api/v1`)
2. Check API key is valid for that endpoint
3. Ensure model name matches provider's naming convention

## Debugging Tips

### Enable Debug Logging
```bash
SHOTGUN_LOG_LEVEL=debug shotgun-cli
```

### Test with Specific Directory
```bash
cd /path/to/directory
shotgun-cli
```

### Run with Race Detector (Development)
```bash
make test-race
```

## Common Development Issues

### Tests Fail After Changing Progress Stages

If you change the order or names of progress stages, update tests in:
- `internal/core/scanner/scanner_test.go`
- `internal/ui/scan_coordinator_test.go`

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

### Config Validation Errors

Check `internal/config/validator.go` for validation rules:
- Size format: Must be like `10MB`, `500KB`, not just numbers for size fields
- Boolean: Must be `true` or `false` (not `yes`/`no`/`1`/`0`)
- Provider: Must be `openai`, `anthropic`, or `gemini`

### Provider Registration Fails

Ensure provider is registered in `internal/app/providers.go`:
```go
DefaultProviderRegistry.Register(llm.ProviderXXX, func(cfg llm.Config) (llm.Provider, error) {
    return xxx.NewClient(cfg)
})
```

## Coverage Issues

### Coverage Below 85%

The CI enforces 85% minimum coverage. To check:
```bash
make coverage
go tool cover -func=coverage.out | grep total
```

Focus on:
- Core logic in `internal/core/`
- Application layer in `internal/app/`
- New code paths

### Coverage HTML Report
```bash
go tool cover -html=coverage.out -o coverage.html
```

## Build Issues

### golangci-lint Failures

The project uses `.golangci-local.yml` for local development:
```bash
make lint
```

Key limits:
- Max line length: 120 chars
- Max cyclomatic complexity: 25

### Module Verification Failures

```bash
go mod tidy
go mod verify
```

## Terminal Size Issues

**Symptom:** Wizard shows "Terminal too small" warning.

**Minimum size:** 40 columns x 10 rows

Check constants in `internal/ui/wizard.go`:
- `minTerminalWidth`
- `minTerminalHeight`
