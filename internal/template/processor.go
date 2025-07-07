package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
		templatesDir: ".",
	}
}

func (p *Processor) LoadTemplate(templateName string) (string, error) {
	templatePath := filepath.Join(p.templatesDir, templateName)
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templateName, err)
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
	templatePath := filepath.Join(p.templatesDir, templateName)
	
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template file %s does not exist", templateName)
	}
	
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateName, err)
	}
	
	templateStr := string(content)
	
	requiredPlaceholders := []string{"{TASK}", "{RULES}", "{FILE_STRUCTURE}"}
	for _, placeholder := range requiredPlaceholders {
		if !strings.Contains(templateStr, placeholder) {
			return fmt.Errorf("template %s is missing required placeholder: %s", templateName, placeholder)
		}
	}
	
	return nil
}