package scanner

import (
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
