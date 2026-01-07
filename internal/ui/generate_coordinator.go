package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/context"
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

type GenerateCoordinator struct {
	generator  context.ContextGenerator
	config     *GenerateConfig
	progressCh chan context.GenProgress
	done       chan bool
	content    string
	genErr     error
	started    bool
}

func NewGenerateCoordinator(gen context.ContextGenerator) *GenerateCoordinator {
	return &GenerateCoordinator{generator: gen}
}

func (c *GenerateCoordinator) Start(cfg *GenerateConfig) tea.Cmd {
	c.config = cfg
	c.progressCh = make(chan context.GenProgress, 100)
	c.done = make(chan bool)
	c.started = false
	c.content = ""
	c.genErr = nil

	return c.iterativeGenerateCmd()
}

func (c *GenerateCoordinator) Poll() tea.Cmd {
	if c.progressCh == nil {
		return nil
	}

	select {
	case progress, ok := <-c.progressCh:
		if !ok {
			return c.finishGenerate()
		}
		return tea.Batch(
			func() tea.Msg {
				return GenerationProgressMsg{
					Stage:   progress.Stage,
					Message: progress.Message,
				}
			},
			c.schedulePoll(),
		)
	case <-c.done:
		return c.finishGenerate()
	default:
		return c.schedulePoll()
	}
}

func (c *GenerateCoordinator) Result() (string, error) {
	return c.content, c.genErr
}

func (c *GenerateCoordinator) IsComplete() bool {
	return c.content != "" || c.genErr != nil
}

func (c *GenerateCoordinator) IsStarted() bool {
	return c.started
}

func (c *GenerateCoordinator) iterativeGenerateCmd() tea.Cmd {
	return func() tea.Msg {
		if !c.started {
			c.started = true
			go func() {
				defer close(c.done)

				genConfig := *c.buildGeneratorConfig()
				content, err := c.generator.Generate(
					c.config.FileTree,
					c.config.Selections,
					genConfig,
				)
				c.content = content
				c.genErr = err
			}()
		}

		return pollGenerateMsg{}
	}
}

func (c *GenerateCoordinator) finishGenerate() tea.Cmd {
	return func() tea.Msg {
		if c.genErr != nil {
			return screens.GenerationErrorMsg{Err: c.genErr}
		}
		return screens.GenerationCompleteMsg{Content: c.content}
	}
}

func (c *GenerateCoordinator) schedulePoll() tea.Cmd {
	return tea.Tick(generatePollInterval, func(time.Time) tea.Msg {
		return pollGenerateMsg{}
	})
}

func (c *GenerateCoordinator) buildGeneratorConfig() *context.GenerateConfig {
	return &context.GenerateConfig{
		TemplateVars: map[string]string{
			"TASK":           c.config.TaskDesc,
			"RULES":          c.config.Rules,
			"FILE_STRUCTURE": "",
			"CURRENT_DATE":   time.Now().Format("2006-01-02"),
		},
		Template:       c.config.Template.Content,
		IncludeTree:    c.config.IncludeTree,
		IncludeSummary: c.config.IncludeSummary,
	}
}

func (c *GenerateCoordinator) Reset() {
	c.config = nil
	c.progressCh = nil
	c.done = nil
	c.content = ""
	c.genErr = nil
	c.started = false
}
