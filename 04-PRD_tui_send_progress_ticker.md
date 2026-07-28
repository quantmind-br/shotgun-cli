# PRD — TUI Send Progress Ticker

**Source**: IDEATION_CODE_IMPROVEMENTS.md
**Generated**: 2026-07-28

## Implementation Order

1. CI-002 — Drive the Review screen's elapsed-time counter with a repeating 1-second tick, and stop resetting the send start time on progress messages

---

## CI-002: Drive the Review Screen's Elapsed-Time Counter During an LLM Send

### Scope

**In scope**:

- Add an `llmTickMsg` message type and an `llmTick()` command helper to `internal/ui/wizard.go`.
- Emit the first tick from `handleSendToLLM` alongside the existing send command.
- Re-arm the tick from `WizardModel.Update` for as long as `m.llmSending` is true, and let it lapse once it is false.
- Delete the `m.llmStartTime = time.Now()` assignment from the `LLMProgressMsg` arm of `ReviewModel.HandleMessage` (`internal/ui/screens/review.go:147-150`).
- Add wizard tests for tick emission, tick re-arming while sending, and tick termination after completion.
- Add a Review-screen test asserting `LLMProgressMsg` does not move `llmStartTime`.

**Out of scope**:

- **Forwarding the provider progress callback.** `internal/ui/wizard.go:927` keeps passing `nil` to `SendToLLMWithProgress`. `llmbase.BaseClient.SendWithProgress` emits exactly two stages — `"Connecting to <provider>..."` at the start and `"Response received"` immediately before returning — so there is no mid-flight signal worth a channel. A `tea.Cmd` closure also cannot deliver more than one message, so forwarding would require a full LLM coordinator; two stages do not justify one.
- Building an `LLMCoordinator` mirroring `ScanCoordinator` / `GenerateCoordinator`.
- Changing `handleLLMProgress` (`internal/ui/wizard.go:857-864`), `LLMProgressMsg`'s shape, or the `internal/app` / `internal/platform` layers.
- Adding a progress bar, spinner or percentage to the send. The elapsed-seconds counter already exists and is the target.
- Changing the send timeout or adding cancellation.

### Technical Approach

- Root cause: `sendToLLMCmd` (`internal/ui/wizard.go:922-938`) is a single `tea.Cmd` that blocks for the whole request and returns one message. Bubble Tea re-renders only in response to a message, so between `handleSendToLLM` and `LLMCompleteMsg` no `View()` call occurs. `internal/ui/screens/review.go:427-430` renders `⏳ Sending to LLM... (%s)` from `time.Since(m.llmStartTime).Round(time.Second)`, so with no re-render the counter is pinned at `(0s)` for up to the configured 300s timeout. A periodic message is the whole fix.
- Add near the existing `copyToastClearMsg` tick (`internal/ui/wizard.go:822-829`), whose shape this mirrors:

  ```go
  const llmTickInterval = time.Second

  // llmTickMsg re-renders the Review screen while a send is in flight so its
  // elapsed-time counter advances. The send itself is one blocking tea.Cmd and
  // emits no intermediate messages.
  type llmTickMsg struct{}

  func llmTick() tea.Cmd {
      return tea.Tick(llmTickInterval, func(time.Time) tea.Msg {
          return llmTickMsg{}
      })
  }
  ```

- In `handleSendToLLM` (`internal/ui/wizard.go:832-855`), extend the existing return:

  ```go
  return tea.Batch(m.sendToLLMCmd(), m.progressComponent.Init(), llmTick())
  ```

- In `WizardModel.Update`, add an arm beside the existing `screens.LLMProgressMsg` / `LLMCompleteMsg` / `LLMErrorMsg` cases (around `internal/ui/wizard.go:253`), using the file's established `cmds = append(...)` accumulation:

  ```go
  case llmTickMsg:
      if m.llmSending {
          cmds = append(cmds, llmTick())
      }
  ```

  The tick is self-terminating: `handleLLMComplete` and `handleLLMError` already set `m.llmSending = false`, so the next tick to arrive is the last. No explicit cancellation, no stored cancel handle, and no risk of a leaked repeating command — a lapsed chain is simply not re-armed. At most one extra tick fires after completion, which is harmless.
- Guard against double-arming: the tick is emitted from exactly two places — `handleSendToLLM` (once, at send start) and the `llmTickMsg` arm (only while sending). `handleSendToLLM` already returns `nil` early when `m.llmSending` is true (`internal/ui/wizard.go:833-835`), so a second send cannot start a second chain.
- In `internal/ui/screens/review.go:147-150`, the `LLMProgressMsg` arm currently reads:

  ```go
  case LLMProgressMsg:
      m.llmSending = true
      m.llmStartTime = time.Now()
      return true, nil
  ```

  Delete the `m.llmStartTime = time.Now()` line. `SetLLMSending(true)` (`internal/ui/screens/review.go:467-473`) already stamps `llmStartTime` once when the send begins; re-stamping it on every progress message would reset the counter to zero mid-flight. Keep `m.llmSending = true` — it is an idempotent restatement and costs nothing.
- Concurrency: `tea.Tick` schedules through Bubble Tea's own message loop, so the tick handler runs on the update goroutine and touches no field the send goroutine writes. No mutex, no coordinator, and nothing new for the race detector to find — but the change must still be verified under `-race`.

### Touchpoints

- `internal/ui/wizard.go` — `llmTickInterval`, `llmTickMsg`, `llmTick()` added; `handleSendToLLM` return extended with `llmTick()`; an `llmTickMsg` arm added to `Update`.
- `internal/ui/screens/review.go` — one line deleted from the `LLMProgressMsg` arm of `HandleMessage`.
- `internal/ui/wizard_test.go` — three tests added.
- `internal/ui/screens/review_test.go` — one test added.

### Contracts

```go
// internal/ui/wizard.go — new, unexported
const llmTickInterval = time.Second

type llmTickMsg struct{}

func llmTick() tea.Cmd

// internal/ui/wizard.go — unchanged signature, extended return
func (m *WizardModel) handleSendToLLM() tea.Cmd
// returns tea.Batch(m.sendToLLMCmd(), m.progressComponent.Init(), llmTick())

// internal/ui/wizard.go — new Update arm
case llmTickMsg:
    if m.llmSending {
        cmds = append(cmds, llmTick())
    }

// internal/ui/screens/review.go — arm after the change
case LLMProgressMsg:
    m.llmSending = true
    return true, nil

// Deliberately unchanged
func (m *WizardModel) sendToLLMCmd() tea.Cmd
// still calls SendToLLMWithProgress(ctx, content, cfg, nil)
```

### Acceptance Criteria

- [ ] `internal/ui/wizard.go` declares `llmTickMsg`, `llmTick()` and `llmTickInterval = time.Second`.
- [ ] `handleSendToLLM` returns a `tea.Batch` that includes `llmTick()`.
- [ ] `WizardModel.Update` has an `llmTickMsg` arm that re-arms the tick only when `m.llmSending` is true.
- [ ] With `m.llmSending = true`, `wizard.Update(llmTickMsg{})` returns a non-nil command whose execution yields another `llmTickMsg`.
- [ ] With `m.llmSending = false`, `wizard.Update(llmTickMsg{})` produces no `llmTickMsg`-yielding command.
- [ ] After `wizard.Update(screens.LLMCompleteMsg{...})`, `m.llmSending` is false and a subsequent `llmTickMsg` does not re-arm.
- [ ] After `wizard.Update(screens.LLMErrorMsg{...})`, `m.llmSending` is false and a subsequent `llmTickMsg` does not re-arm.
- [ ] `grep -n "llmStartTime = time.Now()" internal/ui/screens/review.go` returns exactly two hits, both inside `SetLLMSending`, and none inside `HandleMessage`.
- [ ] A `ReviewModel` test calls `SetLLMSending(true)`, records `llmStartTime`, sends `LLMProgressMsg`, and asserts `llmStartTime` is unchanged.
- [ ] `sendToLLMCmd` still passes `nil` as the fourth argument to `SendToLLMWithProgress`.
- [ ] `grep -rn "LLMCoordinator" internal/ui/` returns no results.
- [ ] Manual QA — run the TUI against a real configured provider, reach Step 5, press F9, and observe `⏳ Sending to LLM... (1s)`, `(2s)`, `(3s)` advancing once per second until the response arrives. This is the gate: the change is not done until the counter has been watched advancing in a real terminal.
- [ ] Manual QA — after the response arrives, the terminal stays idle (no continuing redraw) for at least 10 seconds, confirming the tick chain lapsed.
- [ ] `go test -race ./internal/ui/...` passes.
- [ ] `golangci-lint run --config .golangci.yml ./...` reports no new findings.

### Dependencies

- None.
