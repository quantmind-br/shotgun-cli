package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shotgun-cli/shotgun-cli/internal/core/ignore"
	gitignorelib "github.com/sabhiram/go-gitignore"
)

// FileSystemScanner implements the Scanner interface for local file systems
// PathMatcher defines an interface for evaluating whether paths should be ignored
type PathMatcher interface {
	Matches(path string, isDir bool) bool
}

// gitIgnoreAdapter wraps *gitignorelib.GitIgnore to implement PathMatcher
type gitIgnoreAdapter struct {
	engine *gitignorelib.GitIgnore
}

// Matches implements PathMatcher interface
func (g *gitIgnoreAdapter) Matches(path string, isDir bool) bool {
	if g.engine == nil {
		return false
	}
	return g.engine.MatchesPath(path)
}

type FileSystemScanner struct {
	pathMatcher  PathMatcher
	ignoreEngine ignore.IgnoreEngine
}

// NewFileSystemScanner creates a new file system scanner
func NewFileSystemScanner() *FileSystemScanner {
	return &FileSystemScanner{
		ignoreEngine: ignore.NewIgnoreEngine(),
	}
}

// NewFileSystemScannerWithIgnore creates a new file system scanner with ignore rules
func NewFileSystemScannerWithIgnore(ignoreRules []string) (*FileSystemScanner, error) {
	scanner := NewFileSystemScanner()

	if len(ignoreRules) > 0 {
		if err := scanner.ignoreEngine.AddCustomRules(ignoreRules); err != nil {
			return nil, fmt.Errorf("failed to add custom ignore rules: %w", err)
		}
	}

	return scanner, nil
}

// NewFileSystemScannerWithMatcher creates a FileSystemScanner with a custom PathMatcher
// This is kept for backward compatibility but the new ignore engine is preferred
func NewFileSystemScannerWithMatcher(m PathMatcher) *FileSystemScanner {
	return &FileSystemScanner{
		pathMatcher:  m,
		ignoreEngine: ignore.NewIgnoreEngine(),
	}
}

// LoadGitIgnoreMatcher loads .gitignore rules from the specified root directory and returns a PathMatcher
// This is kept for backward compatibility but the new ignore engine is preferred
func LoadGitIgnoreMatcher(rootPath string) (PathMatcher, error) {
	gitignorePath := filepath.Join(rootPath, ".gitignore")

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil, nil // No .gitignore file
	}

	ignoreEngine, err := gitignorelib.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load .gitignore from %s: %w", gitignorePath, err)
	}

	return &gitIgnoreAdapter{engine: ignoreEngine}, nil
}

// LoadGitIgnore loads .gitignore rules from the specified root directory
func (fs *FileSystemScanner) LoadGitIgnore(rootPath string) error {
	// Use the new ignore engine if available
	if fs.ignoreEngine != nil {
		return fs.ignoreEngine.LoadGitignore(rootPath)
	}

	// Fallback to old PathMatcher approach for backward compatibility
	matcher, err := LoadGitIgnoreMatcher(rootPath)
	if err != nil {
		return err
	}

	if matcher != nil {
		fs.pathMatcher = matcher
	}

	return nil
}

// Scan performs a basic file system scan without progress reporting
func (fs *FileSystemScanner) Scan(rootPath string, config *ScanConfig) (*FileNode, error) {
	return fs.ScanWithProgress(rootPath, config, nil)
}

// ScanWithProgress performs a file system scan with progress reporting
func (fs *FileSystemScanner) ScanWithProgress(rootPath string, config *ScanConfig, progress chan<- Progress) (*FileNode, error) {
	if config == nil {
		config = DefaultScanConfig()
	}

	// Load .gitignore and configure ignore engine with custom patterns
	if fs.ignoreEngine != nil {
		if err := fs.ignoreEngine.LoadGitignore(rootPath); err != nil {
			return nil, fmt.Errorf("failed to load gitignore rules: %w", err)
		}

		// Add custom patterns from config
		if len(config.IgnorePatterns) > 0 {
			if err := fs.ignoreEngine.AddCustomRules(config.IgnorePatterns); err != nil {
				return nil, fmt.Errorf("failed to add custom ignore patterns: %w", err)
			}
		}
	} else {
		// Fallback to old approach if ignore engine is not available
		if fs.pathMatcher == nil {
			if err := fs.LoadGitIgnore(rootPath); err != nil {
				return nil, fmt.Errorf("failed to load ignore rules: %w", err)
			}
		}
	}

	// First pass: count total items for accurate progress reporting
	total, err := fs.countItems(rootPath, config)
	if err != nil {
		return nil, fmt.Errorf("failed to count items: %w", err)
	}

	if progress != nil {
		progress <- Progress{
			Current:   0,
			Total:     total,
			Stage:     "scanning",
			Message:   "Starting scan...",
			Timestamp: time.Now(),
		}
	}

	// Second pass: build the file tree
	root, actualCount, err := fs.walkAndBuild(rootPath, config, progress, total)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// Sort children for consistent ordering
	fs.sortChildren(root)

	if progress != nil {
		progress <- Progress{
			Current:   actualCount,
			Total:     total,
			Stage:     "complete",
			Message:   "Scan completed successfully",
			Timestamp: time.Now(),
		}
	}

	return root, nil
}

// countItems performs a quick count of all items to be processed
func (fs *FileSystemScanner) countItems(rootPath string, config *ScanConfig) (int64, error) {
	var count int64
	var fileCount int64

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories/files
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file limits (MaxFiles applies only to files, not directories)
		if config.MaxFiles > 0 && !d.IsDir() && fileCount >= config.MaxFiles {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil // Skip on error
		}

		// Check if should be ignored
		if fs.shouldIgnore(relPath, d.IsDir(), config) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply MaxFileSize filter for files during counting pass
		if !d.IsDir() && config.MaxFileSize > 0 {
			if info, err := d.Info(); err == nil {
				if info.Size() > config.MaxFileSize {
					return nil // Skip large files, don't count them
				}
			}
			// If we can't get file info, ignore the error gracefully and continue counting
		}

		count++
		// Only count files towards MaxFiles limit
		if !d.IsDir() {
			fileCount++
		}
		return nil
	})

	return count, err
}

// walkAndBuild builds the file tree with progress reporting
func (fs *FileSystemScanner) walkAndBuild(rootPath string, config *ScanConfig, progress chan<- Progress, total int64) (*FileNode, int64, error) {
	var current int64
	var fileCount int64

	// Create root node
	root := &FileNode{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		RelPath:  ".",
		IsDir:    true,
		Children: make([]*FileNode, 0),
		Selected: false,
		Expanded: true,
	}

	// Map to keep track of directory nodes for building the tree
	dirNodes := make(map[string]*FileNode)
	dirNodes[normRel(".")] = root

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories/files but continue scanning
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file limits (MaxFiles applies only to files, not directories)
		if config.MaxFiles > 0 && !d.IsDir() && fileCount >= config.MaxFiles {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil // Skip on error
		}

		// Skip root directory itself
		if relPath == "." {
			return nil
		}

		// Check if should be ignored
		isGitignored, isCustomIgnored := fs.getIgnoreStatus(relPath, d.IsDir(), config)
		if (isGitignored || isCustomIgnored) && !fs.shouldIncludeIgnored(config) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get file size for files
		var size int64
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				size = info.Size()

				// Check file size limits
				if config.MaxFileSize > 0 && size > config.MaxFileSize {
					return nil // Skip large files
				}
			}
		}

		// Create file node
		node := &FileNode{
			Name:            d.Name(),
			Path:            path,
			RelPath:         relPath,
			IsDir:           d.IsDir(),
			Children:        make([]*FileNode, 0),
			Selected:        false,
			IsGitignored:    isGitignored,
			IsCustomIgnored: isCustomIgnored,
			Size:            size,
			Expanded:        false,
		}

		// Find parent node and add this node to it
		parentNode := fs.findParentNode(relPath, dirNodes)
		if parentNode != nil {
			node.Parent = parentNode
			parentNode.Children = append(parentNode.Children, node)
		}

		// If this is a directory, add it to the directory map with normalized path
		if d.IsDir() {
			dirNodes[normRel(relPath)] = node
		}

		current++
		// Only count files towards MaxFiles limit
		if !d.IsDir() {
			fileCount++
		}

		// Report progress every 100 items to avoid UI blocking
		if progress != nil && current%100 == 0 {
			progress <- Progress{
				Current:   current,
				Total:     total,
				Stage:     "scanning",
				Message:   fmt.Sprintf("Processing: %s", relPath),
				Timestamp: time.Now(),
			}
		}

		return nil
	})

	return root, current, err
}

// findParentNode finds the parent directory node for a given relative path
func (fs *FileSystemScanner) findParentNode(relPath string, dirNodes map[string]*FileNode) *FileNode {
	parentPath := filepath.Dir(relPath)
	if parentPath == "." {
		return dirNodes[normRel(".")]
	}

	// Normalize path separators for consistency
	parentPath = normRel(parentPath)
	if node, exists := dirNodes[parentPath]; exists {
		return node
	}

	// Fallback: search for parent by walking up the path
	parts := strings.Split(parentPath, "/")
	for i := len(parts); i > 0; i-- {
		testPath := strings.Join(parts[:i], "/")
		if node, exists := dirNodes[testPath]; exists {
			return node
		}
	}

	return dirNodes[normRel(".")] // Return root if no parent found
}

// sortChildren sorts the children of all directory nodes recursively
func (fs *FileSystemScanner) sortChildren(node *FileNode) {
	if !node.IsDir {
		return
	}

	// Sort children: directories first, then files, both alphabetically
	sort.Slice(node.Children, func(i, j int) bool {
		a, b := node.Children[i], node.Children[j]

		// Directories come first
		if a.IsDir && !b.IsDir {
			return true
		}
		if !a.IsDir && b.IsDir {
			return false
		}

		// Within same type, sort alphabetically (case-insensitive)
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	// Recursively sort children
	for _, child := range node.Children {
		fs.sortChildren(child)
	}
}

// shouldIgnore checks if a path should be ignored based on all rules
func (fs *FileSystemScanner) shouldIgnore(relPath string, isDir bool, config *ScanConfig) bool {
	isGitignored, isCustomIgnored := fs.getIgnoreStatus(relPath, isDir, config)
	return (isGitignored || isCustomIgnored) && !fs.shouldIncludeIgnored(config)
}

// getIgnoreStatus returns the ignore status for both gitignore and custom rules
func (fs *FileSystemScanner) getIgnoreStatus(relPath string, isDir bool, config *ScanConfig) (bool, bool) {
	// Use the new ignore engine if available
	if fs.ignoreEngine != nil {
		// Check if path should be ignored by the layered ignore engine
		ignored, reason := fs.ignoreEngine.ShouldIgnore(relPath)

		if ignored {
			// Determine the ignore type based on the reason
			switch reason {
			case ignore.IgnoreReasonGitignore:
				return true, false // Gitignored but not custom ignored
			case ignore.IgnoreReasonBuiltIn, ignore.IgnoreReasonCustom, ignore.IgnoreReasonExplicit:
				return false, true // Custom ignored but not gitignored
			}
		}

		// Check hidden files/directories if not included and not already ignored
		if !config.IncludeHidden && !ignored {
			baseName := filepath.Base(relPath)
			if strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".." {
				return false, true // Hidden files are treated as custom ignored
			}
		}

		// Use the engine's specific methods for precise classification
		isGitignored := fs.ignoreEngine.IsGitignored(relPath)
		isCustomIgnored := fs.ignoreEngine.IsCustomIgnored(relPath) || ignored && reason != ignore.IgnoreReasonGitignore

		return isGitignored, isCustomIgnored
	}

	// Fallback to old logic if ignore engine is not available
	var isGitignored, isCustomIgnored bool

	// Check .gitignore rules
	if fs.pathMatcher != nil {
		isGitignored = fs.pathMatcher.Matches(relPath, isDir)
	}

	// Check hidden files/directories
	if !config.IncludeHidden {
		baseName := filepath.Base(relPath)
		if strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".." {
			isCustomIgnored = true
		}
	}

	// Check custom ignore patterns
	for _, pattern := range config.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, relPath); matched {
			isCustomIgnored = true
			break
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			isCustomIgnored = true
			break
		}
	}

	return isGitignored, isCustomIgnored
}

// shouldIncludeIgnored determines if ignored files should be included based on config
func (fs *FileSystemScanner) shouldIncludeIgnored(config *ScanConfig) bool {
	// For now, we always exclude ignored files during scanning
	// This can be extended based on configuration needs
	return false
}

// normRel normalizes a relative path for consistent map lookups
func normRel(relPath string) string {
	return filepath.ToSlash(relPath)
}

// SkipDir is a sentinel error used with filepath.WalkDir to skip directories
var SkipDir = filepath.SkipDir