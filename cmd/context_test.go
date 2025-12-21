package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Note: ParseSize, ParseSizeWithDefault, and FormatBytes tests are in internal/utils/conversion_test.go

func TestBuildGenerateConfigDefaults(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().StringSlice("include", []string{"*"}, "")
	cmd.Flags().StringSlice("exclude", nil, "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("max-size", "10MB", "")
	cmd.Flags().Bool("enforce-limit", true, "")

	temp := t.TempDir()
	_ = cmd.Flags().Set("root", temp)

	cfg, err := buildGenerateConfig(cmd)
	if err != nil {
		t.Fatalf("buildGenerateConfig error: %v", err)
	}

	if !strings.HasPrefix(filepath.Base(cfg.Output), "shotgun-prompt-") {
		t.Fatalf("expected generated output name, got %s", cfg.Output)
	}
}

func TestContextGeneratePreRunValidation(t *testing.T) {
	cmd := contextGenerateCmd
	t.Cleanup(func() {
		_ = cmd.Flags().Set("root", ".")
		_ = cmd.Flags().Set("max-size", "10MB")
		_ = cmd.Flags().Set("output", "")
	})
	_ = cmd.Flags().Set("max-size", "1MB")
	_ = cmd.Flags().Set("root", "/does/not/exist")

	err := cmd.PreRunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error for missing root path")
	}
}

func TestBuildGenerateConfigHonorsOutput(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().StringSlice("include", []string{"*"}, "")
	cmd.Flags().StringSlice("exclude", nil, "")
	cmd.Flags().String("output", "custom.md", "")
	cmd.Flags().String("max-size", "1MB", "")
	cmd.Flags().Bool("enforce-limit", true, "")

	dir := t.TempDir()
	_ = cmd.Flags().Set("root", dir)

	cfg, err := buildGenerateConfig(cmd)
	if err != nil {
		t.Fatalf("buildGenerateConfig: %v", err)
	}
	if cfg.Output != "custom.md" {
		t.Fatalf("expected custom output, got %s", cfg.Output)
	}
}

func TestGenerateContextHeadlessInvalidOutput(t *testing.T) {
	// Create a truly invalid path by using a file as directory
	tmpFile, err := os.CreateTemp("", "test_file")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	cfg := GenerateConfig{
		RootPath: filepath.Join(tmpFile.Name(), "subdir"), // Using file as directory
		Include:  []string{"*"},
		Exclude:  nil,
		Output:   filepath.Join(os.TempDir(), "out.md"),
		MaxSize:  1024,
	}

	err = generateContextHeadless(cfg)
	if err == nil {
		t.Fatal("expected error for invalid root path")
	}
}
