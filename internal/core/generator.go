package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	MaxOutputSizeBytes = 0 // No limit
)

var ErrContextTooLong = fmt.Errorf("context is too long")

// GenerateContext generates file context from included files
func (cg *ContextGenerator) GenerateContext(ctx context.Context, files []string) (string, error) {
	if cg.maxSize == 0 {
		cg.maxSize = MaxOutputSizeBytes
	}

	// Initialize worker pool if not set
	if cg.workerPool == nil {
		cg.workerPool = make(chan struct{}, runtime.NumCPU())
	}

	var builder strings.Builder
	builder.Grow(int(cg.maxSize / 4)) // Pre-allocate reasonable capacity

	currentSize := int64(0)
	totalFiles := int64(len(files))

	// Send initial progress
	select {
	case cg.progressChan <- ProgressUpdate{
		Current:     0,
		Total:       totalFiles,
		Percentage:  0,
		CurrentFile: "",
		Phase:       "starting",
	}:
	default:
	}

	// Process files with concurrency control
	var wg sync.WaitGroup
	var mu sync.Mutex
	processed := int64(0)
	errCollector := &ErrorCollector{}

	for _, file := range files {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			// Acquire worker
			cg.workerPool <- struct{}{}
			defer func() { <-cg.workerPool }()

			// Read file content
			content, err := os.ReadFile(filename)
			if err != nil {
				errCollector.Add(&ShotgunError{
					Operation: "read file",
					Path:      filename,
					Err:       err,
				})
				return
			}

			// Get relative path for display
			relPath := filename
			if wd, err := os.Getwd(); err == nil {
				if rel, err := filepath.Rel(wd, filename); err == nil {
					relPath = rel
				}
			}

			// Format file content as XML-like structure (matching existing app)
			formatted := fmt.Sprintf("<file path=\"%s\">\n%s\n</file>\n\n",
				filepath.ToSlash(relPath), string(content))

			// Thread-safe write to builder
			mu.Lock()
			// No size limit check anymore
			builder.WriteString(formatted)
			currentSize += int64(len(formatted))
			processed++

			// Send progress update
			percentage := float64(processed) / float64(totalFiles) * 100
			select {
			case cg.progressChan <- ProgressUpdate{
				Current:     processed,
				Total:       totalFiles,
				Percentage:  percentage,
				CurrentFile: relPath,
				Phase:       "reading files",
			}:
			default:
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	// Check for errors
	if errCollector.HasErrors() {
		errors := errCollector.Errors()
		// Return the first error for simplicity
		return "", errors[0]
	}

	// Send completion progress
	select {
	case cg.progressChan <- ProgressUpdate{
		Current:     totalFiles,
		Total:       totalFiles,
		Percentage:  100,
		CurrentFile: "",
		Phase:       "completed",
	}:
	default:
	}

	return builder.String(), nil
}

// GenerateProjectTree generates a text representation of the project structure
func GenerateProjectTree(root *FileNode, maxDepth int) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s/\n", root.Name))
	generateTreeRecursive(root, "", true, 0, maxDepth, &builder)
	return builder.String()
}

// generateTreeRecursive builds the tree representation recursively
func generateTreeRecursive(node *FileNode, prefix string, isLast bool, depth int, maxDepth int, builder *strings.Builder) {
	if depth >= maxDepth {
		return
	}

	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1

		// Determine the connector
		connector := "├── "
		if isChildLast {
			connector = "└── "
		}

		// Build the full line
		line := prefix + connector + child.Name
		if child.IsDir {
			line += "/"
		}
		builder.WriteString(line + "\n")

		// Recurse for directories
		if child.IsDir && len(child.Children) > 0 {
			newPrefix := prefix
			if isChildLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			generateTreeRecursive(child, newPrefix, isChildLast, depth+1, maxDepth, builder)
		}
	}
}

// GetProgressChannel returns the progress channel for monitoring generation progress
func (cg *ContextGenerator) GetProgressChannel() <-chan ProgressUpdate {
	return cg.progressChan
}

// GenerateContextWithTree generates both tree structure and file content
func (cg *ContextGenerator) GenerateContextWithTree(ctx context.Context, root *FileNode, files []string) (string, error) {
	// Generate the project tree (first part of file structure)
	tree := GenerateProjectTree(root, 5) // Limit tree depth to 5 levels

	// Generate file contents
	content, err := cg.GenerateContext(ctx, files)
	if err != nil {
		return "", err
	}

	// Combine tree and content
	result := fmt.Sprintf("%s\n%s", tree, content)
	return result, nil
}

// EstimateContextSize estimates the size of the context that would be generated
func EstimateContextSize(files []string) (int64, error) {
	var totalSize int64

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue // Skip problematic files
		}

		// Add file content size plus XML wrapper overhead
		fileSize := info.Size()
		relPath := file
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, file); err == nil {
				relPath = rel
			}
		}

		// Estimate XML wrapper size
		wrapperSize := int64(len(fmt.Sprintf("<file path=\"%s\">\n\n</file>\n\n", relPath)))
		totalSize += fileSize + wrapperSize
	}

	return totalSize, nil
}

// FilterFilesBySize filters files to fit within the size limit (now disabled)
func FilterFilesBySize(files []string, maxSize int64) ([]string, int64, error) {
	// No filtering - return all files
	var totalSize int64

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue // Skip problematic files
		}

		// Estimate size with XML wrapper
		relPath := file
		if wd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(wd, file); err == nil {
				relPath = rel
			}
		}
		wrapperSize := int64(len(fmt.Sprintf("<file path=\"%s\">\n\n</file>\n\n", relPath)))
		estimatedSize := info.Size() + wrapperSize
		totalSize += estimatedSize
	}

	return files, totalSize, nil // Return all files
}
