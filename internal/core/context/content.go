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

		if !node.IsSelected() {
			return nil
		}

		if fileCount >= config.MaxFiles {
			return fmt.Errorf("maximum file count exceeded: %d", config.MaxFiles)
		}

		if shouldSkipFile(node, config) {
			return nil
		}

		content, err := readFileContent(node.Path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", node.Path, err)
		}

		if config.SkipBinary && !isTextFile(content) {
			return nil
		}

		if totalSize+int64(len(content)) > config.MaxSize {
			return fmt.Errorf("cumulative content size exceeds limit: %d + %d > %d", totalSize, len(content), config.MaxSize)
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
		return "dockerfile"
	case "makefile":
		return "makefile"
	case "rakefile":
		return "ruby"
	case "gemfile", "gemfile.lock":
		return "ruby"
	case "package.json", "package-lock.json":
		return "json"
	case "composer.json", "composer.lock":
		return "json"
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
		return "ruby"
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
		return "json"
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
		return "dockerfile"
	default:
		return "text"
	}
}

func shouldSkipFile(node *scanner.FileNode, config GenerateConfig) bool {
	if node.IsDir {
		return true
	}

	if node.Size > config.MaxSize {
		return true
	}

	return false
}