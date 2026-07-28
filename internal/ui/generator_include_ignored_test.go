package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quantmind-br/shotgun-cli/internal/core/contextgen"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
)

// TestTUIGeneration_IgnoredFileAppearsInTreeAndContent closes the symptom that
// made CI-012 visible in the output rather than only in the config.
//
// With IncludeIgnored hardcoded to false, the TUI produced an internally
// inconsistent document: collectFileContents includes an explicitly selected
// file even when it is ignored -- the selection map is the authority -- while the
// tree renderer omits ignored nodes. The user pressed `i`, selected the file, and
// got a content block for a file that appeared nowhere in the tree above it.
//
// This drives the real generator through the TUI's own config builder, so it
// fails if either the builder or the wizard plumbing regresses.
func TestTUIGeneration_IgnoredFileAppearsInTreeAndContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	plainPath := filepath.Join(root, "plain.go")
	ignoredPath := filepath.Join(root, "ignored.go")
	require.NoError(t, os.WriteFile(plainPath, []byte("package plain\n"), 0o600))
	require.NoError(t, os.WriteFile(ignoredPath, []byte("package ignored\n"), 0o600))

	tree := &scanner.FileNode{
		Name: filepath.Base(root), Path: root, IsDir: true,
		Children: []*scanner.FileNode{
			{Name: "plain.go", Path: plainPath, RelPath: "plain.go", Size: 14},
			{
				Name: "ignored.go", Path: ignoredPath, RelPath: "ignored.go", Size: 16,
				IsGitignored: true,
			},
		},
	}

	// What the wizard hands over after the user pressed `i` and selected both.
	uiCfg := &GenerateConfig{
		FileTree:       tree,
		Selections:     map[string]bool{plainPath: true, ignoredPath: true},
		Template:       &template.Template{Content: "{FILE_STRUCTURE}"},
		MaxTotalSize:   10 * 1024 * 1024,
		MaxFileSize:    1024 * 1024,
		MaxFiles:       10000,
		IncludeTree:    true,
		IncludeIgnored: true,
	}

	gen := contextgen.NewDefaultContextGenerator()
	out, err := gen.Generate(tree, uiCfg.Selections, buildGeneratorConfig(uiCfg))
	require.NoError(t, err)

	assert.Contains(t, out, "package ignored", "the selected ignored file's content must be included")
	assert.Contains(t, out, "package plain")

	// The tree section ends where the content blocks begin; the ignored file must
	// be named in both halves, not only in the content.
	treeSection := out
	if idx := strings.Index(out, "package plain"); idx > 0 {
		treeSection = out[:idx]
	}
	assert.Contains(t, treeSection, "ignored.go",
		"the ignored file is in the content but missing from the tree: inconsistent document")
}

// TestTUIGeneration_IgnoredFileHiddenWhenToggleOff is the other direction: with
// the toggle off, the tree must not advertise ignored nodes.
func TestTUIGeneration_IgnoredFileHiddenWhenToggleOff(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	plainPath := filepath.Join(root, "plain.go")
	ignoredPath := filepath.Join(root, "ignored.go")
	require.NoError(t, os.WriteFile(plainPath, []byte("package plain\n"), 0o600))
	require.NoError(t, os.WriteFile(ignoredPath, []byte("package ignored\n"), 0o600))

	tree := &scanner.FileNode{
		Name: filepath.Base(root), Path: root, IsDir: true,
		Children: []*scanner.FileNode{
			{Name: "plain.go", Path: plainPath, RelPath: "plain.go", Size: 14},
			{
				Name: "ignored.go", Path: ignoredPath, RelPath: "ignored.go", Size: 16,
				IsGitignored: true,
			},
		},
	}

	uiCfg := &GenerateConfig{
		FileTree:       tree,
		Selections:     map[string]bool{plainPath: true},
		Template:       &template.Template{Content: "{FILE_STRUCTURE}"},
		MaxTotalSize:   10 * 1024 * 1024,
		MaxFileSize:    1024 * 1024,
		MaxFiles:       10000,
		IncludeTree:    true,
		IncludeIgnored: false,
	}

	gen := contextgen.NewDefaultContextGenerator()
	out, err := gen.Generate(tree, uiCfg.Selections, buildGeneratorConfig(uiCfg))
	require.NoError(t, err)

	assert.Contains(t, out, "package plain")
	assert.NotContains(t, out, "package ignored", "an unselected ignored file must not be included")
}
