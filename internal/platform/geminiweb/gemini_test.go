package geminiweb

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Model != "gemini-2.5-flash" {
		t.Errorf("expected default model gemini-2.5-flash, got %s", cfg.Model)
	}

	if cfg.Timeout != 300 {
		t.Errorf("expected default timeout 300, got %d", cfg.Timeout)
	}

	if cfg.BrowserRefresh != "auto" {
		t.Errorf("expected default browser-refresh auto, got %s", cfg.BrowserRefresh)
	}

	if cfg.Verbose {
		t.Error("expected verbose to be false by default")
	}

	if cfg.BinaryPath != "" {
		t.Errorf("expected empty binary path by default, got %s", cfg.BinaryPath)
	}
}

func TestValidModels(t *testing.T) {
	models := ValidModels()

	if len(models) != 3 {
		t.Errorf("expected 3 valid models, got %d", len(models))
	}

	expected := []string{"gemini-2.5-flash", "gemini-2.5-pro", "gemini-3.0-pro"}
	for i, model := range expected {
		if models[i] != model {
			t.Errorf("expected model %s at index %d, got %s", model, i, models[i])
		}
	}
}

func TestIsValidModel(t *testing.T) {
	// Model validation removed - all models should now be valid
	tests := []struct {
		model string
		valid bool
	}{
		{"gemini-2.5-flash", true},
		{"gemini-2.5-pro", true},
		{"gemini-3.0-pro", true},
		{"gemini-3-pro-preview", true}, // Custom/preview models now allowed
		{"gemini-invalid", true},
		{"gpt-4", true},
		{"", true},
		{"any-custom-model", true},
	}

	for _, tt := range tests {
		result := IsValidModel(tt.model)
		if result != tt.valid {
			t.Errorf("IsValidModel(%q) = %v, want %v", tt.model, result, tt.valid)
		}
	}
}

func TestConfigFindBinary_NotFound(t *testing.T) {
	cfg := Config{
		BinaryPath: "/nonexistent/path/to/geminiweb",
	}

	_, err := cfg.FindBinary()
	if err == nil {
		t.Error("expected error for nonexistent binary path")
	}
}

func TestIsAvailable(t *testing.T) {
	// This test is informative - doesn't fail if geminiweb isn't installed
	available := IsAvailable()
	t.Logf("geminiweb available: %v", available)
}

func TestIsConfigured(t *testing.T) {
	// This test is informative
	configured := IsConfigured()
	t.Logf("geminiweb configured: %v", configured)
}

func TestGetCookiesPath(t *testing.T) {
	path := GetCookiesPath()
	if path == "" {
		t.Error("expected non-empty cookies path")
	}
	if !contains(path, ".geminiweb") || !contains(path, "cookies.json") {
		t.Errorf("unexpected cookies path format: %s", path)
	}
}

func TestGetStatus(t *testing.T) {
	status := GetStatus()
	t.Logf("Status: available=%v, configured=%v, binary=%s, error=%s",
		status.Available, status.Configured, status.BinaryPath, status.Error)

	// CookiesPath should always be set
	if status.CookiesPath == "" {
		t.Error("expected non-empty cookies path in status")
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello world"},
		{"\x1b[32mgreen\x1b[0m", "green"},
		{"\x1b[1;31mbold red\x1b[0m", "bold red"},
		{"normal \x1b[4munderline\x1b[0m normal", "normal underline normal"},
		{"\x1b[38;5;196mcolor\x1b[0m", "color"},
		{"no codes here", "no codes here"},
		{"", ""},
		{"\x1b[0m", ""},
	}

	for _, tt := range tests {
		result := StripANSI(tt.input)
		if result != tt.expected {
			t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name: "with gemini header",
			input: `✦ Gemini

Here is the response.

With multiple lines.`,
			expected: `Here is the response.

With multiple lines.`,
		},
		{
			name:     "without header",
			input:    "Just a plain response.",
			expected: "Just a plain response.",
		},
		{
			name: "with whitespace",
			input: `

Response text.

  `,
			expected: "Response text.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseResponse(tt.input)
			if result != tt.expected {
				t.Errorf("ParseResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractCodeBlocks(t *testing.T) {
	input := `Here's some code:

` + "```go" + `
func main() {
    fmt.Println("hello")
}
` + "```" + `

And more:

` + "```python" + `
print('hello')
` + "```" + `

That's all.`

	blocks := ExtractCodeBlocks(input)

	if len(blocks) != 2 {
		t.Fatalf("expected 2 code blocks, got %d", len(blocks))
	}

	if blocks[0].Language != "go" {
		t.Errorf("expected first block language 'go', got %s", blocks[0].Language)
	}

	if blocks[1].Language != "python" {
		t.Errorf("expected second block language 'python', got %s", blocks[1].Language)
	}

	if !contains(blocks[0].Code, "func main()") {
		t.Error("expected first block to contain 'func main()'")
	}

	if !contains(blocks[1].Code, "print") {
		t.Error("expected second block to contain 'print'")
	}
}

func TestExtractCodeBlocks_Empty(t *testing.T) {
	blocks := ExtractCodeBlocks("No code blocks here.")
	if len(blocks) != 0 {
		t.Errorf("expected 0 code blocks, got %d", len(blocks))
	}
}

func TestSummarizeResponse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "short response",
			input:     "Short text.",
			maxLength: 100,
			expected:  "Short text.",
		},
		{
			name:      "truncate at sentence",
			input:     "First sentence. Second sentence. Third sentence.",
			maxLength: 25,
			expected:  "First sentence....",
		},
		{
			name:      "truncate at word",
			input:     "One two three four five six seven eight",
			maxLength: 20,
			expected:  "One two three four...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SummarizeResponse(tt.input, tt.maxLength)
			if len(result) > tt.maxLength+3 { // +3 for "..."
				t.Errorf("result too long: %d > %d", len(result), tt.maxLength+3)
			}
		})
	}
}

func TestContainsError(t *testing.T) {
	tests := []struct {
		input    string
		hasError bool
	}{
		{"Here's your answer.", false},
		{"I apologize, but I cannot help with that.", true},
		{"Error: rate limit exceeded", true},
		{"The authentication failed.", true},
		{"Normal response text.", false},
	}

	for _, tt := range tests {
		hasError, msg := ContainsError(tt.input)
		if hasError != tt.hasError {
			t.Errorf("ContainsError(%q) = %v, want %v (msg: %s)", tt.input, hasError, tt.hasError, msg)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{30 * time.Second, "30.0s"},
		{90 * time.Second, "1m30s"},
		{150 * time.Second, "2m30s"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
		}
	}
}

func TestNewExecutor(t *testing.T) {
	cfg := DefaultConfig()
	executor := NewExecutor(cfg)

	if executor == nil {
		t.Fatal("expected non-nil executor")

		return // unreachable but satisfies staticcheck
	}

	if executor.config.Model != cfg.Model {
		t.Errorf("executor config model mismatch: got %s, want %s", executor.config.Model, cfg.Model)
	}
}

func TestExecutor_buildArgs(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected []string
	}{
		{
			name: "default config",
			config: Config{
				Model:          "gemini-2.5-flash",
				BrowserRefresh: "auto",
			},
			expected: []string{"-m", "gemini-2.5-flash", "--browser-refresh", "auto"},
		},
		{
			name: "custom model",
			config: Config{
				Model: "gemini-3.0-pro",
			},
			expected: []string{"-m", "gemini-3.0-pro"},
		},
		{
			name:     "empty config",
			config:   Config{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.config)
			args := executor.buildArgs()

			if len(args) != len(tt.expected) {
				t.Errorf("expected %d args, got %d: %v", len(tt.expected), len(args), args)

				return
			}

			for i, arg := range args {
				if arg != tt.expected[i] {
					t.Errorf("arg[%d] = %q, want %q", i, arg, tt.expected[i])
				}
			}
		})
	}
}

func TestExecutor_Send_NotAvailable(t *testing.T) {
	cfg := Config{
		BinaryPath: "/nonexistent/geminiweb",
	}
	executor := NewExecutor(cfg)

	_, err := executor.Send(context.Background(), "test content")
	if err == nil {
		t.Error("expected error when binary not available")
	}
}

func TestExtractThoughts(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedThought string
		hasThoughts     bool
	}{
		{
			name:            "no thoughts",
			input:           "Just a regular response.",
			expectedThought: "",
			hasThoughts:     false,
		},
		{
			name:            "with thinking tags",
			input:           "<thinking>Some reasoning</thinking>The actual response.",
			expectedThought: "<thinking>Some reasoning</thinking>",
			hasThoughts:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thoughts, _ := ExtractThoughts(tt.input)
			if tt.hasThoughts && thoughts == "" {
				t.Error("expected thoughts to be extracted")
			}
			if !tt.hasThoughts && thoughts != "" {
				t.Errorf("unexpected thoughts: %s", thoughts)
			}
		})
	}
}

func TestConfig_FindBinary_ExplicitPath(t *testing.T) {
	// Test with a path that exists (using go binary as example)
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go binary not found in PATH")
	}

	cfg := Config{BinaryPath: goPath}
	path, err := cfg.FindBinary()
	if err != nil {
		t.Errorf("expected no error for existing path, got: %v", err)
	}
	if path != goPath {
		t.Errorf("expected path %s, got %s", goPath, path)
	}
}

func TestParseResponse_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "only whitespace",
			input:    "   \n\n   ",
			expected: "",
		},
		{
			name:     "gemini after line 3",
			input:    "Line 1\nLine 2\nLine 3\nGemini is here\nMore text",
			expected: "Line 1\nLine 2\nLine 3\nGemini is here\nMore text",
		},
		{
			name:     "multiple lines after header",
			input:    "✦ Gemini\n\nLine 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseResponse(tt.input)
			if result != tt.expected {
				t.Errorf("ParseResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSummarizeResponse_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
	}{
		{
			name:      "empty string",
			input:     "",
			maxLength: 10,
		},
		{
			name:      "exact length",
			input:     "12345",
			maxLength: 5,
		},
		{
			name:      "very short max",
			input:     "Hello world, this is a test.",
			maxLength: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SummarizeResponse(tt.input, tt.maxLength)
			// Just verify it doesn't panic and returns something
			if len(result) > tt.maxLength+3 && tt.input != "" {
				t.Errorf("result too long: %d > %d", len(result), tt.maxLength+3)
			}
		})
	}
}

func TestContainsError_AllIndicators(t *testing.T) {
	indicators := []string{
		"I apologize for any confusion",
		"I cannot help with that request",
		"I'm unable to process this",
		"Error: something went wrong",
		"An error occurred during processing",
		"Rate limit exceeded, please try again",
		"Quota exceeded for this request",
		"Authentication failed: invalid credentials",
		"Request unauthorized",
	}

	for _, indicator := range indicators {
		hasError, msg := ContainsError(indicator)
		if !hasError {
			t.Errorf("ContainsError(%q) should return true", indicator)
		}
		if msg == "" {
			t.Errorf("ContainsError(%q) should return a message", indicator)
		}
	}
}

func TestFormatDuration_AllRanges(t *testing.T) {
	tests := []struct {
		duration time.Duration
		contains string
	}{
		{0, "0ms"},
		{999 * time.Millisecond, "ms"},
		{1 * time.Second, "s"},
		{59 * time.Second, "s"},
		{60 * time.Second, "m"},
		{5 * time.Minute, "m"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.duration)
		if !strings.Contains(result, tt.contains) {
			t.Errorf("FormatDuration(%v) = %q, expected to contain %q", tt.duration, result, tt.contains)
		}
	}
}

func TestExtractCodeBlocks_VariousLanguages(t *testing.T) {
	input := "```javascript\nconsole.log('hello');\n```\n\n```\nplain code\n```\n\n```rust\nfn main() {}\n```"

	blocks := ExtractCodeBlocks(input)
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(blocks))
	}

	if blocks[0].Language != "javascript" {
		t.Errorf("expected javascript, got %s", blocks[0].Language)
	}

	if blocks[1].Language != "" {
		t.Errorf("expected empty language, got %s", blocks[1].Language)
	}

	if blocks[2].Language != "rust" {
		t.Errorf("expected rust, got %s", blocks[2].Language)
	}
}

func TestStripANSI_ComplexSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"no ansi", "hello", "hello"},
		{"reset only", "\x1b[0m", ""},
		{"bold", "\x1b[1mbold\x1b[0m", "bold"},
		{"color 256", "\x1b[38;5;196mred\x1b[0m", "red"},
		{"rgb color", "\x1b[38;2;255;0;0mred\x1b[0m", "red"},
		{"mixed content", "pre\x1b[32mgreen\x1b[0mpost", "pregreenpost"},
		{"incomplete sequence", "\x1b[", ""},
		{"newlines preserved", "line1\n\x1b[32mgreen\x1b[0m\nline2", "line1\ngreen\nline2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExecutor_BuildArgs_AllOptions(t *testing.T) {
	cfg := Config{
		Model:          "gemini-3.0-pro",
		BrowserRefresh: "chrome",
	}
	executor := NewExecutor(cfg)
	args := executor.buildArgs()

	if len(args) != 4 {
		t.Errorf("expected 4 args, got %d: %v", len(args), args)
	}

	// Check model flag
	if args[0] != "-m" || args[1] != "gemini-3.0-pro" {
		t.Errorf("expected -m gemini-3.0-pro, got %s %s", args[0], args[1])
	}

	// Check browser-refresh flag
	if args[2] != "--browser-refresh" || args[3] != "chrome" {
		t.Errorf("expected --browser-refresh chrome, got %s %s", args[2], args[3])
	}
}

func TestIsTerminatingByte(t *testing.T) {
	// Test lowercase letters
	for b := byte('a'); b <= 'z'; b++ {
		if !isTerminatingByte(b) {
			t.Errorf("expected %c to be terminating", b)
		}
	}

	// Test uppercase letters
	for b := byte('A'); b <= 'Z'; b++ {
		if !isTerminatingByte(b) {
			t.Errorf("expected %c to be terminating", b)
		}
	}

	// Test non-letters
	nonLetters := []byte{'0', '9', ';', '[', ' ', '\n'}
	for _, b := range nonLetters {
		if isTerminatingByte(b) {
			t.Errorf("expected %c to not be terminating", b)
		}
	}
}

func TestGetStatus_Details(t *testing.T) {
	status := GetStatus()

	// CookiesPath should always be set regardless of availability
	if status.CookiesPath == "" {
		t.Error("CookiesPath should never be empty")
	}

	// If available, binary path should be set
	if status.Available && status.BinaryPath == "" {
		t.Error("BinaryPath should be set when available")
	}

	// If not configured but available, should have an error message
	if status.Available && !status.Configured && status.Error == "" {
		t.Error("Error should be set when available but not configured")
	}
}

func TestExecutor_Send_Success(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available for integration test")
	}

	cfg := DefaultConfig()
	executor := NewExecutor(cfg)

	ctx := context.Background()
	result, err := executor.Send(ctx, "Test message for Gemini")

	if err != nil {
		t.Errorf("expected successful send, got error: %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result")

		return
	}

	if result.Response == "" {
		t.Error("expected non-empty response")
	}

	if result.Model != cfg.Model {
		t.Errorf("expected model %s, got %s", cfg.Model, result.Model)
	}

	if result.Duration == 0 {
		t.Error("expected non-zero duration")
	}

	if result.RawResponse == "" {
		t.Error("expected non-empty raw response")
	}
}

func TestExecutor_Send_Timeout(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available")
	}

	cfg := Config{
		Timeout: 1, // Very short timeout
	}
	executor := NewExecutor(cfg)

	ctx := context.Background()
	_, err := executor.Send(ctx, "This should timeout quickly")

	if err == nil {
		t.Error("expected timeout error")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "timeout") && !strings.Contains(errorMsg, "timed out") {
		t.Errorf("expected timeout-related error, got: %s", errorMsg)
	}
}

func TestExecutor_Send_ContextCancelled(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available")
	}

	cfg := DefaultConfig()
	executor := NewExecutor(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := executor.Send(ctx, "Test with canceled context")

	if err == nil {
		t.Error("expected error for canceled context")
	}

	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "cancel") {
		t.Errorf("expected context-related error, got: %s", err.Error())
	}
}

func TestExecutor_Send_NotConfigured(t *testing.T) {
	// Use a fake binary that exists but is not geminiweb
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go binary not found for testing")
	}

	cfg := Config{
		BinaryPath: goPath,
		Timeout:    10, // Higher timeout to avoid timeout errors
	}
	executor := NewExecutor(cfg)

	ctx := context.Background()
	_, err = executor.Send(ctx, "Test with non-gemini binary")

	if err == nil {
		t.Error("expected error when geminiweb is not configured")
	}

	errorMsg := err.Error()
	// Should either get configuration error or execution error
	if !strings.Contains(errorMsg, "configured") &&
		!strings.Contains(errorMsg, "auto-login") &&
		!strings.Contains(errorMsg, "execution failed") {
		t.Errorf("expected configuration or execution error, got: %s", errorMsg)
	}
}

func TestExecutor_Send_ExecutionError(t *testing.T) {
	t.Skip("Skipping: IsConfigured() check happens before command execution, " +
		"so we can't test execution errors without a configured geminiweb")

	// Test with a binary that exists but will fail quickly
	cfg := Config{
		BinaryPath: "/bin/sh",
		Timeout:    5, // Short timeout
	}
	executor := NewExecutor(cfg)

	ctx := context.Background()
	_, err := executor.Send(ctx, "-c 'exit 1'")

	if err == nil {
		t.Error("expected error when execution fails")
	}

	errorMsg := err.Error()
	// Should get either execution failed or timeout, both are acceptable
	if !strings.Contains(errorMsg, "execution failed") &&
		!strings.Contains(errorMsg, "timeout") {
		t.Errorf("expected execution or timeout error, got: %s", errorMsg)
	}
}

func TestConfig_FindBinary_SearchPath(t *testing.T) {
	// Test with empty config - should search PATH
	cfg := Config{}

	path, err := cfg.FindBinary()

	if err != nil {
		// This is OK - geminiweb might not be installed
		t.Logf("geminiweb not found in PATH (this is expected if not installed): %v", err)

		return
	}

	if path == "" {
		t.Error("expected non-empty path when found")
	}

	// Verify it's actually the geminiweb binary
	if !strings.Contains(path, "geminiweb") {
		t.Errorf("expected path to contain 'geminiweb', got: %s", path)
	}
}

func TestConfig_FindBinary_CommonPaths(t *testing.T) {
	// Create a temporary "geminiweb" file in a common location
	tmpDir := t.TempDir()
	commonPath := filepath.Join(tmpDir, "bin", "geminiweb")

	err := os.MkdirAll(filepath.Dir(commonPath), 0o750)
	if err != nil {
		t.Fatal(err)
	}

	// Create a fake geminiweb binary
	file, err := os.Create(commonPath) //nolint:gosec // test file creation
	if err != nil {
		t.Fatal(err)
	}
	_, _ = file.WriteString("#!/bin/bash\necho fake geminiweb")
	_ = file.Close()
	_ = os.Chmod(commonPath, 0o700) //nolint:gosec // test executable needs exec permission

	// Test with GOPATH set to temp dir and clear PATH to avoid finding real geminiweb
	oldGopath := os.Getenv("GOPATH")
	oldPath := os.Getenv("PATH")
	_ = os.Setenv("GOPATH", tmpDir)
	_ = os.Setenv("PATH", tmpDir+":/bin:/usr/bin") // Minimal PATH without real geminiweb
	defer func() { _ = os.Setenv("GOPATH", oldGopath) }()
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	cfg := Config{} // Empty config should search common paths
	path, err := cfg.FindBinary()

	if err != nil {
		t.Errorf("expected to find binary in common path, got error: %v", err)
	}

	if path != commonPath {
		t.Errorf("expected path %s, got %s", commonPath, path)
	}
}

func TestConfig_FindBinary_ExplicitPathNotFound(t *testing.T) {
	cfg := Config{BinaryPath: "/definitely/does/not/exist/geminiweb"}

	_, err := cfg.FindBinary()

	if err == nil {
		t.Error("expected error for nonexistent explicit path")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "not found at specified path") {
		t.Errorf("expected 'not found at specified path' error, got: %s", errorMsg)
	}
}

func TestExecutor_SendWithProgress_Success(t *testing.T) {
	if !IsAvailable() {
		t.Skip("geminiweb not available for integration test")
	}

	cfg := DefaultConfig()
	executor := NewExecutor(cfg)

	var progressStages []string
	progress := func(stage string) {
		progressStages = append(progressStages, stage)
	}

	ctx := context.Background()
	result, err := executor.SendWithProgress(ctx, "Test message for progress", progress)

	if err != nil {
		t.Errorf("expected successful send with progress, got error: %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result")
	}

	if len(progressStages) == 0 {
		t.Error("expected progress updates to be called")
	}

	// Verify expected progress stages were called
	expectedStages := []string{"Locating geminiweb...", "Preparing request...", "Processing response..."}
	for _, expected := range expectedStages {
		found := false
		for _, actual := range progressStages {
			if strings.Contains(actual, expected) {
				found = true

				break
			}
		}
		if !found {
			t.Errorf("expected progress stage '%s' to be called", expected)
		}
	}
}

func TestExtractThoughts_Complex(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedThought string
		hasThoughts     bool
		expectedRest    string
	}{
		{
			name:            "simple thinking tags",
			input:           "<thinking>Some reasoning</thinking>Actual response.",
			expectedThought: "<thinking>Some reasoning</thinking>",
			hasThoughts:     true,
			expectedRest:    "Actual response.",
		},
		{
			name:            "thinking with newlines and double newline ending",
			input:           "<thinking>\nComplex reasoning\nwith multiple lines\n</thinking>\n\nThe actual answer.",
			expectedThought: "<thinking>\nComplex reasoning\nwith multiple lines\n</thinking>",
			hasThoughts:     true,
			expectedRest:    "The actual answer.",
		},
		{
			name:            "emoji thinking marker",
			input:           "💭 This is thinking\n\nActual response.",
			expectedThought: "💭 This is thinking",
			hasThoughts:     true,
			expectedRest:    "Actual response.",
		},
		{
			name:            "markdown thinking marker",
			input:           "**Thinking:** This is thinking\n\nActual response.",
			expectedThought: "**Thinking:** This is thinking",
			hasThoughts:     true,
			expectedRest:    "Actual response.",
		},
		{
			name:            "malformed thinking tags",
			input:           "<thinking unclosed>Some reasoning</thinking>Response",
			expectedThought: "",
			hasThoughts:     false,
			expectedRest:    "<thinking unclosed>Some reasoning</thinking>Response",
		},
		{
			name:            "empty thinking tags",
			input:           "<thinking></thinking>Response",
			expectedThought: "<thinking></thinking>",
			hasThoughts:     true,
			expectedRest:    "Response",
		},
		{
			name:            "multiple thinking sections - first wins",
			input:           "<thinking>First reasoning</thinking>Middle <thinking>Second reasoning</thinking>",
			expectedThought: "<thinking>First reasoning</thinking>",
			hasThoughts:     true,
			expectedRest:    "Middle <thinking>Second reasoning</thinking>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thoughts, rest := ExtractThoughts(tt.input)

			if thoughts != tt.expectedThought {
				t.Errorf("ExtractThoughts() thoughts = %q, want %q", thoughts, tt.expectedThought)
			}

			if rest != tt.expectedRest {
				t.Errorf("ExtractThoughts() rest = %q, want %q", rest, tt.expectedRest)
			}

			hasThoughts := thoughts != ""
			if hasThoughts != tt.hasThoughts {
				t.Errorf("ExtractThoughts() hasThoughts = %v, want %v", hasThoughts, tt.hasThoughts)
			}
		})
	}
}

func TestSendWithProgress_BinaryNotFound(t *testing.T) {
	t.Parallel()

	cfg := Config{BinaryPath: "/nonexistent/path/to/geminiweb"}
	executor := NewExecutor(cfg)

	progress := func(stage string) {
		// No-op progress function for testing
	}
	ctx := context.Background()

	_, err := executor.SendWithProgress(ctx, "test content", progress)

	if err == nil {
		t.Error("expected error for nonexistent binary")
	}
}

func TestSendWithProgress_ContextCancelled(t *testing.T) {
	t.Parallel()

	if !IsAvailable() {
		t.Skip("geminiweb not available")
	}

	cfg := DefaultConfig()
	executor := NewExecutor(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := executor.SendWithProgress(ctx, "test", func(stage string) {})

	if err == nil {
		t.Error("expected error on canceled context")
	}
}

func TestConfigWithAllOptions(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Model:          "gemini-3.0-pro",
		Timeout:        120,
		BrowserRefresh: "never",
		Verbose:        true,
		BinaryPath:     "/custom/path/geminiweb",
	}

	executor := NewExecutor(cfg)
	args := executor.buildArgs()

	// Should contain model
	foundModel := false
	for i, arg := range args {
		if arg == "-m" && i+1 < len(args) && args[i+1] == "gemini-3.0-pro" {
			foundModel = true

			break
		}
	}
	if !foundModel {
		t.Errorf("expected args to contain model flag, got: %v", args)
	}

	// Should contain browser-refresh
	foundBrowser := false
	for i, arg := range args {
		if arg == "--browser-refresh" && i+1 < len(args) && args[i+1] == "never" {
			foundBrowser = true

			break
		}
	}
	if !foundBrowser {
		t.Errorf("expected args to contain browser-refresh flag, got: %v", args)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
