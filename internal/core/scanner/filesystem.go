package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
)

// FileSystemScanner implements the Scanner interface for local file systems
type FileSystemScanner struct {
	ignoreEngine *ignore.GitIgnore
}

// NewFileSystemScanner creates a new file system scanner
func NewFileSystemScanner() *FileSystemScanner {
	return &FileSystemScanner{}
}

// NewFileSystemScannerWithIgnore creates a new file system scanner with ignore rules
func NewFileSystemScannerWithIgnore(ignoreRules []string) (*FileSystemScanner, error) {
	if len(ignoreRules) == 0 {
		return NewFileSystemScanner(), nil
	}

	ignoreEngine := ignore.CompileIgnoreLines(ignoreRules...)

	return &FileSystemScanner{
		ignoreEngine: ignoreEngine,
	}, nil
}

// LoadGitIgnore loads .gitignore rules from the specified root directory
func (fs *FileSystemScanner) LoadGitIgnore(rootPath string) error {
	gitignorePath := filepath.Join(rootPath, ".gitignore")

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil // No .gitignore file, continue without it
	}

	ignoreEngine, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("failed to load .gitignore from %s: %w", gitignorePath, err)
	}

	fs.ignoreEngine = ignoreEngine
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

	// Load .gitignore if not already loaded
	if fs.ignoreEngine == nil {
		if err := fs.LoadGitIgnore(rootPath); err != nil {
			return nil, fmt.Errorf("failed to load ignore rules: %w", err)
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
	root, err := fs.walkAndBuild(rootPath, config, progress, total)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	// Sort children for consistent ordering
	fs.sortChildren(root)

	if progress != nil {
		progress <- Progress{
			Current:   total,
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

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories/files
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check limits
		if config.MaxFiles > 0 && count >= config.MaxFiles {
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

		count++
		return nil
	})

	return count, err
}

// walkAndBuild builds the file tree with progress reporting
func (fs *FileSystemScanner) walkAndBuild(rootPath string, config *ScanConfig, progress chan<- Progress, total int64) (*FileNode, error) {
	var current int64
	var processedFiles int64

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
	dirNodes["."] = root

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories/files but continue scanning
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check limits
		if config.MaxFiles > 0 && processedFiles >= config.MaxFiles {
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

		// If this is a directory, add it to the directory map
		if d.IsDir() {
			dirNodes[relPath] = node
		}

		current++
		processedFiles++

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

	return root, err
}

// findParentNode finds the parent directory node for a given relative path
func (fs *FileSystemScanner) findParentNode(relPath string, dirNodes map[string]*FileNode) *FileNode {
	parentPath := filepath.Dir(relPath)
	if parentPath == "." {
		return dirNodes["."]
	}

	// Normalize path separators for consistency
	parentPath = filepath.ToSlash(parentPath)
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

	return dirNodes["."] // Return root if no parent found
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
	var isGitignored, isCustomIgnored bool

	// Check .gitignore rules
	if fs.ignoreEngine != nil {
		isGitignored = fs.ignoreEngine.MatchesPath(relPath)
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

// SkipDir is a sentinel error used with filepath.WalkDir to skip directories
var SkipDir = filepath.SkipDir