package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shotgun-cli/internal/core/scanner"
)

type TreeRenderer struct {
	showIgnored bool
	maxDepth    int
}

func NewTreeRenderer() *TreeRenderer {
	return &TreeRenderer{
		showIgnored: true,
		maxDepth:    -1, // No limit by default
	}
}

func (tr *TreeRenderer) WithShowIgnored(show bool) *TreeRenderer {
	tr.showIgnored = show
	return tr
}

func (tr *TreeRenderer) WithMaxDepth(depth int) *TreeRenderer {
	tr.maxDepth = depth
	return tr
}

func (tr *TreeRenderer) RenderTree(root *scanner.FileNode) (string, error) {
	if root == nil {
		return "", fmt.Errorf("root node is nil")
	}

	var result strings.Builder
	tr.renderNode(root, "", true, 0, &result)
	return result.String(), nil
}

func (tr *TreeRenderer) renderNode(node *scanner.FileNode, prefix string, isLast bool, depth int, result *strings.Builder) {
	if tr.maxDepth >= 0 && depth > tr.maxDepth {
		return
	}

	if !tr.showIgnored && node.IsIgnored() {
		return
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	name := node.Name
	if node.IsDir {
		name += "/"
	}

	ignoreIndicator := ""
	if node.IsIgnored() {
		if node.IsGitIgnored() {
			ignoreIndicator = " (g)"
		} else {
			ignoreIndicator = " (c)"
		}
	}

	sizeInfo := ""
	if !node.IsDir && node.Size > 0 {
		sizeInfo = fmt.Sprintf(" [%s]", formatFileSize(node.Size))
	}

	result.WriteString(fmt.Sprintf("%s%s%s%s%s\n", prefix, connector, name, ignoreIndicator, sizeInfo))

	if node.IsDir && len(node.Children) > 0 {
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		children := make([]*scanner.FileNode, 0, len(node.Children))
		for _, child := range node.Children {
			if tr.showIgnored || !child.IsIgnored() {
				children = append(children, child)
			}
		}

		sort.Slice(children, func(i, j int) bool {
			if children[i].IsDir != children[j].IsDir {
				return children[i].IsDir
			}
			return children[i].Name < children[j].Name
		})

		for i, child := range children {
			isChildLast := i == len(children)-1
			tr.renderNode(child, childPrefix, isChildLast, depth+1, result)
		}
	}
}

func formatFileSize(bytes int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	if bytes >= GB {
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	} else if bytes >= MB {
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	} else if bytes >= KB {
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%dB", bytes)
}