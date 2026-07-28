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
	if matchesWithAncestors(e.gitignoreMatcher, normalizedPath) {
		return true, IgnoreReasonGitignore
	}

	// 5. Check custom patterns (lowest priority). .shotgunignore feeds this
	// layer, and both use gitignore semantics.
	if matchesWithAncestors(e.customMatcher, normalizedPath) {
		return true, IgnoreReasonCustom
	}

	// Path is not ignored
	return false, IgnoreReasonNone
}

// matchesWithAncestors reports whether matcher excludes normalizedPath, applying
// git's rule that a negation cannot re-include a file whose parent directory is
// excluded.
//
// Without the ancestor check, a nested ignore file re-includes anything it names,
// because every level's patterns compile into one last-match-wins matcher and the
// nested ones come later. Git instead stops descending at an excluded directory,
// so the negation is never consulted. The difference leaks: with `secrets/` in the
// root .gitignore and any `!` line in `secrets/.gitignore`, the named file was
// handed to the LLM despite the user excluding the directory.
func matchesWithAncestors(matcher *gitignore.GitIgnore, normalizedPath string) bool {
	if matcher == nil {
		return false
	}

	if hasIgnoredAncestor(matcher, normalizedPath) {
		return true
	}

	return matcher.MatchesPath(normalizedPath)
}

// hasIgnoredAncestor walks the path's directories from the shallowest down,
// mirroring the order git descends in: the first excluded directory settles the
// question for everything beneath it.
func hasIgnoredAncestor(matcher *gitignore.GitIgnore, normalizedPath string) bool {
	parts := strings.Split(normalizedPath, "/")
	for i := 1; i < len(parts); i++ {
		dir := strings.Join(parts[:i], "/")
		// The trailing slash matters: it is how a directory-only pattern such as
		// `sub/` is matched.
		if matcher.MatchesPath(dir+"/") || matcher.MatchesPath(dir) {
			return true
		}
	}

	return false
}

// LoadGitignore loads .gitignore rules from the specified directory
func (e *LayeredIgnoreEngine) LoadGitignore(rootDir string) error {
	patterns, err := collectIgnorePatterns(rootDir, ".gitignore")
	if err != nil {
		return fmt.Errorf("failed to walk directory for gitignore files: %w", err)
	}

	e.gitignoreMatcher = gitignore.CompileIgnoreLines(patterns...)

	return nil
}

// collectIgnorePatterns walks rootDir gathering the patterns of every ignore
// file called name, rebased so that nested files keep gitignore semantics
// relative to rootDir. Unreadable directories and files are skipped: one
// permission-denied subtree must not discard the rules of the whole project.
func collectIgnorePatterns(rootDir, name string) ([]string, error) {
	var files []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Unreadable entry: skip it and keep walking the rest of the tree.
			return nil //nolint:nilerr // tolerating unreadable subtrees is the point
		}
		if !info.IsDir() && info.Name() == name {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var patterns []string

	for _, file := range files {
		content, err := os.ReadFile(file) //nolint:gosec // path comes from controlled directory walk
		if err != nil {
			continue // Skip files we can't read
		}

		relDir, err := filepath.Rel(rootDir, filepath.Dir(file))
		if err != nil {
			continue
		}
		relDir = filepath.ToSlash(relDir)

		for _, line := range strings.Split(string(content), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue // Skip empty lines and comments
			}
			patterns = append(patterns, rebasePattern(relDir, line))
		}
	}

	return patterns, nil
}

// rebasePattern rewrites a pattern taken from an ignore file in relDir so that
// it means the same thing when matched against paths relative to the scan root.
//
// Git semantics: a pattern without a slash (other than a trailing one) matches
// at any depth below its own directory, so it becomes "relDir/**/pattern"; a
// pattern that is anchored by a slash keeps its position under relDir. The
// trailing slash that restricts a pattern to directories is preserved, which
// filepath.Join would silently strip.
func rebasePattern(relDir, pattern string) string {
	if relDir == "." || relDir == "" {
		return pattern
	}

	negated := strings.HasPrefix(pattern, "!")
	if negated {
		pattern = pattern[1:]
	}

	var rebased string
	switch {
	case strings.HasPrefix(pattern, "/"):
		rebased = relDir + pattern
	case strings.Contains(strings.TrimSuffix(pattern, "/"), "/"):
		rebased = relDir + "/" + pattern
	default:
		rebased = relDir + "/**/" + pattern
	}

	if negated {
		return "!" + rebased
	}

	return rebased
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

// IsGitignored returns true if the path would be ignored by .gitignore rules specifically.
func (e *LayeredIgnoreEngine) IsGitignored(relPath string) bool {
	normalizedPath := filepath.ToSlash(relPath)

	return e.gitignoreMatcher.MatchesPath(normalizedPath)
}

// IsCustomIgnored returns true if the path would be ignored by custom rules specifically.
func (e *LayeredIgnoreEngine) IsCustomIgnored(relPath string) bool {
	normalizedPath := filepath.ToSlash(relPath)

	return e.customMatcher.MatchesPath(normalizedPath)
}

// LoadShotgunignore loads .shotgunignore rules from the specified directory
func (e *LayeredIgnoreEngine) LoadShotgunignore(rootDir string) error {
	patterns, err := collectIgnorePatterns(rootDir, ".shotgunignore")
	if err != nil {
		return fmt.Errorf("failed to walk directory for shotgunignore files: %w", err)
	}

	if len(patterns) == 0 {
		return nil
	}

	return e.AddCustomRules(patterns)
}
