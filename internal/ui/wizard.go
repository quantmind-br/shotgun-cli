package ui

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo464/shotgun-cli/internal/core/scanner"
	"github.com/diogo464/shotgun-cli/internal/core/template"
	"github.com/diogo464/shotgun-cli/internal/core/context"
	"github.com/diogo464/shotgun-cli/internal/platform/clipboard"
	"github.com/diogo464/shotgun-cli/internal/ui/screens"
	"github.com/diogo464/shotgun-cli/internal/ui/components"
	"github.com/diogo464/shotgun-cli/internal/ui/styles"
)

const (
	StepFileSelection = 1
	StepTemplateSelection = 2
	StepTaskInput = 3
	StepRulesInput = 4
	StepReview = 5
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

	rootPath     string
	config       *scanner.Config

	fileSelection     *screens.FileSelectionModel
	templateSelection *screens.TemplateSelectionModel
	taskInput         *screens.TaskInputModel
	rulesInput        *screens.RulesInputModel
	review            *screens.ReviewModel

	progressComponent *components.ProgressModel
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

func NewWizard(rootPath string, config *scanner.Config) *WizardModel {
	return &WizardModel{
		step:          StepFileSelection,
		selectedFiles: make(map[string]bool),
		rootPath:      rootPath,
		config:        config,
		progressComponent: components.NewProgress(),
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		scanDirectoryCmd(m.rootPath, m.config),
	)
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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+q":
			return m, tea.Quit
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

	case GenerationCompleteMsg:
		m.progress.Visible = false
		cmd = clipboard.CopyCmd(msg.Content)
		cmds = append(cmds, cmd)

	case GenerationErrorMsg:
		m.error = msg.Err
		m.progress.Visible = false

	case ClipboardCompleteMsg:
		if !msg.Success && msg.Err != nil {
			m.error = fmt.Errorf("clipboard copy failed: %v", msg.Err)
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, cmd
}

func (m *WizardModel) View() string {
	if m.progress.Visible {
		return m.progressComponent.View()
	}

	if m.error != nil {
		return styles.RenderError(m.error.Error())
	}

	switch m.step {
	case StepFileSelection:
		if m.fileSelection != nil {
			return m.fileSelection.View()
		}
		return "Loading files..."
	case StepTemplateSelection:
		if m.templateSelection != nil {
			return m.templateSelection.View()
		}
		return "Loading templates..."
	case StepTaskInput:
		if m.taskInput != nil {
			return m.taskInput.View()
		}
		return "Initializing task input..."
	case StepRulesInput:
		if m.rulesInput != nil {
			return m.rulesInput.View()
		}
		return "Initializing rules input..."
	case StepReview:
		if m.review != nil {
			return m.review.View()
		}
		return "Initializing review..."
	default:
		return "Unknown step"
	}
}

func (m *WizardModel) canAdvanceStep() bool {
	switch m.step {
	case StepFileSelection:
		return len(m.selectedFiles) > 0
	case StepTemplateSelection:
		return m.template != nil
	case StepTaskInput:
		return len(m.taskDesc) > 0
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

func scanDirectoryCmd(rootPath string, config *scanner.Config) tea.Cmd {
	return func() tea.Msg {
		scanner := scanner.NewFileSystemScanner()

		progressCh := make(chan scanner.Progress, 100)
		done := make(chan bool)

		go func() {
			defer close(done)
			tree, err := scanner.ScanWithProgress(rootPath, config, progressCh)
			if err != nil {
				return
			}

			select {
			case <-done:
				return
			default:
				tea.Send(ScanCompleteMsg{Tree: tree})
			}
		}()

		for {
			select {
			case progress, ok := <-progressCh:
				if !ok {
					<-done
					return nil
				}
				tea.Send(ScanProgressMsg{
					Current: progress.Current,
					Total:   progress.Total,
					Stage:   progress.Stage,
				})
			case <-done:
				return nil
			}
		}
	}
}

func generateContextCmd(fileTree *scanner.FileNode, selectedFiles map[string]bool, template *template.Template, taskDesc, rules, rootPath string) tea.Cmd {
	return func() tea.Msg {
		generator := context.NewGenerator()

		progressCh := make(chan context.GenProgress, 100)
		done := make(chan bool)

		go func() {
			defer close(done)

			config := &context.Config{
				SelectedFiles: selectedFiles,
				Template:      template,
				TaskDesc:      taskDesc,
				Rules:         rules,
			}

			content, err := generator.GenerateWithProgressEx(fileTree, config, progressCh)
			if err != nil {
				tea.Send(GenerationErrorMsg{Err: err})
				return
			}

			timestamp := time.Now().Format("20060102-150405")
			filename := fmt.Sprintf("shotgun-prompt-%s.md", timestamp)
			filePath := filepath.Join(rootPath, filename)

			if err := writeFile(filePath, content); err != nil {
				tea.Send(GenerationErrorMsg{Err: err})
				return
			}

			tea.Send(GenerationCompleteMsg{
				Content:  content,
				FilePath: filePath,
			})
		}()

		for {
			select {
			case progress, ok := <-progressCh:
				if !ok {
					<-done
					return nil
				}
				tea.Send(GenerationProgressMsg{
					Stage:   progress.Stage,
					Message: progress.Message,
				})
			case <-done:
				return nil
			}
		}
	}
}

func writeFile(path, content string) error {
	return nil
}