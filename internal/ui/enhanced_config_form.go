package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"shotgun-cli/internal/core"
)

// EnhancedConfigFormModel represents an enhanced configuration form with Windows compatibility
type EnhancedConfigFormModel struct {
	sections          []EnhancedConfigSection
	activeSection     int
	activeField       int
	editing           bool
	width             int
	height            int
	config            *core.EnhancedConfig
	configMgr         *core.EnhancedConfigManager
	keyMgr            *core.SecureKeyManager
	errors            map[string]string
	showHelp          bool
	validationResults *core.ConfigValidationResult

	// Windows compatibility - enhanced manual text tracking
	editingText     string
	pasteDetected   bool
	inputResetCount int

	// Enhanced features
	connectionTesting  bool
	connectionResult   string
	connectionSuccess  bool
	realTimeValidation bool
	autoSave           bool
	lastAutoSave       time.Time

	// Operation status tracking
	lastOperationStatus  string
	lastOperationSuccess bool
	lastOperationDetails string

	// UI state
	showConnectionTest    bool
	showValidationDetails bool
}

// EnhancedConfigSection represents a section in the enhanced configuration form
type EnhancedConfigSection struct {
	Name        string
	Description string
	Fields      []EnhancedConfigField
	Icon        string
}

// EnhancedConfigField represents a field in the enhanced configuration form
type EnhancedConfigField struct {
	Label       string
	Key         string
	Type        ConfigFieldType
	Input       textinput.Model
	Value       string
	Options     []string // For select fields
	Required    bool
	Sensitive   bool // For password fields
	Placeholder string
	HelpText    string
	Validator   func(string) error // Custom validation function

	// Windows compatibility
	ManualValue   string
	LastResetTime time.Time
}

// NewEnhancedConfigFormModel creates a new enhanced configuration form
func NewEnhancedConfigFormModel(configMgr *core.EnhancedConfigManager, keyMgr *core.SecureKeyManager) *EnhancedConfigFormModel {
	model := &EnhancedConfigFormModel{
		configMgr:             configMgr,
		keyMgr:                keyMgr,
		errors:                make(map[string]string),
		realTimeValidation:    true,
		autoSave:              true,
		showConnectionTest:    true,
		showValidationDetails: true,
	}

	// Initialize with current configuration
	model.config = configMgr.GetEnhanced()
	model.initializeEnhancedSections()

	return model
}

// initializeEnhancedSections sets up the enhanced configuration sections
func (m *EnhancedConfigFormModel) initializeEnhancedSections() {
	m.sections = []EnhancedConfigSection{
		{
			Name:        "OpenAI API",
			Description: "Configure OpenAI API settings and authentication",
			Icon:        "🤖",
			Fields: []EnhancedConfigField{
				m.createEnhancedField("API Key Alias", "openai.apiKeyAlias", FieldText, m.config.OpenAI.APIKeyAlias, false, true, "Enter a name for this API key"),
				m.createEnhancedField("Base URL", "openai.baseUrl", FieldText, m.config.OpenAI.BaseURL, true, false, "OpenAI API base URL"),
				m.createEnhancedField("Model", "openai.model", FieldSelect, m.config.OpenAI.Model, true, false, "GPT model to use"),
				m.createEnhancedField("Timeout (seconds)", "openai.timeout", FieldNumber, fmt.Sprintf("%d", m.config.OpenAI.Timeout), true, false, "Request timeout in seconds"),
				m.createEnhancedField("Max Tokens", "openai.maxTokens", FieldNumber, fmt.Sprintf("%d", m.config.OpenAI.MaxTokens), true, false, "Maximum tokens per request"),
				m.createEnhancedField("Temperature", "openai.temperature", FieldNumber, fmt.Sprintf("%.1f", m.config.OpenAI.Temperature), true, false, "Randomness (0.0-2.0)"),
				m.createEnhancedField("Max Retries", "openai.maxRetries", FieldNumber, fmt.Sprintf("%d", m.config.OpenAI.MaxRetries), true, false, "Maximum retry attempts"),
				m.createEnhancedField("Retry Delay (seconds)", "openai.retryDelay", FieldNumber, fmt.Sprintf("%d", m.config.OpenAI.RetryDelay), true, false, "Delay between retries"),
			},
		},
		{
			Name:        "Translation",
			Description: "Configure automatic translation settings",
			Icon:        "🌐",
			Fields: []EnhancedConfigField{
				m.createEnhancedToggleField("Enable Translation", "translation.enabled", m.config.Translation.Enabled, "Enable automatic translation"),
				m.createEnhancedField("Target Language", "translation.targetLanguage", FieldSelect, m.config.Translation.TargetLanguage, true, false, "Target language for translation"),
				m.createEnhancedField("Context Prompt", "translation.contextPrompt", FieldText, m.config.Translation.ContextPrompt, false, false, "Custom translation context"),
				m.createEnhancedToggleField("Enable Cache", "translation.cacheEnabled", m.config.Translation.CacheEnabled, "Cache translation results"),
				m.createEnhancedField("Cache Size", "translation.cacheSize", FieldNumber, fmt.Sprintf("%d", m.config.Translation.CacheSize), true, false, "Maximum cached translations"),
				m.createEnhancedField("Cache TTL (seconds)", "translation.cacheTTL", FieldNumber, fmt.Sprintf("%d", m.config.Translation.CacheTTL), true, false, "Cache time-to-live"),
			},
		},
		{
			Name:        "Application",
			Description: "Configure application preferences and behavior",
			Icon:        "⚙️",
			Fields: []EnhancedConfigField{
				m.createEnhancedField("Theme", "app.theme", FieldSelect, m.config.App.Theme, true, false, "UI color theme"),
				m.createEnhancedToggleField("Auto Save", "app.autoSave", m.config.App.AutoSave, "Automatically save configuration changes"),
				m.createEnhancedToggleField("Show Line Numbers", "app.showLineNumbers", m.config.App.ShowLineNumbers, "Show line numbers in UI"),
				m.createEnhancedField("Default Template", "app.defaultTemplate", FieldSelect, m.config.App.DefaultTemplate, true, false, "Default prompt template"),
				m.createEnhancedField("Max File Size (MB)", "app.maxFileSize", FieldNumber, fmt.Sprintf("%.1f", float64(m.config.App.MaxFileSize)/1024/1024), true, false, "Maximum file size to process"),
				m.createEnhancedField("Max Directory Depth", "app.maxDirectoryDepth", FieldNumber, fmt.Sprintf("%d", m.config.App.MaxDirectoryDepth), true, false, "Maximum directory scan depth"),
				m.createEnhancedField("Worker Pool Size", "app.workerPoolSize", FieldNumber, fmt.Sprintf("%d", m.config.App.WorkerPoolSize), true, false, "Number of parallel workers"),
				m.createEnhancedField("Refresh Interval (ms)", "app.refreshInterval", FieldNumber, fmt.Sprintf("%d", m.config.App.RefreshInterval), true, false, "UI refresh interval"),
				m.createEnhancedToggleField("Enable Hot Reload", "app.enableHotReload", m.config.App.EnableHotReload, "Automatically reload on config changes"),
			},
		},
	}

	// Set up select field options
	m.setupSelectOptions()
}

// createEnhancedField creates an enhanced configuration field with Windows compatibility
func (m *EnhancedConfigFormModel) createEnhancedField(label, key string, fieldType ConfigFieldType, value string, required, sensitive bool, placeholder string) EnhancedConfigField {
	input := textinput.New()
	input.Placeholder = placeholder
	input.Width = 40

	// Windows compatibility - initialize with clean state
	input.Reset()
	input.SetValue("")
	input.Reset()
	input.Blur()

	if sensitive {
		input.EchoMode = textinput.EchoPassword
		input.EchoCharacter = '*'
	}

	field := EnhancedConfigField{
		Label:         label,
		Key:           key,
		Type:          fieldType,
		Input:         input,
		Value:         value,
		Required:      required,
		Sensitive:     sensitive,
		Placeholder:   placeholder,
		ManualValue:   value, // Initialize manual tracking with current value
		LastResetTime: time.Now(),
	}

	// Set custom validators based on field type
	field.Validator = m.getFieldValidator(key, fieldType)

	return field
}

// createEnhancedToggleField creates a toggle field for boolean values
func (m *EnhancedConfigFormModel) createEnhancedToggleField(label, key string, value bool, helpText string) EnhancedConfigField {
	valueStr := "false"
	if value {
		valueStr = "true"
	}

	field := m.createEnhancedField(label, key, FieldToggle, valueStr, false, false, "")
	field.HelpText = helpText
	return field
}

// setupSelectOptions configures options for select fields
func (m *EnhancedConfigFormModel) setupSelectOptions() {
	for sectionIdx := range m.sections {
		for fieldIdx := range m.sections[sectionIdx].Fields {
			field := &m.sections[sectionIdx].Fields[fieldIdx]

			switch field.Key {
			case "openai.model":
				field.Options = []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-4", "gpt-3.5-turbo", "gpt-3.5-turbo-16k"}
			case "translation.targetLanguage":
				field.Options = []string{"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh"}
			case "app.theme":
				field.Options = []string{"auto", "dark", "light"}
			case "app.defaultTemplate":
				field.Options = []string{"dev", "architect", "debug", "project-manager"}
			}
		}
	}
}

// getFieldValidator returns a validation function for the given field
func (m *EnhancedConfigFormModel) getFieldValidator(key string, fieldType ConfigFieldType) func(string) error {
	return func(value string) error {
		// Create a temporary config with the new value to test validation
		testConfig := m.config.Clone()

		// Apply the value to the appropriate field
		if err := m.applyValueToConfig(testConfig, key, value); err != nil {
			return err
		}

		// Validate the entire config
		validator := core.SetupEnhancedValidator()
		result := testConfig.Validate(validator)

		if !result.Valid {
			// Find errors related to this specific field
			for _, validationErr := range result.Errors {
				if strings.Contains(strings.ToLower(validationErr.Field), strings.Split(key, ".")[1]) {
					return fmt.Errorf("%s", validationErr.Message)
				}
			}
			// Return the first error if no specific match
			if len(result.Errors) > 0 {
				return fmt.Errorf("%s", result.Errors[0].Message)
			}
		}

		return nil
	}
}

// applyValueToConfig applies a field value to a configuration object
func (m *EnhancedConfigFormModel) applyValueToConfig(config *core.EnhancedConfig, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid field key: %s", key)
	}

	section, field := parts[0], parts[1]

	switch section {
	case "openai":
		return m.applyOpenAIValue(config, field, value)
	case "translation":
		return m.applyTranslationValue(config, field, value)
	case "app":
		return m.applyAppValue(config, field, value)
	default:
		return fmt.Errorf("unknown section: %s", section)
	}
}

// Helper methods to apply values to specific config sections
func (m *EnhancedConfigFormModel) applyOpenAIValue(config *core.EnhancedConfig, field, value string) error {
	switch field {
	case "apiKeyAlias":
		config.OpenAI.APIKeyAlias = value
	case "baseUrl":
		config.OpenAI.BaseURL = value
	case "model":
		config.OpenAI.Model = value
	case "timeout":
		if val, err := fmt.Sscanf(value, "%d", &config.OpenAI.Timeout); err != nil || val != 1 {
			return fmt.Errorf("invalid timeout value")
		}
	case "maxTokens":
		if val, err := fmt.Sscanf(value, "%d", &config.OpenAI.MaxTokens); err != nil || val != 1 {
			return fmt.Errorf("invalid max tokens value")
		}
	case "temperature":
		if val, err := fmt.Sscanf(value, "%f", &config.OpenAI.Temperature); err != nil || val != 1 {
			return fmt.Errorf("invalid temperature value")
		}
	case "maxRetries":
		if val, err := fmt.Sscanf(value, "%d", &config.OpenAI.MaxRetries); err != nil || val != 1 {
			return fmt.Errorf("invalid max retries value")
		}
	case "retryDelay":
		if val, err := fmt.Sscanf(value, "%d", &config.OpenAI.RetryDelay); err != nil || val != 1 {
			return fmt.Errorf("invalid retry delay value")
		}
	}
	return nil
}

func (m *EnhancedConfigFormModel) applyTranslationValue(config *core.EnhancedConfig, field, value string) error {
	switch field {
	case "enabled":
		config.Translation.Enabled = value == "true"
	case "targetLanguage":
		config.Translation.TargetLanguage = value
	case "contextPrompt":
		config.Translation.ContextPrompt = value
	case "cacheEnabled":
		config.Translation.CacheEnabled = value == "true"
	case "cacheSize":
		if val, err := fmt.Sscanf(value, "%d", &config.Translation.CacheSize); err != nil || val != 1 {
			return fmt.Errorf("invalid cache size value")
		}
	case "cacheTTL":
		if val, err := fmt.Sscanf(value, "%d", &config.Translation.CacheTTL); err != nil || val != 1 {
			return fmt.Errorf("invalid cache TTL value")
		}
	}
	return nil
}

func (m *EnhancedConfigFormModel) applyAppValue(config *core.EnhancedConfig, field, value string) error {
	switch field {
	case "theme":
		config.App.Theme = value
	case "autoSave":
		config.App.AutoSave = value == "true"
	case "showLineNumbers":
		config.App.ShowLineNumbers = value == "true"
	case "defaultTemplate":
		config.App.DefaultTemplate = value
	case "maxFileSize":
		var sizeMB float64
		if val, err := fmt.Sscanf(value, "%f", &sizeMB); err != nil || val != 1 {
			return fmt.Errorf("invalid file size value")
		}
		config.App.MaxFileSize = int64(sizeMB * 1024 * 1024)
	case "maxDirectoryDepth":
		if val, err := fmt.Sscanf(value, "%d", &config.App.MaxDirectoryDepth); err != nil || val != 1 {
			return fmt.Errorf("invalid directory depth value")
		}
	case "workerPoolSize":
		if val, err := fmt.Sscanf(value, "%d", &config.App.WorkerPoolSize); err != nil || val != 1 {
			return fmt.Errorf("invalid worker pool size value")
		}
	case "refreshInterval":
		if val, err := fmt.Sscanf(value, "%d", &config.App.RefreshInterval); err != nil || val != 1 {
			return fmt.Errorf("invalid refresh interval value")
		}
	case "enableHotReload":
		config.App.EnableHotReload = value == "true"
	}
	return nil
}

// Update handles enhanced input events with Windows compatibility
func (m EnhancedConfigFormModel) Update(msg tea.Msg) (EnhancedConfigFormModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyInput(msg)

	case connectionTestCompleteMsg:
		m.connectionTesting = false
		m.connectionResult = msg.result
		m.connectionSuccess = msg.success
		return m, nil

	case autoSaveCompleteMsg:
		m.lastAutoSave = time.Now()
		return m, nil
	}

	// Update the current field if not in custom input mode
	if m.editing && !m.pasteDetected {
		field := m.getCurrentField()
		if field != nil {
			field.Input, cmd = field.Input.Update(msg)

			// Sync manual tracking with textinput value
			if field.Input.Value() != field.ManualValue {
				field.ManualValue = field.Input.Value()
				m.editingText = field.ManualValue

				// Trigger real-time validation
				if m.realTimeValidation {
					if err := field.Validator(field.ManualValue); err != nil {
						m.errors[field.Key] = err.Error()
					} else {
						delete(m.errors, field.Key)
					}
				}
			}
		}
	}

	return m, cmd
}

// handleKeyInput processes enhanced keyboard input with Windows compatibility
func (m *EnhancedConfigFormModel) handleKeyInput(msg tea.KeyMsg) (EnhancedConfigFormModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c", "esc":
		if m.editing {
			m.stopEditing()
		}
		return *m, tea.Quit

	case "tab", "shift+tab":
		if msg.String() == "tab" {
			m.nextField()
		} else {
			m.prevField()
		}
		return *m, nil

	case "up", "k":
		if !m.editing {
			m.prevField()
		}
		return *m, nil

	case "down", "j":
		if !m.editing {
			m.nextField()
		}
		return *m, nil

	case "left", "h":
		if !m.editing {
			m.prevSection()
		}
		return *m, nil

	case "right", "l":
		if !m.editing {
			m.nextSection()
		}
		return *m, nil

	case "enter":
		if !m.editing {
			m.startEditing()
		} else {
			m.stopEditing()
			if m.autoSave {
				cmd = m.saveConfiguration()
			}
		}
		return *m, cmd

	case "ctrl+s":
		cmd = m.saveConfiguration()
		return *m, cmd

	case "ctrl+t":
		if !m.connectionTesting {
			cmd = m.testConnection()
		}
		return *m, cmd

	case "ctrl+r":
		cmd = m.resetConfiguration()
		return *m, cmd

	case "?":
		m.showHelp = !m.showHelp
		return *m, nil

	case "backspace":
		if m.editing && len(m.editingText) > 0 {
			m.editingText = m.editingText[:len(m.editingText)-1]
			if field := m.getCurrentField(); field != nil {
				field.ManualValue = m.editingText
				field.Input.SetValue(m.editingText)
			}
		}
		return *m, nil

	default:
		if m.editing {
			// Enhanced character input handling
			m.editingText += msg.String()
			if field := m.getCurrentField(); field != nil {
				field.ManualValue = m.editingText
				field.Input.SetValue(m.editingText)

				// Real-time validation
				if m.realTimeValidation {
					if err := field.Validator(m.editingText); err != nil {
						m.errors[field.Key] = err.Error()
					} else {
						delete(m.errors, field.Key)
					}
				}
			}
		}
		return *m, nil
	}
}

// Navigation methods
func (m *EnhancedConfigFormModel) nextSection() {
	if m.activeSection < len(m.sections)-1 {
		m.activeSection++
		m.activeField = 0
	}
}

func (m *EnhancedConfigFormModel) prevSection() {
	if m.activeSection > 0 {
		m.activeSection--
		m.activeField = 0
	}
}

func (m *EnhancedConfigFormModel) nextField() {
	if m.activeField < len(m.sections[m.activeSection].Fields)-1 {
		m.activeField++
	}
}

func (m *EnhancedConfigFormModel) prevField() {
	if m.activeField > 0 {
		m.activeField--
	}
}

// Editing methods with enhanced Windows compatibility
func (m *EnhancedConfigFormModel) startEditing() {
	m.editing = true
	m.pasteDetected = false
	m.inputResetCount = 0

	if field := m.getCurrentField(); field != nil {
		// Enhanced Windows workaround with multiple resets
		field.Input.Reset()
		field.Input.SetValue("")
		field.Input.Reset()
		field.Input.Blur()

		// Small delay to ensure clean state
		field.LastResetTime = time.Now()

		// Set initial values from manual tracking
		m.editingText = field.ManualValue
		field.Input.SetValue(field.ManualValue)
		field.Input.Focus()
	}
}

func (m *EnhancedConfigFormModel) stopEditing() {
	m.editing = false
	m.pasteDetected = false

	if field := m.getCurrentField(); field != nil {
		field.Value = m.editingText
		field.ManualValue = m.editingText
		field.Input.Blur()
	}

	m.editingText = ""
}

// Utility methods
func (m *EnhancedConfigFormModel) getCurrentField() *EnhancedConfigField {
	if m.activeSection >= 0 && m.activeSection < len(m.sections) &&
		m.activeField >= 0 && m.activeField < len(m.sections[m.activeSection].Fields) {
		return &m.sections[m.activeSection].Fields[m.activeField]
	}
	return nil
}

// Action methods
func (m *EnhancedConfigFormModel) saveConfiguration() tea.Cmd {
	// Extract configuration data from form
	config := m.extractEnhancedConfigurationData()

	// Update the configuration manager
	if err := m.configMgr.UpdateEnhanced(config); err != nil {
		m.lastOperationStatus = "Failed to update configuration"
		m.lastOperationSuccess = false
		m.lastOperationDetails = err.Error()
		return nil
	}

	// Save to file
	if err := m.configMgr.Save(); err != nil {
		m.lastOperationStatus = "Failed to save configuration"
		m.lastOperationSuccess = false
		m.lastOperationDetails = err.Error()
		return nil
	}

	m.lastOperationStatus = "Configuration saved successfully"
	m.lastOperationSuccess = true
	m.lastOperationDetails = ""

	return func() tea.Msg {
		return autoSaveCompleteMsg{}
	}
}

func (m *EnhancedConfigFormModel) testConnection() tea.Cmd {
	m.connectionTesting = true

	return func() tea.Msg {
		err := m.configMgr.TestConnection()
		return connectionTestCompleteMsg{
			success: err == nil,
			result:  m.getConnectionResultMessage(err),
		}
	}
}

func (m *EnhancedConfigFormModel) resetConfiguration() tea.Cmd {
	return func() tea.Msg {
		if err := m.configMgr.Reset(); err != nil {
			return configResetErrorMsg{err: err}
		}
		return configResetCompleteMsg{}
	}
}

// Helper methods
func (m *EnhancedConfigFormModel) extractEnhancedConfigurationData() *core.EnhancedConfig {
	config := m.config.Clone()

	// Extract values from all sections and fields
	for _, section := range m.sections {
		for _, field := range section.Fields {
			value := field.ManualValue
			if value == "" {
				value = field.Value
			}
			m.applyValueToConfig(config, field.Key, value)
		}
	}

	return config
}

func (m *EnhancedConfigFormModel) getConnectionResultMessage(err error) string {
	if err == nil {
		return "✓ Connection successful"
	}
	return fmt.Sprintf("✗ Connection failed: %s", err.Error())
}

// Message types for enhanced functionality
type connectionTestCompleteMsg struct {
	success bool
	result  string
}

type autoSaveCompleteMsg struct{}

type configResetCompleteMsg struct{}

type configResetErrorMsg struct {
	err error
}

// Styling
var (
	enhancedSectionTitleStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")).
					Bold(true).
					Padding(0, 1)

	enhancedSectionDescStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("8")).
					Italic(true).
					Padding(0, 2)

	enhancedFieldLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")).
				Bold(true)

	enhancedFieldActiveStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("12")).
					Padding(0, 1)

	enhancedFieldInactiveStyle = lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("8")).
					Padding(0, 1)

	enhancedErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true)

	enhancedSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	enhancedHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Italic(true)
)
