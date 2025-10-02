package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/platform/clipboard"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
	"github.com/spf13/viper"
)

const (
	StepFileSelection     = 1
	StepTemplateSelection = 2
	StepTaskInput         = 3
	StepRulesInput        = 4
	StepReview            = 5
)

type Progress struct {
	Current int64
	Total   int64
	Stage   string
	Message string
	Visible bool
}

type WizardModel struct {
	step          int
	fileTree      *scanner.FileNode
	selectedFiles map[string]bool
	template      *template.Template
	taskDesc      string
	rules         string
	progress      Progress
	error         error
	width         int
	height        int
	showHelp      bool

	rootPath string
	config   *scanner.ScanConfig

	fileSelection     *screens.FileSelectionModel
	templateSelection *screens.TemplateSelectionModel
	taskInput         *screens.TaskInputModel
	rulesInput        *screens.RulesInputModel
	review            *screens.ReviewModel

	progressComponent *components.ProgressModel

	// State for iterative commands
	scanState     *scanState
	generateState *generateState

	// Generation result storage
	generatedFilePath string
}

type ScanProgressMsg struct {
	Current int64
	Total   int64
	Stage   string
}

type ScanCompleteMsg struct {
	Tree *scanner.FileNode
}

type ScanErrorMsg struct {
	Err error
}

type TemplateSelectedMsg struct {
	Template *template.Template
}

type GenerationProgressMsg struct {
	Stage   string
	Message string
}

type GenerationCompleteMsg struct {
	Content  string
	FilePath string
}

type GenerationErrorMsg struct {
	Err error
}

type ClipboardCompleteMsg struct {
	Success bool
	Err     error
}

// Internal state for iterative commands
type scanState struct {
	scanner    *scanner.FileSystemScanner
	rootPath   string
	config     *scanner.ScanConfig
	progressCh chan scanner.Progress
	done       chan bool
	started    bool
}

type generateState struct {
	generator  context.ContextGenerator
	fileTree   *scanner.FileNode
	config     *context.GenerateConfig
	rootPath   string
	progressCh chan context.GenProgress
	done       chan bool
	started    bool
	content    string
}

// New message types for iterative commands
type startScanMsg struct {
	rootPath string
	config   *scanner.ScanConfig
}

type startGenerationMsg struct {
	fileTree      *scanner.FileNode
	selectedFiles map[string]bool
	template      *template.Template
	taskDesc      string
	rules         string
	rootPath      string
}

func NewWizard(rootPath string, config *scanner.ScanConfig) *WizardModel {
	return &WizardModel{
		step:              StepFileSelection,
		selectedFiles:     make(map[string]bool),
		rootPath:          rootPath,
		config:            config,
		progressComponent: components.NewProgress(),
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return scanDirectoryCmd(m.rootPath, m.config)
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.fileSelection != nil {
			m.fileSelection.SetSize(m.width, m.height)
		}
		if m.templateSelection != nil {
			m.templateSelection.SetSize(m.width, m.height)
		}
		if m.taskInput != nil {
			m.taskInput.SetSize(m.width, m.height)
		}
		if m.rulesInput != nil {
			m.rulesInput.SetSize(m.width, m.height)
		}
		if m.review != nil {
			m.review.SetSize(m.width, m.height)
		}

		return m, nil

	case tea.KeyMsg:
		// Always handle quit commands
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		}

		// Process navigation shortcuts ALWAYS (regardless of focus state)
		switch msg.String() {
		case "f8", "ctrl+pgdn":
			if m.step < StepReview {
				if m.canAdvanceStep() {
					m.step++
					cmd = m.initStep()
					cmds = append(cmds, cmd)
				}
			} else if m.step == StepReview {
				cmd = m.generateContext()
				cmds = append(cmds, cmd)
			}
		case "f10", "ctrl+pgup":
			if m.step > StepFileSelection {
				m.step--
				cmd = m.initStep()
				cmds = append(cmds, cmd)
			}
		case "f1":
			m.showHelp = !m.showHelp
		default:
			// For all other keys, pass to the current step
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
		}

	case ScanProgressMsg:
		m.progress = Progress{
			Current: msg.Current,
			Total:   msg.Total,
			Stage:   msg.Stage,
			Visible: true,
		}
		m.progressComponent.Update(msg.Current, msg.Total, msg.Stage, "")
		// Re-enqueue the iterative scan command
		if m.scanState != nil {
			cmd = m.iterativeScanCmd()
			cmds = append(cmds, cmd)
		}

	case ScanCompleteMsg:
		m.fileTree = msg.Tree
		m.progress.Visible = false
		m.fileSelection = screens.NewFileSelection(msg.Tree, m.selectedFiles)
		m.fileSelection.SetSize(m.width, m.height)

	case ScanErrorMsg:
		m.error = msg.Err
		m.progress.Visible = false

	case TemplateSelectedMsg:
		m.template = msg.Template

	case GenerationProgressMsg:
		m.progress = Progress{
			Stage:   msg.Stage,
			Message: msg.Message,
			Visible: true,
		}
		m.progressComponent.UpdateMessage(msg.Stage, msg.Message)
		// Re-enqueue the iterative generate command
		if m.generateState != nil {
			cmd = m.iterativeGenerateCmd()
			cmds = append(cmds, cmd)
		}

	case GenerationCompleteMsg:
		m.progress.Visible = false
		m.generatedFilePath = msg.FilePath
		cmd = m.clipboardCopyCmd(msg.Content)
		cmds = append(cmds, cmd)

	case GenerationErrorMsg:
		m.error = msg.Err
		m.progress.Visible = false

	case ClipboardCompleteMsg:
		// Don't set error, just log the clipboard failure
		// Update review screen with generation status (will show warning if failed)
		if m.review != nil && m.generatedFilePath != "" {
			m.review.SetGenerated(m.generatedFilePath, msg.Success)
		}

	// Handle template messages for TemplateSelection step
	case screens.TemplatesLoadedMsg, screens.TemplatesErrorMsg:
		if m.step == StepTemplateSelection && m.templateSelection != nil {
			cmd = m.templateSelection.HandleMessage(msg)
			cmds = append(cmds, cmd)
		}

	// Handle iterative command messages
	case startScanMsg:
		m.scanState = &scanState{
			scanner:    scanner.NewFileSystemScanner(),
			rootPath:   msg.rootPath,
			config:     msg.config,
			progressCh: make(chan scanner.Progress, 100),
			done:       make(chan bool),
			started:    false,
		}
		cmd = m.iterativeScanCmd()
		cmds = append(cmds, cmd)

	case startGenerationMsg:
		m.generateState = &generateState{
			generator: context.NewDefaultContextGenerator(),
			fileTree:  msg.fileTree,
			config: &context.GenerateConfig{
				TemplateVars: map[string]string{
					"TASK":           msg.taskDesc,
					"RULES":          msg.rules,
					"FILE_STRUCTURE": "",
					"CURRENT_DATE":   time.Now().Format("2006-01-02"),
				},
				Template: msg.template.Content,
			},
			rootPath:   msg.rootPath,
			progressCh: make(chan context.GenProgress, 100),
			done:       make(chan bool),
			started:    false,
		}
		cmd = m.iterativeGenerateCmd()
		cmds = append(cmds, cmd)

	case screens.RescanRequestMsg:
		// Trigger rescan when in file selection step
		if m.step == StepFileSelection {
			cmd = scanDirectoryCmd(m.rootPath, m.config)
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, cmd
}

func (m *WizardModel) View() string {
	var mainView string

	if m.error != nil {
		mainView = styles.RenderError(m.error.Error())
	} else {
		switch m.step {
		case StepFileSelection:
			if m.fileSelection != nil {
				mainView = m.fileSelection.View()
			} else {
				mainView = "Loading files..."
			}
		case StepTemplateSelection:
			if m.templateSelection != nil {
				mainView = m.templateSelection.View()
			} else {
				mainView = "Loading templates..."
			}
		case StepTaskInput:
			if m.taskInput != nil {
				mainView = m.taskInput.View()
			} else {
				mainView = "Initializing task input..."
			}
		case StepRulesInput:
			if m.rulesInput != nil {
				mainView = m.rulesInput.View()
			} else {
				mainView = "Initializing rules input..."
			}
		case StepReview:
			if m.review != nil {
				mainView = m.review.View()
			} else {
				mainView = "Initializing review..."
			}
		default:
			mainView = "Unknown step"
		}
	}

	// Overlay progress if visible
	if m.progress.Visible {
		mainView += "\n" + m.progressComponent.View()
	}

	return mainView
}

func (m *WizardModel) canAdvanceStep() bool {
	switch m.step {
	case StepFileSelection:
		return len(m.selectedFiles) > 0
	case StepTemplateSelection:
		return m.template != nil
	case StepTaskInput:
		return len(strings.TrimSpace(m.taskDesc)) > 0
	case StepRulesInput:
		return true
	case StepReview:
		return true
	default:
		return false
	}
}

func (m *WizardModel) initStep() tea.Cmd {
	switch m.step {
	case StepFileSelection:
		if m.fileTree != nil {
			m.fileSelection = screens.NewFileSelection(m.fileTree, m.selectedFiles)
			m.fileSelection.SetSize(m.width, m.height)
		}
	case StepTemplateSelection:
		m.templateSelection = screens.NewTemplateSelection()
		m.templateSelection.SetSize(m.width, m.height)
		return m.templateSelection.LoadTemplates()
	case StepTaskInput:
		m.taskInput = screens.NewTaskInput(m.taskDesc)
		m.taskInput.SetSize(m.width, m.height)
	case StepRulesInput:
		m.rulesInput = screens.NewRulesInput(m.rules)
		m.rulesInput.SetSize(m.width, m.height)
	case StepReview:
		m.review = screens.NewReview(m.selectedFiles, m.template, m.taskDesc, m.rules)
		m.review.SetSize(m.width, m.height)
	}
	return nil
}

func (m *WizardModel) handleStepInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	switch m.step {
	case StepFileSelection:
		if m.fileSelection != nil {
			cmd = m.fileSelection.Update(msg, m.selectedFiles)
		}
	case StepTemplateSelection:
		if m.templateSelection != nil {
			template, updateCmd := m.templateSelection.Update(msg)
			if template != nil {
				m.template = template
			}
			cmd = updateCmd
		}
	case StepTaskInput:
		if m.taskInput != nil {
			task, updateCmd := m.taskInput.Update(msg)
			m.taskDesc = task
			cmd = updateCmd
		}
	case StepRulesInput:
		if m.rulesInput != nil {
			rules, updateCmd := m.rulesInput.Update(msg)
			m.rules = rules
			cmd = updateCmd
		}
	case StepReview:
		if m.review != nil {
			cmd = m.review.Update(msg)
		}
	}

	return cmd
}

func (m *WizardModel) generateContext() tea.Cmd {
	if m.template == nil || len(m.selectedFiles) == 0 {
		return func() tea.Msg {
			return GenerationErrorMsg{Err: fmt.Errorf("missing template or files")}
		}
	}

	return generateContextCmd(m.fileTree, m.selectedFiles, m.template, m.taskDesc, m.rules, m.rootPath)
}

func scanDirectoryCmd(rootPath string, config *scanner.ScanConfig) tea.Cmd {
	return func() tea.Msg {
		return startScanMsg{
			rootPath: rootPath,
			config:   config,
		}
	}
}

func generateContextCmd(fileTree *scanner.FileNode, selectedFiles map[string]bool, template *template.Template, taskDesc, rules, rootPath string) tea.Cmd {
	return func() tea.Msg {
		return startGenerationMsg{
			fileTree:      fileTree,
			selectedFiles: selectedFiles,
			template:      template,
			taskDesc:      taskDesc,
			rules:         rules,
			rootPath:      rootPath,
		}
	}
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func (m *WizardModel) clipboardCopyCmd(content string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.Copy(content)
		return ClipboardCompleteMsg{
			Success: err == nil,
			Err:     err,
		}
	}
}

// parseSize converts size strings like "10MB" to bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "B") {
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return size * multiplier, nil
}

// Iterative command functions
func (m *WizardModel) iterativeScanCmd() tea.Cmd {
	return func() tea.Msg {
		if m.scanState == nil {
			return ScanErrorMsg{Err: fmt.Errorf("scan state not initialized")}
		}

		// Start the goroutine if not already started
		if !m.scanState.started {
			m.scanState.started = true
			go func() {
				defer close(m.scanState.done)
				_, err := m.scanState.scanner.ScanWithProgress(
					m.scanState.rootPath,
					m.scanState.config,
					m.scanState.progressCh,
				)
				if err != nil {
					return
				}

				// Signal completion and send final result
				select {
				case <-m.scanState.done:
					return
				default:
					// Use a separate channel or mechanism to send completion
					go func() {
						time.Sleep(10 * time.Millisecond) // Small delay to allow progress messages
						m.scanState.progressCh <- scanner.Progress{
							Current: -1, // Special signal for completion
							Total:   -1,
							Stage:   "complete",
						}
					}()
				}
			}()
		}

		// Read one progress update
		select {
		case progress, ok := <-m.scanState.progressCh:
			if !ok || (progress.Current == -1 && progress.Total == -1) {
				// Channel closed or completion signal
				<-m.scanState.done
				if tree, err := m.scanState.scanner.ScanWithProgress(m.scanState.rootPath, m.scanState.config, nil); err != nil {
					return ScanErrorMsg{Err: err}
				} else {
					return ScanCompleteMsg{Tree: tree}
				}
			}
			// Send progress and re-enqueue
			return ScanProgressMsg{
				Current: progress.Current,
				Total:   progress.Total,
				Stage:   progress.Stage,
			}
		case <-m.scanState.done:
			// Completed
			if tree, err := m.scanState.scanner.ScanWithProgress(m.scanState.rootPath, m.scanState.config, nil); err != nil {
				return ScanErrorMsg{Err: err}
			} else {
				return ScanCompleteMsg{Tree: tree}
			}
		default:
			// No progress yet, re-enqueue
			return m.iterativeScanCmd()()
		}
	}
}

func (m *WizardModel) iterativeGenerateCmd() tea.Cmd {
	return func() tea.Msg {
		if m.generateState == nil {
			return GenerationErrorMsg{Err: fmt.Errorf("generation state not initialized")}
		}

		// Start the goroutine if not already started
		if !m.generateState.started {
			m.generateState.started = true
			go func() {
				defer close(m.generateState.done)

				content, err := m.generateState.generator.GenerateWithProgressEx(
					m.generateState.fileTree,
					*m.generateState.config,
					func(p context.GenProgress) {
						m.generateState.progressCh <- p
					},
				)
				if err != nil {
					return
				}

				// Store the content for later use
				m.generateState.content = content

				// Signal completion
				go func() {
					time.Sleep(10 * time.Millisecond)
					m.generateState.progressCh <- context.GenProgress{
						Stage:   "complete",
						Message: "Generation complete",
					}
				}()
			}()
		}

		// Read one progress update
		select {
		case progress, ok := <-m.generateState.progressCh:
			if !ok || progress.Stage == "complete" {
				<-m.generateState.done

				// Use the stored content
				content := m.generateState.content
				if content == "" {
					return GenerationErrorMsg{Err: fmt.Errorf("no content generated")}
				}

				// Check size limit
				maxSizeStr := viper.GetString("context.max-size")
				if maxSizeStr != "" {
					maxSize, err := parseSize(maxSizeStr)
					if err != nil {
						return GenerationErrorMsg{Err: fmt.Errorf("invalid max-size configuration: %v", err)}
					}

					contentSize := int64(len(content))
					if contentSize > maxSize {
						return GenerationErrorMsg{Err: fmt.Errorf("generated content size (%d bytes) exceeds maximum allowed size (%d bytes)", contentSize, maxSize)}
					}
				}

				timestamp := time.Now().Format("20060102-150405")
				filename := fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
				filePath := filepath.Join(m.generateState.rootPath, filename)

				if err := writeFile(filePath, content); err != nil {
					return GenerationErrorMsg{Err: err}
				}

				return GenerationCompleteMsg{
					Content:  content,
					FilePath: filePath,
				}
			}
			// Send progress and re-enqueue
			return GenerationProgressMsg{
				Stage:   progress.Stage,
				Message: progress.Message,
			}
		case <-m.generateState.done:
			// Completed - use stored content
			content := m.generateState.content
			if content == "" {
				return GenerationErrorMsg{Err: fmt.Errorf("no content generated")}
			}

			// Check size limit
			maxSizeStr := viper.GetString("context.max-size")
			if maxSizeStr != "" {
				maxSize, err := parseSize(maxSizeStr)
				if err != nil {
					return GenerationErrorMsg{Err: fmt.Errorf("invalid max-size configuration: %v", err)}
				}

				contentSize := int64(len(content))
				if contentSize > maxSize {
					return GenerationErrorMsg{Err: fmt.Errorf("generated content size (%d bytes) exceeds maximum allowed size (%d bytes)", contentSize, maxSize)}
				}
			}

			timestamp := time.Now().Format("20060102-150405")
			filename := fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
			filePath := filepath.Join(m.generateState.rootPath, filename)

			if err := writeFile(filePath, content); err != nil {
				return GenerationErrorMsg{Err: err}
			}

			return GenerationCompleteMsg{
				Content:  content,
				FilePath: filePath,
			}
		default:
			// No progress yet, re-enqueue
			return m.iterativeGenerateCmd()()
		}
	}
}
