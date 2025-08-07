package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	configMgr     *core.ConfigManager
	keyMgr        *core.SecureKeyManager
	errors        map[string]string
	showHelp      bool
	// Manual text tracking to avoid textinput bugs
	editingText string
	// Operation status tracking
	lastOperationStatus  string
	lastOperationSuccess bool
	lastOperationDetails string
}

// NewConfigFormModel creates a new configuration form
func NewConfigFormModel(config *core.Config, configMgr *core.ConfigManager, keyMgr *core.SecureKeyManager) ConfigFormModel {
	form := ConfigFormModel{
		config:    config,
		configMgr: configMgr,
		keyMgr:    keyMgr,
		errors:    make(map[string]string),
		showHelp:  false,
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
				Value:       m.config.OpenAI.APIKey,
				Type:        FieldText,
				Placeholder: "sk-...",
				Required:    true,
				Masked:      false,
				Input:       m.createTextInput("Enter your OpenAI API key", false),
				Help:        "Your OpenAI API key. Will be stored directly in the configuration file.",
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
				Type:        FieldText,
				Placeholder: "gpt-4o",
				Required:    true,
				Input:       m.createTextInput("Model name", false),
				Help:        "The OpenAI model to use for translation. Enter any model name (e.g., gpt-4o, claude-3-sonnet).",
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
				Type:        FieldText,
				Placeholder: "en",
				Required:    false,
				Input:       m.createTextInput("Target language code", false),
				Help:        "Target language for translations. Common codes: en, es, fr, de, it, pt, ru, ja, ko, zh",
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
		case "ctrl+c":
			return m, tea.Quit

		case "f1":
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

					// Initialize editing text with the actual value
					m.editingText = field.Value

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

		case "f2":
			if !m.editing {
				return m, m.saveConfiguration()
			}

		case "f3":
			if !m.editing {
				return m, m.resetConfiguration()
			}

		case "f4":
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

	field.Value = inputValue

	// Clear any previous error for this field
	delete(m.errors, field.Label)

	// Validate field
	if field.Required && strings.TrimSpace(field.Value) == "" {
		m.errors[field.Label] = "This field is required"
	}
}

// Form data extraction and validation methods

// extractConfigurationData converts form field values into a Config struct
func (m *ConfigFormModel) extractConfigurationData() (*core.Config, error) {
	config := &core.Config{
		Version:     "1.0",
		LastUpdated: time.Now(),
	}

	// Helper function to get field value by label within a section
	getFieldValue := func(sectionIdx int, label string) string {
		for _, field := range m.sections[sectionIdx].Fields {
			if field.Label == label {
				return field.Value
			}
		}
		return ""
	}

	// Helper function to parse integer with fallback
	parseInt := func(value string, fallback int) int {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
		return fallback
	}

	// Helper function to parse float with fallback
	parseFloat := func(value string, fallback float64) float64 {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
		return fallback
	}

	// Helper function to parse boolean with fallback
	parseBool := func(value string, fallback bool) bool {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
		return fallback
	}

	// Extract OpenAI Configuration (section 0)
	config.OpenAI = core.OpenAIConfig{
		APIKey:      getFieldValue(0, "API Key"),
		BaseURL:     getFieldValue(0, "Base URL"),
		Model:       getFieldValue(0, "Model"),
		Timeout:     parseInt(getFieldValue(0, "Timeout (seconds)"), 300),
		MaxTokens:   parseInt(getFieldValue(0, "Max Tokens"), 4096),
		Temperature: parseFloat(getFieldValue(0, "Temperature"), 0.7),
		MaxRetries:  3, // Default value
		RetryDelay:  2, // Default value
	}

	// Extract Translation Configuration (section 1)
	config.Translation = core.TranslationConfig{
		Enabled:        parseBool(getFieldValue(1, "Enable Translation"), false),
		TargetLanguage: getFieldValue(1, "Target Language"),
		ContextPrompt:  getFieldValue(1, "Custom Translation Prompt"),
	}

	// Extract App Configuration (section 2)
	config.App = core.AppConfig{
		AutoSave:        parseBool(getFieldValue(2, "Auto Save Config"), true),
		ShowLineNumbers: parseBool(getFieldValue(2, "Show Line Numbers"), true),
	}

	return config, nil
}

// extractAPIKeyData gets the API key from the form
func (m *ConfigFormModel) extractAPIKeyData() (string, error) {
	// Look for API Key field in OpenAI section (section 0)
	for _, field := range m.sections[0].Fields {
		if field.Label == "API Key" {
			apiKey := strings.TrimSpace(field.Value)
			if apiKey == "" {
				return "", fmt.Errorf("API key is required")
			}
			return apiKey, nil
		}
	}
	return "", fmt.Errorf("API key field not found")
}

// validateFormData performs comprehensive validation of all form fields
func (m *ConfigFormModel) validateFormData() map[string]string {
	errors := make(map[string]string)

	// Helper function to get field value by label within a section
	getFieldValue := func(sectionIdx int, label string) string {
		for _, field := range m.sections[sectionIdx].Fields {
			if field.Label == label {
				return field.Value
			}
		}
		return ""
	}

	// Validate OpenAI Configuration
	baseURL := strings.TrimSpace(getFieldValue(0, "Base URL"))
	if baseURL == "" {
		errors["Base URL"] = "Base URL is required"
	} else if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		errors["Base URL"] = "Base URL must start with http:// or https://"
	}

	model := strings.TrimSpace(getFieldValue(0, "Model"))
	if model == "" {
		errors["Model"] = "Model is required"
	}

	timeoutStr := getFieldValue(0, "Timeout (seconds)")
	if timeout, err := strconv.Atoi(timeoutStr); err != nil {
		errors["Timeout (seconds)"] = "Timeout must be a valid number"
	} else if timeout <= 0 {
		errors["Timeout (seconds)"] = "Timeout must be greater than 0"
	} else if timeout > 3600 {
		errors["Timeout (seconds)"] = "Timeout cannot exceed 3600 seconds (1 hour)"
	}

	maxTokensStr := getFieldValue(0, "Max Tokens")
	if maxTokens, err := strconv.Atoi(maxTokensStr); err != nil {
		errors["Max Tokens"] = "Max Tokens must be a valid number"
	} else if maxTokens <= 0 {
		errors["Max Tokens"] = "Max Tokens must be greater than 0"
	} else if maxTokens > 128000 {
		errors["Max Tokens"] = "Max Tokens cannot exceed 128000"
	}

	temperatureStr := getFieldValue(0, "Temperature")
	if temperature, err := strconv.ParseFloat(temperatureStr, 64); err != nil {
		errors["Temperature"] = "Temperature must be a valid number"
	} else if temperature < 0.0 || temperature > 2.0 {
		errors["Temperature"] = "Temperature must be between 0.0 and 2.0"
	}

	// Validate Translation Configuration
	targetLang := getFieldValue(1, "Target Language")
	if targetLang != "" {
		validLangs := []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"}
		valid := false
		for _, lang := range validLangs {
			if targetLang == lang {
				valid = true
				break
			}
		}
		if !valid {
			errors["Target Language"] = "Invalid target language"
		}
	}

	// Validate App Configuration

	return errors
}

// Configuration operations
func (m *ConfigFormModel) saveConfiguration() tea.Cmd {
	return func() tea.Msg {
		// 1. Validate form data
		errors := m.validateFormData()
		if len(errors) > 0 {
			return configSavedMsg{
				success: false,
				message: "Configuration validation failed",
				errors:  errors,
			}
		}

		// 2. Extract configuration and API key
		config, err := m.extractConfigurationData()
		if err != nil {
			return configSavedMsg{
				success: false,
				message: fmt.Sprintf("Failed to extract configuration: %v", err),
				errors:  make(map[string]string),
			}
		}

		apiKey, err := m.extractAPIKeyData()
		if err != nil {
			return configSavedMsg{
				success: false,
				message: fmt.Sprintf("Failed to extract API key: %v", err),
				errors:  make(map[string]string),
			}
		}

		// 3. Store API key directly in config (no keyring needed)
		if apiKey != "" {
			config.OpenAI.APIKey = apiKey
		}

		// 4. Update ConfigManager
		if err := m.configMgr.Update(config); err != nil {
			return configSavedMsg{
				success: false,
				message: fmt.Sprintf("Failed to update configuration: %v", err),
				errors:  make(map[string]string),
			}
		}

		// 5. Save configuration file
		if err := m.configMgr.Save(); err != nil {
			return configSavedMsg{
				success: false,
				message: fmt.Sprintf("Failed to save configuration file: %v", err),
				errors:  make(map[string]string),
			}
		}

		// 6. Return success message
		return configSavedMsg{
			success: true,
			message: "Configuration saved successfully",
			errors:  make(map[string]string),
		}
	}
}

func (m *ConfigFormModel) resetConfiguration() tea.Cmd {
	return func() tea.Msg {
		// 1. Load default configuration
		defaultConfig := core.DefaultConfig()

		// 2. Update the local config reference
		m.config = defaultConfig

		// 3. Reinitialize form sections with default values
		m.initializeSections()

		// 4. Clear any validation errors
		m.errors = make(map[string]string)

		// 5. Reset editing state
		m.editing = false
		m.editingText = ""

		// 6. Reset form navigation to first field
		m.activeSection = 0
		m.activeField = 0

		return configResetMsg{
			success: true,
			message: "Configuration reset to defaults successfully",
		}
	}
}

func (m *ConfigFormModel) testConnection() tea.Cmd {
	return func() tea.Msg {
		// 1. Validate required fields for connection
		requiredFields := []string{"API Key", "Base URL", "Model"}
		for _, sectionFields := range m.sections {
			for _, field := range sectionFields.Fields {
				for _, required := range requiredFields {
					if field.Label == required && field.Required {
						value := field.Value
						if strings.TrimSpace(value) == "" {
							return connectionTestMsg{
								success: false,
								message: fmt.Sprintf("Missing required field: %s", field.Label),
								details: "Please fill in all required fields before testing",
							}
						}
					}
				}
			}
		}

		// 2. Extract configuration for testing
		testConfig, err := m.extractConfigurationData()
		if err != nil {
			return connectionTestMsg{
				success: false,
				message: "Failed to extract configuration",
				details: fmt.Sprintf("Error: %v", err),
			}
		}

		// 3. Check API key availability
		apiKey, err := m.extractAPIKeyData()
		if err != nil {
			return connectionTestMsg{
				success: false,
				message: "Failed to retrieve API key",
				details: fmt.Sprintf("Error: %v", err),
			}
		}

		// Use direct API key from config if available
		if apiKey == "" && testConfig.OpenAI.APIKey != "" {
			apiKey = testConfig.OpenAI.APIKey
		}

		if apiKey == "" {
			return connectionTestMsg{
				success: false,
				message: "No API key available for testing",
				details: "Please enter an API key",
			}
		}

		// 4. Create temporary enhanced translator for testing
		enhancedConfig := core.FromLegacyConfig(testConfig)
		translator, err := core.NewEnhancedTranslationService(enhancedConfig, m.keyMgr)
		if err != nil {
			return connectionTestMsg{
				success: false,
				message: "Failed to create translator",
				details: fmt.Sprintf("Error: %v", err),
			}
		}

		// 5. Test API connectivity by attempting a simple translation
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Use a simple test translation to verify connectivity
		testResult, err := translator.TranslateText(ctx, "Test connection", "test")
		if err != nil {
			return connectionTestMsg{
				success: false,
				message: "Connection test failed",
				details: fmt.Sprintf("Error: %v", err),
			}
		}

		// Check if we got a reasonable response
		if testResult == nil || testResult.TranslatedText == "" {
			return connectionTestMsg{
				success: false,
				message: "Connection test failed",
				details: "Received empty response from API",
			}
		}

		// 6. Return success with connection details
		return connectionTestMsg{
			success: true,
			message: "Connection test successful",
			details: fmt.Sprintf("Successfully connected to %s using model %s",
				testConfig.OpenAI.BaseURL, testConfig.OpenAI.Model),
		}
	}
}

// Message types for configuration operations
type configSavedMsg struct {
	success bool
	message string
	errors  map[string]string
}

type configResetMsg struct {
	success bool
	message string
}

type connectionTestMsg struct {
	success bool
	message string
	details string
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
