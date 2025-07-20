package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// ViewState represents the current view/step in the application
type ViewState int

const (
	ViewFileExclusion ViewState = iota
	ViewPromptComposition
	ViewGeneration
	ViewComplete
)

// Model represents the main application state
type Model struct {
	// UI State
	currentView   ViewState
	width, height int

	// Components
	fileTree    FileTreeModel
	progressBar progress.Model
	taskInput   textinput.Model
	rulesInput  textarea.Model

	// Business Logic
	scanner   *core.DirectoryScanner
	generator *core.ContextGenerator
	templates *core.SimpleTemplateProcessor
	selection *core.SelectionState

	// Application State
	selectedDir     string
	currentTemplate string
	taskText        string
	rulesText       string
	fileTree_root   *core.FileNode
	includedFiles   []string

	// Generation state
	generating       bool
	generationCtx    context.Context
	generationCancel context.CancelFunc
	progress         core.ProgressUpdate
	outputPath       string

	// Error handling
	lastError error
	showHelp  bool
}

// NewModel creates a new application model
func NewModel() (*Model, error) {
	// Get and validate current working directory
	currentDir, err := getCurrentWorkingDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize working directory: %w", err)
	}

	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())

	// Initialize text input for task
	taskInput := textinput.New()
	taskInput.Placeholder = "Enter your task description..."
	taskInput.Focus()
	taskInput.CharLimit = 500
	taskInput.Width = 50

	// Initialize textarea for rules
	rulesInput := textarea.New()
	rulesInput.Placeholder = "Enter custom rules (optional)..."
	rulesInput.SetWidth(60)
	rulesInput.SetHeight(5)

	// Initialize core components
	scanner := core.NewDirectoryScanner()
	generator := core.NewContextGenerator(0) // No size limit
	templates := core.NewTemplateProcessor()
	selection := core.NewSelectionState()

	// Load templates from templates directory
	templatesDir := "templates"
	if err := templates.LoadTemplatesFromDirectory(templatesDir); err != nil {
		// Try relative path from executable
		if ex, err := os.Executable(); err == nil {
			exPath := filepath.Dir(ex)
			templatesDir = filepath.Join(exPath, "..", "templates")
			if err := templates.LoadTemplatesFromDirectory(templatesDir); err != nil {
				return nil, fmt.Errorf("failed to load templates: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load templates: %v", err)
		}
	}

	return &Model{
		currentView:     ViewFileExclusion,
		progressBar:     prog,
		taskInput:       taskInput,
		rulesInput:      rulesInput,
		scanner:         scanner,
		generator:       generator,
		templates:       templates,
		selection:       selection,
		selectedDir:     currentDir,
		currentTemplate: core.TemplateDevKey, // Default to dev template
		rulesText:       "no additional rules",
	}, nil
}

// getCurrentWorkingDirectory gets and validates the current working directory
func getCurrentWorkingDirectory() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot get current directory: %w", err)
	}
	
	// Validate directory is accessible
	if err := validateDirectoryAccess(dir); err != nil {
		return "", fmt.Errorf("current directory validation failed: %w", err)
	}
	
	return dir, nil
}

// validateDirectoryAccess validates that a directory can be accessed and read
func validateDirectoryAccess(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied accessing directory: %s", path)
		}
		return fmt.Errorf("cannot access directory %s: %w", path, err)
	}
	
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}
	
	// Check read permissions
	_, err = os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("cannot read directory contents: %w", err)
	}
	
	return nil
}

// ValidateCurrentDirectory validates the current directory (for main.go)
func ValidateCurrentDirectory() error {
	_, err := getCurrentWorkingDirectory()
	return err
}

// GetCurrentView returns the current view state (for testing)
func (m *Model) GetCurrentView() ViewState {
	return m.currentView
}

// GetSelectedDir returns the selected directory (for testing)
func (m *Model) GetSelectedDir() string {
	return m.selectedDir
}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.scanDirectory(), // Start scanning immediately
	)
}

// Update handles all application events
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progressBar.Width = msg.Width - 4

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.generating {
				if m.generationCancel != nil {
					m.generationCancel()
				}
			}
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			if m.currentView > ViewFileExclusion {
				m.currentView--
				return m, nil
			}
		}

		// Handle view-specific key events
		switch m.currentView {
		case ViewFileExclusion:
			return m.updateFileExclusion(msg)
		case ViewPromptComposition:
			return m.updatePromptComposition(msg)
		case ViewGeneration:
			return m.updateGeneration(msg)
		case ViewComplete:
			return m.updateComplete(msg)
		}


	case scanCompleteMsg:
		m.fileTree_root = msg.root
		m.fileTree = NewFileTreeModel(msg.root, m.selection)
		return m, nil

	case core.ProgressUpdate:
		m.progress = msg
		return m, nil

	case generationCompleteMsg:
		m.generating = false
		m.outputPath = string(msg)
		m.currentView = ViewComplete
		return m, nil

	case errorMsg:
		m.lastError = msg.err
		m.generating = false
		return m, nil
	}

	// Update components
	var cmd tea.Cmd

	if m.currentView == ViewFileExclusion {
		m.fileTree, cmd = m.fileTree.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.currentView == ViewPromptComposition {
		m.taskInput, cmd = m.taskInput.Update(msg)
		cmds = append(cmds, cmd)

		m.rulesInput, cmd = m.rulesInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the application
func (m *Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	switch m.currentView {
	case ViewFileExclusion:
		return m.renderFileExclusion()
	case ViewPromptComposition:
		return m.renderPromptComposition()
	case ViewGeneration:
		return m.renderGeneration()
	case ViewComplete:
		return m.renderComplete()
	default:
		return "Unknown view"
	}
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A49FA5"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)
)

func (m *Model) renderHelp() string {
	help := `
shotgun-cli - Help

KEYBOARD SHORTCUTS:
  General:
    q, Ctrl+C    Quit application
    ?            Toggle this help
    Esc          Go back to previous step

  File Exclusion:
    hjkl         Navigate file tree (vim-style)
    ↑↓←→         Navigate file tree (arrow keys)
    Space        Toggle file/directory exclusion
    c            Continue to next step
    r            Reset all exclusions
    a            Exclude all files
    A            Include all files

  Prompt Composition:
    Tab          Switch between template/task/rules
    Enter        Confirm and generate
    1-4          Select template (1=Dev, 2=Arch, 3=Debug, 4=PM)

Press ? or Esc to close this help.
`
	return helpStyle.Render(help)
}

// Message types for async operations
type scanCompleteMsg struct{ root *core.FileNode }
type generationCompleteMsg string
type errorMsg struct{ err error }
