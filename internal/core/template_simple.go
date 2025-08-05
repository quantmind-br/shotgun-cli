package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SimpleTemplateProcessor handles template loading and processing with string replacement
type SimpleTemplateProcessor struct {
	templates     map[string]string       // Raw template content
	templateInfos map[string]TemplateInfo // Template metadata
	mu            sync.RWMutex
}

// NewSimpleTemplateProcessor creates a new simple template processor
func NewSimpleTemplateProcessor() *SimpleTemplateProcessor {
	return &SimpleTemplateProcessor{
		templates:     make(map[string]string),
		templateInfos: make(map[string]TemplateInfo),
	}
}

// LoadTemplates loads both builtin and custom templates
func (stp *SimpleTemplateProcessor) LoadTemplates(builtinTemplatesDir, customTemplatesDir string) error {
	stp.mu.Lock()
	defer stp.mu.Unlock()

	// Load builtin templates first
	if err := stp.loadBuiltinTemplates(builtinTemplatesDir); err != nil {
		return fmt.Errorf("failed to load builtin templates: %w", err)
	}

	// Load custom templates (non-breaking - continue if this fails)
	if err := stp.loadCustomTemplates(customTemplatesDir); err != nil {
		// Log the error but don't fail the entire operation
		fmt.Printf("Warning: Failed to load custom templates: %v\n", err)
	}

	return nil
}

// loadBuiltinTemplates loads the standard builtin templates
func (stp *SimpleTemplateProcessor) loadBuiltinTemplates(templatesDir string) error {
	for _, tmplInfo := range AvailableTemplates {
		templatePath := filepath.Join(templatesDir, tmplInfo.Filename)

		// Read template content from filesystem
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return &ShotgunError{
				Operation: "read template file",
				Path:      templatePath,
				Err:       err,
			}
		}

		// Store template content and metadata
		tmplInfo.FilePath = templatePath // Set the full path
		stp.templates[tmplInfo.Key] = string(content)
		stp.templateInfos[tmplInfo.Key] = tmplInfo
	}

	return nil
}

// loadCustomTemplates loads custom templates from user directory
func (stp *SimpleTemplateProcessor) loadCustomTemplates(customTemplatesDir string) error {
	if customTemplatesDir == "" {
		return nil // No custom templates directory specified
	}

	// Load custom templates using the new parsing logic
	customTemplates, customContent, err := loadCustomTemplatesFromDirectory(customTemplatesDir)
	if err != nil {
		return err
	}

	// Validate against conflicts with builtin templates
	var builtinTemplates []TemplateInfo
	for _, tmpl := range AvailableTemplates {
		builtinTemplates = append(builtinTemplates, tmpl)
	}
	validCustomTemplates := validateTemplateKeyConflicts(builtinTemplates, customTemplates)

	// Add valid custom templates
	for _, tmpl := range validCustomTemplates {
		content, exists := customContent[tmpl.Key]
		if !exists {
			continue // Should not happen, but be safe
		}

		stp.templates[tmpl.Key] = content
		stp.templateInfos[tmpl.Key] = tmpl
	}

	return nil
}

// GeneratePrompt generates a prompt using the specified template and data
func (stp *SimpleTemplateProcessor) GeneratePrompt(templateKey string, data TemplateData) (string, error) {
	stp.mu.RLock()
	templateContent, exists := stp.templates[templateKey]
	stp.mu.RUnlock()

	if !exists {
		return "", &ShotgunError{
			Operation: "find template",
			Path:      templateKey,
			Err:       fmt.Errorf("template not found"),
		}
	}

	// Set current date if not provided
	if data.CurrentDate == "" {
		data.CurrentDate = time.Now().Format("2006-01-02")
	}

	// Simple string replacement
	result := templateContent

	// Apply replacements
	result = strings.ReplaceAll(result, "{TASK}", data.Task)
	result = strings.ReplaceAll(result, "{RULES}", data.Rules)
	result = strings.ReplaceAll(result, "{FILE_STRUCTURE}", data.FileStructure)
	result = strings.ReplaceAll(result, "{CURRENT_DATE}", data.CurrentDate)

	return result, nil
}

// GetTemplateContent returns the raw content of a template
func (stp *SimpleTemplateProcessor) GetTemplateContent(templateKey string) (string, error) {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	content, exists := stp.templates[templateKey]
	if !exists {
		return "", &ShotgunError{
			Operation: "find template",
			Path:      templateKey,
			Err:       fmt.Errorf("template not found"),
		}
	}

	return content, nil
}

// ListAvailableTemplates returns a list of available template keys
func (stp *SimpleTemplateProcessor) ListAvailableTemplates() []string {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	keys := make([]string, 0, len(stp.templates))
	for key := range stp.templates {
		keys = append(keys, key)
	}
	return keys
}

// HasTemplate checks if a template exists
func (stp *SimpleTemplateProcessor) HasTemplate(templateKey string) bool {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	_, exists := stp.templates[templateKey]
	return exists
}

// GetTemplateInfo returns template metadata for a specific key
func (stp *SimpleTemplateProcessor) GetTemplateInfo(templateKey string) (TemplateInfo, bool) {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	info, exists := stp.templateInfos[templateKey]
	return info, exists
}

// GetAllTemplateInfos returns all loaded template metadata
func (stp *SimpleTemplateProcessor) GetAllTemplateInfos() []TemplateInfo {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	infos := make([]TemplateInfo, 0, len(stp.templateInfos))
	for _, info := range stp.templateInfos {
		infos = append(infos, info)
	}
	return infos
}

// GetBuiltinTemplateInfos returns only builtin template metadata in the order defined by AvailableTemplates
func (stp *SimpleTemplateProcessor) GetBuiltinTemplateInfos() []TemplateInfo {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	var builtinInfos []TemplateInfo
	// Iterate in the order defined by AvailableTemplates to ensure consistent ordering
	for _, availableTemplate := range AvailableTemplates {
		if info, exists := stp.templateInfos[availableTemplate.Key]; exists && info.Source == TemplateSourceBuiltin {
			builtinInfos = append(builtinInfos, info)
		}
	}
	return builtinInfos
}

// GetCustomTemplateInfos returns only custom template metadata sorted by key for consistent ordering
func (stp *SimpleTemplateProcessor) GetCustomTemplateInfos() []TemplateInfo {
	stp.mu.RLock()
	defer stp.mu.RUnlock()

	var customInfos []TemplateInfo
	var keys []string

	// First collect all custom template keys
	for key, info := range stp.templateInfos {
		if info.Source == TemplateSourceCustom {
			keys = append(keys, key)
		}
	}

	// Sort keys to ensure consistent ordering
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// Add templates in sorted key order
	for _, key := range keys {
		customInfos = append(customInfos, stp.templateInfos[key])
	}

	return customInfos
}
