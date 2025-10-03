package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileNodeBasic(t *testing.T) {
	tests := []struct {
		name     string
		node     *FileNode
		expected map[string]interface{}
	}{
		{
			name: "basic file node",
			node: &FileNode{
				Name:            "test.txt",
				Path:            "/home/user/test.txt",
				RelPath:         "test.txt",
				IsDir:           false,
				Selected:        false,
				IsGitignored:    false,
				IsCustomIgnored: false,
				Size:            1024,
				Expanded:        false,
			},
			expected: map[string]interface{}{
				"name":              "test.txt",
				"path":              "/home/user/test.txt",
				"rel_path":          "test.txt",
				"is_dir":            false,
				"selected":          false,
				"is_gitignored":     false,
				"is_custom_ignored": false,
				"size":              float64(1024),
				"expanded":          false,
			},
		},
		{
			name: "directory node",
			node: &FileNode{
				Name:            "src",
				Path:            "/home/user/src",
				RelPath:         "src",
				IsDir:           true,
				Selected:        true,
				IsGitignored:    true,
				IsCustomIgnored: false,
				Size:            0,
				Expanded:        true,
			},
			expected: map[string]interface{}{
				"name":              "src",
				"path":              "/home/user/src",
				"rel_path":          "src",
				"is_dir":            true,
				"selected":          true,
				"is_gitignored":     true,
				"is_custom_ignored": false,
				"size":              float64(0),
				"expanded":          true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			data, err := json.Marshal(tt.node)
			if err != nil {
				t.Fatalf("Failed to marshal FileNode: %v", err)
			}

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for key, expected := range tt.expected {
				if result[key] != expected {
					t.Errorf("Expected %s to be %v, got %v", key, expected, result[key])
				}
			}
		})
	}
}

func TestFileNodeMethods(t *testing.T) {
	// Create a test file tree
	root := &FileNode{
		Name:     "root",
		IsDir:    true,
		Children: []*FileNode{},
	}

	dir1 := &FileNode{
		Name:     "dir1",
		IsDir:    true,
		Parent:   root,
		Children: []*FileNode{},
	}

	file1 := &FileNode{
		Name:   "file1.txt",
		IsDir:  false,
		Parent: dir1,
		Size:   100,
	}

	file2 := &FileNode{
		Name:   "file2.txt",
		IsDir:  false,
		Parent: dir1,
		Size:   200,
	}

	dir2 := &FileNode{
		Name:     "dir2",
		IsDir:    true,
		Parent:   root,
		Children: []*FileNode{},
	}

	file3 := &FileNode{
		Name:            "file3.txt",
		IsDir:           false,
		Parent:          dir2,
		Size:            300,
		IsGitignored:    true,
		IsCustomIgnored: false,
	}

	// Build the tree
	dir1.Children = []*FileNode{file1, file2}
	dir2.Children = []*FileNode{file3}
	root.Children = []*FileNode{dir1, dir2}

	tests := []struct {
		name     string
		node     *FileNode
		method   string
		expected interface{}
	}{
		{"root count children", root, "CountChildren", 5},
		{"root count files", root, "CountFiles", 3},
		{"root count directories", root, "CountDirectories", 3},
		{"root total size", root, "TotalSize", int64(600)},
		{"dir1 count children", dir1, "CountChildren", 2},
		{"dir1 count files", dir1, "CountFiles", 2},
		{"dir1 count directories", dir1, "CountDirectories", 1},
		{"dir1 total size", dir1, "TotalSize", int64(300)},
		{"file1 count children", file1, "CountChildren", 0},
		{"file1 count files", file1, "CountFiles", 1},
		{"file1 count directories", file1, "CountDirectories", 0},
		{"file1 total size", file1, "TotalSize", int64(100)},
		{"file3 is ignored", file3, "IsIgnored", true},
		{"file1 is ignored", file1, "IsIgnored", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			switch tt.method {
			case "CountChildren":
				result = tt.node.CountChildren()
			case "CountFiles":
				result = tt.node.CountFiles()
			case "CountDirectories":
				result = tt.node.CountDirectories()
			case "TotalSize":
				result = tt.node.TotalSize()
			case "IsIgnored":
				result = tt.node.IsIgnored()
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIgnoreReasons(t *testing.T) {
	tests := []struct {
		name            string
		isGitignored    bool
		isCustomIgnored bool
		expected        string
	}{
		{"not ignored", false, false, "not ignored"},
		{"gitignored only", true, false, "gitignored"},
		{"custom ignored only", false, true, "custom ignored"},
		{"both ignored", true, true, "gitignored and custom ignored"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &FileNode{
				IsGitignored:    tt.isGitignored,
				IsCustomIgnored: tt.isCustomIgnored,
			}

			result := node.GetIgnoreReason()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero bytes", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 1536, "1.5 KB"},
		{"megabytes", 2097152, "2.0 MB"},
		{"gigabytes", 3221225472, "3.0 GB"},
		{"large file", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestProgress(t *testing.T) {
	tests := []struct {
		name       string
		progress   Progress
		percentage float64
		str        string
	}{
		{
			name: "basic progress",
			progress: Progress{
				Current: 50,
				Total:   100,
				Stage:   "scanning",
				Message: "Processing files",
			},
			percentage: 50.0,
			str:        "50.0% (50/100) - scanning: Processing files",
		},
		{
			name: "progress without message",
			progress: Progress{
				Current: 25,
				Total:   100,
				Stage:   "counting",
			},
			percentage: 25.0,
			str:        "25.0% (25/100) - counting",
		},
		{
			name: "zero total",
			progress: Progress{
				Current: 10,
				Total:   0,
				Stage:   "init",
			},
			percentage: 0.0,
			str:        "0.0% (10/0) - init",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.progress.Percentage(); got != tt.percentage {
				t.Errorf("Expected percentage %f, got %f", tt.percentage, got)
			}

			if got := tt.progress.String(); got != tt.str {
				t.Errorf("Expected string %q, got %q", tt.str, got)
			}
		})
	}
}

func TestDefaultScanConfig(t *testing.T) {
	config := DefaultScanConfig()

	if config.MaxFileSize != 0 {
		t.Errorf("Expected MaxFileSize to be 0, got %d", config.MaxFileSize)
	}

	if config.MaxFiles != 0 {
		t.Errorf("Expected MaxFiles to be 0, got %d", config.MaxFiles)
	}

	if config.MaxMemory != 0 {
		t.Errorf("Expected MaxMemory to be 0, got %d", config.MaxMemory)
	}

	if config.SkipBinary != false {
		t.Errorf("Expected SkipBinary to be false, got %t", config.SkipBinary)
	}

	if config.IncludeHidden != false {
		t.Errorf("Expected IncludeHidden to be false, got %t", config.IncludeHidden)
	}

	if config.Workers != 1 {
		t.Errorf("Expected Workers to be 1, got %d", config.Workers)
	}
}

func TestFileSystemScanner(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure
	testFiles := []string{
		"file1.txt",
		"file2.go",
		"subdir/file3.py",
		"subdir/file4.js",
		"subdir/nested/file5.md",
		".hidden.txt",
		".gitignore",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}

		content := "test content for " + file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Create .gitignore content
	gitignoreContent := "*.tmp\n.env\nsubdir/file3.py\n"
	err = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	scanner := NewFileSystemScanner()
	config := DefaultScanConfig()

	// Test basic scanning
	t.Run("basic scan", func(t *testing.T) {
		root, err := scanner.Scan(tempDir, config)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if root == nil {
			t.Fatal("Expected non-nil root node")
		}

		if !root.IsDir {
			t.Error("Expected root to be a directory")
		}

		if len(root.Children) == 0 {
			t.Error("Expected root to have children")
		}
	})

	// Test scanning with progress
	t.Run("scan with progress", func(t *testing.T) {
		progress := make(chan Progress, 100)

		go func() {
			_, err := scanner.ScanWithProgress(tempDir, config, progress)
			if err != nil {
				t.Errorf("ScanWithProgress failed: %v", err)
			}
			close(progress)
		}()

		var progressUpdates []Progress
		for p := range progress {
			progressUpdates = append(progressUpdates, p)
		}

		if len(progressUpdates) == 0 {
			t.Error("Expected at least one progress update")
		}

		// Check first and last progress
		first := progressUpdates[0]
		if first.Stage != "scanning" {
			t.Errorf("Expected first stage to be 'scanning', got %q", first.Stage)
		}

		last := progressUpdates[len(progressUpdates)-1]
		if last.Stage != "complete" {
			t.Errorf("Expected last stage to be 'complete', got %q", last.Stage)
		}

		if last.Current != last.Total {
			t.Errorf("Expected final progress to be complete: %d/%d", last.Current, last.Total)
		}
	})

	// Test with hidden files exclusion
	t.Run("exclude hidden files", func(t *testing.T) {
		config := DefaultScanConfig()
		config.IncludeHidden = false

		root, err := scanner.Scan(tempDir, config)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Check that hidden files are marked as ignored
		var hiddenFound bool
		for _, child := range root.Children {
			if strings.HasPrefix(child.Name, ".") {
				hiddenFound = true
				break
			}
		}

		// Hidden files should not be in the tree when IncludeHidden is false
		if hiddenFound {
			t.Error("Hidden files should be excluded when IncludeHidden is false")
		}
	})

	// Test file size limit
	t.Run("file size limit", func(t *testing.T) {
		config := DefaultScanConfig()
		config.MaxFileSize = 10 // Very small limit

		root, err := scanner.Scan(tempDir, config)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// Count files in the result
		fileCount := root.CountFiles()
		if fileCount > 2 { // Should exclude most files due to size limit
			t.Errorf("Expected few files due to size limit, got %d", fileCount)
		}
	})

	// Test gitignore functionality
	t.Run("gitignore rules", func(t *testing.T) {
		err := scanner.LoadGitIgnore(tempDir)
		if err != nil {
			t.Fatalf("Failed to load gitignore: %v", err)
		}

		root, err := scanner.Scan(tempDir, config)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		// The scan should complete without the gitignored files
		var foundIgnoredFile bool
		var checkNode func(*FileNode)
		checkNode = func(node *FileNode) {
			if node.RelPath == "subdir/file3.py" {
				foundIgnoredFile = true
			}
			for _, child := range node.Children {
				checkNode(child)
			}
		}
		checkNode(root)

		// Since we exclude ignored files during scanning, we shouldn't find the ignored file
		if foundIgnoredFile {
			t.Error("Gitignored file should not be included in scan results")
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkFileSystemScanner(b *testing.B) {
	// Create a larger test directory for benchmarking
	tempDir, err := os.MkdirTemp("", "scanner_benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a reasonable number of files for benchmarking
	numDirs := 10
	numFilesPerDir := 50

	for i := 0; i < numDirs; i++ {
		dirPath := filepath.Join(tempDir, "dir"+string(rune('0'+i)))
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			b.Fatalf("Failed to create directory: %v", err)
		}

		for j := 0; j < numFilesPerDir; j++ {
			fileName := filepath.Join(dirPath, "file"+string(rune('0'+j))+".txt")
			content := "benchmark test content " + string(rune('0'+j))
			err := os.WriteFile(fileName, []byte(content), 0644)
			if err != nil {
				b.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	scanner := NewFileSystemScanner()
	config := DefaultScanConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.Scan(tempDir, config)
		if err != nil {
			b.Fatalf("Scan failed: %v", err)
		}
	}
}

func TestNewFileSystemScannerWithIgnore(t *testing.T) {
	tests := []struct {
		name        string
		ignoreRules []string
		expectError bool
	}{
		{
			name:        "empty rules",
			ignoreRules: []string{},
			expectError: false,
		},
		{
			name:        "valid rules",
			ignoreRules: []string{"*.tmp", "*.log", "node_modules/"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, err := NewFileSystemScannerWithIgnore(tt.ignoreRules)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if scanner == nil {
				t.Error("Expected non-nil scanner")
			}
		})
	}
}

func TestHiddenFileConsistencyWithIgnoreEngine(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "hidden_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test structure with hidden files
	testFiles := []string{
		"normal.txt",
		".hidden.txt",
		".hiddendir/file.txt",
		"subdir/.hidden_in_subdir.txt",
	}

	for _, file := range testFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}

		content := "test content for " + file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	t.Run("hidden files excluded with ignore engine and IncludeHidden=false", func(t *testing.T) {
		scanner, err := NewFileSystemScannerWithIgnore([]string{"*.tmp"}) // Basic ignore rule
		if err != nil {
			t.Fatalf("Failed to create scanner with ignore: %v", err)
		}

		config := DefaultScanConfig()
		config.IncludeHidden = false

		// Count items
		count, err := scanner.countItems(tempDir, config)
		if err != nil {
			t.Fatalf("countItems failed: %v", err)
		}

		// Build tree
		root, buildCount, err := scanner.walkAndBuild(tempDir, config, nil, count)
		if err != nil {
			t.Fatalf("walkAndBuild failed: %v", err)
		}

		// Count and build passes should match
		if count != buildCount {
			t.Errorf("Count mismatch: countItems=%d, walkAndBuild=%d", count, buildCount)
		}

		// Verify no hidden files are present in the tree
		var checkHiddenFiles func(*FileNode) []string
		checkHiddenFiles = func(node *FileNode) []string {
			var hiddenFiles []string
			for _, child := range node.Children {
				if strings.HasPrefix(child.Name, ".") && child.Name != "." && child.Name != ".." {
					hiddenFiles = append(hiddenFiles, child.RelPath)
				}
				hiddenFiles = append(hiddenFiles, checkHiddenFiles(child)...)
			}
			return hiddenFiles
		}

		hiddenFiles := checkHiddenFiles(root)
		if len(hiddenFiles) > 0 {
			t.Errorf("Found hidden files in result when IncludeHidden=false: %v", hiddenFiles)
		}
	})

	t.Run("hidden files included with ignore engine and IncludeHidden=true", func(t *testing.T) {
		scanner, err := NewFileSystemScannerWithIgnore([]string{"*.tmp"}) // Basic ignore rule
		if err != nil {
			t.Fatalf("Failed to create scanner with ignore: %v", err)
		}

		config := DefaultScanConfig()
		config.IncludeHidden = true

		// Count items
		count, err := scanner.countItems(tempDir, config)
		if err != nil {
			t.Fatalf("countItems failed: %v", err)
		}

		// Build tree
		root, buildCount, err := scanner.walkAndBuild(tempDir, config, nil, count)
		if err != nil {
			t.Fatalf("walkAndBuild failed: %v", err)
		}

		// Count and build passes should match
		if count != buildCount {
			t.Errorf("Count mismatch: countItems=%d, walkAndBuild=%d", count, buildCount)
		}

		// Verify hidden files are present in the tree
		var hiddenFiles []string
		var checkHiddenFiles func(*FileNode)
		checkHiddenFiles = func(node *FileNode) {
			for _, child := range node.Children {
				if strings.HasPrefix(child.Name, ".") && child.Name != "." && child.Name != ".." {
					hiddenFiles = append(hiddenFiles, child.RelPath)
				}
				checkHiddenFiles(child)
			}
		}

		checkHiddenFiles(root)
		if len(hiddenFiles) == 0 {
			t.Error("Expected to find hidden files when IncludeHidden=true")
		}

		// Should find at least .hidden.txt
		foundHiddenTxt := false
		for _, path := range hiddenFiles {
			if strings.Contains(path, ".hidden.txt") {
				foundHiddenTxt = true
				break
			}
		}
		if !foundHiddenTxt {
			t.Error("Expected to find .hidden.txt in results")
		}
	})

	t.Run("explicitly included hidden file still excluded if IncludeHidden=false", func(t *testing.T) {
		// Create scanner with explicit include for a hidden file
		scanner, err := NewFileSystemScannerWithIgnore([]string{})
		if err != nil {
			t.Fatalf("Failed to create scanner with ignore: %v", err)
		}

		// Add explicit include for the hidden file
		if scanner.ignoreEngine != nil {
			_ = scanner.ignoreEngine.AddExplicitInclude(".hidden.txt")
		}

		config := DefaultScanConfig()
		config.IncludeHidden = false

		// Count items
		count, err := scanner.countItems(tempDir, config)
		if err != nil {
			t.Fatalf("countItems failed: %v", err)
		}

		// Build tree
		root, buildCount, err := scanner.walkAndBuild(tempDir, config, nil, count)
		if err != nil {
			t.Fatalf("walkAndBuild failed: %v", err)
		}

		// Count and build passes should match
		if count != buildCount {
			t.Errorf("Count mismatch: countItems=%d, walkAndBuild=%d", count, buildCount)
		}

		// Verify no hidden files are present (hidden rule should apply alongside engine decision)
		var checkHiddenFiles func(*FileNode) []string
		checkHiddenFiles = func(node *FileNode) []string {
			var hiddenFiles []string
			for _, child := range node.Children {
				if strings.HasPrefix(child.Name, ".") && child.Name != "." && child.Name != ".." {
					hiddenFiles = append(hiddenFiles, child.RelPath)
				}
				hiddenFiles = append(hiddenFiles, checkHiddenFiles(child)...)
			}
			return hiddenFiles
		}

		hiddenFiles := checkHiddenFiles(root)
		if len(hiddenFiles) > 0 {
			t.Errorf("Explicitly included hidden file should still be excluded when IncludeHidden=false: %v", hiddenFiles)
		}
	})
}

func TestNewFileSystemScannerWithMatcher(t *testing.T) {
	// Create a fake matcher for testing
	fakeMatcher := &fakeMatcher{
		shouldMatch: make(map[string]bool),
	}

	// Configure the fake matcher to ignore specific paths
	fakeMatcher.shouldMatch["test.tmp"] = true
	fakeMatcher.shouldMatch["important.txt"] = false

	// Create scanner with the fake matcher
	scanner := NewFileSystemScannerWithMatcher(fakeMatcher)

	if scanner == nil {
		t.Fatal("Expected non-nil scanner")
	}

	// Verify the scanner uses the injected matcher
	// We can't easily test the full scan without creating files,
	// but we can verify the constructor worked
	if scanner.pathMatcher != fakeMatcher {
		t.Error("Scanner should use the injected matcher")
	}
}

// fakeMatcher is a test implementation of PathMatcher
type fakeMatcher struct {
	shouldMatch map[string]bool
	calls       []matchCall
}

type matchCall struct {
	path  string
	isDir bool
}

func (f *fakeMatcher) Matches(path string, isDir bool) bool {
	f.calls = append(f.calls, matchCall{path: path, isDir: isDir})
	return f.shouldMatch[path]
}

func TestScannerInterface(t *testing.T) {
	// Verify that FileSystemScanner implements the Scanner interface
	var _ Scanner = (*FileSystemScanner)(nil)
}

func TestTreeSorting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sorting_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files and directories in non-alphabetical order
	items := []string{
		"zebra.txt",
		"apple.txt",
		"beta/",
		"alpha/",
		"charlie.txt",
	}

	for _, item := range items {
		fullPath := filepath.Join(tempDir, item)
		if strings.HasSuffix(item, "/") {
			err := os.MkdirAll(fullPath, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", item, err)
			}
		} else {
			err := os.WriteFile(fullPath, []byte("test"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", item, err)
			}
		}
	}

	scanner := NewFileSystemScanner()
	config := DefaultScanConfig()

	root, err := scanner.Scan(tempDir, config)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(root.Children) < 5 {
		t.Fatalf("Expected at least 5 children, got %d", len(root.Children))
	}

	// Check that directories come first, then files, both alphabetically
	var directories []string
	var files []string

	for _, child := range root.Children {
		if child.IsDir {
			directories = append(directories, child.Name)
		} else {
			files = append(files, child.Name)
		}
	}

	// Verify directories are sorted alphabetically
	expectedDirs := []string{"alpha", "beta"}
	if len(directories) != len(expectedDirs) {
		t.Errorf("Expected directories %v, got %v", expectedDirs, directories)
	} else {
		for i, dir := range directories {
			if dir != expectedDirs[i] {
				t.Errorf("Expected directory %s at position %d, got %s", expectedDirs[i], i, dir)
			}
		}
	}

	// Verify files are sorted alphabetically
	expectedFiles := []string{"apple.txt", "charlie.txt", "zebra.txt"}
	if len(files) != len(expectedFiles) {
		t.Errorf("Expected files %v, got %v", expectedFiles, files)
	} else {
		for i, file := range files {
			if file != expectedFiles[i] {
				t.Errorf("Expected file %s at position %d, got %s", expectedFiles[i], i, file)
			}
		}
	}
}
