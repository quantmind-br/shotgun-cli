package template

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed templates/*.md
var embeddedTemplates embed.FS

type Processor struct {
	templatesDir string
}

type PromptData struct {
	Task          string
	Rules         string
	FileStructure string
}

func NewProcessor() *Processor {
	return &Processor{
		templatesDir: "templates",
	}
}

func (p *Processor) LoadTemplate(templateName string) (string, error) {
	// Try to load from local directory first (for development/customization)
	templatePath := filepath.Join(p.templatesDir, templateName)
	if content, err := os.ReadFile(templatePath); err == nil {
		return string(content), nil
	}
	
	// Fallback to embedded templates
	embeddedPath := filepath.Join("templates", templateName)
	content, err := embeddedTemplates.ReadFile(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s (tried local and embedded): %w", templateName, err)
	}
	
	return string(content), nil
}

func (p *Processor) ProcessTemplate(templateContent string, data PromptData) string {
	processed := templateContent
	
	processed = strings.ReplaceAll(processed, "{TASK}", data.Task)
	processed = strings.ReplaceAll(processed, "{RULES}", data.Rules)
	processed = strings.ReplaceAll(processed, "{FILE_STRUCTURE}", data.FileStructure)
	
	return processed
}

func (p *Processor) SavePrompt(content string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("shotgun_prompt_%s.md", timestamp)
	
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to save prompt to %s: %w", filename, err)
	}
	
	return filename, nil
}

func (p *Processor) ValidateTemplate(templateName string) error {
	// Try to load template using the same logic as LoadTemplate
	content, err := p.LoadTemplate(templateName)
	if err != nil {
		return fmt.Errorf("template file %s does not exist or cannot be read: %w", templateName, err)
	}
	
	// Validate required placeholders
	requiredPlaceholders := []string{"{TASK}", "{RULES}", "{FILE_STRUCTURE}"}
	for _, placeholder := range requiredPlaceholders {
		if !strings.Contains(content, placeholder) {
			return fmt.Errorf("template %s is missing required placeholder: %s", templateName, placeholder)
		}
	}
	
	return nil
}