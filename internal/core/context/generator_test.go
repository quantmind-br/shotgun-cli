package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

func TestNewDefaultContextGenerator(t *testing.T) {
	gen := NewDefaultContextGenerator()
	if gen == nil {
		t.Fatal("NewDefaultContextGenerator() returned nil")
	}
	if gen.treeRenderer == nil {
		t.Error("TreeRenderer is nil")
	}
	if gen.templateRenderer == nil {
		t.Error("TemplateRenderer is nil")
	}
}

func TestGenerateConfig_Validation(t *testing.T) {
	gen := NewDefaultContextGenerator()

	tests := []struct {
		name     string
		config   GenerateConfig
		expected GenerateConfig
	}{
		{
			name:   "default values",
			config: GenerateConfig{},
			expected: GenerateConfig{
				MaxSize:      DefaultMaxSize,
				MaxFiles:     DefaultMaxFiles,
				SkipBinary:   false,
				TemplateVars: map[string]string{},
			},
		},
		{
			name: "custom values preserved",
			config: GenerateConfig{
				MaxSize:      5000,
				MaxFiles:     50,
				SkipBinary:   true,
				TemplateVars: map[string]string{"key": "value"},
			},
			expected: GenerateConfig{
				MaxSize:      5000,
				MaxFiles:     50,
				SkipBinary:   true,
				TemplateVars: map[string]string{"key": "value"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			err := gen.validateConfig(&config)
			if err != nil {
				t.Errorf("validateConfig() error = %v", err)
			}
			if config.MaxSize != tt.expected.MaxSize {
				t.Errorf("MaxSize = %v, want %v", config.MaxSize, tt.expected.MaxSize)
			}
			if config.MaxFiles != tt.expected.MaxFiles {
				t.Errorf("MaxFiles = %v, want %v", config.MaxFiles, tt.expected.MaxFiles)
			}
			if config.SkipBinary != tt.expected.SkipBinary {
				t.Errorf("SkipBinary = %v, want %v", config.SkipBinary, tt.expected.SkipBinary)
			}
			if config.TemplateVars == nil {
				t.Error("TemplateVars is nil")
			}
		})
	}
}

func TestGenerate_WithEmptyRoot(t *testing.T) {
	gen := NewDefaultContextGenerator()
	config := GenerateConfig{}

	root := &scanner.FileNode{
		Name:   "test",
		Path:   "/test",
		IsDir:  true,
		Size:   0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	result, err := gen.Generate(root, config)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	if result == "" {
		t.Error("Generate() returned empty result")
	}

	if !strings.Contains(result, "# Project Context") {
		t.Error("Generated context missing expected header")
	}
}

func TestGenerate_WithFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "context_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	testContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a simple FileNode structure
	root := &scanner.FileNode{
		Name:     "test",
		Path:     tmpDir,
		IsDir:    true,
		Size:     0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	fileNode := &scanner.FileNode{
		Name:   "test.go",
		Path:   testFile,
		IsDir:  false,
		Size:   int64(len(testContent)),
		Children: make(map[string]*scanner.FileNode),
	}
	fileNode.SetSelected(true)
	root.Children["test.go"] = fileNode

	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "Test task",
		},
	}

	result, err := gen.Generate(root, config)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	expectedStrings := []string{
		"# Project Context",
		"**Task:** Test task",
		"## File Structure",
		"test.go",
		"## File Contents",
		"### test.go (go)",
		"package main",
		"fmt.Println",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Generated context missing expected string: %s", expected)
		}
	}
}

func TestGenerate_SizeLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "context_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large file
	largeContent := strings.Repeat("x", 1000)
	testFile := filepath.Join(tmpDir, "large.txt")
	err = os.WriteFile(testFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     tmpDir,
		IsDir:    true,
		Size:     0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	fileNode := &scanner.FileNode{
		Name:   "large.txt",
		Path:   testFile,
		IsDir:  false,
		Size:   int64(len(largeContent)),
		Children: make(map[string]*scanner.FileNode),
	}
	fileNode.SetSelected(true)
	root.Children["large.txt"] = fileNode

	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		MaxSize: 500, // Smaller than the content
	}

	_, err = gen.Generate(root, config)
	if err == nil {
		t.Error("Expected error for size limit exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "size") && !strings.Contains(err.Error(), "limit") {
		t.Errorf("Expected size limit error, got: %v", err)
	}
}

func TestGenerate_SkipBinary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "context_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a binary-like file with null bytes
	binaryContent := []byte{0x00, 0x01, 0x02, 0x03}
	binaryFile := filepath.Join(tmpDir, "binary.dat")
	err = os.WriteFile(binaryFile, binaryContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     tmpDir,
		IsDir:    true,
		Size:     0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	fileNode := &scanner.FileNode{
		Name:   "binary.dat",
		Path:   binaryFile,
		IsDir:  false,
		Size:   int64(len(binaryContent)),
		Children: make(map[string]*scanner.FileNode),
	}
	fileNode.SetSelected(true)
	root.Children["binary.dat"] = fileNode

	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		SkipBinary: true,
	}

	result, err := gen.Generate(root, config)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	if strings.Contains(result, "binary.dat") && strings.Contains(result, "## File Contents") {
		t.Error("Binary file content should be skipped when SkipBinary is true")
	}
}

func TestGenerateWithProgress(t *testing.T) {
	gen := NewDefaultContextGenerator()
	config := GenerateConfig{}

	root := &scanner.FileNode{
		Name:   "test",
		Path:   "/test",
		IsDir:  true,
		Size:   0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	var progressMessages []string
	progress := func(msg string) {
		progressMessages = append(progressMessages, msg)
	}

	_, err := gen.GenerateWithProgress(root, config, progress)
	if err != nil {
		t.Errorf("GenerateWithProgress() error = %v", err)
		return
	}

	expectedMessages := []string{
		"Generating file structure...",
		"Collecting file contents...",
		"Rendering template...",
	}

	if len(progressMessages) != len(expectedMessages) {
		t.Errorf("Expected %d progress messages, got %d", len(expectedMessages), len(progressMessages))
	}

	for i, expected := range expectedMessages {
		if i < len(progressMessages) && progressMessages[i] != expected {
			t.Errorf("Progress message %d: expected %q, got %q", i, expected, progressMessages[i])
		}
	}
}

func TestGenerate_CustomTemplate(t *testing.T) {
	gen := NewDefaultContextGenerator()

	root := &scanner.FileNode{
		Name:   "test",
		Path:   "/test",
		IsDir:  true,
		Size:   0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	customTemplate := "Custom template: {{.Task}} - {{.CurrentDate}}"
	config := GenerateConfig{
		Template: customTemplate,
		TemplateVars: map[string]string{
			"TASK": "Custom task",
		},
	}

	result, err := gen.Generate(root, config)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	if !strings.Contains(result, "Custom template: Custom task") {
		t.Error("Custom template not applied correctly")
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewDefaultContextGenerator()

	root := &scanner.FileNode{
		Name:   "test",
		Path:   "/test",
		IsDir:  true,
		Size:   0,
		Children: make(map[string]*scanner.FileNode),
	}
	root.SetSelected(true)

	config := GenerateConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(root, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}