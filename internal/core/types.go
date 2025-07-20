package core

import (
	"sync"
	"text/template"

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
	// Original fields (maintained for backward compatibility)
	Task          string
	Rules         string
	CurrentDate   string
	FileStructure string
	
	// Enhanced formatted fields for numbered content
	FormattedTask  string
	FormattedRules string
	
	// Formatting options and metadata
	FormattingOptions FormattingOptions
	HasNumberedTasks  bool
	HasNumberedRules  bool
}

// SelectionState manages file inclusion/exclusion state
type SelectionState struct {
	included map[string]bool
	excluded map[string]bool
	patterns []string
	mu       sync.RWMutex
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
		included: make(map[string]bool),
		excluded: make(map[string]bool),
		patterns: make([]string, 0),
	}
}

// IsFileIncluded checks if a file should be included in the final output
func (ss *SelectionState) IsFileIncluded(path string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// Check explicit exclusion first
	if ss.excluded[path] {
		return false
	}

	// Default to included
	return true
}

// ToggleFile toggles the exclusion state of a file
func (ss *SelectionState) ToggleFile(path string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.excluded[path] {
		delete(ss.excluded, path)
	} else {
		ss.excluded[path] = true
	}
}

// ExcludeFile explicitly excludes a file
func (ss *SelectionState) ExcludeFile(path string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.excluded[path] = true
}

// IncludeFile explicitly includes a file (removes from excluded)
func (ss *SelectionState) IncludeFile(path string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.excluded, path)
}

// Reset clears all exclusions
func (ss *SelectionState) Reset() {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.excluded = make(map[string]bool)
}

// GetExcludedFiles returns a copy of excluded files
func (ss *SelectionState) GetExcludedFiles() map[string]bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	result := make(map[string]bool)
	for k, v := range ss.excluded {
		result[k] = v
	}
	return result
}
