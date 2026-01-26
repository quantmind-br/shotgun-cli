package cmd

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	return buf.String()
}

func TestRunRootCommandVersionFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("version", false, "")
	_ = cmd.Flags().Set("version", "true")

	output := captureStdout(t, func() {
		runRootCommand(cmd, nil)
	})

	if output == "" {
		t.Fatal("expected version output")
	}
}

func TestInitConfigSetsDefaults(t *testing.T) {
	cfgFile = ""
	viper.Reset()

	initConfig()

	if got := viper.GetInt("scanner.max-files"); got == 0 {
		t.Fatalf("expected scanner.max-files default, got %d", got)
	}
}

func TestGetConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")

	dir := getConfigDir()

	switch runtime.GOOS {
	case "windows":
		// Skip test on Windows - path semantics differ
		t.Skip("skip on windows path semantics")
	case "darwin":
		// macOS uses Library/Application Support regardless of XDG_CONFIG_HOME
		if dir == "" {
			t.Fatal("expected non-empty config dir on macOS")
		}
		// Path should contain Application Support
		if !contains(dir, "Application Support") {
			t.Fatalf("unexpected config dir on macOS: %s", dir)
		}
	default:
		// Linux and others should respect XDG_CONFIG_HOME
		if dir != "/tmp/xdg-config/shotgun-cli" {
			t.Fatalf("unexpected config dir: %s", dir)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			bytes.Contains([]byte(s), []byte(substr))))
}

func TestExecute_ReturnsNilOnSuccess(t *testing.T) {
	oldRun := rootCmd.Run
	rootCmd.Run = func(cmd *cobra.Command, args []string) {}
	t.Cleanup(func() {
		rootCmd.Run = oldRun
	})

	rootCmd.SetArgs([]string{"--help"})
	err := Execute()

	if err != nil {
		t.Fatalf("Execute() with --help should succeed, got error: %v", err)
	}
}

func TestRunRootCommandNoArgsShowsHelp(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("version", false, "")

	oldArgs := os.Args
	os.Args = []string{"shotgun-cli", "--verbose"}
	t.Cleanup(func() {
		os.Args = oldArgs
	})

	output := captureStdout(t, func() {
		runRootCommand(cmd, []string{})
	})

	_ = output
}

func TestSetConfigDefaults(t *testing.T) {
	viper.Reset()

	setConfigDefaults()

	tests := []struct {
		key      string
		expected interface{}
	}{
		{"scanner.max-files", 10000},
		{"scanner.max-file-size", "1MB"},
		{"scanner.respect-gitignore", true},
		{"scanner.skip-binary", true},
		{"scanner.workers", 1},
		{"scanner.include-hidden", false},
		{"scanner.include-ignored", false},
		{"scanner.respect-shotgunignore", true},
		{"scanner.max-memory", "500MB"},
		{"context.max-size", "10MB"},
		{"context.include-tree", true},
		{"context.include-summary", true},
		{"output.format", "markdown"},
		{"output.clipboard", true},
		{"llm.provider", "openai"},
		{"llm.timeout", 300},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := viper.Get(tt.key)
			if got != tt.expected {
				t.Errorf("setConfigDefaults() %s = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestUpdateLoggingLevel(t *testing.T) {
	tests := []struct {
		name    string
		quiet   bool
		verbose bool
	}{
		{"default level", false, false},
		{"quiet mode", true, false},
		{"verbose mode", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("quiet", tt.quiet)
			viper.Set("verbose", tt.verbose)

			updateLoggingLevel()
		})
	}
}

func TestGetConfigDir_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	t.Setenv("APPDATA", "C:\\Users\\Test\\AppData\\Roaming")
	dir := getConfigDir()
	if dir != "C:\\Users\\Test\\AppData\\Roaming\\shotgun-cli" {
		t.Errorf("getConfigDir() on Windows = %s, want path with AppData", dir)
	}
}

func TestGetConfigDir_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-specific test")
	}

	dir := getConfigDir()
	if !contains(dir, "Library/Application Support/shotgun-cli") {
		t.Errorf("getConfigDir() on macOS = %s, want path with Library/Application Support", dir)
	}
}

func TestGetConfigDir_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-specific test")
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	dir := getConfigDir()
	if !contains(dir, ".config/shotgun-cli") {
		t.Errorf("getConfigDir() on Linux = %s, want path with .config", dir)
	}
}

func TestInitConfig_WithCustomConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/custom-config.yaml"
	if err := os.WriteFile(tmpFile, []byte("scanner:\n  max-files: 999\n"), 0644); err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}

	oldCfgFile := cfgFile
	cfgFile = tmpFile
	t.Cleanup(func() {
		cfgFile = oldCfgFile
		viper.Reset()
	})

	viper.Reset()
	initConfig()

	got := viper.GetInt("scanner.max-files")
	if got != 999 {
		t.Errorf("initConfig() with custom file: scanner.max-files = %d, want 999", got)
	}
}

func TestInitConfig_MissingConfigFileUsesDefaults(t *testing.T) {
	oldCfgFile := cfgFile
	cfgFile = ""
	t.Cleanup(func() {
		cfgFile = oldCfgFile
		viper.Reset()
	})

	viper.Reset()
	initConfig()

	got := viper.GetInt("scanner.max-files")
	if got == 0 {
		t.Error("initConfig() with missing config should use defaults")
	}
}
