package styles

import (
	"strings"
	"testing"
)

func TestRenderHeader(t *testing.T) {
	t.Parallel()

	result := RenderHeader(1, "Test Title")

	if !strings.Contains(result, "Step 1/5") {
		t.Fatalf("expected step info in header")
	}
	if !strings.Contains(result, "Test Title") {
		t.Fatalf("expected title in header")
	}
}

func TestRenderHeader_DifferentStep(t *testing.T) {
	t.Parallel()

	result := RenderHeader(3, "Another Title")

	if !strings.Contains(result, "Step 3/5") {
		t.Fatalf("expected step 3 in header")
	}
	if !strings.Contains(result, "Another Title") {
		t.Fatalf("expected title in header")
	}
}

func TestRenderFooter(t *testing.T) {
	t.Parallel()

	shortcuts := []string{"Esc: Back", "Enter: Confirm"}
	result := RenderFooter(shortcuts)

	// Should contain all shortcuts
	if !strings.Contains(result, "Esc: Back") {
		t.Fatalf("expected first shortcut")
	}
	if !strings.Contains(result, "Enter: Confirm") {
		t.Fatalf("expected second shortcut")
	}
	// Should be separated by •
	if !strings.Contains(result, "•") {
		t.Fatalf("expected separator")
	}
}

func TestRenderFooter_Empty(t *testing.T) {
	t.Parallel()

	result := RenderFooter([]string{})

	if result != "" {
		t.Fatalf("expected empty string for empty shortcuts")
	}
}

func TestRenderModal(t *testing.T) {
	t.Parallel()

	content := "Test Content"
	result := RenderModal(content)

	if !strings.Contains(result, "Test Content") {
		t.Fatalf("expected content in modal")
	}
}

func TestRenderError(t *testing.T) {
	t.Parallel()

	result := RenderError("Error message")

	if !strings.Contains(result, "Error message") {
		t.Fatalf("expected error message")
	}
	if !strings.Contains(result, "❌") {
		t.Fatalf("expected error icon")
	}
}

func TestRenderSuccess(t *testing.T) {
	t.Parallel()

	result := RenderSuccess("Success message")

	if !strings.Contains(result, "Success message") {
		t.Fatalf("expected success message")
	}
	if !strings.Contains(result, "✅") {
		t.Fatalf("expected success icon")
	}
}

func TestRenderWarning(t *testing.T) {
	t.Parallel()

	result := RenderWarning("Warning message")

	if !strings.Contains(result, "Warning message") {
		t.Fatalf("expected warning message")
	}
	if !strings.Contains(result, "⚠") {
		t.Fatalf("expected warning icon")
	}
}

func TestRenderProgressBar(t *testing.T) {
	t.Parallel()

	result := RenderProgressBar(50, 100, 10)

	if len(result) == 0 {
		t.Fatalf("expected progress bar")
	}
	// Should contain both filled and unfilled characters
	if !strings.ContainsAny(result, "█░") {
		t.Fatalf("expected progress bar characters")
	}
}

func TestRenderProgressBar_ZeroTotal(t *testing.T) {
	t.Parallel()

	result := RenderProgressBar(50, 0, 10)

	// Should render all empty bars when total is 0
	if result != "░░░░░░░░░░" {
		t.Fatalf("expected all empty bars for zero total")
	}
}

func TestRenderProgressBar_ZeroWidth(t *testing.T) {
	t.Parallel()

	result := RenderProgressBar(50, 100, 0)

	// Should render empty string when width is 0
	if result != "" {
		t.Fatalf("expected empty string for zero width")
	}
}

func TestRenderProgressBar_Overflow(t *testing.T) {
	t.Parallel()

	result := RenderProgressBar(150, 100, 10)

	// Should not exceed width - with overflow (150/100 = 1.5), all should be filled
	filled := strings.Count(result, "█")
	empty := strings.Count(result, "░")
	// Total should be 10, all filled (150 > 100 means 100%+, capped at width)
	if filled+empty != 10 {
		t.Fatalf("expected 10 total bar characters, got %d (filled: %d, empty: %d)", filled+empty, filled, empty)
	}
	if filled != 10 {
		t.Fatalf("expected all 10 characters filled on overflow, got %d filled", filled)
	}
}

func TestRenderBox(t *testing.T) {
	t.Parallel()

	content := "Box content"
	result := RenderBox(content, "Box Title")

	if !strings.Contains(result, "Box content") {
		t.Fatalf("expected content")
	}
	if !strings.Contains(result, "Box Title") {
		t.Fatalf("expected title")
	}
}

func TestRenderBox_NoTitle(t *testing.T) {
	t.Parallel()

	content := "Box content"
	result := RenderBox(content, "")

	if !strings.Contains(result, "Box content") {
		t.Fatalf("expected content")
	}
}

func TestRenderList(t *testing.T) {
	t.Parallel()

	items := []string{"Item 1", "Item 2", "Item 3"}
	result := RenderList(items, 1)

	if !strings.Contains(result, "Item 1") {
		t.Fatalf("expected item 1")
	}
	if !strings.Contains(result, "Item 2") {
		t.Fatalf("expected item 2")
	}
	if !strings.Contains(result, "Item 3") {
		t.Fatalf("expected item 3")
	}
	// Should have selected marker for item 2
	if !strings.Contains(result, "▶") {
		t.Fatalf("expected selection marker")
	}
}

func TestRenderList_Empty(t *testing.T) {
	t.Parallel()

	result := RenderList([]string{}, 0)

	if result != "" {
		t.Fatalf("expected empty string for empty list")
	}
}

func TestRenderList_NoSelection(t *testing.T) {
	t.Parallel()

	items := []string{"Item 1"}
	result := RenderList(items, -1)

	// Should not have selection marker
	if strings.Contains(result, "▶") {
		t.Fatalf("unexpected selection marker")
	}
}

func TestRenderTable(t *testing.T) {
	t.Parallel()

	headers := []string{"Header 1", "Header 2"}
	rows := [][]string{
		{"Row 1 Col 1", "Row 1 Col 2"},
		{"Row 2 Col 1", "Row 2 Col 2"},
	}
	result := RenderTable(headers, rows)

	if !strings.Contains(result, "Header 1") {
		t.Fatalf("expected header 1")
	}
	if !strings.Contains(result, "Row 1 Col 1") {
		t.Fatalf("expected row 1 col 1")
	}
}

func TestRenderTable_EmptyHeaders(t *testing.T) {
	t.Parallel()

	rows := [][]string{
		{"Row 1 Col 1"},
	}
	result := RenderTable([]string{}, rows)

	if result != "" {
		t.Fatalf("expected empty string for empty headers")
	}
}

func TestRenderTable_EmptyRows(t *testing.T) {
	t.Parallel()

	headers := []string{"Header 1"}
	result := RenderTable(headers, [][]string{})

	if result != "" {
		t.Fatalf("expected empty string for empty rows")
	}
}

func TestRenderKeyValue(t *testing.T) {
	t.Parallel()

	result := RenderKeyValue("Key", "Value")

	if !strings.Contains(result, "Key:") {
		t.Fatalf("expected key with colon")
	}
	if !strings.Contains(result, "Value") {
		t.Fatalf("expected value")
	}
}

func TestRenderSeparator(t *testing.T) {
	t.Parallel()

	result := RenderSeparator(20)

	if len(result) == 0 {
		t.Fatalf("expected separator")
	}
	if !strings.ContainsAny(result, "─") {
		t.Fatalf("expected dash characters")
	}
}

func TestRenderSeparator_ZeroWidth(t *testing.T) {
	t.Parallel()

	result := RenderSeparator(0)

	// Should use default width of 50 - count the dash characters
	if strings.Count(result, "─") != 50 {
		t.Fatalf("expected 50 dash characters, got %d", strings.Count(result, "─"))
	}
}

func TestRenderCenter(t *testing.T) {
	t.Parallel()

	result := RenderCenter("Text", 20)

	if !strings.Contains(result, "Text") {
		t.Fatalf("expected text")
	}
	// Should add padding
	if len(result) <= 4 {
		t.Fatalf("expected padding")
	}
}

func TestRenderCenter_ZeroWidth(t *testing.T) {
	t.Parallel()

	text := "Test"
	result := RenderCenter(text, 0)

	// Should return as-is
	if result != text {
		t.Fatalf("expected original text")
	}
}

func TestRenderCenter_WideText(t *testing.T) {
	t.Parallel()

	text := "This is very long text"
	result := RenderCenter(text, 10)

	// Should return as-is when text is wider
	if result != text {
		t.Fatalf("expected original text when text is wider than width")
	}
}

func TestRenderFileTree_DirectoryExpanded(t *testing.T) {
	t.Parallel()

	result := RenderFileTree(true, true, true, "Folder")

	if !strings.Contains(result, "📂") {
		t.Fatalf("expected expanded folder icon")
	}
	if !strings.Contains(result, "Folder") {
		t.Fatalf("expected folder name")
	}
}

func TestRenderFileTree_DirectoryCollapsed(t *testing.T) {
	t.Parallel()

	result := RenderFileTree(true, false, false, "Folder")

	if !strings.Contains(result, "📁") {
		t.Fatalf("expected collapsed folder icon")
	}
}

func TestRenderFileTree_FileSelected(t *testing.T) {
	t.Parallel()

	result := RenderFileTree(false, true, false, "File.txt")

	if !strings.Contains(result, "📄") {
		t.Fatalf("expected file icon")
	}
	if !strings.Contains(result, "[✓]") {
		t.Fatalf("expected selected checkbox")
	}
}

func TestRenderFileTree_FileUnselected(t *testing.T) {
	t.Parallel()

	result := RenderFileTree(false, false, false, "File.txt")

	if !strings.Contains(result, "[ ]") {
		t.Fatalf("expected unselected checkbox")
	}
}

func TestRenderFileName_Selected(t *testing.T) {
	t.Parallel()

	result := RenderFileName("test.txt", SelectionSelected)

	if result == "" {
		t.Fatalf("expected non-empty result")
	}
}

func TestRenderFileName_Partial(t *testing.T) {
	t.Parallel()

	result := RenderFileName("test.txt", SelectionPartial)

	if result == "" {
		t.Fatalf("expected non-empty result")
	}
}

func TestRenderFileName_Unselected(t *testing.T) {
	t.Parallel()

	result := RenderFileName("test.txt", SelectionUnselected)

	if result == "" {
		t.Fatalf("expected non-empty result")
	}
}

func TestRenderFileName_InvalidState(t *testing.T) {
	t.Parallel()

	result := RenderFileName("test.txt", SelectionState(999))

	if result == "" {
		t.Fatalf("expected non-empty result for invalid state")
	}
}
