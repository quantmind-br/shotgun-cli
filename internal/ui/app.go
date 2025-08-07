package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// getUserTemplatesDir returns the user's custom templates directory using XDG standards
func getUserTemplatesDir() (string, error) {
	// Use XDG config directory following the standard: ~/.config/shotgun-cli/templates/
	templatesDir := filepath.Join(xdg.ConfigHome, "shotgun-cli", "templates")
	return templatesDir, nil
}

// ensureTemplatesDirectory creates the custom templates directory if it doesn't exist
func ensureTemplatesDirectory(templatesDir string) error {
	// Check if directory already exists
	if _, err := os.Stat(templatesDir); err == nil {
		log.Printf("Custom templates directory already exists: %s", templatesDir)
		return nil // Directory already exists
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check templates directory %s: %w", templatesDir, err)
	}

	// Create the directory with appropriate permissions
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory %s: %w", templatesDir, err)
	}

	log.Printf("Created custom templates directory: %s", templatesDir)
	return nil
}

// ViewState represents the current view/step in the application
type ViewState int

const (
	ViewFileExclusion ViewState = iota
	ViewTemplateSelection
	ViewTaskDescription
	ViewCustomRules
	ViewConfiguration
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
	taskInput   NumberedTextArea
	rulesInput  NumberedTextArea
	configForm  EnhancedConfigFormModel

	// Business Logic
	scanner    *core.DirectoryScanner
	generator  *core.ContextGenerator
	templates  *core.SimpleTemplateProcessor
	selection  *core.SelectionState
	configMgr  *core.EnhancedConfigManager
	keyMgr     *core.SecureKeyManager
	translator *core.EnhancedTranslationService

	// Application State
	selectedDir     string
	currentTemplate string
	templateIndex   int // Index for arrow key navigation
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

	// Translation state
	translating       bool
	taskTranslated    string
	rulesTranslated   string
	translationStatus string
	translationError  error

	// Error handling
	lastError error
	showHelp  bool
	ready     bool
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

	// Initialize numbered textarea for task (enhanced for full-screen)
	taskInput := NewNumberedTextArea()
	taskInput.SetPlaceholder("Enter your task description...")
	taskInput.Focus()
	taskInput.SetWidth(80) // Increased width for full-screen
	taskInput.SetHeight(8) // Increased height for more content

	// Initialize numbered textarea for rules (enhanced for full-screen)
	rulesInput := NewNumberedTextArea()
	rulesInput.SetPlaceholder("Enter custom rules (optional)...")
	rulesInput.SetWidth(80) // Increased width for full-screen
	rulesInput.SetHeight(8) // Increased height for more content

	// Initialize core components
	scanner := core.NewDirectoryScanner()
	generator := core.NewContextGenerator(0) // No size limit
	templates := core.NewTemplateProcessor()
	selection := core.NewSelectionState()

	// Initialize configuration components
	configMgr, err := core.NewEnhancedConfigManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize config manager: %w", err)
	}
	if err := configMgr.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	keyMgr, err := core.NewSecureKeyManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key manager: %w", err)
	}

	// Initialize translator with current configuration
	config := configMgr.GetEnhanced()

	var translator *core.EnhancedTranslationService
	if config.OpenAI.APIKey != "" {
		var err error
		translator, err = core.NewEnhancedTranslationService(config, keyMgr)
		if err != nil {
			// Translator initialization failed - log the reason for debugging
			// Note: This is not a fatal error, app can work without translation
			translator = nil
			// We could add debug logging here: fmt.Printf("Translation disabled: %v\n", err)
		}
	}

	// Get custom templates directory and ensure it exists
	customTemplatesDir, err := getUserTemplatesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get custom templates directory: %w", err)
	}

	// Ensure custom templates directory exists (non-breaking)
	if err := ensureTemplatesDirectory(customTemplatesDir); err != nil {
		log.Printf("Warning: Failed to create custom templates directory: %v", err)
		// Continue execution - this is not a fatal error
	}

	// Load templates from both builtin and custom directories
	builtinTemplatesDir := "templates"
	if err := templates.LoadTemplates(builtinTemplatesDir, customTemplatesDir); err != nil {
		// Try relative path from executable for builtin templates
		if ex, err := os.Executable(); err == nil {
			exPath := filepath.Dir(ex)
			builtinTemplatesDir = filepath.Join(exPath, "..", "templates")
			if err := templates.LoadTemplates(builtinTemplatesDir, customTemplatesDir); err != nil {
				return nil, fmt.Errorf("failed to load templates: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load templates: %v", err)
		}
	}

	// Initialize configuration form
	configForm := *NewEnhancedConfigFormModel(configMgr, keyMgr)

	return &Model{
		currentView:     ViewFileExclusion,
		progressBar:     prog,
		taskInput:       taskInput,
		rulesInput:      rulesInput,
		configForm:      configForm,
		scanner:         scanner,
		generator:       generator,
		templates:       templates,
		selection:       selection,
		configMgr:       configMgr,
		keyMgr:          keyMgr,
		translator:      translator,
		selectedDir:     currentDir,
		currentTemplate: core.TemplateDevKey, // Default to dev template
		templateIndex:   0,                   // Start with first template
		rulesText:       "no additional rules",
		ready:           false,
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
		m.configForm.SetSize(m.width, m.height)
		m.ready = true

		// Dynamically set the size of text areas for task and rules views
		// This is an approximation based on the static elements in those views.
		// Header: Title(1) + blank(1) + subtitle(1) + blank(1) + instructions(1) + blank(1) = 6 lines
		// Footer: blank(1) + examples(2) + blank(1) + help(1) = 5 lines
		// Translation status can add 2 more lines to the footer area.
		headerHeight := 6
		footerHeight := 7 // Assume translation might be active
		verticalMargin := 2

		textAreaHeight := m.height - headerHeight - footerHeight - verticalMargin
		if textAreaHeight < 5 {
			textAreaHeight = 5 // Minimum height
		}

		// Leave some horizontal margin
		textAreaWidth := m.width - 10
		if textAreaWidth < 40 {
			textAreaWidth = 40
		}
		m.taskInput.SetHeight(textAreaHeight)
		m.taskInput.SetWidth(textAreaWidth)
		m.rulesInput.SetHeight(textAreaHeight)
		m.rulesInput.SetWidth(textAreaWidth)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			if m.generating {
				if m.generationCancel != nil {
					m.generationCancel()
				}
			}
			return m, tea.Quit

		case "o":
			// Access configuration (options) (only from file exclusion view)
			if m.currentView == ViewFileExclusion {
				m.currentView = ViewConfiguration
				return m, nil
			}

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

				// If returning from config view, reload translator
				if m.currentView == ViewFileExclusion {
					config := m.configMgr.GetEnhanced()
					if config.OpenAI.APIKey != "" {
						if newTranslator, err := core.NewEnhancedTranslationService(config, m.keyMgr); err == nil {
							m.translator = newTranslator
						} else {
							m.translator = nil
						}
					} else {
						m.translator = nil
					}
				}
				return m, nil
			}
		}

		// Handle view-specific key events
		switch m.currentView {
		case ViewFileExclusion:
			return m.updateFileExclusion(msg)
		case ViewTemplateSelection:
			return m.updateTemplateSelection(msg)
		case ViewTaskDescription:
			return m.updateTaskDescription(msg)
		case ViewCustomRules:
			return m.updateCustomRules(msg)
		case ViewConfiguration:
			return m.updateConfiguration(msg)
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

	case translationCompleteMsg:
		m.translating = false
		m.translationError = msg.err

		if msg.err != nil {
			m.translationStatus = fmt.Sprintf("Translation failed: %s", msg.err.Error())
			// Continue with original text
		} else {
			switch msg.textType {
			case "task":
				m.taskTranslated = msg.result.TranslatedText
				m.translationStatus = "Task translated successfully"
				// Continue to rules input
				m.currentView = ViewCustomRules
				return m, m.rulesInput.Focus()
			case "rules":
				m.rulesTranslated = msg.result.TranslatedText
				m.translationStatus = "Rules translated successfully"
				// Use translated text for generation
				if m.taskTranslated != "" {
					m.taskText = m.taskTranslated
				}
				if m.rulesTranslated != "" {
					m.rulesText = m.rulesTranslated
				}
				// Continue to generation
				m.currentView = ViewGeneration
				return m, m.generatePrompt()
			}
		}
		return m, nil

	}

	// Update components
	var cmd tea.Cmd

	if m.currentView == ViewFileExclusion {
		m.fileTree, cmd = m.fileTree.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.currentView == ViewTaskDescription {
		m.taskInput, cmd = m.taskInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.currentView == ViewCustomRules {
		m.rulesInput, cmd = m.rulesInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.currentView == ViewConfiguration {
		m.configForm, cmd = m.configForm.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the application
func (m *Model) View() string {
	if !m.ready {
		return "Initializing, please wait..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	switch m.currentView {
	case ViewFileExclusion:
		return m.renderFileExclusion()
	case ViewTemplateSelection:
		return m.renderTemplateSelection()
	case ViewTaskDescription:
		return m.renderTaskDescription()
	case ViewCustomRules:
		return m.renderCustomRules()
	case ViewConfiguration:
		return m.configForm.View()
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

WORKFLOW: File Exclusion → Template Selection → Task Description → Custom Rules → Generation

KEYBOARD SHORTCUTS:
  General:
    Ctrl+Q, Ctrl+C    Quit application
    ?            Toggle this help
    Esc          Go back to previous step

  File Exclusion:
    hjkl         Navigate file tree (vim-style)
    ↑↓←→         Navigate file tree (arrow keys)
    Space        Toggle file/directory exclusion
    c            Continue to next step
    o            Access configuration/settings
    r            Reset all exclusions
    a            Exclude all files
    A            Include all files

  Template Selection:
    ↑/↓ (k/j)    Navigate template options
    1-4          Quick select template
    Enter        Confirm selection and continue

  Task Description:
    Tab          Focus/unfocus input field
    F5           Continue to next step

  Custom Rules:
    Tab          Focus/unfocus input field
    F5           Generate prompt

  Generation/Complete:
    Ctrl+C       Cancel generation
    Enter        Quit (when complete)
    n            Start new prompt (when complete)

Press ? or Esc to close this help.
`
	return helpStyle.Render(help)
}

// Message types for async operations
type scanCompleteMsg struct{ root *core.FileNode }
type generationCompleteMsg string
type errorMsg struct{ err error }
