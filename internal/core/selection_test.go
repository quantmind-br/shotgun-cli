package core

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
)

func TestHierarchicalExclusion(t *testing.T) {
	ss := NewSelectionState()

	// Exclude parent directory using proper method
	ss.ExcludeFile("src/components")

	// Child should be excluded via inheritance
	if !ss.IsPathExcluded("src/components/Button.tsx") {
		t.Error("Child file should inherit exclusion from parent directory")
	}

	// Deeply nested child should also be excluded
	if !ss.IsPathExcluded("src/components/ui/forms/Input.tsx") {
		t.Error("Deeply nested child should inherit exclusion from ancestor directory")
	}

	// Sibling directory should not be affected
	if ss.IsPathExcluded("src/utils/helpers.ts") {
		t.Error("Sibling directory should not be affected by parent exclusion")
	}
}

func TestToggleDirectoryBehavior(t *testing.T) {
	ss := NewSelectionState()

	// Initially directory should be included
	if ss.IsPathExcluded("src/components") {
		t.Error("Directory should be included initially")
	}

	// Toggle directory should exclude it
	ss.ToggleFile("src/components")
	if !ss.IsPathExcluded("src/components") {
		t.Error("Directory should be excluded after first toggle")
	}

	// Toggle again should return to inherit (included)
	ss.ToggleFile("src/components")
	if ss.IsPathExcluded("src/components") {
		t.Error("Directory should return to included state after second toggle")
	}

	// Test that children are affected by parent toggle
	ss.ToggleFile("src")
	if !ss.IsPathExcluded("src/components/Button.tsx") {
		t.Error("Child should be excluded when parent is excluded")
	}
}

func TestCrossPlatformPaths(t *testing.T) {
	ss := NewSelectionState()

	// Test Windows-style path handling
	ss.ExcludeFile("src\\components")
	if !ss.IsPathExcluded("src/components/Button.tsx") {
		t.Error("Should handle mixed path separators correctly")
	}

	// Test path normalization
	ss.Reset()
	ss.ExcludeFile("src/./components/../components")
	if !ss.IsPathExcluded("src/components/Button.tsx") {
		t.Error("Should handle path normalization correctly")
	}

	// Test trailing separators
	ss.Reset()
	ss.ExcludeFile("src/components/")
	if !ss.IsPathExcluded("src/components/Button.tsx") {
		t.Error("Should handle trailing separators correctly")
	}
}

func TestConcurrentAccess(t *testing.T) {
	ss := NewSelectionState()

	// Test thread safety with goroutines
	done := make(chan bool, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ss.ToggleFile(fmt.Sprintf("path%d", i))
		}
		done <- true
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = ss.IsPathExcluded(fmt.Sprintf("path%d", i))
		}
		done <- true
	}()

	wg.Wait()
	// If we reach here without panic, thread safety works
}

func TestParentChildRelationships(t *testing.T) {
	ss := NewSelectionState()

	// Test multiple levels of hierarchy
	ss.ExcludeFile("a")
	ss.ExcludeFile("a/b/c") // This should be redundant due to parent exclusion

	// All children should be excluded
	if !ss.IsPathExcluded("a/b") {
		t.Error("Direct child should be excluded")
	}
	if !ss.IsPathExcluded("a/b/c") {
		t.Error("Grandchild should be excluded")
	}
	if !ss.IsPathExcluded("a/b/c/d") {
		t.Error("Great-grandchild should be excluded")
	}

	// Reset and test explicit child inclusion
	ss.Reset()
	ss.ExcludeFile("a")
	// Child cannot override parent exclusion in this simple model
	// This is by design - parent exclusions are absolute

	if !ss.IsPathExcluded("a/b") {
		t.Error("Child should still be excluded when parent is excluded")
	}
}

func TestEdgeCases(t *testing.T) {
	ss := NewSelectionState()

	// Test root path exclusion
	ss.ExcludeFile(".")
	if !ss.IsPathExcluded("any/file/path.txt") {
		t.Error("All files should be excluded when root is excluded")
	}

	// Test empty path
	ss.Reset()
	if ss.IsPathExcluded("") {
		t.Error("Empty path should not be excluded by default")
	}

	// Test single file exclusion
	ss.ExcludeFile("file.txt")
	if !ss.IsPathExcluded("file.txt") {
		t.Error("Single file should be excluded")
	}

	// Test that parent directory doesn't affect single file in different location
	if ss.IsPathExcluded("dir/file.txt") {
		t.Error("File in different directory should not be affected")
	}
}

func TestGetExcludedFiles(t *testing.T) {
	ss := NewSelectionState()

	// Add some exclusions
	ss.ExcludeFile("src/components")
	ss.ExcludeFile("docs/readme.md")

	excluded := ss.GetExcludedFiles()

	if len(excluded) != 2 {
		t.Errorf("Expected 2 excluded files, got %d", len(excluded))
	}

	// Use filepath.Clean to normalize paths for comparison (cross-platform)
	cleanPath1 := filepath.Clean("src/components")
	cleanPath2 := filepath.Clean("docs/readme.md")

	if !excluded[cleanPath1] {
		t.Errorf("%s should be in excluded files", cleanPath1)
	}

	if !excluded[cleanPath2] {
		t.Errorf("%s should be in excluded files", cleanPath2)
	}
}

func TestReset(t *testing.T) {
	ss := NewSelectionState()

	// Add some exclusions
	ss.selection["src"] = StatusExcluded
	ss.selection["docs"] = StatusExcluded

	// Verify they exist
	if len(ss.selection) != 2 {
		t.Error("Should have 2 exclusions before reset")
	}

	// Reset
	ss.Reset()

	// Verify they're gone
	if len(ss.selection) != 0 {
		t.Error("Should have no exclusions after reset")
	}

	// Verify no paths are excluded
	if ss.IsPathExcluded("src/file.ts") {
		t.Error("No paths should be excluded after reset")
	}
}

// Performance benchmarks

func BenchmarkHierarchicalExclusion(b *testing.B) {
	ss := NewSelectionState()

	// Set up a parent exclusion using proper method
	ss.ExcludeFile("src/components")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This should be O(log d) where d is directory depth
		_ = ss.IsPathExcluded("src/components/ui/forms/deep/nested/Button.tsx")
	}
}

func BenchmarkToggleFile(b *testing.B) {
	ss := NewSelectionState()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("path%d", i%1000) // Cycle through 1000 paths
		ss.ToggleFile(path)
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	ss := NewSelectionState()

	// Pre-populate with some data using proper method
	for i := 0; i < 100; i++ {
		ss.ExcludeFile(fmt.Sprintf("dir%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				_ = ss.IsPathExcluded(fmt.Sprintf("dir%d/file.txt", i%100))
			} else {
				ss.ToggleFile(fmt.Sprintf("path%d", i))
			}
			i++
		}
	})
}

func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("HierarchicalModel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ss := NewSelectionState()

			// Simulate excluding a large directory tree with hierarchical model
			// Only 1 entry needed for entire directory tree
			ss.ExcludeFile("node_modules")

			// Verify it works for many children
			for j := 0; j < 1000; j++ {
				_ = ss.IsPathExcluded(fmt.Sprintf("node_modules/package%d/file.js", j))
			}
		}
	})

	b.Run("FlatModel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate old flat model with individual file tracking
			flatMap := make(map[string]bool)

			// Would need 1000 entries for same exclusion in flat model
			for j := 0; j < 1000; j++ {
				flatMap[fmt.Sprintf("node_modules/package%d/file.js", j)] = true
			}

			// Verify lookups
			for j := 0; j < 1000; j++ {
				_ = flatMap[fmt.Sprintf("node_modules/package%d/file.js", j)]
			}
		}
	})
}

func BenchmarkPathDepthScaling(b *testing.B) {
	ss := NewSelectionState()
	ss.ExcludeFile("a")

	depths := []int{1, 5, 10, 20}

	for _, depth := range depths {
		path := "a"
		for i := 1; i < depth; i++ {
			path = filepath.Join(path, fmt.Sprintf("level%d", i))
		}
		path = filepath.Join(path, "file.txt")

		b.Run(fmt.Sprintf("Depth%d", depth), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = ss.IsPathExcluded(path)
			}
		})
	}
}
