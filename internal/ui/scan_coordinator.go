package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

const scanPollInterval = 50 * time.Millisecond

// ScanCoordinator manages the file scanning state machine.
type ScanCoordinator struct {
	scanner    scanner.Scanner
	rootPath   string
	config     *scanner.ScanConfig
	progressCh chan scanner.Progress
	done       chan bool
	result     *scanner.FileNode
	scanErr    error
	started    bool
}

// NewScanCoordinator creates a new coordinator with the given scanner.
func NewScanCoordinator(s scanner.Scanner) *ScanCoordinator {
	return &ScanCoordinator{
		scanner: s,
	}
}

// Start begins the scan process and returns a command to start polling.
func (c *ScanCoordinator) Start(rootPath string, config *scanner.ScanConfig) tea.Cmd {
	c.rootPath = rootPath
	c.config = config
	c.progressCh = make(chan scanner.Progress, 100)
	c.done = make(chan bool)
	c.started = false
	c.result = nil
	c.scanErr = nil

	return c.iterativeScanCmd()
}

// Poll checks for scan completion or progress and returns the appropriate message.
func (c *ScanCoordinator) Poll() tea.Cmd {
	if c.progressCh == nil {
		return nil
	}

	select {
	case progress, ok := <-c.progressCh:
		if !ok {
			return c.finishScan()
		}
		return tea.Batch(
			func() tea.Msg {
				return ScanProgressMsg{
					Current: progress.Current,
					Total:   progress.Total,
					Stage:   progress.Stage,
				}
			},
			c.schedulePoll(),
		)
	case <-c.done:
		return c.finishScan()
	default:
		return c.schedulePoll()
	}
}

// Result returns the scan result and any error that occurred.
func (c *ScanCoordinator) Result() (*scanner.FileNode, error) {
	return c.result, c.scanErr
}

// IsComplete checks if the scan has finished (either successfully or with an error).
func (c *ScanCoordinator) IsComplete() bool {
	return c.result != nil || c.scanErr != nil
}

// IsStarted returns true if the scan has been started.
func (c *ScanCoordinator) IsStarted() bool {
	return c.started
}

func (c *ScanCoordinator) iterativeScanCmd() tea.Cmd {
	return func() tea.Msg {
		if !c.started {
			c.started = true
			go func() {
				defer close(c.done)
				tree, err := c.scanner.ScanWithProgress(
					c.rootPath,
					c.config,
					c.progressCh,
				)
				c.result = tree
				c.scanErr = err
			}()
		}

		return pollScanMsg{}
	}
}

func (c *ScanCoordinator) finishScan() tea.Cmd {
	return func() tea.Msg {
		if c.scanErr != nil {
			return ScanErrorMsg{Err: c.scanErr}
		}
		return ScanCompleteMsg{Tree: c.result}
	}
}

func (c *ScanCoordinator) schedulePoll() tea.Cmd {
	return tea.Tick(scanPollInterval, func(time.Time) tea.Msg {
		return pollScanMsg{}
	})
}

// Reset clears the coordinator state for reuse.
func (c *ScanCoordinator) Reset() {
	c.rootPath = ""
	c.config = nil
	c.progressCh = nil
	c.done = nil
	c.result = nil
	c.scanErr = nil
	c.started = false
}
