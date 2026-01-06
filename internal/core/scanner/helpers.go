package scanner

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
