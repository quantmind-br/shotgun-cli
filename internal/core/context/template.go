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
}

func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{
		funcs: getTemplateFunctions(),
	}
}

func (tr *TemplateRenderer) RenderTemplate(templateContent string, data ContextData) (string, error) {
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