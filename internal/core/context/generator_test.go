package context

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

// Compile-time contract checks
var _ ContextGenerator = (*DefaultContextGenerator)(nil)

func TestDefaultContextGenerator_validateConfig(t *testing.T) {
	t.Parallel()

	gen := NewDefaultContextGenerator()
	cases := []struct {
		name     string
		input    GenerateConfig
		expected GenerateConfig
	}{
		{
			name:  "defaults applied",
			input: GenerateConfig{},
			expected: GenerateConfig{
				MaxFileSize:  DefaultMaxSize,
				MaxTotalSize: DefaultMaxSize,
				MaxFiles:     DefaultMaxFiles,
				TemplateVars: map[string]string{},
			},
		},
		{
			name: "custom totals respected",
			input: GenerateConfig{
				MaxFileSize: 512, MaxTotalSize: 1024, MaxFiles: 5,
				TemplateVars: map[string]string{"TASK": "x"},
			},
			expected: GenerateConfig{
				MaxFileSize:  512,
				MaxTotalSize: 1024,
				MaxFiles:     5,
				TemplateVars: map[string]string{"TASK": "x"},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := tc.input
			if err := gen.validateConfig(&cfg); err != nil {
				t.Fatalf("validateConfig returned error: %v", err)
			}
			if cfg.MaxFileSize != tc.expected.MaxFileSize ||
				cfg.MaxTotalSize != tc.expected.MaxTotalSize {
				t.Fatalf("unexpected size config: %+v", cfg)
			}
			if cfg.MaxFiles != tc.expected.MaxFiles {
				t.Fatalf("unexpected MaxFiles: got %d want %d", cfg.MaxFiles, tc.expected.MaxFiles)
			}
			if tc.expected.TemplateVars != nil && cfg.TemplateVars == nil {
				t.Fatalf("expected non-nil TemplateVars")
			}
		})
	}
}

type fileSpec struct {
	relPath       string
	content       string
	isDir         bool
	selected      bool
	gitIgnored    bool
	customIgnored bool
}

func buildTestTree(tb testing.TB, specs []fileSpec) (*scanner.FileNode, map[string]bool, func()) {
	tb.Helper()
	rootDir := tb.TempDir()
	root := &scanner.FileNode{
		Name:    filepath.Base(rootDir),
		Path:    rootDir,
		RelPath: ".",
		IsDir:   true,
	}

	selections := make(map[string]bool)
	nodes := map[string]*scanner.FileNode{rootDir: root}

	// Sort to ensure parents created before children on case-insensitive FS
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].relPath < specs[j].relPath
	})

	for _, spec := range specs {
		absPath := filepath.Join(rootDir, filepath.FromSlash(spec.relPath))
		if spec.isDir {
			if err := os.MkdirAll(absPath, 0o750); err != nil {
				tb.Fatalf("mkdir: %v", err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(absPath), 0o750); err != nil {
				tb.Fatalf("mkdir parent: %v", err)
			}
			if err := os.WriteFile(absPath, []byte(spec.content), 0o600); err != nil {
				tb.Fatalf("write file: %v", err)
			}
		}

		node := &scanner.FileNode{
			Name:            filepath.Base(absPath),
			Path:            absPath,
			RelPath:         spec.relPath,
			IsDir:           spec.isDir,
			IsGitignored:    spec.gitIgnored,
			IsCustomIgnored: spec.customIgnored,
		}

		if spec.selected {
			selections[absPath] = true
		}
		if !spec.isDir {
			node.Size = int64(len(spec.content))
		}

		parentPath := filepath.Dir(absPath)
		parent, ok := nodes[parentPath]
		if !ok {
			parent = ensureDirNode(tb, nodes, rootDir, parentPath)
		}
		parent.Children = append(parent.Children, node)
		node.Parent = parent
		if spec.isDir {
			nodes[absPath] = node
		}
	}

	cleanup := func() {
		_ = os.RemoveAll(rootDir)
	}

	return root, selections, cleanup
}

func ensureDirNode(tb testing.TB, nodes map[string]*scanner.FileNode, rootDir, path string) *scanner.FileNode {
	if node, ok := nodes[path]; ok {
		return node
	}
	if path == rootDir {
		return nodes[path]
	}

	parent := ensureDirNode(tb, nodes, rootDir, filepath.Dir(path))
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		tb.Fatalf("failed to compute relative path: %v", err)
	}

	node := &scanner.FileNode{
		Name:    filepath.Base(path),
		Path:    path,
		RelPath: rel,
		IsDir:   true,
		Parent:  parent,
	}
	parent.Children = append(parent.Children, node)
	nodes[path] = node

	return node
}

//nolint:gocyclo // table-driven test with comprehensive scenario coverage
func TestDefaultContextGenerator_GenerateScenarios(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name        string
		specs       []fileSpec
		config      GenerateConfig
		expectError bool
		verify      func(t *testing.T, output string, err error)
	}{
		{
			name:  "single go file",
			specs: []fileSpec{{relPath: "main.go", content: "package main\n// hello", selected: true}},
			config: GenerateConfig{
				MaxTotalSize: 1 << 20,
				MaxFileSize:  1 << 20,
				MaxFiles:     10,
				TemplateVars: map[string]string{"TASK": "Summarize", "RULES": "Be concise"},
				IncludeTree:  true,
			},
			verify: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(output, "# Project Context") {
					t.Fatalf("missing header in output: %s", output)
				}
				if !strings.Contains(output, "**Task:** Summarize") {
					t.Fatalf("TASK variable not substituted")
				}
				if !strings.Contains(output, "**Rules:** Be concise") {
					t.Fatalf("RULES variable not substituted")
				}
				if !strings.Contains(output, "main.go (go)") {
					t.Fatalf("language detection missing: %s", output)
				}
				if !strings.Contains(output, "└── main.go") {
					t.Fatalf("tree output missing: %s", output)
				}
				if strings.Contains(output, "{CURRENT_DATE}") {
					t.Fatalf("CURRENT_DATE placeholder not substituted")
				}
			},
		},
		{
			name: "binary skipped when requested",
			specs: []fileSpec{
				{relPath: "assets/logo.png", content: string([]byte{0x00, 0x01, 0x02}), selected: true},
				{relPath: "README.md", content: "hello", selected: true},
			},
			config: GenerateConfig{
				SkipBinary:   true,
				MaxTotalSize: 1 << 20,
				MaxFileSize:  1 << 20,
				MaxFiles:     10,
				TemplateVars: map[string]string{"TASK": "Describe"},
			},
			verify: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(output, "README.md") {
					t.Fatalf("text file should be present")
				}
				if contents := section(output, "## File Contents"); strings.Contains(contents, "logo.png") {
					t.Fatalf("binary file should not appear in file contents: %s", contents)
				}
			},
		},
		{
			name: "total size limit triggers error",
			specs: []fileSpec{
				{relPath: "a.txt", content: strings.Repeat("A", 600), selected: true},
				{relPath: "b.txt", content: strings.Repeat("B", 600), selected: true},
			},
			config: GenerateConfig{
				MaxTotalSize: 1000,
				MaxFileSize:  1000,
				MaxFiles:     5,
				TemplateVars: map[string]string{"TASK": "Check size"},
			},
			expectError: true,
			verify: func(t *testing.T, _ string, err error) {
				if err == nil || !strings.Contains(err.Error(), "total size") {
					t.Fatalf("expected total size error, got %v", err)
				}
			},
		},
		{
			name: "max files enforcement",
			specs: []fileSpec{
				{relPath: "file1.txt", content: "1", selected: true},
				{relPath: "file2.txt", content: "2", selected: true},
			},
			config: GenerateConfig{
				MaxTotalSize: 1 << 20,
				MaxFileSize:  1 << 20,
				MaxFiles:     1,
				TemplateVars: map[string]string{"TASK": "Limit"},
			},
			expectError: true,
			verify: func(t *testing.T, _ string, err error) {
				if err == nil || !strings.Contains(err.Error(), "maximum file count") {
					t.Fatalf("expected max files error, got %v", err)
				}
			},
		},
		{
			name:  "custom template respected",
			specs: []fileSpec{{relPath: "src/main.go", content: "package main", selected: true}},
			config: GenerateConfig{
				Template:     "Report {{.Task}} with {{.CurrentDate}}",
				MaxTotalSize: 1 << 20,
				MaxFileSize:  1 << 20,
				MaxFiles:     5,
				TemplateVars: map[string]string{"TASK": "Custom"},
			},
			verify: func(t *testing.T, output string, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(output, "Report Custom") {
					t.Fatalf("custom template not rendered: %s", output)
				}
				if strings.Contains(output, "# Project Context") {
					t.Fatalf("default template should not be used: %s", output)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root, selections, cleanup := buildTestTree(t, tc.specs)
			defer cleanup()

			gen := NewDefaultContextGenerator()
			// ensure deterministic time by overriding template renderer required var validation if needed
			tc.config.TemplateVars = ensureTemplateVars(tc.config.TemplateVars)
			tc.config.MaxFileSize = normalizeLimit(tc.config.MaxFileSize)
			tc.config.MaxTotalSize = normalizeLimit(tc.config.MaxTotalSize)
			tc.config.MaxFiles = normalizeFiles(tc.config.MaxFiles)

			output, err := gen.Generate(root, selections, tc.config)
			if tc.expectError {
				tc.verify(t, output, err)

				return
			}
			if err != nil {
				t.Fatalf("Generate returned error: %v", err)
			}
			if strings.Contains(output, now.Add(24*time.Hour).Format("2006-01-02")) {
				t.Fatalf("unexpected future date in output")
			}
			tc.verify(t, output, err)
		})
	}
}

func ensureTemplateVars(vars map[string]string) map[string]string {
	if vars == nil {
		vars = map[string]string{}
	}
	if _, ok := vars["TASK"]; !ok {
		vars["TASK"] = "default task"
	}

	return vars
}

func normalizeLimit(v int64) int64 {
	if v == 0 {
		return DefaultMaxSize
	}

	return v
}

func normalizeFiles(v int) int {
	if v == 0 {
		return DefaultMaxFiles
	}

	return v
}

func section(content, heading string) string {
	idx := strings.Index(content, heading)
	if idx == -1 {
		return ""
	}

	return content[idx:]
}

func TestDefaultContextGenerator_ProgressReporting(t *testing.T) {
	t.Parallel()
	specs := []fileSpec{{relPath: "file.txt", content: "content", selected: true}}
	root, selections, cleanup := buildTestTree(t, specs)
	defer cleanup()

	gen := NewDefaultContextGenerator()
	cfg := GenerateConfig{
		TemplateVars: map[string]string{"TASK": "x"},
		IncludeTree:  true, // Enable tree to get all progress events
	}

	var events []GenProgress
	out, err := gen.GenerateWithProgressEx(root, selections, cfg, func(p GenProgress) {
		events = append(events, p)
	})
	if err != nil {
		t.Fatalf("GenerateWithProgressEx failed: %v", err)
	}
	if out == "" {
		t.Fatalf("expected output")
	}

	expectedStages := []string{"tree_generation", "content_collection", "template_rendering", "complete"}
	if len(events) != len(expectedStages) {
		t.Fatalf("expected %d events, got %d", len(expectedStages), len(events))
	}
	for i, stage := range expectedStages {
		if events[i].Stage != stage {
			t.Fatalf("event %d stage mismatch: got %s want %s", i, events[i].Stage, stage)
		}
		if events[i].Message == "" {
			t.Fatalf("event %d message should not be empty", i)
		}
	}
}

func TestDefaultContextGenerator_ErrorPropagation(t *testing.T) {
	t.Parallel()

	specs := []fileSpec{{relPath: "missing.txt", content: "temp", selected: true}}
	root, selections, cleanup := buildTestTree(t, specs)
	cleanup()
	// remove file to trigger read error
	_ = os.Remove(filepath.Join(root.Path, "missing.txt"))

	gen := NewDefaultContextGenerator()
	cfg := GenerateConfig{TemplateVars: map[string]string{"TASK": "x"}}

	_, err := gen.Generate(root, selections, cfg)
	if err == nil {
		t.Fatalf("expected error when file is missing")
	}
	if !strings.Contains(err.Error(), "failed to collect file contents") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDefaultContextGenerator_IgnoresFlaggedFilesInTree(t *testing.T) {
	t.Parallel()
	specs := []fileSpec{
		{relPath: "src", isDir: true, selected: true},
		{relPath: "src/visible.go", content: "package main", selected: true},
		{relPath: "src/ignored.go", content: "package main", selected: true, gitIgnored: true},
	}
	root, _, cleanup := buildTestTree(t, specs)
	defer cleanup()

	renderer := NewTreeRenderer()
	out, err := renderer.RenderTree(root)
	if err != nil {
		t.Fatalf("RenderTree failed: %v", err)
	}
	if strings.Contains(out, "ignored.go") {
		t.Fatalf("ignored file should not be rendered: %s", out)
	}

	outShowIgnored, err := renderer.WithShowIgnored(true).RenderTree(root)
	if err != nil {
		t.Fatalf("RenderTree with ignored failed: %v", err)
	}
	if !strings.Contains(outShowIgnored, "ignored.go") {
		t.Fatalf("ignored file should be rendered when showIgnored is true")
	}
}

func TestDefaultContextGenerator_TemplateValidation(t *testing.T) {
	t.Parallel()

	specs := []fileSpec{{relPath: "file.txt", content: "hello", selected: true}}
	root, selections, cleanup := buildTestTree(t, specs)
	defer cleanup()

	gen := NewDefaultContextGenerator()
	cfg := GenerateConfig{TemplateVars: map[string]string{}}

	_, err := gen.Generate(root, selections, cfg)
	if err == nil || !strings.Contains(err.Error(), "required template variable 'TASK'") {
		t.Fatalf("expected required variable error, got %v", err)
	}

	cfg.Template = "Custom {{.CurrentDate}}"
	_, err = gen.Generate(root, selections, cfg)
	if err != nil {
		t.Fatalf("custom template should bypass TASK requirement: %v", err)
	}
}

func BenchmarkDefaultContextGenerator(b *testing.B) {
	specs := make([]fileSpec, 0, 50)
	for i := 0; i < 50; i++ {
		specs = append(specs, fileSpec{
			relPath:  filepath.Join("pkg", "module", "file"+strconv.Itoa(i)+".go"),
			content:  "package module\nvar _ = \"benchmark\"",
			selected: true,
		})
	}
	root, selections, cleanup := buildTestTree(b, specs)
	defer cleanup()

	gen := NewDefaultContextGenerator()
	cfg := GenerateConfig{
		MaxTotalSize: 10 << 20, MaxFileSize: 1 << 20, MaxFiles: 100,
		TemplateVars: map[string]string{"TASK": "benchmark"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := gen.Generate(root, selections, cfg); err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}
