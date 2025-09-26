package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/diogo464/shotgun-cli/internal/core/template"
	"github.com/diogo464/shotgun-cli/internal/ui/styles"
)

type ReviewModel struct {
	selectedFiles  map[string]bool
	template       *template.Template
	taskDesc       string
	rules          string
	width          int
	height         int
	generated      bool
	generatedPath  string
	clipboardCopied bool
	estimatedSize  string
}

func NewReview(selectedFiles map[string]bool, template *template.Template, taskDesc, rules string) *ReviewModel {
	return &ReviewModel{
		selectedFiles: selectedFiles,
		template:      template,
		taskDesc:      taskDesc,
		rules:         rules,
	}
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

	// Files summary
	fileCount := len(m.selectedFiles)
	totalSize := m.calculateTotalSize()
	content.WriteString(styles.TitleStyle.Render("📁 Selected Files:"))
	content.WriteString(fmt.Sprintf(" %d files (%s)", fileCount, formatSize(totalSize)))
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

	// Estimated output size
	if m.estimatedSize != "" {
		content.WriteString(styles.HelpStyle.Render(fmt.Sprintf("Estimated output size: %s", m.estimatedSize)))
		content.WriteString("\n\n")
	}

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
			content.WriteString(styles.ErrorStyle.Render("⚠ Failed to copy to clipboard"))
		}
		content.WriteString("\n\n")

		shortcuts := []string{
			"Ctrl+Q: Exit",
			"F1: Help",
		}
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

func (m *ReviewModel) calculateTotalSize() int64 {
	// This would be calculated based on actual file sizes
	// For now, return a placeholder
	return int64(len(m.selectedFiles) * 1024) // Rough estimate
}

func formatSize(bytes int64) string {
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