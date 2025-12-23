package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
	"github.com/spf13/viper"
)

type ReviewModel struct {
	selectedFiles   map[string]bool
	fileTree        *scanner.FileNode
	template        *template.Template
	taskDesc        string
	rules           string
	width           int
	height          int
	generated       bool
	generatedPath   string
	clipboardCopied bool

	// Cached stats calculated on initialization
	totalBytes  int64
	totalTokens int

	// Max size from config
	maxSizeBytes int64
	maxSizeStr   string

	// Gemini integration state
	geminiAvailable  bool
	geminiSending    bool
	geminiComplete   bool
	geminiOutputFile string
	geminiDuration   time.Duration
	geminiError      error
}

// NewReview creates a new ReviewModel with precomputed statistics.
// The fileTree is required for accurate size calculations.
func NewReview(
	selectedFiles map[string]bool,
	fileTree *scanner.FileNode,
	tmpl *template.Template,
	taskDesc, rules string,
) *ReviewModel {
	m := &ReviewModel{
		selectedFiles:   selectedFiles,
		fileTree:        fileTree,
		template:        tmpl,
		taskDesc:        taskDesc,
		rules:           rules,
		geminiAvailable: gemini.IsAvailable() && gemini.IsConfigured(),
	}
	// Calculate stats on initialization (once)
	m.totalBytes, m.totalTokens = m.calculateStats()

	// Get max size from config
	m.maxSizeStr = viper.GetString("context.max-size")
	if m.maxSizeStr != "" {
		m.maxSizeBytes, _ = parseSize(m.maxSizeStr)
	}

	return m
}

func (m *ReviewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ReviewModel) Update(msg tea.KeyMsg) tea.Cmd {
	if msg.String() == "ctrl+c" {
		return tea.Quit
	}

	return nil
}

func (m *ReviewModel) View() string {
	header := styles.RenderHeader(5, "Review & Generate")

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")

	// Files summary with styled token estimate
	fileCount := len(m.selectedFiles)
	filesIcon := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("📁")
	filesLabel := styles.TitleStyle.Render("Selected Files:")
	filesStats := styles.RenderTokenStats(fileCount, formatSizeHelper(m.totalBytes), tokens.FormatTokens(m.totalTokens))

	content.WriteString(filesIcon + " " + filesLabel + " " + filesStats)
	content.WriteString("\n\n")

	// Show some selected files (up to 5) with styling
	count := 0
	for filePath := range m.selectedFiles {
		if count >= 5 {
			remaining := fileCount - 5
			moreText := lipgloss.NewStyle().Foreground(styles.MutedColor).Italic(true).
				Render(fmt.Sprintf("  ... and %d more files", remaining))
			content.WriteString(moreText)
			content.WriteString("\n")

			break
		}
		bullet := lipgloss.NewStyle().Foreground(styles.MutedColor).Render("  • ")
		pathStyled := styles.PathStyle.Render(filePath)
		content.WriteString(bullet + pathStyled)
		content.WriteString("\n")
		count++
	}

	content.WriteString("\n")

	// Template summary
	tmplIcon := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("📋")
	tmplLabel := styles.TitleStyle.Render("Template:")
	content.WriteString(tmplIcon + " " + tmplLabel)
	if m.template != nil {
		tmplName := styles.StatsValueStyle.Render(m.template.Name)
		content.WriteString(" " + tmplName)
		content.WriteString("\n")
		if m.template.Description != "" {
			descStyled := styles.HelpStyle.Render("  " + m.template.Description)
			content.WriteString(descStyled)
			content.WriteString("\n")
		}
	} else {
		noTemplate := styles.WarningStyle.Render(" None selected")
		content.WriteString(noTemplate)
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Task summary
	taskIcon := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("🎯")
	taskLabel := styles.TitleStyle.Render("Task:")
	content.WriteString(taskIcon + " " + taskLabel)
	content.WriteString("\n")
	taskPreview := m.taskDesc
	if len(taskPreview) > 150 {
		taskPreview = taskPreview[:147] + "..."
	}
	taskStyled := lipgloss.NewStyle().Foreground(styles.TextColor).Render("  " + taskPreview)
	content.WriteString(taskStyled)
	content.WriteString("\n\n")

	// Rules summary (if provided)
	if strings.TrimSpace(m.rules) != "" {
		rulesIcon := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("📝")
		rulesLabel := styles.TitleStyle.Render("Rules & Constraints:")
		content.WriteString(rulesIcon + " " + rulesLabel)
		content.WriteString("\n")
		rulesPreview := m.rules
		if len(rulesPreview) > 100 {
			rulesPreview = rulesPreview[:97] + "..."
		}
		rulesStyled := lipgloss.NewStyle().Foreground(styles.TextColor).Render("  " + rulesPreview)
		content.WriteString(rulesStyled)
		content.WriteString("\n\n")
	}

	// Token/Size limits section with visual indicator
	content.WriteString(m.renderSizeLimitSection())
	content.WriteString("\n")

	// Generation status
	if m.generated {
		content.WriteString(m.renderPostGenerationView())
	} else {
		content.WriteString(m.renderPreGenerationView())
	}

	return content.String()
}

func (m *ReviewModel) renderSizeLimitSection() string {
	var section strings.Builder

	limitIcon := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("📊")
	limitLabel := styles.TitleStyle.Render("Context Size:")
	section.WriteString(limitIcon + " " + limitLabel)
	section.WriteString("\n")

	// Current size
	currentSize := formatSizeHelper(m.totalBytes)
	currentTokens := tokens.FormatTokens(m.totalTokens)

	if m.maxSizeBytes > 0 {
		// Calculate percentage
		percentage := float64(m.totalBytes) / float64(m.maxSizeBytes) * 100

		// Determine status color
		var statusStyle lipgloss.Style
		var statusIcon string
		if percentage > 100 {
			statusStyle = lipgloss.NewStyle().Foreground(styles.ErrorColor)
			statusIcon = "⛔"
		} else if percentage > 80 {
			statusStyle = lipgloss.NewStyle().Foreground(styles.WarningColor)
			statusIcon = "⚠️"
		} else {
			statusStyle = lipgloss.NewStyle().Foreground(styles.SuccessColor)
			statusIcon = "✅"
		}

		// Size bar visualization
		barWidth := 30
		filledWidth := int(float64(barWidth) * percentage / 100)
		if filledWidth > barWidth {
			filledWidth = barWidth
		}

		filledStyle := statusStyle
		emptyStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)

		bar := filledStyle.Render(strings.Repeat("█", filledWidth)) +
			emptyStyle.Render(strings.Repeat("░", barWidth-filledWidth))

		sizeInfo := fmt.Sprintf("  %s %s / %s (%.1f%%) ~%s tokens",
			statusIcon, currentSize, m.maxSizeStr, percentage, currentTokens)
		section.WriteString(statusStyle.Render(sizeInfo))
		section.WriteString("\n")
		section.WriteString("  " + bar)
	} else {
		// No limit configured
		sizeInfo := fmt.Sprintf("  %s / ~%s tokens", currentSize, currentTokens)
		section.WriteString(styles.StatsValueStyle.Render(sizeInfo))
		section.WriteString("\n")
		noLimit := styles.HelpStyle.Render("  (no size limit configured)")
		section.WriteString(noLimit)
	}

	return section.String()
}

func (m *ReviewModel) renderPostGenerationView() string {
	var view strings.Builder

	// Success message
	view.WriteString(styles.RenderSuccess("Context generated successfully!"))
	view.WriteString("\n\n")

	// File saved info
	if m.generatedPath != "" {
		fileIcon := lipgloss.NewStyle().Foreground(styles.AccentColor).Render("📄")
		pathStyled := styles.PathStyle.Render(m.generatedPath)
		view.WriteString(fileIcon + " Saved to: " + pathStyled)
		view.WriteString("\n")
	}

	// Clipboard status with better visual
	if m.clipboardCopied {
		clipIcon := lipgloss.NewStyle().Foreground(styles.SuccessColor).Render("📋")
		clipText := styles.SuccessStyle.Render("Copied to clipboard")
		view.WriteString(clipIcon + " " + clipText)
	} else {
		clipIcon := lipgloss.NewStyle().Foreground(styles.WarningColor).Render("📋")
		clipText := styles.WarningStyle.Render("Clipboard copy failed")
		view.WriteString(clipIcon + " " + clipText)
		view.WriteString("\n")
		tip := styles.HelpStyle.Render("   Tip: Copy manually with 'cat " + m.generatedPath + " | wl-copy'")
		view.WriteString(tip)
	}
	view.WriteString("\n\n")

	// Gemini status section
	view.WriteString(m.renderGeminiStatus())
	view.WriteString("\n")

	// Footer
	line1 := []string{
		"c: Copy to Clipboard",
	}
	if m.geminiAvailable && !m.geminiSending && !m.geminiComplete {
		line1 = append(line1, "F9: Send to Gemini")
	}
	line2 := []string{
		"F1: Help",
		"Ctrl+Q: Exit",
	}
	footer := styles.RenderFooter(line1) + "\n" + styles.RenderFooter(line2)
	view.WriteString(footer)

	return view.String()
}

func (m *ReviewModel) renderGeminiStatus() string {
	var status strings.Builder

	geminiIcon := lipgloss.NewStyle().Foreground(styles.Nord15).Render("🤖")
	geminiLabel := styles.TitleStyle.Render("Gemini Integration:")
	status.WriteString(geminiIcon + " " + geminiLabel)
	status.WriteString("\n")

	if m.geminiSending {
		// Sending state with spinner-like indicator
		sendingStyle := lipgloss.NewStyle().Foreground(styles.PrimaryColor)
		status.WriteString("  " + sendingStyle.Render("⏳ Sending to Gemini..."))
	} else if m.geminiComplete {
		// Complete state
		successIcon := lipgloss.NewStyle().Foreground(styles.SuccessColor).Render("✔")
		pathStyled := styles.PathStyle.Render(m.geminiOutputFile)
		status.WriteString("  " + successIcon + " Response saved to: " + pathStyled)
		status.WriteString("\n")

		durationStyled := styles.StatsValueStyle.Render(m.formatDuration(m.geminiDuration))
		status.WriteString("  ⏱ Response time: " + durationStyled)
	} else if m.geminiError != nil {
		// Error state
		errorIcon := lipgloss.NewStyle().Foreground(styles.ErrorColor).Render("✖")
		errorText := styles.ErrorStyle.Render(m.geminiError.Error())
		status.WriteString("  " + errorIcon + " Error: " + errorText)
	} else if m.geminiAvailable {
		// Ready state
		readyStyle := lipgloss.NewStyle().Foreground(styles.AccentColor)
		status.WriteString("  " + readyStyle.Render("● Ready"))
		status.WriteString(" ")
		hint := styles.HelpStyle.Render("(press F9 to send)")
		status.WriteString(hint)
	} else {
		// Not configured
		notConfigured := styles.MutedColor
		status.WriteString("  " + lipgloss.NewStyle().Foreground(notConfigured).Render("○ Not configured"))
	}

	return status.String()
}

func (m *ReviewModel) renderPreGenerationView() string {
	var view strings.Builder

	// Info section
	infoIcon := lipgloss.NewStyle().Foreground(styles.PrimaryColor).Render("ℹ")
	infoLabel := styles.HelpStyle.Render("Ready to generate context. This will:")
	view.WriteString(infoIcon + " " + infoLabel)
	view.WriteString("\n")

	bulletStyle := lipgloss.NewStyle().Foreground(styles.MutedColor)
	itemStyle := lipgloss.NewStyle().Foreground(styles.TextColor)

	items := []string{
		"Create a comprehensive context file",
		"Save it as shotgun-prompt-YYYYMMDD-HHMMSS.md",
		"Copy the content to your clipboard",
	}

	for _, item := range items {
		view.WriteString(bulletStyle.Render("  • ") + itemStyle.Render(item))
		view.WriteString("\n")
	}
	view.WriteString("\n")

	// Footer
	line1 := []string{
		"F7: Back",
		"F8: Generate",
	}
	line2 := []string{
		"F1: Help",
		"Ctrl+Q: Quit",
	}
	footer := styles.RenderFooter(line1) + "\n" + styles.RenderFooter(line2)
	view.WriteString(footer)

	return view.String()
}

func (m *ReviewModel) SetGenerated(filePath string, clipboardSuccess bool) {
	m.generated = true
	m.generatedPath = filePath
	m.clipboardCopied = clipboardSuccess
}

// SetGeminiSending sets the Gemini sending state
func (m *ReviewModel) SetGeminiSending(sending bool) {
	m.geminiSending = sending
	m.geminiError = nil
}

// SetGeminiComplete sets the Gemini complete state
func (m *ReviewModel) SetGeminiComplete(outputFile string, duration time.Duration) {
	m.geminiSending = false
	m.geminiComplete = true
	m.geminiOutputFile = outputFile
	m.geminiDuration = duration
	m.geminiError = nil
}

// SetGeminiError sets the Gemini error state
func (m *ReviewModel) SetGeminiError(err error) {
	m.geminiSending = false
	m.geminiComplete = false
	m.geminiError = err
}

// formatDuration formats a duration for display
func (m *ReviewModel) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}

	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

// calculateStats computes accurate byte and token counts from selected files.
// It traverses the fileTree and sums sizes only for files in selectedFiles.
// It also includes overhead from task description, rules, and template.
func (m *ReviewModel) calculateStats() (int64, int) {
	var fileBytes int64

	// Sum sizes of selected files by traversing the tree
	if m.fileTree != nil {
		m.walkTree(m.fileTree, func(node *scanner.FileNode, path string) {
			if !node.IsDir && m.selectedFiles[path] {
				fileBytes += node.Size
			}
		})
	}

	// Add overhead from task and rules
	overhead := int64(len(m.taskDesc) + len(m.rules))

	// Add template content overhead (approximate)
	if m.template != nil {
		overhead += int64(len(m.template.Content))
	}

	totalBytes := fileBytes + overhead
	totalTokens := tokens.EstimateFromBytes(totalBytes)

	return totalBytes, totalTokens
}

// walkTree recursively visits all nodes in the file tree
func (m *ReviewModel) walkTree(node *scanner.FileNode, fn func(*scanner.FileNode, string)) {
	fn(node, node.Path)
	for _, child := range node.Children {
		m.walkTree(child, fn)
	}
}

func formatSizeHelper(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseSize converts size strings like "10MB" to bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))

	var multiplier int64 = 1
	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "B") {
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	var size int64
	_, err := fmt.Sscanf(sizeStr, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size value: %w", err)
	}

	return size * multiplier, nil
}
