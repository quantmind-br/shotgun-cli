package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	sourceEmbedded = "embedded"
)

// TemplateSource defines the interface for loading templates from different sources
type TemplateSource interface {
	// LoadTemplates loads templates from this source and returns a map of template name -> Template
	LoadTemplates() (map[string]*Template, error)

	// GetSourceName returns a human-readable name for this source
	GetSourceName() string
}

// EmbeddedSource loads templates from an embedded filesystem
type EmbeddedSource struct {
	fsys fs.FS
}

// NewEmbeddedSource creates a new embedded template source
func NewEmbeddedSource(fsys fs.FS) *EmbeddedSource {
	return &EmbeddedSource{fsys: fsys}
}

// LoadTemplates loads all templates from the embedded filesystem
func (s *EmbeddedSource) LoadTemplates() (map[string]*Template, error) {
	return loadTemplatesFromFS(s.fsys, ".", true, sourceEmbedded)
}

// GetSourceName returns the source name
func (s *EmbeddedSource) GetSourceName() string {
	return sourceEmbedded
}

// FilesystemSource loads templates from a filesystem directory
type FilesystemSource struct {
	path       string
	sourceName string
}

// NewFilesystemSource creates a new filesystem template source
// sourceName is used for display purposes (e.g., "user", "custom")
func NewFilesystemSource(path string, sourceName string) *FilesystemSource {
	return &FilesystemSource{
		path:       path,
		sourceName: sourceName,
	}
}

// LoadTemplates loads all templates from the filesystem directory
func (s *FilesystemSource) LoadTemplates() (map[string]*Template, error) {
	fsys := os.DirFS(s.path)
	return loadTemplatesFromFS(fsys, ".", false, s.sourceName)
}

// GetSourceName returns the source name
func (s *FilesystemSource) GetSourceName() string {
	return s.sourceName
}

// loadTemplatesFromFS is a helper function that loads templates from any fs.FS
// basePath is the subdirectory within the filesystem to scan (usually ".")
// isEmbedded indicates whether templates should be marked as embedded
// sourceName is the display name for the source (e.g., "embedded", "user", "/tmp/templates")
func loadTemplatesFromFS(fsys fs.FS, basePath string, isEmbedded bool, sourceName string) (map[string]*Template, error) {
	templates := make(map[string]*Template)

	entries, err := fs.ReadDir(fsys, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		// Skip directories and non-markdown files
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Read template file content
		filePath := entry.Name()
		if basePath != "." {
			filePath = filepath.Join(basePath, entry.Name())
		}

		content, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			// Log warning but continue loading other templates
			continue
		}

		// Parse the template
		template, err := parseTemplate(string(content), entry.Name(), filePath)
		if err != nil {
			// Log warning but continue loading other templates
			continue
		}

		// Set source metadata
		template.IsEmbedded = isEmbedded
		template.Source = sourceName

		// Extract template name (remove "prompt_" prefix and .md suffix)
		templateName := extractTemplateName(entry.Name())

		templates[templateName] = template
	}

	return templates, nil
}
