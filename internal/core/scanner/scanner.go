package scanner

import (
	"fmt"
	"time"
)

// Scanner defines the interface for file system scanning operations
type Scanner interface {
	// Scan performs a basic file system scan without progress reporting
	Scan(rootPath string, config *ScanConfig) (*FileNode, error)

	// ScanWithProgress performs a file system scan with progress reporting
	ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error)
}

// FileNode represents a file or directory in the file system tree
type FileNode struct {
	// Name is the base name of the file or directory
	Name string `json:"name"`

	// Path is the absolute path to the file or directory
	Path string `json:"path"`

	// RelPath is the relative path from the scan root
	RelPath string `json:"rel_path"`

	// IsDir indicates whether this node represents a directory
	IsDir bool `json:"is_dir"`

	// Children contains child nodes for directories
	Children []*FileNode `json:"children,omitempty"`

	// Selected indicates if this node is selected in the UI
	Selected bool `json:"selected"`

	// IsGitignored indicates if this file is ignored by .gitignore
	IsGitignored bool `json:"is_gitignored"`

	// IsCustomIgnored indicates if this file is ignored by custom rules
	IsCustomIgnored bool `json:"is_custom_ignored"`

	// Size is the file size in bytes (0 for directories)
	Size int64 `json:"size"`

	// Expanded indicates if this directory node is expanded in the TUI
	Expanded bool `json:"expanded"`

	// Parent reference for tree navigation (not serialized)
	Parent *FileNode `json:"-"`
}

// Progress represents the current state of a scanning operation
type Progress struct {
	// Current number of items processed
	Current int64 `json:"current"`

	// Total number of items to process
	Total int64 `json:"total"`

	// Stage describes the current scanning stage
	Stage string `json:"stage"`

	// Message provides additional context about the current operation
	Message string `json:"message,omitempty"`

	// Timestamp when this progress update was created
	Timestamp time.Time `json:"timestamp"`
}

// ScanConfig contains configuration options for file scanning
type ScanConfig struct {
	// MaxFileSize limits the size of files to include (in bytes, 0 = no limit)
	MaxFileSize int64 `json:"max_file_size"`

	// MaxFiles limits the total number of files to scan (0 = no limit)
	MaxFiles int64 `json:"max_files"`

	// MaxMemory limits memory usage during scanning (in bytes, 0 = no limit)
	MaxMemory int64 `json:"max_memory"`

	// SkipBinary indicates whether to skip binary files
	SkipBinary bool `json:"skip_binary"`

	// IncludeHidden indicates whether to include hidden files and directories
	IncludeHidden bool `json:"include_hidden"`

	// Workers specifies the number of concurrent workers for scanning
	Workers int `json:"workers"`

	// IgnorePatterns contains custom ignore patterns
	// NOTE: These patterns now use gitignore semantics instead of legacy glob patterns.
	// Use gitignore-style patterns like "*.log", "build/", or "!important.txt" for negation.
	// Legacy glob patterns may not work as expected.
	IgnorePatterns []string `json:"ignore_patterns,omitempty"`

	// IncludePatterns contains patterns for files to include
	// If specified, only files matching these patterns will be included
	// Uses glob-style patterns like "*.go", "*.js", etc.
	IncludePatterns []string `json:"include_patterns,omitempty"`
}

// DefaultScanConfig returns a default scanning configuration
func DefaultScanConfig() *ScanConfig {
	return &ScanConfig{
		MaxFileSize:   0, // No limit
		MaxFiles:      0, // No limit
		MaxMemory:     0, // No limit
		SkipBinary:    false,
		IncludeHidden: false,
		Workers:       1, // Single-threaded by default for simplicity
	}
}

// IsIgnored returns true if the file node is ignored by any rule
func (f *FileNode) IsIgnored() bool {
	return f.IsGitignored || f.IsCustomIgnored
}

// GetIgnoreReason returns a human-readable reason why the file is ignored
func (f *FileNode) GetIgnoreReason() string {
	if f.IsGitignored && f.IsCustomIgnored {
		return "gitignored and custom ignored"
	}
	if f.IsGitignored {
		return "gitignored"
	}
	if f.IsCustomIgnored {
		return "custom ignored"
	}
	return "not ignored"
}

// CountChildren returns the total number of child nodes recursively
func (f *FileNode) CountChildren() int {
	if !f.IsDir {
		return 0
	}

	count := len(f.Children)
	for _, child := range f.Children {
		count += child.CountChildren()
	}
	return count
}

// CountFiles returns the total number of files (non-directories) recursively
func (f *FileNode) CountFiles() int {
	if !f.IsDir {
		return 1
	}

	count := 0
	for _, child := range f.Children {
		count += child.CountFiles()
	}
	return count
}

// CountDirectories returns the total number of directories recursively
func (f *FileNode) CountDirectories() int {
	if !f.IsDir {
		return 0
	}

	count := 1 // Count self
	for _, child := range f.Children {
		count += child.CountDirectories()
	}
	return count
}

// TotalSize returns the total size of all files under this node
func (f *FileNode) TotalSize() int64 {
	if !f.IsDir {
		return f.Size
	}

	var total int64
	for _, child := range f.Children {
		total += child.TotalSize()
	}
	return total
}

// FormatSize returns a human-readable size string
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Percentage calculates the completion percentage for progress reporting
func (p *Progress) Percentage() float64 {
	if p.Total == 0 {
		return 0.0
	}
	return float64(p.Current) / float64(p.Total) * 100.0
}

// String returns a formatted progress string
func (p *Progress) String() string {
	percentage := p.Percentage()
	if p.Message != "" {
		return fmt.Sprintf("%.1f%% (%d/%d) - %s: %s", percentage, p.Current, p.Total, p.Stage, p.Message)
	}
	return fmt.Sprintf("%.1f%% (%d/%d) - %s", percentage, p.Current, p.Total, p.Stage)
}
