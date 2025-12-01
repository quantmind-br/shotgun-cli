package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/core/context"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
)

type ProgressModel struct {
	current int64
	total   int64
	stage   string
	message string
	visible bool
	width   int
	height  int
	spinner spinner.Model
}

func NewProgress() *ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.PrimaryColor)

	return &ProgressModel{
		visible: false,
		spinner: s,
	}
}

// Init returns the initial command for the spinner
func (m *ProgressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles spinner updates
func (m *ProgressModel) UpdateSpinner(msg tea.Msg) (*ProgressModel, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
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

	borderColor := styles.BorderColor
	titleColor := styles.PrimaryColor

	// Top border
	topBorder := lipgloss.NewStyle().Foreground(borderColor).Render(
		"╭" + strings.Repeat("─", modalWidth-2) + "╮")
	content.WriteString(m.centerLine(topBorder))
	content.WriteString("\n")

	// Title with spinner
	title := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render("Processing")
	titleLine := lipgloss.NewStyle().Foreground(borderColor).Render("│") +
		m.padCenter(m.spinner.View()+" "+title, modalWidth-2) +
		lipgloss.NewStyle().Foreground(borderColor).Render("│")
	content.WriteString(m.centerLine(titleLine))
	content.WriteString("\n")

	// Empty line
	emptyLine := lipgloss.NewStyle().Foreground(borderColor).Render("│") +
		strings.Repeat(" ", modalWidth-2) +
		lipgloss.NewStyle().Foreground(borderColor).Render("│")
	content.WriteString(m.centerLine(emptyLine))
	content.WriteString("\n")

	// Stage
	if m.stage != "" {
		truncated := m.truncate(m.stage, modalWidth-4)
		stageStyled := lipgloss.NewStyle().Foreground(styles.TextColor).Render(truncated)
		padding := strings.Repeat(" ", modalWidth-4-len(truncated))
		stageLine := lipgloss.NewStyle().Foreground(borderColor).Render("│ ") +
			stageStyled + padding +
			lipgloss.NewStyle().Foreground(borderColor).Render(" │")
		content.WriteString(m.centerLine(stageLine))
		content.WriteString("\n")
	}

	// Progress bar (if we have totals)
	if m.total > 0 {
		progressBar := m.renderProgressBar(modalWidth - 4)
		progressLine := lipgloss.NewStyle().Foreground(borderColor).Render("│ ") +
			progressBar +
			lipgloss.NewStyle().Foreground(borderColor).Render(" │")
		content.WriteString(m.centerLine(progressLine))
		content.WriteString("\n")

		// Percentage and counts
		percentage := float64(m.current) / float64(m.total) * 100
		stats := fmt.Sprintf("%.1f%% (%d/%d)", percentage, m.current, m.total)
		statsStyled := lipgloss.NewStyle().Foreground(styles.AccentColor).Render(stats)
		statsLine := lipgloss.NewStyle().Foreground(borderColor).Render("│") +
			m.padCenter(statsStyled, modalWidth-2) +
			lipgloss.NewStyle().Foreground(borderColor).Render("│")
		content.WriteString(m.centerLine(statsLine))
		content.WriteString("\n")
	} else if m.message != "" {
		// Show message if no progress totals
		truncated := m.truncate(m.message, modalWidth-4)
		messageStyled := lipgloss.NewStyle().Foreground(styles.TextColor).Render(truncated)
		padding := strings.Repeat(" ", modalWidth-4-len(truncated))
		messageLine := lipgloss.NewStyle().Foreground(borderColor).Render("│ ") +
			messageStyled + padding +
			lipgloss.NewStyle().Foreground(borderColor).Render(" │")
		content.WriteString(m.centerLine(messageLine))
		content.WriteString("\n")

		// Activity indicator using spinner
		activityLine := lipgloss.NewStyle().Foreground(borderColor).Render("│") +
			m.padCenter(m.spinner.View(), modalWidth-2) +
			lipgloss.NewStyle().Foreground(borderColor).Render("│")
		content.WriteString(m.centerLine(activityLine))
		content.WriteString("\n")
	}

	// Bottom border
	bottomBorder := lipgloss.NewStyle().Foreground(borderColor).Render(
		"╰" + strings.Repeat("─", modalWidth-2) + "╯")
	content.WriteString(m.centerLine(bottomBorder))

	return content.String()
}

func (m *ProgressModel) renderProgressBar(width int) string {
	if m.total <= 0 {
		return lipgloss.NewStyle().Foreground(styles.MutedColor).Render(strings.Repeat("░", width))
	}

	filled := int(float64(width) * float64(m.current) / float64(m.total))
	if filled > width {
		filled = width
	}

	// Use Unicode block characters with colors
	filledStyle := lipgloss.NewStyle().Foreground(styles.AccentColor)
	emptyStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)

	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", width-filled))
	return bar
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
	textLen := m.visualWidth(text)
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
	// For now, approximate by stripping ANSI codes
	return lipgloss.Width(text)
}

func (m *ProgressModel) GetProgress() (int64, int64, string, string) {
	return m.current, m.total, m.stage, m.message
}

// GetSpinnerTickCmd returns the tick command for the spinner
func (m *ProgressModel) GetSpinnerTickCmd() tea.Cmd {
	return m.spinner.Tick
}
