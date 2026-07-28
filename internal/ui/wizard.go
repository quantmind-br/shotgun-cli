package ui

import (
	gocontext "context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/llm"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/selection"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/platform/clipboard"
	"github.com/quantmind-br/shotgun-cli/internal/ui/components"
	"github.com/quantmind-br/shotgun-cli/internal/ui/screens"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
	"github.com/quantmind-br/shotgun-cli/internal/utils"
	"github.com/rs/zerolog/log"
)

const (
	StepFileSelection     = 1
	StepTemplateSelection = 2
	StepTaskInput         = 3
	StepRulesInput        = 4
	StepReview            = 5

	minTerminalWidth  = 60
	minTerminalHeight = 16
)

// LLMConfig holds configuration for the LLM provider.
type LLMConfig struct {
	Provider     string
	APIKey       string
	BaseURL      string
	Model        string
	Timeout      int
	SaveResponse bool
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
	step         int
	progress     Progress
	error        error
	width        int
	height       int
	showHelp     bool
	helpViewport viewport.Model

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
	confirmQuit     bool

	selectionStore   *selection.Store
	deselected       []string
	selectionsSeeded bool
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

func NewWizard(
	rootPath string, scanConfig *scanner.ScanConfig, wizardConfig *WizardConfig, svc app.ContextService,
) *WizardModel {
	if wizardConfig == nil {
		wizardConfig = &WizardConfig{}
	}
	if svc == nil {
		svc = app.NewContextService()
	}
	m := &WizardModel{
		step:                StepFileSelection,
		rootPath:            rootPath,
		scanConfig:          scanConfig,
		wizardConfig:        wizardConfig,
		contextService:      svc,
		progressComponent:   components.NewProgress(),
		scanCoordinator:     NewScanCoordinator(scanner.NewFileSystemScanner()),
		generateCoordinator: NewGenerateCoordinator(contextgen.NewDefaultContextGenerator()),
		helpViewport:        viewport.New(0, 0),
	}
	m.helpViewport.SetContent(m.renderHelpContent())
	return m
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
	// -- System & Navigation --
	case tea.WindowSizeMsg:
		cmd = m.handleWindowResize(msg)

		return m, cmd

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	// -- Scan Operations (Async) --
	case ScanProgressMsg:
		// Updates progress bar during directory scanning
		cmd = m.handleScanProgress(msg)
		cmds = append(cmds, cmd)

	case ScanCompleteMsg:
		// Critical transition: Scanning finished, data available
		// Switches from loading state to File Selection screen
		m.handleScanComplete(msg)

	case ScanErrorMsg:
		// Blocking state: Scan failed, shows error screen
		m.handleScanError(msg)

	// -- Template Operations --
	case TemplateSelectedMsg:
		if m.templateSelection == nil {
			m.templateSelection = screens.NewTemplateSelection()
		}
		m.templateSelection.SetSelectedForTest(msg.Template)

	case screens.TemplatesLoadedMsg, screens.TemplatesErrorMsg:
		cmd = m.handleTemplateMessage(msg)
		cmds = append(cmds, cmd)

	// -- Generation Operations (Async) --
	case GenerationProgressMsg:
		// Updates progress bar during context generation
		cmd = m.handleGenerationProgress(msg)
		cmds = append(cmds, cmd)

	case screens.GenerationCompleteMsg:
		// Critical transition: Generation finished, content ready
		// Triggers clipboard copy and enables LLM sending in Review screen
		cmd = m.handleGenerationComplete(msg)
		cmds = append(cmds, cmd)

	case screens.GenerationErrorMsg:
		m.handleGenerationError(msg)

	// -- Clipboard Operations --
	case screens.ClipboardCompleteMsg:
		cmds = append(cmds, m.handleClipboardComplete(msg))

	case copyToastClearMsg:
		if m.review != nil {
			m.review.SetCopyToast(false)
		}

	case screens.ClipboardCopyRequestMsg:
		if m.generatedContent != "" {
			cmds = append(cmds, m.clipboardCopyCmd(m.generatedContent))
		}

	// -- LLM Operations (Async) --
	case screens.LLMProgressMsg:
		m.handleLLMProgress(msg)

	case screens.LLMCompleteMsg:
		// Critical transition: LLM response received
		// Updates Review screen with response file path
		m.handleLLMComplete(msg)

	case screens.LLMErrorMsg:
		m.handleLLMError(msg)

	// -- Async Coordinators --
	case startScanMsg:
		// Triggers the ScanCoordinator to begin async filesystem scan
		cmd = m.handleStartScan(msg)
		cmds = append(cmds, cmd)

	case startGenerationMsg:
		// Triggers the GenerateCoordinator to begin async context generation
		cmd = m.handleStartGeneration(msg)
		cmds = append(cmds, cmd)

	case pollScanMsg:
		// Polls the ScanCoordinator for progress/results
		if m.scanCoordinator != nil {
			cmd = m.scanCoordinator.Poll()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case pollGenerateMsg:
		// Polls the GenerateCoordinator for progress/results
		cmd = m.pollGenerate()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case screens.RescanRequestMsg:
		cmd = m.handleRescanRequest()
		cmds = append(cmds, cmd)

	case screens.ToggleIgnoredScanMsg:
		cmd = m.handleToggleIgnoredScan()
		cmds = append(cmds, cmd)

	// -- Polling & Spinners --
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

	if m.confirmQuit {
		return m.renderConfirmQuit()
	}

	if m.showHelp {
		vw, vh := m.width, m.height
		if vw <= 0 {
			vw = 80
		}
		if vh <= 0 {
			vh = 24
		}
		m.helpViewport.Width = vw
		m.helpViewport.Height = vh
		m.helpViewport.SetContent(m.renderHelpContent())
		return m.helpViewport.View()
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

func (m *WizardModel) renderHelpContent() string {
	var content strings.Builder

	header := styles.RenderHeader(0, "Help - Keyboard Shortcuts")
	content.WriteString(header)
	content.WriteString("\n\n")

	content.WriteString(styles.TitleStyle.Render("Global Shortcuts"))
	content.WriteString("\n")
	content.WriteString("  F1 / ?          Toggle this help screen\n")
	content.WriteString("  F7 / Alt+←      Previous step\n")
	content.WriteString("  F8 / Alt+→      Next step\n")
	content.WriteString("  Tab             Next step alias\n")
	content.WriteString("  Shift+Tab       Previous step alias\n")
	content.WriteString("  q / Ctrl+Q      Quit application\n")
	content.WriteString("\n")

	content.WriteString(styles.TitleStyle.Render("File Selection (Step 1)"))
	content.WriteString("\n")
	content.WriteString("  ↑/↓ or k/j  Navigate up/down\n")
	content.WriteString("  ←/→ or h/l  Collapse/Expand directory\n")
	content.WriteString("  Space       Toggle selection (file or directory)\n")
	content.WriteString("  a           Select all visible files\n")
	content.WriteString("  A           Deselect all visible files\n")
	content.WriteString("  i           Toggle scanning ignored files (rescan)\n")
	content.WriteString("  /           Enter filter mode (fuzzy search)\n")
	content.WriteString("  x           Clear filter\n")
	content.WriteString("  F5 / r      Rescan directory\n")
	content.WriteString("\n")

	content.WriteString(styles.TitleStyle.Render("Template Selection (Step 2)"))
	content.WriteString("\n")
	content.WriteString("  ↑/↓ or k/j  Navigate templates\n")
	content.WriteString("  Space       Select template\n")
	content.WriteString("  Enter       Open preview (modal)\n")
	content.WriteString("  v           View full template (opens modal)\n")
	content.WriteString("\n")
	content.WriteString(styles.TitleStyle.Render("  In Template Preview Modal"))
	content.WriteString("\n")
	content.WriteString("    j/k       Scroll up/down\n")
	content.WriteString("    PgUp/Down Page scroll\n")
	content.WriteString("    g/G       Jump to top/bottom\n")
	content.WriteString("    Esc/q     Close modal\n")
	content.WriteString("\n")

	content.WriteString(styles.TitleStyle.Render("Text Input (Steps 3-4)"))
	content.WriteString("\n")
	content.WriteString("  Type        Enter text\n")
	content.WriteString("  Tab         Toggle focus (edit/done)\n")
	content.WriteString("  Esc         Cancel / leave input\n")
	content.WriteString("  Enter       New line\n")
	content.WriteString("  Backspace   Delete character\n")
	content.WriteString("\n")

	content.WriteString(styles.TitleStyle.Render("Review (Step 5)"))
	content.WriteString("\n")
	content.WriteString("  F8 / g      Generate context\n")
	content.WriteString("  c           Copy to clipboard\n")
	content.WriteString("  F9 / s      Send to LLM (if configured)\n")
	content.WriteString("\n")

	footer := styles.RenderFooter([]string{"F1/?: Close Help", "q: Quit"})
	content.WriteString(footer)

	return content.String()
}

func (m *WizardModel) renderConfirmQuit() string {
	var confirm strings.Builder

	title := "Unsaved Changes"
	if m.llmSending || m.progress.Visible {
		title = "Operation in Progress"
	}
	confirm.WriteString(styles.WarningStyle.Render(title))
	confirm.WriteString("\n\n")

	if m.llmSending {
		confirm.WriteString("LLM send is in progress. Quitting now will cancel it.\n\n")
	} else if m.progress.Visible {
		confirm.WriteString("A background operation is running. Quitting now will cancel it.\n\n")
	} else {
		confirm.WriteString("You have unsaved changes. What would you like to do?\n\n")
	}

	confirm.WriteString("  [Y] Quit without saving\n")
	confirm.WriteString("  [N] Cancel and continue\n")

	return styles.RenderBox(confirm.String(), "")
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

	m.helpViewport.Width = m.width
	m.helpViewport.Height = m.height

	return nil
}

//nolint:gocyclo // key routing switch required by TUI framework
func (m *WizardModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if msg.String() == "ctrl+c" || msg.String() == "ctrl+q" {
		return m, tea.Quit
	}

	if m.confirmQuit {
		switch msg.String() {
		case "y", "Y":
			return m, tea.Quit
		case "n", "N", "esc", "q":
			m.confirmQuit = false
		}
		return m, nil
	}

	// When help is showing, allow scrolling and closing
	if m.showHelp {
		switch msg.String() {
		case "f1", "esc", "?", "q":
			m.showHelp = false
			return m, nil
		case "j", "down":
			count := len(msg.Runes)
			if count <= 0 {
				count = 1
			}
			m.helpViewport.ScrollDown(count)
		case "k", "up":
			count := len(msg.Runes)
			if count <= 0 {
				count = 1
			}
			m.helpViewport.ScrollUp(count)
		case "pgdown", "ctrl+d":
			m.helpViewport.HalfPageDown()
		case "pgup", "ctrl+u":
			m.helpViewport.HalfPageUp()
		case "g", "home":
			m.helpViewport.GotoTop()
		case "G", "end":
			m.helpViewport.GotoBottom()
		}
		return m, nil
	}

	// Process navigation shortcuts
	switch msg.String() {
	case "f8", "alt+right":
		cmd = m.handleNextStep()
		cmds = append(cmds, cmd)
	case "tab":
		if m.isTextInputActive() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		cmd = m.handleNextStep()
		cmds = append(cmds, cmd)
	case "f7", "alt+left":
		cmd = m.handlePrevStep()
		cmds = append(cmds, cmd)
	case "shift+tab":
		if m.isTextInputActive() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		cmd = m.handlePrevStep()
		cmds = append(cmds, cmd)
	case "f9", "s":
		if m.isTextInputActive() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		// Send to LLM (only on review screen after generation)
		cmd = m.handleSendToLLM()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case "f5", "r":
		if m.isTextInputActive() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		if m.step == StepFileSelection && m.fileSelection != nil {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
		}
	case "g":
		if m.isTextInputActive() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		if m.step == StepReview {
			cmd = m.handleNextStep()
			cmds = append(cmds, cmd)
			break
		}
		cmd = m.handleStepInput(msg)
		cmds = append(cmds, cmd)
	case "f1":
		m.showHelp = !m.showHelp
	case "?":
		if !m.isTextInputActive() {
			m.showHelp = !m.showHelp
		} else {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
		}
	case "q":
		if m.isModalOpen() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		if m.isTextInputActive() {
			cmd = m.handleStepInput(msg)
			cmds = append(cmds, cmd)
			break
		}
		if m.progress.Visible || m.llmSending || m.hasEnteredData() {
			m.confirmQuit = true
			return m, nil
		}
		return m, tea.Quit
	default:
		cmd = m.handleStepInput(msg)
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}

	return m, cmd
}

// isTextInputActive returns true when the current step accepts free text input,
// preventing single-key shortcuts (q, ?) from being intercepted.
func (m *WizardModel) isTextInputActive() bool {
	if m.step == StepTaskInput || m.step == StepRulesInput {
		return true
	}
	if m.step == StepFileSelection && m.fileSelection != nil && m.fileSelection.IsFilterMode() {
		return true
	}
	return false
}

// hasEnteredData reports whether the user has provided input that quitting
// would silently discard (selected files or typed task/rules text). Used to
// gate the confirm-quit modal so `q` doesn't throw away work without warning.
func (m *WizardModel) hasEnteredData() bool {
	if m.fileSelection != nil && m.fileSelection.GetSelectedCount() > 0 {
		return true
	}
	if m.taskInput != nil && strings.TrimSpace(m.taskInput.GetValue()) != "" {
		return true
	}
	if m.rulesInput != nil && strings.TrimSpace(m.rulesInput.GetValue()) != "" {
		return true
	}
	return false
}

func (m *WizardModel) isModalOpen() bool {
	if m.step == StepTemplateSelection && m.templateSelection != nil && m.templateSelection.IsShowingFullPreview() {
		return true
	}
	return false
}

func (m *WizardModel) handleNextStep() tea.Cmd {
	if m.step < StepReview {
		if m.canAdvanceStep() {
			m.validationError = ""
			if m.step == StepFileSelection {
				m.persistSelections()
			}
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
		m.fileSelection = screens.NewFileSelection(msg.Tree, nil, m.wizardConfig.Context.MaxSize)
		m.fileSelection.SetSize(m.width, m.height)
	}
	if m.scanConfig != nil {
		m.fileSelection.SetShowIgnored(m.scanConfig.IncludeIgnored)
	}
	if !m.selectionsSeeded {
		includeIgnored := m.scanConfig != nil && m.scanConfig.IncludeIgnored
		m.fileSelection.SetSelections(scanner.SelectAllExcept(msg.Tree, m.deselected, includeIgnored))
		m.selectionsSeeded = true
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

// copyToastClearMsg fires shortly after a successful copy to dismiss the
// transient "Copied to clipboard!" confirmation on the Review screen.
type copyToastClearMsg struct{}

func (m *WizardModel) handleClipboardComplete(msg screens.ClipboardCompleteMsg) tea.Cmd {
	if m.review == nil || m.generatedFilePath == "" {
		return nil
	}
	m.review.SetGenerated(m.generatedFilePath, msg.Success)
	if !msg.Success {
		return nil
	}
	m.review.SetCopyToast(true)
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return copyToastClearMsg{}
	})
}

func (m *WizardModel) handleSendToLLM() tea.Cmd {
	if m.step != StepReview || m.generatedContent == "" || m.llmSending {
		return nil
	}

	if !m.isLLMAvailable() {
		m.validationError = "No LLM configured — run 'shotgun-cli config'"
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
		Provider:     llm.ProviderType(m.wizardConfig.LLM.Provider),
		APIKey:       m.wizardConfig.LLM.APIKey,
		BaseURL:      m.wizardConfig.LLM.BaseURL,
		Model:        m.wizardConfig.LLM.Model,
		SaveResponse: m.wizardConfig.LLM.SaveResponse,
	}

	if cfg.SaveResponse {
		ext := ".md"
		base := strings.TrimSuffix(filepath.Base(m.generatedFilePath), filepath.Ext(m.generatedFilePath))
		dir := filepath.Dir(m.generatedFilePath)
		cfg.OutputPath = filepath.Join(dir, base+"_response"+ext)
	}

	return cfg
}

func (m *WizardModel) isLLMAvailable() bool {
	if m.wizardConfig == nil {
		return false
	}
	// For API providers, we need an API key
	provider := llm.ProviderType(m.wizardConfig.LLM.Provider)
	if provider == "" {
		return false // No provider configured
	}

	// Basic check: require API key for known API providers
	if provider == llm.ProviderOpenAI || provider == llm.ProviderAnthropic || provider == llm.ProviderGemini {
		return m.wizardConfig.LLM.APIKey != ""
	}

	return true
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

	var selections map[string]bool
	if m.fileSelection != nil {
		selections = m.fileSelection.GetSelections()
	}
	m.fileSelection = screens.NewFileSelection(nil, selections, m.wizardConfig.Context.MaxSize)
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

func (m *WizardModel) handleToggleIgnoredScan() tea.Cmd {
	if m.step != StepFileSelection || m.scanConfig == nil {
		return nil
	}
	m.scanConfig.IncludeIgnored = !m.scanConfig.IncludeIgnored
	return scanDirectoryCmd(m.rootPath, m.scanConfig)
}

// SetSelectionStore attaches the persistence store and loads this project's saved deselections.
func (m *WizardModel) SetSelectionStore(store *selection.Store) {
	m.selectionStore = store
	if store == nil {
		return
	}
	if ds, err := store.Load(m.rootPath); err != nil {
		log.Debug().Err(err).Msg("Failed to load saved file selections")
	} else {
		m.deselected = ds
	}
}

func (m *WizardModel) persistSelections() {
	if m.selectionStore == nil {
		return
	}
	tree := m.getFileTree()
	if tree == nil {
		return
	}
	deselected := scanner.CollectDeselected(tree, m.getSelectedFiles())
	if err := m.selectionStore.Save(m.rootPath, deselected); err != nil {
		log.Debug().Err(err).Msg("Failed to persist file selections")
		return
	}
	m.deselected = deselected
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
			maxSize := m.wizardConfig.Context.MaxSize
			m.fileSelection = screens.NewFileSelection(m.getFileTree(), m.getSelectedFiles(), maxSize)
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

	maxSize, err := utils.ParseSize(maxSizeStr)
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
		"Terminal too small (need ≥%dx%d)\n\nCurrent: %dx%d",
		minTerminalWidth, minTerminalHeight,
		m.width, m.height,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.WarningStyle.Render(msg),
	)
}
