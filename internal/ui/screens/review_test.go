package screens

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
)

func TestNewReview(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{
		"/path/to/file1.go": true,
		"/path/to/file2.go": true,
	}
	fileTree := &scanner.FileNode{
		Name:     "root",
		Path:     "/path",
		IsDir:    true,
		Size:     1024,
		Children: []*scanner.FileNode{},
	}
	tmpl := &template.Template{
		Name:        "Test Template",
		Description: "A test template",
		Content:     "Test content",
	}

	m := NewReview(selectedFiles, fileTree, tmpl, "Test task", "Test rules", "")

	if m == nil {
		t.Fatalf("NewReview returned nil")
	}
	if len(m.selectedFiles) != 2 {
		t.Fatalf("expected 2 selected files, got %d", len(m.selectedFiles))
	}
	if m.template == nil || m.template.Name != "Test Template" {
		t.Fatalf("template not set correctly")
	}
	if m.taskDesc != "Test task" {
		t.Fatalf("task description not set correctly")
	}
	if m.rules != "Test rules" {
		t.Fatalf("rules not set correctly")
	}
}

func TestNewReview_NilTemplate(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:     "root",
		Path:     "/path",
		IsDir:    true,
		Children: []*scanner.FileNode{},
	}

	m := NewReview(selectedFiles, fileTree, nil, "Task", "Rules", "")

	if m == nil {
		t.Fatalf("NewReview returned nil")
	}
	if m.template != nil {
		t.Fatalf("template should be nil")
	}
}

func TestReviewModel_SetSize(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "", "", "")
	m.SetSize(100, 50)

	if m.width != 100 {
		t.Fatalf("expected width 100, got %d", m.width)
	}
	if m.height != 50 {
		t.Fatalf("expected height 50, got %d", m.height)
	}
}

func TestReviewModel_UpdateCtrlC(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "", "", "")
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatalf("expected tea.Quit command for Ctrl+C")
	}
}

func TestReviewModel_UpdateOtherKey(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "", "", "")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd := m.Update(msg)

	if cmd != nil {
		t.Fatalf("expected nil command for non-handled key")
	}
}

func TestReviewModel_SetGenerated(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "", "", "")

	m.SetGenerated("/tmp/test.md", true)

	if !m.generated {
		t.Fatalf("expected generated to be true")
	}
	if m.generatedPath != "/tmp/test.md" {
		t.Fatalf("expected generatedPath to be '/tmp/test.md', got %s", m.generatedPath)
	}
	if !m.clipboardCopied {
		t.Fatalf("expected clipboardCopied to be true")
	}
}

func TestReviewModel_SetGenerated_FalseClipboard(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "", "", "")

	m.SetGenerated("/tmp/test.md", false)

	if !m.generated {
		t.Fatalf("expected generated to be true")
	}
	if m.clipboardCopied {
		t.Fatalf("expected clipboardCopied to be false")
	}
}

func TestReviewModel_View_BeforeGeneration(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{
		"/path/to/file1.go": true,
		"/path/to/file2.go": true,
	}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Size:  1024,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
			{
				Name:  "file2.go",
				Path:  "/path/to/file2.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{
		Name:        "Test Template",
		Description: "A test template",
		Content:     "Test content",
	}

	m := NewReview(selectedFiles, fileTree, tmpl, "Test task description", "Test rules", "")
	m.SetSize(80, 24)

	view := m.View()

	if !strings.Contains(view, "Review & Generate") {
		t.Fatalf("expected 'Review & Generate' in view")
	}
	if !strings.Contains(view, "Selected Files:") {
		t.Fatalf("expected 'Selected Files:' in view")
	}
	if !strings.Contains(view, "Test Template") {
		t.Fatalf("expected template name in view")
	}
	if !strings.Contains(view, "Test task description") {
		t.Fatalf("expected task description in view")
	}
	if !strings.Contains(view, "F8: Generate") {
		t.Fatalf("expected 'F8: Generate' in footer")
	}
}

func TestReviewModel_View_AfterGeneration(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{
		"/path/to/file1.go": true,
	}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Size:  512,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{
		Name:    "Test Template",
		Content: "Test content",
	}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 40)
	m.SetGenerated("/tmp/test.md", true)

	view := m.View()

	if !strings.Contains(view, "Context generated successfully") {
		t.Fatalf("expected success message in view")
	}
	if !strings.Contains(view, "/tmp/test.md") {
		t.Fatalf("expected generated path in view")
	}
	if !strings.Contains(view, "📋 Copied to clipboard") {
		t.Fatalf("expected clipboard success message")
	}
}

func TestReviewModel_View_AfterGeneration_ClipboardFailed(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 40)
	m.SetGenerated("/tmp/test.md", false)

	view := m.View()

	if !strings.Contains(view, "Context generated successfully!") {
		t.Fatalf("expected success message")
	}
	if !strings.Contains(view, "Clipboard copy failed") {
		t.Fatalf("expected clipboard failure message")
	}
}

func TestReviewModel_View_NoTemplate(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}

	m := NewReview(selectedFiles, fileTree, nil, "Task", "Rules", "")
	m.SetSize(80, 40)
	view := m.View()

	if !strings.Contains(view, "Template:") {
		t.Fatalf("expected template section")
	}
	if !strings.Contains(view, "None selected") {
		t.Fatalf("expected 'None selected' message")
	}
}

func TestReviewModel_View_NoRules(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "", "")
	m.SetSize(80, 40)
	view := m.View()

	// Rules section should not appear when empty
	if strings.Contains(view, "Rules & Constraints:") {
		t.Fatalf("rules section should not appear for empty rules")
	}
}

func TestReviewModel_View_LongTaskDescription(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	longTask := strings.Repeat("a", 200)
	m := NewReview(selectedFiles, fileTree, tmpl, longTask, "", "")
	m.SetSize(80, 100)
	view := m.View()

	if strings.Contains(view, strings.Repeat("a", 200)) {
		t.Fatalf("full task should not appear - expected truncation")
	}
	if strings.Contains(view, strings.Repeat("a", 148)) {
		t.Fatalf("task should be truncated at 147 chars")
	}
}

func TestReviewModel_View_LongRules(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	longRules := strings.Repeat("r", 150)
	m := NewReview(selectedFiles, fileTree, tmpl, "Task", longRules, "")
	m.SetSize(80, 100)
	view := m.View()

	if !strings.Contains(view, "Rules & Constraints:") {
		t.Fatalf("expected rules section for non-empty rules")
	}
	if strings.Contains(view, strings.Repeat("r", 150)) {
		t.Fatalf("full rules should not appear - expected truncation")
	}
	if strings.Contains(view, strings.Repeat("r", 98)) {
		t.Fatalf("rules should be truncated at 97 chars")
	}
}

func TestReviewModel_calculateStats(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{
		"/path/to/file1.go": true,
		"/path/to/file2.go": true,
		"/path/to/file3.go": false, // Not selected
	}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Size:  1536,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
			{
				Name:  "file2.go",
				Path:  "/path/to/file2.go",
				IsDir: false,
				Size:  512,
			},
			{
				Name:  "file3.go",
				Path:  "/path/to/file3.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{
		Name:        "Test",
		Description: "Test template",
		Content:     "Test content body",
	}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task description", "Rule one, rule two", "")

	totalBytes, totalTokens := m.calculateStats()

	// Should count only selected files (512 + 512 = 1024)
	// Plus overhead from task (15 chars), rules (19 chars), and template (18 chars)
	expectedBytes := int64(1024 + 15 + 19 + 18)
	if totalBytes < expectedBytes-100 || totalBytes > expectedBytes+100 {
		t.Fatalf("expected bytes around %d, got %d", expectedBytes, totalBytes)
	}
	if totalTokens <= 0 {
		t.Fatalf("expected positive token count, got %d", totalTokens)
	}
}

func TestReviewModel_calculateStats_NoFileTree(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	tmpl := &template.Template{Name: "Test", Content: "Content"}
	m := NewReview(selectedFiles, nil, tmpl, "Task", "Rules", "")

	totalBytes, _ := m.calculateStats()

	// Should only count overhead
	if totalBytes <= 0 {
		t.Fatalf("expected positive bytes count")
	}
}

func TestReviewModel_calculateStats_NoTemplate(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	m := NewReview(selectedFiles, fileTree, nil, "Task", "Rules", "")

	totalBytes, _ := m.calculateStats()

	// Should count only file and overhead (no template)
	if totalBytes <= 512 {
		t.Fatalf("expected bytes to include file size")
	}
}

func TestReviewModel_walkTree(t *testing.T) {
	t.Parallel()

	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "child1",
				Path:  "/path/child1",
				IsDir: false,
			},
			{
				Name:  "child2",
				Path:  "/path/child2",
				IsDir: true,
				Children: []*scanner.FileNode{
					{
						Name:  "grandchild",
						Path:  "/path/child2/grandchild",
						IsDir: false,
					},
				},
			},
		},
	}

	m := NewReview(nil, nil, nil, "", "", "")
	visited := make(map[string]bool)

	m.walkTree(fileTree, func(node *scanner.FileNode, path string) {
		visited[path] = true
	})

	// Should have visited root, child1, child2, and grandchild
	if !visited["/path"] {
		t.Fatalf("expected to visit root")
	}
	if !visited["/path/child1"] {
		t.Fatalf("expected to visit child1")
	}
	if !visited["/path/child2"] {
		t.Fatalf("expected to visit child2")
	}
	if !visited["/path/child2/grandchild"] {
		t.Fatalf("expected to visit grandchild")
	}
}

func TestFormatSizeHelper_Bytes(t *testing.T) {
	t.Parallel()

	result := formatSizeHelper(42)
	if result != "42 B" {
		t.Fatalf("expected '42 B', got '%s'", result)
	}
}

func TestFormatSizeHelper_KB(t *testing.T) {
	t.Parallel()

	result := formatSizeHelper(1024)
	if result != "1.0 KB" {
		t.Fatalf("expected '1.0 KB', got '%s'", result)
	}
}

func TestFormatSizeHelper_MB(t *testing.T) {
	t.Parallel()

	result := formatSizeHelper(1024 * 1024)
	if result != "1.0 MB" {
		t.Fatalf("expected '1.0 MB', got '%s'", result)
	}
}

func TestFormatSizeHelper_GB(t *testing.T) {
	t.Parallel()

	result := formatSizeHelper(1024 * 1024 * 1024)
	if result != "1.0 GB" {
		t.Fatalf("expected '1.0 GB', got '%s'", result)
	}
}

func TestFormatSizeHelper_Large(t *testing.T) {
	t.Parallel()

	result := formatSizeHelper(1536) // 1.5 KB
	if !strings.HasPrefix(result, "1.5") {
		t.Fatalf("expected '1.5 KB', got '%s'", result)
	}
}

func TestReview_SetGeminiStates(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "", "", "")

	// Test SetGeminiSending
	m.SetLLMSending(true)
	if !m.llmSending {
		t.Fatalf("expected geminiSending to be true")
	}
	if m.llmError != nil {
		t.Fatalf("expected geminiError to be nil after SetGeminiSending")
	}

	// Test SetGeminiComplete
	duration := 5 * time.Second
	m.SetLLMComplete("/tmp/output.txt", duration)
	if m.llmSending {
		t.Fatalf("expected geminiSending to be false after SetGeminiComplete")
	}
	if !m.llmComplete {
		t.Fatalf("expected geminiComplete to be true")
	}
	if m.llmOutputFile != "/tmp/output.txt" {
		t.Fatalf("expected geminiOutputFile to be '/tmp/output.txt', got %s", m.llmOutputFile)
	}
	if m.llmDuration != duration {
		t.Fatalf("expected geminiDuration to be %v, got %v", duration, m.llmDuration)
	}
	if m.llmError != nil {
		t.Fatalf("expected geminiError to be nil after SetGeminiComplete")
	}

	// Test SetGeminiError
	testErr := errors.New("test error")
	m.SetLLMError(testErr)
	if m.llmSending {
		t.Fatalf("expected geminiSending to be false after SetGeminiError")
	}
	if m.llmComplete {
		t.Fatalf("expected geminiComplete to be false after SetGeminiError")
	}
	if m.llmError != testErr { //nolint:errorlint // testing exact error identity
		t.Fatalf("expected geminiError to be testErr")
	}
}

func TestReview_ViewGeminiSending(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 40)
	m.SetGenerated("/tmp/test.md", true)
	m.SetLLMSending(true)

	view := m.View()

	if !strings.Contains(view, "Sending to LLM...") {
		t.Fatalf("expected 'Sending to LLM...' in view during sending")
	}
}

func TestReview_ViewGeminiComplete(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 40)
	m.SetGenerated("/tmp/test.md", true)
	m.SetLLMComplete("/tmp/gemini-output.txt", 3*time.Second)

	view := m.View()

	// Just verify the view doesn't crash and contains some content
	if len(view) == 0 {
		t.Fatalf("expected non-empty view after gemini completion")
	}
}

func TestReview_ViewGeminiError(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 40)
	m.SetGenerated("/tmp/test.md", true)
	m.SetLLMError(errors.New("gemini connection failed"))

	view := m.View()

	if !strings.Contains(view, "Error:") {
		t.Fatalf("expected 'Error:' in view after error")
	}
	if !strings.Contains(view, "gemini connection failed") {
		t.Fatalf("expected error message in view")
	}
}

func TestReview_ViewGeminiStatus(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  512,
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "Content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 40)
	m.SetGenerated("/tmp/test.md", true)

	// Test view with no LLM state
	// The actual behavior depends on whether LLM is available in the test environment
	view := m.View()

	// Check for LLM Integration section (should always be present)
	if !strings.Contains(view, "LLM Integration:") {
		t.Fatalf("expected 'LLM Integration:' section in view")
	}

	// If LLM is available, we should see "Send to LLM" or "Ready"
	// If not available, we should see "Not configured"
	// We just verify the section exists, regardless of LLM availability
}

func TestReview_ViewSizeLimit(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file1.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "file1.go",
				Path:  "/path/to/file1.go",
				IsDir: false,
				Size:  1024 * 1024, // 1MB file
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: strings.Repeat("content", 1000)}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 24)

	view := m.View()

	// Should show some form of size information
	if !strings.Contains(view, "Size") && !strings.Contains(view, "1.0") {
		t.Fatalf("expected size information in view, got: %s", view)
	}
}

func TestReview_parseSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       string
		expectError bool
		expected    int64
	}{
		{"100", false, 100},
		{"1KB", false, 1024},
		{"2MB", false, 2 * 1024 * 1024},
		{"invalid", true, 0},
		{"", true, 0},
	}

	for _, tt := range tests {
		t.Run("parse "+tt.input, func(t *testing.T) {
			result, err := parseSize(tt.input)
			if tt.expectError && err == nil {
				t.Fatalf("expected error for input '%s'", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Fatalf("unexpected error for input '%s': %v", tt.input, err)
			}
			if !tt.expectError && result != tt.expected {
				t.Fatalf("expected %d, got %d for input '%s'", tt.expected, result, tt.input)
			}
		})
	}
}

func TestReview_calculateStatsLargeFile(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/large.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{
				Name:  "large.go",
				Path:  "/path/to/large.go",
				IsDir: false,
				Size:  10 * 1024 * 1024, // 10MB file
			},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: strings.Repeat("content", 10000)}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")

	totalBytes, totalTokens := m.calculateStats()

	// Should handle large files correctly
	if totalBytes < 10*1024*1024 {
		t.Fatalf("expected bytes to include large file size")
	}
	if totalTokens <= 0 {
		t.Fatalf("expected positive token count for large content")
	}
}

func TestReviewModel_ViewportInitialization(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "task", "rules", "")

	m.SetSize(80, 24)

	if m.viewport.Width != 80 {
		t.Fatalf("expected viewport width 80, got %d", m.viewport.Width)
	}

	expectedHeight := 24 - footerHeight
	if m.viewport.Height != expectedHeight {
		t.Fatalf("expected viewport height %d, got %d", expectedHeight, m.viewport.Height)
	}

	if !m.viewportReady {
		t.Fatalf("expected viewportReady to be true after SetSize")
	}
}

func TestReviewModel_ScrollKeyHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"down arrow", tea.KeyMsg{Type: tea.KeyDown}},
		{"up arrow", tea.KeyMsg{Type: tea.KeyUp}},
		{"j key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}},
		{"k key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}},
		{"pgdown", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
		{"home", tea.KeyMsg{Type: tea.KeyHome}},
		{"end", tea.KeyMsg{Type: tea.KeyEnd}},
		{"g key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}},
		{"G key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewReview(nil, nil, nil, "task", "rules", "")
			m.SetSize(80, 24)

			cmd := m.Update(tt.key)
			if cmd != nil {
				t.Fatalf("scroll keys should not return commands, got non-nil for %s", tt.name)
			}
		})
	}
}

func TestReviewModel_FooterContainsScrollHint(t *testing.T) {
	t.Parallel()

	selectedFiles := map[string]bool{"/path/to/file.go": true}
	fileTree := &scanner.FileNode{
		Name:  "root",
		Path:  "/path",
		IsDir: true,
		Children: []*scanner.FileNode{
			{Name: "file.go", Path: "/path/to/file.go", IsDir: false, Size: 512},
		},
	}
	tmpl := &template.Template{Name: "Test", Content: "content"}

	m := NewReview(selectedFiles, fileTree, tmpl, "Task", "Rules", "")
	m.SetSize(80, 24)

	view := m.View()

	if !strings.Contains(view, "↑/↓: Scroll") {
		t.Fatalf("expected scroll hint in footer")
	}
}

func TestReviewModel_ExistingKeysStillWork(t *testing.T) {
	t.Parallel()

	m := NewReview(nil, nil, nil, "task", "rules", "")
	m.SetGenerated("/tmp/test.md", true)
	m.SetSize(80, 24)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	cmd := m.Update(msg)

	if cmd == nil {
		t.Fatalf("'c' key should return clipboard copy command when generated")
	}
}
