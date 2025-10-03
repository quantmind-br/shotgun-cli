package context

import (
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

func collectFileContents(root *scanner.FileNode, config GenerateConfig) ([]FileContent, error) {
	var files []FileContent
	var totalSize int64
	fileCount := 0

	err := walkSelectedNodes(root, func(node *scanner.FileNode) error {
		if node.IsDir || node.IsIgnored() {
			return nil
		}

		if !node.Selected {
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
			header, err := peekFileHeader(node.Path, 1024)
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
			return fmt.Errorf("cumulative content size exceeds total size limit: %d + %d > %d", totalSize, len(content), config.MaxTotalSize)
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

func peekFileHeader(path string, n int) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	header := make([]byte, n)
	bytesRead, err := file.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return header[:bytesRead], nil
}

func readFileContent(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", err
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

	switch ext {
	case ".go":
		return "go"
	case ".js", ".jsx", ".mjs":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py", ".pyw":
		return "python"
	case ".java":
		return "java"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "cpp"
	case ".h", ".hpp", ".hh", ".hxx", ".h++":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".rb":
		return langRuby
	case ".rs":
		return "rust"
	case ".sh", ".bash", ".zsh":
		return "bash"
	case ".ps1":
		return "powershell"
	case ".sql":
		return "sql"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".scss", ".sass":
		return "scss"
	case ".xml":
		return "xml"
	case ".json":
		return langJSON
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".ini":
		return "ini"
	case ".md", ".markdown":
		return "markdown"
	case ".tex":
		return "latex"
	case ".r":
		return "r"
	case ".m":
		return "matlab"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".clj", ".cljs":
		return "clojure"
	case ".hs":
		return "haskell"
	case ".elm":
		return "elm"
	case ".dart":
		return "dart"
	case ".lua":
		return "lua"
	case ".vim":
		return "vim"
	case ".dockerfile":
		return langDockerfile
	default:
		return "text"
	}
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