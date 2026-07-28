package ui

import (
	"fmt"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

// scanCoordinatorMockScanner is a test double that implements scanner.Scanner
type scanCoordinatorMockScanner struct {
	scanFunc         func(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error)
	scanProgressFunc func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error)
}

func (m *scanCoordinatorMockScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
	return m.scanFunc(rootPath, config)
}

func (m *scanCoordinatorMockScanner) ScanWithProgress(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
	return m.scanProgressFunc(rootPath, config, progress)
}

func TestScanCoordinator_New(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{}
	coordinator := NewScanCoordinator(mockSc)

	if coordinator.scanner != mockSc {
		t.Errorf("scanner not set correctly")
	}

	if coordinator.rootPath != "" {
		t.Errorf("rootPath should be empty initially")
	}

	if coordinator.config != nil {
		t.Errorf("config should be nil initially")
	}

	if coordinator.progressCh != nil {
		t.Errorf("progressCh should not be nil")
	}

	if coordinator.done != nil {
		t.Errorf("done channel should not be nil")
	}

	if coordinator.result != nil {
		t.Errorf("result should be nil initially")
	}

	if coordinator.scanErr != nil {
		t.Errorf("scanErr should be nil initially")
	}

	if coordinator.started != false {
		t.Errorf("should not be started initially")
	}
}

func TestScanCoordinator_Start(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root", Path: rootPath, IsDir: true}, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{
		MaxFiles:        100,
		IgnorePatterns:  []string{"*.log"},
		IncludePatterns: []string{"*.go"},
	}

	cmd := coordinator.Start("/test", cfg)

	if cmd == nil {
		t.Fatal("Start should return a command")
	}

	// Verify coordinator state
	if coordinator.scanner != mockSc {
		t.Errorf("scanner not preserved")
	}
	if coordinator.rootPath != "/test" {
		t.Errorf("rootPath not set: got %s", coordinator.rootPath)
	}

	// Start must hold a private clone, never the caller's pointer: the wizard
	// toggles IncludeIgnored on the struct it passed in and hands the same
	// pointer to the next run.
	if coordinator.config == cfg {
		t.Error("Start retained the caller's ScanConfig pointer")
	}
	if coordinator.config.MaxFiles != cfg.MaxFiles {
		t.Errorf("clone lost MaxFiles: got %d, want %d", coordinator.config.MaxFiles, cfg.MaxFiles)
	}
	if &coordinator.config.IgnorePatterns[0] == &cfg.IgnorePatterns[0] {
		t.Error("IgnorePatterns still aliases the caller's backing array")
	}
	if &coordinator.config.IncludePatterns[0] == &cfg.IncludePatterns[0] {
		t.Error("IncludePatterns still aliases the caller's backing array")
	}

	if coordinator.progressCh == nil {
		t.Errorf("progressCh not initialized")
	}
	if coordinator.done == nil {
		t.Errorf("done channel not initialized")
	}
	if coordinator.started != false {
		t.Errorf("should not be started yet")
	}
	if !coordinator.IsRunning() {
		t.Error("IsRunning should be true once Start accepted the run")
	}
	if coordinator.generation != 1 || coordinator.activeGen != 1 {
		t.Errorf("expected generation/activeGen 1/1, got %d/%d", coordinator.generation, coordinator.activeGen)
	}
	if coordinator.result != nil || coordinator.scanErr != nil {
		t.Errorf("result and scanErr should be nil initially")
	}
}

// TestScanCoordinator_Start_QueuesWhileRunning pins the single-flight contract.
// The first command is deliberately never invoked: running is set inside Start,
// not inside the command, so the gate is already closed.
func TestScanCoordinator_Start_QueuesWhileRunning(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{}
	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}

	cmd1 := coordinator.Start("/test1", cfg)
	if cmd1 == nil {
		t.Fatal("first Start should return a command")
	}

	firstDoneChan := coordinator.done
	firstProgressCh := coordinator.progressCh

	cmd2 := coordinator.Start("/test2", cfg)

	// Returning a command here would start a rival poll chain, and once done
	// closes both chains would emit the terminal message.
	if cmd2 != nil {
		t.Error("Start should return nil while a run is in flight")
	}
	if coordinator.pending == nil {
		t.Fatal("second request should have been recorded as pending")
	}
	if coordinator.pending.rootPath != "/test2" {
		t.Errorf("pending holds the wrong request: got %s", coordinator.pending.rootPath)
	}
	if coordinator.pending.config == cfg {
		t.Error("pending retained the caller's ScanConfig pointer")
	}
	if coordinator.done != firstDoneChan {
		t.Error("queued Start must not replace the live run's done channel")
	}
	if coordinator.progressCh != firstProgressCh {
		t.Error("queued Start must not replace the live run's progress channel")
	}
	if coordinator.generation != 1 {
		t.Errorf("generation should not advance on a queued request, got %d", coordinator.generation)
	}
	if coordinator.result != nil || coordinator.scanErr != nil {
		t.Error("result and scanErr should still be nil")
	}
}

// TestScanCoordinator_Start_AfterReapStartsNewRun is the other half: once Poll
// has reaped the previous run, Start begins immediately instead of queueing.
func TestScanCoordinator_Start_AfterReapStartsNewRun(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root", Path: rootPath}, nil
		},
	}
	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}

	cmd := coordinator.Start("/test1", cfg)
	firstDoneChan := coordinator.done
	cmd()
	<-firstDoneChan

	if terminal := coordinator.Poll(); terminal == nil {
		t.Fatal("Poll should have reaped the completed run")
	}
	if coordinator.IsRunning() {
		t.Error("IsRunning should be false after the run was reaped")
	}

	cmd2 := coordinator.Start("/test2", cfg)
	if cmd2 == nil {
		t.Fatal("Start after a reap should begin a new run")
	}
	if coordinator.done == firstDoneChan {
		t.Error("the new run should own new channels")
	}
	if coordinator.generation != 2 {
		t.Errorf("generation should be 2, got %d", coordinator.generation)
	}
	if coordinator.result != nil || coordinator.scanErr != nil {
		t.Error("the new run should start from clean result state")
	}
}

func TestScanCoordinator_Poll_BeforeStart(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{}
	coordinator := NewScanCoordinator(mockSc)

	// Call Poll before Start - should return nil
	cmd := coordinator.Poll()

	if cmd != nil {
		t.Error("Poll before Start should return nil")
	}
}

func TestScanCoordinator_Poll_DuringScan(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{}
	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}

	// Start the coordinator
	coordinator.Start("/test", cfg)

	// Send a progress update
	go func() {
		coordinator.progressCh <- scanner.Progress{
			Current: 1,
			Total:   10,
			Stage:   "scanning",
		}
	}()

	// Call Poll - should return progress message
	cmd := coordinator.Poll()

	if cmd == nil {
		t.Fatal("Poll should return a command during scan")
	}

	// Verify the command is a Batch with progress and poll schedule
	// We can't easily check this without access to the returned command structure
	// Just verify it's not nil
}

func TestScanCoordinator_Poll_AfterComplete(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root"}, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	pollCmd := coordinator.Poll()
	if pollCmd == nil {
		t.Error("Poll after complete should return finishScan cmd")
	}

	msg := pollCmd()
	if _, ok := msg.(ScanCompleteMsg); !ok {
		t.Errorf("Poll after complete should return ScanCompleteMsg, got %T", msg)
	}
}

func TestScanCoordinator_Result_Success(t *testing.T) {
	t.Parallel()

	expectedTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/test",
		IsDir: true,
	}

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return expectedTree, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	tree, err := coordinator.Result()

	if err != nil {
		t.Errorf("Result should not return error: %v", err)
	}

	if tree == nil {
		t.Fatal("Result should return tree")
	}

	if tree.Name != expectedTree.Name {
		t.Errorf("unexpected tree name: got %s", tree.Name)
	}
	if tree.Path != expectedTree.Path {
		t.Errorf("unexpected tree path: got %s", tree.Path)
	}
}

func TestScanCoordinator_Result_Error(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("permission denied")

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return nil, expectedErr
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	tree, err := coordinator.Result()

	if err != expectedErr {
		t.Errorf("Result should return scan error: got %v, want %v", err, expectedErr)
	}

	if tree != nil {
		t.Error("Result should return nil tree on error")
	}
}

func TestScanCoordinator_IsComplete_NotStarted(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{}
	coordinator := NewScanCoordinator(mockSc)

	if coordinator.IsComplete() {
		t.Error("IsComplete should return false before scan starts")
	}

	if coordinator.IsStarted() {
		t.Error("IsStarted should return false before scan starts")
	}
}

func TestScanCoordinator_IsComplete_AfterStart(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	proceed := make(chan struct{})

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			close(started)
			<-proceed
			return nil, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-started

	if coordinator.IsComplete() {
		t.Error("IsComplete should return false while scanning")
	}

	if !coordinator.IsStarted() {
		t.Error("IsStarted should return true after starting scan")
	}

	close(proceed)
	<-coordinator.done
}

func TestScanCoordinator_IsComplete_AfterSuccess(t *testing.T) {
	t.Parallel()

	expectedTree := &scanner.FileNode{Name: "root"}

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return expectedTree, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	if !coordinator.IsComplete() {
		t.Error("IsComplete should return true after successful scan")
	}

	if coordinator.scanErr != nil {
		t.Error("scanErr should be nil on success")
	}
}

func TestScanCoordinator_IsComplete_AfterError(t *testing.T) {
	t.Parallel()

	expectedErr := fmt.Errorf("permission denied")

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return nil, expectedErr
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	if !coordinator.IsComplete() {
		t.Error("IsComplete should return true even after error")
	}

	if coordinator.scanErr != expectedErr {
		t.Error("scanErr should be set on error")
	}
}

func TestScanCoordinator_IsStarted(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root"}, nil
		},
	}
	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	if !coordinator.IsStarted() {
		t.Error("IsStarted should return true after Start is called")
	}

	<-coordinator.done
}

func TestScanCoordinator_ProgressMessages(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			progress <- scanner.Progress{Current: 1, Total: 10, Stage: "collecting"}
			progress <- scanner.Progress{Current: 5, Total: 10, Stage: "scanning"}
			progress <- scanner.Progress{Current: 10, Total: 10, Stage: "complete"}
			return &scanner.FileNode{Name: "root"}, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	var progressUpdates []scanner.Progress
	for {
		select {
		case progress := <-coordinator.progressCh:
			progressUpdates = append(progressUpdates, progress)
		default:
			goto done
		}
	}
done:

	if len(progressUpdates) < 2 {
		t.Errorf("should receive at least 2 progress updates, got %d", len(progressUpdates))
	}

	if len(progressUpdates) >= 2 {
		if progressUpdates[0].Current != 1 || progressUpdates[0].Total != 10 {
			t.Errorf("unexpected first progress: %+v", progressUpdates[0])
		}
		if progressUpdates[1].Current != 5 || progressUpdates[1].Total != 10 {
			t.Errorf("unexpected second progress: %+v", progressUpdates[1])
		}
	}
}

func TestScanCoordinator_MultipleProgressUpdates(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			for i := int64(1); i <= 100; i++ {
				progress <- scanner.Progress{Current: i, Total: 100, Stage: "processing"}
			}
			return &scanner.FileNode{Name: "root"}, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	var collected []scanner.Progress
	for {
		select {
		case progress := <-coordinator.progressCh:
			collected = append(collected, progress)
		default:
			goto done
		}
	}
done:

	if len(collected) < 3 {
		t.Errorf("should collect multiple progress updates, got %d", len(collected))
	}
}

func TestScanCoordinator_Result_BeforeComplete(t *testing.T) {
	t.Parallel()

	started := make(chan struct{})
	proceed := make(chan struct{})

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			close(started)
			<-proceed
			return &scanner.FileNode{Name: "root"}, nil
		},
	}

	coordinator := NewScanCoordinator(mockSc)
	cmd := coordinator.Start("/test", nil)
	cmd()

	<-started

	tree, err := coordinator.Result()

	if tree != nil {
		t.Error("Result should return nil tree before completion")
	}
	if err != nil {
		t.Error("Result should return nil error before completion")
	}

	close(proceed)
	<-coordinator.done
}

func TestScanCoordinator_Reset(t *testing.T) {
	t.Parallel()

	mockSc := &scanCoordinatorMockScanner{
		scanProgressFunc: func(rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress) (*scanner.FileNode, error) {
			return &scanner.FileNode{Name: "root"}, nil
		},
	}
	coordinator := NewScanCoordinator(mockSc)
	cfg := &scanner.ScanConfig{MaxFiles: 100}
	cmd := coordinator.Start("/test", cfg)
	cmd()

	<-coordinator.done

	// Reap first. Reset deliberately preserves the channels of an unreaped run,
	// so without this Poll the nil-channel assertions below would not hold --
	// and could not, without stranding the live poll chain.
	coordinator.Poll()

	coordinator.Reset()

	if coordinator.rootPath != "" {
		t.Error("rootPath should be empty after reset")
	}
	if coordinator.config != nil {
		t.Error("config should be nil after reset")
	}
	if coordinator.progressCh != nil {
		t.Error("progressCh should be nil after reset")
	}
	if coordinator.done != nil {
		t.Error("done channel should be nil after reset")
	}
	if coordinator.result != nil {
		t.Error("result should be nil after reset")
	}
	if coordinator.scanErr != nil {
		t.Error("scanErr should be nil after reset")
	}
	if coordinator.started != false {
		t.Error("started should be false after reset")
	}
	if coordinator.pending != nil {
		t.Error("pending should be nil after reset")
	}
	if coordinator.generation == 0 {
		t.Error("Reset should increment generation, so an in-flight run's result is discarded")
	}
}
