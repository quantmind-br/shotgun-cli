package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestParseSize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		expect  int64
		wantErr bool
	}{
		{"bytes", "1024", 1024, false},
		{"kb", "1KB", 1024, false},
		{"mb", "2MB", 2 * 1024 * 1024, false},
		{"gb", "1GB", 1024 * 1024 * 1024, false},
		{"invalid", "abc", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSize(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %s", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSize error: %v", err)
			}
			if got != tc.expect {
				t.Fatalf("expected %d, got %d", tc.expect, got)
			}
		})
	}
}

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

func TestParseConfigSizeFallback(t *testing.T) {
	size := parseConfigSize("invalid")
	if size != 1024*1024 {
		t.Fatalf("expected 1MB fallback, got %d", size)
	}
}

func TestFormatBytes(t *testing.T) {
	cases := map[int64]string{
		42:          "42 B",
		1024:        "1.0 KB",
		1024 * 1024: "1.0 MB",
	}

	for input, expect := range cases {
		if got := formatBytes(input); got != expect {
			t.Fatalf("expected %s, got %s", expect, got)
		}
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
	cfg := GenerateConfig{
		RootPath: filepath.Join(os.TempDir(), "missing"),
		Include:  []string{"*"},
		Exclude:  nil,
		Output:   filepath.Join(os.TempDir(), "out.md"),
		MaxSize:  1024,
	}

	err := generateContextHeadless(cfg)
	if err == nil {
		t.Fatal("expected error for invalid root path")
	}
}
