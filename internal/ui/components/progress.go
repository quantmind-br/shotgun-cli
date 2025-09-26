package components

import (
	"fmt"
	"math"
	"strings"

	"github.com/diogo464/shotgun-cli/internal/core/scanner"
	"github.com/diogo464/shotgun-cli/internal/core/context"
	"github.com/diogo464/shotgun-cli/internal/ui/styles"
)

type ProgressModel struct {
	current int64
	total   int64
	stage   string
	message string
	visible bool
	width   int
	height  int
}

func NewProgress() *ProgressModel {
	return &ProgressModel{
		visible: false,
	}
}

func (m *ProgressModel) Update(current, total int64, stage, message string) {
	m.current = current
	m.total = total
	m.stage = stage
	m.message = message
	m.visible = true
}

func (m *ProgressModel) UpdateFromScanner(progress scanner.Progress) {
	m.Update(progress.Current, progress.Total, progress.Stage, "")
}

func (m *ProgressModel) UpdateFromGenerator(progress context.GenProgress) {
	m.Update(0, 0, progress.Stage, progress.Message)
}

func (m *ProgressModel) UpdateMessage(stage, message string) {
	m.stage = stage
	m.message = message
	m.visible = true
}

func (m *ProgressModel) Hide() {
	m.visible = false
}

func (m *ProgressModel) Show() {
	m.visible = true
}

func (m *ProgressModel) IsVisible() bool {
	return m.visible
}

func (m *ProgressModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ProgressModel) View() string {
	if !m.visible {
		return ""
	}

	// Modal dimensions
	modalWidth := 60
	modalHeight := 8

	if m.width > 0 && modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}
	if modalWidth < 30 {
		modalWidth = 30
	}

	// Center the modal
	var content strings.Builder

	// Add vertical padding to center the modal
	if m.height > modalHeight {
		padding := (m.height - modalHeight) / 2
		for i := 0; i < padding; i++ {
			content.WriteString("\n")
		}
	}

	// Top border
	topBorder := "┌" + strings.Repeat("─", modalWidth-2) + "┐"
	content.WriteString(m.centerLine(topBorder))
	content.WriteString("\n")

	// Title
	title := "Processing..."
	titleLine := "│" + m.padCenter(title, modalWidth-2) + "│"
	content.WriteString(m.centerLine(titleLine))
	content.WriteString("\n")

	// Empty line
	emptyLine := "│" + strings.Repeat(" ", modalWidth-2) + "│"
	content.WriteString(m.centerLine(emptyLine))
	content.WriteString("\n")

	// Stage
	if m.stage != "" {
		stageLine := "│ " + m.truncate(m.stage, modalWidth-4) + strings.Repeat(" ", modalWidth-4-len(m.truncate(m.stage, modalWidth-4))) + " │"
		content.WriteString(m.centerLine(stageLine))
		content.WriteString("\n")
	}

	// Progress bar (if we have totals)
	if m.total > 0 {
		progressBar := m.renderProgressBar(modalWidth - 4)
		progressLine := "│ " + progressBar + " │"
		content.WriteString(m.centerLine(progressLine))
		content.WriteString("\n")

		// Percentage and counts
		percentage := float64(m.current) / float64(m.total) * 100
		stats := fmt.Sprintf("%.1f%% (%d/%d)", percentage, m.current, m.total)
		statsLine := "│ " + m.padCenter(stats, modalWidth-4) + " │"
		content.WriteString(m.centerLine(statsLine))
		content.WriteString("\n")
	} else if m.message != "" {
		// Show message if no progress totals
		messageLine := "│ " + m.truncate(m.message, modalWidth-4) + strings.Repeat(" ", modalWidth-4-len(m.truncate(m.message, modalWidth-4))) + " │"
		content.WriteString(m.centerLine(messageLine))
		content.WriteString("\n")

		// Spinner or activity indicator
		spinner := m.getSpinner()
		spinnerLine := "│" + m.padCenter(spinner, modalWidth-2) + "│"
		content.WriteString(m.centerLine(spinnerLine))
		content.WriteString("\n")
	}

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", modalWidth-2) + "┘"
	content.WriteString(m.centerLine(bottomBorder))

	return styles.BorderStyle.Render(content.String())
}

func (m *ProgressModel) renderProgressBar(width int) string {
	if m.total <= 0 {
		return strings.Repeat("░", width)
	}

	filled := int(float64(width) * float64(m.current) / float64(m.total))
	if filled > width {
		filled = width
	}

	// Use Unicode block characters for better visual
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

func (m *ProgressModel) getSpinner() string {
	// Simple spinner animation (would need to be updated with a timer in real implementation)
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	// For now, just return a static spinner
	return spinners[0]
}

func (m *ProgressModel) centerLine(line string) string {
	if m.width <= 0 {
		return line
	}

	lineWidth := m.visualWidth(line)
	if lineWidth >= m.width {
		return line
	}

	padding := (m.width - lineWidth) / 2
	return strings.Repeat(" ", padding) + line
}

func (m *ProgressModel) padCenter(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return m.truncate(text, width)
	}

	leftPad := (width - textLen) / 2
	rightPad := width - textLen - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

func (m *ProgressModel) truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}
	return text[:maxLen-3] + "..."
}

func (m *ProgressModel) visualWidth(text string) int {
	// Simple implementation - in a real scenario, you'd want to handle
	// Unicode characters and ANSI escape sequences properly
	return len(text)
}

func (m *ProgressModel) GetProgress() (int64, int64, string, string) {
	return m.current, m.total, m.stage, m.message
}