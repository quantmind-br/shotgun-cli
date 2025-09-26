package context

import (
	"fmt"
	"time"

	"github.com/shotgun-cli/internal/core/scanner"
)

const (
	DefaultMaxSize  = 10 * 1024 * 1024 // 10MB
	DefaultMaxFiles = 1000
)

type ContextGenerator interface {
	Generate(root *scanner.FileNode, config GenerateConfig) (string, error)
	GenerateWithProgress(root *scanner.FileNode, config GenerateConfig, progress func(string)) (string, error)
}

type GenerateConfig struct {
	MaxSize      int64             `json:"maxSize"`
	MaxFiles     int               `json:"maxFiles"`
	SkipBinary   bool              `json:"skipBinary"`
	TemplateVars map[string]string `json:"templateVars"`
	Template     string            `json:"template,omitempty"`
}

type ContextData struct {
	Task          string        `json:"task"`
	Rules         string        `json:"rules"`
	FileStructure string        `json:"fileStructure"`
	Files         []FileContent `json:"files"`
	CurrentDate   string        `json:"currentDate"`
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
	if err := g.validateConfig(&config); err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}

	if progress != nil {
		progress("Generating file structure...")
	}

	fileStructure, err := g.treeRenderer.RenderTree(root)
	if err != nil {
		return "", fmt.Errorf("failed to render tree: %w", err)
	}

	if progress != nil {
		progress("Collecting file contents...")
	}

	files, err := g.collectFileContents(root, config)
	if err != nil {
		return "", fmt.Errorf("failed to collect file contents: %w", err)
	}

	if progress != nil {
		progress("Rendering template...")
	}

	contextData := ContextData{
		Task:          config.TemplateVars["TASK"],
		Rules:         config.TemplateVars["RULES"],
		FileStructure: fileStructure,
		Files:         files,
		CurrentDate:   time.Now().Format("2006-01-02 15:04:05"),
		Config:        config,
	}

	template := config.Template
	if template == "" {
		template = g.templateRenderer.getDefaultTemplate()
	}

	result, err := g.templateRenderer.RenderTemplate(template, contextData)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	if int64(len(result)) > config.MaxSize {
		return "", fmt.Errorf("generated context exceeds size limit: %d bytes > %d bytes", len(result), config.MaxSize)
	}

	return result, nil
}

func (g *DefaultContextGenerator) validateConfig(config *GenerateConfig) error {
	if config.MaxSize <= 0 {
		config.MaxSize = DefaultMaxSize
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