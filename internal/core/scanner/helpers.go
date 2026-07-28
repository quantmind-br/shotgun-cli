package scanner

import (
	"path/filepath"
	"sort"
)

// CollectSelections recursively collects all non-ignored file paths into a selection map.
func CollectSelections(node *FileNode, selections map[string]bool) map[string]bool {
	if node == nil {
		return selections
	}
	if selections == nil {
		selections = make(map[string]bool)
	}

	if !node.IsIgnored() {
		selections[node.Path] = true
	}

	if node.IsDir {
		for _, child := range node.Children {
			CollectSelections(child, selections)
		}
	}

	return selections
}

// NewSelectAll creates a selection map with all non-ignored files selected.
func NewSelectAll(root *FileNode) map[string]bool {
	return CollectSelections(root, make(map[string]bool))
}

// SelectAllExcept returns a selection map (absolute path -> true) of every node,
// omitting file nodes whose relative path (slash-normalized) appears in deselected.
// Nodes marked as ignored are selected only when includeIgnored is true, which is
// what the caller passes when the scan was asked to include them.
func SelectAllExcept(root *FileNode, deselected []string, includeIgnored bool) map[string]bool {
	set := make(map[string]bool, len(deselected))
	for _, p := range deselected {
		set[filepath.ToSlash(p)] = true
	}
	out := make(map[string]bool)
	selectAllExcept(root, set, out, includeIgnored)
	return out
}

func selectAllExcept(node *FileNode, deselected, out map[string]bool, includeIgnored bool) {
	if node == nil {
		return
	}
	if includeIgnored || !node.IsIgnored() {
		if node.IsDir || !deselected[filepath.ToSlash(node.RelPath)] {
			out[node.Path] = true
		}
	}
	if node.IsDir {
		for _, child := range node.Children {
			selectAllExcept(child, deselected, out, includeIgnored)
		}
	}
}

// CollectDeselected returns the slash-normalized relative paths of non-ignored FILE nodes
// that are not selected in selections. The result is sorted and de-duplicated.
func CollectDeselected(root *FileNode, selections map[string]bool) []string {
	seen := make(map[string]bool)
	collectDeselected(root, selections, seen)
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func collectDeselected(node *FileNode, selections, seen map[string]bool) {
	if node == nil {
		return
	}
	if !node.IsDir && !node.IsIgnored() && !selections[node.Path] {
		seen[filepath.ToSlash(node.RelPath)] = true
	}
	if node.IsDir {
		for _, child := range node.Children {
			collectDeselected(child, selections, seen)
		}
	}
}
