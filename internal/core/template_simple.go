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
	templates map[string]string // Raw template content
	mu        sync.RWMutex
}

// NewSimpleTemplateProcessor creates a new simple template processor
func NewSimpleTemplateProcessor() *SimpleTemplateProcessor {
	return &SimpleTemplateProcessor{
		templates: make(map[string]string),
	}
}

// LoadTemplatesFromDirectory loads all templates from the specified directory
func (stp *SimpleTemplateProcessor) LoadTemplatesFromDirectory(templatesDir string) error {
	stp.mu.Lock()
	defer stp.mu.Unlock()

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

		stp.templates[tmplInfo.Key] = string(content)
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
