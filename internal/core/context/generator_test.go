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
				MaxFileSize:  DefaultMaxSize,
				MaxTotalSize: DefaultMaxSize,
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
				MaxFileSize:  5000,
				MaxTotalSize: 5000,
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
			if config.MaxFileSize != tt.expected.MaxFileSize {
				t.Errorf("MaxFileSize = %v, want %v", config.MaxFileSize, tt.expected.MaxFileSize)
			}
			if config.MaxTotalSize != tt.expected.MaxTotalSize {
				t.Errorf("MaxTotalSize = %v, want %v", config.MaxTotalSize, tt.expected.MaxTotalSize)
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
	config := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "Test with empty root",
		},
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     "/test",
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

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
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	fileNode := &scanner.FileNode{
		Name:   "test.go",
		Path:   testFile,
		IsDir:  false,
		Size:   int64(len(testContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	root.Children = append(root.Children, fileNode)

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
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	fileNode := &scanner.FileNode{
		Name:   "large.txt",
		Path:   testFile,
		IsDir:  false,
		Size:   int64(len(largeContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	root.Children = append(root.Children, fileNode)

	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		MaxTotalSize: 500, // Smaller than the content
		TemplateVars: map[string]string{
			"TASK": "Test size limit",
		},
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
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	fileNode := &scanner.FileNode{
		Name:   "binary.dat",
		Path:   binaryFile,
		IsDir:  false,
		Size:   int64(len(binaryContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	root.Children = append(root.Children, fileNode)

	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		SkipBinary: true,
		TemplateVars: map[string]string{
			"TASK": "Test skip binary",
		},
	}

	result, err := gen.Generate(root, config)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Binary file should not be in the File Contents section
	contentsStart := strings.Index(result, "## File Contents")
	if contentsStart != -1 {
		contentsSection := result[contentsStart:]
		if strings.Contains(contentsSection, "### binary.dat") {
			t.Error("Binary file content should be skipped when SkipBinary is true")
		}
	}
}

func TestGenerateWithProgress(t *testing.T) {
	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "Test with progress",
		},
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     "/test",
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

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
		"Context generation completed",
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
		Name:     "test",
		Path:     "/test",
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

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
		Name:     "test",
		Path:     "/test",
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	config := GenerateConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(root, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestGenerate_MissingRequiredTemplateVars(t *testing.T) {
	gen := NewDefaultContextGenerator()

	root := &scanner.FileNode{
		Name:     "test",
		Path:     "/test",
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	// Test case 1: Missing TASK variable entirely
	config1 := GenerateConfig{
		TemplateVars: map[string]string{}, // No TASK variable
	}

	_, err := gen.Generate(root, config1)
	if err == nil {
		t.Error("Expected error for missing required template variable TASK, got nil")
	}
	if !strings.Contains(err.Error(), "required template variable 'TASK'") {
		t.Errorf("Expected error about missing TASK variable, got: %v", err)
	}

	// Test case 2: Empty TASK variable
	config2 := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "", // Empty TASK variable
		},
	}

	_, err = gen.Generate(root, config2)
	if err == nil {
		t.Error("Expected error for empty required template variable TASK, got nil")
	}
	if !strings.Contains(err.Error(), "required template variable 'TASK'") {
		t.Errorf("Expected error about empty TASK variable, got: %v", err)
	}

	// Test case 3: Whitespace-only TASK variable
	config3 := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "   ", // Whitespace-only TASK variable
		},
	}

	_, err = gen.Generate(root, config3)
	if err == nil {
		t.Error("Expected error for whitespace-only required template variable TASK, got nil")
	}
	if !strings.Contains(err.Error(), "required template variable 'TASK'") {
		t.Errorf("Expected error about whitespace-only TASK variable, got: %v", err)
	}

	// Test case 4: Valid TASK variable should work
	config4 := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "Valid task description",
		},
	}

	_, err = gen.Generate(root, config4)
	if err != nil {
		t.Errorf("Expected no error with valid TASK variable, got: %v", err)
	}

	// Test case 5: Custom template should not require TASK variable
	customTemplate := "Custom template: {{.CurrentDate}}"
	config5 := GenerateConfig{
		Template:     customTemplate,
		TemplateVars: map[string]string{}, // No TASK variable
	}

	_, err = gen.Generate(root, config5)
	if err != nil {
		t.Errorf("Expected no error with custom template and no TASK variable, got: %v", err)
	}
}

func TestGenerate_BinaryDetectionOptimization(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "context_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large binary file
	binaryContent := make([]byte, 10000) // 10KB
	// Add null bytes at the beginning to make it clearly binary
	for i := 0; i < 100; i++ {
		binaryContent[i] = 0x00
	}
	// Fill the rest with random data
	for i := 100; i < len(binaryContent); i++ {
		binaryContent[i] = byte(i % 256)
	}

	binaryFile := filepath.Join(tmpDir, "large.bin")
	err = os.WriteFile(binaryFile, binaryContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a text file for comparison
	textContent := "This is a text file with some content\n"
	textFile := filepath.Join(tmpDir, "text.txt")
	err = os.WriteFile(textFile, []byte(textContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     tmpDir,
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	binaryNode := &scanner.FileNode{
		Name:     "large.bin",
		Path:     binaryFile,
		IsDir:    false,
		Size:     int64(len(binaryContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	textNode := &scanner.FileNode{
		Name:     "text.txt",
		Path:     textFile,
		IsDir:    false,
		Size:     int64(len(textContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	root.Children = append(root.Children, binaryNode, textNode)

	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		SkipBinary: true,
		TemplateVars: map[string]string{
			"TASK": "Test binary detection optimization",
		},
	}

	result, err := gen.Generate(root, config)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Binary file should not be in the content section
	if strings.Contains(result, "### large.bin") && strings.Contains(result, "## File Contents") {
		t.Error("Binary file content should be skipped when SkipBinary is true")
	}

	// Text file should be in the content
	if !strings.Contains(result, "### text.txt") || !strings.Contains(result, "This is a text file") {
		t.Error("Text file should be included in the output")
	}
}

func TestGenerate_SizeLimitDisambiguation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "context_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a large file that exceeds MaxFileSize but not MaxTotalSize
	largeFileContent := strings.Repeat("x", 1000)
	largeFile := filepath.Join(tmpDir, "large.txt")
	err = os.WriteFile(largeFile, []byte(largeFileContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a smaller file
	smallFileContent := strings.Repeat("y", 100)
	smallFile := filepath.Join(tmpDir, "small.txt")
	err = os.WriteFile(smallFile, []byte(smallFileContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     tmpDir,
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	largeNode := &scanner.FileNode{
		Name:     "large.txt",
		Path:     largeFile,
		IsDir:    false,
		Size:     int64(len(largeFileContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	smallNode := &scanner.FileNode{
		Name:     "small.txt",
		Path:     smallFile,
		IsDir:    false,
		Size:     int64(len(smallFileContent)),
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}
	root.Children = append(root.Children, largeNode, smallNode)

	gen := NewDefaultContextGenerator()

	// Test 1: MaxFileSize limits individual files
	config1 := GenerateConfig{
		MaxFileSize:  500,  // Large file exceeds this
		MaxTotalSize: 5000, // But total would be under this
		TemplateVars: map[string]string{
			"TASK": "Test MaxFileSize limit",
		},
	}

	result, err := gen.Generate(root, config1)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Large file should be skipped from File Contents section due to MaxFileSize
	// (it may still appear in file structure)
	contentsStart := strings.Index(result, "## File Contents")
	if contentsStart != -1 {
		contentsSection := result[contentsStart:]
		if strings.Contains(contentsSection, "### large.txt") {
			t.Error("Large file should be skipped from File Contents due to MaxFileSize limit")
		}
	}
	// Small file should be included
	if !strings.Contains(result, "small.txt") {
		t.Error("Small file should be included")
	}

	// Test 2: MaxTotalSize limits cumulative size
	config2 := GenerateConfig{
		MaxFileSize:  2000, // Both files are under this
		MaxTotalSize: 50,   // But total exceeds this
		TemplateVars: map[string]string{
			"TASK": "Test MaxTotalSize limit",
		},
	}

	_, err = gen.Generate(root, config2)
	if err == nil {
		t.Error("Expected error for MaxTotalSize exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "total size limit") {
		t.Errorf("Expected MaxTotalSize limit error, got: %v", err)
	}

	// Test 3: Backward compatibility with MaxSize
	config3 := GenerateConfig{
		MaxSize: 500, // Should be used for both MaxFileSize and MaxTotalSize
		TemplateVars: map[string]string{
			"TASK": "Test backward compatibility",
		},
	}

	result, err = gen.Generate(root, config3)
	if err != nil {
		t.Errorf("Generate() error = %v", err)
		return
	}

	// Large file should be skipped from File Contents section due to MaxSize (acting as MaxFileSize)
	// (it may still appear in file structure)
	contentsStart = strings.Index(result, "## File Contents")
	if contentsStart != -1 {
		contentsSection := result[contentsStart:]
		if strings.Contains(contentsSection, "### large.txt") {
			t.Error("Large file should be skipped from File Contents due to MaxSize acting as MaxFileSize")
		}
	}
}

func TestTreeRenderer_DefaultHidesIgnored(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tree_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	normalFile := filepath.Join(tmpDir, "normal.txt")
	err = os.WriteFile(normalFile, []byte("normal content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     tmpDir,
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	normalNode := &scanner.FileNode{
		Name:         "normal.txt",
		Path:         normalFile,
		IsDir:        false,
		Size:         14,
		Children:     make([]*scanner.FileNode, 0),
		Selected:     true,
		IsGitignored: false,
		IsCustomIgnored: false,
	}

	ignoredNode := &scanner.FileNode{
		Name:         "ignored.txt",
		Path:         filepath.Join(tmpDir, "ignored.txt"),
		IsDir:        false,
		Size:         14,
		Children:     make([]*scanner.FileNode, 0),
		Selected:     true,
		IsGitignored: true,
		IsCustomIgnored: false,
	}

	root.Children = append(root.Children, normalNode, ignoredNode)

	// Test 1: Default TreeRenderer should hide ignored files
	renderer := NewTreeRenderer()
	result, err := renderer.RenderTree(root)
	if err != nil {
		t.Errorf("RenderTree() error = %v", err)
		return
	}

	if !strings.Contains(result, "normal.txt") {
		t.Error("Normal file should be shown by default")
	}
	if strings.Contains(result, "ignored.txt") {
		t.Error("Ignored file should be hidden by default")
	}

	// Test 2: WithShowIgnored(true) should show ignored files
	renderer2 := NewTreeRenderer().WithShowIgnored(true)
	result2, err := renderer2.RenderTree(root)
	if err != nil {
		t.Errorf("RenderTree() error = %v", err)
		return
	}

	if !strings.Contains(result2, "normal.txt") {
		t.Error("Normal file should be shown when showIgnored=true")
	}
	if !strings.Contains(result2, "ignored.txt") {
		t.Error("Ignored file should be shown when showIgnored=true")
	}
	if !strings.Contains(result2, "(g)") {
		t.Error("Ignored file should have gitignore indicator when shown")
	}
}

func TestGenerateWithProgressEx(t *testing.T) {
	gen := NewDefaultContextGenerator()
	config := GenerateConfig{
		TemplateVars: map[string]string{
			"TASK": "Test structured progress",
		},
	}

	root := &scanner.FileNode{
		Name:     "test",
		Path:     "/test",
		IsDir:    true,
		Size:     0,
		Children: make([]*scanner.FileNode, 0),
		Selected: true,
	}

	var progressEvents []GenProgress
	progress := func(p GenProgress) {
		progressEvents = append(progressEvents, p)
	}

	_, err := gen.GenerateWithProgressEx(root, config, progress)
	if err != nil {
		t.Errorf("GenerateWithProgressEx() error = %v", err)
		return
	}

	// Check that we received structured progress events
	expectedStages := []string{"tree_generation", "content_collection", "template_rendering", "complete"}
	expectedMessages := []string{
		"Generating file structure...",
		"Collecting file contents...",
		"Rendering template...",
		"Context generation completed",
	}

	if len(progressEvents) != len(expectedStages) {
		t.Errorf("Expected %d progress events, got %d", len(expectedStages), len(progressEvents))
		return
	}

	for i, expected := range expectedStages {
		if progressEvents[i].Stage != expected {
			t.Errorf("Progress event %d: expected stage %q, got %q", i, expected, progressEvents[i].Stage)
		}
	}

	for i, expected := range expectedMessages {
		if progressEvents[i].Message != expected {
			t.Errorf("Progress event %d: expected message %q, got %q", i, expected, progressEvents[i].Message)
		}
	}
}