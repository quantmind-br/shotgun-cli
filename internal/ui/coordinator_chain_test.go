package ui

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

// chainResult is what driving a poll chain to exhaustion produced.
type chainResult struct {
	// terminal holds every ScanCompleteMsg/ScanErrorMsg (or the generation
	// equivalents) the chain emitted. The contract is exactly one, or zero for a
	// superseded or reset run.
	terminal []tea.Msg
	// progress counts progress messages, to show a reset run stays quiet.
	progress int
	// maxPollsPerStep is the largest number of poll-producing commands the chain
	// yielded from a single step. A rival chain shows up here as 2; it is
	// invisible to a goroutine count, because Bubble Tea drives every Poll on
	// its single update goroutine.
	maxPollsPerStep int
}

// driveScanChain executes a poll chain the way Bubble Tea would: run a command,
// dispatch the message it returns, repeat. tea.Batch is expanded rather than
// executed, matching the real runtime.
//
// extra commands (e.g. a queued Start's return value) can be injected; nil ones
// are dropped, exactly as tea.Batch does.
func driveScanChain(t *testing.T, c *ScanCoordinator, first tea.Cmd, timeout time.Duration) chainResult {
	t.Helper()

	var res chainResult
	queue := []tea.Cmd{first}
	deadline := time.Now().Add(timeout)

	for len(queue) > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("poll chain did not settle within %s", timeout)
		}

		cmd := queue[0]
		queue = queue[1:]
		if cmd == nil {
			continue
		}

		pollsThisStep := 0
		switch msg := cmd().(type) {
		case tea.BatchMsg:
			queue = append(queue, msg...)
		case pollScanMsg:
			next := c.Poll()
			if next != nil {
				pollsThisStep++
				queue = append(queue, next)
			}
		case ScanCompleteMsg:
			res.terminal = append(res.terminal, msg)
		case ScanErrorMsg:
			res.terminal = append(res.terminal, msg)
		case ScanProgressMsg:
			res.progress++
		}

		if pollsThisStep > res.maxPollsPerStep {
			res.maxPollsPerStep = pollsThisStep
		}
	}

	return res
}

// driveGenerateChain is the generation-side twin of driveScanChain.
func driveGenerateChain(t *testing.T, c *GenerateCoordinator, first tea.Cmd, timeout time.Duration) chainResult {
	t.Helper()

	var res chainResult
	queue := []tea.Cmd{first}
	deadline := time.Now().Add(timeout)

	for len(queue) > 0 {
		if time.Now().After(deadline) {
			t.Fatalf("poll chain did not settle within %s", timeout)
		}

		cmd := queue[0]
		queue = queue[1:]
		if cmd == nil {
			continue
		}

		pollsThisStep := 0
		switch msg := cmd().(type) {
		case tea.BatchMsg:
			queue = append(queue, msg...)
		case pollGenerateMsg:
			next := c.Poll()
			if next != nil {
				pollsThisStep++
				queue = append(queue, next)
			}
		case screens.GenerationCompleteMsg:
			res.terminal = append(res.terminal, msg)
		case screens.GenerationErrorMsg:
			res.terminal = append(res.terminal, msg)
		case GenerationProgressMsg:
			res.progress++
		}

		if pollsThisStep > res.maxPollsPerStep {
			res.maxPollsPerStep = pollsThisStep
		}
	}

	return res
}

// gatedScanner blocks inside the walk until the test releases it, and never
// writes to the progress channel. That last part matters: a test that stops
// polling would otherwise risk the worker blocking on a full channel and never
// closing done.
type gatedScanner struct {
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	tree    func(rootPath string) *scanner.FileNode
	err     error
}

func newGatedScanner(tree func(rootPath string) *scanner.FileNode) *gatedScanner {
	return &gatedScanner{
		entered: make(chan struct{}),
		release: make(chan struct{}),
		tree:    tree,
	}
}

func (g *gatedScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
	return g.ScanWithProgress(rootPath, config, nil)
}

func (g *gatedScanner) ScanWithProgress(
	rootPath string, _ *scanner.ScanConfig, _ chan<- scanner.Progress,
) (*scanner.FileNode, error) {
	g.once.Do(func() { close(g.entered) })
	<-g.release

	if g.err != nil {
		return nil, g.err
	}

	return g.tree(rootPath), nil
}

// countingScanner wraps a real scanner and records the high-water mark of
// concurrently live walks. The shared ignoreEngine cannot tolerate more than one.
type countingScanner struct {
	inner scanner.Scanner
	live  atomic.Int32
	max   atomic.Int32
}

func (s *countingScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
	return s.ScanWithProgress(rootPath, config, nil)
}

func (s *countingScanner) ScanWithProgress(
	rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress,
) (*scanner.FileNode, error) {
	n := s.live.Add(1)
	for {
		old := s.max.Load()
		if n <= old || s.max.CompareAndSwap(old, n) {
			break
		}
	}
	defer s.live.Add(-1)

	return s.inner.ScanWithProgress(rootPath, config, progress)
}

// configProbeScanner records the IncludeIgnored it was handed, before and after
// the test mutates the caller's struct. Under -race the mutation is itself the
// oracle: without the clone, the test goroutine's write races this read.
type configProbeScanner struct {
	inner   scanner.Scanner
	entered chan struct{}
	release chan struct{}
	once    sync.Once
	seen    [2]bool
}

func (s *configProbeScanner) Scan(rootPath string, config *scanner.ScanConfig) (*scanner.FileNode, error) {
	return s.ScanWithProgress(rootPath, config, nil)
}

func (s *configProbeScanner) ScanWithProgress(
	rootPath string, config *scanner.ScanConfig, progress chan<- scanner.Progress,
) (*scanner.FileNode, error) {
	s.seen[0] = config.IncludeIgnored
	s.once.Do(func() { close(s.entered) })
	<-s.release
	s.seen[1] = config.IncludeIgnored

	return s.inner.ScanWithProgress(rootPath, config, progress)
}
