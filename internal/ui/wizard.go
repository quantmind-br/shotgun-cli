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
	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/platform/clipboard"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

const (
	StepFileSelection     = 1
	StepTemplateSelection = 2
	StepTaskInput         = 3
	StepRulesInput        = 4
	StepReview            = 5
)

// LLMConfig holds configuration for the LLM provider.
type LLMConfig struct {
	Provider       string
	APIKey         string
	BaseURL        string
	Model          string
	Timeout        int
	SaveResponse   bool
	BinaryPath     string // For GeminiWeb
	BrowserRefresh string // For GeminiWeb
}

// GeminiConfig holds legacy configuration for GeminiWeb.
type GeminiConfig struct {
	BinaryPath     string
	Model          string
	Timeout        int
	BrowserRefresh string
	SaveResponse   bool
}

// ContextConfig holds context generation configuration.
type ContextConfig struct {
	IncludeTree    bool
	IncludeSummary bool
	MaxSize        string
}

// WizardConfig holds all wizard configuration.
type WizardConfig struct {
	LLM     LLMConfig
	Gemini  GeminiConfig
	Context ContextConfig
}

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

	rootPath     string
	scanConfig   *scanner.ScanConfig
	wizardConfig *WizardConfig

	fileSelection     *screens.FileSelectionModel
	templateSelection *screens.TemplateSelectionModel
	taskInput         *screens.TaskInputModel
	rulesInput        *screens.RulesInputModel
	review            *screens.ReviewModel

	progressComponent *components.ProgressModel

	scanState     *scanState
	generateState *generateState

	generatedFilePath string
	generatedContent  string

	geminiSending      bool
	geminiResponseFile string

	// Validation error to display when user tries to advance without completing required fields
	validationError string
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
	result     *scanner.FileNode // Store scan result to avoid double-scanning
	scanErr    error             // Store any error from scan
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

type pollScanMsg struct{}
type pollGenerateMsg struct{}

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

func NewWizard(rootPath string, scanConfig *scanner.ScanConfig, wizardConfig *WizardConfig) *WizardModel {
	if wizardConfig == nil {
		wizardConfig = &WizardConfig{}
	}
	return &WizardModel{
		step:              StepFileSelection,
		selectedFiles:     make(map[string]bool),
		rootPath:          rootPath,
		scanConfig:        scanConfig,
		wizardConfig:      wizardConfig,
		progressComponent: components.NewProgress(),
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return tea.Batch(
		scanDirectoryCmd(m.rootPath, m.scanConfig),
		m.progressComponent.Init(),
	)
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

	case pollScanMsg:
		cmd = m.pollScan()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case pollGenerateMsg:
		cmd = m.pollGenerate()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case screens.RescanRequestMsg:
		cmd = m.handleRescanRequest()
		cmds = append(cmds, cmd)

	case screens.ClipboardCopyRequestMsg:
		if m.generatedContent != "" {
			cmds = append(cmds, m.clipboardCopyCmd(m.generatedContent))
		}

	default:
		if m.progress.Visible {
			var spinnerCmd tea.Cmd
			m.progressComponent, spinnerCmd = m.progressComponent.UpdateSpinner(msg)
			if spinnerCmd != nil {
				cmds = append(cmds, spinnerCmd)
			}
		}
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

	if m.validationError != "" {
		mainView += "\n" + styles.RenderWarning(m.validationError)
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
	content.WriteString("  a           Select all visible files\n")
	content.WriteString("  A           Deselect all visible files\n")
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
	content.WriteString("  v           View full template (opens modal)\n")
	content.WriteString("\n")
	content.WriteString(styles.TitleStyle.Render("  In Template Preview Modal"))
	content.WriteString("\n")
	content.WriteString("    j/k       Scroll up/down\n")
	content.WriteString("    PgUp/Down Page scroll\n")
	content.WriteString("    g/G       Jump to top/bottom\n")
	content.WriteString("    Esc/q     Close modal\n")
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
			m.validationError = ""
			m.step = m.getNextStep()

			return m.initStep()
		}
		m.validationError = m.getValidationErrorMessage()
		return nil
	} else if m.step == StepReview {
		return m.generateContext()
	}

	return nil
}

func (m *WizardModel) getValidationErrorMessage() string {
	switch m.step {
	case StepFileSelection:
		return "Select at least one file to continue"
	case StepTemplateSelection:
		return "Select a template to continue"
	case StepTaskInput:
		return "Enter a task description to continue"
	default:
		return ""
	}
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

	// Create provider based on configuration
	provider, err := m.createLLMProvider()
	if err != nil {
		if m.review != nil {
			m.review.SetGeminiError(err)
		}
		return nil
	}

	// Check availability
	if !provider.IsAvailable() {
		if m.review != nil {
			m.review.SetGeminiError(fmt.Errorf("%s not available", provider.Name()))
		}
		return nil
	}

	// Validate config
	if err := provider.ValidateConfig(); err != nil {
		if m.review != nil {
			m.review.SetGeminiError(err)
		}
		return nil
	}

	m.geminiSending = true
	if m.review != nil {
		m.review.SetGeminiSending(true)
	}

	return m.sendToLLMCmd(provider)
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

// createLLMProvider creates an LLM provider based on wizard configuration.
func (m *WizardModel) createLLMProvider() (llm.Provider, error) {
	cfg := llm.Config{
		Provider:       llm.ProviderType(m.wizardConfig.LLM.Provider),
		APIKey:         m.wizardConfig.LLM.APIKey,
		BaseURL:        m.wizardConfig.LLM.BaseURL,
		Model:          m.wizardConfig.LLM.Model,
		Timeout:        m.wizardConfig.LLM.Timeout,
		BinaryPath:     m.wizardConfig.LLM.BinaryPath,
		BrowserRefresh: m.wizardConfig.LLM.BrowserRefresh,
	}

	// Apply defaults if not set
	cfg.WithDefaults()

	// For GeminiWeb, use legacy config if LLM config is not set
	if cfg.Provider == llm.ProviderGeminiWeb || cfg.Provider == "" {
		if cfg.Model == "" {
			cfg.Model = m.wizardConfig.Gemini.Model
		}
		if cfg.Timeout == 0 {
			cfg.Timeout = m.wizardConfig.Gemini.Timeout
		}
		if cfg.BinaryPath == "" {
			cfg.BinaryPath = m.wizardConfig.Gemini.BinaryPath
		}
		if cfg.BrowserRefresh == "" {
			cfg.BrowserRefresh = m.wizardConfig.Gemini.BrowserRefresh
		}
		cfg.Provider = llm.ProviderGeminiWeb
	}

	return app.DefaultProviderRegistry.Create(cfg)
}

func (m *WizardModel) sendToLLMCmd(provider llm.Provider) tea.Cmd {
	return func() tea.Msg {
		ctx := gocontext.Background()
		result, err := provider.SendWithProgress(ctx, m.generatedContent, func(stage string) {
			// Progress callback - could be used for updates
		})
		if err != nil {
			return GeminiErrorMsg{Err: err}
		}

		outputFile := strings.TrimSuffix(m.generatedFilePath, ".md") + "_response.md"

		saveResponse := m.wizardConfig.LLM.SaveResponse
		if !saveResponse {
			saveResponse = m.wizardConfig.Gemini.SaveResponse
		}

		if saveResponse {
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
			Template:       msg.template.Content,
			IncludeTree:    m.wizardConfig.Context.IncludeTree,
			IncludeSummary: m.wizardConfig.Context.IncludeSummary,
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
		return scanDirectoryCmd(m.rootPath, m.scanConfig)
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
		m.review = screens.NewReview(
			m.selectedFiles, m.fileTree, m.template,
			m.taskDesc, m.rules, m.wizardConfig.Context.MaxSize,
		)
		m.review.SetSize(m.width, m.height)
	}
	return nil
}

func (m *WizardModel) handleStepInput(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd

	m.validationError = ""

	switch m.step {
	case StepFileSelection:
		if m.fileSelection != nil {
			cmd = m.fileSelection.Update(msg)
		}
	case StepTemplateSelection:
		if m.templateSelection != nil {
			cmd = m.templateSelection.Update(msg)
			if tmpl := m.templateSelection.GetSelected(); tmpl != nil {
				m.template = tmpl
			}
		}
	case StepTaskInput:
		if m.taskInput != nil {
			cmd = m.taskInput.Update(msg)
			m.taskDesc = m.taskInput.GetValue()
		}
	case StepRulesInput:
		if m.rulesInput != nil {
			cmd = m.rulesInput.Update(msg)
			m.rules = m.rulesInput.GetValue()
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

func scanDirectoryCmd(rootPath string, scanConfig *scanner.ScanConfig) tea.Cmd {
	return func() tea.Msg {
		return startScanMsg{
			rootPath: rootPath,
			config:   scanConfig,
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

	switch {
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	case strings.HasSuffix(sizeStr, "B"):
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size integer: %w", err)
	}

	return size * multiplier, nil
}

const pollInterval = 50 * time.Millisecond

func (m *WizardModel) schedulePollScan() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollScanMsg{}
	})
}

func (m *WizardModel) schedulePollGenerate() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollGenerateMsg{}
	})
}

func (m *WizardModel) iterativeScanCmd() tea.Cmd {
	return func() tea.Msg {
		if m.scanState == nil {
			return ScanErrorMsg{Err: fmt.Errorf("scan state not initialized")}
		}

		if !m.scanState.started {
			m.scanState.started = true
			go func() {
				defer close(m.scanState.done)
				tree, err := m.scanState.scanner.ScanWithProgress(
					m.scanState.rootPath,
					m.scanState.config,
					m.scanState.progressCh,
				)
				m.scanState.result = tree
				m.scanState.scanErr = err
			}()
		}

		return pollScanMsg{}
	}
}

func (m *WizardModel) pollScan() tea.Cmd {
	if m.scanState == nil {
		return nil
	}

	select {
	case progress, ok := <-m.scanState.progressCh:
		if !ok {
			return m.finishScan()
		}
		return tea.Batch(
			func() tea.Msg {
				return ScanProgressMsg{
					Current: progress.Current,
					Total:   progress.Total,
					Stage:   progress.Stage,
				}
			},
			m.schedulePollScan(),
		)
	case <-m.scanState.done:
		return m.finishScan()
	default:
		return m.schedulePollScan()
	}
}

func (m *WizardModel) finishScan() tea.Cmd {
	return func() tea.Msg {
		if m.scanState.scanErr != nil {
			return ScanErrorMsg{Err: m.scanState.scanErr}
		}
		return ScanCompleteMsg{Tree: m.scanState.result}
	}
}

func (m *WizardModel) iterativeGenerateCmd() tea.Cmd {
	return func() tea.Msg {
		if m.generateState == nil {
			return GenerationErrorMsg{Err: fmt.Errorf("generation state not initialized")}
		}

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
					m.generateState.content = ""
					return
				}

				m.generateState.content = content
			}()
		}

		return pollGenerateMsg{}
	}
}

func (m *WizardModel) pollGenerate() tea.Cmd {
	if m.generateState == nil {
		return nil
	}

	select {
	case progress, ok := <-m.generateState.progressCh:
		if !ok {
			return m.finishGeneration()
		}
		if progress.Stage == "complete" {
			return m.finishGeneration()
		}
		return tea.Batch(
			func() tea.Msg {
				return GenerationProgressMsg{
					Stage:   progress.Stage,
					Message: progress.Message,
				}
			},
			m.schedulePollGenerate(),
		)
	case <-m.generateState.done:
		return m.finishGeneration()
	default:
		return m.schedulePollGenerate()
	}
}

func (m *WizardModel) finishGeneration() tea.Cmd {
	return func() tea.Msg {
		return m.finalizeGeneration()
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
	maxSizeStr := m.wizardConfig.Context.MaxSize
	if maxSizeStr == "" {
		return nil
	}

	maxSize, err := parseSize(maxSizeStr)
	if err != nil {
		return fmt.Errorf("invalid max-size configuration: %w", err)
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

	// #nosec G306 - Generated context files are meant to be world-readable
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}
