// Package ignore provides file and directory ignore pattern matching.
package ignore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

// Reason represents the reason why a path was ignored.
// Deprecated: Use Reason instead of IgnoreReason.
type Reason = IgnoreReason //nolint:revive // keeping IgnoreReason for backward compatibility

// IgnoreReason represents the reason why a path was ignored.
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

	// LoadShotgunignore loads .shotgunignore rules from the specified directory
	LoadShotgunignore(rootDir string) error
}

// LayeredIgnoreEngine implements the IgnoreEngine interface with layered rule support
type LayeredIgnoreEngine struct {
	builtInMatcher   *gitignore.GitIgnore
	gitignoreMatcher *gitignore.GitIgnore
	customMatcher    *gitignore.GitIgnore
	explicitExcludes *gitignore.GitIgnore
	explicitIncludes *gitignore.GitIgnore

	// Store patterns for accumulation across calls
	customPatterns          []string
	explicitExcludePatterns []string
	explicitIncludePatterns []string
}

// NewIgnoreEngine creates a new layered ignore engine with built-in patterns
func NewIgnoreEngine() *LayeredIgnoreEngine {
	engine := &LayeredIgnoreEngine{}

	// Initialize built-in patterns
	builtInPatterns := []string{
		// Shotgun-specific patterns
		"shotgun-prompt*.md",

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
		".pytest_cache/",
		".mypy_cache/",

		// Images and Media
		"*.png", "*.jpg", "*.jpeg", "*.gif", "*.ico", "*.svg", "*.webp",
		"*.mp3", "*.mp4", "*.wav", "*.avi", "*.mov", "*.mkv",

		// Fonts
		"*.ttf", "*.otf", "*.woff", "*.woff2", "*.eot",

		// Documents
		"*.pdf", "*.doc", "*.docx", "*.xls", "*.xlsx", "*.ppt", "*.pptx",

		// Binary executables and libs
		"*.exe", "*.dll", "*.so", "*.dylib",

		// Databases
		"*.sqlite", "*.sqlite3", "*.db",

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
		"*.7z",

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
	// Collect all .gitignore files in the directory tree
	var gitignoreFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == ".gitignore" {
			gitignoreFiles = append(gitignoreFiles, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory for gitignore files: %w", err)
	}

	// If no .gitignore files found, use empty matcher
	if len(gitignoreFiles) == 0 {
		e.gitignoreMatcher = gitignore.CompileIgnoreLines()

		return nil
	}

	// Collect all patterns from all .gitignore files
	var allPatterns []string

	for _, gitignoreFile := range gitignoreFiles {
		// Read the file content
		content, err := os.ReadFile(gitignoreFile) //nolint:gosec // path comes from controlled directory walk
		if err != nil {
			continue // Skip files we can't read
		}

		// Get relative path from root to adjust patterns
		relDir, err := filepath.Rel(rootDir, filepath.Dir(gitignoreFile))
		if err != nil {
			continue
		}

		// Split content into lines and process each pattern
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue // Skip empty lines and comments
			}

			// If this is a nested .gitignore (not root), prefix patterns with relative path
			if relDir != "." && relDir != "" {
				// Adjust pattern for nested location
				if strings.HasPrefix(line, "!") {
					// Negation pattern - prefix after the !
					line = "!" + filepath.Join(relDir, line[1:])
				} else {
					line = filepath.Join(relDir, line)
				}
			}

			allPatterns = append(allPatterns, line)
		}
	}

	// Compile all patterns into a single matcher
	e.gitignoreMatcher = gitignore.CompileIgnoreLines(allPatterns...)

	return nil
}

// AddCustomRule adds a custom ignore pattern
func (e *LayeredIgnoreEngine) AddCustomRule(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Add pattern to accumulated list
	e.customPatterns = append(e.customPatterns, pattern)

	// Recompile matcher with all accumulated patterns
	if len(e.customPatterns) > 0 {
		e.customMatcher = gitignore.CompileIgnoreLines(e.customPatterns...)
	}

	return nil
}

// AddCustomRules adds multiple custom ignore patterns
func (e *LayeredIgnoreEngine) AddCustomRules(patterns []string) error {
	if len(patterns) == 0 {
		return nil
	}

	// Filter out empty patterns and trim whitespace
	validPatterns := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed != "" {
			validPatterns = append(validPatterns, trimmed)
		}
	}

	if len(validPatterns) == 0 {
		return nil
	}

	// Accumulate patterns with existing customPatterns
	e.customPatterns = append(e.customPatterns, validPatterns...)

	// Recompile matcher with all accumulated patterns
	e.customMatcher = gitignore.CompileIgnoreLines(e.customPatterns...)

	return nil
}

// AddExplicitExclude adds a pattern that should always be excluded
func (e *LayeredIgnoreEngine) AddExplicitExclude(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Add pattern to accumulated list
	e.explicitExcludePatterns = append(e.explicitExcludePatterns, pattern)

	// Recompile matcher with all accumulated patterns
	if len(e.explicitExcludePatterns) > 0 {
		e.explicitExcludes = gitignore.CompileIgnoreLines(e.explicitExcludePatterns...)
	}

	return nil
}

// AddExplicitInclude adds a pattern that should always be included
func (e *LayeredIgnoreEngine) AddExplicitInclude(pattern string) error {
	if pattern == "" {
		return nil
	}

	// Add pattern to accumulated list
	e.explicitIncludePatterns = append(e.explicitIncludePatterns, pattern)

	// Recompile matcher with all accumulated patterns
	if len(e.explicitIncludePatterns) > 0 {
		e.explicitIncludes = gitignore.CompileIgnoreLines(e.explicitIncludePatterns...)
	}

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

// LoadShotgunignore loads .shotgunignore rules from the specified directory
func (e *LayeredIgnoreEngine) LoadShotgunignore(rootDir string) error {
	// Collect all .shotgunignore files in the directory tree
	var shotgunignoreFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == ".shotgunignore" {
			shotgunignoreFiles = append(shotgunignoreFiles, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory for shotgunignore files: %w", err)
	}

	// If no .shotgunignore files found, return early
	if len(shotgunignoreFiles) == 0 {
		return nil
	}

	// Collect all patterns from all .shotgunignore files
	var allPatterns []string

	for _, shotgunignoreFile := range shotgunignoreFiles {
		// Read the file content
		content, err := os.ReadFile(shotgunignoreFile) //nolint:gosec // path comes from controlled directory walk
		if err != nil {
			continue // Skip files we can't read
		}

		// Get relative path from root to adjust patterns
		relDir, err := filepath.Rel(rootDir, filepath.Dir(shotgunignoreFile))
		if err != nil {
			continue
		}

		// Split content into lines and process each pattern
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue // Skip empty lines and comments
			}

			// If this is a nested .shotgunignore (not root), prefix patterns with relative path
			if relDir != "." && relDir != "" {
				// Adjust pattern for nested location
				if strings.HasPrefix(line, "!") {
					// Negation pattern - prefix after the !
					line = "!" + filepath.Join(relDir, line[1:])
				} else {
					line = filepath.Join(relDir, line)
				}
			}

			allPatterns = append(allPatterns, line)
		}
	}

	// Add all patterns as custom rules
	if len(allPatterns) > 0 {
		return e.AddCustomRules(allPatterns)
	}

	return nil
}
