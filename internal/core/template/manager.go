package template

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/quantmind-br/shotgun-cli/internal/assets"
)

// TemplateManager defines the interface for template management
type TemplateManager interface {
	ListTemplates() ([]Template, error)
	GetTemplate(name string) (*Template, error)
	RenderTemplate(name string, vars map[string]string) (string, error)
	ValidateTemplate(name string) error
	GetRequiredVariables(name string) ([]string, error)
}

// Manager implements the TemplateManager interface
type Manager struct {
	templates map[string]*Template
	mu        sync.RWMutex
	renderer  *Renderer
}


var templatesFS fs.FS

// NewManager creates a new template manager instance
func NewManager() (*Manager, error) {
	// Create fs.Sub for the templates directory
	var err error
	templatesFS, err = fs.Sub(assets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to create templates filesystem: %w", err)
	}

	manager := &Manager{
		templates: make(map[string]*Template),
		renderer:  NewRenderer(),
	}

	if err := manager.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return manager, nil
}

// loadTemplates loads all embedded templates from the filesystem
func (m *Manager) loadTemplates() error {
	entries, err := fs.ReadDir(templatesFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		content, err := fs.ReadFile(templatesFS, entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", entry.Name(), err)
		}

		template, err := parseTemplate(string(content), entry.Name(), entry.Name())
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", entry.Name(), err)
		}

		templateName := strings.TrimSuffix(entry.Name(), ".md")
		// Remove "prompt_" prefix if present
		if strings.HasPrefix(templateName, "prompt_") {
			templateName = strings.TrimPrefix(templateName, "prompt_")
		}

		m.templates[templateName] = template
	}

	return nil
}

// ListTemplates returns all available templates sorted by name
func (m *Manager) ListTemplates() ([]Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	templates := make([]Template, 0, len(m.templates))
	for _, template := range m.templates {
		templates = append(templates, *template)
	}

	// Sort templates by name for predictable ordering
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// GetTemplate retrieves a specific template by name
func (m *Manager) GetTemplate(name string) (*Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	template, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}

// RenderTemplate renders a template with the provided variables
func (m *Manager) RenderTemplate(name string, vars map[string]string) (string, error) {
	template, err := m.GetTemplate(name)
	if err != nil {
		return "", err
	}

	return m.renderer.RenderTemplate(template, vars)
}

// ValidateTemplate validates a template's content and structure
func (m *Manager) ValidateTemplate(name string) error {
	template, err := m.GetTemplate(name)
	if err != nil {
		return err
	}

	return validateTemplateContent(template.Content)
}

// GetTemplateNames returns a list of all available template names
func (m *Manager) GetTemplateNames() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.templates))
	for name := range m.templates {
		names = append(names, name)
	}

	return names
}

// HasTemplate checks if a template exists
func (m *Manager) HasTemplate(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.templates[name]
	return exists
}

// GetRequiredVariables returns the list of required variables for a template
func (m *Manager) GetRequiredVariables(name string) ([]string, error) {
	template, err := m.GetTemplate(name)
	if err != nil {
		return nil, err
	}

	return m.renderer.GetRequiredVariables(template), nil
}
