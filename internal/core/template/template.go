package template

import (
	"fmt"
	"regexp"
	"strings"
)

// Template represents a template with its metadata and content
type Template struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Content      string   `json:"content"`
	RequiredVars []string `json:"required_vars"`
	FilePath     string   `json:"file_path"`
	IsEmbedded   bool     `json:"is_embedded"` // true if from embedded filesystem
	Source       string   `json:"source"`      // "embedded", "user", or custom path
}

// Common template variable constants
const (
	VarTask          = "TASK"
	VarRules         = "RULES"
	VarFileStructure = "FILE_STRUCTURE"
	VarCurrentDate   = "CURRENT_DATE"
)

// Variable pattern for extracting template variables
var variablePattern = regexp.MustCompile(`\{([A-Z_][A-Z0-9_]*)\}`)

// parseTemplate parses template content and extracts metadata
func parseTemplate(content, fileName, filePath string) (*Template, error) {
	if content == "" {
		return nil, fmt.Errorf("template content is empty")
	}

	template := &Template{
		Name:       extractTemplateName(fileName),
		Content:    content,
		FilePath:   filePath,
		IsEmbedded: true,
	}

	// Extract description from the first comment or header
	template.Description = extractDescription(content, fileName)

	// Extract required variables from the template content
	requiredVars, err := extractRequiredVars(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract required variables: %w", err)
	}
	template.RequiredVars = requiredVars

	return template, nil
}

// extractTemplateName extracts a clean template name from the filename
func extractTemplateName(fileName string) string {
	name := strings.TrimSuffix(fileName, ".md")

	// Remove "prompt_" prefix if present
	name = strings.TrimPrefix(name, "prompt_")

	return name
}

// extractDescription extracts the description from template content
func extractDescription(content, fileName string) string {
	lines := strings.Split(content, "\n")

	// Look for markdown header or comment at the beginning
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for markdown header
		if strings.HasPrefix(line, "#") {
			description := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if description != "" {
				return description
			}
		}

		// Check for HTML comment
		if strings.HasPrefix(line, "<!--") && strings.HasSuffix(line, "-->") {
			description := strings.TrimSpace(line[4 : len(line)-3])
			if description != "" {
				return description
			}
		}

		// If we encounter non-empty, non-comment content, stop looking
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "<!--") {
			break
		}
	}

	// Default description based on file name
	return fmt.Sprintf("Template for %s", extractTemplateName(fileName))
}

// extractRequiredVars extracts all variable placeholders from template content
//
//nolint:unparam // error return reserved for future validation logic
func extractRequiredVars(content string) ([]string, error) {
	matches := variablePattern.FindAllStringSubmatch(content, -1)

	varSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			varSet[match[1]] = true
		}
	}

	vars := make([]string, 0, len(varSet))
	for variable := range varSet {
		vars = append(vars, variable)
	}

	return vars, nil
}

// validateTemplateContent validates the template content for proper formatting
func validateTemplateContent(content string) error {
	if content == "" {
		return fmt.Errorf("template content is empty")
	}

	// Check for malformed variable syntax
	lines := strings.Split(content, "\n")
	inCodeBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track if we're inside a code block (handles both ``` and ```language)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock

			continue
		}

		// Skip validation for lines inside code blocks
		if inCodeBlock {
			continue
		}

		// Check for unmatched braces
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		if openBraces != closeBraces {
			return fmt.Errorf("unmatched braces on line %d: %s", i+1, line)
		}

		// Verify only that occurrences matching variablePattern are well-formed
		if strings.Contains(line, "{") {
			// Find all potential variable patterns
			matches := variablePattern.FindAllString(line, -1)

			// Validate that each match is well-formed (this is redundant but explicit)
			for _, match := range matches {
				if !variablePattern.MatchString(match) {
					return fmt.Errorf("malformed variable pattern on line %d: %s", i+1, match)
				}
			}
		}
	}

	return nil
}

// GetVariableNames returns all variable names found in the template content.
func (t *Template) GetVariableNames() []string {
	vars, _ := extractRequiredVars(t.Content)

	return vars
}

// HasVariable checks if the template contains a specific variable
func (t *Template) HasVariable(varName string) bool {
	return strings.Contains(t.Content, "{"+varName+"}")
}

// GetVariableCount returns the number of times a variable appears in the template
func (t *Template) GetVariableCount(varName string) int {
	return strings.Count(t.Content, "{"+varName+"}")
}

// IsValid checks if the template is valid
func (t *Template) IsValid() error {
	if t.Name == "" {
		return fmt.Errorf("template name is empty")
	}

	if t.Content == "" {
		return fmt.Errorf("template content is empty")
	}

	return validateTemplateContent(t.Content)
}
