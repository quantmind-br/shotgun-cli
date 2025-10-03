package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/quantmind-br/shotgun-cli/internal/assets"
	"github.com/spf13/viper"
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
	manager := &Manager{
		templates: make(map[string]*Template),
		renderer:  NewRenderer(),
	}

	// Create template sources in priority order (first = lowest priority, last = highest priority)
	// Later sources override earlier ones
	sources := []TemplateSource{}

	// 1. Embedded templates (lowest priority)
	var err error
	templatesFS, err = fs.Sub(assets.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("failed to create templates filesystem: %w", err)
	}
	sources = append(sources, NewEmbeddedSource(templatesFS))

	// 2. User config directory templates (XDG compliant)
	userTemplatesDir := filepath.Join(xdg.ConfigHome, "shotgun-cli", "templates")
	if err := os.MkdirAll(userTemplatesDir, 0755); err == nil {
		// Only add if directory creation succeeded or already exists
		sources = append(sources, NewFilesystemSource(userTemplatesDir, "user"))
	}

	// 3. Custom path from config (highest priority)
	customPath := viper.GetString("template.custom-path")
	if customPath != "" {
		// Expand home directory if needed
		if strings.HasPrefix(customPath, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				customPath = filepath.Join(home, customPath[2:])
			}
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(customPath, 0755); err == nil {
			// Use basename of custom path as source name for display
			sourceName := filepath.Base(customPath)
			sources = append(sources, NewFilesystemSource(customPath, sourceName))
		}
	}

	// Load templates from all sources
	if err := manager.loadFromSources(sources); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return manager, nil
}

// loadFromSources loads templates from multiple sources with priority
// Later sources in the slice override earlier ones (by template name)
func (m *Manager) loadFromSources(sources []TemplateSource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load templates from each source in order
	// Later sources override earlier ones with the same name
	for _, source := range sources {
		templates, err := source.LoadTemplates()
		if err != nil {
			// Log warning but continue with other sources
			// This allows the application to start even if one source fails
			continue
		}

		// Merge templates into manager's map
		// Later sources override earlier ones
		for name, template := range templates {
			m.templates[name] = template
		}
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
