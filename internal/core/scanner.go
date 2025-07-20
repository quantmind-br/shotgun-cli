package core

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	gitignore "github.com/sabhiram/go-gitignore"
)

const (
	defaultIgnorePatterns = `.git
.bundle

*.jpg
*.jpeg
*.png
*.gif
*.bmp
*.tiff
*.webp
*.svg
*.ico
*.icns

*.mp3
*.wav
*.ogg
*.aac
*.flac

*.mp4
*.avi
*.mov
*.wmv
*.mkv
*.webm
*.flv

node_modules/
vendor/
.idea/
.vscode/
.vs/
__pycache__/

tmp/
cache/
*.tmp
*.temp
*.bak
*.swp
*.swo

*.zip
*.rar
*.7z
*.tar
*.gz
*.tgz
*.bz2

*.sqlite
*.sqlite3
*.db
*.mdb
*.sql
*.dump

*.exe
*.dll
*.so
*.dylib
*.bin
*.iso
*.dmg`
)

// SetDefaultIgnorePatterns sets up the default ignore patterns
func (ds *DirectoryScanner) SetDefaultIgnorePatterns() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.defaultIgnore = gitignore.CompileIgnoreLines(strings.Split(defaultIgnorePatterns, "\n")...)
	return nil
}

// SetGitIgnore sets the git ignore patterns from a .gitignore file
func (ds *DirectoryScanner) SetGitIgnore(gitignorePath string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		ds.gitIgnore = nil
		return nil
	}

	var err error
	ds.gitIgnore, err = gitignore.CompileIgnoreFile(gitignorePath)
	return err
}

// SetCustomIgnore sets custom ignore patterns
func (ds *DirectoryScanner) SetCustomIgnore(patterns []string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.customIgnore = gitignore.CompileIgnoreLines(patterns...)
	return nil
}

// ScanDirectory scans a directory and returns a file tree
func (ds *DirectoryScanner) ScanDirectory(ctx context.Context, rootPath string) (*FileNode, error) {
	// Set up default patterns if not already set
	if ds.defaultIgnore == nil {
		if err := ds.SetDefaultIgnorePatterns(); err != nil {
			return nil, &ShotgunError{
				Operation: "compile default ignore patterns",
				Path:      rootPath,
				Err:       err,
			}
		}
	}

	// Try to load .gitignore from the root directory
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	if err := ds.SetGitIgnore(gitignorePath); err != nil {
		return nil, &ShotgunError{
			Operation: "load .gitignore",
			Path:      gitignorePath,
			Err:       err,
		}
	}

	// Build the tree recursively
	children, err := ds.buildTreeRecursive(ctx, rootPath, rootPath, 0)
	if err != nil {
		return nil, err
	}

	// Create root node
	rootNode := &FileNode{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		RelPath:  "",
		IsDir:    true,
		Children: children,
	}

	return rootNode, nil
}

// buildTreeRecursive builds the file tree recursively (ported from app.go)
func (ds *DirectoryScanner) buildTreeRecursive(ctx context.Context, currentPath, rootPath string, depth int) ([]*FileNode, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, &ShotgunError{
			Operation: "read directory",
			Path:      currentPath,
			Err:       err,
		}
	}

	var nodes []*FileNode
	for _, entry := range entries {
		nodePath := filepath.Join(currentPath, entry.Name())
		relPath, _ := filepath.Rel(rootPath, nodePath)

		// Check ignore patterns
		isGitignored := false
		isCustomIgnored := false
		isDefaultIgnored := false

		pathToMatch := relPath
		if entry.IsDir() {
			if !strings.HasSuffix(pathToMatch, string(os.PathSeparator)) {
				pathToMatch += string(os.PathSeparator)
			}
		}

		ds.mu.RLock()
		if ds.gitIgnore != nil {
			isGitignored = ds.gitIgnore.MatchesPath(pathToMatch)
		}
		if ds.customIgnore != nil {
			isCustomIgnored = ds.customIgnore.MatchesPath(pathToMatch)
		}
		if ds.defaultIgnore != nil {
			isDefaultIgnored = ds.defaultIgnore.MatchesPath(pathToMatch)
		}
		ds.mu.RUnlock()

		// Send progress update
		select {
		case ds.progressChan <- ProgressUpdate{
			CurrentFile: relPath,
			Phase:       "scanning",
		}:
		default:
		}

		node := &FileNode{
			Name:            entry.Name(),
			Path:            nodePath,
			RelPath:         relPath,
			IsDir:           entry.IsDir(),
			IsGitignored:    isGitignored,
			IsCustomIgnored: isCustomIgnored || isDefaultIgnored,
		}

		if entry.IsDir() {
			// Only recurse if not ignored
			if !isGitignored && !isCustomIgnored && !isDefaultIgnored {
				children, err := ds.buildTreeRecursive(ctx, nodePath, rootPath, depth+1)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil, err
					}
					// Log warning but continue (graceful degradation)
					// In a real implementation, you might want to use a logger here
				} else {
					node.Children = children
				}
			}
		}

		nodes = append(nodes, node)
	}

	// Sort nodes: directories first, then files, then alphabetically
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir && !nodes[j].IsDir {
			return true
		}
		if !nodes[i].IsDir && nodes[j].IsDir {
			return false
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})

	return nodes, nil
}

// GetProgressChannel returns the progress channel for monitoring scan progress
func (ds *DirectoryScanner) GetProgressChannel() <-chan ProgressUpdate {
	return ds.progressChan
}

// CountFiles recursively counts files in a directory tree
func (ds *DirectoryScanner) CountFiles(ctx context.Context, rootPath string) (int64, error) {
	var count int64

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip problematic paths
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !d.IsDir() {
			count++
		}

		return nil
	})

	return count, err
}

// GetIncludedFiles returns a list of files that should be included based on selection state
func GetIncludedFiles(root *FileNode, selection *SelectionState) []string {
	var result []string
	walkFileTree(root, selection, &result)
	return result
}

// walkFileTree recursively walks the file tree and collects included files
func walkFileTree(node *FileNode, selection *SelectionState, result *[]string) {
	if node == nil {
		return
	}

	if !node.IsDir {
		// It's a file, check if it should be included
		if !node.IsGitignored && !node.IsCustomIgnored && selection.IsFileIncluded(node.RelPath) {
			*result = append(*result, node.Path)
		}
	} else {
		// It's a directory, recurse through children
		for _, child := range node.Children {
			walkFileTree(child, selection, result)
		}
	}
}
