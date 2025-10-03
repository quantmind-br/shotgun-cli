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
	w.Close()
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
