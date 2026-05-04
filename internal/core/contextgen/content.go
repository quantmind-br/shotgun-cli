package contextgen

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

const (
	langDockerfile = "dockerfile"
	langJSON       = "json"
	langRuby       = "ruby"
)

type FileContent struct {
	Path     string `json:"path"`
	RelPath  string `json:"relPath"`
	Language string `json:"language"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
}

func collectFileContents(
	root *scanner.FileNode, selections map[string]bool, config GenerateConfig,
) ([]FileContent, error) {
	var files []FileContent
	var totalSize int64
	fileCount := 0

	err := walkSelectedNodes(root, func(node *scanner.FileNode) error {
		if node.IsDir || node.IsIgnored() {
			return nil
		}

		// Check selection against the map
		// If selections map is nil, we assume all non-ignored files are selected
		if selections != nil && !selections[node.Path] {
			return nil
		}

		if fileCount >= config.MaxFiles {
			return fmt.Errorf("maximum file count exceeded: %d", config.MaxFiles)
		}

		if shouldSkipFile(node, config) {
			return nil
		}

		// First peek at the file header to check if it's binary before reading full content
		if config.SkipBinary {
			header, err := peekFileHeader(node.Path)
			if err != nil {
				return fmt.Errorf("failed to peek file header %s: %w", node.Path, err)
			}
			if !isTextFile(string(header)) {
				return nil // Skip binary file without reading full content
			}
		}

		content, err := readFileContent(node.Path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", node.Path, err)
		}

		if totalSize+int64(len(content)) > config.MaxTotalSize {
			return fmt.Errorf(
				"cumulative content size exceeds total size limit: %d + %d > %d",
				totalSize, len(content), config.MaxTotalSize,
			)
		}

		relPath, err := filepath.Rel(root.Path, node.Path)
		if err != nil {
			relPath = node.Path
		}

		fileContent := FileContent{
			Path:     node.Path,
			RelPath:  relPath,
			Language: detectLanguage(node.Name),
			Content:  content,
			Size:     int64(len(content)),
		}

		files = append(files, fileContent)
		totalSize += fileContent.Size
		fileCount++

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func walkSelectedNodes(node *scanner.FileNode, fn func(*scanner.FileNode) error) error {
	if err := fn(node); err != nil {
		return err
	}

	if node.IsDir {
		for _, child := range node.Children {
			if err := walkSelectedNodes(child, fn); err != nil {
				return err
			}
		}
	}

	return nil
}

func peekFileHeader(path string) ([]byte, error) {
	file, err := os.Open(path) //nolint:gosec // path is validated by caller
	if err != nil {
		return nil, fmt.Errorf("failed to open file for header peek: %w", err)
	}
	defer func() { _ = file.Close() }()

	const headerSize = 1024
	header := make([]byte, headerSize)
	bytesRead, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	return header[:bytesRead], nil
}

func readFileContent(path string) (string, error) {
	file, err := os.Open(path) //nolint:gosec // path is validated by caller
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return string(content), nil
}

func isTextFile(content string) bool {
	if len(content) == 0 {
		return true
	}

	checkSize := len(content)
	if checkSize > 1024 {
		checkSize = 1024
	}

	sample := content[:checkSize]

	for _, b := range []byte(sample) {
		if b == 0 {
			return false
		}
	}

	return utf8.ValidString(sample)
}

func detectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	base := strings.ToLower(filepath.Base(filename))

	// Try detecting by basename first
	if lang := detectLanguageByBasename(base); lang != "" {
		return lang
	}

	// Then try by extension
	return detectLanguageByExtension(ext)
}

func detectLanguageByBasename(base string) string {
	switch base {
	case "dockerfile":
		return langDockerfile
	case "makefile":
		return "makefile"
	case "rakefile":
		return langRuby
	case "gemfile", "gemfile.lock":
		return langRuby
	case "package.json", "package-lock.json":
		return langJSON
	case "composer.json", "composer.lock":
		return langJSON
	case "cargo.toml", "cargo.lock":
		return "toml"
	case "go.mod", "go.sum":
		return "go"
	case "requirements.txt", "setup.py", "setup.cfg":
		return "python"
	}

	return ""
}

var extensionToLanguage = map[string]string{
	".go":         "go",
	".js":         "javascript",
	".jsx":        "javascript",
	".mjs":        "javascript",
	".ts":         "typescript",
	".tsx":        "typescript",
	".py":         "python",
	".pyw":        "python",
	".java":       "java",
	".c":          "c",
	".cpp":        "cpp",
	".cc":         "cpp",
	".cxx":        "cpp",
	".c++":        "cpp",
	".h":          "cpp",
	".hpp":        "cpp",
	".hh":         "cpp",
	".hxx":        "cpp",
	".h++":        "cpp",
	".cs":         "csharp",
	".php":        "php",
	".rb":         langRuby,
	".rs":         "rust",
	".sh":         "bash",
	".bash":       "bash",
	".zsh":        "bash",
	".ps1":        "powershell",
	".sql":        "sql",
	".html":       "html",
	".htm":        "html",
	".css":        "css",
	".scss":       "scss",
	".sass":       "scss",
	".xml":        "xml",
	".json":       langJSON,
	".yaml":       "yaml",
	".yml":        "yaml",
	".toml":       "toml",
	".ini":        "ini",
	".md":         "markdown",
	".markdown":   "markdown",
	".tex":        "latex",
	".r":          "r",
	".m":          "matlab",
	".swift":      "swift",
	".kt":         "kotlin",
	".scala":      "scala",
	".clj":        "clojure",
	".cljs":       "clojure",
	".hs":         "haskell",
	".elm":        "elm",
	".dart":       "dart",
	".lua":        "lua",
	".vim":        "vim",
	".dockerfile": langDockerfile,
}

func detectLanguageByExtension(ext string) string {
	if lang, ok := extensionToLanguage[ext]; ok {
		return lang
	}

	return "text"
}

func shouldSkipFile(node *scanner.FileNode, config GenerateConfig) bool {
	if node.IsDir {
		return true
	}

	if node.Size > config.MaxFileSize {
		return true
	}

	return false
}

// renderFileContentBlocks renders file contents in XML-like format
func renderFileContentBlocks(files []FileContent) string {
	var builder strings.Builder

	for _, file := range files {
		builder.WriteString(fmt.Sprintf("<file path=\"%s\">\n", file.RelPath))
		builder.WriteString(file.Content)
		// Ensure content ends with newline before closing tag
		if len(file.Content) > 0 && !strings.HasSuffix(file.Content, "\n") {
			builder.WriteString("\n")
		}
		builder.WriteString("</file>\n")
	}

	return builder.String()
}
