package ui

import (
	gocontext "context"
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
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
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
	generatedContent  string

	// Gemini state
	geminiSending      bool
	geminiResponseFile string
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

// Gemini integration messages
type GeminiSendMsg struct{}

type GeminiProgressMsg struct {
	Stage string
}

type GeminiCompleteMsg struct {
	Response   string
	OutputFile string
	Duration   time.Duration
}

type GeminiErrorMsg struct {
	Err error
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
	selections map[string]bool
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

//nolint:gocyclo // type switch pattern required by Bubble Tea framework
func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.handleWindowResize(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case ScanProgressMsg:
		cmd = m.handleScanProgress(msg)
		cmds = append(cmds, cmd)

	case ScanCompleteMsg:
		m.handleScanComplete(msg)

	case ScanErrorMsg:
		m.handleScanError(msg)

	case TemplateSelectedMsg:
		m.template = msg.Template

	case GenerationProgressMsg:
		cmd = m.handleGenerationProgress(msg)
		cmds = append(cmds, cmd)

	case GenerationCompleteMsg:
		cmd = m.handleGenerationComplete(msg)
		cmds = append(cmds, cmd)

	case GenerationErrorMsg:
		m.handleGenerationError(msg)

	case ClipboardCompleteMsg:
		m.handleClipboardComplete(msg)

	case GeminiProgressMsg:
		m.handleGeminiProgress(msg)

	case GeminiCompleteMsg:
		m.handleGeminiComplete(msg)

	case GeminiErrorMsg:
		m.handleGeminiError(msg)

	case screens.TemplatesLoadedMsg, screens.TemplatesErrorMsg:
		cmd = m.handleTemplateMessage(msg)
		cmds = append(cmds, cmd)

	case startScanMsg:
		cmd = m.handleStartScan(msg)
		cmds = append(cmds, cmd)

	case startGenerationMsg:
		cmd = m.handleStartGeneration(msg)
		cmds = append(cmds, cmd)

	case screens.RescanRequestMsg:
		cmd = m.handleRescanRequest()
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, cmd
}

func (m *WizardModel) View() string {
	// Show help overlay if enabled
	if m.showHelp {
		return m.renderHelp()
	}

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

func (m *WizardModel) renderHelp() string {
	var content strings.Builder

	header := styles.RenderHeader(0, "Help - Keyboard Shortcuts")
	content.WriteString(header)
	content.WriteString("\n\n")

	// Global shortcuts
	content.WriteString(styles.TitleStyle.Render("Global Shortcuts"))
	content.WriteString("\n")
	content.WriteString("  F1          Toggle this help screen\n")
	content.WriteString("  F7          Previous step\n")
	content.WriteString("  F8          Next step\n")
	content.WriteString("  Ctrl+Q      Quit application\n")
	content.WriteString("\n")

	// File selection shortcuts
	content.WriteString(styles.TitleStyle.Render("File Selection (Step 1)"))
	content.WriteString("\n")
	content.WriteString("  ↑/↓ or k/j  Navigate up/down\n")
	content.WriteString("  ←/→ or h/l  Collapse/Expand directory\n")
	content.WriteString("  Space       Toggle selection (file or directory)\n")
	content.WriteString("  i           Toggle showing ignored files\n")
	content.WriteString("  /           Enter filter mode (fuzzy search)\n")
	content.WriteString("  Ctrl+C      Clear filter\n")
	content.WriteString("  F5          Rescan directory\n")
	content.WriteString("\n")

	// Template selection shortcuts
	content.WriteString(styles.TitleStyle.Render("Template Selection (Step 2)"))
	content.WriteString("\n")
	content.WriteString("  ↑/↓ or k/j  Navigate templates\n")
	content.WriteString("  Enter       Select template\n")
	content.WriteString("\n")

	// Text input shortcuts
	content.WriteString(styles.TitleStyle.Render("Text Input (Steps 3-4)"))
	content.WriteString("\n")
	content.WriteString("  Type        Enter text\n")
	content.WriteString("  Enter       New line\n")
	content.WriteString("  Backspace   Delete character\n")
	content.WriteString("\n")

	// Review shortcuts
	content.WriteString(styles.TitleStyle.Render("Review (Step 5)"))
	content.WriteString("\n")
	content.WriteString("  F8          Generate context\n")
	content.WriteString("  c           Copy to clipboard\n")
	content.WriteString("  F9          Send to Gemini (if configured)\n")
	content.WriteString("\n")

	footer := styles.RenderFooter([]string{"F1: Close Help", "Ctrl+Q: Quit"})
	content.WriteString(footer)

	return content.String()
}

//nolint:unparam // tea.Cmd return is part of consistent handler pattern
func (m *WizardModel) handleWindowResize(msg tea.WindowSizeMsg) tea.Cmd {
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
	return nil
}

func (m *WizardModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Always handle quit commands
	if msg.String() == "ctrl+c" || msg.String() == "ctrl+q" {
		return m, tea.Quit
	}

	// Process navigation shortcuts
	switch msg.String() {
	case "f8", "ctrl+pgdn":
		cmd = m.handleNextStep()
		cmds = append(cmds, cmd)
	case "f7":
		cmd = m.handlePrevStep()
		cmds = append(cmds, cmd)
	case "f9":
		// Send to Gemini (only on review screen after generation)
		cmd = m.handleSendToGemini()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case "f10", "ctrl+pgup":
		cmd = m.handlePrevStep()
		cmds = append(cmds, cmd)
	case "f1":
		m.showHelp = !m.showHelp
	default:
		cmd = m.handleStepInput(msg)
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, cmd
}

func (m *WizardModel) handleNextStep() tea.Cmd {
	if m.step < StepReview {
		if m.canAdvanceStep() {
			m.step = m.getNextStep()
			return m.initStep()
		}
	} else if m.step == StepReview {
		return m.generateContext()
	}
	return nil
}

func (m *WizardModel) handlePrevStep() tea.Cmd {
	if m.step > StepFileSelection {
		m.step = m.getPrevStep()
		return m.initStep()
	}
	return nil
}

func (m *WizardModel) handleScanProgress(msg ScanProgressMsg) tea.Cmd {
	m.progress = Progress{
		Current: msg.Current,
		Total:   msg.Total,
		Stage:   msg.Stage,
		Visible: true,
	}
	m.progressComponent.Update(msg.Current, msg.Total, msg.Stage, "")
	if m.scanState != nil {
		return m.iterativeScanCmd()
	}
	return nil
}

func (m *WizardModel) handleScanComplete(msg ScanCompleteMsg) {
	m.fileTree = msg.Tree
	m.progress.Visible = false
	m.fileSelection = screens.NewFileSelection(msg.Tree, m.selectedFiles)
	m.fileSelection.SetSize(m.width, m.height)
}

func (m *WizardModel) handleScanError(msg ScanErrorMsg) {
	m.error = msg.Err
	m.progress.Visible = false
}

func (m *WizardModel) handleGenerationProgress(msg GenerationProgressMsg) tea.Cmd {
	m.progress = Progress{
		Stage:   msg.Stage,
		Message: msg.Message,
		Visible: true,
	}
	m.progressComponent.UpdateMessage(msg.Stage, msg.Message)
	if m.generateState != nil {
		return m.iterativeGenerateCmd()
	}
	return nil
}

func (m *WizardModel) handleGenerationComplete(msg GenerationCompleteMsg) tea.Cmd {
	m.progress.Visible = false
	m.generatedFilePath = msg.FilePath
	m.generatedContent = msg.Content
	return m.clipboardCopyCmd(msg.Content)
}

func (m *WizardModel) handleGenerationError(msg GenerationErrorMsg) {
	m.error = msg.Err
	m.progress.Visible = false
}

func (m *WizardModel) handleClipboardComplete(msg ClipboardCompleteMsg) {
	if m.review != nil && m.generatedFilePath != "" {
		m.review.SetGenerated(m.generatedFilePath, msg.Success)
	}
}

func (m *WizardModel) handleSendToGemini() tea.Cmd {
	// Only allow on review screen after generation
	if m.step != StepReview || m.generatedContent == "" {
		return nil
	}

	// Check if already sending
	if m.geminiSending {
		return nil
	}

	// Check availability
	if !gemini.IsAvailable() {
		if m.review != nil {
			m.review.SetGeminiError(fmt.Errorf("geminiweb not found"))
		}
		return nil
	}

	if !gemini.IsConfigured() {
		if m.review != nil {
			m.review.SetGeminiError(fmt.Errorf("geminiweb not configured - run: geminiweb auto-login"))
		}
		return nil
	}

	m.geminiSending = true
	if m.review != nil {
		m.review.SetGeminiSending(true)
	}

	return m.sendToGeminiCmd()
}

func (m *WizardModel) handleGeminiProgress(msg GeminiProgressMsg) {
	m.progress = Progress{
		Stage:   msg.Stage,
		Message: "Sending to Gemini...",
		Visible: true,
	}
	m.progressComponent.UpdateMessage(msg.Stage, "Sending to Gemini...")
}

func (m *WizardModel) handleGeminiComplete(msg GeminiCompleteMsg) {
	m.geminiSending = false
	m.geminiResponseFile = msg.OutputFile
	m.progress.Visible = false

	if m.review != nil {
		m.review.SetGeminiComplete(msg.OutputFile, msg.Duration)
	}
}

func (m *WizardModel) handleGeminiError(msg GeminiErrorMsg) {
	m.geminiSending = false
	m.progress.Visible = false

	if m.review != nil {
		m.review.SetGeminiError(msg.Err)
	}
}

func (m *WizardModel) sendToGeminiCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := gemini.Config{
			BinaryPath:     viper.GetString("gemini.binary-path"),
			Model:          viper.GetString("gemini.model"),
			Timeout:        viper.GetInt("gemini.timeout"),
			BrowserRefresh: viper.GetString("gemini.browser-refresh"),
		}

		executor := gemini.NewExecutor(cfg)

		ctx := gocontext.Background()
		result, err := executor.Send(ctx, m.generatedContent)
		if err != nil {
			return GeminiErrorMsg{Err: err}
		}

		// Determine output file
		outputFile := strings.TrimSuffix(m.generatedFilePath, ".md") + "_response.md"

		// Save response if configured
		if viper.GetBool("gemini.save-response") {
			if err := os.WriteFile(outputFile, []byte(result.Response), 0600); err != nil {
				return GeminiErrorMsg{Err: fmt.Errorf("failed to save response: %w", err)}
			}
		}

		return GeminiCompleteMsg{
			Response:   result.Response,
			OutputFile: outputFile,
			Duration:   result.Duration,
		}
	}
}

func (m *WizardModel) handleTemplateMessage(msg tea.Msg) tea.Cmd {
	if m.step == StepTemplateSelection && m.templateSelection != nil {
		return m.templateSelection.HandleMessage(msg)
	}
	return nil
}

func (m *WizardModel) handleStartScan(msg startScanMsg) tea.Cmd {
	m.scanState = &scanState{
		scanner:    scanner.NewFileSystemScanner(),
		rootPath:   msg.rootPath,
		config:     msg.config,
		progressCh: make(chan scanner.Progress, 100),
		done:       make(chan bool),
		started:    false,
	}
	return m.iterativeScanCmd()
}

func (m *WizardModel) handleStartGeneration(msg startGenerationMsg) tea.Cmd {
	m.generateState = &generateState{
		generator:  context.NewDefaultContextGenerator(),
		fileTree:   msg.fileTree,
		selections: msg.selectedFiles,
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
	return m.iterativeGenerateCmd()
}

func (m *WizardModel) handleRescanRequest() tea.Cmd {
	if m.step == StepFileSelection {
		return scanDirectoryCmd(m.rootPath, m.config)
	}
	return nil
}

func (m *WizardModel) canAdvanceStep() bool {
	switch m.step {
	case StepFileSelection:
		return len(m.selectedFiles) > 0
	case StepTemplateSelection:
		return m.template != nil
	case StepTaskInput:
		// Only require task description if template has TASK variable
		if m.template != nil && !m.template.HasVariable(template.VarTask) {
			return true
		}
		return len(strings.TrimSpace(m.taskDesc)) > 0
	case StepRulesInput:
		return true
	case StepReview:
		return true
	default:
		return false
	}
}

// requiresTaskInput returns true if the current template requires the TASK variable
func (m *WizardModel) requiresTaskInput() bool {
	return m.template != nil && m.template.HasVariable(template.VarTask)
}

// requiresRulesInput returns true if the current template requires the RULES variable
func (m *WizardModel) requiresRulesInput() bool {
	return m.template != nil && m.template.HasVariable(template.VarRules)
}

// getNextStep returns the next step to navigate to, skipping steps that are not needed
func (m *WizardModel) getNextStep() int {
	switch m.step {
	case StepFileSelection:
		return StepTemplateSelection
	case StepTemplateSelection:
		if !m.requiresTaskInput() {
			if !m.requiresRulesInput() {
				return StepReview
			}
			return StepRulesInput
		}
		return StepTaskInput
	case StepTaskInput:
		if !m.requiresRulesInput() {
			return StepReview
		}
		return StepRulesInput
	case StepRulesInput:
		return StepReview
	default:
		return m.step + 1
	}
}

// getPrevStep returns the previous step to navigate to, skipping steps that were not needed
func (m *WizardModel) getPrevStep() int {
	switch m.step {
	case StepTemplateSelection:
		return StepFileSelection
	case StepTaskInput:
		return StepTemplateSelection
	case StepRulesInput:
		if !m.requiresTaskInput() {
			return StepTemplateSelection
		}
		return StepTaskInput
	case StepReview:
		if !m.requiresRulesInput() {
			if !m.requiresTaskInput() {
				return StepTemplateSelection
			}
			return StepTaskInput
		}
		return StepRulesInput
	default:
		return m.step - 1
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
		// Set skip hint if template doesn't require RULES variable
		m.taskInput.SetWillSkipToReview(!m.requiresRulesInput())
	case StepRulesInput:
		m.rulesInput = screens.NewRulesInput(m.rules)
		m.rulesInput.SetSize(m.width, m.height)
	case StepReview:
		m.review = screens.NewReview(m.selectedFiles, m.fileTree, m.template, m.taskDesc, m.rules)
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

func generateContextCmd(
	fileTree *scanner.FileNode,
	selectedFiles map[string]bool,
	template *template.Template,
	taskDesc, rules, rootPath string,
) tea.Cmd {
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
	// #nosec G306 - Generated context files are meant to be world-readable (contain code/docs, not secrets)
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
					m.generateState.selections,
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
				return m.finalizeGeneration()
			}
			return GenerationProgressMsg{
				Stage:   progress.Stage,
				Message: progress.Message,
			}
		case <-m.generateState.done:
			return m.finalizeGeneration()
		default:
			return m.iterativeGenerateCmd()()
		}
	}
}

func (m *WizardModel) finalizeGeneration() tea.Msg {
	content := m.generateState.content
	if content == "" {
		return GenerationErrorMsg{Err: fmt.Errorf("no content generated")}
	}

	if err := m.validateContentSize(content); err != nil {
		return GenerationErrorMsg{Err: err}
	}

	filePath, err := m.saveGeneratedContent(content)
	if err != nil {
		return GenerationErrorMsg{Err: err}
	}

	return GenerationCompleteMsg{
		Content:  content,
		FilePath: filePath,
	}
}

func (m *WizardModel) validateContentSize(content string) error {
	maxSizeStr := viper.GetString("context.max-size")
	if maxSizeStr == "" {
		return nil
	}

	maxSize, err := parseSize(maxSizeStr)
	if err != nil {
		return fmt.Errorf("invalid max-size configuration: %v", err)
	}

	contentSize := int64(len(content))
	if contentSize > maxSize {
		return fmt.Errorf("generated content size (%d bytes) exceeds maximum allowed size (%d bytes)",
			contentSize, maxSize)
	}
	return nil
}

func (m *WizardModel) saveGeneratedContent(content string) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
	filePath := filepath.Join(m.generateState.rootPath, filename)

	if err := writeFile(filePath, content); err != nil {
		return "", err
	}
	return filePath, nil
}
