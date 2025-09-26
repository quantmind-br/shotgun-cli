package context

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

type TemplateRenderer struct {
	funcs template.FuncMap
	requiredVars []string
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		funcs: getTemplateFunctions(),
		requiredVars: []string{"TASK"}, // Default template requires TASK variable
	}
}

func (tr *TemplateRenderer) RenderTemplate(templateContent string, data ContextData) (string, error) {
	// Validate required variables for default template
	if templateContent == tr.getDefaultTemplate() {
		if err := tr.validateRequiredVars(data); err != nil {
			return "", fmt.Errorf("template variable validation failed: %w", err)
		}
	}

	tmpl, err := template.New("context").Funcs(tr.funcs).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func getTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			if length <= 3 {
				return s[:length]
			}
			return s[:length-3] + "..."
		},
		"formatSize": formatFileSize,
		"detectLang": detectLanguage,
		"now": func() string {
			return time.Now().Format("2006-01-02 15:04:05")
		},
		"join": strings.Join,
		"title": strings.Title,
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
	}
}

func (tr *TemplateRenderer) validateRequiredVars(data ContextData) error {
	for _, varName := range tr.requiredVars {
		value, exists := data.Config.TemplateVars[varName]
		if !exists || strings.TrimSpace(value) == "" {
			return fmt.Errorf("required template variable '%s' is missing or empty", varName)
		}
	}
	return nil
}

func (tr *TemplateRenderer) getDefaultTemplate() string {
	return `# Project Context

**Generated:** {{now}}
{{if .Task}}**Task:** {{.Task}}{{end}}
{{if .Rules}}**Rules:** {{.Rules}}{{end}}

## File Structure

{{.FileStructure}}

## File Contents
{{range .Files}}
### {{.RelPath}}{{if .Language}} ({{.Language}}){{end}}

` + "```" + `{{if .Language}}{{.Language}}{{end}}
{{.Content}}
` + "```" + `
{{end}}

---
*Context generated with {{formatSize .Config.MaxSize}} size limit*`
}