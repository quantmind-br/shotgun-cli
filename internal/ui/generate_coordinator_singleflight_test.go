package ui

import (
	"errors"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

// gatedGenerator blocks inside Generate until released, and reports the task
// description it was handed so a test can tell the runs apart.
type gatedGenerator struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	live    int
	maxLive int
	mu      sync.Mutex
	err     error
}

func newGatedGenerator() *gatedGenerator {
	return &gatedGenerator{
		entered: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (g *gatedGenerator) Generate(
	_ *scanner.FileNode, _ map[string]bool, config contextgen.GenerateConfig,
) (string, error) {
	g.mu.Lock()
	g.live++
	if g.live > g.maxLive {
		g.maxLive = g.live
	}
	g.mu.Unlock()
	defer func() {
		g.mu.Lock()
		g.live--
		g.mu.Unlock()
	}()

	g.once.Do(func() { close(g.entered) })
	<-g.release

	if g.err != nil {
		return "", g.err
	}

	return config.TemplateVars["TASK"], nil
}

func (g *gatedGenerator) GenerateWithProgress(
	root *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig, _ func(string),
) (string, error) {
	return g.Generate(root, selections, config)
}

func (g *gatedGenerator) GenerateWithProgressEx(
	root *scanner.FileNode, selections map[string]bool, config contextgen.GenerateConfig,
	_ func(contextgen.GenProgress),
) (string, error) {
	return g.Generate(root, selections, config)
}

func newTestGenerateConfig(task string) *GenerateConfig {
	return &GenerateConfig{
		FileTree:   &scanner.FileNode{Name: "root", IsDir: true},
		Selections: map[string]bool{"a.go": true},
		Template:   &template.Template{Content: "{TASK}"},
		TaskDesc:   task,
	}
}

// TestGenerateCoordinator_RapidRestart drives three Starts without waiting and
// asserts the delivered content belongs to the final request.
func TestGenerateCoordinator_RapidRestart(t *testing.T) {
	t.Parallel()

	gen := newGatedGenerator()
	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	cmds := []tea.Cmd{
		coord.Start(newTestGenerateConfig("run-a")),
		coord.Start(newTestGenerateConfig("run-b")),
		coord.Start(newTestGenerateConfig("run-c")),
	}

	if cmds[0] == nil {
		t.Fatal("the first Start should return a command")
	}
	if cmds[1] != nil || cmds[2] != nil {
		t.Error("every Start issued while a run is in flight must return nil")
	}

	// Release as soon as the first worker is inside, then let the chain run: each
	// queued handoff reuses the same release channel, which is already closed.
	go func() {
		<-gen.entered
		close(gen.release)
	}()

	res := driveGenerateChain(t, coord, tea.Batch(cmds...), chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d: %#v", len(res.terminal), res.terminal)
	}
	complete, ok := res.terminal[0].(screens.GenerationCompleteMsg)
	if !ok {
		t.Fatalf("expected GenerationCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Content != "run-c" {
		t.Errorf("delivered result is not the final request: got %q, want %q", complete.Content, "run-c")
	}

	gen.mu.Lock()
	maxLive := gen.maxLive
	gen.mu.Unlock()
	if maxLive != 1 {
		t.Errorf("expected at most one live generation, saw %d", maxLive)
	}
	if res.maxPollsPerStep > 1 {
		t.Errorf("more than one poll chain was live: %d polls from a single step", res.maxPollsPerStep)
	}
	if coord.IsRunning() {
		t.Error("IsRunning should be false once the final run completed")
	}
}

// TestGenerateCoordinator_QueuedStartEmitsOneCompletion queues a second request
// mid-run and asserts a single terminal message carrying the second run's output.
func TestGenerateCoordinator_QueuedStartEmitsOneCompletion(t *testing.T) {
	t.Parallel()

	gen := newGatedGenerator()
	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	cmdA := coord.Start(newTestGenerateConfig("run-a"))
	cmdA()
	<-gen.entered

	if cmdB := coord.Start(newTestGenerateConfig("run-b")); cmdB != nil {
		t.Error("the queued Start must return nil")
	}

	close(gen.release)

	res := driveGenerateChain(t, coord, func() tea.Msg { return pollGenerateMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d: %#v", len(res.terminal), res.terminal)
	}
	complete, ok := res.terminal[0].(screens.GenerationCompleteMsg)
	if !ok {
		t.Fatalf("expected GenerationCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Content != "run-b" {
		t.Errorf("expected the second run's content, got %q", complete.Content)
	}
}

// TestGenerateCoordinator_ConfigMutatedMidRun mutates the caller's config while
// a run is in flight, including the Selections map the file-selection screen
// hands over live.
func TestGenerateCoordinator_ConfigMutatedMidRun(t *testing.T) {
	t.Parallel()

	gen := newGatedGenerator()
	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	cfg := newTestGenerateConfig("run-a")
	cmd := coord.Start(cfg)
	cmd()
	<-gen.entered

	// A concurrent write to a shared map would be a fatal error, not a race
	// report; the clone is what makes this safe.
	cfg.Selections["b.go"] = true
	cfg.TaskDesc = "mutated"

	close(gen.release)

	res := driveGenerateChain(t, coord, func() tea.Msg { return pollGenerateMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d", len(res.terminal))
	}
	complete, ok := res.terminal[0].(screens.GenerationCompleteMsg)
	if !ok {
		t.Fatalf("expected GenerationCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Content != "run-a" {
		t.Errorf("the run observed a mid-flight config change: got %q, want %q", complete.Content, "run-a")
	}
}

// TestGenerateCoordinator_ResetMidFlight resets while a run is blocked and
// asserts the orphan writes nothing and emits nothing.
func TestGenerateCoordinator_ResetMidFlight(t *testing.T) {
	t.Parallel()

	gen := newGatedGenerator()
	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	cmd := coord.Start(newTestGenerateConfig("orphan"))
	cmd()
	<-gen.entered

	done := coord.done
	coord.Reset()

	if !coord.IsRunning() {
		t.Error("Reset must leave running set while a run is unreaped")
	}
	if coord.progressCh == nil || coord.done == nil {
		t.Fatal("Reset must leave the live run's channels intact")
	}

	close(gen.release)
	<-done

	res := driveGenerateChain(t, coord, func() tea.Msg { return pollGenerateMsg{} }, chainTimeout)

	if len(res.terminal) != 0 {
		t.Errorf("a reset run must emit no terminal message, got %#v", res.terminal)
	}

	content, err := coord.Result()
	if content != "" || err != nil {
		t.Errorf("a reset run must write no result, got content=%q err=%v", content, err)
	}
	if coord.IsComplete() {
		t.Error("IsComplete should stay false after a reset run finishes")
	}
	if coord.IsRunning() {
		t.Error("the orphan's own chain should have cleared running")
	}
}

// TestGenerateCoordinator_ResetThenStartBeforeRelease is the strand case: a
// request queued after a mid-flight Reset must still run.
func TestGenerateCoordinator_ResetThenStartBeforeRelease(t *testing.T) {
	t.Parallel()

	gen := newGatedGenerator()
	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	cmdA := coord.Start(newTestGenerateConfig("orphan"))
	cmdA()
	<-gen.entered

	coord.Reset()

	if cmdB := coord.Start(newTestGenerateConfig("run-b")); cmdB != nil {
		t.Error("a Start issued while the orphan is unreaped must queue and return nil")
	}

	close(gen.release)

	res := driveGenerateChain(t, coord, func() tea.Msg { return pollGenerateMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("the queued run should deliver exactly one terminal message, got %d: %#v",
			len(res.terminal), res.terminal)
	}
	complete, ok := res.terminal[0].(screens.GenerationCompleteMsg)
	if !ok {
		t.Fatalf("expected GenerationCompleteMsg, got %T", res.terminal[0])
	}
	if complete.Content != "run-b" {
		t.Errorf("expected the queued run's content, got %q", complete.Content)
	}
}

// TestGenerateCoordinator_DeferredTerminalCommand holds the terminal command
// across a later Start and a later Reset.
func TestGenerateCoordinator_DeferredTerminalCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		disturb func(c *GenerateCoordinator)
	}{
		{
			name:    "held across a later Start",
			disturb: func(c *GenerateCoordinator) { c.Start(newTestGenerateConfig("run-b")) },
		},
		{
			name:    "held across a later Reset",
			disturb: func(c *GenerateCoordinator) { c.Reset() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockGen := &mockGenerator{
				generateFunc: func(
					_ *scanner.FileNode, _ map[string]bool, config contextgen.GenerateConfig,
				) (string, error) {
					return config.TemplateVars["TASK"], nil
				},
			}
			coord := NewGenerateCoordinator(mockGen)

			cmd := coord.Start(newTestGenerateConfig("run-a"))
			done := coord.done
			cmd()
			<-done

			terminal := coord.Poll()
			if terminal == nil {
				t.Fatal("Poll should have returned a terminal command")
			}

			tt.disturb(coord)

			msg := terminal()
			complete, ok := msg.(screens.GenerationCompleteMsg)
			if !ok {
				t.Fatalf("expected GenerationCompleteMsg, got %T", msg)
			}
			if complete.Content != "run-a" {
				t.Errorf("the held command delivered the wrong run's content: got %q, want %q",
					complete.Content, "run-a")
			}
		})
	}
}

// TestGenerateCoordinator_IsStartedVsIsRunning pins the distinction between the
// two flags: started records that a run was ever spawned and stays true, while
// running is cleared once Poll reaps the run.
func TestGenerateCoordinator_IsStartedVsIsRunning(t *testing.T) {
	t.Parallel()

	gen := newGatedGenerator()
	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	if coord.IsStarted() || coord.IsRunning() {
		t.Fatal("neither flag should be set before Start")
	}

	cmd := coord.Start(newTestGenerateConfig("run-a"))
	if coord.IsStarted() {
		t.Error("started should only be set once the command runs")
	}
	if !coord.IsRunning() {
		t.Error("running should be set by Start, so a second Start queues")
	}

	cmd()
	<-gen.entered

	if !coord.IsStarted() {
		t.Error("started should be true once the run was spawned")
	}

	close(gen.release)
	driveGenerateChain(t, coord, func() tea.Msg { return pollGenerateMsg{} }, chainTimeout)

	if !coord.IsStarted() {
		t.Error("started must stay true after the run completes")
	}
	if coord.IsRunning() {
		t.Error("running must be false after the run completes with nothing pending")
	}
}

// TestGenerateCoordinator_PollAfterReapIsInert is the generation-side twin of
// TestScanCoordinator_PollAfterReapIsInert: done stays closed on the struct, so
// only the !running guard keeps a stray Poll from emitting a second terminal
// message.
func TestGenerateCoordinator_PollAfterReapIsInert(t *testing.T) {
	t.Parallel()

	mockGen := &mockGenerator{
		generateFunc: func(
			_ *scanner.FileNode, _ map[string]bool, config contextgen.GenerateConfig,
		) (string, error) {
			return config.TemplateVars["TASK"], nil
		},
	}
	coord := NewGenerateCoordinator(mockGen)

	cmd := coord.Start(newTestGenerateConfig("run-a"))
	done := coord.done
	cmd()
	<-done

	first := coord.Poll()
	if first == nil {
		t.Fatal("the first Poll after completion should return the terminal command")
	}
	if _, ok := first().(screens.GenerationCompleteMsg); !ok {
		t.Fatal("expected the first Poll to yield GenerationCompleteMsg")
	}

	for i := range 3 {
		if extra := coord.Poll(); extra != nil {
			t.Fatalf("Poll %d after the reap returned a command (%T); the terminal message would be emitted twice",
				i+2, extra())
		}
	}
}

// TestGenerateCoordinator_ErrorRunEmitsGenerationErrorMsg covers the error path
// now that finishGenerate takes its payload by parameter.
func TestGenerateCoordinator_ErrorRunEmitsGenerationErrorMsg(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("template render failed")
	gen := newGatedGenerator()
	gen.err = wantErr

	coord := NewGenerateCoordinator(gen)
	coord.pollInterval = time.Millisecond

	cmd := coord.Start(newTestGenerateConfig("boom"))
	cmd()
	<-gen.entered
	close(gen.release)

	res := driveGenerateChain(t, coord, func() tea.Msg { return pollGenerateMsg{} }, chainTimeout)

	if len(res.terminal) != 1 {
		t.Fatalf("expected exactly one terminal message, got %d", len(res.terminal))
	}
	genErr, ok := res.terminal[0].(screens.GenerationErrorMsg)
	if !ok {
		t.Fatalf("expected GenerationErrorMsg, got %T", res.terminal[0])
	}
	if !errors.Is(genErr.Err, wantErr) {
		t.Errorf("expected the generator's error, got %v", genErr.Err)
	}
}

// TestBuildGeneratorConfig_NilTemplate guards the worker goroutine: a nil
// Template used to panic there, and a panic in a goroutine cannot be recovered
// by any tea.Cmd -- it takes the process down.
func TestBuildGeneratorConfig_NilTemplate(t *testing.T) {
	t.Parallel()

	cfg := buildGeneratorConfig(&GenerateConfig{TaskDesc: "task", Rules: "rules"})

	if cfg.Template != "" {
		t.Errorf("expected an empty template, got %q", cfg.Template)
	}
	if cfg.TemplateVars["TASK"] != "task" {
		t.Errorf("TASK not propagated: got %q", cfg.TemplateVars["TASK"])
	}
}

// TestCloneGenerateConfig_ClonesSelections pins the Selections clone, which is
// what keeps a live UI map from being read by a worker goroutine.
func TestCloneGenerateConfig_ClonesSelections(t *testing.T) {
	t.Parallel()

	if cloneGenerateConfig(nil) != nil {
		t.Error("cloneGenerateConfig(nil) should be nil")
	}

	original := newTestGenerateConfig("task")
	clone := cloneGenerateConfig(original)

	if clone == original {
		t.Fatal("clone must be a distinct struct")
	}
	clone.Selections["only-in-clone"] = true
	if original.Selections["only-in-clone"] {
		t.Error("Selections is still shared with the caller's map")
	}

	// FileTree is shared deliberately: thousands of nodes, replaced rather than
	// mutated after a scan.
	if clone.FileTree != original.FileTree {
		t.Error("FileTree should be shared, not deep-copied")
	}
}
