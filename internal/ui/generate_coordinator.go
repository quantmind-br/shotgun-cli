package ui

import (
	"maps"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
)

const generatePollInterval = 50 * time.Millisecond

type GenerateConfig struct {
	FileTree       *scanner.FileNode
	Selections     map[string]bool
	Template       *template.Template
	TaskDesc       string
	Rules          string
	RootPath       string
	MaxFileSize    int64
	MaxTotalSize   int64
	MaxFiles       int
	IncludeTree    bool
	IncludeSummary bool
}

// GenerateCoordinator manages the context generation state machine. It mirrors
// ScanCoordinator: single-flight, per-run channels captured as locals, and
// commands that close over snapshots rather than the coordinator. See
// scan_coordinator.go for the reasoning behind each field.
type GenerateCoordinator struct {
	mu        sync.Mutex
	generator contextgen.ContextGenerator
	config    *GenerateConfig

	generation uint64
	activeGen  uint64
	pending    *GenerateConfig
	running    bool

	progressCh chan contextgen.GenProgress
	done       chan bool
	content    string
	genErr     error
	started    bool

	// pollInterval is the tick delay between polls. Zero means
	// generatePollInterval; tests lower it to drive a chain without waiting.
	pollInterval time.Duration
}

func NewGenerateCoordinator(gen contextgen.ContextGenerator) *GenerateCoordinator {
	return &GenerateCoordinator{generator: gen}
}

// cloneGenerateConfig copies the request so an in-flight run cannot observe a
// later mutation by the caller.
//
// Selections is cloned because FileSelectionModel.GetSelections returns the
// screen's live map, not a copy, and the screen mutates it in place on every
// toggle. A concurrent map read/write in Go is a fatal, unrecoverable error, and
// the clone costs O(selected files) next to reading all of them.
//
// FileTree is shared by design: it is thousands of nodes, and after a scan it is
// replaced wholesale rather than mutated.
func cloneGenerateConfig(cfg *GenerateConfig) *GenerateConfig {
	if cfg == nil {
		return nil
	}

	clone := *cfg
	clone.Selections = maps.Clone(cfg.Selections)

	return &clone
}

// Start clones cfg, then begins generation or records it as pending when one is
// already running. Returns nil when the request was queued; see
// ScanCoordinator.Start for why a second poll chain must not be issued.
func (c *GenerateCoordinator) Start(cfg *GenerateConfig) tea.Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	req := cloneGenerateConfig(cfg)

	if c.running {
		c.pending = req

		return nil
	}

	return c.beginLocked(req)
}

// beginLocked installs req as the active run. Caller must hold mu.
func (c *GenerateCoordinator) beginLocked(req *GenerateConfig) tea.Cmd {
	c.generation++
	c.activeGen = c.generation
	c.config = req
	c.content = ""
	c.genErr = nil
	c.started = false
	c.running = true

	progressCh := make(chan contextgen.GenProgress, 100)
	done := make(chan bool)
	c.progressCh = progressCh
	c.done = done

	// progressCh is stored but never written: the coordinator calls
	// generator.Generate, which reports no progress. It still gates Poll (a nil
	// progressCh means "no run"), so it is created per run like the scan side.
	_ = progressCh

	return c.generateCmd(c.activeGen, req, done)
}

func (c *GenerateCoordinator) Poll() tea.Cmd {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.progressCh == nil || !c.running {
		return nil
	}

	select {
	case progress, ok := <-c.progressCh:
		if !ok {
			return c.completeLocked()
		}
		if c.activeGen != c.generation {
			return c.schedulePollLocked()
		}

		return tea.Batch(
			func() tea.Msg {
				return GenerationProgressMsg{
					Stage:   progress.Stage,
					Message: progress.Message,
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

// completeLocked decides what a finished run emits. Caller must hold mu.
func (c *GenerateCoordinator) completeLocked() tea.Cmd {
	if c.pending != nil {
		req := c.pending
		c.pending = nil

		return c.beginLocked(req)
	}

	c.running = false

	if c.activeGen != c.generation {
		return nil
	}

	content, err := c.content, c.genErr

	return finishGenerate(content, err)
}

func (c *GenerateCoordinator) Result() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.content, c.genErr
}

func (c *GenerateCoordinator) IsComplete() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.content != "" || c.genErr != nil
}

// IsStarted returns true if generation has ever been started.
func (c *GenerateCoordinator) IsStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.started
}

// IsRunning reports whether generation is in flight and has not been reaped yet.
func (c *GenerateCoordinator) IsRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.running
}

// generateCmd owns exactly one run and reads no mutable coordinator field.
func (c *GenerateCoordinator) generateCmd(
	gen uint64,
	cfg *GenerateConfig,
	done chan bool,
) tea.Cmd {
	var once sync.Once

	return func() tea.Msg {
		once.Do(func() {
			c.mu.Lock()
			c.started = true
			c.mu.Unlock()

			go func() {
				// Registered first, so it runs after the content write below.
				defer close(done)

				content, err := c.generator.Generate(
					cfg.FileTree,
					cfg.Selections,
					buildGeneratorConfig(cfg),
				)

				c.mu.Lock()
				if gen == c.generation {
					c.content = content
					c.genErr = err
				}
				c.mu.Unlock()
			}()
		})

		return pollGenerateMsg{}
	}
}

// finishGenerate closes over a snapshot taken while Poll held mu, for the same
// reason finishScan does.
func finishGenerate(content string, err error) tea.Cmd {
	return func() tea.Msg {
		if err != nil {
			return screens.GenerationErrorMsg{Err: err}
		}

		return screens.GenerationCompleteMsg{Content: content}
	}
}

func (c *GenerateCoordinator) schedulePollLocked() tea.Cmd {
	interval := c.pollInterval
	if interval <= 0 {
		interval = generatePollInterval
	}

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return pollGenerateMsg{}
	})
}

// buildGeneratorConfig derives the core generator config from a run's own
// request. It takes cfg as a parameter so the run goroutine never dereferences
// c.config.
//
// A nil Template is tolerated: the wizard rejects one before it ever gets here,
// but a panic inside the worker goroutine cannot be recovered by any tea.Cmd and
// would take the whole process down.
func buildGeneratorConfig(cfg *GenerateConfig) contextgen.GenerateConfig {
	var templateContent string
	if cfg.Template != nil {
		templateContent = cfg.Template.Content
	}

	return contextgen.GenerateConfig{
		TemplateVars: map[string]string{
			"TASK":           cfg.TaskDesc,
			"RULES":          cfg.Rules,
			"FILE_STRUCTURE": "",
			"CURRENT_DATE":   time.Now().Format("2006-01-02"),
		},
		Template:       templateContent,
		IncludeTree:    cfg.IncludeTree,
		IncludeSummary: cfg.IncludeSummary,
	}
}

// Reset clears coordinator state and increments generation. While a run is
// unreaped it leaves running, progressCh, done and activeGen intact; see
// ScanCoordinator.Reset for why.
func (c *GenerateCoordinator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.generation++
	c.pending = nil
	c.config = nil
	c.content = ""
	c.genErr = nil
	c.started = false

	if !c.running {
		c.progressCh = nil
		c.done = nil
	}
}
