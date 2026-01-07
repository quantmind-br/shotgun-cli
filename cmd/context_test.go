package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func TestGenerateContextHeadlessTemplateNotFound(t *testing.T) {
	dir := t.TempDir()
	cfg := GenerateConfig{
		RootPath: dir,
		Include:  []string{"*"},
		Exclude:  nil,
		Output:   filepath.Join(dir, "out.md"),
		MaxSize:  1024,
		Template: "/nonexistent/template.md",
	}

	err := generateContextHeadless(cfg)
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
}

func TestGenerateContextHeadlessSuccess(t *testing.T) {
	dir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(dir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := GenerateConfig{
		RootPath:     dir,
		Include:      []string{"*"},
		Exclude:      nil,
		Output:       filepath.Join(dir, "out.md"),
		MaxSize:      1024 * 1024, // 1MB
		ProgressMode: ProgressNone,
	}

	err := generateContextHeadless(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(cfg.Output); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestGenerateContextHeadlessWithProgress(t *testing.T) {
	dir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(dir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := GenerateConfig{
		RootPath:     dir,
		Include:      []string{"*"},
		Exclude:      nil,
		Output:       filepath.Join(dir, "out.md"),
		MaxSize:      1024 * 1024,
		ProgressMode: ProgressHuman,
	}

	err := generateContextHeadless(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(cfg.Output); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestGenerateContextHeadlessWithIncludeExclude(t *testing.T) {
	dir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(dir, "test.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := GenerateConfig{
		RootPath:     dir,
		Include:      []string{"*.go"},
		Exclude:      []string{"*.md"},
		Output:       filepath.Join(dir, "out.md"),
		MaxSize:      1024 * 1024,
		ProgressMode: ProgressNone,
	}

	err := generateContextHeadless(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(cfg.Output); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestGenerateContextHeadlessEmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	cfg := GenerateConfig{
		RootPath:     dir,
		Include:      []string{"*"},
		Exclude:      nil,
		Output:       filepath.Join(dir, "out.md"),
		MaxSize:      1024 * 1024,
		ProgressMode: ProgressNone,
	}

	err := generateContextHeadless(cfg)
	if err != nil {
		t.Fatalf("unexpected error for empty directory: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(cfg.Output); os.IsNotExist(err) {
		t.Fatal("output file was not created")
	}
}

func TestGenerateContextHeadlessWithMaxSize(t *testing.T) {
	dir := t.TempDir()

	// Create a test file larger than max size
	largeContent := strings.Repeat("x", 1024*1024) // 1MB
	testFile := filepath.Join(dir, "large.go")
	if err := os.WriteFile(testFile, []byte(largeContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cfg := GenerateConfig{
		RootPath:     dir,
		Include:      []string{"*"},
		Exclude:      nil,
		Output:       filepath.Join(dir, "out.md"),
		MaxSize:      100, // 100 bytes
		EnforceLimit: true,
		ProgressMode: ProgressNone,
	}

	err := generateContextHeadless(cfg)
	// Should error due to size limit exceeded
	if err == nil {
		t.Fatal("expected error for exceeding max size")
	}
}

func TestRenderProgressHuman_WithTotal(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	p := ProgressOutput{
		Timestamp: "2024-01-01T12:00:00Z",
		Stage:     "scanning",
		Message:   "Processing files",
		Current:   50,
		Total:     100,
		Percent:   50.0,
	}

	renderProgressHuman(p)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected elements
	if !strings.Contains(output, "[scanning]") {
		t.Errorf("output should contain stage '[scanning]', got: %s", output)
	}
	if !strings.Contains(output, "Processing files") {
		t.Errorf("output should contain message 'Processing files', got: %s", output)
	}
	if !strings.Contains(output, "50/100") {
		t.Errorf("output should contain progress '50/100', got: %s", output)
	}
	if !strings.Contains(output, "50.0%") {
		t.Errorf("output should contain percentage '50.0%%', got: %s", output)
	}
}

func TestRenderProgressHuman_WithoutTotal(t *testing.T) {
	// Capture stderr output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	p := ProgressOutput{
		Timestamp: "2024-01-01T12:00:00Z",
		Stage:     "loading",
		Message:   "Loading configuration",
		Current:   0,
		Total:     0,
		Percent:   0,
	}

	renderProgressHuman(p)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected elements
	if !strings.Contains(output, "[loading]") {
		t.Errorf("output should contain stage '[loading]', got: %s", output)
	}
	if !strings.Contains(output, "Loading configuration") {
		t.Errorf("output should contain message 'Loading configuration', got: %s", output)
	}
}

func TestRenderProgressHuman_DifferentStages(t *testing.T) {
	stages := []struct {
		stage   string
		message string
		current int64
		total   int64
	}{
		{"scanning", "Scanning files", 10, 100},
		{"parsing", "Parsing content", 25, 100},
		{"generating", "Generating context", 80, 100},
		{"complete", "Done", 100, 100},
	}

	for _, s := range stages {
		t.Run(s.stage, func(t *testing.T) {
			// Capture stderr output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			var percent float64
			if s.total > 0 {
				percent = float64(s.current) / float64(s.total) * 100
			}

			p := ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     s.stage,
				Message:   s.message,
				Current:   s.current,
				Total:     s.total,
				Percent:   percent,
			}

			renderProgressHuman(p)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if !strings.Contains(output, "["+s.stage+"]") {
				t.Errorf("output should contain stage '[%s]', got: %s", s.stage, output)
			}
		})
	}
}

func TestRenderProgressJSON(t *testing.T) {
	tests := []struct {
		name     string
		progress ProgressOutput
		wantJSON string
	}{
		{
			name: "full progress with total",
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "scanning",
				Message:   "Processing files",
				Current:   50,
				Total:     100,
				Percent:   50.0,
			},
			wantJSON: `{"timestamp":"2024-01-01T12:00:00Z","stage":"scanning","message":"Processing files","current":50,"total":100,"percent":50}`,
		},
		{
			name: "progress without total",
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "generating",
				Message:   "Creating context",
			},
			wantJSON: `{"timestamp":"2024-01-01T12:00:00Z","stage":"generating","message":"Creating context"}`,
		},
		{
			name: "progress with zero percent",
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "starting",
				Message:   "Initializing",
				Current:   0,
				Total:     100,
				Percent:   0,
			},
			wantJSON: `{"timestamp":"2024-01-01T12:00:00Z","stage":"starting","message":"Initializing","total":100}`,
		},
		{
			name: "progress at 100 percent",
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "complete",
				Message:   "Done",
				Current:   100,
				Total:     100,
				Percent:   100,
			},
			wantJSON: `{"timestamp":"2024-01-01T12:00:00Z","stage":"complete","message":"Done","current":100,"total":100,"percent":100}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			renderProgressJSON(tt.progress)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if !strings.Contains(output, tt.wantJSON) {
				t.Errorf("renderProgressJSON() output = %s, want to contain %s", output, tt.wantJSON)
			}
		})
	}
}

func TestRenderProgress(t *testing.T) {
	tests := []struct {
		name        string
		mode        ProgressMode
		progress    ProgressOutput
		wantContain string
	}{
		{
			name: "human mode with total",
			mode: ProgressHuman,
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "scanning",
				Message:   "Processing files",
				Current:   50,
				Total:     100,
				Percent:   50.0,
			},
			wantContain: "[scanning]",
		},
		{
			name: "human mode without total",
			mode: ProgressHuman,
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "generating",
				Message:   "Creating context",
			},
			wantContain: "[generating]",
		},
		{
			name: "json mode",
			mode: ProgressJSON,
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "scanning",
				Message:   "Processing files",
				Current:   50,
				Total:     100,
				Percent:   50.0,
			},
			wantContain: `{"timestamp":"2024-01-01T12:00:00Z","stage":"scanning"`,
		},
		{
			name: "none mode - no output",
			mode: ProgressNone,
			progress: ProgressOutput{
				Timestamp: "2024-01-01T12:00:00Z",
				Stage:     "scanning",
				Message:   "Processing files",
			},
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr output
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			renderProgress(tt.mode, tt.progress)

			_ = w.Close()
			os.Stderr = oldStderr

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := strings.TrimSpace(buf.String())

			if tt.wantContain == "" {
				if output != "" {
					t.Errorf("renderProgress(none) should produce no output, got: %s", output)
				}
			} else {
				if !strings.Contains(output, tt.wantContain) {
					t.Errorf("renderProgress() output = %s, want to contain %s", output, tt.wantContain)
				}
			}
		})
	}
}

func TestLoadTemplateContent_EmptyTemplateName(t *testing.T) {
	content, err := loadTemplateContent("")
	if err != nil {
		t.Fatalf("loadTemplateContent('') should return nil error, got: %v", err)
	}
	if content != "" {
		t.Errorf("loadTemplateContent('') should return empty string, got: %s", content)
	}
}

func TestLoadTemplateContent_NonExistentTemplate(t *testing.T) {
	_, err := loadTemplateContent("nonexistent-template-xyz")
	if err == nil {
		t.Error("loadTemplateContent() with nonexistent template should return error")
	}
}

func TestPrintGenerationSummary(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	result := &app.GenerateResult{
		OutputPath:    "/tmp/output.md",
		FileCount:     10,
		ContentSize:   1024,
		TokenEstimate: 256,
	}
	cfg := GenerateConfig{
		RootPath: "/tmp/project",
		MaxSize:  10 * 1024 * 1024,
	}

	printGenerationSummary(result, cfg)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Context generated successfully") {
		t.Error("output should contain success message")
	}
	if !strings.Contains(output, "/tmp/project") {
		t.Error("output should contain root path")
	}
	if !strings.Contains(output, "/tmp/output.md") {
		t.Error("output should contain output path")
	}
	if !strings.Contains(output, "10") {
		t.Error("output should contain file count")
	}
}

func TestBuildGenerateConfig_InvalidProgressMode(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().StringSlice("include", []string{"*"}, "")
	cmd.Flags().StringSlice("exclude", nil, "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("max-size", "10MB", "")
	cmd.Flags().Bool("enforce-limit", true, "")
	cmd.Flags().Bool("send-gemini", false, "")
	cmd.Flags().String("gemini-model", "", "")
	cmd.Flags().String("gemini-output", "", "")
	cmd.Flags().Int("gemini-timeout", 0, "")
	cmd.Flags().String("template", "", "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("rules", "", "")
	cmd.Flags().StringArray("var", []string{}, "")
	cmd.Flags().Int("workers", 0, "")
	cmd.Flags().Bool("include-hidden", false, "")
	cmd.Flags().Bool("include-ignored", false, "")
	cmd.Flags().String("progress", "invalid", "")

	dir := t.TempDir()
	_ = cmd.Flags().Set("root", dir)

	_, err := buildGenerateConfig(cmd)
	if err == nil {
		t.Error("buildGenerateConfig() with invalid progress mode should return error")
	}
	if !strings.Contains(err.Error(), "invalid --progress value") {
		t.Errorf("error should mention invalid progress value, got: %v", err)
	}
}

func TestBuildGenerateConfig_WithCustomVars(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().StringSlice("include", []string{"*"}, "")
	cmd.Flags().StringSlice("exclude", nil, "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("max-size", "10MB", "")
	cmd.Flags().Bool("enforce-limit", true, "")
	cmd.Flags().Bool("send-gemini", false, "")
	cmd.Flags().String("gemini-model", "", "")
	cmd.Flags().String("gemini-output", "", "")
	cmd.Flags().Int("gemini-timeout", 0, "")
	cmd.Flags().String("template", "", "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("rules", "", "")
	cmd.Flags().StringArray("var", []string{"KEY1=value1", "KEY2=value2"}, "")
	cmd.Flags().Int("workers", 0, "")
	cmd.Flags().Bool("include-hidden", false, "")
	cmd.Flags().Bool("include-ignored", false, "")
	cmd.Flags().String("progress", "none", "")

	dir := t.TempDir()
	_ = cmd.Flags().Set("root", dir)

	cfg, err := buildGenerateConfig(cmd)
	if err != nil {
		t.Fatalf("buildGenerateConfig() error: %v", err)
	}
	if cfg.CustomVars["KEY1"] != "value1" {
		t.Errorf("expected KEY1=value1, got KEY1=%s", cfg.CustomVars["KEY1"])
	}
	if cfg.CustomVars["KEY2"] != "value2" {
		t.Errorf("expected KEY2=value2, got KEY2=%s", cfg.CustomVars["KEY2"])
	}
}

func TestBuildGenerateConfig_InvalidVarFormat(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("root", ".", "")
	cmd.Flags().StringSlice("include", []string{"*"}, "")
	cmd.Flags().StringSlice("exclude", nil, "")
	cmd.Flags().String("output", "", "")
	cmd.Flags().String("max-size", "10MB", "")
	cmd.Flags().Bool("enforce-limit", true, "")
	cmd.Flags().Bool("send-gemini", false, "")
	cmd.Flags().String("gemini-model", "", "")
	cmd.Flags().String("gemini-output", "", "")
	cmd.Flags().Int("gemini-timeout", 0, "")
	cmd.Flags().String("template", "", "")
	cmd.Flags().String("task", "", "")
	cmd.Flags().String("rules", "", "")
	cmd.Flags().StringArray("var", []string{"INVALID_NO_EQUALS"}, "")
	cmd.Flags().Int("workers", 0, "")
	cmd.Flags().Bool("include-hidden", false, "")
	cmd.Flags().Bool("include-ignored", false, "")
	cmd.Flags().String("progress", "none", "")

	dir := t.TempDir()
	_ = cmd.Flags().Set("root", dir)

	_, err := buildGenerateConfig(cmd)
	if err == nil {
		t.Error("buildGenerateConfig() with invalid var format should return error")
	}
	if !strings.Contains(err.Error(), "invalid --var format") {
		t.Errorf("error should mention invalid var format, got: %v", err)
	}
}

func TestBuildScannerConfig(t *testing.T) {
	viper.Reset()
	viper.Set("scanner.max-files", int64(5000))
	viper.Set("scanner.max-file-size", "2MB")
	viper.Set("scanner.max-memory", "200MB")
	viper.Set("scanner.skip-binary", true)
	viper.Set("scanner.include-hidden", false)
	viper.Set("scanner.include-ignored", false)
	viper.Set("scanner.workers", 4)
	viper.Set("scanner.respect-gitignore", true)
	viper.Set("scanner.respect-shotgunignore", true)

	cfg := GenerateConfig{
		Include: []string{"*.go"},
		Exclude: []string{"vendor/*"},
		Workers: 8,
	}

	scanCfg := buildScannerConfig(cfg)

	if scanCfg.MaxFiles != 5000 {
		t.Errorf("expected MaxFiles=5000, got %d", scanCfg.MaxFiles)
	}
	if scanCfg.Workers != 8 {
		t.Errorf("expected Workers=8 (from override), got %d", scanCfg.Workers)
	}
	if len(scanCfg.IncludePatterns) != 1 || scanCfg.IncludePatterns[0] != "*.go" {
		t.Errorf("expected IncludePatterns=[*.go], got %v", scanCfg.IncludePatterns)
	}
	if len(scanCfg.IgnorePatterns) != 1 || scanCfg.IgnorePatterns[0] != "vendor/*" {
		t.Errorf("expected IgnorePatterns=[vendor/*], got %v", scanCfg.IgnorePatterns)
	}
}

func TestBuildTemplateVars(t *testing.T) {
	cfg := GenerateConfig{
		Task:  "Analyze this code",
		Rules: "Be concise",
		CustomVars: map[string]string{
			"CUSTOM_KEY": "custom_value",
		},
	}

	vars := buildTemplateVars(cfg)

	if vars["TASK"] != "Analyze this code" {
		t.Errorf("expected TASK='Analyze this code', got '%s'", vars["TASK"])
	}
	if vars["RULES"] != "Be concise" {
		t.Errorf("expected RULES='Be concise', got '%s'", vars["RULES"])
	}
	if vars["CUSTOM_KEY"] != "custom_value" {
		t.Errorf("expected CUSTOM_KEY='custom_value', got '%s'", vars["CUSTOM_KEY"])
	}
	if vars["CURRENT_DATE"] == "" {
		t.Error("expected CURRENT_DATE to be set")
	}
}

func TestBuildTemplateVars_DefaultTask(t *testing.T) {
	cfg := GenerateConfig{
		Task: "",
	}

	vars := buildTemplateVars(cfg)

	if vars["TASK"] != "Context generation" {
		t.Errorf("expected default TASK='Context generation', got '%s'", vars["TASK"])
	}
}

func TestClearProgressLine(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	clearProgressLine(ProgressHuman)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("clearProgressLine(ProgressHuman) should produce output")
	}
}

func TestClearProgressLine_NoOutput(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	clearProgressLine(ProgressJSON)

	_ = w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if output != "" {
		t.Error("clearProgressLine(ProgressJSON) should produce no output")
	}
}
