package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileStructureWithContentBlocks(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"main.go": `package main

func main() {
    println("Hello, World!")
}`,
		"README.md": `# Test Project

This is a test.`,
		"config.yaml": `server:
  port: 8080`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Build path to shotgun-cli binary
	binaryPath := filepath.Join("..", "..", "build", "shotgun-cli")

	// Generate context
	outputFile := filepath.Join(tmpDir, "output.md")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, //nolint:gosec // test command with controlled args
		"context", "generate",
		"--root", tmpDir,
		"--include", "*.go,*.md,*.yaml",
		"--output", outputFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
	}

	// Read generated output
	content, err := os.ReadFile(outputFile) //nolint:gosec // test reading controlled file
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify ASCII tree is present
	if !strings.Contains(contentStr, "└──") {
		t.Error("Output should contain directory tree with └── connector")
	}

	// Verify XML-like content blocks are present
	expectedBlocks := []string{
		`<file path="main.go">`,
		`package main`,
		`</file>`,
		`<file path="README.md">`,
		`# Test Project`,
		`<file path="config.yaml">`,
		`server:`,
		`port: 8080`,
	}

	for _, expected := range expectedBlocks {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Output should contain %q\nGot:\n%s", expected, contentStr)
		}
	}

	// Verify structure: tree comes before content blocks
	treePattern := "└──"
	fileBlockPattern := `<file path="`

	treeIndex := strings.Index(contentStr, treePattern)
	blockIndex := strings.Index(contentStr, fileBlockPattern)

	if treeIndex == -1 {
		t.Error("Output should contain tree structure")
	}
	if blockIndex == -1 {
		t.Error("Output should contain file content blocks")
	}
	if treeIndex >= blockIndex {
		t.Error("Tree structure should come before file content blocks")
	}

	t.Logf("Successfully validated FILE_STRUCTURE format:\n%s", contentStr)
}

func TestFileStructureOnlyIncludesSelectedFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"included.go":  `package main`,
		"excluded.txt": `This should not be included`,
		"also.go":      `package test`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Build path to shotgun-cli binary
	binaryPath := filepath.Join("..", "..", "build", "shotgun-cli")

	// Generate context with only .go files
	outputFile := filepath.Join(tmpDir, "output.md")
	cmd := exec.Command(binaryPath, //nolint:gosec // test command with controlled args
		"context", "generate",
		"--root", tmpDir,
		"--include", "*.go",
		"--output", outputFile,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command failed: %v\nOutput: %s", err, string(output))
	}

	// Read generated output
	content, err := os.ReadFile(outputFile) //nolint:gosec // test reading controlled file
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Verify .go files are included in content blocks
	if !strings.Contains(contentStr, `<file path="included.go">`) {
		t.Error("Output should contain included.go content block")
	}
	if !strings.Contains(contentStr, `<file path="also.go">`) {
		t.Error("Output should contain also.go content block")
	}

	// Verify .txt file is NOT included in content blocks
	if strings.Contains(contentStr, `<file path="excluded.txt">`) {
		t.Error("Output should NOT contain excluded.txt content block")
	}
	if strings.Contains(contentStr, "This should not be included") {
		t.Error("Output should NOT contain excluded.txt content")
	}
}
