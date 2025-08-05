package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Template definitions
const (
	TemplateDevKey     = "dev"
	TemplateArchKey    = "architect"
	TemplateBugKey     = "debug"
	TemplateProjectKey = "project"
)

// TemplateSource represents the source of a template
type TemplateSource int

const (
	TemplateSourceBuiltin TemplateSource = iota
	TemplateSourceCustom
)

// String returns the string representation of TemplateSource
func (ts TemplateSource) String() string {
	switch ts {
	case TemplateSourceBuiltin:
		return "builtin"
	case TemplateSourceCustom:
		return "custom"
	default:
		return "unknown"
	}
}

// CustomTemplateMetadata represents the YAML frontmatter metadata
type CustomTemplateMetadata struct {
	Key         string `yaml:"key"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ValidateCustomTemplateMetadata validates that required fields are present
func ValidateCustomTemplateMetadata(metadata CustomTemplateMetadata) error {
	if strings.TrimSpace(metadata.Key) == "" {
		return fmt.Errorf("template key is required")
	}
	if strings.TrimSpace(metadata.Name) == "" {
		return fmt.Errorf("template name is required")
	}
	if strings.TrimSpace(metadata.Description) == "" {
		return fmt.Errorf("template description is required")
	}
	return nil
}

// Template metadata
type TemplateInfo struct {
	Key         string
	Name        string
	Description string
	Filename    string
	Source      TemplateSource
	FilePath    string // Full path to the template file
}

var AvailableTemplates = []TemplateInfo{
	{
		Key:         TemplateDevKey,
		Name:        "Dev (makeDiffGitFormat)",
		Description: "Generate git diffs for code changes",
		Filename:    "prompt_makeDiffGitFormat.md",
		Source:      TemplateSourceBuiltin,
		FilePath:    "", // Will be set during loading
	},
	{
		Key:         TemplateArchKey,
		Name:        "Architect (makePlan)",
		Description: "Create design plans and architecture",
		Filename:    "prompt_makePlan.md",
		Source:      TemplateSourceBuiltin,
		FilePath:    "", // Will be set during loading
	},
	{
		Key:         TemplateBugKey,
		Name:        "Debug (analyzeBug)",
		Description: "Bug analysis and debugging",
		Filename:    "prompt_analyzeBug.md",
		Source:      TemplateSourceBuiltin,
		FilePath:    "", // Will be set during loading
	},
	{
		Key:         TemplateProjectKey,
		Name:        "Project Manager",
		Description: "Documentation sync and task management",
		Filename:    "prompt_projectManager.md",
		Source:      TemplateSourceBuiltin,
		FilePath:    "", // Will be set during loading
	},
}

// templateContent stores raw template content for simple string replacement
type templateContent struct {
	key     string
	content string
}

// LoadTemplatesFromDirectory loads all templates from the specified directory
func (tp *TemplateProcessor) LoadTemplatesFromDirectory(templatesDir string) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Use a map to store raw content instead of parsed templates
	if tp.templates == nil {
		tp.templates = make(map[string]*template.Template)
	}

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

		// Store as simple template for string replacement
		tmpl := template.New(tmplInfo.Key)
		tmpl, err = tmpl.Parse(string(content))
		if err != nil {
			return &ShotgunError{
				Operation: "parse template",
				Path:      templatePath,
				Err:       err,
			}
		}

		tp.templates[tmplInfo.Key] = tmpl
	}

	return nil
}

// LoadTemplatesFromEmbedded loads templates from embedded content
func (tp *TemplateProcessor) LoadTemplatesFromEmbedded(templateContents map[string]string) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	for key, content := range templateContents {
		// Create a custom template with functions for variable substitution
		tmpl := template.New(key).Funcs(template.FuncMap{
			"currentDate": func() string {
				return time.Now().Format("2006-01-02")
			},
		})

		// Parse the template
		tmpl, err := tmpl.Parse(content)
		if err != nil {
			return &ShotgunError{
				Operation: "parse embedded template",
				Path:      key,
				Err:       err,
			}
		}

		tp.templates[key] = tmpl
	}

	return nil
}

// GeneratePrompt generates a prompt using the specified template and data
func (tp *TemplateProcessor) GeneratePrompt(templateKey string, data TemplateData) (string, error) {
	tp.mu.RLock()
	tmpl, exists := tp.templates[templateKey]
	tp.mu.RUnlock()

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

	// Use simple string replacement for now (faster than template execution)
	// This matches the existing JavaScript behavior
	tp.mu.RLock()
	templateContent := tmpl.Tree.Root.String() // Get the raw template content
	tp.mu.RUnlock()

	// Manual string replacement (matching existing app behavior)
	result := templateContent
	result = strings.ReplaceAll(result, "{TASK}", data.Task)
	result = strings.ReplaceAll(result, "{RULES}", data.Rules)
	result = strings.ReplaceAll(result, "{FILE_STRUCTURE}", data.FileStructure)
	result = strings.ReplaceAll(result, "{CURRENT_DATE}", data.CurrentDate)

	return result, nil
}

// GetTemplateContent returns the raw content of a template
func (tp *TemplateProcessor) GetTemplateContent(templateKey string) (string, error) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	tmpl, exists := tp.templates[templateKey]
	if !exists {
		return "", &ShotgunError{
			Operation: "find template",
			Path:      templateKey,
			Err:       fmt.Errorf("template not found"),
		}
	}

	// This is a simplified way to get the content - in a real implementation
	// you might want to store the original content separately
	return tmpl.Tree.Root.String(), nil
}

// ListAvailableTemplates returns a list of available template keys
func (tp *TemplateProcessor) ListAvailableTemplates() []string {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	keys := make([]string, 0, len(tp.templates))
	for key := range tp.templates {
		keys = append(keys, key)
	}
	return keys
}

// HasTemplate checks if a template exists
func (tp *TemplateProcessor) HasTemplate(templateKey string) bool {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	_, exists := tp.templates[templateKey]
	return exists
}

// GetTemplateInfo returns information about a template
func GetTemplateInfo(key string) (TemplateInfo, bool) {
	for _, tmpl := range AvailableTemplates {
		if tmpl.Key == key {
			return tmpl, true
		}
	}
	return TemplateInfo{}, false
}
