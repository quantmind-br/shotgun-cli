package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

type FileNode struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	IsDir      bool        `json:"is_dir"`
	IsSelected bool        `json:"is_selected"`
	Children   []*FileNode `json:"children,omitempty"`
	Size       int64       `json:"size,omitempty"`
}

type Scanner struct {
	root       string
	gitignore  *ignore.GitIgnore
	exclusions []string
}

func NewScanner(root string) (*Scanner, error) {
	scanner := &Scanner{
		root:       root,
		exclusions: getDefaultExclusions(),
	}

	if err := scanner.loadGitignore(); err != nil {
		return nil, fmt.Errorf("failed to load gitignore: %w", err)
	}

	return scanner, nil
}

func (s *Scanner) loadGitignore() error {
	gitignorePath := filepath.Join(s.root, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		s.gitignore = ignore.CompileIgnoreLines()
		return nil
	}

	gitIgnore, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("failed to compile gitignore: %w", err)
	}

	s.gitignore = gitIgnore
	return nil
}

func (s *Scanner) ScanDirectory() (*FileNode, error) {
	rootNode := &FileNode{
		Name:       filepath.Base(s.root),
		Path:       ".",
		IsDir:      true,
		IsSelected: false,
		Children:   []*FileNode{},
	}

	err := s.scanRecursive(s.root, rootNode)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return rootNode, nil
}

func (s *Scanner) scanRecursive(currentPath string, node *FileNode) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", currentPath, err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(currentPath, entry.Name())
		relativePath, err := filepath.Rel(s.root, entryPath)
		if err != nil {
			continue
		}

		if s.shouldExclude(relativePath, entry.IsDir()) {
			continue
		}

		childNode := &FileNode{
			Name:       entry.Name(),
			Path:       relativePath,
			IsDir:      entry.IsDir(),
			IsSelected: false,
		}

		if !entry.IsDir() {
			if info, err := entry.Info(); err == nil {
				childNode.Size = info.Size()
			}
		}

		if entry.IsDir() {
			childNode.Children = []*FileNode{}
			if err := s.scanRecursive(entryPath, childNode); err != nil {
				continue
			}
		}

		node.Children = append(node.Children, childNode)
	}

	return nil
}

func (s *Scanner) shouldExclude(relativePath string, isDir bool) bool {
	if s.gitignore.MatchesPath(relativePath) {
		return true
	}

	for _, exclusion := range s.exclusions {
		if matched, _ := filepath.Match(exclusion, relativePath); matched {
			return true
		}
		if matched, _ := filepath.Match(exclusion, filepath.Base(relativePath)); matched {
			return true
		}
	}

	return false
}

func (s *Scanner) GetSelectedFiles(root *FileNode) []string {
	var selected []string
	s.collectSelected(root, &selected)
	return selected
}

func (s *Scanner) collectSelected(node *FileNode, selected *[]string) {
	if node.IsSelected && !node.IsDir {
		*selected = append(*selected, node.Path)
	}

	for _, child := range node.Children {
		s.collectSelected(child, selected)
	}
}

func (s *Scanner) ReadFileContent(relativePath string) (string, error) {
	fullPath := filepath.Join(s.root, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", relativePath, err)
	}
	return string(content), nil
}

func (s *Scanner) GenerateFileStructure(root *FileNode) string {
	var sb strings.Builder
	
	sb.WriteString("# File Structure\n\n")
	
	sb.WriteString("## Directory Tree\n")
	sb.WriteString("```\n")
	s.writeTree(root, &sb, "")
	sb.WriteString("```\n\n")
	
	sb.WriteString("## File Contents\n\n")
	selectedFiles := s.GetSelectedFiles(root)
	
	for _, filePath := range selectedFiles {
		content, err := s.ReadFileContent(filePath)
		if err != nil {
			sb.WriteString(fmt.Sprintf("<file path=\"%s\">\nError reading file: %s\n</file>\n\n", filePath, err.Error()))
			continue
		}
		
		sb.WriteString(fmt.Sprintf("<file path=\"%s\">\n%s\n</file>\n\n", filePath, content))
	}
	
	return sb.String()
}

func (s *Scanner) writeTree(node *FileNode, sb *strings.Builder, prefix string) {
	if node.Path != "." {
		marker := "├── "
		if node.IsSelected {
			marker = "✓ " + marker
		} else {
			marker = "  " + marker
		}
		
		sb.WriteString(fmt.Sprintf("%s%s%s", prefix, marker, node.Name))
		if !node.IsDir && node.Size > 0 {
			sb.WriteString(fmt.Sprintf(" (%d bytes)", node.Size))
		}
		sb.WriteString("\n")
	}

	for i, child := range node.Children {
		childPrefix := prefix
		if node.Path != "." {
			if i == len(node.Children)-1 {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
		}
		s.writeTree(child, sb, childPrefix)
	}
}

func getDefaultExclusions() []string {
	return []string{
		".git",
		".gitignore",
		"node_modules",
		"*.log",
		"*.tmp",
		"*.cache",
		".DS_Store",
		"Thumbs.db",
		".vscode",
		".idea",
		"*.exe",
		"*.dll",
		"*.so",
		"*.dylib",
		"__pycache__",
		"*.pyc",
		"*.pyo",
		"*.pyd",
		".env",
		".env.local",
		".env.*.local",
		"dist",
		"build",
		"target",
		"*.zip",
		"*.tar.gz",
		"*.rar",
	}
}