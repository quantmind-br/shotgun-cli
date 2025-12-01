package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTextFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"empty content", "", true},
		{"simple text", "Hello, World!", true},
		{"multiline text", "line1\nline2\nline3", true},
		{"utf8 text", "Hello, \xe4\xb8\x96\xe7\x95\x8c", true},
		{"binary with null", "Hello\x00World", false},
		{"binary at start", "\x00\x01\x02\x03", false},
		{"valid utf8 special chars", "Line with special chars: \xc3\xa9\xc3\xa0\xc3\xb9", true},
		{"code sample", "func main() {\n\tfmt.Println(\"Hello\")\n}", true},
		{"json content", `{"key": "value", "number": 42}`, true},
		{"xml content", `<?xml version="1.0"?><root><item>test</item></root>`, true},
		{"markdown", "# Title\n\nThis is **bold** text", true},
		{"long text", strings.Repeat("A", 2000), true}, // 2000 characters of text
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTextFile(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// By extension
		{"go file", "main.go", "go"},
		{"javascript", "app.js", "javascript"},
		{"jsx", "component.jsx", "javascript"},
		{"typescript", "service.ts", "typescript"},
		{"tsx", "component.tsx", "typescript"},
		{"python", "script.py", "python"},
		{"java", "Main.java", "java"},
		{"c file", "main.c", "c"},
		{"cpp file", "main.cpp", "cpp"},
		{"csharp", "Program.cs", "csharp"},
		{"php", "index.php", "php"},
		{"ruby", "app.rb", "ruby"},
		{"rust", "main.rs", "rust"},
		{"bash", "script.sh", "bash"},
		{"zsh", "config.zsh", "bash"},
		{"powershell", "script.ps1", "powershell"},
		{"sql", "query.sql", "sql"},
		{"html", "index.html", "html"},
		{"htm", "page.htm", "html"},
		{"css", "styles.css", "css"},
		{"scss", "styles.scss", "scss"},
		{"sass", "styles.sass", "scss"},
		{"xml", "config.xml", "xml"},
		{"json", "package.json", "json"},
		{"yaml", "config.yaml", "yaml"},
		{"yml", "config.yml", "yaml"},
		{"toml", "cargo.toml", "toml"},
		{"ini", "config.ini", "ini"},
		{"markdown", "README.md", "markdown"},
		{"markdown full", "CHANGELOG.markdown", "markdown"},
		{"latex", "document.tex", "latex"},
		{"r language", "analysis.r", "r"},
		{"matlab", "script.m", "matlab"},
		{"swift", "App.swift", "swift"},
		{"kotlin", "Main.kt", "kotlin"},
		{"scala", "App.scala", "scala"},
		{"clojure", "core.clj", "clojure"},
		{"clojurescript", "app.cljs", "clojure"},
		{"haskell", "Main.hs", "haskell"},
		{"elm", "Main.elm", "elm"},
		{"dart", "main.dart", "dart"},
		{"lua", "script.lua", "lua"},
		{"vim", "config.vim", "vim"},
		{"dockerfile ext", "app.dockerfile", "dockerfile"},

		// By basename
		{"Dockerfile", "Dockerfile", "dockerfile"},
		{"Makefile", "Makefile", "makefile"},
		{"Rakefile", "Rakefile", "ruby"},
		{"Gemfile", "Gemfile", "ruby"},
		{"Gemfile.lock", "Gemfile.lock", "ruby"},
		{"package.json", "package.json", "json"},
		{"package-lock.json", "package-lock.json", "json"},
		{"composer.json", "composer.json", "json"},
		{"composer.lock", "composer.lock", "json"},
		{"Cargo.toml", "Cargo.toml", "toml"},
		{"Cargo.lock", "Cargo.lock", "toml"},
		{"go.mod", "go.mod", "go"},
		{"go.sum", "go.sum", "go"},
		{"requirements.txt", "requirements.txt", "python"},
		{"setup.py", "setup.py", "python"},
		{"setup.cfg", "setup.cfg", "python"},

		// Unknown extensions
		{"unknown extension", "file.xyz", "text"},
		{"no extension", "LICENSE", "text"},
		{"dotfile", ".gitignore", "text"},

		// Case insensitivity
		{"uppercase extension", "FILE.GO", "go"},
		{"mixed case", "File.Js", "javascript"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguage(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectLanguageByBasename(t *testing.T) {
	tests := []struct {
		name     string
		basename string
		expected string
	}{
		{"dockerfile", "dockerfile", "dockerfile"},
		{"makefile", "makefile", "makefile"},
		{"rakefile", "rakefile", "ruby"},
		{"gemfile", "gemfile", "ruby"},
		{"gemfile.lock", "gemfile.lock", "ruby"},
		{"package.json", "package.json", "json"},
		{"package-lock.json", "package-lock.json", "json"},
		{"composer.json", "composer.json", "json"},
		{"composer.lock", "composer.lock", "json"},
		{"cargo.toml", "cargo.toml", "toml"},
		{"cargo.lock", "cargo.lock", "toml"},
		{"go.mod", "go.mod", "go"},
		{"go.sum", "go.sum", "go"},
		{"requirements.txt", "requirements.txt", "python"},
		{"setup.py", "setup.py", "python"},
		{"setup.cfg", "setup.cfg", "python"},
		{"unknown", "randomfile", ""},
		{"partial match", "docker", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguageByBasename(tt.basename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectLanguageByExtension(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{"go", ".go", "go"},
		{"js", ".js", "javascript"},
		{"mjs", ".mjs", "javascript"},
		{"ts", ".ts", "typescript"},
		{"py", ".py", "python"},
		{"pyw", ".pyw", "python"},
		{"unknown", ".xyz", "text"},
		{"empty", "", "text"},
		{"cpp variants", ".cc", "cpp"},
		{"cpp cxx", ".cxx", "cpp"},
		{"hpp", ".hpp", "cpp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectLanguageByExtension(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadFileContent(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("read existing file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test.txt")
		expectedContent := "Hello, World!"
		err := os.WriteFile(filePath, []byte(expectedContent), 0644)
		require.NoError(t, err)

		content, err := readFileContent(filePath)
		require.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := readFileContent(filepath.Join(tmpDir, "nonexistent.txt"))
		assert.Error(t, err)
	})

	t.Run("read empty file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(filePath, []byte{}, 0644)
		require.NoError(t, err)

		content, err := readFileContent(filePath)
		require.NoError(t, err)
		assert.Equal(t, "", content)
	})
}

func TestPeekFileHeader(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("peek small file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "small.txt")
		content := "Small content"
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)

		header, err := peekFileHeader(filePath, 1024)
		require.NoError(t, err)
		assert.Equal(t, content, string(header))
	})

	t.Run("peek large file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "large.txt")
		content := make([]byte, 2048)
		for i := range content {
			content[i] = byte('A' + (i % 26))
		}
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		header, err := peekFileHeader(filePath, 1024)
		require.NoError(t, err)
		assert.Len(t, header, 1024)
	})

	t.Run("peek non-existent file", func(t *testing.T) {
		_, err := peekFileHeader(filepath.Join(tmpDir, "nonexistent.txt"), 1024)
		assert.Error(t, err)
	})

	t.Run("peek empty file", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "empty.txt")
		err := os.WriteFile(filePath, []byte{}, 0644)
		require.NoError(t, err)

		header, err := peekFileHeader(filePath, 1024)
		require.NoError(t, err)
		assert.Len(t, header, 0)
	})
}

