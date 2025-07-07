package ui

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"shotgun-cli/internal/file"
	"shotgun-cli/internal/template"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	taskSelectionScreen screen = iota
	taskInputScreen
	rulesInputScreen
	fileSelectionScreen
	processingScreen
	completedScreen
)

type taskType struct {
	name        string
	description string
	template    string
}

func (t taskType) Title() string       { return t.name }
func (t taskType) Description() string { return t.description }
func (t taskType) FilterValue() string { return t.name }

type App struct {
	currentScreen screen
	taskList      list.Model
	selectedTask  taskType
	taskInput     string
	rulesInput    string
	wantsRules    bool
	fileTree      *FileTreeModel
	scanner       *file.Scanner
	processor     *template.Processor
	outputFile    string
	width         int
	height        int
}

var taskTypes = []list.Item{
	taskType{
		name:        "architect",
		description: "Design system architecture and create plans",
		template:    "prompt_makePlan.md",
	},
	taskType{
		name:        "dev",
		description: "Generate code changes and implementations",
		template:    "prompt_makeDiffGitFormat.md",
	},
	taskType{
		name:        "find bug",
		description: "Analyze and debug code issues",
		template:    "prompt_analyzeBug.md",
	},
	taskType{
		name:        "docs-sync",
		description: "Synchronize documentation with code",
		template:    "prompt_projectManager.md",
	},
}

func NewApp() *App {
	taskList := list.New(taskTypes, list.NewDefaultDelegate(), 0, 0)
	taskList.Title = "Select Task Type"
	taskList.SetShowStatusBar(false)
	taskList.SetFilteringEnabled(false)
	
	// Desabilitar todas as teclas do list exceto navegação e enter
	taskList.KeyMap.CursorUp.SetKeys("up", "k")
	taskList.KeyMap.CursorDown.SetKeys("down", "j")
	taskList.KeyMap.Quit.SetKeys("q", "ctrl+c")
	// Remover space das teclas do list
	taskList.KeyMap.ShowFullHelp.SetKeys()
	taskList.KeyMap.CloseFullHelp.SetKeys()
	
	taskList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginLeft(2)

	cwd, _ := os.Getwd()
	scanner, _ := file.NewScanner(cwd)
	processor := template.NewProcessor()

	return &App{
		currentScreen: taskSelectionScreen,
		taskList:      taskList,
		scanner:       scanner,
		processor:     processor,
	}
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.taskList.SetSize(msg.Width-4, msg.Height-4)
		if a.fileTree != nil {
			a.fileTree.width = msg.Width
			a.fileTree.height = msg.Height
		}
		return a, nil

	case fileTreeReady:
		return a, nil
	case processingComplete:
		return a, nil
	case fileError:
		return a, tea.Quit
	case processingError:
		return a, tea.Quit

	case tea.KeyMsg:
		switch a.currentScreen {
		case taskSelectionScreen:
			return a.handleTaskSelection(msg)
		case taskInputScreen:
			return a.handleTaskInput(msg)
		case rulesInputScreen:
			return a.handleRulesInput(msg)
		case fileSelectionScreen:
			return a.handleFileSelection(msg)
		case completedScreen:
			return a, tea.Quit
		}

	}

	return a, nil
}

func (a *App) handleTaskSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return a, tea.Quit
	case "enter":
		if selectedItem, ok := a.taskList.SelectedItem().(taskType); ok {
			a.selectedTask = selectedItem
			a.currentScreen = taskInputScreen
			// Limpar qualquer input anterior
			a.taskInput = ""
			return a, nil
		}
	}

	// Só permitir que o List processe teclas de navegação específicas
	switch msg.String() {
	case "up", "down", "k", "j":
		var cmd tea.Cmd
		a.taskList, cmd = a.taskList.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) handleTaskInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle key combinations FIRST (Alt+D, Ctrl+C, etc.)
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.currentScreen = taskSelectionScreen
		return a, nil
	case "alt+d":
		if strings.TrimSpace(a.taskInput) != "" {
			a.currentScreen = rulesInputScreen
			return a, nil
		}
		return a, nil
	case " ":
		// Explicit space handling (Bubble Tea best practice)
		a.taskInput += " "
		return a, nil
	}

	// Handle special keys
	switch msg.Type {
	case tea.KeySpace:
		// Handle tea.KeySpace explicitly for maximum compatibility
		a.taskInput += " "
		return a, nil
	case tea.KeyBackspace:
		if len(a.taskInput) > 0 {
			a.taskInput = a.taskInput[:len(a.taskInput)-1]
		}
		return a, nil
	case tea.KeyEnter:
		a.taskInput += "\n"
		return a, nil
	case tea.KeyRunes:
		// Handle regular character input (excluding spaces, already handled above)
		a.taskInput += string(msg.Runes)
		return a, nil
	}

	return a, nil
}

func (a *App) handleRulesInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle key combinations FIRST
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.currentScreen = taskInputScreen
		return a, nil
	case "y", "Y":
		if !a.wantsRules {
			a.wantsRules = true
			return a, nil
		}
	case "n", "N":
		if !a.wantsRules {
			a.rulesInput = "1. There's no user rules."
			return a, a.initFileSelection()
		}
	case "alt+d":
		if a.wantsRules && strings.TrimSpace(a.rulesInput) != "" {
			return a, a.initFileSelection()
		}
	case " ":
		// Explicit space handling for rules input (Bubble Tea best practice)
		if a.wantsRules {
			a.rulesInput += " "
		}
		return a, nil
	}

	// Handle text input only when in rules input mode
	if a.wantsRules {
		// Handle special keys for text input
		switch msg.Type {
		case tea.KeySpace:
			// Handle tea.KeySpace explicitly for maximum compatibility
			a.rulesInput += " "
			return a, nil
		case tea.KeyBackspace:
			if len(a.rulesInput) > 0 {
				a.rulesInput = a.rulesInput[:len(a.rulesInput)-1]
			}
			return a, nil
		case tea.KeyEnter:
			a.rulesInput += "\n"
			return a, nil
		case tea.KeyRunes:
			// Handle regular character input (excluding spaces, already handled above)
			a.rulesInput += string(msg.Runes)
			return a, nil
		}
	}

	return a, nil
}

func (a *App) initFileSelection() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			root, err := a.scanner.ScanDirectory()
			if err != nil {
				return fileError{err}
			}
			a.fileTree = NewFileTreeModel(root)
			a.currentScreen = fileSelectionScreen
			return fileTreeReady{}
		},
	)
}

func (a *App) handleFileSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	case "esc":
		a.currentScreen = rulesInputScreen
		return a, nil
	case "alt+d":
		return a, a.processAndGenerate()
	}

	if a.fileTree != nil {
		var cmd tea.Cmd
		m, cmd := a.fileTree.Update(msg)
		if ftm, ok := m.(*FileTreeModel); ok {
			a.fileTree = ftm
		}
		return a, cmd
	}

	return a, nil
}

func (a *App) processAndGenerate() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			a.currentScreen = processingScreen

			templateContent, err := a.processor.LoadTemplate(a.selectedTask.template)
			if err != nil {
				return processingError{err}
			}

			fileStructure := a.scanner.GenerateFileStructure(a.fileTree.root)

			promptData := template.PromptData{
				Task:          a.taskInput,
				Rules:         a.rulesInput,
				FileStructure: fileStructure,
			}

			finalPrompt := a.processor.ProcessTemplate(templateContent, promptData)

			outputFile, err := a.processor.SavePrompt(finalPrompt)
			if err != nil {
				return processingError{err}
			}

			a.outputFile = outputFile
			a.currentScreen = completedScreen
			return processingComplete{}
		},
	)
}

type fileError struct{ error }
type fileTreeReady struct{}
type processingError struct{ error }
type processingComplete struct{}

func (a *App) View() string {
	switch a.currentScreen {
	case taskSelectionScreen:
		return a.renderTaskSelection()
	case taskInputScreen:
		return a.renderTaskInput()
	case rulesInputScreen:
		return a.renderRulesInput()
	case fileSelectionScreen:
		return a.renderFileSelection()
	case processingScreen:
		return a.renderProcessing()
	case completedScreen:
		return a.renderCompleted()
	}
	return ""
}

func (a *App) renderTaskSelection() string {
	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		a.taskList.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Use ↑/↓ to navigate, Enter to select, q to quit"),
		"",
	)
}

func (a *App) renderTaskInput() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	lines := strings.Split(a.taskInput, "\n")
	numberedLines := make([]string, len(lines))
	for i, line := range lines {
		numberedLines[i] = fmt.Sprintf("%3d. %s", i+1, line)
	}

	content := fmt.Sprintf("Task: %s\n\n%s\n\n%s",
		a.selectedTask.name,
		strings.Join(numberedLines, "\n"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Alt+D when done, Esc to go back"),
	)

	return style.Render(content)
}

func (a *App) renderRulesInput() string {
	if !a.wantsRules {
		return fmt.Sprintf("\n%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("Do you want to add project rules?"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Y for yes, N for no"),
			"",
		)
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(a.width - 4).
		Height(a.height - 8)

	lines := strings.Split(a.rulesInput, "\n")
	numberedLines := make([]string, len(lines))
	for i, line := range lines {
		numberedLines[i] = fmt.Sprintf("%3d. %s", i+1, line)
	}

	content := fmt.Sprintf("Project Rules:\n\n%s\n\n%s",
		strings.Join(numberedLines, "\n"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Press Alt+D when done, Esc to go back"),
	)

	return style.Render(content)
}

func (a *App) renderFileSelection() string {
	if a.fileTree == nil {
		return fmt.Sprintf("\n%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("File Selection"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Loading file tree..."),
			"Please wait...",
		)
	}

	header := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("File Selection")
	subheader := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("✓ = Selected (will be included) | ✗ = Deselected (will be excluded)")
	
	return fmt.Sprintf("\n%s\n%s\n\n%s",
		header,
		subheader,
		a.fileTree.View(),
	)
}

func (a *App) renderProcessing() string {
	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("Processing..."),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Generating your prompt..."),
		"",
	)
}

func (a *App) renderCompleted() string {
	message := "Your prompt has been generated successfully."
	if a.outputFile != "" {
		message = fmt.Sprintf("Your prompt has been saved to: %s", a.outputFile)
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("✓ Complete!"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(message),
		"Press any key to exit",
	)
}

func (a *App) Run() error {
	// Configuração mais robusta para Windows
	options := []tea.ProgramOption{}
	
	// No Windows, evitar AltScreen se houver problemas de TTY
	if runtime.GOOS == "windows" {
		// Tentar detectar se estamos em um terminal válido
		if os.Getenv("TERM") == "" {
			os.Setenv("TERM", "xterm-256color")
		}
		options = append(options, tea.WithoutSignalHandler())
	} else {
		options = append(options, tea.WithAltScreen())
	}
	
	p := tea.NewProgram(a, options...)
	_, err := p.Run()
	return err
}
