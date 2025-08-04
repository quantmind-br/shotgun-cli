package core

import (
	"path/filepath"
	"sync"
	"text/template"
	"time"

	gitignore "github.com/sabhiram/go-gitignore"
)

// FileNode represents a file or directory in the project tree
type FileNode struct {
	Name            string      `json:"name"`
	Path            string      `json:"path"`
	RelPath         string      `json:"relPath"`
	IsDir           bool        `json:"isDir"`
	Children        []*FileNode `json:"children,omitempty"`
	IsGitignored    bool        `json:"isGitignored"`
	IsCustomIgnored bool        `json:"isCustomIgnored"`
	IsExcluded      bool        `json:"isExcluded"` // User exclusion state
}

// DirectoryScanner handles recursive directory scanning with filtering
type DirectoryScanner struct {
	gitIgnore     *gitignore.GitIgnore
	customIgnore  *gitignore.GitIgnore
	defaultIgnore *gitignore.GitIgnore
	progressChan  chan ProgressUpdate
	mu            sync.RWMutex
}

// ProgressUpdate represents progress information during scanning
type ProgressUpdate struct {
	Current     int64
	Total       int64
	Percentage  float64
	CurrentFile string
	Phase       string
}

// ContextGenerator handles context generation with size limits
type ContextGenerator struct {
	maxSize      int64
	progressChan chan ProgressUpdate
	workerPool   chan struct{}
	mu           sync.Mutex
}

// TemplateProcessor handles template loading and processing
type TemplateProcessor struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// TemplateData contains data for template substitution
type TemplateData struct {
	Task          string
	Rules         string
	CurrentDate   string
	FileStructure string
}

// SelectionStatus represents the selection state of a path
type SelectionStatus int

const (
	StatusInherit  SelectionStatus = iota // Inherit from parent (default)
	StatusExcluded                        // Explicitly excluded
)

// SelectionState manages file inclusion/exclusion state
type SelectionState struct {
	selection map[string]SelectionStatus // Only explicit decisions (new hierarchical system)
	mu        sync.RWMutex               // Thread safety (preserve existing)
}

// ShotgunError provides structured error information
type ShotgunError struct {
	Operation string
	Path      string
	Err       error
}

func (e *ShotgunError) Error() string {
	return "shotgun-cli: " + e.Operation + " failed for " + e.Path + ": " + e.Err.Error()
}

func (e *ShotgunError) Unwrap() error {
	return e.Err
}

// ErrorCollector aggregates multiple errors
type ErrorCollector struct {
	errors []error
	mu     sync.Mutex
}

func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.mu.Lock()
		ec.errors = append(ec.errors, err)
		ec.mu.Unlock()
	}
}

func (ec *ErrorCollector) HasErrors() bool {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return len(ec.errors) > 0
}

func (ec *ErrorCollector) Errors() []error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	result := make([]error, len(ec.errors))
	copy(result, ec.errors)
	return result
}

// NewDirectoryScanner creates a new directory scanner
func NewDirectoryScanner() *DirectoryScanner {
	return &DirectoryScanner{
		progressChan: make(chan ProgressUpdate, 100),
	}
}

// NewContextGenerator creates a new context generator
func NewContextGenerator(maxSize int64) *ContextGenerator {
	return &ContextGenerator{
		maxSize:      0, // No limit
		progressChan: make(chan ProgressUpdate, 100),
		workerPool:   make(chan struct{}, 10),
	}
}

// NewTemplateProcessor creates a new template processor
func NewTemplateProcessor() *SimpleTemplateProcessor {
	return NewSimpleTemplateProcessor()
}

// NewSelectionState creates a new selection state
func NewSelectionState() *SelectionState {
	return &SelectionState{
		selection: make(map[string]SelectionStatus),
	}
}

// IsPathExcluded checks if a path is excluded, checking parent hierarchy
func (ss *SelectionState) IsPathExcluded(path string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// Clean path for cross-platform compatibility
	cleanPath := filepath.Clean(path)

	// Check each parent level for exclusions
	currentPath := cleanPath
	for {
		// Check explicit exclusion at this level
		if status, exists := ss.selection[currentPath]; exists {
			return status == StatusExcluded
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root without finding exclusion
			break
		}
		currentPath = parentPath
	}

	// Default: included (no exclusions found)
	return false
}

// IsFileIncluded checks if a file should be included in the final output
func (ss *SelectionState) IsFileIncluded(path string) bool {
	// Use the new hierarchical exclusion checking
	return !ss.IsPathExcluded(path)
}

// ToggleFile toggles the exclusion state of a file
func (ss *SelectionState) ToggleFile(path string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	cleanPath := filepath.Clean(path)

	// Simple state cycling: inherit -> excluded -> inherit
	current := ss.selection[cleanPath] // Zero value is StatusInherit
	if current == StatusInherit {
		ss.selection[cleanPath] = StatusExcluded
	} else {
		delete(ss.selection, cleanPath) // Return to inherit state
	}
}

// ExcludeFile explicitly excludes a file
func (ss *SelectionState) ExcludeFile(path string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.selection[filepath.Clean(path)] = StatusExcluded
}

// IncludeFile explicitly includes a file (removes from excluded)
func (ss *SelectionState) IncludeFile(path string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.selection, filepath.Clean(path))
}

// Reset clears all exclusions
func (ss *SelectionState) Reset() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.selection = make(map[string]SelectionStatus)
}

// GetExcludedFiles returns a copy of excluded files
func (ss *SelectionState) GetExcludedFiles() map[string]bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	result := make(map[string]bool)
	for k, v := range ss.selection {
		if v == StatusExcluded {
			result[k] = true
		}
	}
	return result
}

// Configuration types for OpenAI translation integration

// Config represents the application settings
type Config struct {
	// OpenAI API Settings
	OpenAI OpenAIConfig `json:"openai"`

	// Translation Settings
	Translation TranslationConfig `json:"translation"`

	// Application Settings
	App AppConfig `json:"app"`

	// Metadata
	Version     string    `json:"version"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// OpenAIConfig contains OpenAI API configuration
type OpenAIConfig struct {
	// API Key (stored in keyring, this holds reference/alias)
	APIKeyAlias string `json:"apiKeyAlias,omitempty"`

	// API Configuration
	BaseURL string `json:"baseUrl"` // Default: "https://api.openai.com/v1"
	Model   string `json:"model"`   // Default: "gpt-4o"

	// Request Settings
	Timeout     int     `json:"timeout"`     // Seconds, Default: 300 (5 minutes)
	MaxTokens   int     `json:"maxTokens"`   // Default: 4096
	Temperature float64 `json:"temperature"` // Default: 0.7

	// Retry Settings
	MaxRetries int `json:"maxRetries"` // Default: 3
	RetryDelay int `json:"retryDelay"` // Seconds, Default: 2
}

// TranslationConfig contains translation settings
type TranslationConfig struct {
	Enabled        bool   `json:"enabled"`        // Default: false
	TargetLanguage string `json:"targetLanguage"` // Default: "en"
	ContextPrompt  string `json:"contextPrompt"`  // Custom translation context
}

// AppConfig contains application preferences
type AppConfig struct {
	Theme           string `json:"theme"`           // Default: "auto" (auto/dark/light)
	AutoSave        bool   `json:"autoSave"`        // Default: true
	ShowLineNumbers bool   `json:"showLineNumbers"` // Default: true
	DefaultTemplate string `json:"defaultTemplate"` // Default: "dev"
}
