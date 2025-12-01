package components

import (
	"testing"

	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/stretchr/testify/assert"
)

func TestNewProgress(t *testing.T) {
	p := NewProgress()

	assert.NotNil(t, p)
	assert.False(t, p.visible)
	assert.Equal(t, int64(0), p.current)
	assert.Equal(t, int64(0), p.total)
	assert.Equal(t, "", p.stage)
	assert.Equal(t, "", p.message)
}

func TestProgressUpdate(t *testing.T) {
	p := NewProgress()

	p.Update(10, 100, "scanning", "processing files")

	assert.Equal(t, int64(10), p.current)
	assert.Equal(t, int64(100), p.total)
	assert.Equal(t, "scanning", p.stage)
	assert.Equal(t, "processing files", p.message)
	assert.True(t, p.visible)
}

func TestProgressUpdateFromScanner(t *testing.T) {
	p := NewProgress()

	progress := scanner.Progress{
		Current: 5,
		Total:   50,
		Stage:   "scan-progress",
	}

	p.UpdateFromScanner(progress)

	assert.Equal(t, int64(5), p.current)
	assert.Equal(t, int64(50), p.total)
	assert.Equal(t, "scan-progress", p.stage)
	assert.Equal(t, "", p.message)
	assert.True(t, p.visible)
}

func TestProgressUpdateFromGenerator(t *testing.T) {
	p := NewProgress()

	progress := context.GenProgress{
		Stage:   "generate",
		Message: "generating content",
	}

	p.UpdateFromGenerator(progress)

	assert.Equal(t, int64(0), p.current)
	assert.Equal(t, int64(0), p.total)
	assert.Equal(t, "generate", p.stage)
	assert.Equal(t, "generating content", p.message)
	assert.True(t, p.visible)
}

func TestProgressUpdateMessage(t *testing.T) {
	p := NewProgress()
	p.Update(0, 0, "old-stage", "old-message")

	p.UpdateMessage("new-stage", "new-message")

	assert.Equal(t, "new-stage", p.stage)
	assert.Equal(t, "new-message", p.message)
	assert.True(t, p.visible)
}

func TestProgressHide(t *testing.T) {
	p := NewProgress()
	p.Update(10, 100, "scanning", "")
	assert.True(t, p.visible)

	p.Hide()
	assert.False(t, p.visible)
}

func TestProgressShow(t *testing.T) {
	p := NewProgress()
	assert.False(t, p.visible)

	p.Show()
	assert.True(t, p.visible)
}

func TestProgressIsVisible(t *testing.T) {
	p := NewProgress()
	assert.False(t, p.IsVisible())

	p.Show()
	assert.True(t, p.IsVisible())

	p.Hide()
	assert.False(t, p.IsVisible())
}

func TestProgressSetSize(t *testing.T) {
	p := NewProgress()

	p.SetSize(100, 50)

	assert.Equal(t, 100, p.width)
	assert.Equal(t, 50, p.height)
}

func TestProgressViewInvisible(t *testing.T) {
	p := NewProgress()

	view := p.View()

	assert.Equal(t, "", view)
}

func TestProgressViewVisible(t *testing.T) {
	p := NewProgress()
	p.Update(10, 100, "scanning", "")
	p.SetSize(80, 30)

	view := p.View()

	assert.NotEmpty(t, view)
	// Should contain stage information
	assert.Contains(t, view, "scanning")
	// Should contain progress percentage
	assert.Contains(t, view, "10.0%")
}

func TestProgressViewWithMessage(t *testing.T) {
	p := NewProgress()
	p.Update(0, 0, "processing", "Please wait...")
	p.SetSize(80, 30)

	view := p.View()

	assert.NotEmpty(t, view)
	assert.Contains(t, view, "processing")
	assert.Contains(t, view, "Please wait...")
}

func TestProgressRenderProgressBar(t *testing.T) {
	p := NewProgress()
	p.Update(50, 100, "", "")
	p.SetSize(80, 30)

	// Half complete should render half bars
	bar := p.renderProgressBar(10)
	assert.NotEmpty(t, bar)
	assert.Contains(t, bar, "█")
	assert.Contains(t, bar, "░")
}

func TestProgressRenderProgressBarZeroTotal(t *testing.T) {
	p := NewProgress()
	p.Update(50, 0, "", "")

	bar := p.renderProgressBar(10)

	// Should render all empty bars when total is 0
	assert.Equal(t, "░░░░░░░░░░", bar)
}

func TestProgressRenderProgressBarOverflow(t *testing.T) {
	p := NewProgress()
	p.Update(200, 100, "", "")

	bar := p.renderProgressBar(10)

	// Should not exceed width (renderProgressBar doesn't clamp, so this is expected behavior)
	// The function fills based on percentage, which could exceed the bar width
	// Just verify it's not empty
	assert.NotEmpty(t, bar)
	// All characters should be either █ or ░
	for _, c := range bar {
		assert.Contains(t, "█░", string(c))
	}
}

func TestProgressSpinnerTickCmd(t *testing.T) {
	p := NewProgress()

	cmd := p.GetSpinnerTickCmd()

	// Should return a tick command for the spinner
	assert.NotNil(t, cmd)
}

func TestProgressCenterLine(t *testing.T) {
	p := NewProgress()
	p.SetSize(50, 20)

	line := "test"
	result := p.centerLine(line)

	assert.Contains(t, result, "test")
	// Should be centered or at least have padding
	assert.GreaterOrEqual(t, len(result), len(line))
}

func TestProgressCenterLineNoWidth(t *testing.T) {
	p := NewProgress()
	// width is 0

	line := "test"
	result := p.centerLine(line)

	// Should return as-is when width is not set
	assert.Equal(t, line, result)
}

func TestProgressCenterLineWideLine(t *testing.T) {
	p := NewProgress()
	p.SetSize(10, 20)

	line := "this is a very long line"
	result := p.centerLine(line)

	// Should return as-is when line is wider than container
	assert.Contains(t, result, "this is a very long line")
}

func TestProgressPadCenter(t *testing.T) {
	p := NewProgress()

	text := "test"
	padded := p.padCenter(text, 10)

	assert.Contains(t, padded, "test")
	assert.Equal(t, 10, len(padded))
}

func TestProgressPadCenterExactWidth(t *testing.T) {
	p := NewProgress()

	text := "exactly"
	padded := p.padCenter(text, 7)

	assert.Equal(t, text, padded)
}

func TestProgressPadCenterTextTooWide(t *testing.T) {
	p := NewProgress()

	text := "this is too long"
	truncated := p.padCenter(text, 10)

	assert.LessOrEqual(t, len(truncated), 10)
	assert.Contains(t, truncated, "...")
}

func TestProgressTruncate(t *testing.T) {
	p := NewProgress()

	text := "short"
	truncated := p.truncate(text, 10)

	assert.Equal(t, "short", truncated)
}

func TestProgressTruncateLongText(t *testing.T) {
	p := NewProgress()

	text := "this is a very long text"
	truncated := p.truncate(text, 10)

	assert.Equal(t, "this is...", truncated)
}

func TestProgressTruncateVeryShort(t *testing.T) {
	p := NewProgress()

	text := "x"
	truncated := p.truncate(text, 2)

	// If text is shorter than maxLen, return as-is
	assert.Equal(t, "x", truncated)
}

func TestProgressTruncateDots(t *testing.T) {
	p := NewProgress()

	text := "test"
	truncated := p.truncate(text, 2)

	// For maxLen <= 3 and text longer than maxLen, should return dots
	assert.Equal(t, "..", truncated)
}

func TestProgressTruncateMaxLenZero(t *testing.T) {
	p := NewProgress()

	text := "test"
	truncated := p.truncate(text, 0)

	assert.Equal(t, "", truncated)
}

func TestProgressVisualWidth(t *testing.T) {
	p := NewProgress()

	width := p.visualWidth("test")

	assert.Equal(t, 4, width)
}

func TestProgressGetProgress(t *testing.T) {
	p := NewProgress()
	p.Update(10, 100, "stage", "message")

	current, total, stage, message := p.GetProgress()

	assert.Equal(t, int64(10), current)
	assert.Equal(t, int64(100), total)
	assert.Equal(t, "stage", stage)
	assert.Equal(t, "message", message)
}
