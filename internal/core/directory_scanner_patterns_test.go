package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDirectoryScannerPatterns tests the DirectoryScanner pattern functionality
func TestDirectoryScannerPatterns(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test file structure
	testFiles := []string{
		"important.log",             // Should be force-included
		"temp/cache.tmp",            // Should be custom ignored
		"build/output.js",           // Should be custom ignored
		"src/main.go",               // Regular source file
		"config/app.yml",            // Should be force-included
		"logs/debug.log",            // Should be custom ignored
		"docs/README.md",            // Regular documentation
		"node_modules/pkg/index.js", // Should be custom ignored
	}

	// Create directories and files
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	t.Run("Default Scanner Behavior", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Scan without any custom patterns
		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Count all files found
		fileCount := countFiles(root)
		assert.Greater(t, fileCount, 0, "Should find files with default scanner")
	})

	t.Run("Custom Ignore Patterns", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Set custom ignore patterns
		customIgnorePatterns := []string{
			"*.tmp",         // Ignore temp files
			"*.log",         // Ignore log files
			"build/",        // Ignore build directory
			"node_modules/", // Ignore node modules
		}

		err := scanner.SetCustomIgnore(customIgnorePatterns)
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Verify that custom ignored files are marked
		foundFiles := collectAllFiles(root)
		for _, file := range foundFiles {
			if matchesPatterns(file.RelPath, customIgnorePatterns) {
				assert.True(t, file.IsCustomIgnored, "File %s should be custom ignored", file.RelPath)
			}
		}
	})

	t.Run("Force Include Patterns", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Set custom ignore patterns that would normally exclude these files
		customIgnorePatterns := []string{
			"*.log", // This would ignore important.log
			"*.yml", // This would ignore config/app.yml
		}
		err := scanner.SetCustomIgnore(customIgnorePatterns)
		require.NoError(t, err)

		// Set force include patterns to override ignores
		forceIncludePatterns := []string{
			"important.log", // Force include this specific log
			"config/*.yml",  // Force include config YAML files
		}
		err = scanner.SetForceIncludePatterns(forceIncludePatterns)
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Verify that force included files are not marked as ignored
		foundFiles := collectAllFiles(root)
		for _, file := range foundFiles {
			if file.RelPath == "important.log" || file.RelPath == filepath.Join("config", "app.yml") {
				assert.False(t, file.IsCustomIgnored, "Force included file %s should not be marked as ignored", file.RelPath)
			}
		}
	})

	t.Run("Combined Patterns Priority", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Aggressive ignore patterns
		customIgnorePatterns := []string{
			"*.log",
			"*.tmp",
			"*.yml",
			"build/",
			"node_modules/",
		}
		err := scanner.SetCustomIgnore(customIgnorePatterns)
		require.NoError(t, err)

		// Strategic force include patterns
		forceIncludePatterns := []string{
			"important.log", // Override log ignore for this specific file
			"config/*.yml",  // Override yml ignore for config directory
		}
		err = scanner.SetForceIncludePatterns(forceIncludePatterns)
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		foundFiles := collectAllFiles(root)

		// Verify priority: force include > custom ignore > gitignore > default
		for _, file := range foundFiles {
			switch file.RelPath {
			case "important.log":
				// Should NOT be ignored despite matching *.log pattern
				assert.False(t, file.IsCustomIgnored, "important.log should be force included")
			case filepath.Join("config", "app.yml"):
				// Should NOT be ignored despite matching *.yml pattern
				assert.False(t, file.IsCustomIgnored, "config/app.yml should be force included")
			case filepath.Join("logs", "debug.log"):
				// Should be ignored by *.log pattern and NOT force included
				assert.True(t, file.IsCustomIgnored, "logs/debug.log should be custom ignored")
			case filepath.Join("temp", "cache.tmp"):
				// Should be ignored by *.tmp pattern
				assert.True(t, file.IsCustomIgnored, "temp/cache.tmp should be custom ignored")
			}
		}
	})

	t.Run("Invalid Pattern Handling", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Test with invalid patterns (empty strings, malformed patterns)
		invalidPatterns := []string{
			"",         // Empty pattern
			"*.tmp",    // Valid pattern
			"[invalid", // Malformed bracket pattern
			"build/",   // Valid pattern
		}

		// SetCustomIgnore should handle invalid patterns gracefully
		err := scanner.SetCustomIgnore(invalidPatterns)
		// Should not error for patterns that gitignore can't parse, it filters them out
		assert.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Scanner should still work despite invalid patterns
		fileCount := countFiles(root)
		assert.Greater(t, fileCount, 0, "Scanner should work despite invalid patterns")
	})

	t.Run("Empty Pattern Lists", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Test with empty pattern lists
		err := scanner.SetCustomIgnore([]string{})
		require.NoError(t, err)

		err = scanner.SetForceIncludePatterns([]string{})
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Should behave like default scanner (but default patterns may still apply)
		foundFiles := collectAllFiles(root)

		// Check that files matching default patterns are still ignored
		// but files not matching default patterns are not custom ignored
		for _, file := range foundFiles {
			if !matchesDefaultPatterns(file.RelPath) {
				assert.False(t, file.IsCustomIgnored, "Non-default pattern file %s should not be custom ignored with empty custom patterns", file.RelPath)
			}
		}
	})

	t.Run("Complex Real World Patterns", func(t *testing.T) {
		scanner := NewDirectoryScanner()

		// Real world web development ignore patterns
		customIgnorePatterns := []string{
			"node_modules/", // Dependencies
			"*.log",         // Log files
			"*.tmp",         // Temporary files
			"build/",        // Build output
			"dist/",         // Distribution
			".env.local",    // Local environment
			"coverage/",     // Test coverage
			"*.cache",       // Cache files
		}

		// Important files to force include despite ignores
		forceIncludePatterns := []string{
			"important.log", // Critical log file
			"config/*.yml",  // Configuration files
			"scripts/*.js",  // Build scripts
			"docs/**/*.md",  // All documentation
		}

		err := scanner.SetCustomIgnore(customIgnorePatterns)
		require.NoError(t, err)

		err = scanner.SetForceIncludePatterns(forceIncludePatterns)
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Verify the filtering is working
		foundFiles := collectAllFiles(root)
		assert.Greater(t, len(foundFiles), 0, "Should find files even with complex patterns")

		// Check specific cases
		for _, file := range foundFiles {
			t.Logf("File: %s, IsCustomIgnored: %v", file.RelPath, file.IsCustomIgnored)
		}
	})
}

// TestDirectoryScannerPatternConcurrency tests pattern functionality under concurrent access
func TestDirectoryScannerPatternConcurrency(t *testing.T) {
	tempDir := t.TempDir()

	// Create some test files
	testFiles := []string{
		"test1.tmp",
		"test2.log",
		"important.txt",
		"build/output.js",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("content"), 0644)
		require.NoError(t, err)
	}

	scanner := NewDirectoryScanner()

	// Set initial patterns
	err := scanner.SetCustomIgnore([]string{"*.tmp"})
	require.NoError(t, err)

	// Concurrent scanning with pattern updates
	done := make(chan bool, 10)

	// Start concurrent scans
	for i := 0; i < 5; i++ {
		go func(index int) {
			defer func() { done <- true }()

			ctx := context.Background()
			root, err := scanner.ScanDirectory(ctx, tempDir)
			assert.NoError(t, err)
			assert.NotNil(t, root)
		}(i)
	}

	// Start concurrent pattern updates
	for i := 0; i < 5; i++ {
		go func(index int) {
			defer func() { done <- true }()

			patterns := []string{"*.log", "build/"}
			err := scanner.SetCustomIgnore(patterns)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper functions

// countFiles recursively counts all files in a file tree
func countFiles(node *FileNode) int {
	if node == nil {
		return 0
	}

	count := 0
	if !node.IsDir {
		count = 1
	}

	for _, child := range node.Children {
		count += countFiles(child)
	}

	return count
}

// collectAllFiles collects all file nodes from a tree into a flat slice
func collectAllFiles(node *FileNode) []*FileNode {
	if node == nil {
		return nil
	}

	var files []*FileNode

	if !node.IsDir {
		files = append(files, node)
	}

	for _, child := range node.Children {
		files = append(files, collectAllFiles(child)...)
	}

	return files
}

// matchesPatterns checks if a path matches any of the given patterns (simple implementation)
func matchesPatterns(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		// Simple pattern matching for testing
		if pattern == path {
			return true
		}

		// Handle basic wildcards for directories
		if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
			dir := pattern[:len(pattern)-1]
			if filepath.Dir(path) == dir || path == dir {
				return true
			}
		}

		// Handle basic file extension patterns
		if len(pattern) > 1 && pattern[0] == '*' && pattern[1] == '.' {
			ext := pattern[1:]
			if filepath.Ext(path) == ext {
				return true
			}
		}
	}

	return false
}

// matchesDefaultPatterns checks if a path would be ignored by default patterns
func matchesDefaultPatterns(path string) bool {
	// Simple check for common default patterns used in tests
	defaultPatterns := []string{
		"*.tmp", "*.temp", "*.bak", "*.swp", "*.swo",
		"tmp/", "cache/", "node_modules/", "vendor/",
		".idea/", ".vscode/", ".vs/", "__pycache__/",
		"*.exe", "*.dll", "*.so", "*.dylib", "*.bin",
		"*.zip", "*.rar", "*.7z", "*.tar", "*.gz",
		"*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp",
		"*.mp3", "*.wav", "*.ogg", "*.aac", "*.flac",
		"*.mp4", "*.avi", "*.mov", "*.wmv", "*.mkv",
		"*.sqlite", "*.sqlite3", "*.db", "*.mdb",
	}

	return matchesPatterns(path, defaultPatterns)
}
