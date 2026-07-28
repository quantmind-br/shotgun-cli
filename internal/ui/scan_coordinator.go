package ui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

const scanPollInterval = 50 * time.Millisecond

// scanRunRequest holds an already-cloned config; the caller's pointer is never
// retained, so a later in-place mutation cannot reach a queued or running scan.
type scanRunRequest struct {
	rootPath string
	config   *scanner.ScanConfig
}

// ScanCoordinator manages the file scanning state machine.
//
// At most one scan goroutine is ever in flight. This is not a policy choice:
// FileSystemScanner owns a single mutable ignoreEngine that ScanWithProgress
// writes to (LoadGitignore, LoadShotgunignore, AddCustomRules) and reads
// throughout the walk, so two overlapping scans would blend each other's rules.
//
// The governing rule for everything below: a returned tea.Cmd captures values,
// never the coordinator. Bubble Tea runs each command on its own goroutine, so
// any field a closure dereferences is read after mu was released and after later
// messages may have replaced it.
type ScanCoordinator struct {
	mu       sync.Mutex
	scanner  scanner.Scanner
	rootPath string
	config   *scanner.ScanConfig

	// generation is the newest run's number; activeGen is the number of the run
	// whose channels currently sit on the struct. They diverge only after a Reset
	// during a live run, which is exactly the signal Poll uses to suppress a
	// stale terminal message.
	generation uint64
	activeGen  uint64

	// pending holds a request that arrived while a run was in flight. The active
	// run's own poll chain launches it on completion.
	pending *scanRunRequest

	// running means "a run is in flight and has not been reaped yet". Distinct
	// from started, which only records that a run was ever spawned. Cleared by
	// Poll, never by the worker goroutine -- see Reset.
	running bool

	progressCh chan scanner.Progress
	done       chan bool
	result     *scanner.FileNode
	scanErr    error
	started    bool

	// pollInterval is the tick delay between polls. Zero means scanPollInterval;
	// tests lower it so a poll chain can be driven without real waiting.
	pollInterval time.Duration
}

// NewScanCoordinator creates a new coordinator with the given scanner.
func NewScanCoordinator(s scanner.Scanner) *ScanCoordinator {
	return &ScanCoordinator{
		scanner: s,
	}
}

// cloneScanConfig deep-copies a ScanConfig, including IgnorePatterns and
// IncludePatterns, so an in-flight run cannot observe a later in-place mutation
// by the caller: the wizard toggles IncludeIgnored on the shared struct and then
// hands the same pointer to the next run. A shallow struct copy is not enough,
// because ScanWithProgress passes config.IgnorePatterns straight to
// AddCustomRules.
//
// nil is passed through: ScanWithProgress applies its own default, and
// duplicating that policy here would let the two drift.
func cloneScanConfig(c *scanner.ScanConfig) *scanner.ScanConfig {
	if c == nil {
		return nil
	}

	clone := *c
	clone.IgnorePatterns = append([]string(nil), c.IgnorePatterns...)
	clone.IncludePatterns = append([]string(nil), c.IncludePatterns...)

	return &clone
}

// Start clones config, then begins a scan or records it as pending when one is
// already running.
//
// Returns nil when the request was queued. The active run's existing poll chain
// performs the handoff; issuing a second chain would double-emit the terminal
// message once done is closed, because a closed channel is always ready and both
// chains would reach the terminal branch. Callers batch the result through
// tea.Batch, which discards nil commands.
//
// The returned command must be executed exactly once. running is set here rather
// than inside the command, so a caller that discards it leaves the gate closed
// with no worker to open it.
func (c *ScanCoordinator) Start(rootPath string, config *scanner.ScanConfig) tea.Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clone before the running check, not inside beginLocked: a queued request
	// must not retain the caller's pointer either.
	req := &scanRunRequest{rootPath: rootPath, config: cloneScanConfig(config)}

	if c.running {
		c.pending = req

		return nil
	}

	return c.beginLocked(req)
}

// beginLocked installs req as the active run and returns its command.
// Caller must hold mu. It must not call anything that locks.
func (c *ScanCoordinator) beginLocked(req *scanRunRequest) tea.Cmd {
	c.generation++
	c.activeGen = c.generation
	c.rootPath = req.rootPath
	c.config = req.config
	c.result = nil
	c.scanErr = nil
	c.started = false
	c.running = true

	// Locals, so the run goroutine and the terminal command never dereference
	// the struct fields these are stored in.
	progressCh := make(chan scanner.Progress, 100)
	done := make(chan bool)
	c.progressCh = progressCh
	c.done = done

	return c.scanCmd(c.activeGen, req.rootPath, req.config, progressCh, done)
}

// Poll checks for scan completion or progress and returns the appropriate
// message.
func (c *ScanCoordinator) Poll() tea.Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	// !running means there is no unreaped run: either nothing started, or a
	// previous Poll already emitted the terminal message. Without this guard a
	// second Poll would find done still closed on the struct and emit that
	// message again.
	if c.progressCh == nil || !c.running {
		return nil
	}

	// The select has a default, so it never blocks while holding mu.
	select {
	case progress, ok := <-c.progressCh:
		if !ok {
			// Unreachable in practice: nothing in the codebase closes progressCh,
			// done is the only completion signal. Handled for safety.
			return c.completeLocked()
		}
		if c.activeGen != c.generation {
			// Run was reset mid-flight. Keep draining so the worker cannot block,
			// but do not paint progress for a scan the wizard abandoned.
			return c.schedulePollLocked()
		}

		return tea.Batch(
			func() tea.Msg {
				return ScanProgressMsg{
					Current: progress.Current,
					Total:   progress.Total,
					Stage:   progress.Stage,
				}
			},
			c.schedulePollLocked(),
		)
	case <-c.done:
		return c.completeLocked()
	default:
		return c.schedulePollLocked()
	}
}

// completeLocked decides, in one atomic step, what a finished run emits.
// Caller must hold mu.
func (c *ScanCoordinator) completeLocked() tea.Cmd {
	if c.pending != nil {
		// Silent A->B handoff. running stays true, so the gate is never open
		// between the two runs, and no terminal message is emitted for A: the
		// wizard would otherwise commit the superseded tree and consume its
		// one-shot selectionsSeeded latch with it.
		req := c.pending
		c.pending = nil

		return c.beginLocked(req)
	}

	c.running = false

	if c.activeGen != c.generation {
		// Reset landed mid-flight: nobody is waiting on this run.
		return nil
	}

	// Snapshot under the same lock. Reading these when the command executes
	// would be too late -- a later Start or Reset can have cleared them.
	tree, err := c.result, c.scanErr

	return finishScan(tree, err)
}

// Result returns the scan result and any error that occurred.
func (c *ScanCoordinator) Result() (*scanner.FileNode, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.result, c.scanErr
}

// IsComplete checks if the scan has finished (either successfully or with an error).
func (c *ScanCoordinator) IsComplete() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.result != nil || c.scanErr != nil
}

// IsStarted returns true if a scan has ever been started. It stays true after a
// run completes; use IsRunning to ask whether one is in flight.
func (c *ScanCoordinator) IsStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.started
}

// IsRunning reports whether a scan is in flight and has not been reaped yet.
func (c *ScanCoordinator) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.running
}

// scanCmd owns exactly one run. Its goroutine reads no mutable coordinator
// field: the channels, root path, cloned config and generation all arrive as
// parameters.
func (c *ScanCoordinator) scanCmd(
	gen uint64,
	rootPath string,
	config *scanner.ScanConfig,
	progressCh chan scanner.Progress,
	done chan bool,
) tea.Cmd {
	// Guards against a caller invoking the same command twice, which would spawn
	// two workers and close(done) twice -- a panic.
	var once sync.Once

	return func() tea.Msg {
		once.Do(func() {
			c.mu.Lock()
			c.started = true
			c.mu.Unlock()

			go func() {
				// Registered first, so it runs last: the result write below is
				// therefore visible to anyone who observed done being closed.
				defer close(done)

				tree, err := c.scanner.ScanWithProgress(rootPath, config, progressCh)

				c.mu.Lock()
				if gen == c.generation {
					c.result = tree
					c.scanErr = err
				}
				c.mu.Unlock()
			}()
		})

		return pollScanMsg{}
	}
}

// finishScan closes over a snapshot taken while Poll held mu. It must never
// dereference the coordinator: Bubble Tea executes this command on its own
// goroutine, by which time a later Start or Reset may have cleared or replaced
// result and scanErr.
func finishScan(tree *scanner.FileNode, err error) tea.Cmd {
	return func() tea.Msg {
		if err != nil {
			return ScanErrorMsg{Err: err}
		}

		return ScanCompleteMsg{Tree: tree}
	}
}

// schedulePollLocked returns the next tick. Caller must hold mu; tea.Tick only
// builds a command here, it does not wait.
func (c *ScanCoordinator) schedulePollLocked() tea.Cmd {
	interval := c.pollInterval
	if interval <= 0 {
		interval = scanPollInterval
	}

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return pollScanMsg{}
	})
}

// Reset clears coordinator state and increments generation, so a run still in
// flight can no longer write its result and its terminal message is suppressed.
//
// While a run is in flight it deliberately leaves running, progressCh, done and
// activeGen intact. Clearing running would let the next Start spawn a concurrent
// walk against the scanner's shared ignoreEngine. Clearing the channels would
// make the next Poll return nil and kill the live chain -- and because a
// post-reset Start queues into pending and returns nil, nothing would be left to
// launch it: the request would strand and the wizard would wait forever. The
// orphan's own chain performs that handoff, then clears running.
//
// "In flight" means unreaped, not merely still executing: running is cleared by
// Poll, never by the worker. A Reset issued after the worker returned but before
// any Poll therefore still takes the in-flight branch, deliberately -- a
// scheduled tick may still sit in Bubble Tea's queue, and nil'ing the channels
// would kill that chain.
//
// With no run in flight it clears progressCh and done as before.
func (c *ScanCoordinator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	c.pending = nil
	c.rootPath = ""
	c.config = nil
	c.result = nil
	c.scanErr = nil
	c.started = false

	if !c.running {
		c.progressCh = nil
		c.done = nil
	}
}
