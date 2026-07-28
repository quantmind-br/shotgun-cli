package ui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

const chainTimeout = 10 * time.Second

// newScanFixture builds a small tree under t.TempDir(). Small matters:
// FileSystemScanner throttles progress to every 100th entry, so a handful of
// files produces ~2 events into a 100-slot channel. The worker therefore never
// blocks waiting to be drained, which is what makes the reset-mid-flight and
// deferred-terminal-command scenarios below implementable without deadlock.
func newScanFixture(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "c.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("package main\n"), 0o600); err != nil {
			t.Fatalf("fixture write failed: %v", err)
		}
	}
	sub := filepath.Join(root, "pkg")
	if err := os.Mkdir(sub, 0o750); err != nil {
		t.Fatalf("fixture mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub, "d.go"), []byte("package pkg\n"), 0o600); err != nil {
		t.Fatalf("fixture write failed: %v", err)
	}

	return root
}

func newTestScanConfig() *scanner.ScanConfig {
	cfg := scanner.DefaultScanConfig()
	// Keep the walk hermetic: the fixture has no .gitignore, and loading the
	// repo's would depend on where the test runs from.
	cfg.RespectGitignore = false
	cfg.RespectShotgunignore = false

	return cfg
}

// TestScanCoordinator_RapidRestart_RealScanner drives three Starts in a row
// against a real FileSystemScanner without waiting in between. A mock would
// exercise neither the shared ignoreEngine nor the shared-config race.
func TestScanCoordinator_RapidRestart_RealScanner(t *testing.T) {
	t.Parallel()

	rootA := newScanFixture(t)
	rootB := newScanFixture(t)
	rootC := newScanFixture(t)

	counting := &countingScanner{inner: scanner.NewFileSystemScanner()}
	coordinator := NewScanCoordinator(counting)
	coordinator.pollInterval = time.Millisecond

	cfg := newTestScanConfig()

	// No Poll between these, so runs B and C queue: running tracks reaping, not
	// worker exit, which makes the queueing deterministic without gating.
	cmds := []tea.Cmd{
		coordinator.Start(rootA, cfg),
		coordinator.Start(rootB, cfg),
		coordinator.Start(rootC, cfg),
	}

	if cmds[0] == nil {
		t.Fatal("the first Start should return a command")
	}
	if cmds[1] != nil || cmds[2] != nil {
		t.Error("every Start issued while a run is in flight must return nil")
	}

	res := driveScanChain(t, coordinator, tea.Batch(cmds...), chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d: %#v", len(res.terminal), res.terminal)
	}
	complete, ok := res.terminal[0].(ScanCompleteMsg)
	if !ok {
		t.Fatalf("expected ScanCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Tree == nil {
		t.Fatal("terminal message carried a nil tree")
	}
	if complete.Tree.Path != rootC {
		t.Errorf("delivered result is not the final request: got %s, want %s", complete.Tree.Path, rootC)
	}

	if got := counting.max.Load(); got != 1 {
		t.Errorf("expected at most one live scan goroutine, saw %d", got)
	}
	if res.maxPollsPerStep > 1 {
		t.Errorf("more than one poll chain was live: %d polls from a single step", res.maxPollsPerStep)
	}
	if coordinator.IsRunning() {
		t.Error("IsRunning should be false once the final run completed with nothing pending")
	}
}

// TestScanCoordinator_ConfigMutatedMidScan flips IncludeIgnored on the caller's
// struct while a scan is in flight. Under -race the mutation is the oracle:
// without the clone, this write races the walk's reads.
func TestScanCoordinator_ConfigMutatedMidScan(t *testing.T) {
	t.Parallel()

	root := newScanFixture(t)
	probe := &configProbeScanner{
		inner:   scanner.NewFileSystemScanner(),
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
	coordinator := NewScanCoordinator(probe)
	coordinator.pollInterval = time.Millisecond

	cfg := newTestScanConfig()
	cfg.IncludeIgnored = false

	cmd := coordinator.Start(root, cfg)
	cmd()
	<-probe.entered

	// Exactly what handleToggleIgnoredScan does: mutate in place and reuse the
	// pointer. The active run must not see it.
	cfg.IncludeIgnored = true
	cfg.IgnorePatterns = append(cfg.IgnorePatterns, "*.go")

	close(probe.release)
	<-coordinator.done

	if probe.seen[0] != probe.seen[1] {
		t.Errorf("the active run observed a mid-flight config change: %v then %v", probe.seen[0], probe.seen[1])
	}
	if probe.seen[0] {
		t.Error("the active run should have observed IncludeIgnored=false, the value at Start time")
	}
}

// TestScanCoordinator_QueuedStartEmitsOneCompletion queues a second request
// mid-run and asserts a single terminal message carrying the second run's tree.
// Emitting the first run's tree is a permanent wrong-state bug in the wizard: it
// consumes the one-shot selectionsSeeded latch.
func TestScanCoordinator_QueuedStartEmitsOneCompletion(t *testing.T) {
	t.Parallel()

	rootB := newScanFixture(t)
	gated := newGatedScanner(func(rootPath string) *scanner.FileNode {
		return &scanner.FileNode{Name: "root", Path: rootPath, IsDir: true}
	})
	coordinator := NewScanCoordinator(gated)
	coordinator.pollInterval = time.Millisecond

	cfg := newTestScanConfig()

	cmdA := coordinator.Start("/run-a", cfg)
	cmdA()
	<-gated.entered

	cmdB := coordinator.Start(rootB, cfg)
	if cmdB != nil {
		t.Error("the queued Start must return nil")
	}

	close(gated.release)

	res := driveScanChain(t, coordinator, func() tea.Msg { return pollScanMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d: %#v", len(res.terminal), res.terminal)
	}
	complete, ok := res.terminal[0].(ScanCompleteMsg)
	if !ok {
		t.Fatalf("expected ScanCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Tree == nil || complete.Tree.Path != rootB {
		t.Errorf("expected the second run's tree (%s), got %#v", rootB, complete.Tree)
	}
	if res.maxPollsPerStep > 1 {
		t.Errorf("a rival poll chain was live: %d polls from a single step", res.maxPollsPerStep)
	}
}

// TestScanCoordinator_ResetMidFlight resets while a run is blocked, then
// releases it. The orphan must write no result and emit no terminal message.
func TestScanCoordinator_ResetMidFlight(t *testing.T) {
	t.Parallel()

	gated := newGatedScanner(func(rootPath string) *scanner.FileNode {
		return &scanner.FileNode{Name: "orphan", Path: rootPath, IsDir: true}
	})
	coordinator := NewScanCoordinator(gated)
	coordinator.pollInterval = time.Millisecond

	cmd := coordinator.Start("/orphan", newTestScanConfig())
	cmd()
	<-gated.entered

	done := coordinator.done
	coordinator.Reset()

	// Reset must not disarm the live run: doing so would let the next Start walk
	// concurrently against the shared ignoreEngine.
	if !coordinator.IsRunning() {
		t.Error("Reset must leave running set while a run is unreaped")
	}
	if coordinator.progressCh == nil || coordinator.done == nil {
		t.Fatal("Reset must leave the live run's channels intact, or its poll chain dies")
	}
	if coordinator.activeGen == coordinator.generation {
		t.Error("Reset should have advanced generation past activeGen")
	}

	close(gated.release)
	<-done

	res := driveScanChain(t, coordinator, func() tea.Msg { return pollScanMsg{} }, chainTimeout)

	if len(res.terminal) != 0 {
		t.Errorf("a reset run must emit no terminal message, got %#v", res.terminal)
	}
	if res.progress != 0 {
		t.Errorf("a reset run must emit no progress, got %d messages", res.progress)
	}

	tree, err := coordinator.Result()
	if tree != nil || err != nil {
		t.Errorf("a reset run must write no result, got tree=%#v err=%v", tree, err)
	}
	if coordinator.IsComplete() {
		t.Error("IsComplete should stay false after a reset run finishes")
	}
	if coordinator.IsRunning() {
		t.Error("the orphan's own chain should have cleared running")
	}
}

// TestScanCoordinator_ResetThenStartBeforeRelease is the case that strands if
// Reset nils a live run's channels: the queued request would have nothing left
// to launch it, because a post-reset Start returns nil.
func TestScanCoordinator_ResetThenStartBeforeRelease(t *testing.T) {
	t.Parallel()

	rootB := newScanFixture(t)
	gated := newGatedScanner(func(rootPath string) *scanner.FileNode {
		return &scanner.FileNode{Name: "root", Path: rootPath, IsDir: true}
	})
	coordinator := NewScanCoordinator(gated)
	coordinator.pollInterval = time.Millisecond

	cfg := newTestScanConfig()

	cmdA := coordinator.Start("/orphan", cfg)
	cmdA()
	<-gated.entered

	coordinator.Reset()

	cmdB := coordinator.Start(rootB, cfg)
	if cmdB != nil {
		t.Error("a Start issued while the orphan is unreaped must queue and return nil")
	}

	close(gated.release)

	res := driveScanChain(t, coordinator, func() tea.Msg { return pollScanMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("the queued run should deliver exactly one terminal message, got %d: %#v",
			len(res.terminal), res.terminal)
	}
	complete, ok := res.terminal[0].(ScanCompleteMsg)
	if !ok {
		t.Fatalf("expected ScanCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Tree == nil || complete.Tree.Path != rootB {
		t.Errorf("expected the queued run's tree (%s), got %#v", rootB, complete.Tree)
	}
	if coordinator.IsRunning() {
		t.Error("IsRunning should be false after the queued run completed")
	}
}

// TestScanCoordinator_DeferredTerminalCommand holds the terminal command across
// a later Start and a later Reset. This is the case a lock inside finishScan
// does not fix: the read has to happen at decision time, not execution time.
func TestScanCoordinator_DeferredTerminalCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		disturb func(c *ScanCoordinator, cfg *scanner.ScanConfig)
	}{
		{
			name: "held across a later Start",
			disturb: func(c *ScanCoordinator, cfg *scanner.ScanConfig) {
				c.Start("/run-b", cfg)
			},
		},
		{
			name: "held across a later Reset",
			disturb: func(c *ScanCoordinator, _ *scanner.ScanConfig) {
				c.Reset()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockSc := &scanCoordinatorMockScanner{
				scanProgressFunc: func(
					rootPath string, _ *scanner.ScanConfig, _ chan<- scanner.Progress,
				) (*scanner.FileNode, error) {
					return &scanner.FileNode{Name: "run-a", Path: rootPath, IsDir: true}, nil
				},
			}
			coordinator := NewScanCoordinator(mockSc)
			cfg := newTestScanConfig()

			cmd := coordinator.Start("/run-a", cfg)
			done := coordinator.done
			cmd()
			<-done

			// Obtain the terminal command but do NOT invoke it yet.
			terminal := coordinator.Poll()
			if terminal == nil {
				t.Fatal("Poll should have returned a terminal command")
			}

			// After the reap, running is false, so this Start begins immediately
			// rather than queueing.
			tt.disturb(coordinator, cfg)

			msg := terminal()
			complete, ok := msg.(ScanCompleteMsg)
			if !ok {
				t.Fatalf("expected ScanCompleteMsg, got %T", msg)
			}
			if complete.Tree == nil {
				t.Fatal("the held command delivered a nil tree; it read coordinator state too late")
			}
			if complete.Tree.Path != "/run-a" {
				t.Errorf("the held command delivered the wrong run's tree: got %s, want /run-a",
					complete.Tree.Path)
			}
		})
	}
}

// TestScanCoordinator_PollAfterReapIsInert covers the guard at the top of Poll.
// After the terminal branch, done stays closed on the struct, and a closed
// channel is always ready -- so without the !running check an extra Poll would
// take the completion branch again and emit a second terminal message, breaking
// the exactly-one-completion contract.
//
// A stray Poll is reachable: a tick scheduled before completion can still be
// sitting in Bubble Tea's queue.
func TestScanCoordinator_PollAfterReapIsInert(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(
			rootPath string, _ *scanner.ScanConfig, _ chan<- scanner.Progress,
		) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root", Path: rootPath, IsDir: true}, nil
		},
	}
	coordinator := NewScanCoordinator(mockSc)

	cmd := coordinator.Start("/run-a", newTestScanConfig())
	done := coordinator.done
	cmd()
	<-done

	first := coordinator.Poll()
	if first == nil {
		t.Fatal("the first Poll after completion should return the terminal command")
	}
	if _, ok := first().(ScanCompleteMsg); !ok {
		t.Fatal("expected the first Poll to yield ScanCompleteMsg")
	}

	for i := range 3 {
		if extra := coordinator.Poll(); extra != nil {
			t.Fatalf("Poll %d after the reap returned a command (%T); the terminal message would be emitted twice",
				i+2, extra())
		}
	}
}

// TestScanCoordinator_ErrorRunEmitsScanErrorMsg keeps the error path covered
// now that finishScan takes its payload by parameter.
func TestScanCoordinator_ErrorRunEmitsScanErrorMsg(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("permission denied")
	gated := newGatedScanner(nil)
	gated.err = wantErr

	coordinator := NewScanCoordinator(gated)
	coordinator.pollInterval = time.Millisecond

	cmd := coordinator.Start("/boom", newTestScanConfig())
	cmd()
	<-gated.entered
	close(gated.release)

	res := driveScanChain(t, coordinator, func() tea.Msg { return pollScanMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d", len(res.terminal))
	}
	scanErr, ok := res.terminal[0].(ScanErrorMsg)
	if !ok {
		t.Fatalf("expected ScanErrorMsg, got %T", res.terminal[0])
	}
	if !errors.Is(scanErr.Err, wantErr) {
		t.Errorf("expected the scanner's error, got %v", scanErr.Err)
	}
}
