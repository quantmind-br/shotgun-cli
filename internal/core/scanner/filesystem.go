package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/ignore"
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
		ignoreEngine: nil, // Don't initialize ignore engine to give pathMatcher precedence
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
			return fs.handleCountError(d)
		}

		if fs.shouldStopCounting(config, d, fileCount) {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil || relPath == "." {
			return nil
		}

		if fs.shouldIgnore(relPath, d.IsDir(), config) {
			return fs.skipIfDirectory(d)
		}

		if fs.shouldSkipLargeFile(d, config) {
			return nil
		}

		count++
		if !d.IsDir() {
			fileCount++
		}
		return nil
	})

	return count, err
}

func (fs *FileSystemScanner) handleCountError(d os.DirEntry) error {
	if d != nil && d.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

func (fs *FileSystemScanner) shouldStopCounting(config *ScanConfig, d os.DirEntry, fileCount int64) bool {
	return config.MaxFiles > 0 && !d.IsDir() && fileCount >= config.MaxFiles
}

func (fs *FileSystemScanner) skipIfDirectory(d os.DirEntry) error {
	if d.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

func (fs *FileSystemScanner) shouldSkipLargeFile(d os.DirEntry, config *ScanConfig) bool {
	if d.IsDir() || config.MaxFileSize <= 0 {
		return false
	}
	if info, err := d.Info(); err == nil {
		return info.Size() > config.MaxFileSize
	}
	return false
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
			return fs.handleWalkError(d)
		}

		if fs.shouldStopWalking(config, d, fileCount) {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil || relPath == "." {
			return nil
		}

		if fs.shouldIgnore(relPath, d.IsDir(), config) {
			return fs.skipIfDirectory(d)
		}

		size, skipFile := fs.getFileSize(d, config)
		if skipFile {
			return nil
		}

		node := fs.createFileNode(path, relPath, d, size, config)
		fs.addNodeToTree(node, relPath, dirNodes)

		current++
		if !d.IsDir() {
			fileCount++
		}

		fs.reportProgress(progress, current, total, relPath)
		return nil
	})

	return root, current, err
}

func (fs *FileSystemScanner) handleWalkError(d os.DirEntry) error {
	if d != nil && d.IsDir() {
		return filepath.SkipDir
	}
	return nil
}

func (fs *FileSystemScanner) shouldStopWalking(config *ScanConfig, d os.DirEntry, fileCount int64) bool {
	return config.MaxFiles > 0 && !d.IsDir() && fileCount >= config.MaxFiles
}

func (fs *FileSystemScanner) getFileSize(d os.DirEntry, config *ScanConfig) (int64, bool) {
	if d.IsDir() {
		return 0, false
	}
	if info, err := d.Info(); err == nil {
		size := info.Size()
		if config.MaxFileSize > 0 && size > config.MaxFileSize {
			return 0, true
		}
		return size, false
	}
	return 0, false
}

func (fs *FileSystemScanner) createFileNode(path, relPath string, d os.DirEntry, size int64, config *ScanConfig) *FileNode {
	isGitignored, isCustomIgnored := fs.getIgnoreStatus(relPath, d.IsDir(), config)

	return &FileNode{
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
}

func (fs *FileSystemScanner) addNodeToTree(node *FileNode, relPath string, dirNodes map[string]*FileNode) {
	parentNode := fs.findParentNode(relPath, dirNodes)
	if parentNode != nil {
		node.Parent = parentNode
		parentNode.Children = append(parentNode.Children, node)
	}

	if node.IsDir {
		dirNodes[normRel(relPath)] = node
	}
}

func (fs *FileSystemScanner) reportProgress(progress chan<- Progress, current, total int64, relPath string) {
	if progress != nil && current%100 == 0 {
		progress <- Progress{
			Current:   current,
			Total:     total,
			Stage:     "scanning",
			Message:   fmt.Sprintf("Processing: %s", relPath),
			Timestamp: time.Now(),
		}
	}
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

// matchesIncludePatterns checks if a file matches any include patterns
func (fs *FileSystemScanner) matchesIncludePatterns(relPath string, isDir bool, config *ScanConfig) bool {
	// If no include patterns specified, include everything
	if len(config.IncludePatterns) == 0 {
		return true
	}

	// Always include directories to allow traversal
	if isDir {
		return true
	}

	// Check if file matches any include pattern
	fileName := filepath.Base(relPath)
	for _, pattern := range config.IncludePatterns {
		// Try matching against both relative path and filename
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, fileName); matched {
			return true
		}
	}

	return false
}

// shouldIgnore checks if a path should be ignored based on all rules
func (fs *FileSystemScanner) shouldIgnore(relPath string, isDir bool, config *ScanConfig) bool {
	// First check if file matches include patterns (if any)
	if !fs.matchesIncludePatterns(relPath, isDir, config) {
		return true
	}

	// Use the new ignore engine if available - it properly handles explicit includes/excludes
	if fs.ignoreEngine != nil {
		ignored, _ := fs.ignoreEngine.ShouldIgnore(relPath)
		if ignored {
			return !fs.shouldIncludeIgnored(config)
		}
		// If not ignored by engine, check hidden file exclusion
		if !config.IncludeHidden {
			baseName := filepath.Base(relPath)
			if strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".." {
				return true
			}
		}
		return false
	}
	
	// Fallback to old logic for backward compatibility when using pathMatcher
	isGitignored, isCustomIgnored := fs.getIgnoreStatus(relPath, isDir, config)
	return (isGitignored || isCustomIgnored) && !fs.shouldIncludeIgnored(config)
}

// getIgnoreStatus returns the ignore status for both gitignore and custom rules
func (fs *FileSystemScanner) getIgnoreStatus(relPath string, isDir bool, config *ScanConfig) (bool, bool) {
	if fs.ignoreEngine != nil {
		return fs.getIgnoreStatusWithEngine(relPath, config)
	}
	return fs.getIgnoreStatusLegacy(relPath, isDir, config)
}

func (fs *FileSystemScanner) getIgnoreStatusWithEngine(relPath string, config *ScanConfig) (bool, bool) {
	ignored, reason := fs.ignoreEngine.ShouldIgnore(relPath)

	if ignored {
		return fs.classifyIgnoreReason(reason)
	}

	if fs.isHiddenFile(relPath, config) {
		return false, true
	}

	isGitignored := fs.ignoreEngine.IsGitignored(relPath)
	isCustomIgnored := fs.ignoreEngine.IsCustomIgnored(relPath) || (ignored && reason != ignore.IgnoreReasonGitignore)

	return isGitignored, isCustomIgnored
}

func (fs *FileSystemScanner) classifyIgnoreReason(reason ignore.IgnoreReason) (bool, bool) {
	switch reason {
	case ignore.IgnoreReasonGitignore:
		return true, false
	case ignore.IgnoreReasonBuiltIn, ignore.IgnoreReasonCustom, ignore.IgnoreReasonExplicit:
		return false, true
	}
	return false, false
}

func (fs *FileSystemScanner) isHiddenFile(relPath string, config *ScanConfig) bool {
	if config.IncludeHidden {
		return false
	}
	baseName := filepath.Base(relPath)
	return strings.HasPrefix(baseName, ".") && baseName != "." && baseName != ".."
}

func (fs *FileSystemScanner) getIgnoreStatusLegacy(relPath string, isDir bool, config *ScanConfig) (bool, bool) {
	var isGitignored, isCustomIgnored bool

	if fs.pathMatcher != nil {
		isGitignored = fs.pathMatcher.Matches(relPath, isDir)
	}

	if fs.isHiddenFile(relPath, config) {
		isCustomIgnored = true
	}

	if fs.matchesCustomPatterns(relPath, config.IgnorePatterns) {
		isCustomIgnored = true
	}

	return isGitignored, isCustomIgnored
}

func (fs *FileSystemScanner) matchesCustomPatterns(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(relPath)); matched {
			return true
		}
	}
	return false
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