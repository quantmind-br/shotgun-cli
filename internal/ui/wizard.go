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
	"github.com/charmbracelet/lipgloss"
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

	minTerminalWidth  = 40
	minTerminalHeight = 10
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
	step     int
	progress Progress
	error    error
	width    int
	height   int
	showHelp bool

	rootPath       string
	scanConfig     *scanner.ScanConfig
	wizardConfig   *WizardConfig
	contextService app.ContextService

	fileSelection     *screens.FileSelectionModel
	templateSelection *screens.TemplateSelectionModel
	taskInput         *screens.TaskInputModel
	rulesInput        *screens.RulesInputModel
	review            *screens.ReviewModel

	progressComponent *components.ProgressModel

	scanCoordinator     *ScanCoordinator
	generateCoordinator *GenerateCoordinator

	generatedFilePath string
	generatedContent  string

	llmSending      bool
	llmResponseFile string

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

type LLMSendMsg struct{}

type generateCoordinator struct {
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

func NewWizard(rootPath string, scanConfig *scanner.ScanConfig, wizardConfig *WizardConfig, svc app.ContextService) *WizardModel {
	if wizardConfig == nil {
		wizardConfig = &WizardConfig{}
	}
	if svc == nil {
		svc = app.NewContextService()
	}
	return &WizardModel{
		step:                StepFileSelection,
		rootPath:            rootPath,
		scanConfig:          scanConfig,
		wizardConfig:        wizardConfig,
		contextService:      svc,
		progressComponent:   components.NewProgress(),
		scanCoordinator:     NewScanCoordinator(scanner.NewFileSystemScanner()),
		generateCoordinator: NewGenerateCoordinator(context.NewDefaultContextGenerator()),
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
		if m.templateSelection == nil {
			m.templateSelection = screens.NewTemplateSelection()
		}
		m.templateSelection.SetSelectedForTest(msg.Template)

	case GenerationProgressMsg:
		cmd = m.handleGenerationProgress(msg)
		cmds = append(cmds, cmd)

	case screens.GenerationCompleteMsg:
		cmd = m.handleGenerationComplete(msg)
		cmds = append(cmds, cmd)

	case screens.GenerationErrorMsg:
		m.handleGenerationError(msg)

	case screens.ClipboardCompleteMsg:
		m.handleClipboardComplete(msg)

	case screens.LLMProgressMsg:
		m.handleLLMProgress(msg)

	case screens.LLMCompleteMsg:
		m.handleLLMComplete(msg)

	case screens.LLMErrorMsg:
		m.handleLLMError(msg)

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
		if m.scanCoordinator != nil {
			cmd = m.scanCoordinator.Poll()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
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
		if m.step == StepFileSelection && m.fileSelection != nil && m.fileSelection.IsLoading() {
			spinnerCmd := m.fileSelection.Update(msg)
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
	if m.isTerminalTooSmall() {
		return m.renderSmallScreenWarning()
	}

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
	content.WriteString("  F1              Toggle this help screen\n")
	content.WriteString("  F7 / Ctrl+P     Previous step\n")
	content.WriteString("  F8 / Ctrl+N     Next step\n")
	content.WriteString("  Ctrl+Q          Quit application\n")
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
	content.WriteString("  F9          Send to LLM (if configured)\n")
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
	case "f8", "ctrl+pgdn", "ctrl+n":
		cmd = m.handleNextStep()
		cmds = append(cmds, cmd)
	case "f7", "ctrl+p":
		cmd = m.handlePrevStep()
		cmds = append(cmds, cmd)
	case "f9":
		// Send to LLM (only on review screen after generation)
		cmd = m.handleSendToLLM()
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
	return nil
}

func (m *WizardModel) handleScanComplete(msg ScanCompleteMsg) {
	m.progress.Visible = false
	if m.fileSelection != nil {
		m.fileSelection.SetFileTree(msg.Tree)
	} else {
		m.fileSelection = screens.NewFileSelection(msg.Tree, nil)
		m.fileSelection.SetSize(m.width, m.height)
	}
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
	return nil
}

func (m *WizardModel) handleGenerationComplete(msg screens.GenerationCompleteMsg) tea.Cmd {
	m.progress.Visible = false

	content := msg.Content
	filePath := msg.FilePath

	if filePath == "" {
		if err := m.validateContentSize(content); err != nil {
			m.handleGenerationError(screens.GenerationErrorMsg{Err: err})
			return nil
		}

		var err error
		filePath, err = m.saveGeneratedContent(content)
		if err != nil {
			m.handleGenerationError(screens.GenerationErrorMsg{Err: err})
			return nil
		}
	}

	m.generatedFilePath = filePath
	m.generatedContent = content

	return m.clipboardCopyCmd(content)
}

func (m *WizardModel) handleGenerationError(msg screens.GenerationErrorMsg) {
	m.error = msg.Err
	m.progress.Visible = false
}

func (m *WizardModel) handleClipboardComplete(msg screens.ClipboardCompleteMsg) {
	if m.review != nil && m.generatedFilePath != "" {
		m.review.SetGenerated(m.generatedFilePath, msg.Success)
	}
}

func (m *WizardModel) handleSendToLLM() tea.Cmd {
	if m.step != StepReview || m.generatedContent == "" || m.llmSending {
		return nil
	}

	m.llmSending = true
	if m.review != nil {
		m.review.SetLLMSending(true)
	}

	m.progress = Progress{
		Stage:   "sending",
		Message: "Sending to LLM...",
		Visible: true,
	}
	m.progressComponent.UpdateMessage("", "Sending to LLM...")

	return tea.Batch(m.sendToLLMCmd(), m.progressComponent.Init())
}

func (m *WizardModel) handleLLMProgress(msg screens.LLMProgressMsg) {
	m.progress = Progress{
		Stage:   msg.Stage,
		Message: "Sending to LLM...",
		Visible: true,
	}
	m.progressComponent.UpdateMessage("", "Sending to LLM...")
}

func (m *WizardModel) handleLLMComplete(msg screens.LLMCompleteMsg) {
	m.llmSending = false
	m.llmResponseFile = msg.OutputFile
	m.progress.Visible = false

	if m.review != nil {
		m.review.SetLLMComplete(msg.OutputFile, msg.Duration)
	}
}

func (m *WizardModel) handleLLMError(msg screens.LLMErrorMsg) {
	m.llmSending = false
	m.progress.Visible = false

	if m.review != nil {
		m.review.SetLLMError(msg.Err)
	}
}

func (m *WizardModel) buildLLMSendConfig() app.LLMSendConfig {
	cfg := app.LLMSendConfig{
		Provider:       llm.ProviderType(m.wizardConfig.LLM.Provider),
		APIKey:         m.wizardConfig.LLM.APIKey,
		BaseURL:        m.wizardConfig.LLM.BaseURL,
		Model:          m.wizardConfig.LLM.Model,
		Timeout:        m.wizardConfig.LLM.Timeout,
		BinaryPath:     m.wizardConfig.LLM.BinaryPath,
		BrowserRefresh: m.wizardConfig.LLM.BrowserRefresh,
	}

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

	saveResponse := m.wizardConfig.LLM.SaveResponse
	if !saveResponse {
		saveResponse = m.wizardConfig.Gemini.SaveResponse
	}
	cfg.SaveResponse = saveResponse
	cfg.OutputPath = strings.TrimSuffix(m.generatedFilePath, ".md") + "_response.md"

	return cfg
}

func (m *WizardModel) isLLMAvailable() bool {
	cfg := m.buildLLMSendConfig()
	llmCfg := llm.Config{
		Provider:       cfg.Provider,
		APIKey:         cfg.APIKey,
		BaseURL:        cfg.BaseURL,
		Model:          cfg.Model,
		Timeout:        cfg.Timeout,
		BinaryPath:     cfg.BinaryPath,
		BrowserRefresh: cfg.BrowserRefresh,
	}
	llmCfg.WithDefaults()

	provider, err := app.DefaultProviderRegistry.Create(llmCfg)
	if err != nil {
		return false
	}
	return provider.IsAvailable() && provider.ValidateConfig() == nil
}

func (m *WizardModel) sendToLLMCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := m.buildLLMSendConfig()
		ctx := gocontext.Background()

		result, err := m.contextService.SendToLLMWithProgress(ctx, m.generatedContent, cfg, nil)
		if err != nil {
			return screens.LLMErrorMsg{Err: err}
		}

		return screens.LLMCompleteMsg{
			Response:   result.Response,
			OutputFile: cfg.OutputPath,
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
	if m.scanCoordinator == nil {
		m.scanCoordinator = NewScanCoordinator(scanner.NewFileSystemScanner())
	}

	m.fileSelection = screens.NewFileSelection(nil, nil)
	m.fileSelection.SetSize(m.width, m.height)

	return tea.Batch(m.fileSelection.Init(), m.scanCoordinator.Start(msg.rootPath, msg.config))
}

func (m *WizardModel) handleStartGeneration(msg startGenerationMsg) tea.Cmd {
	cfg := &GenerateConfig{
		FileTree:       msg.fileTree,
		Selections:     msg.selectedFiles,
		Template:       msg.template,
		TaskDesc:       msg.taskDesc,
		Rules:          msg.rules,
		RootPath:       msg.rootPath,
		IncludeTree:    m.wizardConfig.Context.IncludeTree,
		IncludeSummary: m.wizardConfig.Context.IncludeSummary,
	}

	return m.generateCoordinator.Start(cfg)
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
		return len(m.getSelectedFiles()) > 0
	case StepTemplateSelection:
		return m.getSelectedTemplate() != nil
	case StepTaskInput:
		tmpl := m.getSelectedTemplate()
		if tmpl != nil && !tmpl.HasVariable(template.VarTask) {
			return true
		}
		return len(strings.TrimSpace(m.getTaskDesc())) > 0
	case StepRulesInput:
		return true
	case StepReview:
		return true
	default:
		return false
	}
}

func (m *WizardModel) requiresTaskInput() bool {
	tmpl := m.getSelectedTemplate()
	return tmpl != nil && tmpl.HasVariable(template.VarTask)
}

func (m *WizardModel) requiresRulesInput() bool {
	tmpl := m.getSelectedTemplate()
	return tmpl != nil && tmpl.HasVariable(template.VarRules)
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
		if m.getFileTree() != nil {
			m.fileSelection = screens.NewFileSelection(m.getFileTree(), m.getSelectedFiles())
			m.fileSelection.SetSize(m.width, m.height)
		}
	case StepTemplateSelection:
		m.templateSelection = screens.NewTemplateSelection()
		m.templateSelection.SetSize(m.width, m.height)
		return m.templateSelection.LoadTemplates()
	case StepTaskInput:
		m.taskInput = screens.NewTaskInput(m.getTaskDesc())
		m.taskInput.SetSize(m.width, m.height)
		// Set skip hint if template doesn't require RULES variable
		m.taskInput.SetWillSkipToReview(!m.requiresRulesInput())
	case StepRulesInput:
		m.rulesInput = screens.NewRulesInput(m.getRules())
		m.rulesInput.SetSize(m.width, m.height)
	case StepReview:
		m.review = screens.NewReview(
			m.getSelectedFiles(), m.getFileTree(), m.getSelectedTemplate(),
			m.getTaskDesc(), m.getRules(), m.wizardConfig.Context.MaxSize,
		)
		m.review.SetSize(m.width, m.height)
		m.review.SetLLMAvailable(m.isLLMAvailable())
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
		}
	case StepTaskInput:
		if m.taskInput != nil {
			cmd = m.taskInput.Update(msg)
		}
	case StepRulesInput:
		if m.rulesInput != nil {
			cmd = m.rulesInput.Update(msg)
		}
	case StepReview:
		if m.review != nil {
			cmd = m.review.Update(msg)
		}
	}

	return cmd
}

func (m *WizardModel) generateContext() tea.Cmd {
	tmpl := m.getSelectedTemplate()
	if tmpl == nil || len(m.getSelectedFiles()) == 0 {
		return func() tea.Msg {
			return screens.GenerationErrorMsg{Err: fmt.Errorf("missing template or files")}
		}
	}

	return generateContextCmd(m.getFileTree(), m.getSelectedFiles(), tmpl, m.getTaskDesc(), m.getRules(), m.rootPath)
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
		return screens.ClipboardCompleteMsg{
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

func (m *WizardModel) schedulePollGenerate() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg {
		return pollGenerateMsg{}
	})
}

func (m *WizardModel) pollGenerate() tea.Cmd {
	if m.generateCoordinator == nil {
		return nil
	}
	return m.generateCoordinator.Poll()
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
	filePath := filepath.Join(m.rootPath, filename)

	// #nosec G306 - Generated context files are meant to be world-readable
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return filePath, nil
}

func (m *WizardModel) getFileTree() *scanner.FileNode {
	if m.fileSelection != nil {
		return m.fileSelection.GetFileTree()
	}
	return nil
}

func (m *WizardModel) getSelectedTemplate() *template.Template {
	if m.templateSelection != nil {
		return m.templateSelection.GetSelected()
	}
	return nil
}

func (m *WizardModel) getSelectedFiles() map[string]bool {
	if m.fileSelection != nil {
		return m.fileSelection.GetSelections()
	}
	return nil
}

func (m *WizardModel) getTaskDesc() string {
	if m.taskInput != nil {
		return m.taskInput.GetValue()
	}
	return ""
}

func (m *WizardModel) getRules() string {
	if m.rulesInput != nil {
		return m.rulesInput.GetValue()
	}
	return ""
}

func (m *WizardModel) isTerminalTooSmall() bool {
	return m.width > 0 && m.height > 0 &&
		(m.width < minTerminalWidth || m.height < minTerminalHeight)
}

func (m *WizardModel) renderSmallScreenWarning() string {
	msg := fmt.Sprintf(
		"Terminal too small (%dx%d)\n\nPlease resize to at least %dx%d",
		m.width, m.height,
		minTerminalWidth, minTerminalHeight,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.WarningStyle.Render(msg),
	)
}
