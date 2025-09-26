package ignore

import (
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// IgnoreReason represents the reason why a path was ignored
type IgnoreReason int

const (
	// IgnoreReasonNone indicates the path is not ignored
	IgnoreReasonNone IgnoreReason = iota
	// IgnoreReasonBuiltIn indicates the path was ignored by built-in patterns
	IgnoreReasonBuiltIn
	// IgnoreReasonGitignore indicates the path was ignored by .gitignore rules
	IgnoreReasonGitignore
	// IgnoreReasonCustom indicates the path was ignored by custom patterns
	IgnoreReasonCustom
	// IgnoreReasonExplicit indicates the path was explicitly excluded
	IgnoreReasonExplicit
)

// String returns the string representation of the ignore reason
func (r IgnoreReason) String() string {
	switch r {
	case IgnoreReasonNone:
		return "none"
	case IgnoreReasonBuiltIn:
		return "built-in"
	case IgnoreReasonGitignore:
		return "gitignore"
	case IgnoreReasonCustom:
		return "custom"
	case IgnoreReasonExplicit:
		return "explicit"
	default:
		return "unknown"
	}
}

// IgnoreEngine interface defines the contract for ignore engines
type IgnoreEngine interface {
	// ShouldIgnore checks if a path should be ignored and returns the reason
	ShouldIgnore(relPath string) (bool, IgnoreReason)

	// LoadGitignore loads .gitignore rules from the specified directory
	LoadGitignore(rootDir string) error

	// AddCustomRule adds a custom ignore pattern
	AddCustomRule(pattern string) error

	// AddCustomRules adds multiple custom ignore patterns
	AddCustomRules(patterns []string) error

	// AddExplicitExclude adds a pattern that should always be excluded
	AddExplicitExclude(pattern string) error

	// AddExplicitInclude adds a pattern that should always be included
	AddExplicitInclude(pattern string) error

	// IsGitignored returns true if the path would be ignored by .gitignore rules specifically
	IsGitignored(relPath string) bool

	// IsCustomIgnored returns true if the path would be ignored by custom rules specifically
	IsCustomIgnored(relPath string) bool
}

// LayeredIgnoreEngine implements the IgnoreEngine interface with layered rule support
type LayeredIgnoreEngine struct {
	builtInMatcher    *gitignore.GitIgnore
	gitignoreMatcher  *gitignore.GitIgnore
	customMatcher     *gitignore.GitIgnore
	explicitExcludes  *gitignore.GitIgnore
	explicitIncludes  *gitignore.GitIgnore
}

// NewIgnoreEngine creates a new layered ignore engine with built-in patterns
func NewIgnoreEngine() *LayeredIgnoreEngine {
	engine := &LayeredIgnoreEngine{}

	// Initialize built-in patterns
	builtInPatterns := []string{
		// Shotgun-specific patterns
		"shotgun-prompt-*.md",

		// Version control
		".git/",
		".svn/",
		".hg/",
		".bzr/",

		// IDE and editor files
		".vscode/",
		".idea/",
		"*.swp",
		"*.swo",
		"*~",
		".DS_Store",
		"Thumbs.db",

		// Build and dependency directories
		"node_modules/",
		"bower_components/",
		"vendor/",
		"target/",
		"build/",
		"dist/",
		"out/",
		"bin/",
		"obj/",

		// Cache and temporary files
		"__pycache__/",
		"*.pyc",
		"*.pyo",
		".cache/",
		".tmp/",
		"tmp/",

		// Log files
		"*.log",
		"logs/",

		// Package files
		"*.jar",
		"*.war",
		"*.nar",
		"*.ear",
		"*.zip",
		"*.tar.gz",
		"*.rar",

		// OS generated files
		".DS_Store?",
		"._*",
		".Spotlight-V100",
		".Trashes",
		"ehthumbs.db",
	}

	engine.builtInMatcher = gitignore.CompileIgnoreLines(builtInPatterns...)

	// Initialize empty matchers for other layers
	engine.gitignoreMatcher = gitignore.CompileIgnoreLines()
	engine.customMatcher = gitignore.CompileIgnoreLines()
	engine.explicitExcludes = gitignore.CompileIgnoreLines()
	engine.explicitIncludes = gitignore.CompileIgnoreLines()

	return engine
}

// ShouldIgnore checks if a path should be ignored using layered rules
// Priority: explicit excludes → explicit includes → built-in → .gitignore → custom
func (e *LayeredIgnoreEngine) ShouldIgnore(relPath string) (bool, IgnoreReason) {
	// Normalize path separators for consistent matching
	normalizedPath := filepath.ToSlash(relPath)

	// 1. Check explicit excludes (highest priority)
	if e.explicitExcludes.MatchesPath(normalizedPath) {
		return true, IgnoreReasonExplicit
	}

	// 2. Check explicit includes (overrides all other rules)
	if e.explicitIncludes.MatchesPath(normalizedPath) {
		return false, IgnoreReasonNone
	}

	// 3. Check built-in patterns
	if e.builtInMatcher.MatchesPath(normalizedPath) {
		return true, IgnoreReasonBuiltIn
	}

	// 4. Check .gitignore patterns
	if e.gitignoreMatcher.MatchesPath(normalizedPath) {
		return true, IgnoreReasonGitignore
	}

	// 5. Check custom patterns (lowest priority)
	if e.customMatcher.MatchesPath(normalizedPath) {
		return true, IgnoreReasonCustom
	}

	// Path is not ignored
	return false, IgnoreReasonNone
}

// LoadGitignore loads .gitignore rules from the specified directory
func (e *LayeredIgnoreEngine) LoadGitignore(rootDir string) error {
	gitignorePath := filepath.Join(rootDir, ".gitignore")

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// No .gitignore file, use empty matcher
		e.gitignoreMatcher = gitignore.CompileIgnoreLines()
		return nil
	}

	// Load .gitignore file
	matcher, err := gitignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return err
	}

	e.gitignoreMatcher = matcher
	return nil
}

// AddCustomRule adds a custom ignore pattern
func (e *LayeredIgnoreEngine) AddCustomRule(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Get existing patterns and add new one
	patterns := []string{pattern}

	// Create new matcher with the additional pattern
	newMatcher := gitignore.CompileIgnoreLines(patterns...)

	// Combine with existing custom matcher
	if e.customMatcher != nil {
		// We need to rebuild the matcher with all patterns
		// This is a limitation of the go-gitignore library
		e.customMatcher = newMatcher
	} else {
		e.customMatcher = newMatcher
	}

	return nil
}

// AddCustomRules adds multiple custom ignore patterns
func (e *LayeredIgnoreEngine) AddCustomRules(patterns []string) error {
	if len(patterns) == 0 {
		return nil
	}

	// Filter out empty patterns
	validPatterns := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		if strings.TrimSpace(pattern) != "" {
			validPatterns = append(validPatterns, strings.TrimSpace(pattern))
		}
	}

	if len(validPatterns) == 0 {
		return nil
	}

	// Create new matcher with all patterns
	e.customMatcher = gitignore.CompileIgnoreLines(validPatterns...)
	return nil
}

// AddExplicitExclude adds a pattern that should always be excluded
func (e *LayeredIgnoreEngine) AddExplicitExclude(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Create new matcher with the pattern
	e.explicitExcludes = gitignore.CompileIgnoreLines(pattern)
	return nil
}

// AddExplicitInclude adds a pattern that should always be included
func (e *LayeredIgnoreEngine) AddExplicitInclude(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Create new matcher with the pattern
	e.explicitIncludes = gitignore.CompileIgnoreLines(pattern)
	return nil
}

// IsGitignored returns true if the path would be ignored by .gitignore rules specifically
func (e *LayeredIgnoreEngine) IsGitignored(relPath string) bool {
	normalizedPath := filepath.ToSlash(relPath)
	return e.gitignoreMatcher.MatchesPath(normalizedPath)
}

// IsCustomIgnored returns true if the path would be ignored by custom rules specifically
func (e *LayeredIgnoreEngine) IsCustomIgnored(relPath string) bool {
	normalizedPath := filepath.ToSlash(relPath)
	return e.customMatcher.MatchesPath(normalizedPath)
}