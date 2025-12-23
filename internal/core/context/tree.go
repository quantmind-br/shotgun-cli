package context

import (
	"fmt"
	"sort"
	"strings"

	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
)

type TreeRenderer struct {
	showIgnored bool
	maxDepth    int
}

func NewTreeRenderer() *TreeRenderer {
	return &TreeRenderer{
		showIgnored: false, // Default to hiding ignored files
		maxDepth:    -1,    // No limit by default
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

func (tr *TreeRenderer) renderNode(
	node *scanner.FileNode, prefix string, isLast bool, depth int, result *strings.Builder,
) {
	if tr.shouldSkipNode(node, depth) {
		return
	}

	line := tr.formatNodeLine(node, prefix, isLast)
	result.WriteString(line)

	if node.IsDir && len(node.Children) > 0 {
		tr.renderChildren(node, prefix, isLast, depth, result)
	}
}

func (tr *TreeRenderer) shouldSkipNode(node *scanner.FileNode, depth int) bool {
	if tr.maxDepth >= 0 && depth > tr.maxDepth {
		return true
	}
	return !tr.showIgnored && node.IsIgnored()
}

func (tr *TreeRenderer) formatNodeLine(node *scanner.FileNode, prefix string, isLast bool) string {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	name := node.Name
	if node.IsDir {
		name += "/"
	}

	ignoreIndicator := tr.getIgnoreIndicator(node)
	sizeInfo := tr.getSizeInfo(node)

	return fmt.Sprintf("%s%s%s%s%s\n", prefix, connector, name, ignoreIndicator, sizeInfo)
}

func (tr *TreeRenderer) getIgnoreIndicator(node *scanner.FileNode) string {
	if !node.IsIgnored() {
		return ""
	}
	if node.IsGitignored {
		return " (g)"
	}
	return " (c)"
}

func (tr *TreeRenderer) getSizeInfo(node *scanner.FileNode) string {
	if node.IsDir || node.Size == 0 {
		return ""
	}
	return fmt.Sprintf(" [%s]", formatFileSize(node.Size))
}

func (tr *TreeRenderer) renderChildren(
	node *scanner.FileNode, prefix string, isLast bool, depth int, result *strings.Builder,
) {
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	children := tr.getVisibleChildren(node)
	tr.sortChildren(children)

	for i, child := range children {
		isChildLast := i == len(children)-1
		tr.renderNode(child, childPrefix, isChildLast, depth+1, result)
	}
}

func (tr *TreeRenderer) getVisibleChildren(node *scanner.FileNode) []*scanner.FileNode {
	children := make([]*scanner.FileNode, 0, len(node.Children))
	for _, child := range node.Children {
		if tr.showIgnored || !child.IsIgnored() {
			children = append(children, child)
		}
	}
	return children
}

func (tr *TreeRenderer) sortChildren(children []*scanner.FileNode) {
	sort.Slice(children, func(i, j int) bool {
		if children[i].IsDir != children[j].IsDir {
			return children[i].IsDir
		}
		return children[i].Name < children[j].Name
	})
}

func formatFileSize(bytes int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
