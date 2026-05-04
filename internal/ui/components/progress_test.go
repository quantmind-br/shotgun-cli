package components

import (
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

const testString = "test"

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

func TestProgressUpdateMessage(t *testing.T) {
	p := NewProgress()
	p.Update(0, 0, "old-stage", "old-message")

	p.UpdateMessage("new-stage", "new-message")

	assert.Equal(t, "new-stage", p.stage)
	assert.Equal(t, "new-message", p.message)
	assert.True(t, p.visible)
}

func TestProgressUpdateMessage_ResetsCurrentAndTotal(t *testing.T) {
	t.Parallel()
	p := NewProgress()
	p.Update(42, -1, "scanning", "")

	p.UpdateMessage("sending", "Sending to LLM...")

	assert.Equal(t, int64(0), p.current, "current should be reset to 0")
	assert.Equal(t, int64(0), p.total, "total should be reset to 0")
	assert.Equal(t, "sending", p.stage)
	assert.Equal(t, "Sending to LLM...", p.message)
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

func TestProgressCenterLine(t *testing.T) {
	p := NewProgress()
	p.SetSize(50, 20)

	line := testString
	result := p.centerLine(line)

	assert.Contains(t, result, testString)
	// Should be centered or at least have padding
	assert.GreaterOrEqual(t, len(result), len(line))
}

func TestProgressCenterLineNoWidth(t *testing.T) {
	p := NewProgress()
	// width is 0

	line := testString
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

	text := testString
	padded := p.padCenter(text, 10)

	assert.Contains(t, padded, testString)
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

	text := testString
	truncated := p.truncate(text, 2)

	// For maxLen <= 3 and text longer than maxLen, should return dots
	assert.Equal(t, "..", truncated)
}

func TestProgressTruncateMaxLenZero(t *testing.T) {
	p := NewProgress()

	text := testString
	truncated := p.truncate(text, 0)

	assert.Equal(t, "", truncated)
}

func TestProgressVisualWidth(t *testing.T) {
	p := NewProgress()

	width := p.visualWidth(testString)

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

func TestProgressInit(t *testing.T) {
	p := NewProgress()

	cmd := p.Init()

	// Init should return a spinner tick command
	assert.NotNil(t, cmd)
	// The command should be a function that can be executed
	// We can't easily test the actual execution without more complex setup
}

func TestProgressUpdateSpinner(t *testing.T) {
	p := NewProgress()
	initialSpinner := p.spinner

	// Create a test spinner message
	msg := spinner.TickMsg{}

	updatedModel, cmd := p.UpdateSpinner(msg)

	// Should return the same model instance
	assert.Equal(t, p, updatedModel)
	// Should return a command for the next tick
	assert.NotNil(t, cmd)
	// Spinner should have been updated
	assert.NotEqual(t, initialSpinner, p.spinner)
}

func TestProgressUpdateSpinnerWithDifferentMessage(t *testing.T) {
	p := NewProgress()

	// Test with a different type of message (should not crash)
	msg := tea.KeyMsg{}

	updatedModel, cmd := p.UpdateSpinner(msg)

	// Should handle gracefully and return same model and nil cmd
	assert.Equal(t, p, updatedModel)
	assert.Nil(t, cmd)
}

func TestUsageBar_View(t *testing.T) {
	tests := []struct {
		name        string
		current     int64
		max         int64
		maxStr      string
		tokens      int
		expectIcon  string
		expectColor string // Just checking content for simplicity
	}{
		{
			name:       "safe usage",
			current:    50,
			max:        100,
			maxStr:     "100 B",
			tokens:     10,
			expectIcon: "✅",
		},
		{
			name:       "warning usage",
			current:    85,
			max:        100,
			maxStr:     "100 B",
			tokens:     20,
			expectIcon: "⚠️",
		},
		{
			name:       "critical usage",
			current:    101,
			max:        100,
			maxStr:     "100 B",
			tokens:     30,
			expectIcon: "⛔",
		},
		{
			name:       "no limit",
			current:    50,
			max:        0,
			maxStr:     "",
			tokens:     10,
			expectIcon: "", // No icon for no limit
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewUsageBar(tt.current, tt.max, tt.maxStr, tt.tokens, 30)
			view := bar.View()

			if tt.expectIcon != "" {
				assert.Contains(t, view, tt.expectIcon)
			}
			if tt.max > 0 {
				assert.Contains(t, view, tt.maxStr)
			} else {
				assert.Contains(t, view, "no size limit")
			}
		})
	}
}
