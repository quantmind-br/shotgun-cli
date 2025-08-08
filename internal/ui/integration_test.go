package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"shotgun-cli/internal/core"
)

// TestPatternConfigurationWorkflowIntegration tests the pattern functionality
// integration between UI components and core pattern processing
func TestPatternConfigurationWorkflowIntegration(t *testing.T) {
	// Create a temporary directory with test files
	tempDir := t.TempDir()

	// Create test file structure
	testFiles := []string{
		"important.log",             // Should be force-included despite log ignore
		"temp/cache.tmp",            // Should be custom ignored
		"build/output.js",           // Should be custom ignored
		"src/main.go",               // Regular source file
		"config/app.yml",            // Should be force-included
		"logs/debug.log",            // Should be custom ignored
		"docs/README.md",            // Regular documentation
		"node_modules/pkg/index.js", // Should be custom ignored
		"test.bak",                  // Should be custom ignored
	}

	// Create directories and files
	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
	}

	t.Run("Pattern Configuration UI and Core Integration", func(t *testing.T) {
		// Create enhanced config manager
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		// Verify initial state - no custom patterns configured
		initialConfig := configManager.GetEnhanced()
		assert.Empty(t, initialConfig.App.CustomIgnorePatterns)
		assert.Empty(t, initialConfig.App.ForceIncludePatterns)

		// Step 1: Create and test pattern configuration UI component
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)
		require.NotNil(t, patternConfig)

		// Step 2: Add custom ignore patterns through UI
		patterns := []string{
			"*.log",         // Ignore all log files
			"*.tmp",         // Ignore temp files
			"*.bak",         // Ignore backup files
			"build/",        // Ignore build directory
			"node_modules/", // Ignore node modules
		}

		for _, pattern := range patterns {
			patternConfig.startAdd()
			patternConfig.addInput.SetValue(pattern)
			patternConfig.addPattern()
		}

		// Step 3: Add force include patterns through UI (switch to force include tab)
		patternConfig.currentTab = 1
		forcePatterns := []string{
			"important.log", // Force include this specific log
			"config/*.yml",  // Force include config files
		}

		for _, pattern := range forcePatterns {
			patternConfig.startAdd()
			patternConfig.addInput.SetValue(pattern)
			patternConfig.addPattern()
		}

		// Step 4: Get updated configuration from UI component
		updatedConfig := patternConfig.GetConfig()

		// Verify patterns were added to configuration
		for _, pattern := range patterns {
			assert.Contains(t, updatedConfig.App.CustomIgnorePatterns, pattern)
		}
		for _, pattern := range forcePatterns {
			assert.Contains(t, updatedConfig.App.ForceIncludePatterns, pattern)
		}

		// Step 5: Apply configuration and test core integration
		err = configManager.UpdateEnhanced(updatedConfig)
		require.NoError(t, err)

		// Step 6: Test directory scanner integration with patterns
		scanner := core.NewDirectoryScanner()

		// Set patterns from configuration
		err = scanner.SetCustomIgnore(updatedConfig.App.CustomIgnorePatterns)
		require.NoError(t, err)

		err = scanner.SetForceIncludePatterns(updatedConfig.App.ForceIncludePatterns)
		require.NoError(t, err)

		// Scan directory with patterns applied
		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)

		// Step 7: Verify pattern application in scanned file tree
		verifyPatternApplicationInScannedTree(t, root, map[string]PatternExpectation{
			"important.log":                                  {ShouldBeVisible: true, ShouldBeIgnored: false}, // Force included
			filepath.Join("config", "app.yml"):               {ShouldBeVisible: true, ShouldBeIgnored: false}, // Force included
			filepath.Join("temp", "cache.tmp"):               {ShouldBeVisible: true, ShouldBeIgnored: true},  // Custom ignored
			filepath.Join("build", "output.js"):              {ShouldBeVisible: true, ShouldBeIgnored: true},  // Custom ignored
			filepath.Join("logs", "debug.log"):               {ShouldBeVisible: true, ShouldBeIgnored: true},  // Custom ignored
			"test.bak":                                       {ShouldBeVisible: true, ShouldBeIgnored: true},  // Custom ignored
			filepath.Join("node_modules", "pkg", "index.js"): {ShouldBeVisible: true, ShouldBeIgnored: true},  // Custom ignored
			filepath.Join("src", "main.go"):                  {ShouldBeVisible: true, ShouldBeIgnored: false}, // Not ignored
			filepath.Join("docs", "README.md"):               {ShouldBeVisible: true, ShouldBeIgnored: false}, // Not ignored
		})
	})

	t.Run("Pattern Modification and Persistence", func(t *testing.T) {
		// Create config manager
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		// Set initial patterns
		initialConfig := configManager.GetEnhanced()
		initialConfig.App.CustomIgnorePatterns = []string{"*.tmp", "build/"}
		initialConfig.App.ForceIncludePatterns = []string{"important.log"}
		err = configManager.UpdateEnhanced(initialConfig)
		require.NoError(t, err)

		// Create pattern config UI with initial patterns
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)
		require.NotNil(t, patternConfig)

		// Verify initial patterns are loaded in UI
		assert.Len(t, patternConfig.config.App.CustomIgnorePatterns, 2)
		assert.Len(t, patternConfig.config.App.ForceIncludePatterns, 1)

		// Modify patterns - remove one ignore pattern
		patternConfig.currentTab = 0 // Custom ignore tab
		for i, p := range patternConfig.config.App.CustomIgnorePatterns {
			if p == "*.tmp" {
				patternConfig.customPatterns.Select(i)
				patternConfig.deleteSelected()
				break
			}
		}

		// Add one force include pattern
		patternConfig.currentTab = 1 // Force include tab
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("config/*.yml")
		patternConfig.addPattern()

		// Get updated configuration
		updatedConfig := patternConfig.GetConfig()
		err = configManager.UpdateEnhanced(updatedConfig)
		require.NoError(t, err)

		// Verify configuration persistence
		finalConfig := configManager.GetEnhanced()
		assert.NotContains(t, finalConfig.App.CustomIgnorePatterns, "*.tmp")
		assert.Contains(t, finalConfig.App.CustomIgnorePatterns, "build/")
		assert.Contains(t, finalConfig.App.ForceIncludePatterns, "important.log")
		assert.Contains(t, finalConfig.App.ForceIncludePatterns, "config/*.yml")

		// Test with directory scanner
		scanner := core.NewDirectoryScanner()
		err = scanner.SetCustomIgnore(finalConfig.App.CustomIgnorePatterns)
		require.NoError(t, err)
		err = scanner.SetForceIncludePatterns(finalConfig.App.ForceIncludePatterns)
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)

		// Verify pattern changes:
		// - temp/cache.tmp is STILL ignored because *.tmp is in the DEFAULT ignore patterns (not custom)
		// - config/app.yml should now be force included (overriding any ignore patterns)
		// - build/output.js should still be ignored (build/ pattern remains)
		verifyPatternApplicationInScannedTree(t, root, map[string]PatternExpectation{
			filepath.Join("temp", "cache.tmp"):  {ShouldBeVisible: true, ShouldBeIgnored: true},  // Still ignored by default *.tmp pattern
			filepath.Join("config", "app.yml"):  {ShouldBeVisible: true, ShouldBeIgnored: false}, // Force included
			filepath.Join("build", "output.js"): {ShouldBeVisible: true, ShouldBeIgnored: true},  // Still ignored by build/
		})
	})

	t.Run("Pattern Validation Integration", func(t *testing.T) {
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		initialConfig := configManager.GetEnhanced()
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)
		require.NotNil(t, patternConfig)

		// Verify validation is enabled by default
		assert.True(t, patternConfig.validationEnabled)

		// Try to add invalid pattern (empty pattern)
		// Note: The current UI implementation returns early for empty patterns without setting validation error
		initialPatternCount := len(patternConfig.config.App.CustomIgnorePatterns)
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("") // Empty pattern
		patternConfig.addPattern()

		// Verify empty pattern was not added (returns early without validation error)
		assert.Equal(t, initialPatternCount, len(patternConfig.config.App.CustomIgnorePatterns))
		assert.True(t, patternConfig.isAddingPattern) // Should remain in add mode due to early return (cancelAdd() not called)

		// Cancel add mode and try with valid pattern
		patternConfig.cancelAdd()
		assert.False(t, patternConfig.isAddingPattern)

		// Add valid pattern
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("*.valid")
		patternConfig.addPattern()

		// Verify valid pattern was added successfully
		assert.Empty(t, patternConfig.validationError)
		assert.False(t, patternConfig.isAddingPattern)
		assert.Contains(t, patternConfig.config.App.CustomIgnorePatterns, "*.valid")

		// Test validation toggle
		patternConfig.validationEnabled = false
		assert.False(t, patternConfig.validationEnabled)

		// With validation disabled, should be able to add patterns without validation
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("any-pattern")
		patternConfig.addPattern()

		// Should succeed even with potentially invalid pattern when validation disabled
		assert.False(t, patternConfig.isAddingPattern)
		assert.Contains(t, patternConfig.config.App.CustomIgnorePatterns, "any-pattern")
	})

	t.Run("Configuration Persistence and State Management", func(t *testing.T) {
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		// Test 1: Create pattern config, add patterns, save configuration
		initialConfig := configManager.GetEnhanced()
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)

		// Add test pattern
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("*.test")
		patternConfig.addPattern()

		// Save configuration
		updatedConfig := patternConfig.GetConfig()
		err = configManager.UpdateEnhanced(updatedConfig)
		require.NoError(t, err)
		err = configManager.Save()
		require.NoError(t, err)

		// Test 2: Create new config manager and verify persistence
		newConfigManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)
		err = newConfigManager.Load()
		require.NoError(t, err)

		loadedConfig := newConfigManager.GetEnhanced()
		assert.Contains(t, loadedConfig.App.CustomIgnorePatterns, "*.test")

		// Test 3: Create new pattern config UI with loaded state
		newPatternConfig := NewPatternConfigModel(loadedConfig, 100, 30)
		assert.Contains(t, newPatternConfig.config.App.CustomIgnorePatterns, "*.test")

		// Test 4: Verify state independence - modify new config without affecting original
		originalPatternsCount := len(loadedConfig.App.CustomIgnorePatterns)

		newPatternConfig.startAdd()
		newPatternConfig.addInput.SetValue("*.new")
		newPatternConfig.addPattern()

		// Original config should be unchanged
		assert.Equal(t, originalPatternsCount, len(loadedConfig.App.CustomIgnorePatterns))

		// New config should have additional pattern
		newConfigFromUI := newPatternConfig.GetConfig()
		assert.Equal(t, originalPatternsCount+1, len(newConfigFromUI.App.CustomIgnorePatterns))
		assert.Contains(t, newConfigFromUI.App.CustomIgnorePatterns, "*.new")
	})
}

// PatternExpectation defines what we expect for a file in pattern testing
type PatternExpectation struct {
	ShouldBeVisible bool
	ShouldBeIgnored bool
}

// Helper functions for integration testing

func verifyPatternApplicationInScannedTree(t *testing.T, root *core.FileNode, expectations map[string]PatternExpectation) {
	// Walk through the file tree and verify each file meets expectations
	walkFileTreeForVerification(t, root, "", expectations)
}

func walkFileTreeForVerification(t *testing.T, node *core.FileNode, currentPath string, expectations map[string]PatternExpectation) {
	if node == nil {
		return
	}

	// Build the relative path for this node
	var nodePath string
	if currentPath == "" {
		nodePath = node.RelPath
	} else {
		nodePath = node.RelPath
	}

	// Check if this file has expectations
	if expectation, exists := expectations[nodePath]; exists {
		if !node.IsDir {
			// For files, verify ignore status matches expectation
			actuallyIgnored := node.IsGitignored || node.IsCustomIgnored
			assert.Equal(t, expectation.ShouldBeIgnored, actuallyIgnored,
				"File %s ignore status mismatch. Expected ignored: %v, Actually ignored: %v (IsGitignored: %v, IsCustomIgnored: %v)",
				nodePath, expectation.ShouldBeIgnored, actuallyIgnored, node.IsGitignored, node.IsCustomIgnored)
		}
	}

	// Recurse through children
	for _, child := range node.Children {
		walkFileTreeForVerification(t, child, nodePath, expectations)
	}
}

// TestPatternConfigurationEdgeCases tests edge cases in UI integration
func TestPatternConfigurationEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Empty Directory Pattern Configuration", func(t *testing.T) {
		// Test pattern configuration with empty directory
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		initialConfig := configManager.GetEnhanced()
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)

		// Should be able to configure patterns even with empty directory
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("*.tmp")
		patternConfig.addPattern()

		updatedConfig := patternConfig.GetConfig()
		err = configManager.UpdateEnhanced(updatedConfig)
		require.NoError(t, err)

		// Verify configuration was saved
		finalConfig := configManager.GetEnhanced()
		assert.Contains(t, finalConfig.App.CustomIgnorePatterns, "*.tmp")

		// Test scanning empty directory with patterns
		scanner := core.NewDirectoryScanner()
		err = scanner.SetCustomIgnore(finalConfig.App.CustomIgnorePatterns)
		require.NoError(t, err)

		ctx := context.Background()
		root, err := scanner.ScanDirectory(ctx, tempDir)
		require.NoError(t, err)
		require.NotNil(t, root)
	})

	t.Run("Large Number of Patterns", func(t *testing.T) {
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		initialConfig := configManager.GetEnhanced()
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)

		// Add many patterns to test UI performance and stability
		var manyPatterns []string
		for i := 0; i < 50; i++ {
			pattern := "*.ext" + string(rune(97+i%26)) // *.exta, *.extb, etc.
			manyPatterns = append(manyPatterns, pattern)

			patternConfig.startAdd()
			patternConfig.addInput.SetValue(pattern)
			patternConfig.addPattern()
		}

		updatedConfig := patternConfig.GetConfig()
		err = configManager.UpdateEnhanced(updatedConfig)
		require.NoError(t, err)

		// Verify all patterns were saved
		finalConfig := configManager.GetEnhanced()
		assert.Len(t, finalConfig.App.CustomIgnorePatterns, 50)

		// Verify all expected patterns are present
		for _, pattern := range manyPatterns {
			assert.Contains(t, finalConfig.App.CustomIgnorePatterns, pattern)
		}
	})

	t.Run("Rapid Pattern Modifications", func(t *testing.T) {
		configManager, err := core.NewEnhancedConfigManager()
		require.NoError(t, err)

		initialConfig := configManager.GetEnhanced()
		patternConfig := NewPatternConfigModel(initialConfig, 100, 30)

		// Simulate rapid add/remove operations
		for i := 0; i < 10; i++ {
			pattern := "*.rapid" + string(rune(97+i))

			// Add pattern
			patternConfig.startAdd()
			patternConfig.addInput.SetValue(pattern)
			patternConfig.addPattern()

			// Remove pattern - find and delete
			for j, p := range patternConfig.config.App.CustomIgnorePatterns {
				if p == pattern {
					patternConfig.customPatterns.Select(j)
					patternConfig.deleteSelected()
					break
				}
			}
		}

		// Add final pattern
		patternConfig.startAdd()
		patternConfig.addInput.SetValue("*.final")
		patternConfig.addPattern()

		// Save configuration
		updatedConfig := patternConfig.GetConfig()
		err = configManager.UpdateEnhanced(updatedConfig)
		require.NoError(t, err)

		// Only the final pattern should remain
		finalConfig := configManager.GetEnhanced()
		assert.Len(t, finalConfig.App.CustomIgnorePatterns, 1)
		assert.Contains(t, finalConfig.App.CustomIgnorePatterns, "*.final")
	})
}

// TestPatternConfigurationPerformance benchmarks pattern configuration operations
func TestPatternConfigurationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()

	// Create large file structure for performance testing
	for i := 0; i < 100; i++ {
		for j := 0; j < 10; j++ {
			dirName := fmt.Sprintf("dir%03d", i) // Use numeric directory names to avoid invalid characters
			fileName := fmt.Sprintf("file%03d.txt", j)
			filePath := filepath.Join(tempDir, dirName, fileName)
			err := os.MkdirAll(filepath.Dir(filePath), 0755)
			require.NoError(t, err)
			err = os.WriteFile(filePath, []byte("content"), 0644)
			require.NoError(t, err)
		}
	}

	configManager, err := core.NewEnhancedConfigManager()
	require.NoError(t, err)

	// Measure pattern configuration UI performance
	start := time.Now()
	initialConfig := configManager.GetEnhanced()
	patternConfig := NewPatternConfigModel(initialConfig, 100, 30)

	// Add patterns that won't ignore all the test files
	patterns := []string{"*.log", "build/", "temp/"} // Use patterns that won't match our .txt test files
	for _, pattern := range patterns {
		patternConfig.startAdd()
		patternConfig.addInput.SetValue(pattern)
		patternConfig.addPattern()
	}

	updatedConfig := patternConfig.GetConfig()
	err = configManager.UpdateEnhanced(updatedConfig)
	require.NoError(t, err)
	configTime := time.Since(start)

	t.Logf("Pattern configuration took: %v", configTime)
	assert.Less(t, configTime, 1*time.Second, "Pattern configuration should complete within 1 second")

	// Measure directory scanning performance with patterns
	start = time.Now()
	scanner := core.NewDirectoryScanner()
	err = scanner.SetCustomIgnore(updatedConfig.App.CustomIgnorePatterns)
	require.NoError(t, err)

	ctx := context.Background()
	root, err := scanner.ScanDirectory(ctx, tempDir)
	require.NoError(t, err)
	require.NotNil(t, root)
	scanTime := time.Since(start)

	t.Logf("Directory scanning with patterns took: %v", scanTime)
	assert.Less(t, scanTime, 5*time.Second, "Directory scanning should complete within 5 seconds")

	// Count files to verify scan worked
	fileCount := countFilesInTree(root)
	t.Logf("Found %d files in scanned tree", fileCount)
	assert.Greater(t, fileCount, 0, "Should find files in directory")
}

// Helper function to count files in tree
func countFilesInTree(node *core.FileNode) int {
	if node == nil {
		return 0
	}

	count := 0
	if !node.IsDir {
		count = 1
	}

	for _, child := range node.Children {
		count += countFilesInTree(child)
	}

	return count
}
