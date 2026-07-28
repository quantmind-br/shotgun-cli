package scanner

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectSelections_NilNode(t *testing.T) {
	selections := CollectSelections(nil, nil)
	assert.Nil(t, selections)
}

func TestCollectSelections_NilSelections(t *testing.T) {
	node := &FileNode{
		Name:  "file.txt",
		Path:  "/root/file.txt",
		IsDir: false,
	}

	selections := CollectSelections(node, nil)
	require.NotNil(t, selections)
	assert.True(t, selections["/root/file.txt"])
}

func TestCollectSelections_SingleFile(t *testing.T) {
	node := &FileNode{
		Name:  "file.txt",
		Path:  "/root/file.txt",
		IsDir: false,
	}

	selections := make(map[string]bool)
	result := CollectSelections(node, selections)

	assert.Equal(t, selections, result, "should return same map")
	assert.True(t, selections["/root/file.txt"])
	assert.Len(t, selections, 1)
}

func TestCollectSelections_IgnoredFile(t *testing.T) {
	node := &FileNode{
		Name:         "ignored.txt",
		Path:         "/root/ignored.txt",
		IsDir:        false,
		IsGitignored: true,
	}

	selections := make(map[string]bool)
	CollectSelections(node, selections)

	assert.False(t, selections["/root/ignored.txt"])
	assert.Len(t, selections, 0)
}

func TestCollectSelections_CustomIgnoredFile(t *testing.T) {
	node := &FileNode{
		Name:            "custom.txt",
		Path:            "/root/custom.txt",
		IsDir:           false,
		IsCustomIgnored: true,
	}

	selections := make(map[string]bool)
	CollectSelections(node, selections)

	assert.Len(t, selections, 0)
}

func TestCollectSelections_DirectoryWithChildren(t *testing.T) {
	root := &FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*FileNode{
			{
				Name:  "file1.txt",
				Path:  "/root/file1.txt",
				IsDir: false,
			},
			{
				Name:  "file2.txt",
				Path:  "/root/file2.txt",
				IsDir: false,
			},
		},
	}

	selections := make(map[string]bool)
	CollectSelections(root, selections)

	assert.True(t, selections["/root"])
	assert.True(t, selections["/root/file1.txt"])
	assert.True(t, selections["/root/file2.txt"])
	assert.Len(t, selections, 3)
}

func TestCollectSelections_NestedDirectories(t *testing.T) {
	root := &FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*FileNode{
			{
				Name:  "src",
				Path:  "/root/src",
				IsDir: true,
				Children: []*FileNode{
					{
						Name:  "main.go",
						Path:  "/root/src/main.go",
						IsDir: false,
					},
				},
			},
		},
	}

	selections := make(map[string]bool)
	CollectSelections(root, selections)

	assert.True(t, selections["/root"])
	assert.True(t, selections["/root/src"])
	assert.True(t, selections["/root/src/main.go"])
	assert.Len(t, selections, 3)
}

func TestCollectSelections_MixedIgnored(t *testing.T) {
	root := &FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*FileNode{
			{
				Name:  "visible.txt",
				Path:  "/root/visible.txt",
				IsDir: false,
			},
			{
				Name:         "gitignored.txt",
				Path:         "/root/gitignored.txt",
				IsDir:        false,
				IsGitignored: true,
			},
			{
				Name:            "customignored.txt",
				Path:            "/root/customignored.txt",
				IsDir:           false,
				IsCustomIgnored: true,
			},
		},
	}

	selections := make(map[string]bool)
	CollectSelections(root, selections)

	assert.True(t, selections["/root"])
	assert.True(t, selections["/root/visible.txt"])
	assert.False(t, selections["/root/gitignored.txt"])
	assert.False(t, selections["/root/customignored.txt"])
	assert.Len(t, selections, 2)
}

func TestCollectSelections_IgnoredDirectory(t *testing.T) {
	root := &FileNode{
		Name:  "root",
		Path:  "/root",
		IsDir: true,
		Children: []*FileNode{
			{
				Name:         "node_modules",
				Path:         "/root/node_modules",
				IsDir:        true,
				IsGitignored: true,
				Children: []*FileNode{
					{
						Name:  "package.json",
						Path:  "/root/node_modules/package.json",
						IsDir: false,
					},
				},
			},
		},
	}

	selections := make(map[string]bool)
	CollectSelections(root, selections)

	assert.True(t, selections["/root"])
	assert.False(t, selections["/root/node_modules"])
	assert.True(t, selections["/root/node_modules/package.json"])
}

func TestCollectSelections_PreserveExisting(t *testing.T) {
	node := &FileNode{
		Name:  "new.txt",
		Path:  "/root/new.txt",
		IsDir: false,
	}

	selections := map[string]bool{
		"/root/existing.txt": true,
	}
	CollectSelections(node, selections)

	assert.True(t, selections["/root/existing.txt"])
	assert.True(t, selections["/root/new.txt"])
	assert.Len(t, selections, 2)
}

func TestNewSelectAll_NilRoot(t *testing.T) {
	selections := NewSelectAll(nil)
	require.NotNil(t, selections)
	assert.Len(t, selections, 0)
}

func TestNewSelectAll_SingleFile(t *testing.T) {
	node := &FileNode{
		Name:  "file.txt",
		Path:  "/root/file.txt",
		IsDir: false,
	}

	selections := NewSelectAll(node)
	assert.True(t, selections["/root/file.txt"])
	assert.Len(t, selections, 1)
}

func TestNewSelectAll_DirectoryTree(t *testing.T) {
	root := &FileNode{
		Name:  "project",
		Path:  "/project",
		IsDir: true,
		Children: []*FileNode{
			{
				Name:  "main.go",
				Path:  "/project/main.go",
				IsDir: false,
			},
			{
				Name:  "lib",
				Path:  "/project/lib",
				IsDir: true,
				Children: []*FileNode{
					{
						Name:  "util.go",
						Path:  "/project/lib/util.go",
						IsDir: false,
					},
				},
			},
			{
				Name:         "vendor",
				Path:         "/project/vendor",
				IsDir:        true,
				IsGitignored: true,
			},
		},
	}

	selections := NewSelectAll(root)

	assert.True(t, selections["/project"])
	assert.True(t, selections["/project/main.go"])
	assert.True(t, selections["/project/lib"])
	assert.True(t, selections["/project/lib/util.go"])
	assert.False(t, selections["/project/vendor"])
	assert.Len(t, selections, 4)
}

func TestNewSelectAll_EmptyDirectory(t *testing.T) {
	root := &FileNode{
		Name:     "empty",
		Path:     "/empty",
		IsDir:    true,
		Children: []*FileNode{},
	}

	selections := NewSelectAll(root)
	assert.True(t, selections["/empty"])
	assert.Len(t, selections, 1)
}

// newSelectionTestTree builds a tree with non-ignored files, a nested subdirectory
// with a file, and both a gitignored and a custom-ignored file, exercising the full
// SelectAllExcept / CollectDeselected surface.
//
//	/project (dir, RelPath ".")
//	├── main.go            RelPath "main.go"
//	├── util.go            RelPath "util.go"
//	├── gitignored.txt     RelPath "gitignored.txt"     (IsGitignored)
//	├── customignored.txt  RelPath "customignored.txt"  (IsCustomIgnored)
//	└── sub (dir)          RelPath "sub"
//	    └── child.go       RelPath "sub/child.go"
func newSelectionTestTree() *FileNode {
	return &FileNode{
		Name:    "project",
		Path:    "/project",
		RelPath: ".",
		IsDir:   true,
		Children: []*FileNode{
			{Name: "main.go", Path: "/project/main.go", RelPath: "main.go", IsDir: false},
			{Name: "util.go", Path: "/project/util.go", RelPath: "util.go", IsDir: false},
			{
				Name:         "gitignored.txt",
				Path:         "/project/gitignored.txt",
				RelPath:      "gitignored.txt",
				IsDir:        false,
				IsGitignored: true,
			},
			{
				Name:            "customignored.txt",
				Path:            "/project/customignored.txt",
				RelPath:         "customignored.txt",
				IsDir:           false,
				IsCustomIgnored: true,
			},
			{
				Name:    "sub",
				Path:    "/project/sub",
				RelPath: "sub",
				IsDir:   true,
				Children: []*FileNode{
					{Name: "child.go", Path: "/project/sub/child.go", RelPath: "sub/child.go", IsDir: false},
				},
			},
		},
	}
}

func TestSelectAllExcept_NilDeselectedEqualsSelectAll(t *testing.T) {
	root := newSelectionTestTree()

	assert.Equal(t, NewSelectAll(root), SelectAllExcept(root, nil, false),
		"SelectAllExcept(root, nil, false) must be identical to NewSelectAll(root)")
}

func TestSelectAllExcept_OmitsDeselectedFileKeepsSiblingsAndDirs(t *testing.T) {
	root := newSelectionTestTree()

	result := SelectAllExcept(root, []string{"util.go"}, false)

	// The deselected file is gone.
	assert.False(t, result["/project/util.go"], "deselected file must be omitted")
	// Sibling files remain selected.
	assert.True(t, result["/project/main.go"], "sibling file must stay selected")
	assert.True(t, result["/project/sub/child.go"], "nested file must stay selected")
	// Directory keys are always kept for non-ignored dirs.
	assert.True(t, result["/project"], "root dir must stay selected")
	assert.True(t, result["/project/sub"], "nested dir must stay selected")
	// Ignored nodes never appear.
	assert.False(t, result["/project/gitignored.txt"])
	assert.False(t, result["/project/customignored.txt"])
	// 5 non-ignored nodes minus the 1 deselected file.
	assert.Len(t, result, 4)
}

func TestSelectAllExcept_NeverIncludesIgnoredNodes(t *testing.T) {
	root := newSelectionTestTree()

	// Even when the ignored files' relpaths are passed as deselected (and when
	// they are not), ignored nodes must never be present in the selection map.
	for _, deselected := range [][]string{
		nil,
		{"gitignored.txt", "customignored.txt", "does-not-exist.go"},
	} {
		result := SelectAllExcept(root, deselected, false)
		assert.False(t, result["/project/gitignored.txt"],
			"gitignored node must never be selected (deselected=%v)", deselected)
		assert.False(t, result["/project/customignored.txt"],
			"custom-ignored node must never be selected (deselected=%v)", deselected)
	}
}

func TestSelectAllExcept_SlashNormalizedMatching(t *testing.T) {
	// The scanner stores RelPath with the OS-native separator (via filepath.Join),
	// while the persistence store records slash-form paths. Matching must survive
	// that separator mismatch because both sides are run through filepath.ToSlash.
	// (On Linux the native separator is already "/", so this exercises the real
	// Windows-relevant conversion while remaining correct here.)
	nestedRel := filepath.Join("sub", "child.go")
	root := &FileNode{
		Name: "project", Path: "/project", RelPath: ".", IsDir: true,
		Children: []*FileNode{
			{Name: "main.go", Path: "/project/main.go", RelPath: "main.go", IsDir: false},
			{
				Name: "sub", Path: "/project/sub", RelPath: "sub", IsDir: true,
				Children: []*FileNode{
					{Name: "child.go", Path: "/project/sub/child.go", RelPath: nestedRel, IsDir: false},
				},
			},
		},
	}

	// Deselect using the slash-form the store would persist.
	result := SelectAllExcept(root, []string{"sub/child.go"}, false)

	assert.False(t, result["/project/sub/child.go"], "slash-form deselected entry must omit the native-separator file")
	assert.True(t, result["/project/main.go"], "unrelated sibling must remain")
	assert.True(t, result["/project/sub"], "parent directory must remain")
}

func TestSelectAllExcept_NilRootReturnsEmptyMap(t *testing.T) {
	result := SelectAllExcept(nil, []string{"anything.go"}, false)

	require.NotNil(t, result, "must return a non-nil map")
	assert.Len(t, result, 0)
}

func TestCollectDeselected_SortedSlashFormExcludesDirsAndIgnored(t *testing.T) {
	// Children are deliberately in reverse-alphabetical order so tree-traversal
	// order differs from the required sorted output.
	root := &FileNode{
		Name:    "r",
		Path:    "/r",
		RelPath: ".",
		IsDir:   true,
		Children: []*FileNode{
			{Name: "zebra.go", Path: "/r/zebra.go", RelPath: "zebra.go", IsDir: false},
			{Name: "alpha.go", Path: "/r/alpha.go", RelPath: "alpha.go", IsDir: false},
			{
				Name:    "mid",
				Path:    "/r/mid",
				RelPath: "mid",
				IsDir:   true,
				Children: []*FileNode{
					{Name: "beta.go", Path: "/r/mid/beta.go", RelPath: "mid/beta.go", IsDir: false},
				},
			},
			{Name: "selected.go", Path: "/r/selected.go", RelPath: "selected.go", IsDir: false},
			{
				Name:         "ignored.txt",
				Path:         "/r/ignored.txt",
				RelPath:      "ignored.txt",
				IsDir:        false,
				IsGitignored: true,
			},
		},
	}

	// Only selected.go (plus dir keys, which CollectDeselected ignores) is selected.
	selections := map[string]bool{
		"/r":             true,
		"/r/mid":         true,
		"/r/selected.go": true,
	}

	result := CollectDeselected(root, selections)

	// Exactly the non-ignored, unselected FILE nodes, sorted ascending. Directories
	// (/r, /r/mid), the selected file, and the ignored file are all excluded.
	assert.Equal(t, []string{"alpha.go", "mid/beta.go", "zebra.go"}, result)
}

func TestCollectDeselected_NilRootReturnsEmptySlice(t *testing.T) {
	result := CollectDeselected(nil, nil)

	assert.Len(t, result, 0)
}

func TestCollectDeselected_RoundTripsWithSelectAllExcept(t *testing.T) {
	root := newSelectionTestTree()

	// Mixed input: a plain slash-form file, a native-separator file (what the
	// scanner produces), a path that is not a node, and an ignored file. Only real
	// non-ignored files should survive the round trip.
	ds := []string{"util.go", filepath.FromSlash("sub/child.go"), "does-not-exist.go", "gitignored.txt"}

	got := CollectDeselected(root, SelectAllExcept(root, ds, false))

	// Sorted, slash-normalized ds restricted to actual non-ignored files in the tree.
	assert.Equal(t, []string{"sub/child.go", "util.go"}, got)
}

// TestSelectAllExcept_IncludeIgnored garante que, quando o scan foi pedido com
// --include-ignored, os nós marcados como ignorados entram na seleção — sem
// isso a flag não teria efeito nenhum no modo headless.
func TestSelectAllExcept_IncludeIgnored(t *testing.T) {
	root := &FileNode{
		Name: "root", Path: "/root", RelPath: ".", IsDir: true,
		Children: []*FileNode{
			{Name: "main.go", Path: "/root/main.go", RelPath: "main.go"},
			{Name: "secret.env", Path: "/root/secret.env", RelPath: "secret.env", IsGitignored: true},
		},
	}

	without := SelectAllExcept(root, nil, false)
	assert.True(t, without["/root/main.go"])
	assert.False(t, without["/root/secret.env"], "sem includeIgnored o nó ignorado fica de fora")

	with := SelectAllExcept(root, nil, true)
	assert.True(t, with["/root/main.go"])
	assert.True(t, with["/root/secret.env"], "com includeIgnored o nó ignorado é selecionado")

	deselected := SelectAllExcept(root, []string{"secret.env"}, true)
	assert.False(t, deselected["/root/secret.env"], "desmarcação explícita continua valendo")
}
