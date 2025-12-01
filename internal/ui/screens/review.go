package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/quantmind-br/shotgun-cli/internal/core/scanner"
	"github.com/quantmind-br/shotgun-cli/internal/core/template"
	"github.com/quantmind-br/shotgun-cli/internal/core/tokens"
	"github.com/quantmind-br/shotgun-cli/internal/platform/gemini"
	"github.com/quantmind-br/shotgun-cli/internal/ui/styles"
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
	return m
}

func (m *ReviewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ReviewModel) Update(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		return tea.Quit
	}
	return nil
}

func (m *ReviewModel) View() string {
	header := styles.RenderHeader(5, "Review & Generate")

	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n\n")

	// Files summary with token estimate
	fileCount := len(m.selectedFiles)
	content.WriteString(styles.TitleStyle.Render("📁 Selected Files:"))
	content.WriteString(fmt.Sprintf(" %d files (%s / ~%s tokens)",
		fileCount,
		formatSizeHelper(m.totalBytes),
		tokens.FormatTokens(m.totalTokens)))
	content.WriteString("\n\n")

	// Show some selected files (up to 5)
	count := 0
	for filePath := range m.selectedFiles {
		if count >= 5 {
			remaining := fileCount - 5
			content.WriteString(fmt.Sprintf("  ... and %d more files", remaining))
			content.WriteString("\n")
			break
		}
		content.WriteString(fmt.Sprintf("  • %s", filePath))
		content.WriteString("\n")
		count++
	}

	content.WriteString("\n")

	// Template summary
	content.WriteString(styles.TitleStyle.Render("📋 Template:"))
	if m.template != nil {
		content.WriteString(fmt.Sprintf(" %s", m.template.Name))
		content.WriteString("\n")
		if m.template.Description != "" {
			content.WriteString(fmt.Sprintf("  %s", m.template.Description))
			content.WriteString("\n")
		}
	} else {
		content.WriteString(" None selected")
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Task summary
	content.WriteString(styles.TitleStyle.Render("🎯 Task:"))
	content.WriteString("\n")
	taskPreview := m.taskDesc
	if len(taskPreview) > 150 {
		taskPreview = taskPreview[:147] + "..."
	}
	content.WriteString(fmt.Sprintf("  %s", taskPreview))
	content.WriteString("\n\n")

	// Rules summary (if provided)
	if strings.TrimSpace(m.rules) != "" {
		content.WriteString(styles.TitleStyle.Render("📝 Rules & Constraints:"))
		content.WriteString("\n")
		rulesPreview := m.rules
		if len(rulesPreview) > 100 {
			rulesPreview = rulesPreview[:97] + "..."
		}
		content.WriteString(fmt.Sprintf("  %s", rulesPreview))
		content.WriteString("\n\n")
	}

	// Estimated output size summary
	content.WriteString(styles.HelpStyle.Render(
		fmt.Sprintf("Estimated context: %s / ~%s tokens",
			formatSizeHelper(m.totalBytes),
			tokens.FormatTokens(m.totalTokens))))
	content.WriteString("\n\n")

	// Generation status
	if m.generated {
		content.WriteString(styles.SuccessStyle.Render("✅ Context generated successfully!"))
		content.WriteString("\n")
		if m.generatedPath != "" {
			content.WriteString(fmt.Sprintf("📄 Saved to: %s", m.generatedPath))
			content.WriteString("\n")
		}
		if m.clipboardCopied {
			content.WriteString(styles.SuccessStyle.Render("📋 Copied to clipboard"))
		} else {
			content.WriteString(styles.HelpStyle.Render("📋 Clipboard copy failed (file saved successfully)"))
			content.WriteString("\n")
			tip := "   Tip: Copy manually from the file or use 'cat " + m.generatedPath + " | wl-copy'"
			content.WriteString(styles.HelpStyle.Render(tip))
		}
		content.WriteString("\n")

		// Gemini status
		content.WriteString("\n")
		if m.geminiSending {
			content.WriteString(styles.HelpStyle.Render("🤖 Sending to Gemini..."))
		} else if m.geminiComplete {
			content.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("🤖 Gemini response saved to: %s", m.geminiOutputFile)))
			content.WriteString("\n")
			content.WriteString(styles.HelpStyle.Render(fmt.Sprintf("   Response time: %s", m.formatDuration(m.geminiDuration))))
		} else if m.geminiError != nil {
			content.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("🤖 Gemini error: %v", m.geminiError)))
		} else if m.geminiAvailable {
			content.WriteString(styles.HelpStyle.Render("🤖 Gemini: Ready (press F9 to send)"))
		} else {
			content.WriteString(styles.HelpStyle.Render("🤖 Gemini: Not configured"))
		}
		content.WriteString("\n\n")

		shortcuts := []string{
			"Ctrl+Q: Exit",
		}
		if m.geminiAvailable && !m.geminiSending && !m.geminiComplete {
			shortcuts = append([]string{"F9: Send to Gemini"}, shortcuts...)
		}
		shortcuts = append(shortcuts, "F1: Help")
		footer := styles.RenderFooter(shortcuts)
		content.WriteString(footer)
	} else {
		// Pre-generation view
		content.WriteString(styles.HelpStyle.Render("Ready to generate context. This will:"))
		content.WriteString("\n")
		content.WriteString("  • Create a comprehensive context file")
		content.WriteString("\n")
		content.WriteString("  • Save it as shotgun-prompt-YYYYMMDD-HHMMSS.md")
		content.WriteString("\n")
		content.WriteString("  • Copy the content to your clipboard")
		content.WriteString("\n\n")

		shortcuts := []string{
			"F8: Generate",
			"F10: Back to Edit",
			"F1: Help",
			"Ctrl+Q: Quit",
		}
		footer := styles.RenderFooter(shortcuts)
		content.WriteString(footer)
	}

	return content.String()
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
