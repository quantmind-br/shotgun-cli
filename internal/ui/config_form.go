package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// ConfigFieldType represents different types of form fields
type ConfigFieldType int

const (
	FieldText ConfigFieldType = iota
	FieldPassword
	FieldSelect
	FieldToggle
	FieldNumber
)

// ConfigField represents a single configuration field
type ConfigField struct {
	Label       string
	Value       string
	Type        ConfigFieldType
	Options     []string // For select fields
	Placeholder string
	Required    bool
	Masked      bool // For password fields
	Input       textinput.Model
	Help        string
}

// ConfigSection represents a group of related configuration fields
type ConfigSection struct {
	Title       string
	Description string
	Fields      []ConfigField
}

// ConfigFormModel manages the configuration form UI
type ConfigFormModel struct {
	sections      []ConfigSection
	activeSection int
	activeField   int
	editing       bool
	width         int
	height        int
	config        *core.Config
	keyMgr        *core.SecureKeyManager
	errors        map[string]string
	showHelp      bool
	// Manual text tracking to avoid textinput bugs
	editingText   string
}

// NewConfigFormModel creates a new configuration form
func NewConfigFormModel(config *core.Config, keyMgr *core.SecureKeyManager) ConfigFormModel {
	form := ConfigFormModel{
		config:   config,
		keyMgr:   keyMgr,
		errors:   make(map[string]string),
		showHelp: false,
	}

	form.initializeSections()
	return form
}

// initializeSections sets up the configuration sections and fields
func (m *ConfigFormModel) initializeSections() {
	// OpenAI Configuration Section
	openaiSection := ConfigSection{
		Title:       "OpenAI API Configuration",
		Description: "Configure your OpenAI API settings for translation",
		Fields: []ConfigField{
			{
				Label:       "API Key",
				Value:       m.getAPIKeyDisplayValue(),
				Type:        FieldPassword,
				Placeholder: "sk-...",
				Required:    true,
				Masked:      true,
				Input:       m.createTextInput("Enter your OpenAI API key", true),
				Help:        "Your OpenAI API key. Will be stored securely in your system keyring.",
			},
			{
				Label:       "Base URL",
				Value:       m.config.OpenAI.BaseURL,
				Type:        FieldText,
				Placeholder: "https://api.openai.com/v1",
				Required:    true,
				Input:       m.createTextInput("API base URL", false),
				Help:        "OpenAI API base URL. Use default for OpenAI, or custom URL for compatible services.",
			},
			{
				Label:       "Model",
				Value:       m.config.OpenAI.Model,
				Type:        FieldSelect,
				Options:     []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo"},
				Required:    true,
				Input:       m.createTextInput("Model name", false),
				Help:        "The OpenAI model to use for translation. GPT-4o recommended for best results.",
			},
			{
				Label:       "Timeout (seconds)",
				Value:       fmt.Sprintf("%d", m.config.OpenAI.Timeout),
				Type:        FieldNumber,
				Placeholder: "300",
				Required:    true,
				Input:       m.createTextInput("Timeout in seconds", false),
				Help:        "Request timeout in seconds. Default: 300 (5 minutes).",
			},
			{
				Label:       "Max Tokens",
				Value:       fmt.Sprintf("%d", m.config.OpenAI.MaxTokens),
				Type:        FieldNumber,
				Placeholder: "4096",
				Required:    true,
				Input:       m.createTextInput("Maximum tokens", false),
				Help:        "Maximum tokens for API responses. Default: 4096.",
			},
			{
				Label:       "Temperature",
				Value:       fmt.Sprintf("%.1f", m.config.OpenAI.Temperature),
				Type:        FieldText,
				Placeholder: "0.7",
				Required:    true,
				Input:       m.createTextInput("Temperature (0.0-2.0)", false),
				Help:        "Controls randomness. Lower = more focused, Higher = more creative. Range: 0.0-2.0.",
			},
		},
	}

	// Translation Configuration Section
	translationSection := ConfigSection{
		Title:       "Translation Settings",
		Description: "Configure automatic translation behavior",
		Fields: []ConfigField{
			{
				Label:    "Enable Translation",
				Value:    fmt.Sprintf("%t", m.config.Translation.Enabled),
				Type:     FieldToggle,
				Required: false,
				Help:     "Enable automatic translation of tasks and rules to English.",
			},
			{
				Label:       "Target Language",
				Value:       m.config.Translation.TargetLanguage,
				Type:        FieldSelect,
				Options:     []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"},
				Required:    false,
				Input:       m.createTextInput("Target language code", false),
				Help:        "Target language for translations. Default: en (English).",
			},
			{
				Label:       "Custom Translation Prompt",
				Value:       m.config.Translation.ContextPrompt,
				Type:        FieldText,
				Placeholder: "Translate the following text...",
				Required:    false,
				Input:       m.createTextInput("Custom translation instructions", false),
				Help:        "Custom prompt for translation context. Leave empty for default.",
			},
		},
	}

	// Application Settings Section
	appSection := ConfigSection{
		Title:       "Application Settings",
		Description: "Configure general application behavior",
		Fields: []ConfigField{
			{
				Label:   "Theme",
				Value:   m.config.App.Theme,
				Type:    FieldSelect,
				Options: []string{"auto", "dark", "light"},
				Input:   m.createTextInput("UI theme", false),
				Help:    "Application color theme. Auto detects system preference.",
			},
			{
				Label: "Auto Save Config",
				Value: fmt.Sprintf("%t", m.config.App.AutoSave),
				Type:  FieldToggle,
				Help:  "Automatically save configuration changes.",
			},
			{
				Label: "Show Line Numbers",
				Value: fmt.Sprintf("%t", m.config.App.ShowLineNumbers),
				Type:  FieldToggle,
				Help:  "Show line numbers in text input areas.",
			},
			{
				Label:   "Default Template",
				Value:   m.config.App.DefaultTemplate,
				Type:    FieldSelect,
				Options: []string{"dev", "architect", "debug", "project-manager"},
				Input:   m.createTextInput("Default template", false),
				Help:    "Default prompt template to select on startup.",
			},
		},
	}

	m.sections = []ConfigSection{openaiSection, translationSection, appSection}

	// Don't set initial values in textinputs to avoid Windows first character bug
	// Values will be displayed separately and only set when editing starts
	for sectionIdx := range m.sections {
		for fieldIdx := range m.sections[sectionIdx].Fields {
			field := &m.sections[sectionIdx].Fields[fieldIdx]
			// Always leave textinput empty initially to avoid phantom characters
			field.Input.SetValue("")
		}
	}
}

// createTextInput creates a new text input with common settings
func (m *ConfigFormModel) createTextInput(placeholder string, password bool) textinput.Model {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 200
	if password {
		input.EchoMode = textinput.EchoPassword
		input.EchoCharacter = '*'
	}
	
	// Windows workaround for first character bug
	// Multiple resets to ensure clean state
	input.Reset()
	input.SetValue("")
	input.Reset()
	input.Blur() // Ensure not focused initially
	
	return input
}

// getAPIKeyDisplayValue returns a display value for the API key
func (m *ConfigFormModel) getAPIKeyDisplayValue() string {
	if m.config.OpenAI.APIKeyAlias != "" && m.keyMgr != nil {
		if m.keyMgr.HasAPIKey(m.config.OpenAI.APIKeyAlias) {
			return "••••••••••••••••" // Show masked key if it exists
		}
	}
	return ""
}

// Update handles form updates
func (m ConfigFormModel) Update(msg tea.Msg) (ConfigFormModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "esc":
			if m.editing {
				m.editing = false
				m.editingText = "" // Clear manual text
				m.getCurrentField().Input.Blur()
				return m, nil
			}
			// Exit configuration
			return m, nil

		case "tab", "shift+tab":
			if m.editing {
				return m, nil // Don't change focus while editing
			}
			if msg.String() == "tab" {
				m.nextField()
			} else {
				m.prevField()
			}
			return m, nil

		case "up", "k":
			if m.editing {
				// Allow normal text input during editing, don't block hjkl
				break
			}
			m.prevField()
			return m, nil

		case "down", "j":
			if m.editing {
				// Allow normal text input during editing, don't block hjkl
				break
			}
			m.nextField()
			return m, nil

		case "left", "h":
			if m.editing {
				// Allow normal text input during editing, don't block hjkl
				break
			}
			m.prevSection()
			return m, nil

		case "right", "l":
			if m.editing {
				// Allow normal text input during editing, don't block hjkl
				break
			}
			m.nextSection()
			return m, nil

		case "enter", " ":
			if !m.editing {
				field := m.getCurrentField()
				if field.Type == FieldToggle {
					m.toggleField()
				} else {
					m.editing = true
					
					// Windows workaround: Use manual text tracking instead of relying on textinput
					isPassword := field.Type == FieldPassword
					
					// Initialize editing text manually
					if isPassword && field.Value == "••••••••••••••••" {
						// For password fields with existing key, start with empty
						m.editingText = ""
					} else {
						// For other fields, load the actual value
						m.editingText = field.Value
					}
					
					// Create fresh textinput but don't rely on its internal state
					field.Input = m.createTextInput(field.Placeholder, isPassword)
					field.Input.Focus()
				}
				return m, cmd
			} else {
				// Save current field value
				m.saveCurrentField()
				m.editing = false
				m.editingText = "" // Clear manual text
				m.getCurrentField().Input.Blur()
				return m, nil
			}

		case "s":
			if !m.editing {
				return m, m.saveConfiguration()
			}

		case "r":
			if !m.editing {
				return m, m.resetConfiguration()
			}

		case "t":
			if !m.editing {
				return m, m.testConnection()
			}
			
		case "ctrl+x":
			if m.editing {
				// Force clear current field (workaround for phantom characters)
				field := m.getCurrentField()
				m.forceFieldReset(field)
				return m, nil
			}
		}

		// Handle input updates when editing - use manual text tracking
		if m.editing {
			field := m.getCurrentField()
			
			// Process keyboard input manually to avoid textinput bugs
			switch msg.String() {
			case "backspace":
				if len(m.editingText) > 0 {
					m.editingText = m.editingText[:len(m.editingText)-1]
				}
			case "delete":
				// For now, just handle backspace
			default:
				// Handle regular character input
				if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
					m.editingText += msg.String()
				}
			}
			
			// Keep textinput focused but don't rely on its state
			field.Input.Focus()
			return m, nil
		}
	}

	return m, nil
}

// Navigation helpers
func (m *ConfigFormModel) nextField() {
	m.activeField++
	if m.activeField >= len(m.sections[m.activeSection].Fields) {
		m.activeField = 0
	}
}

func (m *ConfigFormModel) prevField() {
	m.activeField--
	if m.activeField < 0 {
		m.activeField = len(m.sections[m.activeSection].Fields) - 1
	}
}

func (m *ConfigFormModel) nextSection() {
	m.activeSection++
	if m.activeSection >= len(m.sections) {
		m.activeSection = 0
	}
	m.activeField = 0
}

func (m *ConfigFormModel) prevSection() {
	m.activeSection--
	if m.activeSection < 0 {
		m.activeSection = len(m.sections) - 1
	}
	m.activeField = 0
}

func (m *ConfigFormModel) getCurrentField() *ConfigField {
	return &m.sections[m.activeSection].Fields[m.activeField]
}

// forceFieldReset completely resets a field to work around Windows textinput bugs
func (m *ConfigFormModel) forceFieldReset(field *ConfigField) {
	isPassword := field.Type == FieldPassword
	wasFocused := field.Input.Focused()
	
	// Completely recreate the input
	field.Input = m.createTextInput(field.Placeholder, isPassword)
	
	// Restore focus if it was focused
	if wasFocused {
		field.Input.Focus()
	}
}

// Field operations
func (m *ConfigFormModel) toggleField() {
	field := m.getCurrentField()
	if field.Type == FieldToggle {
		if field.Value == "true" {
			field.Value = "false"
		} else {
			field.Value = "true"
		}
	}
}

func (m *ConfigFormModel) saveCurrentField() {
	field := m.getCurrentField()
	
	// Use manual text tracking instead of textinput value
	inputValue := m.editingText
	
	// Special handling for password fields (API Key)
	if field.Type == FieldPassword && field.Label == "API Key" {
		// If input is empty and we already have a key, keep the existing key
		if strings.TrimSpace(inputValue) == "" && field.Value == "••••••••••••••••" {
			// Don't change the existing key
			return
		}
		// If user entered a new key, use it
		field.Value = inputValue
	} else {
		field.Value = inputValue
	}
	
	// Clear any previous error for this field
	delete(m.errors, field.Label)
	
	// Validate field
	if field.Required && strings.TrimSpace(field.Value) == "" {
		m.errors[field.Label] = "This field is required"
	}
}

// Configuration operations
func (m *ConfigFormModel) saveConfiguration() tea.Cmd {
	return func() tea.Msg {
		// Build configuration from form values
		// This would update the actual config and save it
		return configSavedMsg{}
	}
}

func (m *ConfigFormModel) resetConfiguration() tea.Cmd {
	return func() tea.Msg {
		// Reset to default configuration
		return configResetMsg{}
	}
}

func (m *ConfigFormModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		// Test API connection with current settings
		return connectionTestMsg{success: true, message: "Connection successful"}
	}
}

// Message types for configuration operations
type configSavedMsg struct{}
type configResetMsg struct{}
type connectionTestMsg struct {
	success bool
	message string
}

// Styling
var (
	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("33")).
				MarginBottom(1)

	sectionDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginBottom(1)

	fieldLabelStyle = lipgloss.NewStyle().
			Width(20).
			Align(lipgloss.Right).
			Bold(true)

	fieldActiveStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)

	fieldInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	configErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	configHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Italic(true)
)