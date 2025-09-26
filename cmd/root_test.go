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
	cmd.Flags().Set("version", "true")

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
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows path semantics")
	}

	dir := getConfigDir()
	if dir != "/tmp/xdg-config/shotgun-cli" {
		t.Fatalf("unexpected config dir: %s", dir)
	}
}
