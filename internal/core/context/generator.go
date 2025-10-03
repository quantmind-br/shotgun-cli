package context

import (
	"fmt"
	"strings"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

const (
	DefaultMaxSize  = 10 * 1024 * 1024 // 10MB
	DefaultMaxFiles = 1000
)

// GenProgress represents structured progress information
type GenProgress struct {
	Stage   string `json:"stage"`
	Message string `json:"message"`
}

type ContextGenerator interface {
	Generate(root *scanner.FileNode, config GenerateConfig) (string, error)
	GenerateWithProgress(root *scanner.FileNode, config GenerateConfig, progress func(string)) (string, error)
	GenerateWithProgressEx(root *scanner.FileNode, config GenerateConfig, progress func(GenProgress)) (string, error)
}

type GenerateConfig struct {
	MaxSize      int64             `json:"maxSize"`      // Deprecated: use MaxTotalSize for cumulative and MaxFileSize for per-file limits
	MaxFileSize  int64             `json:"maxFileSize"`  // Maximum size for individual files
	MaxTotalSize int64             `json:"maxTotalSize"` // Maximum total size of all content
	MaxFiles     int               `json:"maxFiles"`
	SkipBinary   bool              `json:"skipBinary"`
	TemplateVars map[string]string `json:"templateVars"`
	Template     string            `json:"template,omitempty"`
}

type ContextData struct {
	Task          string         `json:"task"`
	Rules         string         `json:"rules"`
	FileStructure string         `json:"fileStructure"`
	Files         []FileContent  `json:"files"`
	CurrentDate   string         `json:"currentDate"`
	Config        GenerateConfig `json:"config"`
}

type DefaultContextGenerator struct {
	treeRenderer     *TreeRenderer
	templateRenderer *TemplateRenderer
}

func NewDefaultContextGenerator() *DefaultContextGenerator {
	return &DefaultContextGenerator{
		treeRenderer:     NewTreeRenderer(),
		templateRenderer: NewTemplateRenderer(),
	}
}

func (g *DefaultContextGenerator) Generate(root *scanner.FileNode, config GenerateConfig) (string, error) {
	return g.GenerateWithProgress(root, config, nil)
}

func (g *DefaultContextGenerator) GenerateWithProgress(root *scanner.FileNode, config GenerateConfig, progress func(string)) (string, error) {
	// Use the new structured progress internally, adapting to the old interface
	var adaptedProgress func(GenProgress)
	if progress != nil {
		adaptedProgress = func(p GenProgress) {
			progress(p.Message)
		}
	}
	return g.GenerateWithProgressEx(root, config, adaptedProgress)
}

func (g *DefaultContextGenerator) GenerateWithProgressEx(root *scanner.FileNode, config GenerateConfig, progress func(GenProgress)) (string, error) {
	if err := g.validateConfig(&config); err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}

	if progress != nil {
		progress(GenProgress{Stage: "tree_generation", Message: "Generating file structure..."})
	}

	fileStructure, err := g.treeRenderer.RenderTree(root)
	if err != nil {
		return "", fmt.Errorf("failed to render tree: %w", err)
	}

	if progress != nil {
		progress(GenProgress{Stage: "content_collection", Message: "Collecting file contents..."})
	}

	files, err := g.collectFileContents(root, config)
	if err != nil {
		return "", fmt.Errorf("failed to collect file contents: %w", err)
	}

	// Combine tree structure with file content blocks
	fileStructureComplete := g.buildCompleteFileStructure(fileStructure, files)

	if progress != nil {
		progress(GenProgress{Stage: "template_rendering", Message: "Rendering template..."})
	}

	contextData := ContextData{
		Task:          config.TemplateVars["TASK"],
		Rules:         config.TemplateVars["RULES"],
		FileStructure: fileStructureComplete,
		Files:         files,
		CurrentDate:   time.Now().Format("2006-01-02 15:04:05"),
		Config:        config,
	}

	template := config.Template
	if template == "" {
		template = g.templateRenderer.getDefaultTemplate()
	}

	// Convert {VARIABLE} syntax to {{.Variable}} syntax for Go templates
	template = convertTemplateVariables(template)

	result, err := g.templateRenderer.RenderTemplate(template, contextData)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	if int64(len(result)) > config.MaxTotalSize {
		return "", fmt.Errorf("generated context exceeds total size limit: %d bytes > %d bytes", len(result), config.MaxTotalSize)
	}

	if progress != nil {
		progress(GenProgress{Stage: "complete", Message: "Context generation completed"})
	}

	return result, nil
}

func (g *DefaultContextGenerator) validateConfig(config *GenerateConfig) error {
	// Handle backward compatibility with MaxSize
	if config.MaxSize > 0 {
		// If MaxSize is set but new fields are not, use MaxSize for both
		if config.MaxFileSize == 0 {
			config.MaxFileSize = config.MaxSize
		}
		if config.MaxTotalSize == 0 {
			config.MaxTotalSize = config.MaxSize
		}
	} else {
		// Set defaults if neither old nor new fields are set
		if config.MaxFileSize == 0 {
			config.MaxFileSize = DefaultMaxSize
		}
		if config.MaxTotalSize == 0 {
			config.MaxTotalSize = DefaultMaxSize
		}
		// Keep MaxSize for backward compatibility in templates
		config.MaxSize = config.MaxTotalSize
	}

	if config.MaxFiles <= 0 {
		config.MaxFiles = DefaultMaxFiles
	}
	if config.TemplateVars == nil {
		config.TemplateVars = make(map[string]string)
	}
	return nil
}

func (g *DefaultContextGenerator) collectFileContents(root *scanner.FileNode, config GenerateConfig) ([]FileContent, error) {
	return collectFileContents(root, config)
}

// buildCompleteFileStructure combines ASCII tree with file content blocks
func (g *DefaultContextGenerator) buildCompleteFileStructure(tree string, files []FileContent) string {
	var builder strings.Builder

	// First part: ASCII tree structure
	builder.WriteString(tree)

	// Add separator if there are files
	if len(files) > 0 {
		builder.WriteString("\n")

		// Second part: File content blocks in XML-like format
		builder.WriteString(renderFileContentBlocks(files))
	}

	return builder.String()
}

// convertTemplateVariables converts {VARIABLE} syntax to {{.Variable}} syntax for Go templates
func convertTemplateVariables(template string) string {
	// Map of variable conversions from {UPPERCASE} to {{.TitleCase}}
	conversions := map[string]string{
		"{TASK}":           "{{.Task}}",
		"{RULES}":          "{{.Rules}}",
		"{FILE_STRUCTURE}": "{{.FileStructure}}",
		"{CURRENT_DATE}":   "{{.CurrentDate}}",
	}

	result := template
	for old, new := range conversions {
		result = strings.Replace(result, old, new, -1)
	}

	return result
}
