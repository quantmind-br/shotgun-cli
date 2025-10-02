package context

import (
	"strings"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

func TestFileStructureWithContentBlocks(t *testing.T) {
	// Create a mock file tree
	root := &scanner.FileNode{
		Name:     "project",
		Path:     "/tmp/project",
		RelPath:  ".",
		IsDir:    true,
		Selected: true,
		Children: []*scanner.FileNode{
			{
				Name:     "main.go",
				Path:     "/tmp/project/main.go",
				RelPath:  "main.go",
				IsDir:    false,
				Selected: true,
				Size:     50,
			},
			{
				Name:     "README.md",
				Path:     "/tmp/project/README.md",
				RelPath:  "README.md",
				IsDir:    false,
				Selected: true,
				Size:     20,
			},
		},
	}

	// Mock file contents
	files := []FileContent{
		{
			Path:     "/tmp/project/main.go",
			RelPath:  "main.go",
			Language: "go",
			Content:  "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
			Size:     50,
		},
		{
			Path:     "/tmp/project/README.md",
			RelPath:  "README.md",
			Language: "markdown",
			Content:  "# Test Project\n",
			Size:     20,
		},
	}

	// Test tree rendering
	treeRenderer := NewTreeRenderer()
	tree, err := treeRenderer.RenderTree(root)
	if err != nil {
		t.Fatalf("Failed to render tree: %v", err)
	}

	// Test that tree is generated
	if tree == "" {
		t.Error("Tree should not be empty")
	}

	// Test file content blocks rendering
	contentBlocks := renderFileContentBlocks(files)

	// Verify content blocks format
	if !strings.Contains(contentBlocks, `<file path="main.go">`) {
		t.Error("Content blocks should contain <file path=\"main.go\">")
	}
	if !strings.Contains(contentBlocks, `<file path="README.md">`) {
		t.Error("Content blocks should contain <file path=\"README.md\">")
	}
	if !strings.Contains(contentBlocks, `package main`) {
		t.Error("Content blocks should contain main.go content")
	}
	if !strings.Contains(contentBlocks, `# Test Project`) {
		t.Error("Content blocks should contain README.md content")
	}
	if !strings.Contains(contentBlocks, `</file>`) {
		t.Error("Content blocks should contain closing </file> tags")
	}

	// Test complete file structure (tree + content blocks)
	generator := NewDefaultContextGenerator()
	completeStructure := generator.buildCompleteFileStructure(tree, files)

	// Verify complete structure has both parts
	if !strings.Contains(completeStructure, "project/") {
		t.Error("Complete structure should contain tree")
	}
	if !strings.Contains(completeStructure, `<file path="main.go">`) {
		t.Error("Complete structure should contain file content blocks")
	}

	// Verify format: tree comes first, then content blocks
	treeIndex := strings.Index(completeStructure, "project/")
	contentIndex := strings.Index(completeStructure, `<file path="main.go">`)

	if treeIndex == -1 || contentIndex == -1 {
		t.Error("Complete structure should contain both tree and content blocks")
	}
	if treeIndex >= contentIndex {
		t.Error("Tree should come before content blocks")
	}

	t.Logf("Complete FILE_STRUCTURE:\n%s", completeStructure)
}

func TestRenderFileContentBlocks(t *testing.T) {
	tests := []struct {
		name     string
		files    []FileContent
		expected []string
	}{
		{
			name:     "empty files",
			files:    []FileContent{},
			expected: []string{},
		},
		{
			name: "single file",
			files: []FileContent{
				{
					RelPath: "test.go",
					Content: "package test\n",
				},
			},
			expected: []string{
				`<file path="test.go">`,
				"package test",
				`</file>`,
			},
		},
		{
			name: "multiple files",
			files: []FileContent{
				{
					RelPath: "file1.txt",
					Content: "content 1",
				},
				{
					RelPath: "dir/file2.txt",
					Content: "content 2",
				},
			},
			expected: []string{
				`<file path="file1.txt">`,
				"content 1",
				`</file>`,
				`<file path="dir/file2.txt">`,
				"content 2",
				`</file>`,
			},
		},
		{
			name: "file without trailing newline",
			files: []FileContent{
				{
					RelPath: "test.txt",
					Content: "no newline",
				},
			},
			expected: []string{
				`<file path="test.txt">`,
				"no newline",
				`</file>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderFileContentBlocks(tt.files)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got:\n%s", expected, result)
				}
			}
		})
	}
}
