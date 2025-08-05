package core

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontmatterParseResult represents the result of parsing YAML frontmatter
type FrontmatterParseResult struct {
	Metadata CustomTemplateMetadata
	Content  string
}

// parseFrontmatter parses YAML frontmatter from a Markdown file
func parseFrontmatter(content string) (*FrontmatterParseResult, error) {
	lines := strings.Split(content, "\n")

	// Check if file starts with YAML frontmatter delimiter
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("no YAML frontmatter found - file must start with '---'")
	}

	// Find the closing delimiter
	var yamlEndIndex int = -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			yamlEndIndex = i
			break
		}
	}

	if yamlEndIndex == -1 {
		return nil, fmt.Errorf("YAML frontmatter not properly closed - missing closing '---'")
	}

	// Extract YAML content (between the delimiters)
	yamlContent := strings.Join(lines[1:yamlEndIndex], "\n")

	// Extract markdown content (after the closing delimiter)
	var markdownContent string
	if yamlEndIndex+1 < len(lines) {
		markdownContent = strings.Join(lines[yamlEndIndex+1:], "\n")
	}

	// Parse YAML frontmatter
	var metadata CustomTemplateMetadata
	if err := yaml.Unmarshal([]byte(yamlContent), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Validate required fields
	if err := ValidateCustomTemplateMetadata(metadata); err != nil {
		return nil, fmt.Errorf("invalid template metadata: %w", err)
	}

	return &FrontmatterParseResult{
		Metadata: metadata,
		Content:  strings.TrimSpace(markdownContent),
	}, nil
}

// loadCustomTemplate loads and parses a single custom template file
func loadCustomTemplate(filePath string) (*TemplateInfo, string, error) {
	// Read the template file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read template file %s: %w", filePath, err)
	}

	// Parse frontmatter
	result, err := parseFrontmatter(string(content))
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse template %s: %w", filePath, err)
	}

	// Create TemplateInfo from parsed metadata
	templateInfo := &TemplateInfo{
		Key:         result.Metadata.Key,
		Name:        result.Metadata.Name,
		Description: result.Metadata.Description,
		Filename:    filepath.Base(filePath),
		Source:      TemplateSourceCustom,
		FilePath:    filePath,
	}

	return templateInfo, result.Content, nil
}

// loadCustomTemplatesFromDirectory scans a directory for custom template files
func loadCustomTemplatesFromDirectory(templatesDir string) ([]TemplateInfo, map[string]string, error) {
	var templates []TemplateInfo
	templateContent := make(map[string]string)

	// Check if directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// Directory doesn't exist, return empty results (not an error)
		log.Printf("Custom templates directory does not exist: %s", templatesDir)
		return templates, templateContent, nil
	}

	// Walk through the directory looking for .md files
	err := filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Warning: Error accessing %s: %v", path, err)
			return nil // Continue walking, don't fail the entire process
		}

		// Skip directories and non-.md files
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Try to load the custom template
		templateInfo, content, err := loadCustomTemplate(path)
		if err != nil {
			log.Printf("Warning: Failed to load custom template %s: %v", path, err)
			return nil // Continue processing other templates
		}

		// Check for key conflicts with existing templates
		for _, existing := range templates {
			if existing.Key == templateInfo.Key {
				log.Printf("Warning: Duplicate template key '%s' in %s, skipping (first occurrence takes precedence)", templateInfo.Key, path)
				return nil
			}
		}

		// Add the template
		templates = append(templates, *templateInfo)
		templateContent[templateInfo.Key] = content
		log.Printf("Loaded custom template: %s (%s)", templateInfo.Name, templateInfo.Key)

		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan custom templates directory %s: %w", templatesDir, err)
	}

	log.Printf("Successfully loaded %d custom templates from %s", len(templates), templatesDir)
	return templates, templateContent, nil
}

// validateTemplateKeyConflicts checks for conflicts between builtin and custom templates
func validateTemplateKeyConflicts(builtinTemplates []TemplateInfo, customTemplates []TemplateInfo) []TemplateInfo {
	// Create a map of builtin template keys for fast lookup
	builtinKeys := make(map[string]bool)
	for _, tmpl := range builtinTemplates {
		builtinKeys[tmpl.Key] = true
	}

	// Filter out custom templates that conflict with builtin templates
	var validCustomTemplates []TemplateInfo
	for _, customTmpl := range customTemplates {
		if builtinKeys[customTmpl.Key] {
			log.Printf("Warning: Custom template with key '%s' conflicts with builtin template, ignoring custom template", customTmpl.Key)
			continue
		}
		validCustomTemplates = append(validCustomTemplates, customTmpl)
	}

	return validCustomTemplates
}

// GetTemplateInfoByKey finds a template by key in a slice of TemplateInfo
func GetTemplateInfoByKey(templates []TemplateInfo, key string) (TemplateInfo, bool) {
	for _, tmpl := range templates {
		if tmpl.Key == key {
			return tmpl, true
		}
	}
	return TemplateInfo{}, false
}
