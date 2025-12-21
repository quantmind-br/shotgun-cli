package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreReason_String(t *testing.T) {
	tests := []struct {
		reason   IgnoreReason
		expected string
	}{
		{IgnoreReasonNone, "none"},
		{IgnoreReasonBuiltIn, "built-in"},
		{IgnoreReasonGitignore, "gitignore"},
		{IgnoreReasonCustom, "custom"},
		{IgnoreReasonExplicit, "explicit"},
		{IgnoreReason(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.reason.String(); got != tt.expected {
				t.Errorf("IgnoreReason.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewIgnoreEngine(t *testing.T) {
	engine := NewIgnoreEngine()

	if engine == nil {
		t.Fatal("NewIgnoreEngine() returned nil")
	}

	if engine.builtInMatcher == nil {
		t.Error("builtInMatcher is nil")
	}

	if engine.gitignoreMatcher == nil {
		t.Error("gitignoreMatcher is nil")
	}

	if engine.customMatcher == nil {
		t.Error("customMatcher is nil")
	}

	if engine.explicitExcludes == nil {
		t.Error("explicitExcludes is nil")
	}

	if engine.explicitIncludes == nil {
		t.Error("explicitIncludes is nil")
	}
}

func TestLayeredIgnoreEngine_BuiltInPatterns(t *testing.T) {
	engine := NewIgnoreEngine()

	tests := []struct {
		path     string
		ignored  bool
		reason   IgnoreReason
		testName string
	}{
		// Shotgun-specific patterns
		{"shotgun-prompt-test.md", true, IgnoreReasonBuiltIn, "shotgun-prompt file"},
		{"docs/shotgun-prompt-feature.md", true, IgnoreReasonBuiltIn, "shotgun-prompt in subdirectory"},
		{"shotgun-prompt.md", true, IgnoreReasonBuiltIn, "shotgun-prompt without suffix"},

		// Version control
		{".git/config", true, IgnoreReasonBuiltIn, "git directory"},
		{".svn/entries", true, IgnoreReasonBuiltIn, "svn directory"},
		{"src/.git/hooks", true, IgnoreReasonBuiltIn, "nested git directory"},

		// IDE and editor files
		{".vscode/settings.json", true, IgnoreReasonBuiltIn, "vscode directory"},
		{".idea/workspace.xml", true, IgnoreReasonBuiltIn, "idea directory"},
		{"file.swp", true, IgnoreReasonBuiltIn, "vim swap file"},
		{".DS_Store", true, IgnoreReasonBuiltIn, "macos ds store"},

		// Build and dependency directories
		{"node_modules/package/index.js", true, IgnoreReasonBuiltIn, "node modules"},
		{"target/classes/Main.class", true, IgnoreReasonBuiltIn, "maven target"},
		{"build/output.jar", true, IgnoreReasonBuiltIn, "build directory"},

		// Cache and temporary files
		{"__pycache__/module.pyc", true, IgnoreReasonBuiltIn, "python cache"},
		{"file.pyc", true, IgnoreReasonBuiltIn, "python compiled"},
		{".cache/data", true, IgnoreReasonBuiltIn, "cache directory"},

		// Log files
		{"app.log", true, IgnoreReasonBuiltIn, "log file"},
		{"logs/error.log", true, IgnoreReasonBuiltIn, "logs directory"},

		// Files that should NOT be ignored
		{"README.md", false, IgnoreReasonNone, "readme file"},
		{"src/main.go", false, IgnoreReasonNone, "source file"},
		{"config.yaml", false, IgnoreReasonNone, "config file"},
		{"test.txt", false, IgnoreReasonNone, "text file"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ignored, reason := engine.ShouldIgnore(tt.path)
			if ignored != tt.ignored {
				t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
			}
			if reason != tt.reason {
				t.Errorf("ShouldIgnore(%q) reason = %v, want %v", tt.path, reason, tt.reason)
			}
		})
	}
}

func TestLayeredIgnoreEngine_LoadGitignore(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "ignore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	engine := NewIgnoreEngine()

	t.Run("missing gitignore file", func(t *testing.T) {
		err := engine.LoadGitignore(tmpDir)
		if err != nil {
			t.Errorf("LoadGitignore() with missing file should not error, got %v", err)
		}

		// Should not ignore files when no .gitignore exists
		ignored, reason := engine.ShouldIgnore("test.txt")
		if ignored && reason == IgnoreReasonGitignore {
			t.Error("Should not ignore files when no .gitignore exists")
		}
	})

	t.Run("valid gitignore file", func(t *testing.T) {
		// Create .gitignore file with patterns that won't conflict with built-in patterns
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		gitignoreContent := `*.tmp
/gitignore-test/
unique-folder/
!important.tmp
`
		err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		err = engine.LoadGitignore(tmpDir)
		if err != nil {
			t.Errorf("LoadGitignore() error = %v", err)
		}

		tests := []struct {
			path     string
			ignored  bool
			reason   IgnoreReason
			testName string
		}{
			{"test.tmp", true, IgnoreReasonGitignore, "tmp file should be ignored"},
			{"gitignore-test/result.txt", true, IgnoreReasonGitignore, "gitignore-test directory should be ignored"},
			{"unique-folder/package.json", true, IgnoreReasonGitignore, "unique-folder should be ignored"},
			{"important.tmp", false, IgnoreReasonNone, "negated pattern should not be ignored"},
			{"test.txt", false, IgnoreReasonNone, "non-matching file should not be ignored"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				ignored, reason := engine.ShouldIgnore(tt.path)
				if ignored != tt.ignored {
					t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
				}
				if ignored && reason != tt.reason {
					t.Errorf("ShouldIgnore(%q) reason = %v, want %v", tt.path, reason, tt.reason)
				}
			})
		}
	})

	t.Run("invalid gitignore file", func(t *testing.T) {
		// Create invalid .gitignore file (should still work, gitignore is forgiving)
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		err := os.WriteFile(gitignorePath, []byte(""), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		err = engine.LoadGitignore(tmpDir)
		if err != nil {
			t.Errorf("LoadGitignore() with empty file should not error, got %v", err)
		}
	})
}

func TestLayeredIgnoreEngine_CustomRules(t *testing.T) {
	engine := NewIgnoreEngine()

	t.Run("add custom rule", func(t *testing.T) {
		err := engine.AddCustomRule("*.custom")
		if err != nil {
			t.Errorf("AddCustomRule() error = %v", err)
		}

		ignored, reason := engine.ShouldIgnore("test.custom")
		if !ignored {
			t.Error("Custom rule should ignore matching files")
		}
		if reason != IgnoreReasonCustom {
			t.Errorf("Expected IgnoreReasonCustom, got %v", reason)
		}

		// Non-matching file should not be ignored
		ignored, _ = engine.ShouldIgnore("test.txt")
		if ignored {
			t.Error("Custom rule should not ignore non-matching files")
		}
	})

	t.Run("add multiple custom rules", func(t *testing.T) {
		patterns := []string{"*.test", "temp/", "*.debug"}
		err := engine.AddCustomRules(patterns)
		if err != nil {
			t.Errorf("AddCustomRules() error = %v", err)
		}

		tests := []struct {
			path     string
			ignored  bool
			testName string
		}{
			{"file.test", true, "test extension"},
			{"temp/data.txt", true, "temp directory"},
			{"app.debug", true, "debug extension"},
			{"normal.txt", false, "non-matching file"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				ignored, reason := engine.ShouldIgnore(tt.path)
				if ignored != tt.ignored {
					t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
				}
				if ignored && reason != IgnoreReasonCustom {
					t.Errorf("Expected IgnoreReasonCustom, got %v", reason)
				}
			})
		}
	})

	t.Run("empty custom rules", func(t *testing.T) {
		err := engine.AddCustomRule("")
		if err != nil {
			t.Errorf("AddCustomRule() with empty string should not error")
		}

		err = engine.AddCustomRules([]string{})
		if err != nil {
			t.Errorf("AddCustomRules() with empty slice should not error")
		}

		err = engine.AddCustomRules([]string{"", "  ", "\t"})
		if err != nil {
			t.Errorf("AddCustomRules() with whitespace-only patterns should not error")
		}
	})
}

func TestLayeredIgnoreEngine_ExplicitRules(t *testing.T) {
	engine := NewIgnoreEngine()

	t.Run("explicit exclude", func(t *testing.T) {
		err := engine.AddExplicitExclude("*.secret")
		if err != nil {
			t.Errorf("AddExplicitExclude() error = %v", err)
		}

		ignored, reason := engine.ShouldIgnore("config.secret")
		if !ignored {
			t.Error("Explicit exclude should ignore matching files")
		}
		if reason != IgnoreReasonExplicit {
			t.Errorf("Expected IgnoreReasonExplicit, got %v", reason)
		}
	})

	t.Run("explicit include", func(t *testing.T) {
		err := engine.AddExplicitInclude("*.important")
		if err != nil {
			t.Errorf("AddExplicitInclude() error = %v", err)
		}

		// Add a built-in pattern that would normally ignore this
		err = engine.AddCustomRule("*.important")
		if err != nil {
			t.Errorf("AddCustomRule() error = %v", err)
		}

		// Explicit include should override custom rule
		ignored, reason := engine.ShouldIgnore("data.important")
		if ignored {
			t.Error("Explicit include should override other ignore rules")
		}
		if reason != IgnoreReasonNone {
			t.Errorf("Expected IgnoreReasonNone, got %v", reason)
		}
	})

	t.Run("empty explicit rules", func(t *testing.T) {
		err := engine.AddExplicitExclude("")
		if err != nil {
			t.Errorf("AddExplicitExclude() with empty string should not error")
		}

		err = engine.AddExplicitInclude("")
		if err != nil {
			t.Errorf("AddExplicitInclude() with empty string should not error")
		}
	})
}

func TestLayeredIgnoreEngine_RulePrecedence(t *testing.T) {
	engine := NewIgnoreEngine()

	// Set up rules in different layers
	err := engine.AddCustomRule("*.test")
	if err != nil {
		t.Fatal(err)
	}

	err = engine.AddExplicitExclude("*.exclude")
	if err != nil {
		t.Fatal(err)
	}

	err = engine.AddExplicitInclude("*.include")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path     string
		ignored  bool
		reason   IgnoreReason
		testName string
	}{
		// Explicit exclude should have highest priority
		{"file.exclude", true, IgnoreReasonExplicit, "explicit exclude"},

		// Explicit include should override everything except explicit exclude
		{"file.include", false, IgnoreReasonNone, "explicit include"},

		// Built-in patterns should override gitignore and custom
		{"shotgun-prompt-test.md", true, IgnoreReasonBuiltIn, "built-in pattern"},

		// Custom patterns should have lowest priority among ignore rules
		{"file.test", true, IgnoreReasonCustom, "custom pattern"},

		// No matching rules
		{"normal.txt", false, IgnoreReasonNone, "no matching rules"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ignored, reason := engine.ShouldIgnore(tt.path)
			if ignored != tt.ignored {
				t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
			}
			if reason != tt.reason {
				t.Errorf("ShouldIgnore(%q) reason = %v, want %v", tt.path, reason, tt.reason)
			}
		})
	}
}

func TestLayeredIgnoreEngine_IsGitignored(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "ignore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .gitignore file
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	gitignoreContent := `*.tmp
/gitignore-test/
`
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	engine := NewIgnoreEngine()
	err = engine.LoadGitignore(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path       string
		gitignored bool
		testName   string
	}{
		{"test.tmp", true, "gitignored file"},
		{"gitignore-test/result.txt", true, "gitignored directory"},
		{"normal.txt", false, "non-gitignored file"},
		{"shotgun-prompt-test.md", false, "built-in ignored but not gitignored"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gitignored := engine.IsGitignored(tt.path)
			if gitignored != tt.gitignored {
				t.Errorf("IsGitignored(%q) = %v, want %v", tt.path, gitignored, tt.gitignored)
			}
		})
	}
}

func TestLayeredIgnoreEngine_IsCustomIgnored(t *testing.T) {
	engine := NewIgnoreEngine()

	err := engine.AddCustomRules([]string{"*.custom", "temp/"})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path          string
		customIgnored bool
		testName      string
	}{
		{"test.custom", true, "custom ignored file"},
		{"temp/data.txt", true, "custom ignored directory"},
		{"normal.txt", false, "non-custom ignored file"},
		{"shotgun-prompt-test.md", false, "built-in ignored but not custom ignored"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			customIgnored := engine.IsCustomIgnored(tt.path)
			if customIgnored != tt.customIgnored {
				t.Errorf("IsCustomIgnored(%q) = %v, want %v", tt.path, customIgnored, tt.customIgnored)
			}
		})
	}
}

func TestLayeredIgnoreEngine_PathNormalization(t *testing.T) {
	engine := NewIgnoreEngine()

	err := engine.AddCustomRule("temp/")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path     string
		ignored  bool
		testName string
	}{
		{"temp/file.txt", true, "forward slash path"},
		{filepath.Join("temp", "file.txt"), true, "os-specific path separator"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			ignored, _ := engine.ShouldIgnore(tt.path)
			if ignored != tt.ignored {
				t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
			}
		})
	}
}

func TestLayeredIgnoreEngine_LoadShotgunignore(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "shotgunignore_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("missing shotgunignore file", func(t *testing.T) {
		engine := NewIgnoreEngine()
		err := engine.LoadShotgunignore(tmpDir)
		if err != nil {
			t.Errorf("LoadShotgunignore() with missing file should not error, got %v", err)
		}

		// Should not ignore files when no .shotgunignore exists
		ignored, reason := engine.ShouldIgnore("test.txt")
		if ignored && reason == IgnoreReasonCustom {
			t.Error("Should not ignore files when no .shotgunignore exists")
		}
	})

	t.Run("valid shotgunignore file", func(t *testing.T) {
		engine := NewIgnoreEngine()

		// Create .shotgunignore file with patterns
		shotgunignorePath := filepath.Join(tmpDir, ".shotgunignore")
		shotgunignoreContent := `# Test files
*_test.go
test/**
*.spec.js
!important_test.go
`
		err := os.WriteFile(shotgunignorePath, []byte(shotgunignoreContent), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		err = engine.LoadShotgunignore(tmpDir)
		if err != nil {
			t.Errorf("LoadShotgunignore() error = %v", err)
		}

		tests := []struct {
			path     string
			ignored  bool
			reason   IgnoreReason
			testName string
		}{
			{"main_test.go", true, IgnoreReasonCustom, "test file should be ignored"},
			{"test/unit/example.go", true, IgnoreReasonCustom, "test directory should be ignored"},
			{"app.spec.js", true, IgnoreReasonCustom, "spec file should be ignored"},
			{"important_test.go", false, IgnoreReasonNone, "negated pattern should not be ignored"},
			{"main.go", false, IgnoreReasonNone, "non-matching file should not be ignored"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				ignored, reason := engine.ShouldIgnore(tt.path)
				if ignored != tt.ignored {
					t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
				}
				if ignored && reason != tt.reason {
					t.Errorf("ShouldIgnore(%q) reason = %v, want %v", tt.path, reason, tt.reason)
				}
			})
		}
	})

	t.Run("nested shotgunignore file", func(t *testing.T) {
		engine := NewIgnoreEngine()

		// Create nested directory structure
		nestedDir := filepath.Join(tmpDir, "nested")
		err := os.MkdirAll(nestedDir, 0o750)
		if err != nil {
			t.Fatal(err)
		}

		// Create .shotgunignore file in nested directory
		nestedShotgunignorePath := filepath.Join(nestedDir, ".shotgunignore")
		nestedShotgunignoreContent := `*.local
temp/
`
		err = os.WriteFile(nestedShotgunignorePath, []byte(nestedShotgunignoreContent), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		err = engine.LoadShotgunignore(tmpDir)
		if err != nil {
			t.Errorf("LoadShotgunignore() error = %v", err)
		}

		tests := []struct {
			path     string
			ignored  bool
			testName string
		}{
			{"nested/config.local", true, "nested local file should be ignored"},
			{"nested/temp/data.txt", true, "nested temp directory should be ignored"},
			{"config.local", false, "root local file should not be ignored by nested rule"},
		}

		for _, tt := range tests {
			t.Run(tt.testName, func(t *testing.T) {
				ignored, _ := engine.ShouldIgnore(tt.path)
				if ignored != tt.ignored {
					t.Errorf("ShouldIgnore(%q) ignored = %v, want %v", tt.path, ignored, tt.ignored)
				}
			})
		}
	})

	t.Run("empty shotgunignore file", func(t *testing.T) {
		engine := NewIgnoreEngine()

		// Create empty .shotgunignore file
		shotgunignorePath := filepath.Join(tmpDir, ".shotgunignore")
		err := os.WriteFile(shotgunignorePath, []byte(""), 0o600)
		if err != nil {
			t.Fatal(err)
		}

		err = engine.LoadShotgunignore(tmpDir)
		if err != nil {
			t.Errorf("LoadShotgunignore() with empty file should not error, got %v", err)
		}
	})
}

// Benchmark tests
func BenchmarkLayeredIgnoreEngine_ShouldIgnore(b *testing.B) {
	engine := NewIgnoreEngine()

	// Add some custom rules
	_ = engine.AddCustomRules([]string{"*.test", "temp/", "build/"})

	paths := []string{
		"src/main.go",
		"test.tmp",
		"shotgun-prompt-feature.md",
		"node_modules/package/index.js",
		"normal.txt",
		"temp/data.txt",
		"build/output.jar",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			engine.ShouldIgnore(path)
		}
	}
}

func BenchmarkLayeredIgnoreEngine_NewIgnoreEngine(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewIgnoreEngine()
	}
}
